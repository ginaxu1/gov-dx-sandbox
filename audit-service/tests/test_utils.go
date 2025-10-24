package tests

import (
	"os"
	"testing"

	"github.com/gov-dx-sandbox/audit-service/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// getEnvOrDefault gets an environment variable with a fallback default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
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
