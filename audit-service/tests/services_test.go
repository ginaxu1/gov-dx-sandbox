package tests

import (
	"testing"

	"github.com/gov-dx-sandbox/audit-service/models"
)

// TestAuditService tests the audit service layer
func TestAuditService(t *testing.T) {
	server := SetupTestServer(t)
	defer server.Close()

	t.Run("CreateLog_Success", func(t *testing.T) {
		logReq := &models.LogRequest{
			Status:        "success",
			RequestedData: "query { test }",
			ConsumerID:    "test-consumer",
			ProviderID:    "test-provider",
		}

		log, err := server.AuditService.CreateLog(logReq)
		if err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}

		// Validate created log
		if log.Status != logReq.Status {
			t.Errorf("Expected status '%s', got '%s'", logReq.Status, log.Status)
		}
		if log.RequestedData != logReq.RequestedData {
			t.Errorf("Expected requestedData '%s', got '%s'", logReq.RequestedData, log.RequestedData)
		}
		if log.ConsumerID != logReq.ConsumerID {
			t.Errorf("Expected consumerId '%s', got '%s'", logReq.ConsumerID, log.ConsumerID)
		}
		if log.ProviderID != logReq.ProviderID {
			t.Errorf("Expected providerId '%s', got '%s'", logReq.ProviderID, log.ProviderID)
		}
		if log.ID == "" {
			t.Error("Expected ID to be generated")
		}
		if log.Timestamp.IsZero() {
			t.Error("Expected timestamp to be set")
		}
	})

	t.Run("GetLogs_AllLogs", func(t *testing.T) {
		// Create test data
		testLogs := []*models.LogRequest{
			{
				Status:        "success",
				RequestedData: "query { test1 }",
				ConsumerID:    "consumer-1",
				ProviderID:    "provider-1",
			},
			{
				Status:        "failure",
				RequestedData: "query { test2 }",
				ConsumerID:    "consumer-2",
				ProviderID:    "provider-2",
			},
		}

		for _, logReq := range testLogs {
			_, err := server.AuditService.CreateLog(logReq)
			if err != nil {
				t.Fatalf("Failed to create test log: %v", err)
			}
		}

		// Get all logs
		filter := &models.LogFilter{}
		response, err := server.AuditService.GetLogs(filter)
		if err != nil {
			t.Fatalf("Failed to get logs: %v", err)
		}

		// Validate response
		if len(response.Logs) < 2 {
			t.Errorf("Expected at least 2 logs, got %d", len(response.Logs))
		}
		if response.Total < 2 {
			t.Errorf("Expected total at least 2, got %d", response.Total)
		}
	})

	t.Run("GetLogs_ByConsumerId", func(t *testing.T) {
		// Create test data
		logReq := &models.LogRequest{
			Status:        "success",
			RequestedData: "query { consumerTest }",
			ConsumerID:    "test-consumer-filter",
			ProviderID:    "test-provider",
		}

		_, err := server.AuditService.CreateLog(logReq)
		if err != nil {
			t.Fatalf("Failed to create test log: %v", err)
		}

		// Get logs by consumer ID
		filter := &models.LogFilter{
			ConsumerID: "test-consumer-filter",
		}
		response, err := server.AuditService.GetLogs(filter)
		if err != nil {
			t.Fatalf("Failed to get logs: %v", err)
		}

		// Validate response
		if len(response.Logs) < 1 {
			t.Errorf("Expected at least 1 log for consumer, got %d", len(response.Logs))
		}
		if response.Total < 1 {
			t.Errorf("Expected total at least 1, got %d", response.Total)
		}

		// Verify all logs belong to the specified consumer
		for _, log := range response.Logs {
			if log.ConsumerID != "test-consumer-filter" {
				t.Errorf("Expected consumerId 'test-consumer-filter', got '%s'", log.ConsumerID)
			}
		}
	})

	t.Run("GetLogs_ByProviderId", func(t *testing.T) {
		// Create test data
		logReq := &models.LogRequest{
			Status:        "success",
			RequestedData: "query { providerTest }",
			ConsumerID:    "test-consumer",
			ProviderID:    "test-provider-filter",
		}

		_, err := server.AuditService.CreateLog(logReq)
		if err != nil {
			t.Fatalf("Failed to create test log: %v", err)
		}

		// Get logs by provider ID
		filter := &models.LogFilter{
			ProviderID: "test-provider-filter",
		}
		response, err := server.AuditService.GetLogs(filter)
		if err != nil {
			t.Fatalf("Failed to get logs: %v", err)
		}

		// Validate response
		if len(response.Logs) < 1 {
			t.Errorf("Expected at least 1 log for provider, got %d", len(response.Logs))
		}
		if response.Total < 1 {
			t.Errorf("Expected total at least 1, got %d", response.Total)
		}

		// Verify all logs belong to the specified provider
		for _, log := range response.Logs {
			if log.ProviderID != "test-provider-filter" {
				t.Errorf("Expected providerId 'test-provider-filter', got '%s'", log.ProviderID)
			}
		}
	})

	t.Run("GetLogs_ByStatus", func(t *testing.T) {
		// Create test data
		logReq := &models.LogRequest{
			Status:        "failure",
			RequestedData: "query { statusTest }",
			ConsumerID:    "test-consumer",
			ProviderID:    "test-provider",
		}

		_, err := server.AuditService.CreateLog(logReq)
		if err != nil {
			t.Fatalf("Failed to create test log: %v", err)
		}

		// Get logs by status
		filter := &models.LogFilter{
			Status: "failure",
		}
		response, err := server.AuditService.GetLogs(filter)
		if err != nil {
			t.Fatalf("Failed to get logs: %v", err)
		}

		// Validate response
		if len(response.Logs) < 1 {
			t.Errorf("Expected at least 1 failure log, got %d", len(response.Logs))
		}
		if response.Total < 1 {
			t.Errorf("Expected total at least 1, got %d", response.Total)
		}

		// Verify all logs have the specified status
		for _, log := range response.Logs {
			if log.Status != "failure" {
				t.Errorf("Expected status 'failure', got '%s'", log.Status)
			}
		}
	})

	t.Run("GetLogs_WithPagination", func(t *testing.T) {
		// Create multiple test logs
		for i := 0; i < 5; i++ {
			logReq := &models.LogRequest{
				Status:        "success",
				RequestedData: "query { paginationTest }",
				ConsumerID:    "test-consumer",
				ProviderID:    "test-provider",
			}

			_, err := server.AuditService.CreateLog(logReq)
			if err != nil {
				t.Fatalf("Failed to create test log: %v", err)
			}
		}

		// Get logs with pagination
		filter := &models.LogFilter{
			Limit:  2,
			Offset: 1,
		}
		response, err := server.AuditService.GetLogs(filter)
		if err != nil {
			t.Fatalf("Failed to get logs: %v", err)
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
	})

	t.Run("GetLogs_CombinedFilters", func(t *testing.T) {
		// Create test data
		logReq := &models.LogRequest{
			Status:        "success",
			RequestedData: "query { combinedTest }",
			ConsumerID:    "test-consumer-combined",
			ProviderID:    "test-provider-combined",
		}

		_, err := server.AuditService.CreateLog(logReq)
		if err != nil {
			t.Fatalf("Failed to create test log: %v", err)
		}

		// Get logs with combined filters
		filter := &models.LogFilter{
			ConsumerID: "test-consumer-combined",
			ProviderID: "test-provider-combined",
			Status:     "success",
		}
		response, err := server.AuditService.GetLogs(filter)
		if err != nil {
			t.Fatalf("Failed to get logs: %v", err)
		}

		// Validate response
		if len(response.Logs) < 1 {
			t.Errorf("Expected at least 1 log with combined filters, got %d", len(response.Logs))
		}
		if response.Total < 1 {
			t.Errorf("Expected total at least 1, got %d", response.Total)
		}

		// Verify all logs match the combined filters
		for _, log := range response.Logs {
			if log.ConsumerID != "test-consumer-combined" {
				t.Errorf("Expected consumerId 'test-consumer-combined', got '%s'", log.ConsumerID)
			}
			if log.ProviderID != "test-provider-combined" {
				t.Errorf("Expected providerId 'test-provider-combined', got '%s'", log.ProviderID)
			}
			if log.Status != "success" {
				t.Errorf("Expected status 'success', got '%s'", log.Status)
			}
		}
	})

	t.Run("GetLogs_EmptyResult", func(t *testing.T) {
		// Get logs with filters that should return no results
		filter := &models.LogFilter{
			ConsumerID: "nonexistent-consumer",
		}
		response, err := server.AuditService.GetLogs(filter)
		if err != nil {
			t.Fatalf("Failed to get logs: %v", err)
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
