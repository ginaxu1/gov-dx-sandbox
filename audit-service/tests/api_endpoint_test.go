package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gov-dx-sandbox/audit-service/models"
)

// TestHealthEndpoint tests the GET /health endpoint
func TestHealthEndpoint(t *testing.T) {
	server := SetupTestServer(t)
	defer server.Close()

	t.Run("HealthCheck_Success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		// Simulate the health check handler
		healthHandler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Audit Service is healthy"))
		}

		healthHandler(w, req)

		// Verify response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		expectedBody := "Audit Service is healthy"
		if w.Body.String() != expectedBody {
			t.Errorf("Expected body '%s', got '%s'", expectedBody, w.Body.String())
		}
	})
}

// TestPOSTLogsEndpoint tests the POST /api/logs endpoint
func TestPOSTLogsEndpoint(t *testing.T) {
	server := SetupTestServer(t)
	defer server.Close()

	t.Run("CreateLog_Success", func(t *testing.T) {
		reqBody := models.LogRequest{
			Status:        "success",
			RequestedData: "query { personInfo(nic: \"199512345678\") { fullName } }",
			ApplicationID: "app-123",
			SchemaID:      "schema-456",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/logs", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.Handler.CreateLog(w, req)

		// Verify response
		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		var response models.Log
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Validate response fields
		if response.Status != "success" {
			t.Errorf("Expected status 'success', got '%s'", response.Status)
		}
		if response.RequestedData != reqBody.RequestedData {
			t.Errorf("Expected requestedData '%s', got '%s'", reqBody.RequestedData, response.RequestedData)
		}
		if response.ApplicationID != reqBody.ApplicationID {
			t.Errorf("Expected applicationId '%s', got '%s'", reqBody.ApplicationID, response.ApplicationID)
		}
		if response.SchemaID != reqBody.SchemaID {
			t.Errorf("Expected schemaId '%s', got '%s'", reqBody.SchemaID, response.SchemaID)
		}
		if response.ID == "" {
			t.Error("Expected ID to be generated")
		}
		if response.Timestamp.IsZero() {
			t.Error("Expected timestamp to be set")
		}
	})

	t.Run("CreateLog_Failure", func(t *testing.T) {
		reqBody := models.LogRequest{
			Status:        "failure",
			RequestedData: "query { vehicleInfo(plate: \"ABC-1234\") { model } }",
			ApplicationID: "app-789",
			SchemaID:      "schema-456",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/logs", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.Handler.CreateLog(w, req)

		// Verify response
		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		var response models.Log
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Validate response fields
		if response.Status != "failure" {
			t.Errorf("Expected status 'failure', got '%s'", response.Status)
		}
	})

	t.Run("CreateLog_InvalidStatus", func(t *testing.T) {
		reqBody := models.LogRequest{
			Status:        "invalid",
			RequestedData: "query { test }",
			ApplicationID: "app-123",
			SchemaID:      "schema-456",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/logs", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.Handler.CreateLog(w, req)

		// Verify response
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("CreateLog_MissingStatus", func(t *testing.T) {
		reqBody := models.LogRequest{
			RequestedData: "query { test }",
			ApplicationID: "app-123",
			SchemaID:      "schema-456",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/logs", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.Handler.CreateLog(w, req)

		// Verify response
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("CreateLog_MissingRequestedData", func(t *testing.T) {
		reqBody := models.LogRequest{
			Status:        "success",
			ApplicationID: "app-123",
			SchemaID:      "schema-456",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/logs", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.Handler.CreateLog(w, req)

		// Verify response
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("CreateLog_InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/logs", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.Handler.CreateLog(w, req)

		// Verify response
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

// TestGETLogsEndpoint tests the GET /api/logs endpoint
func TestGETLogsEndpoint(t *testing.T) {
	server := SetupTestServer(t)
	defer server.Close()

	// Create test data
	testLogs := []models.LogRequest{
		{
			Status:        "success",
			RequestedData: "query { personInfo(nic: \"199512345678\") { fullName } }",
			ApplicationID: "app-123",
			SchemaID:      "schema-456",
		},
		{
			Status:        "failure",
			RequestedData: "query { vehicleInfo(plate: \"ABC-1234\") { model } }",
			ApplicationID: "app-789",
			SchemaID:      "schema-456",
		},
		{
			Status:        "success",
			RequestedData: "query { citizenInfo(nic: \"199012345678\") { address } }",
			ApplicationID: "app-123",
			SchemaID:      "schema-789",
		},
	}

	// Insert test data
	for _, logReq := range testLogs {
		_, err := server.AuditService.CreateLog(context.Background(), &logReq)
		if err != nil {
			t.Fatalf("Failed to create test log: %v", err)
		}
	}

	t.Run("GetAllLogs_AdminPortal", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/logs", nil)
		w := httptest.NewRecorder()

		server.Handler.GetLogs(w, req)

		// Verify response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response models.LogResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Validate response
		if len(response.Logs) != 3 {
			t.Errorf("Expected 3 logs, got %d", len(response.Logs))
		}
		if response.Total != 3 {
			t.Errorf("Expected total 3, got %d", response.Total)
		}
		if response.Limit != 50 {
			t.Errorf("Expected limit 50, got %d", response.Limit)
		}
		if response.Offset != 0 {
			t.Errorf("Expected offset 0, got %d", response.Offset)
		}

		// Verify logs are ordered by timestamp DESC (newest first)
		if len(response.Logs) >= 2 {
			if response.Logs[0].Timestamp.Before(response.Logs[1].Timestamp) {
				t.Error("Expected logs to be ordered by timestamp DESC")
			}
		}
	})

	t.Run("GetLogs_ByConsumerId_ConsumerPortal", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/logs?consumerId=test-consumer", nil)
		w := httptest.NewRecorder()

		server.Handler.GetLogs(w, req)

		// Verify response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response models.LogResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Validate response - all test logs have test-consumer
		if len(response.Logs) != 3 {
			t.Errorf("Expected 3 logs for test-consumer, got %d", len(response.Logs))
		}
		if response.Total != 3 {
			t.Errorf("Expected total 3, got %d", response.Total)
		}

		// Verify all logs belong to test-consumer
		for _, log := range response.Logs {
			if log.ConsumerID != "test-consumer" {
				t.Errorf("Expected consumerId 'test-consumer', got '%s'", log.ConsumerID)
			}
		}
	})

	t.Run("GetLogs_ByProviderId_ProviderPortal", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/logs?providerId=test-provider", nil)
		w := httptest.NewRecorder()

		server.Handler.GetLogs(w, req)

		// Verify response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response models.LogResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Validate response - all test logs have test-provider
		if len(response.Logs) != 3 {
			t.Errorf("Expected 3 logs for test-provider, got %d", len(response.Logs))
		}
		if response.Total != 3 {
			t.Errorf("Expected total 3, got %d", response.Total)
		}

		// Verify all logs belong to test-provider
		for _, log := range response.Logs {
			if log.ProviderID != "test-provider" {
				t.Errorf("Expected providerId 'test-provider', got '%s'", log.ProviderID)
			}
		}
	})

	t.Run("GetLogs_ByStatus", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/logs?status=success", nil)
		w := httptest.NewRecorder()

		server.Handler.GetLogs(w, req)

		// Verify response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response models.LogResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Validate response
		if len(response.Logs) != 2 {
			t.Errorf("Expected 2 success logs, got %d", len(response.Logs))
		}
		if response.Total != 2 {
			t.Errorf("Expected total 2, got %d", response.Total)
		}

		// Verify all logs have success status
		for _, log := range response.Logs {
			if log.Status != "success" {
				t.Errorf("Expected status 'success', got '%s'", log.Status)
			}
		}
	})

	t.Run("GetLogs_WithPagination", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/logs?limit=2&offset=1", nil)
		w := httptest.NewRecorder()

		server.Handler.GetLogs(w, req)

		// Verify response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response models.LogResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Validate response
		if len(response.Logs) != 2 {
			t.Errorf("Expected 2 logs with limit=2, got %d", len(response.Logs))
		}
		if response.Limit != 2 {
			t.Errorf("Expected limit 2, got %d", response.Limit)
		}
		if response.Offset != 1 {
			t.Errorf("Expected offset 1, got %d", response.Offset)
		}
		if response.Total != 3 {
			t.Errorf("Expected total 3, got %d", response.Total)
		}
	})

	t.Run("GetLogs_CombinedFilters", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/logs?consumerId=consumer-123&status=success", nil)
		w := httptest.NewRecorder()

		server.Handler.GetLogs(w, req)

		// Verify response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response models.LogResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Validate response
		if len(response.Logs) != 2 {
			t.Errorf("Expected 2 logs for consumer-123 with success status, got %d", len(response.Logs))
		}
		if response.Total != 2 {
			t.Errorf("Expected total 2, got %d", response.Total)
		}

		// Verify all logs match the filters
		for _, log := range response.Logs {
			if log.ConsumerID != "consumer-123" {
				t.Errorf("Expected consumerId 'consumer-123', got '%s'", log.ConsumerID)
			}
			if log.Status != "success" {
				t.Errorf("Expected status 'success', got '%s'", log.Status)
			}
		}
	})

	t.Run("GetLogs_NoResults", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/logs?consumerId=nonexistent", nil)
		w := httptest.NewRecorder()

		server.Handler.GetLogs(w, req)

		// Verify response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response models.LogResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Validate response
		if len(response.Logs) != 0 {
			t.Errorf("Expected 0 logs for nonexistent consumer, got %d", len(response.Logs))
		}
		if response.Total != 0 {
			t.Errorf("Expected total 0, got %d", response.Total)
		}
	})
}

// TestLogsEndpointIntegration tests the full integration of the logs endpoint
func TestLogsEndpointIntegration(t *testing.T) {
	server := SetupTestServer(t)
	defer server.Close()

	t.Run("FullWorkflow_AdminPortal", func(t *testing.T) {
		// 1. Create a log entry
		createReq := models.LogRequest{
			Status:        "success",
			RequestedData: "query { integrationTest }",
			ApplicationID: "integration-app",
			SchemaID:      "integration-schema",
		}

		jsonBody, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/logs", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.Handler.CreateLog(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		var createdLog models.Log
		if err := json.Unmarshal(w.Body.Bytes(), &createdLog); err != nil {
			t.Fatalf("Failed to unmarshal created log: %v", err)
		}

		// 2. Retrieve all logs (admin portal)
		req = httptest.NewRequest("GET", "/api/logs", nil)
		w = httptest.NewRecorder()

		server.Handler.GetLogs(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response models.LogResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// 3. Verify the created log is in the response
		found := false
		for _, log := range response.Logs {
			if log.ID == createdLog.ID {
				found = true
				break
			}
		}
		if !found {
			t.Error("Created log not found in admin portal response")
		}

		// 4. Retrieve logs by consumer (consumer portal)
		req = httptest.NewRequest("GET", "/api/logs?consumerId=integration-consumer", nil)
		w = httptest.NewRecorder()

		server.Handler.GetLogs(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal consumer response: %v", err)
		}

		// 5. Verify the created log is in the consumer response
		found = false
		for _, log := range response.Logs {
			if log.ID == createdLog.ID {
				found = true
				break
			}
		}
		if !found {
			t.Error("Created log not found in consumer portal response")
		}

		// 6. Retrieve logs by provider (provider portal)
		req = httptest.NewRequest("GET", "/api/logs?providerId=integration-provider", nil)
		w = httptest.NewRecorder()

		server.Handler.GetLogs(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal provider response: %v", err)
		}

		// 7. Verify the created log is in the provider response
		found = false
		for _, log := range response.Logs {
			if log.ID == createdLog.ID {
				found = true
				break
			}
		}
		if !found {
			t.Error("Created log not found in provider portal response")
		}
	})
}
