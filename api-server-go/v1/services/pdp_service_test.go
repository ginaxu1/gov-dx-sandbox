package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"github.com/stretchr/testify/assert"
)

func TestPDPService_CreatePolicyMetadata(t *testing.T) {
	t.Run("CreatePolicyMetadata_Success", func(t *testing.T) {
		// Create a mock HTTP server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/api/v1/policy/metadata", r.URL.Path)
			assert.Equal(t, "test-api-key", r.Header.Get("apikey"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			// Return success response
			response := models.PolicyMetadataCreateResponse{
				Records: []models.PolicyMetadataResponse{
					{FieldName: "test.field", SchemaID: "schema-123"},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		service := NewPDPService(server.URL, "test-api-key")
		result, err := service.CreatePolicyMetadata("schema-123", "type Query { test: String }")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Records, 1)
	})

	t.Run("CreatePolicyMetadata_InvalidSDL", func(t *testing.T) {
		service := NewPDPService("http://localhost:9999", "test-api-key")
		result, err := service.CreatePolicyMetadata("schema-123", "invalid sdl syntax {")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to parse SDL")
	})

	t.Run("CreatePolicyMetadata_ServerError", func(t *testing.T) {
		// Create a mock HTTP server that returns error
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}))
		defer server.Close()

		service := NewPDPService(server.URL, "test-api-key")
		result, err := service.CreatePolicyMetadata("schema-123", "type Query { test: String }")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "PDP returned status 500")
	})

	t.Run("CreatePolicyMetadata_NetworkError", func(t *testing.T) {
		// Use a non-existent URL to simulate network error
		service := NewPDPService("http://localhost:9999", "test-api-key")
		result, err := service.CreatePolicyMetadata("schema-123", "type Query { test: String }")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to send request to PDP")
	})
}

func TestPDPService_UpdateAllowList(t *testing.T) {
	t.Run("UpdateAllowList_Success", func(t *testing.T) {
		// Create a mock HTTP server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/api/v1/policy/update-allowlist", r.URL.Path)
			assert.Equal(t, "test-api-key", r.Header.Get("apikey"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			// Parse request body
			var req models.AllowListUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "app-123", req.ApplicationID)

			// Return success response
			response := models.AllowListUpdateResponse{
				Records: []models.AllowListUpdateResponseRecord{
					{FieldName: "field1", SchemaID: "schema-123", ExpiresAt: "2024-12-31T00:00:00Z", UpdatedAt: "2024-01-01T00:00:00Z"},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		service := NewPDPService(server.URL, "test-api-key")
		req := models.AllowListUpdateRequest{
			ApplicationID: "app-123",
			Records: []models.SelectedFieldRecord{
				{FieldName: "field1", SchemaID: "schema-123"},
			},
			GrantDuration: models.GrantDurationTypeOneMonth,
		}

		result, err := service.UpdateAllowList(req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Records, 1)
	})

	t.Run("UpdateAllowList_ServerError", func(t *testing.T) {
		// Create a mock HTTP server that returns error
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad Request"))
		}))
		defer server.Close()

		service := NewPDPService(server.URL, "test-api-key")
		req := models.AllowListUpdateRequest{
			ApplicationID: "app-123",
			Records: []models.SelectedFieldRecord{
				{FieldName: "field1", SchemaID: "schema-123"},
			},
			GrantDuration: models.GrantDurationTypeOneMonth,
		}

		result, err := service.UpdateAllowList(req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "PDP returned status 400")
	})

	t.Run("UpdateAllowList_NetworkError", func(t *testing.T) {
		// Use a non-existent URL to simulate network error
		service := NewPDPService("http://localhost:9999", "test-api-key")
		req := models.AllowListUpdateRequest{
			ApplicationID: "app-123",
			Records: []models.SelectedFieldRecord{
				{FieldName: "field1", SchemaID: "schema-123"},
			},
			GrantDuration: models.GrantDurationTypeOneMonth,
		}

		result, err := service.UpdateAllowList(req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to send request to PDP")
	})
}

func TestPDPService_HealthCheck(t *testing.T) {
	t.Run("HealthCheck_Success", func(t *testing.T) {
		// Create a mock HTTP server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/health", r.URL.Path)
			assert.Equal(t, "test-api-key", r.Header.Get("apikey"))
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		service := NewPDPService(server.URL, "test-api-key")
		err := service.HealthCheck()

		assert.NoError(t, err)
	})

	t.Run("HealthCheck_ServerError", func(t *testing.T) {
		// Create a mock HTTP server that returns error
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		service := NewPDPService(server.URL, "test-api-key")
		err := service.HealthCheck()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "health check failed")
	})

	t.Run("HealthCheck_NetworkError", func(t *testing.T) {
		// Use a non-existent URL to simulate network error
		service := NewPDPService("http://localhost:9999", "test-api-key")
		err := service.HealthCheck()

		assert.Error(t, err)
	})
}

func TestPDPService_setAuthHeader(t *testing.T) {
	t.Run("setAuthHeader_SetsAPIKey", func(t *testing.T) {
		service := NewPDPService("http://localhost:8082", "test-api-key")
		req, _ := http.NewRequest("GET", "http://localhost:8082/health", nil)

		service.setAuthHeader(req)

		assert.Equal(t, "test-api-key", req.Header.Get("apikey"))
	})
}
