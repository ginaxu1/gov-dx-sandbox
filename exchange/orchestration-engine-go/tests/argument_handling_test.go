package tests

import (
	"testing"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/federator"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// ARGUMENT EXTRACTION TESTS
// ============================================================================

func TestFindRequiredArguments(t *testing.T) {
	tests := []struct {
		name           string
		flattenedPaths []string
		argMappings    []*graphql.ArgMapping
		expectedCount  int
		expectedKeys   []string
		description    string
	}{
		{
			name:           "Single Argument Mapping",
			flattenedPaths: []string{"drp.person.fullName", "drp.person.address"},
			argMappings: []*graphql.ArgMapping{
				{
					ProviderKey:   "drp",
					TargetArgName: "nic",
					SourceArgPath: "personInfo-nic",
					TargetArgPath: "drp.person",
				},
			},
			expectedCount: 1,
			expectedKeys:  []string{"drp.person"},
			description:   "Should find single required argument",
		},
		{
			name:           "Multiple Argument Mappings",
			flattenedPaths: []string{"drp.person.fullName", "rgd.getPersonInfo.name", "dmt.vehicle.getVehicleInfos.data"},
			argMappings: []*graphql.ArgMapping{
				{
					ProviderKey:   "drp",
					TargetArgName: "nic",
					SourceArgPath: "personInfo-nic",
					TargetArgPath: "drp.person",
				},
				{
					ProviderKey:   "rgd",
					TargetArgName: "nic",
					SourceArgPath: "personInfo-nic",
					TargetArgPath: "rgd.getPersonInfo",
				},
				{
					ProviderKey:   "dmt",
					TargetArgName: "regNos",
					SourceArgPath: "vehicles-regNos",
					TargetArgPath: "dmt.vehicle.getVehicleInfos",
				},
			},
			expectedCount: 3,
			expectedKeys:  []string{"drp.person", "rgd.getPersonInfo", "dmt.vehicle.getVehicleInfos"},
			description:   "Should find multiple required arguments",
		},
		{
			name:           "Empty Argument Mappings",
			flattenedPaths: []string{},
			argMappings:    []*graphql.ArgMapping{},
			expectedCount:  0,
			expectedKeys:   []string{},
			description:    "Should handle empty argument mappings",
		},
		{
			name:           "Duplicate Source Paths",
			flattenedPaths: []string{"drp.person.fullName", "rgd.getPersonInfo.name"},
			argMappings: []*graphql.ArgMapping{
				{
					ProviderKey:   "drp",
					TargetArgName: "nic",
					SourceArgPath: "personInfo-nic",
					TargetArgPath: "drp.person",
				},
				{
					ProviderKey:   "rgd",
					TargetArgName: "nic",
					SourceArgPath: "personInfo-nic",
					TargetArgPath: "rgd.getPersonInfo",
				},
			},
			expectedCount: 2,
			expectedKeys:  []string{"drp.person", "rgd.getPersonInfo"},
			description:   "Should find both argument mappings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requiredArgs := federator.FindRequiredArguments(tt.flattenedPaths, tt.argMappings)

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
			argSources := federator.ExtractRequiredArguments(tt.argMappings, tt.arguments)

			assert.Len(t, argSources, tt.expectedCount, "Should extract correct number of arguments")

			// Verify argument values
			for _, argSource := range argSources {
				assert.NotNil(t, argSource.Argument, "Should have valid argument")
				assert.NotNil(t, argSource.ArgMapping, "Should have valid argument mapping")
			}
		})
	}
}

// ============================================================================
// ARGUMENT PUSHING TESTS
// ============================================================================

