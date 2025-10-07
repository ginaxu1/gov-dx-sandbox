package tests

import (
	"testing"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/federator"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/stretchr/testify/assert"
)

// TestCompleteFederationFlow tests the complete federation flow from query parsing
// to response accumulation, focusing on object and array responses (not array arguments).
func TestCompleteFederationFlow(t *testing.T) {
	t.Skip("Skipping integration test - requires config initialization")
	t.Run("Single Object Federation", func(t *testing.T) {
		// Test complete flow for single object query
		query := `
			query {
				personInfo(nic: "123456789V") {
					fullName
					name
					address
				}
			}
		`

		// Step 1: Parse query
		queryDoc := ParseTestQuery(t, query)
		assert.NotNil(t, queryDoc, "Should parse query successfully")

		// Step 2: Load schema
		schema := CreateTestSchema(t)
		assert.NotNil(t, schema, "Should load schema successfully")

		// Step 3: Extract source info directives
		response, err := federator.ProviderSchemaCollector(schema, queryDoc)
		assert.NoError(t, err, "Should extract source info directives")
		assert.Len(t, response.ProviderFieldMap, 3, "Should extract 3 response.ProviderFieldMap")
		assert.Len(t, response.Arguments, 1, "Should extract 1 argument")

		// Step 4: Build provider queries
		_ = createMockArgMappings()
		argSources := createMockArgSources()
		requests, err := federator.QueryBuilder(response.ProviderFieldMap, argSources)
		assert.NoError(t, err, "Should build provider queries")
		assert.Len(t, requests, 2, "Should create 2 provider requests")

		// Step 5: Mock federated responses
		federatedResponse := &federator.FederationResponse{
			Responses: []federator.ProviderResponse{
				{
					ServiceKey: "drp",
					Response: graphql.Response{
						Data: map[string]interface{}{
							"person": map[string]interface{}{
								"fullName":         "John Doe",
								"permanentAddress": "123 Main St",
							},
						},
					},
				},
				{
					ServiceKey: "rgd",
					Response: graphql.Response{
						Data: map[string]interface{}{
							"getPersonInfo": map[string]interface{}{
								"name": "John",
							},
						},
					},
				},
			},
		}

		// Step 6: Accumulate response
		accumulatedResponse := federator.AccumulateResponse(queryDoc, federatedResponse)
		assert.NotNil(t, accumulatedResponse.Data, "Should have response data")
		assert.Contains(t, accumulatedResponse.Data, "personInfo", "Should contain personInfo")

		// Step 7: Verify final response structure
		personInfo := accumulatedResponse.Data["personInfo"].(map[string]interface{})
		assert.Equal(t, "John Doe", personInfo["fullName"])
		assert.Equal(t, "John", personInfo["name"])
		assert.Equal(t, "123 Main St", personInfo["address"])
	})

	t.Run("Array Field Federation", func(t *testing.T) {
		// Test complete flow for query with array response.ProviderFieldMap
		query := `
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
		`

		// Step 1: Parse query
		queryDoc := ParseTestQuery(t, query)
		assert.NotNil(t, queryDoc, "Should parse query successfully")

		// Step 2: Load schema
		schema := CreateTestSchema(t)
		assert.NotNil(t, schema, "Should load schema successfully")

		// Step 3: Extract source info directives
		response, err := federator.ProviderSchemaCollector(schema, queryDoc)
		assert.NoError(t, err, "Should extract source info directives")
		assert.Len(t, response.ProviderFieldMap, 5, "Should extract 5 response.ProviderFieldMap (including array response.ProviderFieldMap)")
		assert.Len(t, response.Arguments, 1, "Should extract 1 argument")

		// Step 4: Build provider queries
		argSources := createMockArgSources()
		requests, err := federator.QueryBuilder(response.ProviderFieldMap, argSources)
		assert.NoError(t, err, "Should build provider queries")
		assert.Len(t, requests, 2, "Should create 2 provider requests")

		// Step 5: Mock federated responses with array data
		federatedResponse := &federator.FederationResponse{
			Responses: []federator.ProviderResponse{
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
											"model":              "Camry",
										},
										map[string]interface{}{
											"registrationNumber": "XYZ789",
											"make":               "Honda",
											"model":              "Civic",
										},
									},
								},
							},
						},
					},
				},
			},
		}

		// Step 6: Accumulate response
		accumulatedResponse := federator.AccumulateResponse(queryDoc, federatedResponse)
		assert.NotNil(t, accumulatedResponse.Data, "Should have response data")
		assert.Contains(t, accumulatedResponse.Data, "personInfo", "Should contain personInfo")

		// Step 7: Verify final response structure with array
		personInfo := accumulatedResponse.Data["personInfo"].(map[string]interface{})
		assert.Equal(t, "John Doe", personInfo["fullName"])

		// Verify array field
		ownedVehicles := personInfo["ownedVehicles"].([]interface{})
		assert.Len(t, ownedVehicles, 2, "Should have 2 vehicles")

		// Verify first vehicle
		vehicle1 := ownedVehicles[0].(map[string]interface{})
		assert.Equal(t, "ABC123", vehicle1["regNo"])
		assert.Equal(t, "Toyota", vehicle1["make"])
		assert.Equal(t, "Camry", vehicle1["model"])

		// Verify second vehicle
		vehicle2 := ownedVehicles[1].(map[string]interface{})
		assert.Equal(t, "XYZ789", vehicle2["regNo"])
		assert.Equal(t, "Honda", vehicle2["make"])
		assert.Equal(t, "Civic", vehicle2["model"])
	})

	t.Run("Object with Array Field Federation", func(t *testing.T) {
		// Test complete flow for object with array field
		// This demonstrates support for personInfo.ownedVehicles: [VehicleInfo]
		query := `
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
		`

		// Step 1: Parse query
		queryDoc := ParseTestQuery(t, query)
		assert.NotNil(t, queryDoc, "Should parse query successfully")

		// Step 2: Load schema
		schema := CreateTestSchema(t)
		assert.NotNil(t, schema, "Should load schema successfully")

		// Step 3: Extract source info directives
		response, err := federator.ProviderSchemaCollector(schema, queryDoc)
		assert.NoError(t, err, "Should extract source info directives")
		assert.Len(t, response.ProviderFieldMap, 3, "Should extract 3 response.ProviderFieldMap")
		assert.Len(t, response.Arguments, 1, "Should extract 1 argument")

		// Step 4: Build provider queries
		argSources := createMockBulkArgSources()
		requests, err := federator.QueryBuilder(response.ProviderFieldMap, argSources)
		assert.NoError(t, err, "Should build provider queries")
		assert.Len(t, requests, 2, "Should create 2 provider requests")

		// Step 5: Mock federated responses with bulk array data
		federatedResponse := &federator.FederationResponse{
			Responses: []federator.ProviderResponse{
				{
					ServiceKey: "drp",
					Response: graphql.Response{
						Data: map[string]interface{}{
							"persons": []interface{}{
								map[string]interface{}{
									"fullName":         "John Doe",
									"permanentAddress": "123 Main St",
								},
								map[string]interface{}{
									"fullName":         "Jane Smith",
									"permanentAddress": "456 Oak Ave",
								},
							},
						},
					},
				},
				{
					ServiceKey: "rgd",
					Response: graphql.Response{
						Data: map[string]interface{}{
							"getPersonInfos": []interface{}{
								map[string]interface{}{
									"name": "John",
								},
								map[string]interface{}{
									"name": "Jane",
								},
							},
						},
					},
				},
			},
		}

		// Step 6: Accumulate response
		accumulatedResponse := federator.AccumulateResponse(queryDoc, federatedResponse)
		assert.NotNil(t, accumulatedResponse.Data, "Should have response data")
		assert.Contains(t, accumulatedResponse.Data, "personInfos", "Should contain personInfos")

		// Step 7: Verify final response structure with bulk array
		personInfos := accumulatedResponse.Data["personInfos"].([]interface{})
		assert.Len(t, personInfos, 2, "Should have 2 persons")

		// Verify first person
		person1 := personInfos[0].(map[string]interface{})
		assert.Equal(t, "John Doe", person1["fullName"])
		assert.Equal(t, "John", person1["name"])
		assert.Equal(t, "123 Main St", person1["address"])

		// Verify second person
		person2 := personInfos[1].(map[string]interface{})
		assert.Equal(t, "Jane Smith", person2["fullName"])
		assert.Equal(t, "Jane", person2["name"])
		assert.Equal(t, "456 Oak Ave", person2["address"])
	})

	t.Run("Mixed Object and Array Federation", func(t *testing.T) {
		// Test complete flow for query with both object and array response.ProviderFieldMap
		query := `
			query {
				personInfo(nic: "123456789V") {
					fullName
					ownedVehicles {
						regNo
						make
					}
				}
				vehicles(regNos: ["ABC123", "XYZ789"]) {
					regNo
					make
					model
				}
			}
		`

		// Step 1: Parse query
		queryDoc := ParseTestQuery(t, query)
		assert.NotNil(t, queryDoc, "Should parse query successfully")

		// Step 2: Load schema
		schema := CreateTestSchema(t)
		assert.NotNil(t, schema, "Should load schema successfully")

		// Step 3: Extract source info directives
		response, err := federator.ProviderSchemaCollector(schema, queryDoc)
		assert.NoError(t, err, "Should extract source info directives")
		assert.Len(t, response.ProviderFieldMap, 6, "Should extract 6 response.ProviderFieldMap")
		assert.Len(t, response.Arguments, 2, "Should extract 2 arguments")

		// Step 4: Build provider queries
		argSources := createMockMixedArgSources()
		requests, err := federator.QueryBuilder(response.ProviderFieldMap, argSources)
		assert.NoError(t, err, "Should build provider queries")
		assert.Len(t, requests, 2, "Should create 2 provider requests")

		// Step 5: Mock federated responses with mixed data
		federatedResponse := &federator.FederationResponse{
			Responses: []federator.ProviderResponse{
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
											"model":              "Camry",
										},
										map[string]interface{}{
											"registrationNumber": "XYZ789",
											"make":               "Honda",
											"model":              "Civic",
										},
									},
								},
							},
						},
					},
				},
			},
		}

		// Step 6: Accumulate response
		accumulatedResponse := federator.AccumulateResponse(queryDoc, federatedResponse)
		assert.NotNil(t, accumulatedResponse.Data, "Should have response data")

		// Step 7: Verify final response structure with mixed data
		// Verify personInfo object
		assert.Contains(t, accumulatedResponse.Data, "personInfo")
		personInfo := accumulatedResponse.Data["personInfo"].(map[string]interface{})
		assert.Equal(t, "John Doe", personInfo["fullName"])

		// Verify ownedVehicles array
		ownedVehicles := personInfo["ownedVehicles"].([]interface{})
		assert.Len(t, ownedVehicles, 2, "Should have 2 owned vehicles")

		// Verify vehicles array (bulk query result)
		assert.Contains(t, accumulatedResponse.Data, "vehicles")
		vehicles := accumulatedResponse.Data["vehicles"].([]interface{})
		assert.Len(t, vehicles, 2, "Should have 2 vehicles")
	})
}

