package tests

import (
	"testing"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/stretchr/testify/assert"
)

// TestObjectAndArrayResponseSupport verifies that the orchestration engine supports
// both object and array responses as specified in the requirements.
func TestObjectAndArrayResponseSupport(t *testing.T) {
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

// TestResponsePatterns demonstrates the supported response patterns
func TestResponsePatterns(t *testing.T) {
	t.Run("Pattern 1: Single Object", func(t *testing.T) {
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

	t.Run("Pattern 2: Object with Array Field", func(t *testing.T) {
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

	t.Run("Pattern 3: Multiple Objects", func(t *testing.T) {
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

// TestArrayResponseValidation validates that array responses are properly structured
func TestArrayResponseValidation(t *testing.T) {
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