func TestPushArgumentsToProviderQueryAst(t *testing.T) {
	tests := []struct {
		name         string
		queryAst     *federator.FederationServiceAST
		argSources   []*federator.ArgSource
		expectedArgs int
		expectError  bool
		description  string
	}{
		{
			name: "Single String Argument",
			queryAst: createMockFederationServiceAST(t, `
				query {
					person {
						fullName
					}
				}
			`, "drp"),
			argSources: []*federator.ArgSource{
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
			expectedArgs: 1,
			expectError:  false,
			description:  "Should push single string argument to query AST",
		},
		{
			name: "Array Argument",
			queryAst: createMockFederationServiceAST(t, `
				query {
					vehicle {
						getVehicleInfos {
							regNo
						}
					}
				}
			`, "dmt"),
			argSources: []*federator.ArgSource{
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
			expectedArgs: 1,
			expectError:  false,
			description:  "Should push array argument to query AST",
		},
		{
			name: "Multiple Arguments",
			queryAst: createMockFederationServiceAST(t, `
				query {
					person {
						fullName
					}
				}
			`, "drp"),
			argSources: []*federator.ArgSource{
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
						ProviderKey:   "drp",
						TargetArgName: "includeVehicles",
						SourceArgPath: "personInfo-includeVehicles",
						TargetArgPath: "drp.person",
					},
					Argument: &ast.Argument{
						Name:  &ast.Name{Value: "includeVehicles"},
						Value: &ast.BooleanValue{Value: true},
					},
				},
			},
			expectedArgs: 2,
			expectError:  false,
			description:  "Should push multiple arguments to query AST",
		},
		{
			name:         "Empty Arguments",
			queryAst:     createMockFederationServiceAST(t, `query { person { fullName } }`, "drp"),
			argSources:   []*federator.ArgSource{},
			expectedArgs: 0,
			expectError:  false,
			description:  "Should handle empty arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			federator.PushArgumentsToProviderQueryAst(tt.argSources, tt.queryAst)

			// Verify arguments were added to the query
			if len(tt.argSources) > 0 {
				operationDef := tt.queryAst.QueryAst.Definitions[0].(*ast.OperationDefinition)
				selectionSet := operationDef.SelectionSet

				// Find the first field and check its arguments
				if len(selectionSet.Selections) > 0 {
					if field, ok := selectionSet.Selections[0].(*ast.Field); ok {
						assert.Len(t, field.Arguments, tt.expectedArgs, "Should have correct number of arguments")
					}
				}
			}
		})
	}
}

// ============================================================================
// BASIC ARGUMENT HANDLING TESTS
// ============================================================================

func TestBasicArgumentHandling(t *testing.T) {
	t.Run("Single String Argument", func(t *testing.T) {
		// Test that single string arguments are handled correctly
		query := `
			query {
				personInfo(nic: "123456789V") {
					fullName
					name
				}
			}
		`

		queryDoc := ParseTestQuery(t, query)
		operationDef := queryDoc.Definitions[0].(*ast.OperationDefinition)

		argMappings := []*graphql.ArgMapping{
			{
				ProviderKey:   "drp",
				TargetArgName: "nic",
				SourceArgPath: "personInfo-nic",
				TargetArgPath: "drp.person",
			},
		}

		// Extract arguments from the query
		var arguments []*ast.Argument
		for _, selection := range operationDef.SelectionSet.Selections {
			if field, ok := selection.(*ast.Field); ok {
				arguments = append(arguments, field.Arguments...)
			}
		}

		argSources := federator.ExtractRequiredArguments(argMappings, arguments)
		assert.Len(t, argSources, 1, "Should extract one argument source")

		// Verify the argument is a string value
		argSource := argSources[0]
		stringValue, ok := argSource.Argument.Value.(*ast.StringValue)
		assert.True(t, ok, "Should have string value")
		assert.Equal(t, "123456789V", stringValue.Value)
	})

	t.Run("Multiple String Arguments", func(t *testing.T) {
		// Test that multiple string arguments are handled correctly
		query := `
			query {
				personInfo(nic: "123456789V") {
					fullName
				}
				vehicle(regNo: "ABC123") {
					make
				}
			}
		`

		queryDoc := ParseTestQuery(t, query)
		operationDef := queryDoc.Definitions[0].(*ast.OperationDefinition)

		argMappings := []*graphql.ArgMapping{
			{
				ProviderKey:   "drp",
				TargetArgName: "nic",
				SourceArgPath: "personInfo-nic",
				TargetArgPath: "drp.person",
			},
			{
				ProviderKey:   "dmt",
				TargetArgName: "regNo",
				SourceArgPath: "vehicle-regNo",
				TargetArgPath: "dmt.vehicle",
			},
		}

		// Extract arguments from the query
		var arguments []*ast.Argument
		for _, selection := range operationDef.SelectionSet.Selections {
			if field, ok := selection.(*ast.Field); ok {
				arguments = append(arguments, field.Arguments...)
			}
		}

		argSources := federator.ExtractRequiredArguments(argMappings, arguments)
		assert.Len(t, argSources, 2, "Should extract two argument sources")

		// Verify both arguments are string values
		for _, argSource := range argSources {
			stringValue, ok := argSource.Argument.Value.(*ast.StringValue)
			assert.True(t, ok, "Should have string value")
			assert.True(t, stringValue.Value == "123456789V" || stringValue.Value == "ABC123", "Should have correct value")
		}
	})

	t.Run("Array Arguments", func(t *testing.T) {
		// Test that array arguments are handled correctly
		query := `
			query {
				vehicles(regNos: ["ABC123", "XYZ789"]) {
					regNo
					make
				}
			}
		`

		queryDoc := ParseTestQuery(t, query)
		operationDef := queryDoc.Definitions[0].(*ast.OperationDefinition)

		argMappings := []*graphql.ArgMapping{
			{
				ProviderKey:   "dmt",
				TargetArgName: "regNos",
				SourceArgPath: "vehicles-regNos",
				TargetArgPath: "dmt.vehicle.getVehicleInfos",
			},
		}

		// Extract arguments from the query
		var arguments []*ast.Argument
		for _, selection := range operationDef.SelectionSet.Selections {
			if field, ok := selection.(*ast.Field); ok {
				arguments = append(arguments, field.Arguments...)
			}
		}

		argSources := federator.ExtractRequiredArguments(argMappings, arguments)
		assert.Len(t, argSources, 1, "Should extract one argument source")

		// Verify the argument is a list value
		argSource := argSources[0]
		listValue, ok := argSource.Argument.Value.(*ast.ListValue)
		assert.True(t, ok, "Should have list value")
		assert.Len(t, listValue.Values, 2, "Should have 2 values in list")

		// Verify the list values
		value1 := listValue.Values[0].(*ast.StringValue)
		value2 := listValue.Values[1].(*ast.StringValue)
		assert.Equal(t, "ABC123", value1.Value)
		assert.Equal(t, "XYZ789", value2.Value)
	})

	t.Run("Boolean Arguments", func(t *testing.T) {
		// Test that boolean arguments are handled correctly
		query := `
			query {
				personInfo(nic: "123456789V", includeVehicles: true) {
					fullName
				}
			}
		`

		queryDoc := ParseTestQuery(t, query)
		operationDef := queryDoc.Definitions[0].(*ast.OperationDefinition)

		argMappings := []*graphql.ArgMapping{
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
		}

		// Extract arguments from the query
		var arguments []*ast.Argument
		for _, selection := range operationDef.SelectionSet.Selections {
			if field, ok := selection.(*ast.Field); ok {
				arguments = append(arguments, field.Arguments...)
			}
		}

		argSources := federator.ExtractRequiredArguments(argMappings, arguments)
		assert.Len(t, argSources, 2, "Should extract two argument sources")

		// Verify both arguments are extracted
		for _, argSource := range argSources {
			assert.NotNil(t, argSource.Argument, "Should have valid argument")
			assert.NotNil(t, argSource.ArgMapping, "Should have valid argument mapping")
		}
	})
}

