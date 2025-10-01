package tests

import (
	"testing"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/federator"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/stretchr/testify/assert"
)

func TestProviderSchemaCollector(t *testing.T) {
	t.Skip("Skipping mapper test - requires config initialization")
	schema := CreateTestSchema(t)

	tests := []struct {
		name           string
		query          string
		expectedFields []string
		expectedArgs   int
		expectError    bool
		description    string
	}{
		{
			name: "Single Entity Query",
			query: `
				query {
					personInfo(nic: "123456789V") {
						fullName
						name
						address
					}
				}
			`,
			expectedFields: []string{
				"drp.person.fullName",
				"rgd.getPersonInfo.name",
				"drp.person.permanentAddress",
			},
			expectedArgs: 1,
			expectError:  false,
			description:  "Should extract source info directives for single entity query",
		},
		{
			name: "Query with Array Field",
			query: `
				query {
					personInfo(nic: "123456789V") {
						fullName
						ownedVehicles {
							regNo
							make
							model
						}
					}
				}
			`,
			expectedFields: []string{
				"drp.person.fullName",
				"dmt.vehicle.getVehicleInfos.data",
				"dmt.vehicle.getVehicleInfos.data.registrationNumber",
				"dmt.vehicle.getVehicleInfos.data.make",
				"dmt.vehicle.getVehicleInfos.data.model",
			},
			expectedArgs: 1,
			expectError:  false,
			description:  "Should extract source info directives for query with array field",
		},
		{
			name: "Bulk Query",
			query: `
				query {
					personInfos(nics: ["123456789V", "987654321V"]) {
						fullName
						name
					}
				}
			`,
			expectedFields: []string{
				"drp.person.fullName",
				"rgd.getPersonInfo.name",
			},
			expectedArgs: 1,
			expectError:  false,
			description:  "Should extract source info directives for bulk query",
		},
		{
			name: "Query with Variables",
			query: `
				query GetPersonInfo($nic: String!) {
					personInfo(nic: $nic) {
						fullName
						address
					}
				}
			`,
			expectedFields: []string{
				"drp.person.fullName",
				"drp.person.permanentAddress",
			},
			expectedArgs: 1,
			expectError:  false,
			description:  "Should extract source info directives for query with variables",
		},
		{
			name: "Invalid Query - Mutation",
			query: `
				mutation {
					updatePerson(nic: "123456789V") {
						fullName
					}
				}
			`,
			expectedFields: nil,
			expectedArgs:   0,
			expectError:    true,
			description:    "Should reject mutation queries",
		},
		{
			name: "Invalid Query - Multiple Operations",
			query: `
				query Query1 {
					personInfo(nic: "123456789V") {
						fullName
					}
				}
				query Query2 {
					personInfo(nic: "987654321V") {
						name
					}
				}
			`,
			expectedFields: nil,
			expectedArgs:   0,
			expectError:    true,
			description:    "Should reject queries with multiple operations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryDoc := ParseTestQuery(t, tt.query)

			fields, args, err := federator.ProviderSchemaCollector(schema, queryDoc)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, fields)
				assert.Nil(t, args)
			} else {
				assert.NoError(t, err, tt.description)
				assert.Len(t, fields, len(tt.expectedFields), "Should extract correct number of fields")
				assert.Len(t, args, tt.expectedArgs, "Should extract correct number of arguments")

				// Verify extracted fields
				for _, expectedField := range tt.expectedFields {
					assert.Contains(t, fields, expectedField, "Should contain field: %s", expectedField)
				}
			}
		})
	}
}

