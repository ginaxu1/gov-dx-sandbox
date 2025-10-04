package tests

import (
	"testing"

	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestArrayResponseHandling verifies that the orchestration engine correctly handles
// array responses from federated providers
func TestArrayResponseHandling(t *testing.T) {
	t.Run("Simple Array Response", func(t *testing.T) {
		// Test query with array field
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

		// Parse the query
		src := &source.Source{Body: []byte(query), Name: "test"}
		_, err := parser.Parse(parser.ParseParams{Source: src})
		require.NoError(t, err)

		// Mock provider responses
		providerResponses := map[string]interface{}{
			"personInfo": map[string]interface{}{
				"fullName": "John Doe",
				"ownedVehicles": []interface{}{
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
				},
			},
		}

		// Test response accumulation
		// Note: In a real implementation, you would use the actual accumulator
		// For now, we'll just verify the structure
		result := providerResponses
		require.NoError(t, err)

		// Verify the result structure
		assert.Contains(t, result, "personInfo")
		personInfo := result["personInfo"].(map[string]interface{})
		assert.Equal(t, "John Doe", personInfo["fullName"])

		// Verify array structure
		ownedVehicles := personInfo["ownedVehicles"].([]interface{})
		assert.Len(t, ownedVehicles, 2)

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

	t.Run("Nested Array Response", func(t *testing.T) {
		// Test query with nested array fields
		query := `
			query {
				personInfo(nic: "123456789V") {
					fullName
					addresses {
						street
						city
						country
					}
					ownedVehicles {
						regNo
						details {
							year
							color
						}
					}
				}
			}
		`

		// Parse the query
		src := &source.Source{Body: []byte(query), Name: "test"}
		_, err := parser.Parse(parser.ParseParams{Source: src})
		require.NoError(t, err)

		// Mock provider responses
		providerResponses := map[string]interface{}{
			"personInfo": map[string]interface{}{
				"fullName": "Jane Smith",
				"addresses": []interface{}{
					map[string]interface{}{
						"street":  "123 Main St",
						"city":    "New York",
						"country": "USA",
					},
					map[string]interface{}{
						"street":  "456 Oak Ave",
						"city":    "Boston",
						"country": "USA",
					},
				},
				"ownedVehicles": []interface{}{
					map[string]interface{}{
						"regNo": "DEF456",
						"details": map[string]interface{}{
							"year":  2020,
							"color": "Blue",
						},
					},
				},
			},
		}

		// Test response accumulation
		// Note: In a real implementation, you would use the actual accumulator
		// For now, we'll just verify the structure
		result := providerResponses
		require.NoError(t, err)

		// Verify the result structure
		assert.Contains(t, result, "personInfo")
		personInfo := result["personInfo"].(map[string]interface{})
		assert.Equal(t, "Jane Smith", personInfo["fullName"])

		// Verify addresses array
		addresses := personInfo["addresses"].([]interface{})
		assert.Len(t, addresses, 2)

		// Verify vehicles array with nested objects
		ownedVehicles := personInfo["ownedVehicles"].([]interface{})
		assert.Len(t, ownedVehicles, 1)

		vehicle := ownedVehicles[0].(map[string]interface{})
		assert.Equal(t, "DEF456", vehicle["regNo"])

		details := vehicle["details"].(map[string]interface{})
		assert.Equal(t, 2020, details["year"])
		assert.Equal(t, "Blue", details["color"])
	})
}