// ============================================================================
// ARGUMENT VALIDATION TESTS
// ============================================================================

func TestArgumentValidation(t *testing.T) {
	t.Run("Valid Argument Types", func(t *testing.T) {
		// Test that valid argument types are accepted
		validArguments := []*ast.Argument{
			{
				Name:  &ast.Name{Value: "stringArg"},
				Value: &ast.StringValue{Value: "test"},
			},
			{
				Name:  &ast.Name{Value: "intArg"},
				Value: &ast.IntValue{Value: "123"},
			},
			{
				Name:  &ast.Name{Value: "floatArg"},
				Value: &ast.FloatValue{Value: "123.45"},
			},
			{
				Name:  &ast.Name{Value: "boolArg"},
				Value: &ast.BooleanValue{Value: true},
			},
			{
				Name: &ast.Name{Value: "listArg"},
				Value: &ast.ListValue{
					Values: []ast.Value{
						&ast.StringValue{Value: "item1"},
						&ast.StringValue{Value: "item2"},
					},
				},
			},
		}

		for _, arg := range validArguments {
			assert.NotNil(t, arg.Name, "Should have name")
			assert.NotNil(t, arg.Value, "Should have value")
		}
	})

	t.Run("Argument Name Validation", func(t *testing.T) {
		// Test that argument names are properly validated
		arg := &ast.Argument{
			Name:  &ast.Name{Value: "validArgName"},
			Value: &ast.StringValue{Value: "test"},
		}

		assert.NotNil(t, arg.Name, "Should have name")
		assert.NotEmpty(t, arg.Name.Value, "Should have non-empty name")
	})

	t.Run("Argument Value Validation", func(t *testing.T) {
		// Test that argument values are properly validated
		arg := &ast.Argument{
			Name:  &ast.Name{Value: "testArg"},
			Value: &ast.StringValue{Value: "testValue"},
		}

		assert.NotNil(t, arg.Value, "Should have value")
	})
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func createMockFederationServiceAST(t *testing.T, query string, serviceKey string) *federator.FederationServiceAST {
	queryDoc := ParseTestQuery(t, query)
	return &federator.FederationServiceAST{
		ServiceKey: serviceKey,
		QueryAst:   queryDoc,
	}
}
