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

// TestAuthHandlers tests authentication endpoints
func TestAuthHandlers(t *testing.T) {
	apiServer := handlers.NewAPIServer()
	mux := http.NewServeMux()
	apiServer.SetupRoutes(mux)

	t.Run("POST /auth/token", func(t *testing.T) {
		// Create test consumer and approved application
		consumer := createTestConsumerViaAPI(t, mux)
		app := createAndApproveTestAppViaAPI(t, mux, consumer.ConsumerID)

		// Get the application from the API server's service to get the generated credentials
		updatedApp, err := apiServer.GetConsumerService().GetConsumerApp(app.SubmissionID)
		if err != nil {
			t.Fatalf("Failed to get application: %v", err)
		}

		// Test valid authentication
		authReq := models.AuthRequest{
			ConsumerID: consumer.ConsumerID,
			Secret:     updatedApp.Credentials.APISecret,
		}

		jsonBody, _ := json.Marshal(authReq)
		req := httptest.NewRequest("POST", "/auth/token", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		// Response should contain data directly
	})

	t.Run("POST /auth/validate", func(t *testing.T) {
		validateReq := map[string]string{
			"token": "test-token",
		}

		jsonBody, _ := json.Marshal(validateReq)
		req := httptest.NewRequest("POST", "/auth/validate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		// Should return 200, 500, or 503 depending on Asgardeo configuration
		if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError && w.Code != http.StatusServiceUnavailable {
			t.Errorf("Expected status 200, 500, or 503, got %d", w.Code)
		}
	})

	t.Run("POST /auth/exchange", func(t *testing.T) {
		// Create test consumer and approved application
		consumer := createTestConsumerViaAPI(t, mux)
		app := createAndApproveTestAppViaAPI(t, mux, consumer.ConsumerID)

		// Get the application from the API server's service to get the generated credentials
		updatedApp, err := apiServer.GetConsumerService().GetConsumerApp(app.SubmissionID)
		if err != nil {
			t.Fatalf("Failed to get application: %v", err)
		}

		// Test token exchange
		exchangeReq := models.TokenExchangeRequest{
			APIKey:    updatedApp.Credentials.APIKey,
			APISecret: updatedApp.Credentials.APISecret,
		}

		jsonBody, _ := json.Marshal(exchangeReq)
		req := httptest.NewRequest("POST", "/auth/exchange", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})
}

// TestConsumerHandlers tests consumer management endpoints
func TestConsumerHandlers(t *testing.T) {
	apiServer := handlers.NewAPIServer()
	mux := http.NewServeMux()
	apiServer.SetupRoutes(mux)

	t.Run("POST /consumers", func(t *testing.T) {
		reqBody := models.CreateConsumerRequest{
			ConsumerName: "Test Consumer",
			ContactEmail: "test@example.com",
			PhoneNumber:  "1234567890",
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

		// Response should contain data directly
	})

	t.Run("GET /consumers", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/consumers", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("POST /consumer-applications/{consumerId}", func(t *testing.T) {
		// Create consumer first
		consumer := createTestConsumerViaAPI(t, mux)

		appReq := models.CreateConsumerAppRequest{
			ConsumerID:     consumer.ConsumerID,
			RequiredFields: map[string]bool{"person.fullName": true},
		}

		jsonBody, _ := json.Marshal(appReq)
		req := httptest.NewRequest("POST", "/consumer-applications/"+consumer.ConsumerID, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}
	})
}

// Health endpoints are tested in main_test.go

// Helper functions for API testing
func createTestConsumerViaAPI(t *testing.T, mux *http.ServeMux) *models.Consumer {
	reqBody := models.CreateConsumerRequest{
		ConsumerName: "Test Consumer",
		ContactEmail: "test@example.com",
		PhoneNumber:  "1234567890",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/consumers", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create consumer: status %d", w.Code)
	}

	var consumerData map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &consumerData)

	return &models.Consumer{
		ConsumerID:   consumerData["consumerId"].(string),
		ConsumerName: consumerData["consumerName"].(string),
		ContactEmail: consumerData["contactEmail"].(string),
		PhoneNumber:  consumerData["phoneNumber"].(string),
	}
}

func createAndApproveTestAppViaAPI(t *testing.T, mux *http.ServeMux, consumerID string) *models.ConsumerApp {
	// Create application
	appReq := models.CreateConsumerAppRequest{
		ConsumerID:     consumerID,
		RequiredFields: map[string]bool{"person.fullName": true},
	}

	jsonBody, _ := json.Marshal(appReq)
	req := httptest.NewRequest("POST", "/consumer-applications/"+consumerID, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create application: status %d", w.Code)
	}

	var appData map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &appData)

	submissionID := appData["submissionId"].(string)

	// Approve application
	updateReq := models.UpdateConsumerAppRequest{
		Status: &[]models.ApplicationStatus{models.StatusApproved}[0],
	}

	jsonBody, _ = json.Marshal(updateReq)
	req = httptest.NewRequest("PUT", "/consumer-applications/"+submissionID, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to approve application: status %d", w.Code)
	}

	// Return the approved application - credentials will be generated by the service
	return &models.ConsumerApp{
		SubmissionID: submissionID,
		ConsumerID:   consumerID,
		Status:       models.StatusApproved,
		Credentials:  nil, // Will be generated by the service
	}
}
