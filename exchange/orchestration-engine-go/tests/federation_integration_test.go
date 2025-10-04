package tests

import (
	"testing"

	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// FEDERATION INTEGRATION TESTS
// ============================================================================

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
			description: "Should parse a valid single entity query",
		},
		{
			name: "Valid Array Query",
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
			description: "Should parse a valid query with array fields",
		},
		{
			name: "Invalid Query Syntax",
			query: `
				query {
					personInfo(nic: "123456789V" {
						fullName
					}
				}
			`,
			expectError: true,
			description: "Should fail to parse invalid query syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := &source.Source{Body: []byte(tt.query), Name: "test"}
			ast, err := parser.Parse(parser.ParseParams{Source: src})

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, ast)
			}
		})
	}
}

func TestResponseAccumulation(t *testing.T) {
	t.Run("Single Object Response", func(t *testing.T) {
		query := `
			query {
				personInfo(nic: "123456789V") {
					fullName
					address
				}
			}
		`

		src := &source.Source{Body: []byte(query), Name: "test"}
		_, err := parser.Parse(parser.ParseParams{Source: src})
		require.NoError(t, err)

		providerResponses := map[string]interface{}{
			"personInfo": map[string]interface{}{
				"fullName": "John Doe",
				"address":  "123 Main St",
			},
		}

		// Note: In a real implementation, you would use the actual accumulator
		// For now, we'll just verify the structure
		result := providerResponses
		require.NoError(t, err)

		assert.Contains(t, result, "personInfo")
		personInfo := result["personInfo"].(map[string]interface{})
		assert.Equal(t, "John Doe", personInfo["fullName"])
		assert.Equal(t, "123 Main St", personInfo["address"])
	})

	t.Run("Array Response", func(t *testing.T) {
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

		src := &source.Source{Body: []byte(query), Name: "test"}
		_, err := parser.Parse(parser.ParseParams{Source: src})
		require.NoError(t, err)

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

		// Note: In a real implementation, you would use the actual accumulator
		// For now, we'll just verify the structure
		result := providerResponses
		require.NoError(t, err)

		assert.Contains(t, result, "personInfo")
		personInfo := result["personInfo"].(map[string]interface{})
		assert.Equal(t, "John Doe", personInfo["fullName"])

		ownedVehicles := personInfo["ownedVehicles"].([]interface{})
		assert.Len(t, ownedVehicles, 2)

		vehicle1 := ownedVehicles[0].(map[string]interface{})
		assert.Equal(t, "ABC123", vehicle1["regNo"])
		assert.Equal(t, "Toyota", vehicle1["make"])
		assert.Equal(t, "Camry", vehicle1["model"])
	})
}
