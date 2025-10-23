package tests

import (
	"io/ioutil"
	"testing"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/federator"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

func TestSchemaResolutionDebug(t *testing.T) {
	// Load the schema from schema.graphql
	schemaContent, err := ioutil.ReadFile("../schema.graphql")
	if err != nil {
		t.Fatalf("Failed to read schema file: %v", err)
	}

	// Parse the schema
	src := source.NewSource(&source.Source{
		Body: []byte(schemaContent),
		Name: "Schema",
	})

	schema, err := parser.Parse(parser.ParseParams{Source: src})
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}

	t.Logf("Schema loaded successfully with %d definitions", len(schema.Definitions))

	// Test cases for schema resolution
	testCases := []struct {
		fieldPath   string
		expected    bool
		description string
	}{
		{
			fieldPath:   "vehicleInfo.regNo",
			expected:    true,
			description: "Simple field in VehicleInfo type",
		},
		{
			fieldPath:   "vehicleInfo.class",
			expected:    true,
			description: "Array field in VehicleInfo type",
		},
		{
			fieldPath:   "VehicleClass.className",
			expected:    true,
			description: "Field in VehicleClass type",
		},
		{
			fieldPath:   "vehicleInfo.class.className",
			expected:    true,
			description: "Nested field in array type",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			t.Logf("Testing field path: %s", tc.fieldPath)

			// Test schema resolution
			sourceInfo := federator.ExtractSourceInfoFromSchema(schema, tc.fieldPath)

			if tc.expected {
				if sourceInfo == nil {
					t.Errorf("Expected to find source info for path '%s', but got nil", tc.fieldPath)
				} else {
					t.Logf("✅ Found source info: ProviderKey=%s, ProviderField=%s",
						sourceInfo.ProviderKey, sourceInfo.ProviderField)
				}
			} else {
				if sourceInfo != nil {
					t.Errorf("Expected no source info for path '%s', but got: %+v", tc.fieldPath, sourceInfo)
				} else {
					t.Logf("✅ Correctly found no source info for path '%s'", tc.fieldPath)
				}
			}
		})
	}
}

func TestSchemaFieldDefinitions(t *testing.T) {
	// Load and parse schema
	schemaContent, err := ioutil.ReadFile("../schema.graphql")
	if err != nil {
		t.Fatalf("Failed to read schema file: %v", err)
	}

	src := source.NewSource(&source.Source{
		Body: []byte(schemaContent),
		Name: "Schema",
	})

	schema, err := parser.Parse(parser.ParseParams{Source: src})
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}

	// Debug: Print all type definitions
	t.Logf("=== Schema Type Definitions ===")
	for _, def := range schema.Definitions {
		if objType, ok := def.(*ast.ObjectDefinition); ok {
			t.Logf("Type: %s", objType.Name.Value)
			for _, field := range objType.Fields {
				hasDirectives := len(field.Directives) > 0
				t.Logf("  - Field: %s (has directives: %v)", field.Name.Value, hasDirectives)
				if hasDirectives {
					for _, dir := range field.Directives {
						t.Logf("    Directive: %s", dir.Name.Value)
					}
				}
			}
		}
	}

	// Debug: Test specific field lookups
	t.Logf("=== Testing Field Lookups ===")

	// Test VehicleInfo.class field
	fieldDef := federator.FindFieldDefinitionInSchema(schema, "VehicleInfo", "class")
	if fieldDef != nil {
		t.Logf("✅ Found VehicleInfo.class field")
		sourceInfo := federator.ExtractSourceInfoFromSchemaField(fieldDef)
		if sourceInfo != nil {
			t.Logf("  Source info: %+v", sourceInfo)
		} else {
			t.Logf("  ❌ No source info found")
		}
	} else {
		t.Logf("❌ VehicleInfo.class field not found")
	}

	// Test VehicleClass type existence
	vehicleClassFound := false
	for _, def := range schema.Definitions {
		if objType, ok := def.(*ast.ObjectDefinition); ok {
			if objType.Name.Value == "VehicleClass" {
				vehicleClassFound = true
				t.Logf("✅ Found VehicleClass type")
				break
			}
		}
	}
	if !vehicleClassFound {
		t.Logf("❌ VehicleClass type not found in schema")
	}
}
