package tests

import (
	"testing"

	"github.com/gov-dx-sandbox/exchange/orchestration-engine-go/federator"
	"github.com/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/graphql"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFixedNestedArrayWithSchema tests the fix using the new schema-aware accumulator
func TestFixedNestedArrayWithSchema(t *testing.T) {
	t.Run("TestNestedArrayWithSchema", func(t *testing.T) {
		// Create a schema that includes the class field as a nested array
		schemaSDL := `
			directive @sourceInfo(
				providerKey: String!
				providerField: String!
			) on FIELD_DEFINITION

			type Query {
				vehicleInfo(regNo: String!): VehicleInfo
			}

			type VehicleInfo {
				regNo: String @sourceInfo(providerKey: "dmt", providerField: "vehicle.registrationNumber")
				make: String @sourceInfo(providerKey: "dmt", providerField: "vehicle.make")
				model: String @sourceInfo(providerKey: "dmt", providerField: "vehicle.model")
				class: [VehicleClass] @sourceInfo(providerKey: "dmt", providerField: "vehicle.classes")
			}

			type VehicleClass {
				className: String @sourceInfo(providerKey: "dmt", providerField: "vehicle.classes.className")
				classCode: String @sourceInfo(providerKey: "dmt", providerField: "vehicle.classes.classCode")
			}
		`

		// Parse the schema
		src := source.NewSource(&source.Source{
			Body: []byte(schemaSDL),
			Name: "TestSchema",
		})
		schema, err := parser.Parse(parser.ParseParams{Source: src})
		require.NoError(t, err, "Should parse schema successfully")

		// Test query that requests nested array
		query := `
			query {
				vehicleInfo(regNo: "ABC123") {
					regNo
					make
					model
					class {
						className
						classCode
					}
				}
			}
		`

		queryDoc := ParseTestQuery(t, query)
		assert.NotNil(t, queryDoc, "Should parse query successfully")

		// Mock provider response with nested array data
		mockProviderResponse := map[string]interface{}{
			"vehicle": map[string]interface{}{
				"registrationNumber": "ABC123",
				"make":               "Toyota",
				"model":              "Camry",
				"classes": []interface{}{
					map[string]interface{}{
						"className": "Sedan",
						"classCode": "SED",
					},
					map[string]interface{}{
						"className": "Passenger Vehicle",
						"classCode": "PV",
					},
				},
			},
		}

		// Create federated response
		federatedResponse := &federator.FederationResponse{
			ServiceKey: "test",
			Responses: []*federator.ProviderResponse{
				&federator.ProviderResponse{
					ServiceKey: "dmt",
					Response: graphql.Response{
						Data: mockProviderResponse,
					},
				},
			},
		}

		// Test the new schema-aware accumulator
		result := federator.AccumulateResponseWithSchema(queryDoc, federatedResponse, schema)
		assert.NotNil(t, result, "Should return result")
		assert.NotNil(t, result.Data, "Should have data")

		// Verify the response structure
		resultMap := result.Data
		assert.Contains(t, resultMap, "vehicleInfo", "Should contain vehicleInfo")

		vehicleInfo := resultMap["vehicleInfo"].(map[string]interface{})

		// Verify scalar fields
		assert.Equal(t, "ABC123", vehicleInfo["regNo"], "Should have regNo")
		assert.Equal(t, "Toyota", vehicleInfo["make"], "Should have make")
		assert.Equal(t, "Camry", vehicleInfo["model"], "Should have model")

		// Verify nested array field - THIS SHOULD NOW WORK
		classArray, exists := vehicleInfo["class"]
		assert.True(t, exists, "Should have class field")
		assert.NotNil(t, classArray, "Class field should not be nil")

		classSlice, ok := classArray.([]map[string]interface{})
		assert.True(t, ok, "Class should be an array of maps")
		assert.Len(t, classSlice, 2, "Should have 2 class items")

		// Verify first class item
		class1 := classSlice[0]
		assert.Equal(t, "Sedan", class1["className"], "Should have correct className")
		assert.Equal(t, "SED", class1["classCode"], "Should have correct classCode")

		// Verify second class item
		class2 := classSlice[1]
		assert.Equal(t, "Passenger Vehicle", class2["className"], "Should have correct className")
		assert.Equal(t, "PV", class2["classCode"], "Should have correct classCode")
	})

	t.Run("TestNestedArrayWithEmptyData", func(t *testing.T) {
		// Create a schema
		schemaSDL := `
			directive @sourceInfo(
				providerKey: String!
				providerField: String!
			) on FIELD_DEFINITION

			type Query {
				vehicleInfo(regNo: String!): VehicleInfo
			}

			type VehicleInfo {
				regNo: String @sourceInfo(providerKey: "dmt", providerField: "vehicle.registrationNumber")
				class: [VehicleClass] @sourceInfo(providerKey: "dmt", providerField: "vehicle.classes")
			}

			type VehicleClass {
				className: String @sourceInfo(providerKey: "dmt", providerField: "vehicle.classes.className")
			}
		`

		src := source.NewSource(&source.Source{
			Body: []byte(schemaSDL),
			Name: "TestSchema",
		})
		schema, err := parser.Parse(parser.ParseParams{Source: src})
		require.NoError(t, err, "Should parse schema successfully")

		// Test query with empty class array
		query := `
			query {
				vehicleInfo(regNo: "XYZ789") {
					regNo
					class {
						className
					}
				}
			}
		`

		queryDoc := ParseTestQuery(t, query)

		mockProviderResponse := map[string]interface{}{
			"vehicle": map[string]interface{}{
				"registrationNumber": "XYZ789",
				"classes":            []interface{}{}, // Empty array
			},
		}

		federatedResponse := &federator.FederationResponse{
			ServiceKey: "test",
			Responses: []*federator.ProviderResponse{
				&federator.ProviderResponse{
					ServiceKey: "dmt",
					Response: graphql.Response{
						Data: mockProviderResponse,
					},
				},
			},
		}

		result := federator.AccumulateResponseWithSchema(queryDoc, federatedResponse, schema)
		assert.NotNil(t, result, "Should return result")

		resultMap := result.Data
		vehicleInfo := resultMap["vehicleInfo"].(map[string]interface{})

		// Should have empty array, not nil
		classArray := vehicleInfo["class"]
		assert.NotNil(t, classArray, "Class field should not be nil even if empty")

		classSlice, ok := classArray.([]map[string]interface{})
		assert.True(t, ok, "Class should be an array of maps")
		assert.Len(t, classSlice, 0, "Should have empty array")
	})
}
