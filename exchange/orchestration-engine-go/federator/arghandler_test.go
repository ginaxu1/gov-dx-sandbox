package federator

import (
	"testing"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/stretchr/testify/assert"
)

func TestFindRequiredArguments(t *testing.T) {
	tests := []struct {
		name           string
		flattenedPaths *[]ProviderLevelFieldRecord
		argMappings    []*graphql.ArgMapping
		expectedCount  int
		expectedKeys   []string
		description    string
	}{
		{
			name: "Single Argument Mapping",
			flattenedPaths: &[]ProviderLevelFieldRecord{
				{
					ServiceKey: "drp",
					SchemaId:   "v1",
					FieldPath:  "person.fullName",
				},
				{
					ServiceKey: "drp",
					SchemaId:   "v1",
					FieldPath:  "person.address",
				},
			},
			argMappings: []*graphql.ArgMapping{
				{
					ProviderKey:   "drp",
					SchemaID:      "v1",
					TargetArgName: "nic",
					SourceArgPath: "personInfo-nic",
					TargetArgPath: "person",
				},
			},
			expectedCount: 1,
			expectedKeys:  []string{"drp.person"},
			description:   "Should find single required argument",
		},
		{
			name: "Multiple Argument Mappings",
			flattenedPaths: &[]ProviderLevelFieldRecord{
				{
					ServiceKey: "drp",
					SchemaId:   "drp-v1",
					FieldPath:  "person.fullName",
				},
				{
					ServiceKey: "rgd",
					SchemaId:   "rgd-v2",
					FieldPath:  "getPersonInfo.name",
				},
				{
					ServiceKey: "dmt",
					SchemaId:   "rgd-v1",
					FieldPath:  "vehicle.getVehicleInfos.data",
				},
			},
			argMappings: []*graphql.ArgMapping{
				{
					ProviderKey:   "drp",
					SchemaID:      "drp-v1",
					TargetArgName: "nic",
					SourceArgPath: "personInfo-nic",
					TargetArgPath: "person",
				},
				{
					ProviderKey:   "rgd",
					SchemaID:      "rgd-v2",
					TargetArgName: "nic",
					SourceArgPath: "personInfo-nic",
					TargetArgPath: "getPersonInfo",
				},
				{
					ProviderKey:   "dmt",
					TargetArgName: "regNos",
					SchemaID:      "rgd-v1",
					SourceArgPath: "vehicles-regNos",
					TargetArgPath: "vehicle.getVehicleInfos",
				},
			},
			expectedCount: 3,
			expectedKeys:  []string{"drp.person", "rgd.getPersonInfo", "dmt.vehicle.getVehicleInfos"},
			description:   "Should find multiple required arguments",
		},
		{
			name:           "Empty Argument Mappings",
			flattenedPaths: &[]ProviderLevelFieldRecord{},
			argMappings:    []*graphql.ArgMapping{},
			expectedCount:  0,
			expectedKeys:   []string{},
			description:    "Should handle empty argument mappings",
		},
		{
			name: "Duplicate Source Paths",
			flattenedPaths: &[]ProviderLevelFieldRecord{
				{
					ServiceKey: "drp",
					SchemaId:   "11",
					FieldPath:  "person.fullName",
				},
				{
					ServiceKey: "rgd",
					SchemaId:   "12",
					FieldPath:  "getPersonInfo.name",
				},
			},
			argMappings: []*graphql.ArgMapping{
				{
					ProviderKey:   "drp",
					SchemaID:      "11",
					TargetArgName: "nic",
					SourceArgPath: "personInfo-nic",
					TargetArgPath: "person",
				},
				{
					ProviderKey:   "rgd",
					TargetArgName: "nic",
					SchemaID:      "12",
					SourceArgPath: "personInfo-nic",
					TargetArgPath: "getPersonInfo",
				},
			},
			expectedCount: 2,
			expectedKeys:  []string{"person", "getPersonInfo"},
			description:   "Should find both argument mappings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requiredArgs := FindRequiredArguments(tt.flattenedPaths, tt.argMappings)

			assert.Len(t, requiredArgs, tt.expectedCount, tt.description)

			// Verify target paths
			actualKeys := make([]string, len(requiredArgs))
			for i, arg := range requiredArgs {
				actualKeys[i] = arg.TargetArgPath
			}
			assert.ElementsMatch(t, tt.expectedKeys, actualKeys, "Should have correct target paths")
		})
	}
}

