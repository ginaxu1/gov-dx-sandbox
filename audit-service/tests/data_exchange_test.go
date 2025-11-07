package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gov-dx-sandbox/audit-service/models"
)

// TestDataExchangeEndpoint tests the POST /v1/audit/exchange endpoint
func TestDataExchangeEndpoint(t *testing.T) {
	server := SetupTestServerWithGORM(t)
	defer server.Close()

	t.Run("CreateDataExchangeEvent_Success", func(t *testing.T) {
		reqBody := models.DataExchangeEvent{
			EventID:           "550e8400-e29b-41d4-a716-446655440000",
			Timestamp:         "2024-01-15T10:30:00Z",
			ActorUserID:       "user-123",
			ConsumerAppID:     "passport-app",
			ConsumerID:        "member-consumer-123",
			OnBehalfOfOwnerID: "citizen-abc",
			ProviderSchemaID:  "hospital-schema-v1",
			ProviderID:        "member-provider-456",
			RequestedFields:   []string{"personInfo.name", "personInfo.address"},
			Status:            "SUCCESS",
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		req := httptest.NewRequest("POST", "/v1/audit/exchange", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.DataExchangeHandler.CreateDataExchangeEvent(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
		}

		var response models.Log
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.ApplicationID != reqBody.ConsumerAppID {
			t.Errorf("Expected applicationId %s, got %s", reqBody.ConsumerAppID, response.ApplicationID)
		}

		if response.SchemaID != reqBody.ProviderSchemaID {
			t.Errorf("Expected schemaId %s, got %s", reqBody.ProviderSchemaID, response.SchemaID)
		}

		if response.Status != "success" {
			t.Errorf("Expected status 'success', got %s", response.Status)
		}

		if response.ConsumerID != reqBody.ConsumerID {
			t.Errorf("Expected consumerId %s, got %s", reqBody.ConsumerID, response.ConsumerID)
		}

		if response.ProviderID != reqBody.ProviderID {
			t.Errorf("Expected providerId %s, got %s", reqBody.ProviderID, response.ProviderID)
		}
	})

	t.Run("CreateDataExchangeEvent_Failure", func(t *testing.T) {
		reqBody := models.DataExchangeEvent{
			EventID:           "550e8400-e29b-41d4-a716-446655440001",
			Timestamp:         "2024-01-15T10:30:00Z",
			ActorUserID:       "user-123",
			ConsumerAppID:     "passport-app",
			ConsumerID:        "member-consumer-123",
			OnBehalfOfOwnerID: "citizen-abc",
			ProviderSchemaID:  "hospital-schema-v1",
			ProviderID:        "member-provider-456",
			RequestedFields:   []string{"personInfo.name"},
			Status:            "FAILURE",
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		req := httptest.NewRequest("POST", "/v1/audit/exchange", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.DataExchangeHandler.CreateDataExchangeEvent(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
		}

		var response models.Log
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Status != "failure" {
			t.Errorf("Expected status 'failure', got %s", response.Status)
		}
	})

	t.Run("CreateDataExchangeEvent_MissingConsumerAppID", func(t *testing.T) {
		reqBody := models.DataExchangeEvent{
			EventID:          "550e8400-e29b-41d4-a716-446655440002",
			ProviderSchemaID: "hospital-schema-v1",
			ProviderID:       "member-provider-456",
			Status:           "SUCCESS",
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		req := httptest.NewRequest("POST", "/v1/audit/exchange", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.DataExchangeHandler.CreateDataExchangeEvent(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})

	t.Run("CreateDataExchangeEvent_MissingConsumerID", func(t *testing.T) {
		reqBody := models.DataExchangeEvent{
			EventID:          "550e8400-e29b-41d4-a716-446655440006",
			ConsumerAppID:    "passport-app",
			ProviderSchemaID: "hospital-schema-v1",
			ProviderID:       "member-provider-456",
			Status:           "SUCCESS",
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		req := httptest.NewRequest("POST", "/v1/audit/exchange", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.DataExchangeHandler.CreateDataExchangeEvent(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})

	t.Run("CreateDataExchangeEvent_MissingProviderID", func(t *testing.T) {
		reqBody := models.DataExchangeEvent{
			EventID:          "550e8400-e29b-41d4-a716-446655440007",
			ConsumerAppID:    "passport-app",
			ConsumerID:       "member-consumer-123",
			ProviderSchemaID: "hospital-schema-v1",
			Status:           "SUCCESS",
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		req := httptest.NewRequest("POST", "/v1/audit/exchange", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.DataExchangeHandler.CreateDataExchangeEvent(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})

	t.Run("CreateDataExchangeEvent_MissingProviderSchemaID", func(t *testing.T) {
		reqBody := models.DataExchangeEvent{
			EventID:       "550e8400-e29b-41d4-a716-446655440003",
			ConsumerAppID: "passport-app",
			ConsumerID:    "member-consumer-123",
			ProviderID:    "member-provider-456",
			Status:        "SUCCESS",
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		req := httptest.NewRequest("POST", "/v1/audit/exchange", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.DataExchangeHandler.CreateDataExchangeEvent(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})

	t.Run("CreateDataExchangeEvent_InvalidStatus", func(t *testing.T) {
		reqBody := models.DataExchangeEvent{
			EventID:          "550e8400-e29b-41d4-a716-446655440004",
			ConsumerAppID:    "passport-app",
			ConsumerID:       "member-consumer-123",
			ProviderSchemaID: "hospital-schema-v1",
			ProviderID:       "member-provider-456",
			Status:           "INVALID",
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		req := httptest.NewRequest("POST", "/v1/audit/exchange", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.DataExchangeHandler.CreateDataExchangeEvent(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})

	t.Run("CreateDataExchangeEvent_InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/v1/audit/exchange", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.DataExchangeHandler.CreateDataExchangeEvent(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("CreateDataExchangeEvent_EmptyRequestedFields", func(t *testing.T) {
		reqBody := models.DataExchangeEvent{
			EventID:          "550e8400-e29b-41d4-a716-446655440005",
			ConsumerAppID:    "passport-app",
			ConsumerID:       "member-consumer-123",
			ProviderSchemaID: "hospital-schema-v1",
			ProviderID:       "member-provider-456",
			RequestedFields:  []string{},
			Status:           "SUCCESS",
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		req := httptest.NewRequest("POST", "/v1/audit/exchange", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.DataExchangeHandler.CreateDataExchangeEvent(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
		}
	})
}
