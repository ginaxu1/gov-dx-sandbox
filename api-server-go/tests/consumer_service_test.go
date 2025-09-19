package tests

import (
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/models"
	"github.com/gov-dx-sandbox/api-server-go/services"
)

// TestConsumerService_CRUD tests basic CRUD operations
func TestConsumerService_CRUD(t *testing.T) {
	consumerService := services.NewConsumerService()

	t.Run("CreateConsumer", func(t *testing.T) {
		req := models.CreateConsumerRequest{
			ConsumerName: "Test Consumer",
			ContactEmail: "test@example.com",
			PhoneNumber:  "1234567890",
		}

		consumer, err := consumerService.CreateConsumer(req)
		if err != nil {
			t.Fatalf("Failed to create consumer: %v", err)
		}

		if consumer.ConsumerID == "" {
			t.Error("Expected consumer ID")
		}
		if consumer.ConsumerName != req.ConsumerName {
			t.Errorf("Expected name %s, got %s", req.ConsumerName, consumer.ConsumerName)
		}
	})

	t.Run("GetConsumer", func(t *testing.T) {
		// Create consumer first
		req := models.CreateConsumerRequest{
			ConsumerName: "Test Consumer 2",
			ContactEmail: "test2@example.com",
			PhoneNumber:  "1234567891",
		}
		createdConsumer, err := consumerService.CreateConsumer(req)
		if err != nil {
			t.Fatalf("Failed to create consumer: %v", err)
		}

		// Get consumer
		consumer, err := consumerService.GetConsumer(createdConsumer.ConsumerID)
		if err != nil {
			t.Fatalf("Failed to get consumer: %v", err)
		}

		if consumer.ConsumerID != createdConsumer.ConsumerID {
			t.Errorf("Expected consumer ID %s, got %s", createdConsumer.ConsumerID, consumer.ConsumerID)
		}
	})

	t.Run("UpdateConsumer", func(t *testing.T) {
		// Create consumer first
		req := models.CreateConsumerRequest{
			ConsumerName: "Test Consumer 3",
			ContactEmail: "test3@example.com",
			PhoneNumber:  "1234567892",
		}
		createdConsumer, err := consumerService.CreateConsumer(req)
		if err != nil {
			t.Fatalf("Failed to create consumer: %v", err)
		}

		// Update consumer
		updatedName := "Updated Consumer"
		updatedEmail := "updated@example.com"
		updateReq := models.UpdateConsumerRequest{
			ConsumerName: &updatedName,
			ContactEmail: &updatedEmail,
		}
		updatedConsumer, err := consumerService.UpdateConsumer(createdConsumer.ConsumerID, updateReq)
		if err != nil {
			t.Fatalf("Failed to update consumer: %v", err)
		}

		if updatedConsumer.ConsumerName != *updateReq.ConsumerName {
			t.Errorf("Expected name %s, got %s", *updateReq.ConsumerName, updatedConsumer.ConsumerName)
		}
	})

	t.Run("DeleteConsumer", func(t *testing.T) {
		// Create consumer first
		req := models.CreateConsumerRequest{
			ConsumerName: "Test Consumer 4",
			ContactEmail: "test4@example.com",
			PhoneNumber:  "1234567893",
		}
		createdConsumer, err := consumerService.CreateConsumer(req)
		if err != nil {
			t.Fatalf("Failed to create consumer: %v", err)
		}

		// Delete consumer
		err = consumerService.DeleteConsumer(createdConsumer.ConsumerID)
		if err != nil {
			t.Fatalf("Failed to delete consumer: %v", err)
		}

		// Verify deletion
		_, err = consumerService.GetConsumer(createdConsumer.ConsumerID)
		if err == nil {
			t.Error("Expected error when getting deleted consumer")
		}
	})
}

// TestConsumerService_Applications tests consumer application management
func TestConsumerService_Applications(t *testing.T) {
	consumerService := services.NewConsumerService()

	// Create test consumer
	consumerReq := models.CreateConsumerRequest{
		ConsumerName: "Test Consumer",
		ContactEmail: "test@example.com",
		PhoneNumber:  "1234567890",
	}
	consumer, err := consumerService.CreateConsumer(consumerReq)
	if err != nil {
		t.Fatalf("Failed to create consumer: %v", err)
	}

	t.Run("CreateApplication", func(t *testing.T) {
		appReq := models.CreateConsumerAppRequest{
			ConsumerID:     consumer.ConsumerID,
			RequiredFields: map[string]bool{"person.fullName": true, "person.email": true},
		}

		app, err := consumerService.CreateConsumerApp(appReq)
		if err != nil {
			t.Fatalf("Failed to create application: %v", err)
		}

		if app.ConsumerID != consumer.ConsumerID {
			t.Errorf("Expected consumer ID %s, got %s", consumer.ConsumerID, app.ConsumerID)
		}
		if app.Status != models.StatusPending {
			t.Errorf("Expected status %s, got %s", models.StatusPending, app.Status)
		}
	})

	t.Run("UpdateApplication", func(t *testing.T) {
		// Create application first
		appReq := models.CreateConsumerAppRequest{
			ConsumerID:     consumer.ConsumerID,
			RequiredFields: map[string]bool{"person.fullName": true},
		}
		app, err := consumerService.CreateConsumerApp(appReq)
		if err != nil {
			t.Fatalf("Failed to create application: %v", err)
		}

		// Update application
		updateReq := models.UpdateConsumerAppRequest{
			Status: &[]models.ApplicationStatus{models.StatusApproved}[0],
		}
		updatedApp, err := consumerService.UpdateConsumerApp(app.SubmissionID, updateReq)
		if err != nil {
			t.Fatalf("Failed to update application: %v", err)
		}

		if updatedApp.Status != models.StatusApproved {
			t.Errorf("Expected status %s, got %s", models.StatusApproved, updatedApp.Status)
		}
	})

	t.Run("GetApplication", func(t *testing.T) {
		// Create application first
		appReq := models.CreateConsumerAppRequest{
			ConsumerID:     consumer.ConsumerID,
			RequiredFields: map[string]bool{"person.fullName": true},
		}
		createdApp, err := consumerService.CreateConsumerApp(appReq)
		if err != nil {
			t.Fatalf("Failed to create application: %v", err)
		}

		// Get application
		app, err := consumerService.GetConsumerApp(createdApp.SubmissionID)
		if err != nil {
			t.Fatalf("Failed to get application: %v", err)
		}

		if app.SubmissionID != createdApp.SubmissionID {
			t.Errorf("Expected submission ID %s, got %s", createdApp.SubmissionID, app.SubmissionID)
		}
	})

	// Note: DeleteConsumerApp method is not implemented in the service
}
