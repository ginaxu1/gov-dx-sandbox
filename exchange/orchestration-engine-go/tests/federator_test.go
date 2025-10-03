package tests

import (
	"testing"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/federator"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryParsing(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		expectError bool
		description string
	}{
		{
			name: "Valid Single Entity Query",
			query: `
				query {
					personInfo(nic: "123456789V") {
						fullName
						name
						address
					}
				}
			`,
			expectError: false,
			description: "Should parse a valid single entity query successfully",
		},
		{
			name: "Valid Array Field Query",
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
			expectError: false,
			description: "Should parse a query with array fields successfully",
		},
		{
			name: "Valid Bulk Query (Future Implementation)",
			query: `
				query {
					personInfos(nics: ["123456789V", "987654321V"]) {
						fullName
						name
						address
					}
				}
			`,
			expectError: false,
			description: "Should parse a bulk query for multiple entities",
		},
		{
			name: "Invalid GraphQL Syntax",
			query: `
				query {
					personInfo(nic: "123456789V" {
						fullName
					}
				}
			`,
			expectError: true,
			description: "Should fail to parse invalid GraphQL syntax",
		},
		{
			name:        "Empty Query",
			query:       "",
			expectError: false,
			description: "Should parse empty query as valid document",
		},
		{
			name: "Query with Variables",
			query: `
				query GetPersonInfo($nic: String!) {
					personInfo(nic: $nic) {
						fullName
						name
					}
				}
			`,
			expectError: false,
			description: "Should parse query with variables successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert query string to AST
			src := source.NewSource(&source.Source{
				Body: []byte(tt.query),
				Name: "TestQuery",
			})

			doc, err := parser.Parse(parser.ParseParams{Source: src})

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, doc)
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, doc)
				assert.Equal(t, "Document", doc.Kind)
				// For empty queries, definitions can be 0
				if tt.name != "Empty Query" {
					assert.Greater(t, len(doc.Definitions), 0)
				}
			}
		})
	}
}

func TestSchemaCollection(t *testing.T) {
	t.Skip("Skipping schema collection test - requires config initialization")
	// Create a mock schema with @sourceInfo directives
	schemaSDL := `
		directive @sourceInfo(
			providerKey: String!
			providerField: String!
		) on FIELD_DEFINITION

		type Query {
			personInfo(nic: String!): PersonInfo
			personInfos(nics: [String!]!): [PersonInfo]
		}

		type PersonInfo {
			fullName: String @sourceInfo(providerKey: "drp", providerField: "person.fullName")
			name: String @sourceInfo(providerKey: "rgd", providerField: "getPersonInfo.name")
			address: String @sourceInfo(providerKey: "drp", providerField: "person.permanentAddress")
			ownedVehicles: [VehicleInfo] @sourceInfo(providerKey: "dmt", providerField: "vehicle.getVehicleInfos.data")
		}

		type VehicleInfo {
			regNo: String @sourceInfo(providerKey: "dmt", providerField: "vehicle.getVehicleInfos.data.registrationNumber")
			make: String @sourceInfo(providerKey: "dmt", providerField: "vehicle.getVehicleInfos.data.make")
			model: String @sourceInfo(providerKey: "dmt", providerField: "vehicle.getVehicleInfos.data.model")
		}
	`

	// Parse the schema
	src := source.NewSource(&source.Source{
		Body: []byte(schemaSDL),
		Name: "TestSchema",
	})

	schema, err := parser.Parse(parser.ParseParams{Source: src})
	require.NoError(t, err, "Should parse schema successfully")

	tests := []struct {
		name           string
		query          string
		expectedFields []string
		expectedArgs   int
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
			description:  "Should extract source info directives for query with array field",
		},
		{
			name: "Bulk Query (Future Implementation)",
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
			description:  "Should extract source info directives for query with variables",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the query
			querySrc := source.NewSource(&source.Source{
				Body: []byte(tt.query),
				Name: "TestQuery",
			})

			queryDoc, err := parser.Parse(parser.ParseParams{Source: querySrc})
			require.NoError(t, err, "Should parse query successfully")

			// Extract source info directives
			fields, args, err := federator.ProviderSchemaCollector(schema, queryDoc)

			assert.NoError(t, err, tt.description)
			assert.Len(t, fields, len(tt.expectedFields), "Should extract correct number of fields")
			assert.Len(t, args, tt.expectedArgs, "Should extract correct number of arguments")

			// Verify extracted fields
			for _, expectedField := range tt.expectedFields {
				assert.Contains(t, fields, expectedField, "Should contain field: %s", expectedField)
			}
		})
	}
}

