package testutils

import (
	"fmt"
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SetupPostgresTestDB creates a PostgreSQL test database connection for integration tests
// This connects to the same database that the service uses (via docker-compose)
func SetupPostgresTestDB(t *testing.T) *gorm.DB {
	// Use environment variables that match docker-compose.test.yml
	// These can be overridden for local testing
	host := getEnvOrDefault("TEST_DB_HOST", "localhost")
	port := getEnvOrDefault("TEST_DB_PORT", "5433") // docker-compose maps to 5433
	user := getEnvOrDefault("TEST_DB_USERNAME", "postgres")
	password := getEnvOrDefault("TEST_DB_PASSWORD", "password")
	database := getEnvOrDefault("TEST_DB_DATABASE", "policy_db")
	sslmode := getEnvOrDefault("TEST_DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, database, sslmode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Skipf("Skipping database verification: could not connect to test database: %v", err)
		return nil
	}

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		t.Skipf("Skipping database verification: failed to get sql.DB: %v", err)
		return nil
	}

	if err := sqlDB.Ping(); err != nil {
		t.Skipf("Skipping database verification: failed to ping database: %v", err)
		return nil
	}

	t.Logf("Connected to PostgreSQL test database: %s@%s:%s/%s", user, host, port, database)

	return db
}

// CleanupTestData removes all test data from the database
func CleanupTestData(t *testing.T, db *gorm.DB) {
	if db == nil {
		return
	}

	// Delete all policy metadata
	if err := db.Exec("DELETE FROM policy_metadata").Error; err != nil {
		t.Logf("Warning: could not cleanup policy_metadata: %v", err)
	}
}

// VerifyPolicyMetadataExists checks if a policy metadata record exists in the database
func VerifyPolicyMetadataExists(t *testing.T, db *gorm.DB, schemaID, fieldName string) bool {
	if db == nil {
		t.Skip("Database connection not available for verification")
		return false
	}

	var count int64
	// Use raw SQL to avoid importing models package
	err := db.Table("policy_metadata").
		Where("schema_id = ? AND field_name = ?", schemaID, fieldName).
		Count(&count).Error

	if err != nil {
		t.Logf("Warning: could not verify policy metadata: %v", err)
		return false
	}

	return count > 0
}

// GetPolicyMetadataCount returns the count of policy metadata records
func GetPolicyMetadataCount(t *testing.T, db *gorm.DB) int64 {
	if db == nil {
		return 0
	}

	var count int64
	// Use raw SQL to avoid importing models package
	if err := db.Table("policy_metadata").Count(&count).Error; err != nil {
		t.Logf("Warning: could not count policy metadata: %v", err)
		return 0
	}

	return count
}
