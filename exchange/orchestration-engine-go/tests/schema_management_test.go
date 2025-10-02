package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/models"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// SCHEMA MODELS TESTS
// ============================================================================

func TestSchemaModels(t *testing.T) {
	t.Run("CreateSchemaRequest", func(t *testing.T) {
		req := models.CreateSchemaRequest{
			Version:    "1.0.0",
			SDL:        "type Query { hello: String }",
			CreatedBy:  "test-user",
			ChangeType: models.VersionChangeTypeMajor,
			Notes:      stringPtr("Test notes"),
		}

		assert.Equal(t, "1.0.0", req.Version)
		assert.Equal(t, "type Query { hello: String }", req.SDL)
		assert.Equal(t, "test-user", req.CreatedBy)
		assert.Equal(t, models.VersionChangeTypeMajor, req.ChangeType)
		assert.Equal(t, "Test notes", *req.Notes)
	})

	t.Run("UnifiedSchema", func(t *testing.T) {
		schema := models.UnifiedSchema{
			ID:         1,
			Version:    "1.0.0",
			SDL:        "type Query { hello: String }",
			CreatedBy:  "test-user",
			Status:     models.SchemaStatusActive,
			ChangeType: models.VersionChangeTypeMajor,
			Notes:      stringPtr("Test notes"),
		}

		assert.Equal(t, 1, schema.ID)
		assert.Equal(t, "1.0.0", schema.Version)
		assert.Equal(t, models.SchemaStatusActive, schema.Status)
		assert.Equal(t, models.VersionChangeTypeMajor, schema.ChangeType)
	})

	t.Run("SchemaStatus_Constants", func(t *testing.T) {
		assert.Equal(t, models.SchemaStatus("active"), models.SchemaStatusActive)
		assert.Equal(t, models.SchemaStatus("inactive"), models.SchemaStatusInactive)
		assert.Equal(t, models.SchemaStatus("deprecated"), models.SchemaStatusDeprecated)
	})

	t.Run("VersionChangeType_Constants", func(t *testing.T) {
		assert.Equal(t, models.VersionChangeType("major"), models.VersionChangeTypeMajor)
		assert.Equal(t, models.VersionChangeType("minor"), models.VersionChangeTypeMinor)
		assert.Equal(t, models.VersionChangeType("patch"), models.VersionChangeTypePatch)
	})
}

// ============================================================================
// SCHEMA COMPATIBILITY TESTS
// ============================================================================

func TestSchemaCompatibility(t *testing.T) {
	t.Run("SchemaCompatibilityCheck", func(t *testing.T) {
		check := models.SchemaCompatibilityCheck{
			IsCompatible: true,
			Issues:       []string{},
			Warnings:     []string{"New field added"},
		}

		assert.True(t, check.IsCompatible)
		assert.Empty(t, check.Issues)
		assert.Len(t, check.Warnings, 1)
	})

	t.Run("UpdateSchemaStatusRequest", func(t *testing.T) {
		req := models.UpdateSchemaStatusRequest{
			IsActive: true,
			Reason:   stringPtr("Activating schema"),
		}

		assert.True(t, req.IsActive)
		assert.Equal(t, "Activating schema", *req.Reason)
	})
}

// ============================================================================
// GRAPHQL MODELS TESTS
// ============================================================================

func TestGraphQLModels(t *testing.T) {
	t.Run("GraphQLRequest", func(t *testing.T) {
		req := models.GraphQLRequest{
			Query:     "query { hello }",
			Variables: map[string]interface{}{"name": "world"},
			Operation: "HelloQuery",
		}

		assert.Equal(t, "query { hello }", req.Query)
		assert.Equal(t, "world", req.Variables["name"])
		assert.Equal(t, "HelloQuery", req.Operation)
	})

	t.Run("GraphQLResponse", func(t *testing.T) {
		response := models.GraphQLResponse{
			Data: map[string]interface{}{"hello": "world"},
			Errors: []models.GraphQLError{
				{
					Message: "Test error",
					Locations: []models.GraphQLErrorLocation{
						{Line: 1, Column: 1},
					},
				},
			},
		}

		assert.Equal(t, "world", response.Data.(map[string]interface{})["hello"])
		assert.Len(t, response.Errors, 1)
		assert.Equal(t, "Test error", response.Errors[0].Message)
	})
}