func TestQueryBuilder(t *testing.T) {
	tests := []struct {
		name          string
		fieldsMap     []string
		args          []*federator.ArgSource
		expectedCount int
		expectedKeys  []string
		expectError   bool
		description   string
	}{
		{
			name:      "Single Provider Query",
			fieldsMap: []string{"drp.person.fullName", "drp.person.address"},
			args: []*federator.ArgSource{
				{
					ArgMapping: &graphql.ArgMapping{
						ProviderKey:   "drp",
						TargetArgName: "nic",
						SourceArgPath: "personInfo-nic",
						TargetArgPath: "drp.person",
					},
					Argument: &ast.Argument{
						Name:  &ast.Name{Value: "nic"},
						Value: &ast.StringValue{Value: "123456789V"},
					},
				},
			},
			expectedCount: 1,
			expectedKeys:  []string{"drp"},
			expectError:   false,
			description:   "Should build single provider query with arguments",
		},
		{
			name:      "Multiple Provider Queries",
			fieldsMap: []string{"drp.person.fullName", "rgd.getPersonInfo.name", "dmt.vehicle.data"},
			args: []*federator.ArgSource{
				{
					ArgMapping: &graphql.ArgMapping{
						ProviderKey:   "drp",
						TargetArgName: "nic",
						SourceArgPath: "personInfo-nic",
						TargetArgPath: "drp.person",
					},
					Argument: &ast.Argument{
						Name:  &ast.Name{Value: "nic"},
						Value: &ast.StringValue{Value: "123456789V"},
					},
				},
				{
					ArgMapping: &graphql.ArgMapping{
						ProviderKey:   "rgd",
						TargetArgName: "nic",
						SourceArgPath: "personInfo-nic",
						TargetArgPath: "rgd.getPersonInfo",
					},
					Argument: &ast.Argument{
						Name:  &ast.Name{Value: "nic"},
						Value: &ast.StringValue{Value: "123456789V"},
					},
				},
			},
			expectedCount: 3,
			expectedKeys:  []string{"drp", "rgd", "dmt"},
			expectError:   false,
			description:   "Should build multiple provider queries",
		},
		{
			name:          "Empty Fields Map",
			fieldsMap:     []string{},
			args:          []*federator.ArgSource{},
			expectedCount: 0,
			expectedKeys:  []string{},
			expectError:   false,
			description:   "Should handle empty fields map",
		},
		{
			name:      "Query with Array Arguments",
			fieldsMap: []string{"dmt.vehicle.data"},
			args: []*federator.ArgSource{
				{
					ArgMapping: &graphql.ArgMapping{
						ProviderKey:   "dmt",
						TargetArgName: "regNos",
						SourceArgPath: "vehicles-regNos",
						TargetArgPath: "dmt.vehicle.getVehicleInfos",
					},
					Argument: &ast.Argument{
						Name: &ast.Name{Value: "regNos"},
						Value: &ast.ListValue{
							Values: []ast.Value{
								&ast.StringValue{Value: "ABC123"},
								&ast.StringValue{Value: "XYZ789"},
							},
						},
					},
				},
			},
			expectedCount: 1,
			expectedKeys:  []string{"dmt"},
			expectError:   false,
			description:   "Should handle array arguments for bulk queries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requests, err := federator.QueryBuilder(tt.fieldsMap, tt.args)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				assert.Len(t, requests, tt.expectedCount, "Should create correct number of requests")

				// Verify service keys
				actualKeys := make([]string, len(requests))
				for i, request := range requests {
					actualKeys[i] = request.ServiceKey
				}
				assert.ElementsMatch(t, tt.expectedKeys, actualKeys, "Should have correct service keys")

				// Verify each request has a valid GraphQL query
				for _, request := range requests {
					assert.NotEmpty(t, request.GraphQLRequest.Query, "Should have non-empty query")
					assert.NotNil(t, request.GraphQLRequest, "Should have valid GraphQL request")
				}
			}
		})
	}
}