// TestFederationErrorHandling tests error handling in the federation flow
func TestFederationErrorHandling(t *testing.T) {
	t.Skip("Skipping error handling test - requires @sourceInfo directives")
	t.Run("Provider Error Handling", func(t *testing.T) {
		// Test that provider errors are handled gracefully
		query := `
			query {
				personInfo(nic: "123456789V") {
					fullName
					ownedVehicles {
						regNo
					}
				}
			}
		`

		queryDoc := ParseTestQuery(t, query)
		_ = CreateTestSchema(t)

		// Mock federated response with provider error
		federatedResponse := &federator.FederationResponse{
			Responses: []federator.ProviderResponse{
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
				{
					ServiceKey: "dmt",
					Response: graphql.Response{
						Data: nil,
						Errors: []interface{}{
							map[string]interface{}{
								"message": "Provider unavailable",
								"extensions": map[string]interface{}{
									"code": "PROVIDER_ERROR",
								},
							},
						},
					},
				},
			},
		}

		// Accumulate response
		accumulatedResponse := federator.AccumulateResponse(queryDoc, federatedResponse)

		// Verify error handling
		assert.NotNil(t, accumulatedResponse.Data, "Should have response data")
		assert.Contains(t, accumulatedResponse.Data, "personInfo", "Should contain personInfo")

		personInfo := accumulatedResponse.Data["personInfo"].(map[string]interface{})
		assert.Equal(t, "John Doe", personInfo["fullName"])

		// Verify that ownedVehicles is not present due to provider error
		_, exists := personInfo["ownedVehicles"]
		assert.False(t, exists, "ownedVehicles should not be present due to provider error")
	})

	t.Run("Partial Array Failure", func(t *testing.T) {
		// Test that partial failures in array responses are handled gracefully
		query := `
			query {
				personInfos(nics: ["123456789V", "987654321V"]) {
					fullName
					name
				}
			}
		`

		queryDoc := ParseTestQuery(t, query)
		_ = CreateTestSchema(t)

		// Mock federated response with partial failure
		federatedResponse := &federator.FederationResponse{
			Responses: []federator.ProviderResponse{
				{
					ServiceKey: "drp",
					Response: graphql.Response{
						Data: map[string]interface{}{
							"persons": []interface{}{
								map[string]interface{}{
									"fullName": "John Doe",
								},
								// Missing second person due to error
							},
						},
					},
				},
				{
					ServiceKey: "rgd",
					Response: graphql.Response{
						Data: map[string]interface{}{
							"getPersonInfos": []interface{}{
								map[string]interface{}{
									"name": "John",
								},
								// Missing second person due to error
							},
						},
					},
				},
			},
		}

		// Accumulate response
		accumulatedResponse := federator.AccumulateResponse(queryDoc, federatedResponse)

		// Verify partial array response structure
		assert.NotNil(t, accumulatedResponse.Data, "Should have response data")
		assert.Contains(t, accumulatedResponse.Data, "personInfos", "Should contain personInfos")

		personInfos := accumulatedResponse.Data["personInfos"].([]interface{})
		assert.Len(t, personInfos, 1, "Should have 1 person due to partial failure")

		// Verify the successful person
		person1 := personInfos[0].(map[string]interface{})
		assert.Equal(t, "John Doe", person1["fullName"])
		assert.Equal(t, "John", person1["name"])
	})
}

