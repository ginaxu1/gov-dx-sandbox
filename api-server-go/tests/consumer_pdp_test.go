package tests

import (
	"testing"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/models"
)

func TestConsumerAppPDPUpdate(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	service := ts.APIServer.GetConsumerService()

	// Create a consumer
	consumer, err := service.CreateConsumer(models.CreateConsumerRequest{
		ConsumerName: "Test Consumer",
		ContactEmail: "test@example.com",
		PhoneNumber:  "1234567890",
	})
	if err != nil {
		t.Fatalf("Failed to create consumer: %v", err)
	}

	// Create a consumer app with required fields
	app, err := service.CreateConsumerApp(models.CreateConsumerAppRequest{
		ConsumerID: consumer.ConsumerID,
		RequiredFields: []models.RequiredField{
			{FieldName: "person.fullName", GrantDuration: "P30D"},
			{FieldName: "person.nic", GrantDuration: "P60D"},
		},
	})
	if err != nil {
		t.Fatalf("Failed to create consumer app: %v", err)
	}

	// Verify required fields are stored
	if len(app.RequiredFields) != 2 {
		t.Errorf("Expected 2 required fields, got %d", len(app.RequiredFields))
	}

	// Update the app status to approved
	updateReq := models.UpdateConsumerAppRequest{
		Status: &[]models.ApplicationStatus{models.StatusApproved}[0],
	}

	// Note: This will trigger the PDP update logic
	// The actual PDP update will fail in tests since there's no real PDP server,
	// but we can verify the required fields are present in the app
	response, err := service.UpdateConsumerApp(app.SubmissionID, updateReq)
	if err != nil {
		t.Fatalf("Failed to update consumer app: %v", err)
	}

	// Verify the app was updated
	if response.ConsumerApp.Status != models.StatusApproved {
		t.Errorf("Expected status to be approved, got %s", response.ConsumerApp.Status)
	}

	// Verify required fields are still present after update
	if len(response.ConsumerApp.RequiredFields) != 2 {
		t.Errorf("Expected 2 required fields after update, got %d", len(response.ConsumerApp.RequiredFields))
	}

	// Verify the specific fields
	fieldNames := make([]string, len(response.ConsumerApp.RequiredFields))
	for i, field := range response.ConsumerApp.RequiredFields {
		fieldNames[i] = field.FieldName
	}

	expectedFields := []string{"person.fullName", "person.nic"}
	for _, expected := range expectedFields {
		found := false
		for _, actual := range fieldNames {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected field %s not found in required fields", expected)
		}
	}

	t.Logf("Consumer app updated successfully with required fields: %v", fieldNames)
}

func TestConsumerAppRequiredFieldsPersistence(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	service := ts.APIServer.GetConsumerService()

	// Create a consumer
	consumer, err := service.CreateConsumer(models.CreateConsumerRequest{
		ConsumerName: "Test Consumer",
		ContactEmail: "test@example.com",
		PhoneNumber:  "1234567890",
	})
	if err != nil {
		t.Fatalf("Failed to create consumer: %v", err)
	}

	// Create a consumer app with required fields
	app, err := service.CreateConsumerApp(models.CreateConsumerAppRequest{
		ConsumerID: consumer.ConsumerID,
		RequiredFields: []models.RequiredField{
			{FieldName: "person.fullName", GrantDuration: "P30D"},
			{FieldName: "person.nic", GrantDuration: "P60D"},
			{FieldName: "person.address", GrantDuration: "P90D"},
		},
	})
	if err != nil {
		t.Fatalf("Failed to create consumer app: %v", err)
	}

	// Wait a moment to ensure database operations complete
	time.Sleep(100 * time.Millisecond)

	// Retrieve the app from the database
	retrievedApp, err := service.GetConsumerApp(app.SubmissionID)
	if err != nil {
		t.Fatalf("Failed to retrieve consumer app: %v", err)
	}

	// Verify all required fields are present
	if len(retrievedApp.RequiredFields) != 3 {
		t.Errorf("Expected 3 required fields, got %d", len(retrievedApp.RequiredFields))
	}

	// Verify each field
	expectedFields := map[string]string{
		"person.fullName": "P30D",
		"person.nic":      "P60D",
		"person.address":  "P90D",
	}

	for _, field := range retrievedApp.RequiredFields {
		expectedDuration, exists := expectedFields[field.FieldName]
		if !exists {
			t.Errorf("Unexpected field: %s", field.FieldName)
		}
		if field.GrantDuration != expectedDuration {
			t.Errorf("Expected grant duration %s for field %s, got %s", expectedDuration, field.FieldName, field.GrantDuration)
		}
	}

	t.Logf("Required fields persisted correctly: %+v", retrievedApp.RequiredFields)
}
