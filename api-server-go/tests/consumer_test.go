package tests

import (
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/models"
)

func TestConsumerService_CreateConsumer(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	service := ts.APIServer.GetConsumerService()

	req := models.CreateConsumerRequest{
		ConsumerName: "Test Consumer",
		ContactEmail: "test@example.com",
		PhoneNumber:  "1234567890",
	}

	consumer, err := service.CreateConsumer(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
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
}

func TestConsumerService_GetConsumer(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	service := ts.APIServer.GetConsumerService()

	// First create a consumer
	req := models.CreateConsumerRequest{
		ConsumerName: "Test Consumer",
		ContactEmail: "test@example.com",
		PhoneNumber:  "1234567890",
	}

	createdConsumer, err := service.CreateConsumer(req)
	if err != nil {
		t.Fatalf("Failed to create consumer: %v", err)
	}

	// Now get the consumer
	consumer, err := service.GetConsumer(createdConsumer.ConsumerID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if consumer.ConsumerID != createdConsumer.ConsumerID {
		t.Errorf("Expected ConsumerID %s, got %s", createdConsumer.ConsumerID, consumer.ConsumerID)
	}
}

func TestConsumerService_GetConsumer_NotFound(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	service := ts.APIServer.GetConsumerService()

	_, err := service.GetConsumer("non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent consumer")
	}
}

func TestConsumerService_GetAllConsumers(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	service := ts.APIServer.GetConsumerService()

	// Create multiple consumers
	consumers := []models.CreateConsumerRequest{
		{ConsumerName: "Consumer 1", ContactEmail: "consumer1@example.com", PhoneNumber: "1111111111"},
		{ConsumerName: "Consumer 2", ContactEmail: "consumer2@example.com", PhoneNumber: "2222222222"},
	}

	for _, req := range consumers {
		_, err := service.CreateConsumer(req)
		if err != nil {
			t.Fatalf("Failed to create consumer: %v", err)
		}
	}

	// Get all consumers
	allConsumers, err := service.GetAllConsumers()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(allConsumers) != len(consumers) {
		t.Errorf("Expected %d consumers, got %d", len(consumers), len(allConsumers))
	}
}

func TestConsumerService_UpdateConsumer(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	service := ts.APIServer.GetConsumerService()

	// Create a consumer
	req := models.CreateConsumerRequest{
		ConsumerName: "Test Consumer",
		ContactEmail: "test@example.com",
		PhoneNumber:  "1234567890",
	}

	createdConsumer, err := service.CreateConsumer(req)
	if err != nil {
		t.Fatalf("Failed to create consumer: %v", err)
	}

	// Update the consumer
	updateReq := models.UpdateConsumerRequest{
		ConsumerName: stringPtr("Updated Consumer"),
		ContactEmail: stringPtr("updated@example.com"),
	}

	updatedConsumer, err := service.UpdateConsumer(createdConsumer.ConsumerID, updateReq)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if updatedConsumer.ConsumerName != "Updated Consumer" {
		t.Errorf("Expected ConsumerName 'Updated Consumer', got %s", updatedConsumer.ConsumerName)
	}
	if updatedConsumer.ContactEmail != "updated@example.com" {
		t.Errorf("Expected ContactEmail 'updated@example.com', got %s", updatedConsumer.ContactEmail)
	}
}

func TestConsumerService_DeleteConsumer(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	service := ts.APIServer.GetConsumerService()

	// Create a consumer
	req := models.CreateConsumerRequest{
		ConsumerName: "Test Consumer",
		ContactEmail: "test@example.com",
		PhoneNumber:  "1234567890",
	}

	createdConsumer, err := service.CreateConsumer(req)
	if err != nil {
		t.Fatalf("Failed to create consumer: %v", err)
	}

	// Delete the consumer
	err = service.DeleteConsumer(createdConsumer.ConsumerID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Try to get the deleted consumer
	_, err = service.GetConsumer(createdConsumer.ConsumerID)
	if err == nil {
		t.Error("Expected error for deleted consumer")
	}
}

func TestConsumerService_CreateConsumerApp(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	service := ts.APIServer.GetConsumerService()

	// First create a consumer
	consumerReq := models.CreateConsumerRequest{
		ConsumerName: "Test Consumer",
		ContactEmail: "test@example.com",
		PhoneNumber:  "1234567890",
	}

	consumer, err := service.CreateConsumer(consumerReq)
	if err != nil {
		t.Fatalf("Failed to create consumer: %v", err)
	}

	// Create a consumer app
	appReq := models.CreateConsumerAppRequest{
		ConsumerID: consumer.ConsumerID,
		RequiredFields: map[string]bool{
			"person.fullName": true,
			"person.nic":      true,
		},
	}

	app, err := service.CreateConsumerApp(appReq)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if app.SubmissionID == "" {
		t.Error("Expected SubmissionID to be generated")
	}
	if app.ConsumerID != consumer.ConsumerID {
		t.Errorf("Expected ConsumerID %s, got %s", consumer.ConsumerID, app.ConsumerID)
	}
}

func TestConsumerService_GetConsumerApp(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	service := ts.APIServer.GetConsumerService()

	// Create a consumer and app
	consumer, _ := service.CreateConsumer(models.CreateConsumerRequest{
		ConsumerName: "Test Consumer",
		ContactEmail: "test@example.com",
		PhoneNumber:  "1234567890",
	})

	app, _ := service.CreateConsumerApp(models.CreateConsumerAppRequest{
		ConsumerID: consumer.ConsumerID,
		RequiredFields: map[string]bool{
			"person.fullName": true,
		},
	})

	// Get the app
	retrievedApp, err := service.GetConsumerApp(app.SubmissionID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if retrievedApp.SubmissionID != app.SubmissionID {
		t.Errorf("Expected SubmissionID %s, got %s", app.SubmissionID, retrievedApp.SubmissionID)
	}
}

func TestConsumerService_GetAllConsumerApps(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	service := ts.APIServer.GetConsumerService()

	// Create consumers and apps
	consumer1, _ := service.CreateConsumer(models.CreateConsumerRequest{
		ConsumerName: "Consumer 1",
		ContactEmail: "consumer1@example.com",
		PhoneNumber:  "1111111111",
	})

	consumer2, _ := service.CreateConsumer(models.CreateConsumerRequest{
		ConsumerName: "Consumer 2",
		ContactEmail: "consumer2@example.com",
		PhoneNumber:  "2222222222",
	})

	// Create apps for both consumers
	service.CreateConsumerApp(models.CreateConsumerAppRequest{
		ConsumerID:     consumer1.ConsumerID,
		RequiredFields: map[string]bool{"person.fullName": true},
	})

	service.CreateConsumerApp(models.CreateConsumerAppRequest{
		ConsumerID:     consumer2.ConsumerID,
		RequiredFields: map[string]bool{"person.nic": true},
	})

	// Get all apps
	apps, err := service.GetAllConsumerApps()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(apps) != 2 {
		t.Errorf("Expected 2 apps, got %d", len(apps))
	}
}

func TestConsumerService_GetConsumerAppsByConsumerID(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	service := ts.APIServer.GetConsumerService()

	// Create a consumer
	consumer, _ := service.CreateConsumer(models.CreateConsumerRequest{
		ConsumerName: "Test Consumer",
		ContactEmail: "test@example.com",
		PhoneNumber:  "1234567890",
	})

	// Create multiple apps for the consumer
	service.CreateConsumerApp(models.CreateConsumerAppRequest{
		ConsumerID:     consumer.ConsumerID,
		RequiredFields: map[string]bool{"person.fullName": true},
	})

	service.CreateConsumerApp(models.CreateConsumerAppRequest{
		ConsumerID:     consumer.ConsumerID,
		RequiredFields: map[string]bool{"person.nic": true},
	})

	// Get apps for the specific consumer
	apps, err := service.GetConsumerAppsByConsumerID(consumer.ConsumerID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(apps) != 2 {
		t.Errorf("Expected 2 apps for consumer, got %d", len(apps))
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
