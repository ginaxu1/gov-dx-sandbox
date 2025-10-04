package tests

import (
	"testing"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/federator"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// RESPONSE ACCUMULATION TESTS
// ============================================================================

func TestAccumulateResponse_SingleObject(t *testing.T) {
	// Test query with @sourceInfo directives
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

	// Mock federated response
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

	// Mock schema with source info directives
	_ = CreateTestSchema(t)

	// Accumulate response
	response := federator.AccumulateResponse(queryDoc, federatedResponse)

	// Verify response structure
	assert.NotNil(t, response.Data)
	assert.Contains(t, response.Data, "personInfo")

	personInfo := response.Data["personInfo"].(map[string]interface{})
	assert.Equal(t, "John Doe", personInfo["fullName"])
	assert.Equal(t, "John", personInfo["name"])
	assert.Equal(t, "123 Main St", personInfo["address"])
}

func TestAccumulateResponse_ArrayField(t *testing.T) {
	// Test query with array field and @sourceInfo directives
	query := `
		query {
			personInfo(nic: "123456789V") {
				fullName @sourceInfo(providerKey: "drp", providerField: "person.fullName")
				ownedVehicles @sourceInfo(providerKey: "dmt", providerField: "vehicle.getVehicleInfos.data") {
					regNo @sourceInfo(providerKey: "dmt", providerField: "vehicle.getVehicleInfos.data.registrationNumber")
					make @sourceInfo(providerKey: "dmt", providerField: "vehicle.getVehicleInfos.data.make")
					model @sourceInfo(providerKey: "dmt", providerField: "vehicle.getVehicleInfos.data.model")
				}
			}
		}
	`

	queryDoc := ParseTestQuery(t, query)

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

	// Mock schema with source info directives
	_ = CreateTestSchema(t)

	// Accumulate response
	response := federator.AccumulateResponse(queryDoc, federatedResponse)

	// Verify response structure
	assert.NotNil(t, response.Data)
	assert.Contains(t, response.Data, "personInfo")

	personInfo := response.Data["personInfo"].(map[string]interface{})
	assert.Equal(t, "John Doe", personInfo["fullName"])

	// Verify array field
	ownedVehicles := personInfo["ownedVehicles"].([]map[string]interface{})
	assert.Len(t, ownedVehicles, 2)

	// Verify first vehicle
	vehicle1 := ownedVehicles[0]
	assert.Equal(t, "ABC123", vehicle1["regNo"])
	assert.Equal(t, "Toyota", vehicle1["make"])
	assert.Equal(t, "Camry", vehicle1["model"])

	// Verify second vehicle
	vehicle2 := ownedVehicles[1]
	assert.Equal(t, "XYZ789", vehicle2["regNo"])
	assert.Equal(t, "Honda", vehicle2["make"])
	assert.Equal(t, "Civic", vehicle2["model"])
}

func TestAccumulateResponse_EmptyArray(t *testing.T) {
	// Test that empty arrays are handled correctly
	query := `
		query {
			personInfo(nic: "123456789V") {
				fullName @sourceInfo(providerKey: "drp", providerField: "person.fullName")
				ownedVehicles @sourceInfo(providerKey: "dmt", providerField: "vehicle.getVehicleInfos.data") {
					regNo @sourceInfo(providerKey: "dmt", providerField: "registrationNumber")
					make @sourceInfo(providerKey: "dmt", providerField: "make")
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
}

func TestAccumulateResponse_NestedArray(t *testing.T) {
	// Test that deeply nested arrays work correctly
	query := `
		query {
			personInfo(nic: "123456789V") {
				fullName @sourceInfo(providerKey: "drp", providerField: "person.fullName")
				ownedVehicles @sourceInfo(providerKey: "dmt", providerField: "vehicle.getVehicleInfos.data") {
					regNo @sourceInfo(providerKey: "dmt", providerField: "registrationNumber")
					maintenanceRecords @sourceInfo(providerKey: "dmt", providerField: "maintenanceRecords") {
						date @sourceInfo(providerKey: "dmt", providerField: "date")
						description @sourceInfo(providerKey: "dmt", providerField: "description")
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
}

// ============================================================================
// RESPONSE PATTERN TESTS
// ============================================================================

func TestResponsePatterns(t *testing.T) {
	t.Run("Single Object Response", func(t *testing.T) {
		// Test: personInfo(nic: String): PersonInfo
		// Expected: {personInfo: {fullName: "John Doe", name: "John"}}

		obj := map[string]interface{}{}
		path := "personInfo"
		value := map[string]interface{}{
			"fullName": "John Doe",
			"name":     "John",
			"address":  "123 Main St",
		}

		result, err := PushValue(obj, path, value)
		assert.NoError(t, err, "Should not return error")
		assert.NotNil(t, result, "Should return result")

		// Verify object response structure
		resultMap := result.(map[string]interface{})
		assert.Contains(t, resultMap, "personInfo", "Should contain personInfo")

		personInfo := resultMap["personInfo"].(map[string]interface{})
		assert.Equal(t, "John Doe", personInfo["fullName"], "Should have correct fullName")
		assert.Equal(t, "John", personInfo["name"], "Should have correct name")
		assert.Equal(t, "123 Main St", personInfo["address"], "Should have correct address")
	})

	t.Run("Object with Array Field Response", func(t *testing.T) {
		// Test: personInfo(nic: String): { fullName: String, ownedVehicles: [VehicleInfo] }
		// Expected: {personInfo: {fullName: "John Doe", ownedVehicles: [...]}}

		obj := map[string]interface{}{}

		// First add the person info
		personPath := "personInfo"
		personValue := map[string]interface{}{
			"fullName": "John Doe",
		}

		result, err := PushValue(obj, personPath, personValue)
		assert.NoError(t, err, "Should not return error")

		// Then add the array field
		vehiclesPath := "personInfo.ownedVehicles"
		vehiclesValue := []interface{}{
			map[string]interface{}{
				"regNo": "ABC123",
				"make":  "Toyota",
				"model": "Camry",
			},
			map[string]interface{}{
				"regNo": "XYZ789",
				"make":  "Honda",
				"model": "Civic",
			},
		}

		result, err = PushValue(result, vehiclesPath, vehiclesValue)
		assert.NoError(t, err, "Should not return error")
		assert.NotNil(t, result, "Should return result")

		// Verify object with array field response structure
		resultMap := result.(map[string]interface{})
		assert.Contains(t, resultMap, "personInfo", "Should contain personInfo")

		personInfo := resultMap["personInfo"].(map[string]interface{})
		assert.Equal(t, "John Doe", personInfo["fullName"], "Should have correct fullName")

		// Verify array field
		ownedVehicles := personInfo["ownedVehicles"].([]interface{})
		assert.Len(t, ownedVehicles, 2, "Should have 2 vehicles")

		// Verify first vehicle
		vehicle1 := ownedVehicles[0].(map[string]interface{})
		assert.Equal(t, "ABC123", vehicle1["regNo"], "Should have correct first vehicle regNo")
		assert.Equal(t, "Toyota", vehicle1["make"], "Should have correct first vehicle make")
		assert.Equal(t, "Camry", vehicle1["model"], "Should have correct first vehicle model")

		// Verify second vehicle
		vehicle2 := ownedVehicles[1].(map[string]interface{})
		assert.Equal(t, "XYZ789", vehicle2["regNo"], "Should have correct second vehicle regNo")
		assert.Equal(t, "Honda", vehicle2["make"], "Should have correct second vehicle make")
		assert.Equal(t, "Civic", vehicle2["model"], "Should have correct second vehicle model")
	})

	t.Run("Nested Array Response", func(t *testing.T) {
		// Test: personInfo.ownedVehicles.maintenanceRecords: [MaintenanceRecord]
		// Expected: {personInfo: {ownedVehicles: [{maintenanceRecords: [...]}]}}

		obj := map[string]interface{}{}

		// Create nested array structure
		maintenancePath := "personInfo.ownedVehicles.0.maintenanceRecords"
		maintenanceValue := []interface{}{
			map[string]interface{}{
				"date":        "2023-01-15",
				"description": "Oil change",
			},
			map[string]interface{}{
				"date":        "2023-06-15",
				"description": "Brake inspection",
			},
		}

		result, err := PushValue(obj, maintenancePath, maintenanceValue)
		assert.NoError(t, err, "Should not return error")
		assert.NotNil(t, result, "Should return result")

		// Verify nested array response structure
		resultMap := result.(map[string]interface{})
		personInfo := resultMap["personInfo"].(map[string]interface{})

		// Handle both array and map structures
		var ownedVehicles []interface{}
		if vehiclesArray, ok := personInfo["ownedVehicles"].([]interface{}); ok {
			ownedVehicles = vehiclesArray
		} else if vehiclesMap, ok := personInfo["ownedVehicles"].(map[string]interface{}); ok {
			// Convert map to array for testing
			ownedVehicles = []interface{}{vehiclesMap["0"]}
		} else {
			t.Fatalf("Expected array or map, got %T", personInfo["ownedVehicles"])
		}

		// Verify we have one vehicle
		assert.Len(t, ownedVehicles, 1, "Should have 1 vehicle")
		vehicle := ownedVehicles[0].(map[string]interface{})
		maintenanceRecords := vehicle["maintenanceRecords"].([]interface{})
		assert.Len(t, maintenanceRecords, 2, "Should have 2 maintenance records")

		// Verify first maintenance record
		record1 := maintenanceRecords[0].(map[string]interface{})
		assert.Equal(t, "2023-01-15", record1["date"], "Should have correct first record date")
		assert.Equal(t, "Oil change", record1["description"], "Should have correct first record description")

		// Verify second maintenance record
		record2 := maintenanceRecords[1].(map[string]interface{})
		assert.Equal(t, "2023-06-15", record2["date"], "Should have correct second record date")
		assert.Equal(t, "Brake inspection", record2["description"], "Should have correct second record description")
	})

	t.Run("Empty Array Response", func(t *testing.T) {
		// Test: personInfo.ownedVehicles: [] (empty array)
		// Expected: {personInfo: {ownedVehicles: []}}

		obj := map[string]interface{}{}
		path := "personInfo.ownedVehicles"
		value := []interface{}{} // Empty array

		result, err := PushValue(obj, path, value)
		assert.NoError(t, err, "Should not return error")
		assert.NotNil(t, result, "Should return result")

		// Verify empty array response structure
		resultMap := result.(map[string]interface{})
		personInfo := resultMap["personInfo"].(map[string]interface{})
		ownedVehicles := personInfo["ownedVehicles"].([]interface{})
		assert.Len(t, ownedVehicles, 0, "Should have empty array")
	})

	t.Run("Mixed Object and Array Fields", func(t *testing.T) {
		// Test: personInfo with both scalar fields and array fields
		// Expected: {personInfo: {fullName: "John", ownedVehicles: [...]}}

		obj := map[string]interface{}{}

		// Add scalar field
		scalarPath := "personInfo.fullName"
		scalarValue := "John Doe"
		result, err := PushValue(obj, scalarPath, scalarValue)
		assert.NoError(t, err, "Should not return error")

		// Add array field
		arrayPath := "personInfo.ownedVehicles"
		arrayValue := []interface{}{
			map[string]interface{}{"regNo": "ABC123"},
		}
		result, err = PushValue(result, arrayPath, arrayValue)
		assert.NoError(t, err, "Should not return error")

		// Verify mixed response structure
		resultMap := result.(map[string]interface{})
		personInfo := resultMap["personInfo"].(map[string]interface{})

		// Verify scalar field
		assert.Equal(t, "John Doe", personInfo["fullName"], "Should have scalar field")

		// Verify array field
		ownedVehicles := personInfo["ownedVehicles"].([]interface{})
		assert.Len(t, ownedVehicles, 1, "Should have array field")
		assert.Equal(t, "ABC123", ownedVehicles[0].(map[string]interface{})["regNo"], "Should have correct array element")
	})
}

// ============================================================================
// RESPONSE VALIDATION TESTS
// ============================================================================

func TestResponseValidation(t *testing.T) {
	t.Run("Array Element Structure", func(t *testing.T) {
		// Test that array elements maintain proper structure
		vehicles := []interface{}{
			map[string]interface{}{
				"regNo": "ABC123",
				"make":  "Toyota",
				"model": "Camry",
			},
			map[string]interface{}{
				"regNo": "XYZ789",
				"make":  "Honda",
				"model": "Civic",
			},
		}

		// Verify array structure
		assert.Len(t, vehicles, 2, "Should have 2 vehicles")

		for i, vehicle := range vehicles {
			vehicleMap := vehicle.(map[string]interface{})
			assert.Contains(t, vehicleMap, "regNo", "Vehicle %d should have regNo", i)
			assert.Contains(t, vehicleMap, "make", "Vehicle %d should have make", i)
			assert.Contains(t, vehicleMap, "model", "Vehicle %d should have model", i)
		}
	})

	t.Run("Array Response Path Extraction", func(t *testing.T) {
		// Test that we can extract values from array responses
		data := map[string]interface{}{
			"personInfo": map[string]interface{}{
				"ownedVehicles": []interface{}{
					map[string]interface{}{
						"regNo": "ABC123",
						"make":  "Toyota",
					},
				},
			},
		}

		// Extract array from path
		value, err := GetValueAtPath(data, "personInfo.ownedVehicles")
		assert.NoError(t, err, "Should extract array value")

		vehicles := value.([]interface{})
		assert.Len(t, vehicles, 1, "Should have 1 vehicle")

		vehicle := vehicles[0].(map[string]interface{})
		assert.Equal(t, "ABC123", vehicle["regNo"], "Should have correct regNo")
		assert.Equal(t, "Toyota", vehicle["make"], "Should have correct make")
	})
}

// ============================================================================
// UTILITY FUNCTION TESTS
// ============================================================================

func TestPushValue_ArrayHandling(t *testing.T) {
	tests := []struct {
		name        string
		obj         interface{}
		path        string
		value       interface{}
		expected    interface{}
		expectError bool
		description string
	}{
		{
			name:        "Push to empty object",
			obj:         map[string]interface{}{},
			path:        "personInfo.fullName",
			value:       "John Doe",
			expected:    map[string]interface{}{"personInfo": map[string]interface{}{"fullName": "John Doe"}},
			expectError: false,
			description: "Should push value to empty object",
		},
		{
			name:        "Push to existing object",
			obj:         map[string]interface{}{"personInfo": map[string]interface{}{"name": "John"}},
			path:        "personInfo.fullName",
			value:       "John Doe",
			expected:    map[string]interface{}{"personInfo": map[string]interface{}{"name": "John", "fullName": "John Doe"}},
			expectError: false,
			description: "Should push value to existing object",
		},
		{
			name:        "Push to array field",
			obj:         map[string]interface{}{},
			path:        "personInfo.ownedVehicles",
			value:       []interface{}{map[string]interface{}{"regNo": "ABC123"}},
			expected:    map[string]interface{}{"personInfo": map[string]interface{}{"ownedVehicles": []interface{}{map[string]interface{}{"regNo": "ABC123"}}}},
			expectError: false,
			description: "Should push array value to object",
		},
		{
			name:        "Push to nested array",
			obj:         map[string]interface{}{"personInfo": map[string]interface{}{"ownedVehicles": []interface{}{}}},
			path:        "personInfo.ownedVehicles.regNo",
			value:       "ABC123",
			expected:    map[string]interface{}{"personInfo": map[string]interface{}{"ownedVehicles": []interface{}{}}},
			expectError: false,
			description: "Should push value to nested array (applies to all elements)",
		},
		{
			name:        "Push to bulk array",
			obj:         map[string]interface{}{},
			path:        "personInfos",
			value:       []interface{}{map[string]interface{}{"fullName": "John"}, map[string]interface{}{"fullName": "Jane"}},
			expected:    map[string]interface{}{"personInfos": []interface{}{map[string]interface{}{"fullName": "John"}, map[string]interface{}{"fullName": "Jane"}}},
			expectError: false,
			description: "Should push bulk array to object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := federator.PushValue(tt.obj, tt.path, tt.value)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				assert.Equal(t, tt.expected, result, tt.description)
			}
		})
	}
}

func TestGetValueAtPath_ArrayHandling(t *testing.T) {
	tests := []struct {
		name        string
		data        interface{}
		path        string
		expected    interface{}
		expectError bool
		description string
	}{
		{
			name: "Get value from object",
			data: map[string]interface{}{
				"person": map[string]interface{}{
					"fullName": "John Doe",
				},
			},
			path:        "person.fullName",
			expected:    "John Doe",
			expectError: false,
			description: "Should get value from object path",
		},
		{
			name: "Get value from array",
			data: map[string]interface{}{
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
			path:        "vehicle.getVehicleInfos.data",
			expected:    []interface{}{map[string]interface{}{"registrationNumber": "ABC123", "make": "Toyota"}, map[string]interface{}{"registrationNumber": "XYZ789", "make": "Honda"}},
			expectError: false,
			description: "Should get array value from path",
		},
		{
			name: "Get value from array elements",
			data: map[string]interface{}{
				"persons": []interface{}{
					map[string]interface{}{
						"fullName": "John Doe",
					},
					map[string]interface{}{
						"fullName": "Jane Smith",
					},
				},
			},
			path:        "persons.fullName",
			expected:    []interface{}{"John Doe", "Jane Smith"},
			expectError: false,
			description: "Should get values from all array elements",
		},
		{
			name: "Get value from nested array elements",
			data: map[string]interface{}{
				"personInfo": map[string]interface{}{
					"ownedVehicles": []interface{}{
						map[string]interface{}{
							"regNo": "ABC123",
							"make":  "Toyota",
						},
					},
				},
			},
			path:        "personInfo.ownedVehicles.regNo",
			expected:    []interface{}{"ABC123"},
			expectError: false,
			description: "Should get values from all nested array elements",
		},
		{
			name: "Get non-existent key",
			data: map[string]interface{}{
				"person": map[string]interface{}{
					"name": "John",
				},
			},
			path:        "person.fullName",
			expected:    nil,
			expectError: true,
			description: "Should return error for non-existent key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := federator.GetValueAtPath(tt.data, tt.path)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				assert.Equal(t, tt.expected, result, tt.description)
			}
		})
	}
}

// ============================================================================
// QUERY PATTERN TESTS
// ============================================================================

func TestQueryPatterns(t *testing.T) {
	t.Run("Single Object Query", func(t *testing.T) {
		// Pattern: personInfo(nic: String): PersonInfo
		// Response: {personInfo: {...}}

		query := `
			query {
				personInfo(nic: "123456789V") {
					fullName
					name
					address
				}
			}
		`

		queryDoc := ParseTestQuery(t, query)
		assert.NotNil(t, queryDoc, "Should parse single object query")

		// Verify query structure
		operationDef := queryDoc.Definitions[0].(*ast.OperationDefinition)
		selectionSet := operationDef.SelectionSet
		assert.Len(t, selectionSet.Selections, 1, "Should have one selection")

		field := selectionSet.Selections[0].(*ast.Field)
		assert.Equal(t, "personInfo", field.Name.Value, "Should have personInfo field")
	})

	t.Run("Object with Array Field Query", func(t *testing.T) {
		// Pattern: personInfo(nic: String): { fullName: String, ownedVehicles: [VehicleInfo] }
		// Response: {personInfo: {fullName: "...", ownedVehicles: [...]}}

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
		assert.NotNil(t, queryDoc, "Should parse object with array field query")

		// Verify query structure
		operationDef := queryDoc.Definitions[0].(*ast.OperationDefinition)
		selectionSet := operationDef.SelectionSet
		assert.Len(t, selectionSet.Selections, 1, "Should have one selection")

		field := selectionSet.Selections[0].(*ast.Field)
		assert.Equal(t, "personInfo", field.Name.Value, "Should have personInfo field")

		// Verify nested selection set (ownedVehicles array)
		nestedSelectionSet := field.SelectionSet
		assert.Len(t, nestedSelectionSet.Selections, 2, "Should have fullName and ownedVehicles")

		// Find ownedVehicles field
		var ownedVehiclesField *ast.Field
		for _, selection := range nestedSelectionSet.Selections {
			if field, ok := selection.(*ast.Field); ok && field.Name.Value == "ownedVehicles" {
				ownedVehiclesField = field
				break
			}
		}
		assert.NotNil(t, ownedVehiclesField, "Should have ownedVehicles field")

		// Verify ownedVehicles has nested fields
		vehiclesSelectionSet := ownedVehiclesField.SelectionSet
		assert.Len(t, vehiclesSelectionSet.Selections, 3, "Should have regNo, make, model fields")
	})

	t.Run("Multiple Objects Query", func(t *testing.T) {
		// Pattern: personInfo and vehicle queries
		// Response: {personInfo: {...}, vehicle: {...}}

		query := `
			query {
				personInfo(nic: "123456789V") {
					fullName
				}
				vehicle(regNo: "ABC123") {
					make
					model
				}
			}
		`

		queryDoc := ParseTestQuery(t, query)
		assert.NotNil(t, queryDoc, "Should parse multiple objects query")

		// Verify query structure
		operationDef := queryDoc.Definitions[0].(*ast.OperationDefinition)
		selectionSet := operationDef.SelectionSet
		assert.Len(t, selectionSet.Selections, 2, "Should have two selections")

		// Verify both fields exist
		fieldNames := make([]string, len(selectionSet.Selections))
		for i, selection := range selectionSet.Selections {
			if field, ok := selection.(*ast.Field); ok {
				fieldNames[i] = field.Name.Value
			}
		}
		assert.Contains(t, fieldNames, "personInfo", "Should contain personInfo")
		assert.Contains(t, fieldNames, "vehicle", "Should contain vehicle")
	})
}
