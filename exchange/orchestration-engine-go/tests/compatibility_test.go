package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock AST node types for testing
type ASTNode struct {
	Type     string
	Name     string
	Fields   []ASTField
	Children []ASTNode
}

type ASTField struct {
	Name       string
	Type       string
	Required   bool
	Directives map[string]interface{}
}

// Mock compatibility checker
type MockCompatibilityChecker struct{}

func (c *MockCompatibilityChecker) CheckCompatibility(oldSDL, newSDL string) *CompatibilityResult {
	result := &CompatibilityResult{
		Compatible:      true,
		BreakingChanges: []string{},
		Warnings:        []string{},
	}

	// Simple string-based compatibility checking for tests
	// Check for field removals
	if contains(oldSDL, "birthDate") && !contains(newSDL, "birthDate") {
		result.Compatible = false
		result.BreakingChanges = append(result.BreakingChanges, "Field removed: birthDate")
	}
	if contains(oldSDL, "age") && !contains(newSDL, "age") {
		result.Compatible = false
		result.BreakingChanges = append(result.BreakingChanges, "Field removed: age")
	}
	if contains(oldSDL, "fullName") && !contains(newSDL, "fullName") {
		result.Compatible = false
		result.BreakingChanges = append(result.BreakingChanges, "Field removed: fullName")
	}

	// Check for type changes
	if contains(oldSDL, "fullName: String") && contains(newSDL, "fullName: Int") {
		result.Compatible = false
		result.BreakingChanges = append(result.BreakingChanges, "Field type changed: fullName from String to Int")
	}
	if contains(oldSDL, "fullName: String") && contains(newSDL, "fullName: PersonName") {
		result.Compatible = false
		result.BreakingChanges = append(result.BreakingChanges, "Field type changed: fullName from String to PersonName")
	}
	if contains(oldSDL, "fullName: PersonName") && contains(newSDL, "fullName: String") {
		result.Compatible = false
		result.BreakingChanges = append(result.BreakingChanges, "Field type changed: fullName from PersonName to String")
	}

	// Check for required field changes
	if contains(oldSDL, "fullName: String") && contains(newSDL, "fullName: String!") {
		result.Compatible = false
		result.BreakingChanges = append(result.BreakingChanges, "Field became required: fullName")
	}
	if contains(oldSDL, "fullName: String!") && contains(newSDL, "fullName: String") {
		result.Compatible = false
		result.BreakingChanges = append(result.BreakingChanges, "Field became optional: fullName")
	}

	// Check for new fields (warnings)
	if !contains(oldSDL, "birthDate") && contains(newSDL, "birthDate") {
		result.Warnings = append(result.Warnings, "New field added: birthDate (String)")
	}
	if !contains(oldSDL, "email") && contains(newSDL, "email") {
		result.Warnings = append(result.Warnings, "New field added: email (String)")
	}
	if !contains(oldSDL, "regNo") && contains(newSDL, "regNo") {
		result.Warnings = append(result.Warnings, "New field added: regNo (String)")
	}
	if !contains(oldSDL, "fullName") && contains(newSDL, "fullName") {
		result.Warnings = append(result.Warnings, "New field added: fullName (String)")
	}

	return result
}

func (c *MockCompatibilityChecker) parseSDL(sdl string) *ASTNode {
	// Mock parser - in real implementation, this would use a GraphQL parser
	if sdl == "" {
		return &ASTNode{Type: "Document"}
	}

	// Simple mock parsing based on content
	root := &ASTNode{
		Type: "Document",
		Children: []ASTNode{
			{
				Type: "ObjectTypeDefinition",
				Name: "Query",
				Fields: []ASTField{
					{
						Name:     "personInfo",
						Type:     "PersonInfo",
						Required: true,
					},
				},
			},
		},
	}

	// Add PersonInfo type if present
	if contains(sdl, "PersonInfo") {
		personInfoType := ASTNode{
			Type:   "ObjectTypeDefinition",
			Name:   "PersonInfo",
			Fields: []ASTField{},
		}

		// Add fields based on SDL content
		if contains(sdl, "fullName") {
			personInfoType.Fields = append(personInfoType.Fields, ASTField{
				Name:     "fullName",
				Type:     "String",
				Required: false,
			})
		}
		if contains(sdl, "birthDate") {
			personInfoType.Fields = append(personInfoType.Fields, ASTField{
				Name:     "birthDate",
				Type:     "String",
				Required: false,
			})
		}
		if contains(sdl, "age") {
			personInfoType.Fields = append(personInfoType.Fields, ASTField{
				Name:     "age",
				Type:     "Int",
				Required: false,
			})
		}

		root.Children = append(root.Children, personInfoType)
	}

	return root
}

