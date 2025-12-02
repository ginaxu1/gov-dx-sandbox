package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gov-dx-sandbox/portal-backend/models"
	"github.com/stretchr/testify/assert"
)

func TestNewPDPClient(t *testing.T) {
	client := NewPDPClient("http://localhost:8082")
	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:8082", client.baseURL)
	assert.NotNil(t, client.httpClient)
}

func TestPDPClient_UpdateProviderMetadata(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/metadata/update", r.URL.Path)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var req models.ProviderMetadataUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)

			response := models.ProviderMetadataUpdateResponse{
				Success: true,
				Updated: 1,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewPDPClient(server.URL)
		req := models.ProviderMetadataUpdateRequest{
			ApplicationID: "app-123",
			Fields: []models.ProviderFieldGrant{
				{FieldName: "field1", GrantDuration: "1h"},
			},
		}

		resp, err := client.UpdateProviderMetadata(req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Success)
	})

	t.Run("HTTPError", func(t *testing.T) {
		client := NewPDPClient("http://invalid-url:9999")
		req := models.ProviderMetadataUpdateRequest{
			ApplicationID: "app-123",
		}

		resp, err := client.UpdateProviderMetadata(req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("Non200Status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad Request"))
		}))
		defer server.Close()

		client := NewPDPClient(server.URL)
		req := models.ProviderMetadataUpdateRequest{
			ApplicationID: "app-123",
		}

		resp, err := client.UpdateProviderMetadata(req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "status 400")
	})

	t.Run("InvalidJSONResponse", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		client := NewPDPClient(server.URL)
		req := models.ProviderMetadataUpdateRequest{
			ApplicationID: "app-123",
		}

		resp, err := client.UpdateProviderMetadata(req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestPDPClient_HealthCheck(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/health", r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewPDPClient(server.URL)
		err := client.HealthCheck()

		assert.NoError(t, err)
	})

	t.Run("HTTPError", func(t *testing.T) {
		client := NewPDPClient("http://invalid-url:9999")
		err := client.HealthCheck()

		assert.Error(t, err)
	})

	t.Run("Non200Status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer server.Close()

		client := NewPDPClient(server.URL)
		err := client.HealthCheck()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "status 503")
	})
}
