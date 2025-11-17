package federator

import (
	"testing"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	logger.Init()
}

func TestIsArrayFieldInSchema(t *testing.T) {
	schemaSDL := `
		type Query {
			personInfo: PersonInfo
		}
		type PersonInfo {
			fullName: String
			ownedVehicles: [VehicleInfo]
		}
		type VehicleInfo {
			regNo: String
		}
	`

	schema := parseSchemaHelper(t, schemaSDL)

	tests := []struct {
		name       string
		parentType string
		fieldName  string
		expected   bool
	}{
		{
			name:       "Array field",
			parentType: "PersonInfo",
			fieldName:  "ownedVehicles",
			expected:   true,
		},
		{
			name:       "Non-array field",
			parentType: "PersonInfo",
			fieldName:  "fullName",
			expected:   false,
		},
		{
			name:       "Non-existent type",
			parentType: "NonExistent",
			fieldName:  "test",
			expected:   false,
		},
		{
			name:       "Non-existent field",
			parentType: "PersonInfo",
			fieldName:  "nonExistent",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isArrayFieldInSchema(schema, tt.parentType, tt.fieldName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsArrayFieldInSchema_NilSchema(t *testing.T) {
	result := isArrayFieldInSchema(nil, "PersonInfo", "ownedVehicles")
	assert.False(t, result)
}

func TestGetArrayElementTypeNameFromSchema(t *testing.T) {
	schemaSDL := `
		type Query {
			personInfo: PersonInfo
		}
		type PersonInfo {
			ownedVehicles: [VehicleInfo]
		}
		type VehicleInfo {
			regNo: String
		}
	`

	schema := parseSchemaHelper(t, schemaSDL)

	tests := []struct {
		name           string
		parentType     string
		arrayFieldName string
		expected       string
	}{
		{
			name:           "Get element type from array field",
			parentType:     "PersonInfo",
			arrayFieldName: "ownedVehicles",
			expected:       "VehicleInfo",
		},
		{
			name:           "Non-array field",
			parentType:     "PersonInfo",
			arrayFieldName: "fullName",
			expected:       "",
		},
		{
			name:           "Non-existent type",
			parentType:     "NonExistent",
			arrayFieldName: "test",
			expected:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getArrayElementTypeNameFromSchema(schema, tt.parentType, tt.arrayFieldName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetArrayElementTypeNameFromSchema_NilSchema(t *testing.T) {
	result := getArrayElementTypeNameFromSchema(nil, "PersonInfo", "ownedVehicles")
	assert.Empty(t, result)
}

func TestFindFieldInSelectionSet(t *testing.T) {
	query := `
		query {
			personInfo {
				fullName
				ownedVehicles {
					regNo
					make
				}
			}
		}
	`

	queryDoc := parseQueryHelper(t, query)

	// Get the personInfo field's selection set
	operationDef := queryDoc.Definitions[0].(*ast.OperationDefinition)
	personInfoField := operationDef.SelectionSet.Selections[0].(*ast.Field)
	selectionSet := personInfoField.SelectionSet

	tests := []struct {
		name      string
		fieldName string
		expected  bool
	}{
		{
			name:      "Find existing field",
			fieldName: "fullName",
			expected:  true,
		},
		{
			name:      "Find nested field",
			fieldName: "regNo",
			expected:  true,
		},
		{
			name:      "Find non-existent field",
			fieldName: "nonExistent",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := findFieldInSelectionSet(selectionSet, tt.fieldName)
			if tt.expected {
				assert.NotNil(t, field)
				assert.Equal(t, tt.fieldName, field.Name.Value)
			} else {
				assert.Nil(t, field)
			}
		})
	}
}

func TestFindFieldInSelectionSet_NilSelectionSet(t *testing.T) {
	field := findFieldInSelectionSet(nil, "test")
	assert.Nil(t, field)
}

func TestProcessArrayFieldSimple_EmptyArray(t *testing.T) {
	responseData := make(map[string]interface{})
	path := []string{"personInfo", "ownedVehicles"}
	fieldName := "ownedVehicles"
	sourceArray := []interface{}{}

	query := `query { personInfo { ownedVehicles { regNo } } }`
	queryDoc := parseQueryHelper(t, query)
	operationDef := queryDoc.Definitions[0].(*ast.OperationDefinition)
	personInfoField := operationDef.SelectionSet.Selections[0].(*ast.Field)
	ownedVehiclesField := personInfoField.SelectionSet.Selections[0].(*ast.Field)
	selectionSet := ownedVehiclesField.SelectionSet

	federatedResponse := &FederationResponse{
		Responses: []ProviderResponse{},
	}

	processArrayFieldSimple(responseData, path, fieldName, sourceArray, selectionSet, federatedResponse, nil)

	// Should handle empty array gracefully
	personInfoVal, ok := responseData["personInfo"]
	if !ok {
		t.Fatalf("personInfo not found in response data: %#v", responseData)
	}
	personInfo, ok := personInfoVal.(map[string]interface{})
	if !ok {
		t.Fatalf("personInfo is not a map: %#v", personInfoVal)
	}
	vehiclesVal, ok := personInfo["ownedVehicles"]
	if !ok {
		t.Fatalf("ownedVehicles not found in personInfo: %#v", personInfo)
	}
	switch v := vehiclesVal.(type) {
	case []map[string]interface{}:
		assert.Len(t, v, 0)
	case []interface{}:
		assert.Len(t, v, 0)
	default:
		t.Fatalf("ownedVehicles has unexpected type: %T", vehiclesVal)
	}
}

func TestProcessArrayFieldSimple_NonArrayValue(t *testing.T) {
	responseData := make(map[string]interface{})
	path := []string{"personInfo", "ownedVehicles"}
	fieldName := "ownedVehicles"
	sourceArray := "not an array" // Wrong type

	query := `query { personInfo { ownedVehicles { regNo } } }`
	queryDoc := parseQueryHelper(t, query)
	operationDef := queryDoc.Definitions[0].(*ast.OperationDefinition)
	personInfoField := operationDef.SelectionSet.Selections[0].(*ast.Field)
	ownedVehiclesField := personInfoField.SelectionSet.Selections[0].(*ast.Field)
	selectionSet := ownedVehiclesField.SelectionSet

	federatedResponse := &FederationResponse{
		Responses: []ProviderResponse{},
	}

	processArrayFieldSimple(responseData, path, fieldName, sourceArray, selectionSet, federatedResponse, nil)

	// Should handle non-array gracefully (logs warning, doesn't crash)
	assert.NotNil(t, responseData)
}

func TestProcessSimpleField_Success(t *testing.T) {
	responseData := map[string]interface{}{
		"personInfo": map[string]interface{}{},
	}
	path := []string{"personInfo"}
	fieldName := "fullName"

	schemaInfo := &SourceSchemaInfo{
		ProviderKey:   "drp",
		ProviderField: "person.fullName",
	}

	federatedResponse := &FederationResponse{
		Responses: []ProviderResponse{
			{
				ServiceKey: "drp",
				Response: graphql.Response{
					Data: map[string]interface{}{
						"person": map[string]interface{}{
							"fullName": "John Doe",
						},
					},
				},
			},
		},
	}

	processSimpleField(responseData, path, fieldName, schemaInfo, federatedResponse)

	personInfo := responseData["personInfo"].(map[string]interface{})
	assert.Equal(t, "John Doe", personInfo["fullName"])
}

func TestProcessArrayFieldWithSchema_Success(t *testing.T) {
	responseData := map[string]interface{}{}
	path := []string{"personInfo"}
	fieldName := "ownedVehicles"

	schemaInfo := &SourceSchemaInfo{
		ProviderKey:            "dmt",
		IsArray:                true,
		ProviderArrayFieldPath: "vehicle.getVehicleInfos.data",
		SubFieldSchemaInfos: map[string]*SourceSchemaInfo{
			"regNo": {
				ProviderKey:   "dmt",
				ProviderField: "registrationNumber",
			},
			"make": {
				ProviderKey:   "dmt",
				ProviderField: "make",
			},
		},
	}

	federatedResponse := &FederationResponse{
		Responses: []ProviderResponse{
			{
				ServiceKey: "dmt",
				Response: graphql.Response{
					Data: map[string]interface{}{
						"vehicle": map[string]interface{}{
							"getVehicleInfos": map[string]interface{}{
								"data": []interface{}{
									map[string]interface{}{
										"registrationNumber": "ABC123",
										"make":               "Toyota",
									},
									map[string]interface{}{
										"registrationNumber": "XYZ789",
										"make":               "Honda",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	processArrayFieldWithSchema(responseData, path, fieldName, schemaInfo, federatedResponse)

	personInfoVal, ok := responseData["personInfo"]
	require.True(t, ok)

	personInfo, ok := personInfoVal.(map[string]interface{})
	require.True(t, ok)

	ownedVehiclesVal, ok := personInfo["ownedVehicles"]
	require.True(t, ok)

	vehicles, ok := ownedVehiclesVal.([]map[string]interface{})
	if !ok {
		converted := make([]map[string]interface{}, 0, len(ownedVehiclesVal.([]interface{})))
		for _, v := range ownedVehiclesVal.([]interface{}) {
			converted = append(converted, v.(map[string]interface{}))
		}
		vehicles = converted
	}

	require.Len(t, vehicles, 2)
	assert.Equal(t, "ABC123", vehicles[0]["regNo"])
	assert.Equal(t, "Toyota", vehicles[0]["make"])
}

// Helper functions
func parseSchemaHelper(t *testing.T, schemaSDL string) *ast.Document {
	src := source.NewSource(&source.Source{
		Body: []byte(schemaSDL),
		Name: "TestSchema",
	})

	schema, err := parser.Parse(parser.ParseParams{Source: src})
	require.NoError(t, err, "Should parse schema successfully")
	return schema
}

func parseQueryHelper(t *testing.T, query string) *ast.Document {
	src := source.NewSource(&source.Source{
		Body: []byte(query),
		Name: "TestQuery",
	})

	doc, err := parser.Parse(parser.ParseParams{Source: src})
	require.NoError(t, err, "Should parse query successfully")
	return doc
}
