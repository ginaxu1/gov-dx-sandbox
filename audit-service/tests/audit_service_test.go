package tests

import (
	"context"
	"os"
	"testing"

	"github.com/gov-dx-sandbox/audit-service/models"
	"github.com/gov-dx-sandbox/audit-service/services"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
		logReq := createTestLogRequest("success", "test-app", "test-schema")
		log, err := auditService.CreateLog(context.Background(), logReq)
		if err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}

		validateLog(t, log, logReq)
	})

	t.Run("CreateLog_Failure", func(t *testing.T) {
		logReq := createTestLogRequest("failure", "test-app", "test-schema")
		log, err := auditService.CreateLog(context.Background(), logReq)
		if err != nil {
			t.Fatalf("Failed to create failure log: %v", err)
		}

		if log.Status != "failure" {
			t.Errorf("Expected Status 'failure', got '%s'", log.Status)
		}
	})

	t.Run("GetAuditLogs_WithPagination", func(t *testing.T) {
		// Create test data
		createTestLogs(t, auditService, 5)

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

	t.Run("DirectDatabaseInsert", func(t *testing.T) {
		// Test direct database insertion (simulating Redis consumer behavior)
		auditLog := createTestAuditLog("test-event-123", "test-consumer", "test-provider")

		result := gormDB.Create(auditLog)
		if result.Error != nil {
			t.Fatalf("Failed to insert audit log: %v", result.Error)
		}

		// Verify the event was saved
		var savedEvent models.AuditLog
		result = gormDB.Where("event_id = ?", auditLog.EventID).First(&savedEvent)
		if result.Error != nil {
			t.Fatalf("Failed to find saved event: %v", result.Error)
		}

		validateAuditLog(t, &savedEvent, auditLog)
	})

	t.Run("DirectInsert_WithResponseData", func(t *testing.T) {
		responseData := `{"data": {"test": {"id": "123"}}}`
		auditLog := createTestAuditLog("test-event-456", "test-consumer", "test-provider")
		auditLog.ResponseData = &responseData

		result := gormDB.Create(auditLog)
		if result.Error != nil {
			t.Fatalf("Failed to insert audit log with response data: %v", result.Error)
		}

		// Verify response data was saved
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

	t.Run("DirectInsert_WithIPAddress", func(t *testing.T) {
		ipAddress := "127.0.0.1"
		auditLog := createTestAuditLog("test-event-789", "test-consumer", "test-provider")
		auditLog.IPAddress = &ipAddress

		result := gormDB.Create(auditLog)
		if result.Error != nil {
			t.Fatalf("Failed to insert audit log with IP address: %v", result.Error)
		}

		// Verify IP address was saved
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

// Helper functions for DRY principle

func createTestLogRequest(status, appID, schemaID string) *models.LogRequest {
	return &models.LogRequest{
		Status:        status,
		RequestedData: `{"query": "query { test { id } }"}`,
		ApplicationID: appID,
		SchemaID:      schemaID,
	}
}

func createTestAuditLog(eventID, consumerID, providerID string) *models.AuditLog {
	return &models.AuditLog{
		EventID:           eventID,
		ConsumerID:        consumerID,
		ProviderID:        providerID,
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
}

func createTestLogs(t *testing.T, auditService *services.AuditService, count int) {
	for i := 0; i < count; i++ {
		logReq := createTestLogRequest("success", "test-app", "test-schema")
		_, err := auditService.CreateLog(context.Background(), logReq)
		if err != nil {
			t.Fatalf("Failed to create test log %d: %v", i, err)
		}
	}
}

func validateLog(t *testing.T, log *models.Log, logReq *models.LogRequest) {
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
}

func validateAuditLog(t *testing.T, saved, expected *models.AuditLog) {
	if saved.ConsumerID != expected.ConsumerID {
		t.Errorf("Expected ConsumerID '%s', got '%s'", expected.ConsumerID, saved.ConsumerID)
	}
	if saved.ProviderID != expected.ProviderID {
		t.Errorf("Expected ProviderID '%s', got '%s'", expected.ProviderID, saved.ProviderID)
	}
	if saved.Status != expected.Status {
		t.Errorf("Expected Status '%s', got '%s'", expected.Status, saved.Status)
	}
}

// setupTestDatabase creates a test database connection
func setupTestDatabase(t *testing.T) *gorm.DB {
	// Use test database configuration
	host := getEnvOrDefault("TEST_DB_HOST", "localhost")
	port := getEnvOrDefault("TEST_DB_PORT", "5432")
	username := getEnvOrDefault("TEST_DB_USERNAME", "postgres")
	password := getEnvOrDefault("TEST_DB_PASSWORD", "password")
	database := getEnvOrDefault("TEST_DB_DATABASE", "audit_service_test")
	sslmode := getEnvOrDefault("TEST_DB_SSLMODE", "disable")

	dsn := "host=" + host + " port=" + port + " user=" + username +
		" password=" + password + " dbname=" + database + " sslmode=" + sslmode

	// Create GORM database connection
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping PostgreSQL test: failed to connect to database: %v", err)
	}

	// Test the connection
	sqlDB, err := gormDB.DB()
	if err != nil {
		t.Skipf("Skipping PostgreSQL test: failed to get underlying DB: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		t.Skipf("Skipping PostgreSQL test: failed to ping database: %v", err)
	}

	// Auto-migrate the schema
	err = gormDB.AutoMigrate(&models.AuditLog{})
	if err != nil {
		t.Fatalf("Failed to auto-migrate schema: %v", err)
	}

	// Clean up test data
	gormDB.Exec("DELETE FROM audit_logs")

	return gormDB
}

// getEnvOrDefault gets an environment variable with a fallback default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
