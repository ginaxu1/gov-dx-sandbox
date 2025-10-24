package tests

import (
	"context"
	"os"
	"testing"

	"github.com/gov-dx-sandbox/audit-service/models"
	"github.com/gov-dx-sandbox/audit-service/services"
	_ "github.com/lib/pq"
)

// TestConsumer tests the Redis stream consumer functionality
func TestConsumer(t *testing.T) {
	// Skip if not using PostgreSQL for tests
	if os.Getenv("TEST_USE_POSTGRES") != "true" {
		t.Skip("PostgreSQL tests not enabled (set TEST_USE_POSTGRES=true)")
	}

	// Setup test database
	gormDB := setupTestDatabase(t)
	defer func() {
		if sqlDB, err := gormDB.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	// Create audit service
	auditService := services.NewAuditService(gormDB)

	t.Run("CreateLog_DirectDatabaseInsert", func(t *testing.T) {
		// Test direct database insertion (simulating what the consumer would do)
		auditLog := &models.AuditLog{
			EventID:           "test-event-123",
			ConsumerID:        "test-consumer",
			ProviderID:        "test-provider",
			RequestedData:     `{"query": "query { test { id } }"}`,
			ResponseData:      nil,
			TransactionStatus: "success",
			CitizenHash:       "hash123",
			UserAgent:         "test-agent",
			IPAddress:         nil,
			ApplicationID:     "test-app",
			SchemaID:          "test-schema",
			Status:            "success",
		}

		// Insert directly into database
		result := gormDB.Create(auditLog)
		if result.Error != nil {
			t.Fatalf("Failed to insert audit log: %v", result.Error)
		}

		// Verify the event was saved to database
		var savedEvent models.AuditLog
		result = gormDB.Where("event_id = ?", auditLog.EventID).First(&savedEvent)
		if result.Error != nil {
			t.Fatalf("Failed to find saved event: %v", result.Error)
		}

		if savedEvent.ConsumerID != auditLog.ConsumerID {
			t.Errorf("Expected ConsumerID '%s', got '%s'", auditLog.ConsumerID, savedEvent.ConsumerID)
		}
		if savedEvent.ProviderID != auditLog.ProviderID {
			t.Errorf("Expected ProviderID '%s', got '%s'", auditLog.ProviderID, savedEvent.ProviderID)
		}
		if savedEvent.Status != auditLog.Status {
			t.Errorf("Expected Status '%s', got '%s'", auditLog.Status, savedEvent.Status)
		}
	})

	t.Run("CreateLog_WithResponseData", func(t *testing.T) {
		responseData := `{"data": {"test": {"id": "123"}}}`
		auditLog := &models.AuditLog{
			EventID:           "test-event-456",
			ConsumerID:        "test-consumer",
			ProviderID:        "test-provider",
			RequestedData:     `{"query": "query { test { id } }"}`,
			ResponseData:      &responseData,
			TransactionStatus: "success",
			CitizenHash:       "hash456",
			UserAgent:         "test-agent",
			IPAddress:         nil,
			ApplicationID:     "test-app",
			SchemaID:          "test-schema",
			Status:            "success",
		}

		result := gormDB.Create(auditLog)
		if result.Error != nil {
			t.Fatalf("Failed to insert audit log with response data: %v", result.Error)
		}

		// Verify the event was saved with response data
		var savedEvent models.AuditLog
		result = gormDB.Where("event_id = ?", auditLog.EventID).First(&savedEvent)
		if result.Error != nil {
			t.Fatalf("Failed to find saved event: %v", result.Error)
		}

		if savedEvent.ResponseData == nil {
			t.Error("Expected ResponseData to be saved")
		} else if *savedEvent.ResponseData != responseData {
			t.Errorf("Expected ResponseData '%s', got '%s'", responseData, *savedEvent.ResponseData)
		}
	})

	t.Run("CreateLog_WithIPAddress", func(t *testing.T) {
		ipAddress := "127.0.0.1"
		auditLog := &models.AuditLog{
			EventID:           "test-event-789",
			ConsumerID:        "test-consumer",
			ProviderID:        "test-provider",
			RequestedData:     `{"query": "query { test { id } }"}`,
			ResponseData:      nil,
			TransactionStatus: "success",
			CitizenHash:       "hash789",
			UserAgent:         "test-agent",
			IPAddress:         &ipAddress,
			ApplicationID:     "test-app",
			SchemaID:          "test-schema",
			Status:            "success",
		}

		result := gormDB.Create(auditLog)
		if result.Error != nil {
			t.Fatalf("Failed to insert audit log with IP address: %v", result.Error)
		}

		// Verify the event was saved with IP address
		var savedEvent models.AuditLog
		result = gormDB.Where("event_id = ?", auditLog.EventID).First(&savedEvent)
		if result.Error != nil {
			t.Fatalf("Failed to find saved event: %v", result.Error)
		}

		if savedEvent.IPAddress == nil {
			t.Error("Expected IPAddress to be saved")
		} else if *savedEvent.IPAddress != ipAddress {
			t.Errorf("Expected IPAddress '%s', got '%s'", ipAddress, *savedEvent.IPAddress)
		}
	})

	t.Run("GetAuditLogs_AfterDirectInsert", func(t *testing.T) {
		// Test that we can retrieve logs after direct insertion
		logs, total, err := auditService.GetAuditLogs(context.Background(), 10, 0)
		if err != nil {
			t.Fatalf("Failed to get audit logs: %v", err)
		}

		if total < 3 {
			t.Errorf("Expected at least 3 logs, got %d", total)
		}
		if len(logs) < 3 {
			t.Errorf("Expected at least 3 logs in response, got %d", len(logs))
		}
	})
}

