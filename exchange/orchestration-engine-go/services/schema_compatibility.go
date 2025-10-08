package services

import (
	"fmt"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

// CompatibilityResult represents the result of a schema compatibility check
type CompatibilityResult struct {
	IsCompatible bool                   `json:"is_compatible"`
	Reason       string                 `json:"reason"`
	Changes      map[string]interface{} `json:"changes"`
}

// BreakingChange represents a breaking change in the schema
type BreakingChange struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Field       string `json:"field,omitempty"`
	OldType     string `json:"old_type,omitempty"`
	NewType     string `json:"new_type,omitempty"`
}

// NonBreakingChange represents a non-breaking change in the schema
type NonBreakingChange struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Field       string `json:"field,omitempty"`
}

// FieldDefinition represents a GraphQL field definition
type FieldDefinition struct {
	Name       string               `json:"name"`
	Type       string               `json:"type"`
	IsRequired bool                 `json:"is_required"`
	IsList     bool                 `json:"is_list"`
	IsNonNull  bool                 `json:"is_non_null"`
	Arguments  []ArgumentDefinition `json:"arguments,omitempty"`
}

// ArgumentDefinition represents a GraphQL argument definition
type ArgumentDefinition struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	IsRequired bool   `json:"is_required"`
	IsList     bool   `json:"is_list"`
	IsNonNull  bool   `json:"is_non_null"`
}

// SchemaCompatibilityChecker provides methods for checking GraphQL schema compatibility
type SchemaCompatibilityChecker struct{}

// NewSchemaCompatibilityChecker creates a new schema compatibility checker
func NewSchemaCompatibilityChecker() *SchemaCompatibilityChecker {
	return &SchemaCompatibilityChecker{}
}

// CheckCompatibility performs a comprehensive compatibility check between two GraphQL schemas
func (s *SchemaCompatibilityChecker) CheckCompatibility(oldSDL, newSDL string) *CompatibilityResult {
	result := &CompatibilityResult{
		IsCompatible: true,
		Reason:       "compatible",
		Changes: map[string]interface{}{
			"breaking":     []BreakingChange{},
			"non_breaking": []NonBreakingChange{},
			"warnings":     []string{},
		},
	}

	// Parse both schemas
	oldDoc, err := s.parseSDL(oldSDL)
	if err != nil {
		result.IsCompatible = false
		result.Reason = fmt.Sprintf("failed to parse old schema: %v", err)
		return result
	}

	newDoc, err := s.parseSDL(newSDL)
	if err != nil {
		result.IsCompatible = false
		result.Reason = fmt.Sprintf("failed to parse new schema: %v", err)
		return result
	}

	// Extract type definitions
	oldTypes := s.extractTypeDefinitions(oldDoc)
	newTypes := s.extractTypeDefinitions(newDoc)

	// Check for breaking changes
	breakingChanges := s.checkBreakingChanges(oldTypes, newTypes)
	if len(breakingChanges) > 0 {
		result.IsCompatible = false
		result.Reason = "breaking changes detected"
		result.Changes["breaking"] = breakingChanges
	}

	// Check for non-breaking changes
	nonBreakingChanges := s.checkNonBreakingChanges(oldTypes, newTypes)
	result.Changes["non_breaking"] = nonBreakingChanges

	// Check for warnings
	warnings := s.checkWarnings(oldTypes, newTypes)
	result.Changes["warnings"] = warnings

	return result
}

