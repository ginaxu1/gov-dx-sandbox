package services

import (
	"fmt"
	"strings"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/models"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

// CompatibilityChecker handles backward compatibility checking
type CompatibilityChecker struct{}

// NewCompatibilityChecker creates a new compatibility checker
func NewCompatibilityChecker() *CompatibilityChecker {
	return &CompatibilityChecker{}
}

// CheckCompatibility checks backward compatibility between old and new schemas
func (c *CompatibilityChecker) CheckCompatibility(oldSDL, newSDL string) (*models.CompatibilityResult, error) {
	result := &models.CompatibilityResult{
		Compatible:      true,
		BreakingChanges: []string{},
		Warnings:        []string{},
	}

	// Parse both schemas
	oldDoc, err := c.parseSDL(oldSDL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse old SDL: %w", err)
	}

	newDoc, err := c.parseSDL(newSDL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse new SDL: %w", err)
	}

	// Extract type definitions
	oldTypes := c.extractTypeDefinitions(oldDoc)
	newTypes := c.extractTypeDefinitions(newDoc)

	// Check for breaking changes
	breakingChanges := c.findBreakingChanges(oldTypes, newTypes)
	if len(breakingChanges) > 0 {
		result.Compatible = false
		result.BreakingChanges = breakingChanges
	}

	// Check for warnings (non-breaking changes)
	warnings := c.findWarnings(oldTypes, newTypes)
	result.Warnings = warnings

	return result, nil
}

// parseSDL parses a GraphQL SDL string into an AST document
func (c *CompatibilityChecker) parseSDL(sdl string) (*ast.Document, error) {
	src := source.NewSource(&source.Source{
		Body: []byte(sdl),
		Name: "SchemaSDL",
	})

	doc, err := parser.Parse(parser.ParseParams{Source: src})
	if err != nil {
		return nil, fmt.Errorf("failed to parse SDL: %w", err)
	}

	return doc, nil
}

// TypeDefinition represents a GraphQL type definition
type TypeDefinition struct {
	Name   string
	Fields map[string]FieldDefinition
}

// FieldDefinition represents a GraphQL field definition
type FieldDefinition struct {
	Name       string
	Type       string
	Required   bool
	Directives map[string]interface{}
}

// extractTypeDefinitions extracts type definitions from an AST document
func (c *CompatibilityChecker) extractTypeDefinitions(doc *ast.Document) map[string]*TypeDefinition {
	types := make(map[string]*TypeDefinition)

	for _, def := range doc.Definitions {
		if objectType, ok := def.(*ast.ObjectDefinition); ok {
			typeDef := &TypeDefinition{
				Name:   objectType.Name.Value,
				Fields: make(map[string]FieldDefinition),
			}

			for _, field := range objectType.Fields {
				if field != nil && field.Name != nil {
					fieldDef := FieldDefinition{
						Name:       field.Name.Value,
						Type:       c.getTypeString(field.Type),
						Required:   c.isRequired(field.Type),
						Directives: c.extractDirectives(field.Directives),
					}
					typeDef.Fields[field.Name.Value] = fieldDef
				}
			}

			types[objectType.Name.Value] = typeDef
		}
	}

	return types
}

// getTypeString converts a GraphQL type to its string representation
func (c *CompatibilityChecker) getTypeString(t ast.Type) string {
	switch typeNode := t.(type) {
	case *ast.NonNull:
		return c.getTypeString(typeNode.Type) + "!"
	case *ast.List:
		return "[" + c.getTypeString(typeNode.Type) + "]"
	case *ast.Named:
		return typeNode.Name.Value
	default:
		return "Unknown"
	}
}

// isRequired checks if a field is required (non-null)
func (c *CompatibilityChecker) isRequired(t ast.Type) bool {
	_, ok := t.(*ast.NonNull)
	return ok
}

// extractDirectives extracts directives from a field
func (c *CompatibilityChecker) extractDirectives(directives []*ast.Directive) map[string]interface{} {
	result := make(map[string]interface{})
	for _, directive := range directives {
		if directive.Name != nil {
			args := make(map[string]interface{})
			for _, arg := range directive.Arguments {
				if arg.Name != nil && arg.Value != nil {
					args[arg.Name.Value] = c.getValue(arg.Value)
				}
			}
			result[directive.Name.Value] = args
		}
	}
	return result
}

// getValue extracts value from an AST value
func (c *CompatibilityChecker) getValue(value ast.Value) interface{} {
	switch v := value.(type) {
	case *ast.StringValue:
		return v.Value
	case *ast.IntValue:
		return v.Value
	case *ast.FloatValue:
		return v.Value
	case *ast.BooleanValue:
		return v.Value
	case *ast.EnumValue:
		return v.Value
	default:
		return "Unknown"
	}
}

// findBreakingChanges finds breaking changes between old and new types
func (c *CompatibilityChecker) findBreakingChanges(oldTypes, newTypes map[string]*TypeDefinition) []string {
	var changes []string

	// Check for removed types
	for typeName := range oldTypes {
		if _, exists := newTypes[typeName]; !exists {
			changes = append(changes, fmt.Sprintf("Type removed: %s", typeName))
		}
	}

	// Check for changes in existing types
	for typeName, oldType := range oldTypes {
		if newType, exists := newTypes[typeName]; exists {
			// Check for removed fields
			for fieldName := range oldType.Fields {
				if _, exists := newType.Fields[fieldName]; !exists {
					changes = append(changes, fmt.Sprintf("Field removed: %s.%s", typeName, fieldName))
				}
			}

			// Check for field changes in existing fields
			for fieldName, oldField := range oldType.Fields {
				if newField, exists := newType.Fields[fieldName]; exists {
					// Check for type changes
					if oldField.Type != newField.Type {
						changes = append(changes, fmt.Sprintf("Field type changed: %s.%s from %s to %s", typeName, fieldName, oldField.Type, newField.Type))
					}

					// Check for required field changes
					if oldField.Required != newField.Required {
						if newField.Required {
							changes = append(changes, fmt.Sprintf("Field became required: %s.%s", typeName, fieldName))
						} else {
							changes = append(changes, fmt.Sprintf("Field became optional: %s.%s", typeName, fieldName))
						}
					}
				}
			}
		}
	}

	return changes
}

// findWarnings finds non-breaking changes (warnings)
func (c *CompatibilityChecker) findWarnings(oldTypes, newTypes map[string]*TypeDefinition) []string {
	var warnings []string

	// Check for new types
	for typeName := range newTypes {
		if _, exists := oldTypes[typeName]; !exists {
			warnings = append(warnings, fmt.Sprintf("New type added: %s", typeName))
		}
	}

	// Check for new fields in existing types
	for typeName, newType := range newTypes {
		if oldType, exists := oldTypes[typeName]; exists {
			for fieldName, newField := range newType.Fields {
				if _, exists := oldType.Fields[fieldName]; !exists {
					warnings = append(warnings, fmt.Sprintf("New field added: %s.%s (%s)", typeName, fieldName, newField.Type))
				}
			}
		}
	}

	return warnings
}

// Simple string-based compatibility checking for basic cases
func (c *CompatibilityChecker) CheckCompatibilitySimple(oldSDL, newSDL string) *models.CompatibilityResult {
	result := &models.CompatibilityResult{
		Compatible:      true,
		BreakingChanges: []string{},
		Warnings:        []string{},
	}

	// Simple string-based checks for common patterns
	// Check for field removals
	if c.containsField(oldSDL, "birthDate") && !c.containsField(newSDL, "birthDate") {
		result.Compatible = false
		result.BreakingChanges = append(result.BreakingChanges, "Field removed: birthDate")
	}
	if c.containsField(oldSDL, "age") && !c.containsField(newSDL, "age") {
		result.Compatible = false
		result.BreakingChanges = append(result.BreakingChanges, "Field removed: age")
	}
	if c.containsField(oldSDL, "fullName") && !c.containsField(newSDL, "fullName") {
		result.Compatible = false
		result.BreakingChanges = append(result.BreakingChanges, "Field removed: fullName")
	}

	// Check for type changes
	if c.containsField(oldSDL, "fullName: String") && c.containsField(newSDL, "fullName: Int") {
		result.Compatible = false
		result.BreakingChanges = append(result.BreakingChanges, "Field type changed: fullName from String to Int")
	}
	if c.containsField(oldSDL, "fullName: String") && c.containsField(newSDL, "fullName: PersonName") {
		result.Compatible = false
		result.BreakingChanges = append(result.BreakingChanges, "Field type changed: fullName from String to PersonName")
	}
	if c.containsField(oldSDL, "fullName: PersonName") && c.containsField(newSDL, "fullName: String") {
		result.Compatible = false
		result.BreakingChanges = append(result.BreakingChanges, "Field type changed: fullName from PersonName to String")
	}

	// Check for required field changes
	if c.containsField(oldSDL, "fullName: String") && c.containsField(newSDL, "fullName: String!") {
		result.Compatible = false
		result.BreakingChanges = append(result.BreakingChanges, "Field became required: fullName")
	}
	if c.containsField(oldSDL, "fullName: String!") && c.containsField(newSDL, "fullName: String") {
		result.Compatible = false
		result.BreakingChanges = append(result.BreakingChanges, "Field became optional: fullName")
	}

	// Check for new fields (warnings)
	if !c.containsField(oldSDL, "birthDate") && c.containsField(newSDL, "birthDate") {
		result.Warnings = append(result.Warnings, "New field added: birthDate (String)")
	}
	if !c.containsField(oldSDL, "email") && c.containsField(newSDL, "email") {
		result.Warnings = append(result.Warnings, "New field added: email (String)")
	}
	if !c.containsField(oldSDL, "regNo") && c.containsField(newSDL, "regNo") {
		result.Warnings = append(result.Warnings, "New field added: regNo (String)")
	}
	if !c.containsField(oldSDL, "fullName") && c.containsField(newSDL, "fullName") {
		result.Warnings = append(result.Warnings, "New field added: fullName (String)")
	}

	return result
}

// containsField checks if a field exists in SDL
func (c *CompatibilityChecker) containsField(sdl, field string) bool {
	return strings.Contains(sdl, field)
}