func TestRecursivelyExtractSourceSchemaInfo(t *testing.T) {
	t.Skip("Skipping mapper test - requires config initialization")
	schema := CreateTestSchema(t)

	tests := []struct {
		name           string
		query          string
		expectedFields []string
		expectedArgs   int
		description    string
	}{
		{
			name: "Nested Object Query",
			query: `
				query {
					personInfo(nic: "123456789V") {
						fullName
						birthInfo {
							birthRegistrationNumber
							birthPlace
						}
					}
				}
			`,
			expectedFields: []string{
				"drp.person.fullName",
				"rgd.getPersonInfo.brNo",
				"rgd.getPersonInfo.birthPlace",
			},
			expectedArgs: 1,
			description:  "Should extract source info from nested objects",
		},
		{
			name: "Array Field Query",
			query: `
				query {
					personInfo(nic: "123456789V") {
						fullName
						ownedVehicles {
							regNo
							make
							model
						}
					}
				}
			`,
			expectedFields: []string{
				"drp.person.fullName",
				"dmt.vehicle.getVehicleInfos.data",
				"dmt.vehicle.getVehicleInfos.data.registrationNumber",
				"dmt.vehicle.getVehicleInfos.data.make",
				"dmt.vehicle.getVehicleInfos.data.model",
			},
			expectedArgs: 1,
			description:  "Should extract source info from array fields",
		},
		{
			name: "Complex Nested Query",
			query: `
				query {
					personInfo(nic: "123456789V") {
						fullName
						birthInfo {
							birthRegistrationNumber
							district
						}
						ownedVehicles {
							regNo
							make
						}
					}
				}
			`,
			expectedFields: []string{
				"drp.person.fullName",
				"rgd.getPersonInfo.brNo",
				"rgd.getPersonInfo.district",
				"dmt.vehicle.getVehicleInfos.data",
			},
			expectedArgs: 1,
			description:  "Should extract source info from complex nested structure (array fields handled by new implementation)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryDoc := ParseTestQuery(t, tt.query)
			operationDef := queryDoc.Definitions[0].(*ast.OperationDefinition)
			selectionSet := operationDef.SelectionSet

			// Get query object definition from schema
			queryObjectDef := federator.GetQueryObjectDefinition(schema)
			assert.NotNil(t, queryObjectDef, "Should find query object definition")

			// Extract source schema info
			directives, arguments := federator.RecursivelyExtractSourceSchemaInfo(
				selectionSet, schema, queryObjectDef, nil, nil)

			// Convert directives to field map
			fieldMap := federator.ProviderFieldMap(directives)

			assert.Len(t, fieldMap, len(tt.expectedFields), "Should extract correct number of fields")
			assert.Len(t, arguments, tt.expectedArgs, "Should extract correct number of arguments")

			// Verify extracted fields
			for _, expectedField := range tt.expectedFields {
				assert.Contains(t, fieldMap, expectedField, "Should contain field: %s", expectedField)
			}
		})
	}
}

func TestFindFieldDefinitionFromFieldName(t *testing.T) {
	schema := CreateTestSchema(t)

	tests := []struct {
		name           string
		fieldName      string
		parentObject   string
		expectedExists bool
		description    string
	}{
		{
			name:           "Existing Field in PersonInfo",
			fieldName:      "fullName",
			parentObject:   "PersonInfo",
			expectedExists: true,
			description:    "Should find existing field in PersonInfo",
		},
		{
			name:           "Existing Field in VehicleInfo",
			fieldName:      "regNo",
			parentObject:   "VehicleInfo",
			expectedExists: true,
			description:    "Should find existing field in VehicleInfo",
		},
		{
			name:           "Non-existent Field",
			fieldName:      "nonExistentField",
			parentObject:   "PersonInfo",
			expectedExists: false,
			description:    "Should return nil for non-existent field",
		},
		{
			name:           "Non-existent Parent Object",
			fieldName:      "fullName",
			parentObject:   "NonExistentType",
			expectedExists: false,
			description:    "Should return nil for non-existent parent object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fieldDef := federator.FindFieldDefinitionFromFieldName(tt.fieldName, schema, tt.parentObject)

			if tt.expectedExists {
				assert.NotNil(t, fieldDef, tt.description)
				assert.Equal(t, tt.fieldName, fieldDef.Name.Value, "Should have correct field name")
			} else {
				assert.Nil(t, fieldDef, tt.description)
			}
		})
	}
}

func TestGetQueryObjectDefinition(t *testing.T) {
	schema := CreateTestSchema(t)

	queryDef := federator.GetQueryObjectDefinition(schema)
	assert.NotNil(t, queryDef, "Should find query object definition")
	assert.Equal(t, "Query", queryDef.Name.Value, "Should have correct name")
	assert.Greater(t, len(queryDef.Fields), 0, "Should have fields")
}

// Helper functions