// parseSDL parses a GraphQL SDL string into an AST document
func (s *SchemaCompatibilityChecker) parseSDL(sdl string) (*ast.Document, error) {
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

// extractTypeDefinitions extracts type definitions from an AST document
func (s *SchemaCompatibilityChecker) extractTypeDefinitions(doc *ast.Document) map[string]*ast.ObjectDefinition {
	types := make(map[string]*ast.ObjectDefinition)

	for _, def := range doc.Definitions {
		if objDef, ok := def.(*ast.ObjectDefinition); ok {
			types[objDef.Name.Value] = objDef
		}
	}

	return types
}

// checkBreakingChanges checks for breaking changes between old and new type definitions
func (s *SchemaCompatibilityChecker) checkBreakingChanges(oldTypes, newTypes map[string]*ast.ObjectDefinition) []BreakingChange {
	var breakingChanges []BreakingChange

	// Check for removed types
	for typeName := range oldTypes {
		if _, exists := newTypes[typeName]; !exists {
			breakingChanges = append(breakingChanges, BreakingChange{
				Type:        "type_removed",
				Description: fmt.Sprintf("Type '%s' has been removed", typeName),
			})
		}
	}

	// Check for changes in existing types
	for typeName, oldType := range oldTypes {
		if newType, exists := newTypes[typeName]; exists {
			// Check for removed fields
			oldFields := s.extractFieldDefinitions(oldType)
			newFields := s.extractFieldDefinitions(newType)

			for fieldName, oldField := range oldFields {
				if newField, exists := newFields[fieldName]; !exists {
					breakingChanges = append(breakingChanges, BreakingChange{
						Type:        "field_removed",
						Description: fmt.Sprintf("Field '%s' has been removed from type '%s'", fieldName, typeName),
						Field:       fieldName,
					})
				} else {
					// Check for type changes
					if s.hasTypeChanged(oldField, newField) {
						breakingChanges = append(breakingChanges, BreakingChange{
							Type:        "field_type_changed",
							Description: fmt.Sprintf("Field '%s' type changed from '%s' to '%s' in type '%s'", fieldName, s.getTypeString(oldField), s.getTypeString(newField), typeName),
							Field:       fieldName,
							OldType:     s.getTypeString(oldField),
							NewType:     s.getTypeString(newField),
						})
					}

					// Check for argument changes
					argChanges := s.checkArgumentChanges(oldField, newField, typeName, fieldName)
					breakingChanges = append(breakingChanges, argChanges...)
				}
			}
		}
	}

	return breakingChanges
}

// checkNonBreakingChanges checks for non-breaking changes between old and new type definitions
func (s *SchemaCompatibilityChecker) checkNonBreakingChanges(oldTypes, newTypes map[string]*ast.ObjectDefinition) []NonBreakingChange {
	var nonBreakingChanges []NonBreakingChange

	// Check for new types
	for typeName := range newTypes {
		if _, exists := oldTypes[typeName]; !exists {
			nonBreakingChanges = append(nonBreakingChanges, NonBreakingChange{
				Type:        "type_added",
				Description: fmt.Sprintf("New type '%s' has been added", typeName),
			})
		}
	}

	// Check for new fields in existing types
	for typeName, oldType := range oldTypes {
		if newType, exists := newTypes[typeName]; exists {
			oldFields := s.extractFieldDefinitions(oldType)
			newFields := s.extractFieldDefinitions(newType)

			for fieldName := range newFields {
				if _, exists := oldFields[fieldName]; !exists {
					nonBreakingChanges = append(nonBreakingChanges, NonBreakingChange{
						Type:        "field_added",
						Description: fmt.Sprintf("New field '%s' has been added to type '%s'", fieldName, typeName),
						Field:       fieldName,
					})
				}
			}
		}
	}

	return nonBreakingChanges
}

// checkWarnings checks for potential issues that should be warned about
func (s *SchemaCompatibilityChecker) checkWarnings(oldTypes, newTypes map[string]*ast.ObjectDefinition) []string {
	var warnings []string

	// Check for deprecated fields (simplified check)
	for _, newType := range newTypes {
		for _, field := range newType.Fields {
			for _, directive := range field.Directives {
				if directive.Name.Value == "deprecated" {
					warnings = append(warnings, fmt.Sprintf("Field '%s' in type '%s' is marked as deprecated", field.Name.Value, newType.Name.Value))
				}
			}
		}
	}

	return warnings
}

// extractFieldDefinitions extracts field definitions from an object type definition
func (s *SchemaCompatibilityChecker) extractFieldDefinitions(objDef *ast.ObjectDefinition) map[string]*ast.FieldDefinition {
	fields := make(map[string]*ast.FieldDefinition)

	for _, field := range objDef.Fields {
		fields[field.Name.Value] = field
	}

	return fields
}

// hasTypeChanged checks if a field's type has changed
func (s *SchemaCompatibilityChecker) hasTypeChanged(oldField, newField *ast.FieldDefinition) bool {
	oldTypeStr := s.getTypeString(oldField)
	newTypeStr := s.getTypeString(newField)
	return oldTypeStr != newTypeStr
}

// getTypeString converts a GraphQL type to a string representation
func (s *SchemaCompatibilityChecker) getTypeString(field *ast.FieldDefinition) string {
	return s.typeToString(field.Type)
}

// typeToString converts an AST type to a string
func (s *SchemaCompatibilityChecker) typeToString(typ ast.Type) string {
	switch t := typ.(type) {
	case *ast.NonNull:
		return s.typeToString(t.Type) + "!"
	case *ast.List:
		return "[" + s.typeToString(t.Type) + "]"
	case *ast.Named:
		return t.Name.Value
	default:
		return "Unknown"
	}
}

// checkArgumentChanges checks for breaking changes in field arguments
func (s *SchemaCompatibilityChecker) checkArgumentChanges(oldField, newField *ast.FieldDefinition, typeName, fieldName string) []BreakingChange {
	var breakingChanges []BreakingChange

	oldArgs := s.extractArgumentDefinitions(oldField)
	newArgs := s.extractArgumentDefinitions(newField)

	// Check for removed arguments
	for argName := range oldArgs {
		if _, exists := newArgs[argName]; !exists {
			breakingChanges = append(breakingChanges, BreakingChange{
				Type:        "argument_removed",
				Description: fmt.Sprintf("Argument '%s' has been removed from field '%s' in type '%s'", argName, fieldName, typeName),
				Field:       fieldName,
			})
		}
	}

	// Check for argument type changes
	for argName, oldArg := range oldArgs {
		if newArg, exists := newArgs[argName]; exists {
			oldTypeStr := s.typeToString(oldArg.Type)
			newTypeStr := s.typeToString(newArg.Type)
			if oldTypeStr != newTypeStr {
				breakingChanges = append(breakingChanges, BreakingChange{
					Type:        "argument_type_changed",
					Description: fmt.Sprintf("Argument '%s' type changed from '%s' to '%s' in field '%s' of type '%s'", argName, oldTypeStr, newTypeStr, fieldName, typeName),
					Field:       fieldName,
					OldType:     oldTypeStr,
					NewType:     newTypeStr,
				})
			}
		}
	}

	// Check for new required arguments
	for argName, newArg := range newArgs {
		if _, exists := oldArgs[argName]; !exists {
			// New argument - check if it's required
			if _, isNonNull := newArg.Type.(*ast.NonNull); isNonNull {
				breakingChanges = append(breakingChanges, BreakingChange{
					Type:        "required_argument_added",
					Description: fmt.Sprintf("New required argument '%s' has been added to field '%s' in type '%s'", argName, fieldName, typeName),
					Field:       fieldName,
				})
			}
		}
	}

	return breakingChanges
}

// extractArgumentDefinitions extracts argument definitions from a field definition
func (s *SchemaCompatibilityChecker) extractArgumentDefinitions(field *ast.FieldDefinition) map[string]*ast.InputValueDefinition {
	args := make(map[string]*ast.InputValueDefinition)

	for _, arg := range field.Arguments {
		args[arg.Name.Value] = arg
	}

	return args
}

// IsBackwardCompatible is a simple wrapper function for backward compatibility
func (s *SchemaCompatibilityChecker) IsBackwardCompatible(oldSDL, newSDL string) bool {
	result := s.CheckCompatibility(oldSDL, newSDL)
	return result.IsCompatible
}
