package tests

import (
	"context"
	"os"
	"testing"

	"github.com/gov-dx-sandbox/audit-service/models"
	"github.com/gov-dx-sandbox/audit-service/services"
	_ "github.com/lib/pq"
)

// TestAuditService tests the GORM-based audit service
func TestAuditService(t *testing.T) {
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

	t.Run("CreateLog_Success", func(t *testing.T) {
		logReq := &models.LogRequest{
			Status:        "success",
			RequestedData: `{"query": "query { test { id } }"}`,
			ApplicationID: "test-app",
			SchemaID:      "test-schema",
		}

		log, err := auditService.CreateLog(context.Background(), logReq)
		if err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}

		// Validate created log
		if log.Status != logReq.Status {
			t.Errorf("Expected Status '%s', got '%s'", logReq.Status, log.Status)
		}
		if log.RequestedData != logReq.RequestedData {
			t.Errorf("Expected RequestedData '%s', got '%s'", logReq.RequestedData, log.RequestedData)
		}
		if log.ApplicationID != logReq.ApplicationID {
			t.Errorf("Expected ApplicationID '%s', got '%s'", logReq.ApplicationID, log.ApplicationID)
		}
		if log.SchemaID != logReq.SchemaID {
			t.Errorf("Expected SchemaID '%s', got '%s'", logReq.SchemaID, log.SchemaID)
		}
		if log.ID == "" {
			t.Error("Expected ID to be generated")
		}
		if log.CreatedAt.IsZero() {
			t.Error("Expected CreatedAt to be set")
		}
	})

	t.Run("GetAuditLogs_WithPagination", func(t *testing.T) {
		// Create test data
		for i := 0; i < 5; i++ {
			logReq := &models.LogRequest{
				Status:        "success",
				RequestedData: `{"query": "query { test { id } }"}`,
				ApplicationID: "test-app",
				SchemaID:      "test-schema",
			}
			_, err := auditService.CreateLog(context.Background(), logReq)
			if err != nil {
				t.Fatalf("Failed to create test log %d: %v", i, err)
			}
		}

		// Test pagination
		logs, total, err := auditService.GetAuditLogs(context.Background(), 3, 0)
		if err != nil {
			t.Fatalf("Failed to get audit logs: %v", err)
		}

		if len(logs) != 3 {
			t.Errorf("Expected 3 logs, got %d", len(logs))
		}
		if total < 5 {
			t.Errorf("Expected total >= 5, got %d", total)
		}
	})

	t.Run("GetAuditLogs_WithOffset", func(t *testing.T) {
		// Test offset pagination
		logs, total, err := auditService.GetAuditLogs(context.Background(), 2, 2)
		if err != nil {
			t.Fatalf("Failed to get audit logs with offset: %v", err)
		}

		if len(logs) > 2 {
			t.Errorf("Expected at most 2 logs, got %d", len(logs))
		}
		if total < 5 {
			t.Errorf("Expected total >= 5, got %d", total)
		}
	})

	t.Run("CreateLog_WithMinimalData", func(t *testing.T) {
		logReq := &models.LogRequest{
			Status:        "success",
			RequestedData: "minimal data",
			ApplicationID: "minimal-app",
			SchemaID:      "minimal-schema",
		}

		log, err := auditService.CreateLog(context.Background(), logReq)
		if err != nil {
			t.Fatalf("Failed to create minimal log: %v", err)
		}

		if log.Status != logReq.Status {
			t.Errorf("Expected Status '%s', got '%s'", logReq.Status, log.Status)
		}
		if log.RequestedData != logReq.RequestedData {
			t.Errorf("Expected RequestedData '%s', got '%s'", logReq.RequestedData, log.RequestedData)
		}
		if log.ApplicationID != logReq.ApplicationID {
			t.Errorf("Expected ApplicationID '%s', got '%s'", logReq.ApplicationID, log.ApplicationID)
		}
		if log.SchemaID != logReq.SchemaID {
			t.Errorf("Expected SchemaID '%s', got '%s'", logReq.SchemaID, log.SchemaID)
		}
	})

	t.Run("GetAuditLogs_EmptyResult", func(t *testing.T) {
		// Clean up all data
		gormDB.Exec("DELETE FROM audit_logs")

		logs, total, err := auditService.GetAuditLogs(context.Background(), 10, 0)
		if err != nil {
			t.Fatalf("Failed to get empty audit logs: %v", err)
		}

		if len(logs) != 0 {
			t.Errorf("Expected 0 logs, got %d", len(logs))
		}
		if total != 0 {
			t.Errorf("Expected total 0, got %d", total)
		}
	})
}