func TestExtractRequiredArguments(t *testing.T) {
	tests := []struct {
		name          string
		argMappings   []*graphql.ArgMapping
		arguments     []*ast.Argument
		expectedCount int
		expectError   bool
		description   string
	}{
		{
			name: "Single String Argument",
			argMappings: []*graphql.ArgMapping{
				{
					ProviderKey:   "drp",
					TargetArgName: "nic",
					SourceArgPath: "personInfo-nic",
					TargetArgPath: "drp.person",
				},
			},
			arguments: []*ast.Argument{
				{
					Name:  &ast.Name{Value: "nic"},
					Value: &ast.StringValue{Value: "123456789V"},
				},
			},
			expectedCount: 1,
			expectError:   false,
			description:   "Should extract single string argument",
		},
		{
			name: "Array Argument",
			argMappings: []*graphql.ArgMapping{
				{
					ProviderKey:   "dmt",
					TargetArgName: "regNos",
					SourceArgPath: "vehicles-regNos",
					TargetArgPath: "dmt.vehicle.getVehicleInfos",
				},
			},
			arguments: []*ast.Argument{
				{
					Name: &ast.Name{Value: "regNos"},
					Value: &ast.ListValue{
						Values: []ast.Value{
							&ast.StringValue{Value: "ABC123"},
							&ast.StringValue{Value: "XYZ789"},
						},
					},
				},
			},
			expectedCount: 1,
			expectError:   false,
			description:   "Should extract array argument",
		},
		{
			name: "Multiple Arguments",
			argMappings: []*graphql.ArgMapping{
				{
					ProviderKey:   "drp",
					TargetArgName: "nic",
					SourceArgPath: "personInfo-nic",
					TargetArgPath: "drp.person",
				},
				{
					ProviderKey:   "drp",
					TargetArgName: "includeVehicles",
					SourceArgPath: "personInfo-includeVehicles",
					TargetArgPath: "drp.person",
				},
			},
			arguments: []*ast.Argument{
				{
					Name:  &ast.Name{Value: "nic"},
					Value: &ast.StringValue{Value: "123456789V"},
				},
				{
					Name:  &ast.Name{Value: "includeVehicles"},
					Value: &ast.BooleanValue{Value: true},
				},
			},
			expectedCount: 2,
			expectError:   false,
			description:   "Should extract multiple arguments",
		},
		{
			name:          "Empty Arguments",
			argMappings:   []*graphql.ArgMapping{},
			arguments:     []*ast.Argument{},
			expectedCount: 0,
			expectError:   false,
			description:   "Should handle empty arguments",
		},
		{
			name: "No Matching Arguments",
			argMappings: []*graphql.ArgMapping{
				{
					ProviderKey:   "drp",
					TargetArgName: "nic",
					SourceArgPath: "personInfo-nic",
					TargetArgPath: "drp.person",
				},
			},
			arguments: []*ast.Argument{
				{
					Name:  &ast.Name{Value: "differentArg"},
					Value: &ast.StringValue{Value: "value"},
				},
			},
			expectedCount: 0,
			expectError:   false,
			description:   "Should handle no matching arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argSources := ExtractRequiredArguments(tt.argMappings, tt.arguments)

			assert.Len(t, argSources, tt.expectedCount, "Should extract correct number of arguments")

			// Verify argument values
			for _, argSource := range argSources {
				assert.NotNil(t, argSource.Argument, "Should have valid argument")
				assert.NotNil(t, argSource.ArgMapping, "Should have valid argument mapping")
			}
		})
	}
}

//func TestPushArgumentsToProviderQueryAst(t *testing.T) {
//	tests := []struct {
//		name         string
//		queryAst     *FederationServiceAST
//		argSources   []*ArgSource
//		expectedArgs int
//		expectError  bool
//		description  string
//	}{
//		{
//			name: "Single String Argument",
//			queryAst: createMockFederationServiceAST(t, `
//				query {
//					person {
//						fullName
//					}
//				}
//			`, "drp"),
//			argSources: []*ArgSource{
//				{
//					ArgMapping: &graphql.ArgMapping{
//						ProviderKey:   "drp",
//						TargetArgName: "nic",
//						SourceArgPath: "personInfo-nic",
//						TargetArgPath: "drp.person",
//					},
//					Argument: &ast.Argument{
//						Name:  &ast.Name{Value: "nic"},
//						Value: &ast.StringValue{Value: "123456789V"},
//					},
//				},
//			},
//			expectedArgs: 1,
//			expectError:  false,
//			description:  "Should push single string argument to query AST",
//		},
//		{
//			name: "Array Argument",
//			queryAst: createMockFederationServiceAST(t, `
//				query {
//					vehicle {
//						getVehicleInfos {
//							regNo
//						}
//					}
//				}
//			`, "dmt"),
//			argSources: []*ArgSource{
//				{
//					ArgMapping: &graphql.ArgMapping{
//						ProviderKey:   "dmt",
//						TargetArgName: "regNos",
//						SourceArgPath: "vehicles-regNos",
//						TargetArgPath: "dmt.vehicle.getVehicleInfos",
//					},
//					Argument: &ast.Argument{
//						Name: &ast.Name{Value: "regNos"},
//						Value: &ast.ListValue{
//							Values: []ast.Value{
//								&ast.StringValue{Value: "ABC123"},
//								&ast.StringValue{Value: "XYZ789"},
//							},
//						},
//					},
//				},
//			},
//			expectedArgs: 1,
//			expectError:  false,
//			description:  "Should push array argument to query AST",
//		},
//		{
//			name: "Multiple Arguments",
//			queryAst: createMockFederationServiceAST(t, `
//				query {
//					person {
//						fullName
//					}
//				}
//			`, "drp"),
//			argSources: []*ArgSource{
//				{
//					ArgMapping: &graphql.ArgMapping{
//						ProviderKey:   "drp",
//						TargetArgName: "nic",
//						SourceArgPath: "personInfo-nic",
//						TargetArgPath: "drp.person",
//					},
//					Argument: &ast.Argument{
//						Name:  &ast.Name{Value: "nic"},
//						Value: &ast.StringValue{Value: "123456789V"},
//					},
//				},
//				{
//					ArgMapping: &graphql.ArgMapping{
//						ProviderKey:   "drp",
//						TargetArgName: "includeVehicles",
//						SourceArgPath: "personInfo-includeVehicles",
//						TargetArgPath: "drp.person",
//					},
//					Argument: &ast.Argument{
//						Name:  &ast.Name{Value: "includeVehicles"},
//						Value: &ast.BooleanValue{Value: true},
//					},
//				},
//			},
//			expectedArgs: 2,
//			expectError:  false,
//			description:  "Should push multiple arguments to query AST",
//		},
//		{
//			name:         "Empty Arguments",
//			queryAst:     createMockFederationServiceAST(t, `query { person { fullName } }`, "drp"),
//			argSources:   []*ArgSource{},
//			expectedArgs: 0,
//			expectError:  false,
//			description:  "Should handle empty arguments",
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			PushArgumentsToProviderQueryAst(tt.argSources, tt.queryAst)
//
//			// Verify arguments were added to the query
//			if len(tt.argSources) > 0 {
//				operationDef := tt.queryAst.QueryAst.Definitions[0].(*ast.OperationDefinition)
//				selectionSet := operationDef.SelectionSet
//
//				// Find the first field and check its arguments
//				if len(selectionSet.Selections) > 0 {
//					if field, ok := selectionSet.Selections[0].(*ast.Field); ok {
//						assert.Len(t, field.Arguments, tt.expectedArgs, "Should have correct number of arguments")
//					}
//				}
//			}
//		})
//	}
//}

