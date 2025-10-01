package tests

import (
	"testing"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/stretchr/testify/assert"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/models"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/services"
)

func TestSchemaService_CheckCompatibility(t *testing.T) {
	tests := []struct {
		name                    string
		version                 string
		currentVersion          string
		currentSDL              string
		newSDL                  string
		expectedError           bool
		expectedType            string
		expectedBreakingChanges []string
		expectedNewFields       []string
	}{
		{
			name:                    "Minor version - add field (compatible)",
			version:                 "1.1.0",
			currentVersion:          "1.0.0",
			currentSDL:              "type Query { hello: String }",
			newSDL:                  "type Query { hello: String, world: String }",
			expectedError:           false,
			expectedType:            "minor",
			expectedBreakingChanges: []string{},
			expectedNewFields:       []string{"New field 'Query.world' added"},
		},
		{
			name:                    "Minor version - remove field (incompatible)",
			version:                 "1.1.0",
			currentVersion:          "1.0.0",
			currentSDL:              "type Query { hello: String, world: String }",
			newSDL:                  "type Query { hello: String }",
			expectedError:           true,
			expectedType:            "minor",
			expectedBreakingChanges: []string{"Field 'Query.world' was removed"},
		},
		{
			name:                    "Minor version - change field type (incompatible)",
			version:                 "1.1.0",
			currentVersion:          "1.0.0",
			currentSDL:              "type Query { hello: String }",
			newSDL:                  "type Query { hello: Int }",
			expectedError:           true,
			expectedType:            "minor",
			expectedBreakingChanges: []string{"Field 'Query.hello' type changed from String to Int"},
		},
		{
			name:                    "Major version - breaking changes allowed",
			version:                 "2.0.0",
			currentVersion:          "1.0.0",
			currentSDL:              "type Query { hello: String }",
			newSDL:                  "type Query { greeting: String }",
			expectedError:           false,
			expectedType:            "major",
			expectedBreakingChanges: []string{"Field 'Query.hello' was removed"},
			expectedNewFields:       []string{"New field 'Query.greeting' added"},
		},
		{
			name:                    "Patch version - no changes (compatible)",
			version:                 "1.0.1",
			currentVersion:          "1.0.0",
			currentSDL:              "type Query { hello: String }",
			newSDL:                  "type Query { hello: String }",
			expectedError:           false,
			expectedType:            "patch",
			expectedBreakingChanges: []string{},
			expectedNewFields:       []string{},
		},
		{
			name:           "Invalid version format",
			version:        "invalid",
			currentVersion: "1.0.0",
			currentSDL:     "type Query { hello: String }",
			newSDL:         "type Query { hello: String }",
			expectedError:  true,
		},
		{
			name:           "Version not higher than current",
			version:        "1.0.0",
			currentVersion: "1.0.0",
			currentSDL:     "type Query { hello: String }",
			newSDL:         "type Query { hello: String }",
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse schemas
			currentSchema, err := parseSDL(tt.currentSDL)
			assert.NoError(t, err)

			newSchema, err := parseSDL(tt.newSDL)
			assert.NoError(t, err)

			// Create service
			service := &services.SchemaService{}

			// Mock getCurrentActiveSchema
			service.GetCurrentActiveSchema = func() (string, *ast.Document, error) {
				return tt.currentVersion, currentSchema, nil
			}

			// Execute test
			compatibility, err := service.CheckCompatibility(tt.version, newSchema)

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedType, compatibility.ChangeType)
				assert.Equal(t, tt.expectedBreakingChanges, compatibility.BreakingChanges)
				assert.Equal(t, tt.expectedNewFields, compatibility.NewFields)
			}
		})
	}
}

