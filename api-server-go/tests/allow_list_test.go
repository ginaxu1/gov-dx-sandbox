package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/models"
	"github.com/gov-dx-sandbox/api-server-go/services"
)

func TestAllowListManagement(t *testing.T) {
	// Create test server
	ts := NewTestServer()

	// Create test data
	consumerService := services.NewConsumerService()
	providerService := services.NewProviderService()

	// Create a consumer
	consumerReq := models.CreateConsumerRequest{
		ConsumerName: "Test Consumer",
		ContactEmail: "test@example.com",
		PhoneNumber:  "1234567890",
	}
	consumer, err := consumerService.CreateConsumer(consumerReq)
	if err != nil {
		t.Fatalf("Failed to create consumer: %v", err)
	}

	// Create a provider and approve it
	providerReq := models.CreateProviderSubmissionRequest{
		ProviderName: "Test Department",
		ContactEmail: "dept@example.com",
		PhoneNumber:  "1234567890",
		ProviderType: models.ProviderTypeGovernment,
	}
	providerSubmission, err := providerService.CreateProviderSubmission(providerReq)
	if err != nil {
		t.Fatalf("Failed to create provider submission: %v", err)
	}

	// Approve the provider submission
	status := models.SubmissionStatusApproved
	_, err = providerService.UpdateProviderSubmission(providerSubmission.SubmissionID, models.UpdateProviderSubmissionRequest{
		Status: &status,
	})
	if err != nil {
		t.Fatalf("Failed to approve provider submission: %v", err)
	}

	fieldName := "person.permanentAddress"

	t.Run("Allow List Management", func(t *testing.T) {
		// Test GET /admin/fields/{fieldName}/allow-list (empty list)
		w := ts.MakeGETRequest("/admin/fields/" + fieldName + "/allow-list")
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["count"] != float64(0) {
			t.Errorf("Expected empty allow list, got count %v", response["count"])
		}

		// Test POST /admin/fields/{fieldName}/allow-list (add consumer)
		addRequest := models.AllowListManagementRequest{
			ConsumerID:    consumer.ConsumerID,
			ExpiresAt:     1757560679,
			GrantDuration: "30d",
			Reason:        "Test consent approval",
			UpdatedBy:     "admin",
		}

		w = ts.MakePOSTRequest("/admin/fields/"+fieldName+"/allow-list", addRequest)
		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d. Response: %s", w.Code, w.Body.String())
		}

		// Test GET /admin/fields/{fieldName}/allow-list (with consumer)
		w = ts.MakeGETRequest("/admin/fields/" + fieldName + "/allow-list")
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["count"] != float64(1) {
			t.Errorf("Expected 1 consumer in allow list, got count %v", response["count"])
		}

		// Test PUT /admin/fields/{fieldName}/allow-list/{consumerId} (update consumer)
		updateRequest := models.AllowListManagementRequest{
			ConsumerID:    consumer.ConsumerID,
			ExpiresAt:     1757560679,
			GrantDuration: "60d",
			Reason:        "Extended access period",
			UpdatedBy:     "admin",
		}

		w = ts.MakePUTRequest("/admin/fields/"+fieldName+"/allow-list/"+consumer.ConsumerID, updateRequest)
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
		}

		// Test DELETE /admin/fields/{fieldName}/allow-list/{consumerId} (remove consumer)
		w = ts.MakeDELETERequest("/admin/fields/" + fieldName + "/allow-list/" + consumer.ConsumerID)
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
		}

		// Verify consumer is removed
		w = ts.MakeGETRequest("/admin/fields/" + fieldName + "/allow-list")
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["count"] != float64(0) {
			t.Errorf("Expected empty allow list after deletion, got count %v", response["count"])
		}
	})

	t.Run("Allow List Error Cases", func(t *testing.T) {
		// Test invalid field name
		w := ts.MakeGETRequest("/admin/fields//allow-list")
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for empty field name, got %d", w.Code)
		}

		// Test non-existent consumer
		addRequest := models.AllowListManagementRequest{
			ConsumerID:    "non-existent-consumer",
			ExpiresAt:     1757560679,
			GrantDuration: "30d",
			Reason:        "Test",
			UpdatedBy:     "admin",
		}

		w = ts.MakePOSTRequest("/admin/fields/"+fieldName+"/allow-list", addRequest)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for non-existent consumer, got %d", w.Code)
		}

		// Test invalid JSON
		w = ts.MakeRequest("POST", "/admin/fields/"+fieldName+"/allow-list", "invalid json")
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid JSON, got %d", w.Code)
		}

		// Test non-existent consumer for update
		updateRequest := models.AllowListManagementRequest{
			ConsumerID:    "non-existent-consumer",
			ExpiresAt:     1757560679,
			GrantDuration: "30d",
			Reason:        "Test",
			UpdatedBy:     "admin",
		}

		w = ts.MakePUTRequest("/admin/fields/"+fieldName+"/allow-list/non-existent-consumer", updateRequest)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for non-existent consumer update, got %d", w.Code)
		}

		// Test non-existent consumer for deletion
		w = ts.MakeDELETERequest("/admin/fields/" + fieldName + "/allow-list/non-existent-consumer")
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for non-existent consumer deletion, got %d", w.Code)
		}
	})
}

func TestAllowListService(t *testing.T) {
	grantsService := services.NewGrantsService()
	fieldName := "person.permanentAddress"

	// Test GetAllowListForField with empty list
	allowList, err := grantsService.GetAllowListForField(fieldName)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if allowList.Count != 0 {
		t.Errorf("Expected empty allow list, got count %d", allowList.Count)
	}

	// Test AddToAllowList
	addRequest := models.AllowListManagementRequest{
		ConsumerID:    "test-consumer",
		ExpiresAt:     1757560679,
		GrantDuration: "30d",
		Reason:        "Test consent approval",
		UpdatedBy:     "admin",
	}

	_, err = grantsService.AddConsumerToAllowList(fieldName, addRequest)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test GetAllowListForField with consumer
	allowList, err = grantsService.GetAllowListForField(fieldName)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if allowList.Count != 1 {
		t.Errorf("Expected 1 consumer in allow list, got count %d", allowList.Count)
	}

	// Test UpdateAllowListEntry
	updateRequest := models.AllowListManagementRequest{
		ConsumerID:    "test-consumer",
		ExpiresAt:     1757560679,
		GrantDuration: "60d",
		Reason:        "Extended access period",
		UpdatedBy:     "admin",
	}

	_, err = grantsService.UpdateConsumerInAllowList(fieldName, "test-consumer", updateRequest)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test RemoveFromAllowList
	_, err = grantsService.RemoveConsumerFromAllowList(fieldName, "test-consumer")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify consumer is removed
	allowList, err = grantsService.GetAllowListForField(fieldName)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if allowList.Count != 0 {
		t.Errorf("Expected empty allow list after deletion, got count %d", allowList.Count)
	}
}