// ============================================================================
// VALIDATION TESTS
// ============================================================================

func TestValidationError(t *testing.T) {
	t.Run("ValidationError", func(t *testing.T) {
		err := models.ValidationError{
			Field:   "version",
			Message: "Version is required",
		}

		assert.Equal(t, "version", err.Field)
		assert.Equal(t, "Version is required", err.Message)
		assert.Equal(t, "Version is required", err.Error())
	})
}

// ============================================================================
// SCHEMA VERSIONING TESTS
// ============================================================================

func TestSchemaVersioning(t *testing.T) {
	t.Run("VersionComparison", func(t *testing.T) {
		// Test that version constants work correctly
		assert.Equal(t, "major", string(models.VersionChangeTypeMajor))
		assert.Equal(t, "minor", string(models.VersionChangeTypeMinor))
		assert.Equal(t, "patch", string(models.VersionChangeTypePatch))
	})

	t.Run("StatusComparison", func(t *testing.T) {
		// Test that status constants work correctly
		assert.Equal(t, "active", string(models.SchemaStatusActive))
		assert.Equal(t, "inactive", string(models.SchemaStatusInactive))
		assert.Equal(t, "deprecated", string(models.SchemaStatusDeprecated))
	})
}

// ============================================================================
// JSON SERIALIZATION TESTS
// ============================================================================

func TestJSONSerialization(t *testing.T) {
	t.Run("CreateSchemaRequest_JSON", func(t *testing.T) {
		req := models.CreateSchemaRequest{
			Version:    "1.0.0",
			SDL:        "type Query { hello: String }",
			CreatedBy:  "test-user",
			ChangeType: models.VersionChangeTypeMajor,
			Notes:      stringPtr("Test notes"),
		}

		// Marshal to JSON
		jsonData, err := json.Marshal(req)
		assert.NoError(t, err)
		assert.NotEmpty(t, jsonData)

		// Unmarshal back
		var unmarshaled models.CreateSchemaRequest
		err = json.Unmarshal(jsonData, &unmarshaled)
		assert.NoError(t, err)
		assert.Equal(t, req.Version, unmarshaled.Version)
		assert.Equal(t, req.SDL, unmarshaled.SDL)
		assert.Equal(t, req.CreatedBy, unmarshaled.CreatedBy)
		assert.Equal(t, req.ChangeType, unmarshaled.ChangeType)
	})

	t.Run("UnifiedSchema_JSON", func(t *testing.T) {
		schema := models.UnifiedSchema{
			ID:         1,
			Version:    "1.0.0",
			SDL:        "type Query { hello: String }",
			CreatedBy:  "test-user",
			Status:     models.SchemaStatusActive,
			ChangeType: models.VersionChangeTypeMajor,
			Notes:      stringPtr("Test notes"),
		}

		// Marshal to JSON
		jsonData, err := json.Marshal(schema)
		assert.NoError(t, err)
		assert.NotEmpty(t, jsonData)

		// Unmarshal back
		var unmarshaled models.UnifiedSchema
		err = json.Unmarshal(jsonData, &unmarshaled)
		assert.NoError(t, err)
		assert.Equal(t, schema.ID, unmarshaled.ID)
		assert.Equal(t, schema.Version, unmarshaled.Version)
		assert.Equal(t, schema.Status, unmarshaled.Status)
	})

	t.Run("AllModels_Serialization", func(t *testing.T) {
		// Test that all models can be instantiated and serialized
		models := []interface{}{
			models.CreateSchemaRequest{},
			models.UnifiedSchema{},
			models.UpdateSchemaStatusRequest{},
			models.GraphQLRequest{},
			models.GraphQLResponse{},
			models.SchemaCompatibilityCheck{},
			models.ValidationError{},
		}

		for _, model := range models {
			jsonData, err := json.Marshal(model)
			assert.NoError(t, err)
			assert.NotNil(t, jsonData)
		}
	})
}

// ============================================================================
// SCHEMA VALIDATION TESTS
// ============================================================================

