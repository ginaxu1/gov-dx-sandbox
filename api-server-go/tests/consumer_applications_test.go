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

func TestConsumerApplicationsAPI(t *testing.T) {
	apiServer := handlers.NewAPIServer()
	mux := http.NewServeMux()
	apiServer.SetupRoutes(mux)

	// First, create a consumer
	consumerReq := models.CreateConsumerRequest{
		ConsumerName: "Test Consumer",
		ContactEmail: "test@example.com",
		PhoneNumber:  "+1-555-0123",
	}

	jsonBody, _ := json.Marshal(consumerReq)
	req := httptest.NewRequest("POST", "/consumers", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create consumer: %d", w.Code)
	}

	var consumer models.Consumer
	if err := json.Unmarshal(w.Body.Bytes(), &consumer); err != nil {
		t.Fatalf("Failed to unmarshal consumer response: %v", err)
	}

	consumerID := consumer.ConsumerID
	t.Logf("Created consumer with ID: %s", consumerID)

	t.Run("Create Consumer Application", func(t *testing.T) {
		appReq := models.CreateConsumerAppRequest{
			RequiredFields: map[string]bool{
				"person.fullName":  true,
				"person.birthDate": true,
				"person.nic":       true,
			},
		}

		jsonBody, _ := json.Marshal(appReq)
		req := httptest.NewRequest("POST", "/consumer-applications/"+consumerID, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d. Response: %s", w.Code, w.Body.String())
		}

		var app models.ConsumerApp
		if err := json.Unmarshal(w.Body.Bytes(), &app); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if app.SubmissionID == "" {
			t.Error("Expected SubmissionID to be generated")
		}

		if app.ConsumerID != consumerID {
			t.Errorf("Expected ConsumerID %s, got %s", consumerID, app.ConsumerID)
		}

		if app.Status != models.StatusPending {
			t.Errorf("Expected status 'pending', got %s", app.Status)
		}

		if len(app.RequiredFields) != 3 {
			t.Errorf("Expected 3 required fields, got %d", len(app.RequiredFields))
		}
	})

	t.Run("Get Consumer Applications", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/consumer-applications/"+consumerID, nil)

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		// Should have at least one application from the previous test
		if items, ok := response["items"].([]interface{}); ok {
			if len(items) == 0 {
				t.Error("Expected at least one application")
			}
		} else {
			t.Error("Expected 'items' field in response")
		}
	})

	t.Run("Get All Consumer Applications (Admin View)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/consumer-applications", nil)

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		// Should have at least one application
		if items, ok := response["items"].([]interface{}); ok {
			if len(items) == 0 {
				t.Error("Expected at least one application")
			}
		} else {
			t.Error("Expected 'items' field in response")
		}
	})

	t.Run("Create Consumer Application - Consumer Not Found", func(t *testing.T) {
		appReq := models.CreateConsumerAppRequest{
			RequiredFields: map[string]bool{
				"person.fullName": true,
			},
		}

		jsonBody, _ := json.Marshal(appReq)
		req := httptest.NewRequest("POST", "/consumer-applications/nonexistent-consumer", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d. Response: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Get Consumer Applications - Consumer Not Found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/consumer-applications/nonexistent-consumer", nil)

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d. Response: %s", w.Code, w.Body.String())
		}
	})
}

