package federator

import (
	"strings"

	"github.com/graphql-go/graphql/language/ast"
)

// FindFieldDefinitionInSchema finds a field definition in the schema by type name and field name
func FindFieldDefinitionInSchema(schema *ast.Document, typeName, fieldName string) *ast.FieldDefinition {
	// Check if typeName is empty to avoid panic
	if len(typeName) == 0 {
		return nil
	}

	for _, def := range schema.Definitions {
		if objType, ok := def.(*ast.ObjectDefinition); ok {
			// Convert to PascalCase for type matching (vehicleInfo -> VehicleInfo)
			pascalTypeName := strings.ToUpper(typeName[:1]) + typeName[1:]
			if objType.Name.Value == pascalTypeName {
				for _, field := range objType.Fields {
					if field.Name.Value == fieldName {
						return field
					}
				}
			}
		}
	}
	return nil
}

// ExtractSourceInfoFromSchema extracts @sourceInfo directive from schema using field path
func ExtractSourceInfoFromSchema(schema *ast.Document, fieldPath string) *SourceInfo {
	// Parse field path like "vehicleInfo.regNo" or "vehicleInfo.class.className" or "VehicleClass.className"
	parts := strings.Split(fieldPath, ".")
	if len(parts) < 2 {
		return nil
	}

	// Handle different path formats
	switch len(parts) {
	case 2:
		// Direct type.field format (e.g., "VehicleClass.className", "vehicleInfo.regNo")
		return findAndExtractSourceInfo(schema, parts[0], parts[1])

	case 3:
		// Nested path format (e.g., "vehicleInfo.class.className")
		// This means we have: parentType.arrayField.nestedField
		parentType := parts[0]  // "vehicleInfo"
		arrayField := parts[1]  // "class"
		nestedField := parts[2] // "className"

		// First, find the array field in the parent type to determine the array element type
		arrayFieldDef := FindFieldDefinitionInSchema(schema, parentType, arrayField)
		if arrayFieldDef == nil {
			return nil
		}

		// Determine the array element type name from the schema
		arrayElementTypeName := ""

		// Check if the array field is a List type and get its element type
		if listType, ok := arrayFieldDef.Type.(*ast.List); ok {
			if namedType, ok := listType.Type.(*ast.Named); ok {
				arrayElementTypeName = namedType.Name.Value
			}
		}

		if arrayElementTypeName == "" {
			return nil
		}

		// Find the nested field in the array element type
		return findAndExtractSourceInfo(schema, arrayElementTypeName, nestedField)

	default:
		// For paths with more than 3 parts, fall back to the first two parts
		return findAndExtractSourceInfo(schema, parts[0], parts[1])
	}
}

// findAndExtractSourceInfo is a helper function to find a field and extract its source info
func findAndExtractSourceInfo(schema *ast.Document, typeName, fieldName string) *SourceInfo {
	fieldDef := FindFieldDefinitionInSchema(schema, typeName, fieldName)
	if fieldDef == nil {
		return nil
	}
	return ExtractSourceInfoFromSchemaField(fieldDef)
}
