package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/models"
	"github.com/gov-dx-sandbox/api-server-go/services"
)

func TestConsumerService_CreateApplication(t *testing.T) {
	service := services.NewConsumerService()

	req := models.CreateApplicationRequest{
		RequiredFields: map[string]bool{
			"person.fullName": true,
			"person.nic":      true,
		},
	}

	app, err := service.CreateApplication(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if app.SubmissionID == "" {
		t.Error("Expected SubmissionID to be generated")
	}

	if app.Status != models.StatusPending {
		t.Errorf("Expected status %s, got %s", models.StatusPending, app.Status)
	}

	if len(app.RequiredFields) != 2 {
		t.Errorf("Expected 2 required fields, got %d", len(app.RequiredFields))
	}
}

func TestConsumerService_GetApplication(t *testing.T) {
	service := services.NewConsumerService()

	// Create an application first
	req := models.CreateApplicationRequest{
		RequiredFields: map[string]bool{"person.fullName": true},
	}
	createdApp, err := service.CreateApplication(req)
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	// Retrieve the application
	app, err := service.GetApplication(createdApp.SubmissionID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if app.SubmissionID != createdApp.SubmissionID {
		t.Errorf("Expected SubmissionID %s, got %s", createdApp.SubmissionID, app.SubmissionID)
	}
}

func TestConsumerService_GetApplication_NotFound(t *testing.T) {
	service := services.NewConsumerService()

	_, err := service.GetApplication("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent application")
	}
}

func TestConsumerService_UpdateApplication(t *testing.T) {
	service := services.NewConsumerService()

	// Create an application
	req := models.CreateApplicationRequest{
		RequiredFields: map[string]bool{"person.fullName": true},
	}
	createdApp, err := service.CreateApplication(req)
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	// Update the application
	updateReq := models.UpdateApplicationRequest{
		Status: &[]models.ApplicationStatus{models.StatusApproved}[0],
	}

	updatedApp, err := service.UpdateApplication(createdApp.SubmissionID, updateReq)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if updatedApp.Status != models.StatusApproved {
		t.Errorf("Expected status %s, got %s", models.StatusApproved, updatedApp.Status)
	}

	// Check that credentials were generated for approved application
	if updatedApp.Credentials == nil {
		t.Error("Expected credentials to be generated for approved application")
	}

	if updatedApp.Credentials.APIKey == "" || updatedApp.Credentials.APISecret == "" {
		t.Error("Expected credentials to have API key and secret")
	}
}

func TestConsumerService_DeleteApplication(t *testing.T) {
	service := services.NewConsumerService()

	// Create an application
	req := models.CreateApplicationRequest{
		RequiredFields: map[string]bool{"person.fullName": true},
	}
	createdApp, err := service.CreateApplication(req)
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	// Delete the application
	err = service.DeleteApplication(createdApp.SubmissionID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify it's deleted
	_, err = service.GetApplication(createdApp.SubmissionID)
	if err == nil {
		t.Error("Expected error for deleted application")
	}
}

func TestConsumerService_GetAllApplications(t *testing.T) {
	service := services.NewConsumerService()

	// Create multiple applications
	req1 := models.CreateApplicationRequest{
		RequiredFields: map[string]bool{"person.fullName": true},
	}
	req2 := models.CreateApplicationRequest{
		RequiredFields: map[string]bool{"person.nic": true},
	}

	_, err := service.CreateApplication(req1)
	if err != nil {
		t.Fatalf("Failed to create first application: %v", err)
	}

	_, err = service.CreateApplication(req2)
	if err != nil {
		t.Fatalf("Failed to create second application: %v", err)
	}

	// Get all applications
	apps, err := service.GetAllApplications()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(apps) != 2 {
		t.Errorf("Expected 2 applications, got %d", len(apps))
	}
}

// HTTP endpoint tests
func TestConsumerEndpoints(t *testing.T) {
	ts := NewTestServer()

	t.Run("Consumer Management", func(t *testing.T) {
		// Test GET /consumers (empty initially)
		w := ts.MakeGETRequest("/consumers")
		AssertResponseStatus(t, w, http.StatusOK)

		// Test POST /consumers
		consumerID := ts.CreateTestConsumer(t, "Test Consumer", "consumer@example.com", "1234567890")

		// Test GET /consumers/{consumerId}
		w = ts.MakeGETRequest("/consumers/" + consumerID)
		AssertResponseStatus(t, w, http.StatusOK)

		// Test PUT /consumers/{consumerId}
		updateReq := map[string]string{
			"consumerName": "Updated Consumer",
		}
		w = ts.MakePUTRequest("/consumers/"+consumerID, updateReq)
		AssertResponseStatus(t, w, http.StatusOK)

		// Test DELETE /consumers/{consumerId}
		w = ts.MakeDELETERequest("/consumers/" + consumerID)
		AssertResponseStatus(t, w, http.StatusNoContent)
	})

	t.Run("Consumer Applications", func(t *testing.T) {
		// Create a consumer first
		consumerID := ts.CreateTestConsumer(t, "Test Consumer", "consumer@example.com", "1234567890")

		// Test GET /consumer-applications (empty initially)
		w := ts.MakeGETRequest("/consumer-applications")
		AssertResponseStatus(t, w, http.StatusOK)

		// Test POST /consumer-applications
		requiredFields := map[string]bool{
			"name":  true,
			"email": true,
		}
		submissionID := ts.CreateTestConsumerApp(t, consumerID, requiredFields)

		// Test GET /consumer-applications/{submissionId}
		w = ts.MakeGETRequest("/consumer-applications/" + submissionID)
		AssertResponseStatus(t, w, http.StatusOK)

		// Test PUT /consumer-applications/{submissionId} (admin approval)
		updateReq := map[string]string{
			"status": "approved",
		}
		w = ts.MakePUTRequest("/consumer-applications/"+submissionID, updateReq)
		AssertResponseStatus(t, w, http.StatusOK)
	})

	t.Run("Error Cases", func(t *testing.T) {
		// Test 404 for non-existent consumer
		w := ts.MakeGETRequest("/consumers/non-existent")
		AssertResponseStatus(t, w, http.StatusNotFound)

		// Test 405 for unsupported method
		w = ts.MakeDELETERequest("/consumers")
		AssertResponseStatus(t, w, http.StatusMethodNotAllowed)

		// Test 400 for invalid JSON
		invalidReq := httptest.NewRequest("POST", "/consumers", nil)
		invalidReq.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		ts.Mux.ServeHTTP(w, invalidReq)
		AssertResponseStatus(t, w, http.StatusBadRequest)
	})
}