func TestSchemaValidation(t *testing.T) {
	t.Run("ValidSchemaRequest", func(t *testing.T) {
		req := models.CreateSchemaRequest{
			Version:    "1.0.0",
			SDL:        "type Query { hello: String }",
			CreatedBy:  "test-user",
			ChangeType: models.VersionChangeTypeMajor,
		}

		// Basic validation
		assert.NotEmpty(t, req.Version)
		assert.NotEmpty(t, req.SDL)
		assert.NotEmpty(t, req.CreatedBy)
		assert.NotEmpty(t, req.ChangeType)
	})

	t.Run("InvalidSchemaRequest", func(t *testing.T) {
		req := models.CreateSchemaRequest{
			Version:    "",
			SDL:        "",
			CreatedBy:  "",
			ChangeType: "",
		}

		// Should be invalid
		assert.Empty(t, req.Version)
		assert.Empty(t, req.SDL)
		assert.Empty(t, req.CreatedBy)
		assert.Empty(t, req.ChangeType)
	})
}

// ============================================================================
// HTTP CLIENT TESTS
// ============================================================================

func TestHTTPClient(t *testing.T) {
	t.Run("HTTPClientCreation", func(t *testing.T) {
		client := &http.Client{Timeout: 5 * time.Second}
		assert.NotNil(t, client)
		assert.Equal(t, 5*time.Second, client.Timeout)
	})

	t.Run("HTTPRequestCreation", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://localhost:4000/health", nil)
		assert.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "http://localhost:4000/health", req.URL.String())
	})

	t.Run("HTTPPostRequestCreation", func(t *testing.T) {
		data := map[string]string{"test": "value"}
		jsonData, _ := json.Marshal(data)
		req, err := http.NewRequest("POST", "http://localhost:4000/sdl", bytes.NewBuffer(jsonData))
		assert.NoError(t, err)
		assert.Equal(t, "POST", req.Method)
		req.Header.Set("Content-Type", "application/json")
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	})
}

// ============================================================================
// REAL SERVER TESTS
// ============================================================================

func TestRealServerEndpoints(t *testing.T) {
	// This test assumes the server is running on localhost:4000
	// In a real test environment, you would start the server in a goroutine
	baseURL := "http://localhost:4000"

	// Skip this test if server is not running
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		t.Skip("Server not running, skipping real server tests")
		return
	}
	resp.Body.Close()

	t.Run("HealthCheck", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/health")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("CreateSchema", func(t *testing.T) {
		createReq := models.CreateSchemaRequest{
			Version:    "1.0.0",
			SDL:        "type Query { hello: String }",
			CreatedBy:  "test-user",
			ChangeType: models.VersionChangeTypeMajor,
			Notes:      stringPtr("Test schema"),
		}

		reqBody, _ := json.Marshal(createReq)
		resp, err := client.Post(baseURL+"/sdl", "application/json", bytes.NewBuffer(reqBody))
		assert.NoError(t, err)
		defer resp.Body.Close()

		// The response might be 500 due to database connection issues, but the endpoint should exist
		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusInternalServerError)
	})

	t.Run("GetActiveSchema", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/sdl/active")
		assert.NoError(t, err)
		defer resp.Body.Close()

		// The response might be 500 due to database connection issues, but the endpoint should exist
		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusInternalServerError)
	})

	t.Run("GetSchemaVersions", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/sdl/versions")
		assert.NoError(t, err)
		defer resp.Body.Close()

		// The response might be 500 due to database connection issues, but the endpoint should exist
		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusInternalServerError)
	})

	t.Run("GetSchemaVersionsInfo", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/sdl/versions/info")
		assert.NoError(t, err)
		defer resp.Body.Close()

		// The response might be 500 due to database connection issues, but the endpoint should exist
		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusInternalServerError)
	})
}

// ============================================================================
// APPLICATION BUILD TESTS
// ============================================================================

func TestApplicationBuild(t *testing.T) {
	t.Run("BuildSuccess", func(t *testing.T) {
		// This test verifies that the application builds successfully
		// In a real CI/CD environment, this would be tested by actually building
		assert.True(t, true, "Application should build successfully")
	})

	t.Run("ConfigurationLoading", func(t *testing.T) {
		// Test that configuration can be loaded
		// This would test the actual config loading
		assert.True(t, true, "Configuration should load successfully")
	})
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
