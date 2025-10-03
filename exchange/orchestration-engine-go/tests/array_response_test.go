package tests

import (
	"testing"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/federator"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

// TestArrayResponseHandling verifies that the orchestration engine can properly handle
// both object and array responses. Focus on response structure, not array arguments.
func TestArrayResponseHandling(t *testing.T) {
	t.Run("Single Object Response", func(t *testing.T) {
		// Test that single object responses work correctly
		query := `
			query {
				personInfo(nic: "123456789V") {
					fullName @sourceInfo(providerKey: "drp", providerField: "person.fullName")
					name @sourceInfo(providerKey: "rgd", providerField: "getPersonInfo.name")
					address @sourceInfo(providerKey: "drp", providerField: "person.permanentAddress")
				}
			}
		`

		queryDoc := ParseTestQuery(t, query)
		_ = CreateTestSchema(t)

		// Note: Using test-specific configuration instead of modifying global config

		// Note: This test focuses on response accumulation, not query building

		// Mock federated response for single object
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

		// Accumulate response using the original query document
		response := federator.AccumulateResponse(queryDoc, federatedResponse)

		// Verify single object response structure
		assert.NotNil(t, response.Data)
		assert.Contains(t, response.Data, "personInfo")

		personInfo := response.Data["personInfo"].(map[string]interface{})
		assert.Equal(t, "John Doe", personInfo["fullName"])
		assert.Equal(t, "John", personInfo["name"])
		assert.Equal(t, "123 Main St", personInfo["address"])
	})

	t.Run("Array Field Response", func(t *testing.T) {
		// Test that array fields within objects work correctly
		query := `
			query {
				personInfo(nic: "123456789V") {
					fullName @sourceInfo(providerKey: "drp", providerField: "person.fullName")
					ownedVehicles @sourceInfo(providerKey: "dmt", providerField: "vehicle.getVehicleInfos.data") {
						regNo @sourceInfo(providerKey: "dmt", providerField: "registrationNumber")
						make @sourceInfo(providerKey: "dmt", providerField: "make")
						model @sourceInfo(providerKey: "dmt", providerField: "model")
					}
				}
			}
		`

		queryDoc := ParseTestQuery(t, query)
		_ = CreateTestSchema(t)

		// Note: Using test-specific configuration instead of modifying global config

		// Note: This test focuses on response accumulation, not query building

		// Mock federated response with array data
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

		// Accumulate response using the original query document
		response := federator.AccumulateResponse(queryDoc, federatedResponse)

		// Verify array field response structure
		assert.NotNil(t, response.Data)
		assert.Contains(t, response.Data, "personInfo")

		personInfo := response.Data["personInfo"].(map[string]interface{})
		assert.Equal(t, "John Doe", personInfo["fullName"])

		// Verify array field - the accumulator returns an array of maps for array fields
		ownedVehicles := personInfo["ownedVehicles"]
		assert.NotNil(t, ownedVehicles, "ownedVehicles should be present")

		// The accumulator returns an array of maps with structured objects
		vehiclesArray, ok := ownedVehicles.([]map[string]interface{})
		assert.True(t, ok, "ownedVehicles should be an array")
		assert.Len(t, vehiclesArray, 2, "Should have 2 vehicles")

		// Each vehicle is a map with individual values
		vehicle1 := vehiclesArray[0]
		vehicle2 := vehiclesArray[1]

		// Check that we have the expected fields
		assert.Contains(t, vehicle1, "regNo")
		assert.Contains(t, vehicle1, "make")
		assert.Contains(t, vehicle1, "model")

		// Each field should contain individual values (not arrays)
		regNo1 := vehicle1["regNo"]
		make1 := vehicle1["make"]
		model1 := vehicle1["model"]

		regNo2 := vehicle2["regNo"]
		make2 := vehicle2["make"]
		model2 := vehicle2["model"]

		// Verify the values - each vehicle should have one value per field
		assert.Equal(t, "ABC123", regNo1)
		assert.Equal(t, "Toyota", make1)
		assert.Equal(t, "Camry", model1)

		// Second vehicle should have different values
		assert.Equal(t, "XYZ789", regNo2)
		assert.Equal(t, "Honda", make2)
		assert.Equal(t, "Civic", model2)
	})

	t.Skip("Skipping bulk query test - not implemented yet")
	t.Run("Array Field Response (Future Enhancement)", func(t *testing.T) {
		// Test that array fields within objects work correctly
		// This demonstrates support for arrays in responses like ownedVehicles: [VehicleInfo]
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

		queryDoc := ParseTestQuery(t, query)
		_ = CreateTestSchema(t)

		// Mock federated response with bulk array data
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

		// Accumulate response
		response := federator.AccumulateResponse(queryDoc, federatedResponse)

		// Verify bulk array response structure
		assert.NotNil(t, response.Data)
		assert.Contains(t, response.Data, "personInfos")

		personInfos := response.Data["personInfos"].([]map[string]interface{})
		assert.Len(t, personInfos, 2)

		// Verify first person
		person1 := personInfos[0]
		assert.Equal(t, "John Doe", person1["fullName"])
		assert.Equal(t, "John", person1["name"])
		assert.Equal(t, "123 Main St", person1["address"])

		// Verify second person
		person2 := personInfos[1]
		assert.Equal(t, "Jane Smith", person2["fullName"])
		assert.Equal(t, "Jane", person2["name"])
		assert.Equal(t, "456 Oak Ave", person2["address"])
	})

	t.Run("Mixed Object and Array Response", func(t *testing.T) {
		// Test that queries with both object and array fields work correctly
		// This demonstrates support for both objects and arrays in the same response
		query := `
			query {
				personInfo(nic: "123456789V") {
					fullName
					ownedVehicles {
						regNo
						make
					}
				}
			}
		`

		queryDoc := ParseTestQuery(t, query)
		_ = CreateTestSchema(t)

		// Mock federated response with mixed data
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

		// Accumulate response
		response := federator.AccumulateResponse(queryDoc, federatedResponse)

		// Verify mixed response structure
		assert.NotNil(t, response.Data)

		// Verify personInfo object
		assert.Contains(t, response.Data, "personInfo")
		personInfo := response.Data["personInfo"].(map[string]interface{})
		assert.Equal(t, "John Doe", personInfo["fullName"])

		// Verify ownedVehicles array
		ownedVehicles := personInfo["ownedVehicles"].([]map[string]interface{})
		assert.Len(t, ownedVehicles, 2)

		// Verify vehicles array (bulk query result)
		assert.Contains(t, response.Data, "vehicles")
		vehicles := response.Data["vehicles"].([]interface{})
		assert.Len(t, vehicles, 2)
	})

	t.Run("Empty Array Response", func(t *testing.T) {
		// Test that empty arrays are handled correctly
		query := `
			query {
				personInfo(nic: "123456789V") {
					fullName
					ownedVehicles {
						regNo
						make
					}
				}
			}
		`

		queryDoc := ParseTestQuery(t, query)
		_ = CreateTestSchema(t)

		// Mock federated response with empty array
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
									"data": []interface{}{},
								},
							},
						},
					},
				},
			},
		}

		// Accumulate response
		response := federator.AccumulateResponse(queryDoc, federatedResponse)

		// Verify empty array response structure
		assert.NotNil(t, response.Data)
		assert.Contains(t, response.Data, "personInfo")

		personInfo := response.Data["personInfo"].(map[string]interface{})
		assert.Equal(t, "John Doe", personInfo["fullName"])

		// Verify empty array
		ownedVehicles := personInfo["ownedVehicles"].([]map[string]interface{})
		assert.Len(t, ownedVehicles, 0)
	})

	t.Run("Nested Array Response", func(t *testing.T) {
		// Test that deeply nested arrays work correctly
		query := `
			query {
				personInfo(nic: "123456789V") {
					fullName
					ownedVehicles {
						regNo
						maintenanceRecords {
							date
							description
						}
					}
				}
			}
		`

		queryDoc := ParseTestQuery(t, query)
		_ = CreateTestSchema(t)

		// Mock federated response with nested arrays
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
											"maintenanceRecords": []interface{}{
												map[string]interface{}{
													"date":        "2023-01-15",
													"description": "Oil change",
												},
												map[string]interface{}{
													"date":        "2023-06-15",
													"description": "Brake inspection",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		// Accumulate response
		response := federator.AccumulateResponse(queryDoc, federatedResponse)

		// Verify nested array response structure
		assert.NotNil(t, response.Data)
		assert.Contains(t, response.Data, "personInfo")

		personInfo := response.Data["personInfo"].(map[string]interface{})
		assert.Equal(t, "John Doe", personInfo["fullName"])

		// Verify outer array
		ownedVehicles := personInfo["ownedVehicles"].([]map[string]interface{})
		assert.Len(t, ownedVehicles, 1)

		// Verify inner array
		vehicle := ownedVehicles[0]
		assert.Equal(t, "ABC123", vehicle["regNo"])

		maintenanceRecords := vehicle["maintenanceRecords"].([]interface{})
		assert.Len(t, maintenanceRecords, 2)

		// Verify first maintenance record
		record1 := maintenanceRecords[0].(map[string]interface{})
		assert.Equal(t, "2023-01-15", record1["date"])
		assert.Equal(t, "Oil change", record1["description"])

		// Verify second maintenance record
		record2 := maintenanceRecords[1].(map[string]interface{})
		assert.Equal(t, "2023-06-15", record2["date"])
		assert.Equal(t, "Brake inspection", record2["description"])
	})
}

// TestArrayResponseErrorHandling verifies that errors in array responses are handled correctly
func TestArrayResponseErrorHandling(t *testing.T) {
	t.Skip("Skipping error handling test - requires @sourceInfo directives")
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
							"personInfos": []interface{}{
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
		response := federator.AccumulateResponse(queryDoc, federatedResponse)

		// Verify partial array response structure
		assert.NotNil(t, response.Data)
		assert.Contains(t, response.Data, "personInfos")

		// Handle both array and object responses
		personInfosValue := response.Data["personInfos"]
		var personInfos []interface{}

		if arr, ok := personInfosValue.([]interface{}); ok {
			personInfos = arr
		} else if obj, ok := personInfosValue.(map[string]interface{}); ok {
			// Convert object to array for testing
			personInfos = []interface{}{obj}
		} else {
			t.Fatalf("Expected array or object, got %T", personInfosValue)
		}

		assert.Len(t, personInfos, 1) // Only one person due to partial failure

		// Verify the successful person
		person1 := personInfos[0].(map[string]interface{})
		assert.Equal(t, "John Doe", person1["fullName"])
		assert.Equal(t, "John", person1["name"])
	})

	t.Run("Provider Response Error", func(t *testing.T) {
		// Test that provider errors are handled correctly
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
		response := federator.AccumulateResponse(queryDoc, federatedResponse)

		// Verify error handling
		assert.NotNil(t, response.Data)
		assert.Contains(t, response.Data, "personInfo")

		personInfo := response.Data["personInfo"].(map[string]interface{})
		assert.Equal(t, "John Doe", personInfo["fullName"])

		// Verify that ownedVehicles is not present due to provider error
		_, exists := personInfo["ownedVehicles"]
		assert.False(t, exists, "ownedVehicles should not be present due to provider error")
	})
}

// Helper functions
