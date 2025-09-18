package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/handlers"
	"github.com/gov-dx-sandbox/api-server-go/models"
)

// TestAuthenticationFlow tests the complete authentication flow
func TestAuthenticationFlow(t *testing.T) {
	// Setup
	apiServer := handlers.NewAPIServer()
	mux := http.NewServeMux()
	apiServer.SetupRoutes(mux)

	// Test 1: Create Consumer
	t.Run("CreateConsumer", func(t *testing.T) {
		reqBody := models.CreateConsumerRequest{
			ConsumerName: "Test Consumer Corp",
			ContactEmail: "test@example.com",
			PhoneNumber:  "+1-555-0123",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/consumers", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		if response["consumerId"] == nil {
			t.Error("Expected consumerId in response")
		}
	})

	// Test 2: Create Consumer Application
	t.Run("CreateConsumerApplication", func(t *testing.T) {
		// First create a consumer
		consumerReq := models.CreateConsumerRequest{
			ConsumerName: "Test Consumer Corp",
			ContactEmail: "test@example.com",
			PhoneNumber:  "+1-555-0123",
		}

		jsonBody, _ := json.Marshal(consumerReq)
		req := httptest.NewRequest("POST", "/consumers", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var consumerResponse map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &consumerResponse)
		consumerID := consumerResponse["consumerId"].(string)

		// Create application
		appReq := models.CreateApplicationRequest{
			ConsumerID: consumerID,
			RequiredFields: map[string]bool{
				"citizen_name": true,
				"citizen_age":  true,
			},
		}

		jsonBody, _ = json.Marshal(appReq)
		req = httptest.NewRequest("POST", "/consumer-applications/"+consumerID, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		if response["submissionId"] == nil {
			t.Error("Expected submissionId in response")
		}
	})

	// Test 3: Approve Application
	t.Run("ApproveApplication", func(t *testing.T) {
		// Create consumer and application first
		consumerID := createTestConsumer(t, mux)
		submissionID := createTestApplication(t, mux, consumerID)

		// Approve application
		approvalReq := models.UpdateApplicationRequest{
			Status: statusPtr(models.StatusApproved),
		}

		jsonBody, _ := json.Marshal(approvalReq)
		req := httptest.NewRequest("PUT", "/consumer-applications/"+submissionID, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		credentials := response["credentials"].(map[string]interface{})

		if credentials["apiKey"] == nil || credentials["apiSecret"] == nil {
			t.Error("Expected API credentials in response")
		}
	})
}

// TestTokenExchange tests token exchange functionality
func TestTokenExchange(t *testing.T) {
	apiServer := handlers.NewAPIServer()
	mux := http.NewServeMux()
	apiServer.SetupRoutes(mux)

	// Create consumer, application, and approve to get credentials
	consumerID := createTestConsumer(t, mux)
	submissionID := createTestApplication(t, mux, consumerID)
	apiKey, apiSecret := approveTestApplication(t, mux, submissionID)

	// Test valid token exchange
	t.Run("ValidTokenExchange", func(t *testing.T) {
		reqBody := models.TokenExchangeRequest{
			APIKey:    apiKey,
			APISecret: apiSecret,
			Scope:     "gov-dx-api",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/auth/exchange", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		// Should succeed or fail with configuration error (not credential error)
		if w.Code == http.StatusOK {
			var response models.TokenExchangeResponse
			json.Unmarshal(w.Body.Bytes(), &response)

			if response.AccessToken == "" {
				t.Error("Expected access token in response")
			}
		} else if w.Code == http.StatusInternalServerError {
			// Expected if Asgardeo not configured
			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			if !contains(response["error"].(string), "credential mapping system not properly configured") {
				t.Errorf("Expected configuration error, got: %v", response["error"])
			}
		} else {
			t.Errorf("Expected status 200 or 500, got %d", w.Code)
		}
	})

	// Test invalid token exchange
	t.Run("InvalidTokenExchange", func(t *testing.T) {
		reqBody := models.TokenExchangeRequest{
			APIKey:    "invalid_key",
			APISecret: "invalid_secret",
			Scope:     "gov-dx-api",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/auth/exchange", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		if !contains(response["error"].(string), "invalid credentials") {
			t.Errorf("Expected invalid credentials error, got: %v", response["error"])
		}
	})

	// Test missing fields
	t.Run("MissingFields", func(t *testing.T) {
		reqBody := map[string]string{
			"apiKey": "test_key",
			// Missing apiSecret
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/auth/exchange", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		if !contains(response["error"].(string), "apiKey and apiSecret are required") {
			t.Errorf("Expected missing field error, got: %v", response["error"])
		}
	})
}

// TestInputValidation tests input validation
func TestInputValidation(t *testing.T) {
	apiServer := handlers.NewAPIServer()
	mux := http.NewServeMux()
	apiServer.SetupRoutes(mux)

	t.Run("SuspiciousInput", func(t *testing.T) {
		reqBody := models.TokenExchangeRequest{
			APIKey:    "<script>alert('xss')</script>",
			APISecret: "test_secret",
			Scope:     "gov-dx-api",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/auth/exchange", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		// Should be rejected by authentication (401) or input validation (400)
		if w.Code != http.StatusBadRequest && w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 400 or 401 for suspicious input, got %d", w.Code)
		}
	})

	t.Run("PathTraversal", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/../etc/passwd", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		// Path traversal might result in redirect (301) or bad request (400)
		if w.Code != http.StatusBadRequest && w.Code != http.StatusMovedPermanently {
			t.Errorf("Expected status 400 or 301 for path traversal, got %d", w.Code)
		}
	})

	t.Run("InvalidContentType", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/exchange", bytes.NewBufferString("test"))
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid content type, got %d", w.Code)
		}
	})
}

// Helper functions

func createTestConsumer(t *testing.T, mux *http.ServeMux) string {
	reqBody := models.CreateConsumerRequest{
		ConsumerName: "Test Consumer Corp",
		ContactEmail: "test@example.com",
		PhoneNumber:  "+1-555-0123",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/consumers", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create consumer: %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	return response["consumerId"].(string)
}

func createTestApplication(t *testing.T, mux *http.ServeMux, consumerID string) string {
	reqBody := models.CreateApplicationRequest{
		ConsumerID: consumerID,
		RequiredFields: map[string]bool{
			"citizen_name": true,
			"citizen_age":  true,
		},
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/consumer-applications/"+consumerID, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create application: %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	return response["submissionId"].(string)
}

func approveTestApplication(t *testing.T, mux *http.ServeMux, submissionID string) (string, string) {
	reqBody := models.UpdateApplicationRequest{
		Status: statusPtr(models.StatusApproved),
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/consumer-applications/"+submissionID, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to approve application: %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	credentials := response["credentials"].(map[string]interface{})

	return credentials["apiKey"].(string), credentials["apiSecret"].(string)
}

func stringPtr(s string) *string {
	return &s
}

func statusPtr(status models.ApplicationStatus) *models.ApplicationStatus {
	return &status
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && contains(s[1:], substr)
}