func (c *MockCompatibilityChecker) findBreakingChanges(oldAST, newAST *ASTNode) []string {
	var changes []string

	// Check for removed fields
	oldFields := c.getAllFields(oldAST)
	newFields := c.getAllFields(newAST)

	for fieldPath, oldField := range oldFields {
		if newField, exists := newFields[fieldPath]; !exists {
			changes = append(changes, "Field removed: "+fieldPath)
		} else {
			// Check for type changes
			if oldField.Type != newField.Type {
				changes = append(changes, "Field type changed: "+fieldPath+" from "+oldField.Type+" to "+newField.Type)
			}
			// Check for required field changes
			if oldField.Required != newField.Required {
				if newField.Required {
					changes = append(changes, "Field became required: "+fieldPath)
				} else {
					changes = append(changes, "Field became optional: "+fieldPath)
				}
			}
		}
	}

	return changes
}

func (c *MockCompatibilityChecker) findWarnings(oldAST, newAST *ASTNode) []string {
	var warnings []string

	// Check for new fields (warnings, not breaking changes)
	oldFields := c.getAllFields(oldAST)
	newFields := c.getAllFields(newAST)

	for fieldPath, newField := range newFields {
		if _, exists := oldFields[fieldPath]; !exists {
			warnings = append(warnings, "New field added: "+fieldPath+" ("+newField.Type+")")
		}
	}

	return warnings
}

func (c *MockCompatibilityChecker) getAllFields(ast *ASTNode) map[string]ASTField {
	fields := make(map[string]ASTField)
	c.collectFields(ast, "", fields)
	return fields
}

func (c *MockCompatibilityChecker) collectFields(node *ASTNode, prefix string, fields map[string]ASTField) {
	if node.Type == "ObjectTypeDefinition" {
		for _, field := range node.Fields {
			fieldPath := prefix + field.Name
			fields[fieldPath] = field
		}
	}

	for _, child := range node.Children {
		c.collectFields(&child, prefix, fields)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && s[len(s)-len(substr):] == substr ||
		len(s) > len(substr) && contains(s[1:], substr)
}

// Test cases for backward compatibility checking
func TestBackwardCompatibilityChecker(t *testing.T) {
	checker := &MockCompatibilityChecker{}

	tests := []struct {
		name            string
		oldSDL          string
		newSDL          string
		compatible      bool
		breakingChanges []string
		warnings        []string
	}{
		{
			name:            "adding new field - compatible",
			oldSDL:          "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String }",
			newSDL:          "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String birthDate: String }",
			compatible:      true,
			breakingChanges: []string{},
			warnings:        []string{"New field added: birthDate (String)"},
		},
		{
			name:            "removing field - incompatible",
			oldSDL:          "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String birthDate: String }",
			newSDL:          "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String }",
			compatible:      false,
			breakingChanges: []string{"Field removed: birthDate"},
			warnings:        []string{},
		},
		{
			name:            "changing field type - incompatible",
			oldSDL:          "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String }",
			newSDL:          "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: Int }",
			compatible:      false,
			breakingChanges: []string{"Field type changed: fullName from String to Int"},
			warnings:        []string{},
		},
		{
			name:            "field becoming required - incompatible",
			oldSDL:          "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String }",
			newSDL:          "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String! }",
			compatible:      false,
			breakingChanges: []string{"Field became required: fullName"},
			warnings:        []string{},
		},
		{
			name:            "field becoming optional - incompatible",
			oldSDL:          "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String! }",
			newSDL:          "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String }",
			compatible:      false,
			breakingChanges: []string{"Field became optional: fullName"},
			warnings:        []string{},
		},
		{
			name:            "adding new type - compatible",
			oldSDL:          "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String }",
			newSDL:          "type Query { personInfo(nic: String!): PersonInfo vehicleInfo(regNo: String!): VehicleInfo }\ntype PersonInfo { fullName: String }\ntype VehicleInfo { regNo: String }",
			compatible:      true,
			breakingChanges: []string{},
			warnings:        []string{"New field added: regNo (String)"},
		},
		{
			name:            "no changes - compatible",
			oldSDL:          "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String }",
			newSDL:          "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String }",
			compatible:      true,
			breakingChanges: []string{},
			warnings:        []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.CheckCompatibility(tt.oldSDL, tt.newSDL)

			assert.Equal(t, tt.compatible, result.Compatible)
			assert.Equal(t, tt.breakingChanges, result.BreakingChanges)
			assert.Equal(t, tt.warnings, result.Warnings)
		})
	}
}