// Helper functions

func createMockArgMappings() []*graphql.ArgMapping {
	return []*graphql.ArgMapping{
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
	}
}

func createMockArgSources() []*federator.ArgSource {
	return []*federator.ArgSource{
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
	}
}

func createMockBulkArgSources() []*federator.ArgSource {
	return []*federator.ArgSource{
		{
			ArgMapping: &graphql.ArgMapping{
				ProviderKey:   "drp",
				TargetArgName: "nics",
				SourceArgPath: "personInfos-nics",
				TargetArgPath: "drp.persons",
			},
			Argument: &ast.Argument{
				Name: &ast.Name{Value: "nics"},
				Value: &ast.ListValue{
					Values: []ast.Value{
						&ast.StringValue{Value: "123456789V"},
						&ast.StringValue{Value: "987654321V"},
					},
				},
			},
		},
		{
			ArgMapping: &graphql.ArgMapping{
				ProviderKey:   "rgd",
				TargetArgName: "nics",
				SourceArgPath: "personInfos-nics",
				TargetArgPath: "rgd.getPersonInfos",
			},
			Argument: &ast.Argument{
				Name: &ast.Name{Value: "nics"},
				Value: &ast.ListValue{
					Values: []ast.Value{
						&ast.StringValue{Value: "123456789V"},
						&ast.StringValue{Value: "987654321V"},
					},
				},
			},
		},
	}
}

func createMockMixedArgSources() []*federator.ArgSource {
	return []*federator.ArgSource{
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
	}
}