func TestConsumerApplicationBySubmissionID(t *testing.T) {
	apiServer := handlers.NewAPIServer()
	mux := http.NewServeMux()
	apiServer.SetupRoutes(mux)

	// Create a consumer and application
	consumerReq := models.CreateConsumerRequest{
		ConsumerName: "Test Consumer",
		ContactEmail: "test@example.com",
		PhoneNumber:  "+1-555-0123",
	}

	jsonBody, _ := json.Marshal(consumerReq)
	req := httptest.NewRequest("POST", "/consumers", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var consumer models.Consumer
	json.Unmarshal(w.Body.Bytes(), &consumer)

	// Create an application
	appReq := models.CreateConsumerAppRequest{
		RequiredFields: map[string]bool{
			"person.fullName": true,
		},
	}

	jsonBody, _ = json.Marshal(appReq)
	req = httptest.NewRequest("POST", "/consumer-applications/"+consumer.ConsumerID, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var app models.ConsumerApp
	json.Unmarshal(w.Body.Bytes(), &app)
	submissionID := app.SubmissionID

	t.Run("Get Specific Consumer Application", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/consumer-applications/"+submissionID, nil)

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
		}

		var responseApp models.ConsumerApp
		if err := json.Unmarshal(w.Body.Bytes(), &responseApp); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if responseApp.SubmissionID != submissionID {
			t.Errorf("Expected SubmissionID %s, got %s", submissionID, responseApp.SubmissionID)
		}
	})

	t.Run("Update Consumer Application (Admin Approval)", func(t *testing.T) {
		updateReq := models.UpdateConsumerAppRequest{
			Status: &[]models.ApplicationStatus{models.StatusApproved}[0],
		}

		jsonBody, _ := json.Marshal(updateReq)
		req := httptest.NewRequest("PUT", "/consumer-applications/"+submissionID, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
		}

		var response models.UpdateConsumerAppResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response.Status != models.StatusApproved {
			t.Errorf("Expected status 'approved', got %s", response.Status)
		}

		if response.Credentials == nil {
			t.Error("Expected credentials to be generated for approved application")
		}
	})

	t.Run("Get Specific Consumer Application - Not Found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/consumer-applications/nonexistent-submission", nil)

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d. Response: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Update Consumer Application - Not Found", func(t *testing.T) {
		updateReq := models.UpdateConsumerAppRequest{
			Status: &[]models.ApplicationStatus{models.StatusApproved}[0],
		}

		jsonBody, _ := json.Marshal(updateReq)
		req := httptest.NewRequest("PUT", "/consumer-applications/nonexistent-submission", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d. Response: %s", w.Code, w.Body.String())
		}
	})
}

func TestConsumerAPI(t *testing.T) {
	apiServer := handlers.NewAPIServer()
	mux := http.NewServeMux()
	apiServer.SetupRoutes(mux)

	t.Run("Create Consumer", func(t *testing.T) {
		req := models.CreateConsumerRequest{
			ConsumerName: "Passport App Test",
			ContactEmail: "contact@test.com",
			PhoneNumber:  "+1-555-0789",
		}

		jsonBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/consumers", bytes.NewBuffer(jsonBody))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httpReq)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d. Response: %s", w.Code, w.Body.String())
		}

		var consumer models.Consumer
		if err := json.Unmarshal(w.Body.Bytes(), &consumer); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if consumer.ConsumerID == "" {
			t.Error("Expected ConsumerID to be generated")
		}

		if consumer.ConsumerName != req.ConsumerName {
			t.Errorf("Expected ConsumerName %s, got %s", req.ConsumerName, consumer.ConsumerName)
		}

		if consumer.ContactEmail != req.ContactEmail {
			t.Errorf("Expected ContactEmail %s, got %s", req.ContactEmail, consumer.ContactEmail)
		}

		if consumer.PhoneNumber != req.PhoneNumber {
			t.Errorf("Expected PhoneNumber %s, got %s", req.PhoneNumber, consumer.PhoneNumber)
		}

		if consumer.CreatedAt.IsZero() {
			t.Error("Expected CreatedAt to be set")
		}
	})

	t.Run("Get All Consumers", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/consumers", nil)

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		// Should have at least one consumer from the previous test
		if items, ok := response["items"].([]interface{}); ok {
			if len(items) == 0 {
				t.Error("Expected at least one consumer")
			}
		} else {
			t.Error("Expected 'items' field in response")
		}
	})

	t.Run("Get Specific Consumer", func(t *testing.T) {
		// First create a consumer
		createReq := models.CreateConsumerRequest{
			ConsumerName: "Test Consumer",
			ContactEmail: "test@example.com",
			PhoneNumber:  "+1-555-0123",
		}

		jsonBody, _ := json.Marshal(createReq)
		createHttpReq := httptest.NewRequest("POST", "/consumers", bytes.NewBuffer(jsonBody))
		createHttpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, createHttpReq)

		var consumer models.Consumer
		json.Unmarshal(w.Body.Bytes(), &consumer)

		// Now get the specific consumer
		req := httptest.NewRequest("GET", "/consumers/"+consumer.ConsumerID, nil)

		w = httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
		}

		var responseConsumer models.Consumer
		if err := json.Unmarshal(w.Body.Bytes(), &responseConsumer); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if responseConsumer.ConsumerID != consumer.ConsumerID {
			t.Errorf("Expected ConsumerID %s, got %s", consumer.ConsumerID, responseConsumer.ConsumerID)
		}
	})

	t.Run("Get Specific Consumer - Not Found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/consumers/nonexistent-consumer", nil)

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d. Response: %s", w.Code, w.Body.String())
		}
	})
}