func TestProviderFieldMap(t *testing.T) {
	tests := []struct {
		name           string
		directives     []*ast.Directive
		expectedFields []string
		description    string
	}{
		{
			name: "Single Source Info Directive",
			directives: []*ast.Directive{
				{
					Name: &ast.Name{Value: "sourceInfo"},
					Arguments: []*ast.Argument{
						{
							Name:  &ast.Name{Value: "providerKey"},
							Value: &ast.StringValue{Value: "drp"},
						},
						{
							Name:  &ast.Name{Value: "providerField"},
							Value: &ast.StringValue{Value: "person.fullName"},
						},
					},
				},
			},
			expectedFields: []string{"drp.person.fullName"},
			description:    "Should map single source info directive correctly",
		},
		{
			name: "Multiple Source Info Directives",
			directives: []*ast.Directive{
				{
					Name: &ast.Name{Value: "sourceInfo"},
					Arguments: []*ast.Argument{
						{
							Name:  &ast.Name{Value: "providerKey"},
							Value: &ast.StringValue{Value: "drp"},
						},
						{
							Name:  &ast.Name{Value: "providerField"},
							Value: &ast.StringValue{Value: "person.fullName"},
						},
					},
				},
				{
					Name: &ast.Name{Value: "sourceInfo"},
					Arguments: []*ast.Argument{
						{
							Name:  &ast.Name{Value: "providerKey"},
							Value: &ast.StringValue{Value: "rgd"},
						},
						{
							Name:  &ast.Name{Value: "providerField"},
							Value: &ast.StringValue{Value: "getPersonInfo.name"},
						},
					},
				},
			},
			expectedFields: []string{"drp.person.fullName", "rgd.getPersonInfo.name"},
			description:    "Should map multiple source info directives correctly",
		},
		{
			name:           "Empty Directives",
			directives:     []*ast.Directive{},
			expectedFields: []string{},
			description:    "Should handle empty directives list",
		},
		{
			name: "Non Source Info Directives",
			directives: []*ast.Directive{
				{
					Name: &ast.Name{Value: "deprecated"},
					Arguments: []*ast.Argument{
						{
							Name:  &ast.Name{Value: "reason"},
							Value: &ast.StringValue{Value: "No longer supported"},
						},
					},
				},
			},
			expectedFields: []string{},
			description:    "Should ignore non-source info directives",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := federator.ProviderFieldMap(tt.directives)
			assert.Equal(t, tt.expectedFields, result, tt.description)
		})
	}
}

func TestBuildProviderLevelQuery(t *testing.T) {
	tests := []struct {
		name           string
		fieldsMap      []string
		expectedKeys   []string
		expectedFields map[string][]string
		description    string
	}{
		{
			name:         "Single Provider Fields",
			fieldsMap:    []string{"drp.person.fullName", "drp.person.address"},
			expectedKeys: []string{"drp"},
			expectedFields: map[string][]string{
				"drp": {"person"},
			},
			description: "Should group fields by single provider",
		},
		{
			name:         "Multiple Provider Fields",
			fieldsMap:    []string{"drp.person.fullName", "rgd.getPersonInfo.name", "dmt.vehicle.data"},
			expectedKeys: []string{"drp", "rgd", "dmt"},
			expectedFields: map[string][]string{
				"drp": {"person"},
				"rgd": {"getPersonInfo"},
				"dmt": {"vehicle"},
			},
			description: "Should group fields by multiple providers",
		},
		{
			name:           "Empty Fields Map",
			fieldsMap:      []string{},
			expectedKeys:   []string{},
			expectedFields: map[string][]string{},
			description:    "Should handle empty fields map",
		},
		{
			name:         "Nested Fields",
			fieldsMap:    []string{"drp.person.fullName", "drp.person.address.street", "drp.person.address.city"},
			expectedKeys: []string{"drp"},
			expectedFields: map[string][]string{
				"drp": {"person"},
			},
			description: "Should handle nested field paths",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queries := federator.BuildProviderLevelQuery(tt.fieldsMap)

			assert.Len(t, queries, len(tt.expectedKeys), "Should create correct number of provider queries")

			// Verify service keys
			actualKeys := make([]string, len(queries))
			for i, query := range queries {
				actualKeys[i] = query.ServiceKey
			}
			assert.ElementsMatch(t, tt.expectedKeys, actualKeys, "Should have correct service keys")

			// Verify field structure in each query
			for _, query := range queries {
				expectedFields, exists := tt.expectedFields[query.ServiceKey]
				assert.True(t, exists, "Should have expected fields for provider: %s", query.ServiceKey)

				// Extract field names from the query AST
				operationDef := query.QueryAst.Definitions[0].(*ast.OperationDefinition)
				actualFields := extractFieldNames(operationDef.SelectionSet)
				assert.ElementsMatch(t, expectedFields, actualFields, "Should have correct fields for provider: %s", query.ServiceKey)
			}
		})
	}
}

// Helper function to extract field names from selection set
func extractFieldNames(selectionSet *ast.SelectionSet) []string {
	var fields []string
	for _, selection := range selectionSet.Selections {
		if field, ok := selection.(*ast.Field); ok {
			fields = append(fields, field.Name.Value)
		}
	}
	return fields
}
