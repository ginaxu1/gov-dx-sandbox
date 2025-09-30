package tests

import (
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/models"
)

func TestAdminService_GetMetrics(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	adminService := ts.APIServer.GetAdminService()

	// Create some test data
	consumerService := ts.APIServer.GetConsumerService()
	providerService := ts.APIServer.GetProviderService()

	// Create consumers
	consumer1, err := consumerService.CreateConsumer(models.CreateConsumerRequest{
		ConsumerName: "Test Consumer 1",
		ContactEmail: "consumer1@example.com",
		PhoneNumber:  "1111111111",
	})
	if err != nil {
		t.Fatalf("Failed to create consumer: %v", err)
	}

	consumer2, err := consumerService.CreateConsumer(models.CreateConsumerRequest{
		ConsumerName: "Test Consumer 2",
		ContactEmail: "consumer2@example.com",
		PhoneNumber:  "2222222222",
	})
	if err != nil {
		t.Fatalf("Failed to create consumer: %v", err)
	}

	// Create consumer apps
	_, err = consumerService.CreateConsumerApp(models.CreateConsumerAppRequest{
		ConsumerID: consumer1.ConsumerID,
		RequiredFields: []models.RequiredField{
			{FieldName: "person.fullName"},
			{FieldName: "person.nic"},
		},
	})
	if err != nil {
		t.Fatalf("Failed to create consumer app: %v", err)
	}

	_, err = consumerService.CreateConsumerApp(models.CreateConsumerAppRequest{
		ConsumerID: consumer2.ConsumerID,
		RequiredFields: []models.RequiredField{
			{FieldName: "person.address"},
		},
	})
	if err != nil {
		t.Fatalf("Failed to create consumer app: %v", err)
	}

	// Create provider submission
	_, err = providerService.CreateProviderSubmission(models.CreateProviderSubmissionRequest{
		ProviderName: "Test Provider",
		ContactEmail: "provider@example.com",
		PhoneNumber:  "3333333333",
		ProviderType: "government",
	})
	if err != nil {
		t.Fatalf("Failed to create provider submission: %v", err)
	}

	// Get dashboard metrics
	metrics, err := adminService.GetMetrics()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check metrics structure
	if len(metrics) == 0 {
		t.Error("Expected metrics to have data")
	}

	// Verify we have the expected metrics
	expectedKeys := []string{"total_consumers", "total_consumer_apps", "total_provider_submissions"}
	for _, key := range expectedKeys {
		if _, exists := metrics[key]; !exists {
			t.Errorf("Expected metric key %s to exist", key)
		}
	}
}

func TestAdminService_GetStatistics(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	adminService := ts.APIServer.GetAdminService()

	// Create test data
	consumerService := ts.APIServer.GetConsumerService()
	providerService := ts.APIServer.GetProviderService()

	// Create a consumer
	consumer, err := consumerService.CreateConsumer(models.CreateConsumerRequest{
		ConsumerName: "Statistics Test Consumer",
		ContactEmail: "stats@example.com",
		PhoneNumber:  "4444444444",
	})
	if err != nil {
		t.Fatalf("Failed to create consumer: %v", err)
	}

	// Create a consumer app
	_, err = consumerService.CreateConsumerApp(models.CreateConsumerAppRequest{
		ConsumerID: consumer.ConsumerID,
		RequiredFields: []models.RequiredField{
			{FieldName: "person.fullName"},
		},
	})
	if err != nil {
		t.Fatalf("Failed to create consumer app: %v", err)
	}

	// Create a provider submission
	_, err = providerService.CreateProviderSubmission(models.CreateProviderSubmissionRequest{
		ProviderName: "Statistics Test Provider",
		ContactEmail: "stats-provider@example.com",
		PhoneNumber:  "5555555555",
		ProviderType: "private",
	})
	if err != nil {
		t.Fatalf("Failed to create provider submission: %v", err)
	}

	// Get statistics
	stats, err := adminService.GetStatistics()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check statistics structure
	if stats == nil {
		t.Error("Expected statistics data to be returned")
	}

	// Verify we have some data
	if len(stats) == 0 {
		t.Error("Expected statistics to have data")
	}
}