func TestSchemaService_CheckMinorCompatibility(t *testing.T) {
	tests := []struct {
		name                    string
		currentSDL              string
		newSDL                  string
		expectedError           bool
		expectedBreakingChanges []string
		expectedNewFields       []string
	}{
		{
			name:                    "Add new field - compatible",
			currentSDL:              "type Query { hello: String }",
			newSDL:                  "type Query { hello: String, world: String }",
			expectedError:           false,
			expectedBreakingChanges: []string{},
			expectedNewFields:       []string{"New field 'Query.world' added"},
		},
		{
			name:                    "Add new type - compatible",
			currentSDL:              "type Query { hello: String }",
			newSDL:                  "type Query { hello: String }\ntype User { name: String }",
			expectedError:           false,
			expectedBreakingChanges: []string{},
			expectedNewFields:       []string{"New type 'User' added"},
		},
		{
			name:                    "Remove field - incompatible",
			currentSDL:              "type Query { hello: String, world: String }",
			newSDL:                  "type Query { hello: String }",
			expectedError:           true,
			expectedBreakingChanges: []string{"Field 'Query.world' was removed"},
		},
		{
			name:                    "Change field type - incompatible",
			currentSDL:              "type Query { hello: String }",
			newSDL:                  "type Query { hello: Int }",
			expectedError:           true,
			expectedBreakingChanges: []string{"Field 'Query.hello' type changed from String to Int"},
		},
		{
			name:                    "Remove type - incompatible",
			currentSDL:              "type Query { hello: String }\ntype User { name: String }",
			newSDL:                  "type Query { hello: String }",
			expectedError:           true,
			expectedBreakingChanges: []string{"Type 'User' was removed"},
		},
		{
			name:                    "Complex schema - add fields to multiple types",
			currentSDL:              "type Query { hello: String }\ntype User { name: String }",
			newSDL:                  "type Query { hello: String, world: String }\ntype User { name: String, email: String }",
			expectedError:           false,
			expectedBreakingChanges: []string{},
			expectedNewFields:       []string{"New field 'Query.world' added", "New field 'User.email' added"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse schemas
			currentSchema, err := parseSDL(tt.currentSDL)
			assert.NoError(t, err)

			newSchema, err := parseSDL(tt.newSDL)
			assert.NoError(t, err)

			// Create service
			service := &services.SchemaService{}

			// Create compatibility object
			compatibility := &models.VersionCompatibility{}

			// Execute test
			err = service.CheckMinorCompatibility(currentSchema, newSchema, compatibility)

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedBreakingChanges, compatibility.BreakingChanges)
			assert.Equal(t, tt.expectedNewFields, compatibility.NewFields)
		})
	}
}

func TestSchemaService_CheckMajorCompatibility(t *testing.T) {
	tests := []struct {
		name                    string
		currentSDL              string
		newSDL                  string
		expectedBreakingChanges []string
		expectedNewFields       []string
		expectedRemovedFields   []string
		expectedModifiedFields  []string
	}{
		{
			name:       "Complete schema replacement",
			currentSDL: "type Query { hello: String }\ntype User { name: String }",
			newSDL:     "type Query { greeting: String }\ntype Person { fullName: String }",
			expectedBreakingChanges: []string{
				"Type 'User' was removed",
				"Field 'Query.hello' was removed",
			},
			expectedNewFields: []string{
				"New type 'Person' added",
				"New field 'Query.greeting' added",
			},
		},
		{
			name:                    "Add new types and fields",
			currentSDL:              "type Query { hello: String }",
			newSDL:                  "type Query { hello: String, world: String }\ntype User { name: String }\ntype Post { title: String }",
			expectedBreakingChanges: []string{},
			expectedNewFields: []string{
				"New type 'User' added",
				"New type 'Post' added",
				"New field 'Query.world' added",
			},
		},
		{
			name:       "Remove all existing content",
			currentSDL: "type Query { hello: String, world: String }\ntype User { name: String }",
			newSDL:     "type Query { }",
			expectedBreakingChanges: []string{
				"Type 'User' was removed",
				"Field 'Query.hello' was removed",
				"Field 'Query.world' was removed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse schemas
			currentSchema, err := parseSDL(tt.currentSDL)
			assert.NoError(t, err)

			newSchema, err := parseSDL(tt.newSDL)
			assert.NoError(t, err)

			// Create service
			service := &services.SchemaService{}

			// Create compatibility object
			compatibility := &models.VersionCompatibility{}

			// Execute test
			service.CheckMajorCompatibility(currentSchema, newSchema, compatibility)

			// Assertions
			assert.Equal(t, tt.expectedBreakingChanges, compatibility.BreakingChanges)
			assert.Equal(t, tt.expectedNewFields, compatibility.NewFields)
		})
	}
}