//
//func TestBasicArgumentHandling(t *testing.T) {
//	t.Run("Single String Argument", func(t *testing.T) {
//		// Test that single string arguments are handled correctly
//		query := `
//			query {
//				personInfo(nic: "123456789V") {
//					fullName
//					name
//				}
//			}
//		`
//
//		queryDoc := tests.ParseTestQuery(t, query)
//		operationDef := queryDoc.Definitions[0].(*ast.OperationDefinition)
//
//		argMappings := []*graphql.ArgMapping{
//			{
//				ProviderKey:   "drp",
//				TargetArgName: "nic",
//				SourceArgPath: "personInfo-nic",
//				TargetArgPath: "drp.person",
//			},
//		}
//
//		// Extract arguments from the query
//		var arguments []*ast.Argument
//		for _, selection := range operationDef.SelectionSet.Selections {
//			if field, ok := selection.(*ast.Field); ok {
//				arguments = append(arguments, field.Arguments...)
//			}
//		}
//
//		argSources := ExtractRequiredArguments(argMappings, arguments)
//		assert.Len(t, argSources, 1, "Should extract one argument source")
//
//		// Verify the argument is a string value
//		argSource := argSources[0]
//		stringValue, ok := argSource.Argument.Value.(*ast.StringValue)
//		assert.True(t, ok, "Should have string value")
//		assert.Equal(t, "123456789V", stringValue.Value)
//	})
//
//	t.Run("Multiple String Arguments", func(t *testing.T) {
//		// Test that multiple string arguments are handled correctly
//		query := `
//			query {
//				personInfo(nic: "123456789V") {
//					fullName
//				}
//				vehicle(regNo: "ABC123") {
//					make
//				}
//			}
//		`
//
//		queryDoc := tests.ParseTestQuery(t, query)
//		operationDef := queryDoc.Definitions[0].(*ast.OperationDefinition)
//
//		argMappings := []*graphql.ArgMapping{
//			{
//				ProviderKey:   "drp",
//				TargetArgName: "nic",
//				SourceArgPath: "personInfo-nic",
//				TargetArgPath: "drp.person",
//			},
//			{
//				ProviderKey:   "dmt",
//				TargetArgName: "regNo",
//				SourceArgPath: "vehicle-regNo",
//				TargetArgPath: "dmt.vehicle",
//			},
//		}
//
//		// Extract arguments from the query
//		var arguments []*ast.Argument
//		for _, selection := range operationDef.SelectionSet.Selections {
//			if field, ok := selection.(*ast.Field); ok {
//				arguments = append(arguments, field.Arguments...)
//			}
//		}
//
//		argSources := ExtractRequiredArguments(argMappings, arguments)
//		assert.Len(t, argSources, 2, "Should extract two argument sources")
//
//		// Verify both arguments are string values
//		for _, argSource := range argSources {
//			stringValue, ok := argSource.Argument.Value.(*ast.StringValue)
//			assert.True(t, ok, "Should have string value")
//			assert.True(t, stringValue.Value == "123456789V" || stringValue.Value == "ABC123", "Should have correct value")
//		}
//	})
//}

// Helper functions