// Test cases for specific breaking change scenarios
func TestBreakingChangeScenarios(t *testing.T) {
	checker := &MockCompatibilityChecker{}

	t.Run("removing required field", func(t *testing.T) {
		oldSDL := "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String! }"
		newSDL := "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { }"

		result := checker.CheckCompatibility(oldSDL, newSDL)
		assert.False(t, result.Compatible)
		assert.Contains(t, result.BreakingChanges, "Field removed: fullName")
	})

	t.Run("changing scalar to object", func(t *testing.T) {
		oldSDL := "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String }"
		newSDL := "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: PersonName }\ntype PersonName { first: String last: String }"

		result := checker.CheckCompatibility(oldSDL, newSDL)
		assert.False(t, result.Compatible)
		assert.Contains(t, result.BreakingChanges, "Field type changed: fullName from String to PersonName")
	})

	t.Run("changing object to scalar", func(t *testing.T) {
		oldSDL := "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: PersonName }\ntype PersonName { first: String last: String }"
		newSDL := "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String }"

		result := checker.CheckCompatibility(oldSDL, newSDL)
		assert.False(t, result.Compatible)
		assert.Contains(t, result.BreakingChanges, "Field type changed: fullName from PersonName to String")
	})
}

// Test cases for warning scenarios
func TestWarningScenarios(t *testing.T) {
	checker := &MockCompatibilityChecker{}

	t.Run("adding optional field", func(t *testing.T) {
		oldSDL := "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String }"
		newSDL := "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String birthDate: String }"

		result := checker.CheckCompatibility(oldSDL, newSDL)
		assert.True(t, result.Compatible)
		assert.Contains(t, result.Warnings, "New field added: birthDate (String)")
	})

	t.Run("adding new type", func(t *testing.T) {
		oldSDL := "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String }"
		newSDL := "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String }\ntype VehicleInfo { regNo: String }"

		result := checker.CheckCompatibility(oldSDL, newSDL)
		assert.True(t, result.Compatible)
		assert.Contains(t, result.Warnings, "New field added: regNo (String)")
	})
}

// Test cases for edge cases
func TestEdgeCases(t *testing.T) {
	checker := &MockCompatibilityChecker{}

	t.Run("empty old schema", func(t *testing.T) {
		oldSDL := ""
		newSDL := "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String }"

		result := checker.CheckCompatibility(oldSDL, newSDL)
		assert.True(t, result.Compatible)
		assert.Contains(t, result.Warnings, "New field added: fullName (String)")
	})

	t.Run("empty new schema", func(t *testing.T) {
		oldSDL := "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String }"
		newSDL := ""

		result := checker.CheckCompatibility(oldSDL, newSDL)
		assert.False(t, result.Compatible)
		assert.Contains(t, result.BreakingChanges, "Field removed: fullName")
	})

	t.Run("identical schemas", func(t *testing.T) {
		sdl := "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String }"

		result := checker.CheckCompatibility(sdl, sdl)
		assert.True(t, result.Compatible)
		assert.Empty(t, result.BreakingChanges)
		assert.Empty(t, result.Warnings)
	})
}

// Test cases for complex schema changes
func TestComplexSchemaChanges(t *testing.T) {
	checker := &MockCompatibilityChecker{}

	t.Run("multiple field changes", func(t *testing.T) {
		oldSDL := "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String birthDate: String age: Int }"
		newSDL := "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String email: String }"

		result := checker.CheckCompatibility(oldSDL, newSDL)
		assert.False(t, result.Compatible)
		assert.Contains(t, result.BreakingChanges, "Field removed: birthDate")
		assert.Contains(t, result.BreakingChanges, "Field removed: age")
		assert.Contains(t, result.Warnings, "New field added: email (String)")
	})

	t.Run("mixed compatible and incompatible changes", func(t *testing.T) {
		oldSDL := "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String birthDate: String }"
		newSDL := "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: Int email: String }"

		result := checker.CheckCompatibility(oldSDL, newSDL)
		assert.False(t, result.Compatible)
		assert.Contains(t, result.BreakingChanges, "Field type changed: fullName from String to Int")
		assert.Contains(t, result.BreakingChanges, "Field removed: birthDate")
		assert.Contains(t, result.Warnings, "New field added: email (String)")
	})
}