func TestSchemaService_ParseVersion(t *testing.T) {
	tests := []struct {
		name          string
		version       string
		expectedMajor int
		expectedMinor int
		expectedPatch int
		expectedError bool
	}{
		{
			name:          "Valid semantic version",
			version:       "1.2.3",
			expectedMajor: 1,
			expectedMinor: 2,
			expectedPatch: 3,
			expectedError: false,
		},
		{
			name:          "Version with leading zeros",
			version:       "01.02.03",
			expectedMajor: 1,
			expectedMinor: 2,
			expectedPatch: 3,
			expectedError: false,
		},
		{
			name:          "Invalid format - missing patch",
			version:       "1.2",
			expectedError: true,
		},
		{
			name:          "Invalid format - non-numeric",
			version:       "1.2.x",
			expectedError: true,
		},
		{
			name:          "Empty version",
			version:       "",
			expectedError: true,
		},
		{
			name:          "Invalid format - too many parts",
			version:       "1.2.3.4",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create service
			service := &services.SchemaService{}

			// Execute test
			major, minor, patch, err := service.ParseVersion(tt.version)

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMajor, major)
				assert.Equal(t, tt.expectedMinor, minor)
				assert.Equal(t, tt.expectedPatch, patch)
			}
		})
	}
}

func TestSchemaService_ExtractTypes(t *testing.T) {
	tests := []struct {
		name     string
		sdl      string
		expected map[string]TypeInfo
	}{
		{
			name: "Simple query type",
			sdl:  "type Query { hello: String }",
			expected: map[string]TypeInfo{
				"Query": {
					Fields: map[string]FieldInfo{
						"hello": {Type: "String"},
					},
				},
			},
		},
		{
			name: "Multiple types",
			sdl:  "type Query { hello: String }\ntype User { name: String, age: Int }",
			expected: map[string]TypeInfo{
				"Query": {
					Fields: map[string]FieldInfo{
						"hello": {Type: "String"},
					},
				},
				"User": {
					Fields: map[string]FieldInfo{
						"name": {Type: "String"},
						"age":  {Type: "Int"},
					},
				},
			},
		},
		{
			name: "Complex types with lists and non-null",
			sdl:  "type Query { users: [User!]!, hello: String }",
			expected: map[string]TypeInfo{
				"Query": {
					Fields: map[string]FieldInfo{
						"users": {Type: "[User!]!"},
						"hello": {Type: "String"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse schema
			schema, err := parseSDL(tt.sdl)
			assert.NoError(t, err)

			// Create service
			service := &services.SchemaService{}

			// Execute test
			types := service.ExtractTypes(schema)

			// Assertions
			assert.Equal(t, len(tt.expected), len(types))
			for typeName, expectedType := range tt.expected {
				actualType, exists := types[typeName]
				assert.True(t, exists, "Type %s should exist", typeName)
				assert.Equal(t, expectedType.Fields, actualType.Fields)
			}
		})
	}
}

func TestSchemaService_FieldTypesEqual(t *testing.T) {
	tests := []struct {
		name     string
		type1    string
		type2    string
		expected bool
	}{
		{
			name:     "Identical types",
			type1:    "String",
			type2:    "String",
			expected: true,
		},
		{
			name:     "Different types",
			type1:    "String",
			type2:    "Int",
			expected: false,
		},
		{
			name:     "List vs non-list",
			type1:    "String",
			type2:    "[String]",
			expected: false,
		},
		{
			name:     "Non-null vs nullable",
			type1:    "String",
			type2:    "String!",
			expected: false,
		},
		{
			name:     "Complex types - identical",
			type1:    "[User!]!",
			type2:    "[User!]!",
			expected: true,
		},
		{
			name:     "Complex types - different",
			type1:    "[User!]!",
			type2:    "[Post!]!",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create service
			service := &services.SchemaService{}

			// Execute test
			result := service.FieldTypesEqual(tt.type1, tt.type2)

			// Assertions
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper types for testing
type TypeInfo struct {
	Fields map[string]FieldInfo
}

type FieldInfo struct {
	Type string
}

// Helper function to parse SDL
func parseSDL(sdl string) (*ast.Document, error) {
	parser := &ast.Parser{}
	return parser.ParseString(&ast.Source{Input: sdl})
}
