package tests

import (
	"context"
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
			ApplicationID: "test-app",
			SchemaID:      "test-schema",
		}

		log, err := server.AuditService.CreateLog(context.Background(), logReq)
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
		if log.ApplicationID != logReq.ApplicationID {
			t.Errorf("Expected applicationId '%s', got '%s'", logReq.ApplicationID, log.ApplicationID)
		}
		if log.SchemaID != logReq.SchemaID {
			t.Errorf("Expected schemaId '%s', got '%s'", logReq.SchemaID, log.SchemaID)
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
				ApplicationID: "app-1",
				SchemaID:      "schema-1",
			},
			{
				Status:        "failure",
				RequestedData: "query { test2 }",
				ApplicationID: "app-2",
				SchemaID:      "schema-2",
			},
		}

		for _, logReq := range testLogs {
			_, err := server.AuditService.CreateLog(context.Background(), logReq)
			if err != nil {
				t.Fatalf("Failed to create test log: %v", err)
			}
		}

		// Get all logs
		filter := &models.LogFilter{}
		response, err := server.AuditService.GetLogs(context.Background(), filter)
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
			ApplicationID: "test-app-filter",
			SchemaID:      "test-schema",
		}

		_, err := server.AuditService.CreateLog(context.Background(), logReq)
		if err != nil {
			t.Fatalf("Failed to create test log: %v", err)
		}

		// Get logs by consumer ID (using the test consumer from the view)
		filter := &models.LogFilter{
			ConsumerID: "test-consumer",
		}
		response, err := server.AuditService.GetLogs(context.Background(), filter)
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
			if log.ConsumerID != "test-consumer" {
				t.Errorf("Expected consumerId 'test-consumer', got '%s'", log.ConsumerID)
			}
		}
	})

	t.Run("GetLogs_ByProviderId", func(t *testing.T) {
		// Create test data
		logReq := &models.LogRequest{
			Status:        "success",
			RequestedData: "query { providerTest }",
			ApplicationID: "test-app",
			SchemaID:      "test-schema-filter",
		}

		_, err := server.AuditService.CreateLog(context.Background(), logReq)
		if err != nil {
			t.Fatalf("Failed to create test log: %v", err)
		}

		// Get logs by provider ID (using the test provider from the view)
		filter := &models.LogFilter{
			ProviderID: "test-provider",
		}
		response, err := server.AuditService.GetLogs(context.Background(), filter)
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
			if log.ProviderID != "test-provider" {
				t.Errorf("Expected providerId 'test-provider', got '%s'", log.ProviderID)
			}
		}
	})

	t.Run("GetLogs_ByStatus", func(t *testing.T) {
		// Create test data
		logReq := &models.LogRequest{
			Status:        "failure",
			RequestedData: "query { statusTest }",
			ApplicationID: "test-app",
			SchemaID:      "test-schema",
		}

		_, err := server.AuditService.CreateLog(context.Background(), logReq)
		if err != nil {
			t.Fatalf("Failed to create test log: %v", err)
		}

		// Get logs by status
		filter := &models.LogFilter{
			Status: "failure",
		}
		response, err := server.AuditService.GetLogs(context.Background(), filter)
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
				ApplicationID: "test-app",
				SchemaID:      "test-schema",
			}

			_, err := server.AuditService.CreateLog(context.Background(), logReq)
			if err != nil {
				t.Fatalf("Failed to create test log: %v", err)
			}
		}

		// Get logs with pagination
		filter := &models.LogFilter{
			Limit:  2,
			Offset: 1,
		}
		response, err := server.AuditService.GetLogs(context.Background(), filter)
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
			ApplicationID: "test-app-combined",
			SchemaID:      "test-schema-combined",
		}

		_, err := server.AuditService.CreateLog(context.Background(), logReq)
		if err != nil {
			t.Fatalf("Failed to create test log: %v", err)
		}

		// Get logs with combined filters (using test values from the view)
		filter := &models.LogFilter{
			ConsumerID: "test-consumer",
			ProviderID: "test-provider",
			Status:     "success",
		}
		response, err := server.AuditService.GetLogs(context.Background(), filter)
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
			if log.ConsumerID != "test-consumer" {
				t.Errorf("Expected consumerId 'test-consumer', got '%s'", log.ConsumerID)
			}
			if log.ProviderID != "test-provider" {
				t.Errorf("Expected providerId 'test-provider', got '%s'", log.ProviderID)
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
		response, err := server.AuditService.GetLogs(context.Background(), filter)
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
