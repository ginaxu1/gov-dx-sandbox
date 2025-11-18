package services

import (
	"os"
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// getEnvOrDefault returns the environment variable value or a default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SetupPostgresTestDB creates a PostgreSQL test database connection
// Uses environment variables for configuration (TEST_DB_*)
//
// Returns nil if the database connection cannot be established or if database
// migration fails. In this case, the test is automatically skipped using t.Skipf().
//
// IMPORTANT: Callers MUST check for nil return value before using the database:
//
//	db := SetupPostgresTestDB(t)
//	if db == nil {
//	    return // test was skipped
//	}
//
// Environment variables:
//   - TEST_DB_HOST (optional, default: "localhost"): Database host
//   - TEST_DB_PORT (optional, default: "5432"): Database port
//   - TEST_DB_USERNAME (optional, default: "postgres"): Database username
//   - TEST_DB_PASSWORD (optional, default: "password"): Database password
//   - TEST_DB_DATABASE (optional, default: "api_server_test"): Database name
//   - TEST_DB_SSLMODE (optional, default: "disable"): SSL mode
//
// Exported for use in handler tests
func SetupPostgresTestDB(t *testing.T) *gorm.DB {
	host := getEnvOrDefault("TEST_DB_HOST", "localhost")
	port := getEnvOrDefault("TEST_DB_PORT", "5432")
	user := getEnvOrDefault("TEST_DB_USERNAME", "postgres")
	password := getEnvOrDefault("TEST_DB_PASSWORD", "password")
	database := getEnvOrDefault("TEST_DB_DATABASE", "api_server_test")
	sslmode := getEnvOrDefault("TEST_DB_SSLMODE", "disable")

	dsn := "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + database + " sslmode=" + sslmode

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Skipf("Skipping test: could not connect to test database: %v", err)
		return nil // IMPORTANT: Callers must check for nil before using the database
	}

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.Member{},
		&models.Application{},
		&models.ApplicationSubmission{},
		&models.Schema{},
		&models.SchemaSubmission{},
	)
	if err != nil {
		t.Skipf("Skipping test: could not migrate test database: %v", err)
		return nil // IMPORTANT: Callers must check for nil before using the database
	}

	// Clean up test data before each test
	CleanupTestData(t, db)

	return db
}

// CleanupTestData removes all test data from the database
// Exported for use in handler tests
func CleanupTestData(t *testing.T, db *gorm.DB) {
	// Delete in reverse order of dependencies
	if err := db.Exec("DELETE FROM application_submissions").Error; err != nil {
		t.Logf("Warning: failed to cleanup application_submissions: %v", err)
	}
	if err := db.Exec("DELETE FROM schema_submissions").Error; err != nil {
		t.Logf("Warning: failed to cleanup schema_submissions: %v", err)
	}
	if err := db.Exec("DELETE FROM applications").Error; err != nil {
		t.Logf("Warning: failed to cleanup applications: %v", err)
	}
	if err := db.Exec("DELETE FROM schemas").Error; err != nil {
		t.Logf("Warning: failed to cleanup schemas: %v", err)
	}
	if err := db.Exec("DELETE FROM members").Error; err != nil {
		t.Logf("Warning: failed to cleanup members: %v", err)
	}
}

// RequireTestDB is a helper function that sets up a test database and fails the test
// if the database cannot be established. This provides a cleaner API for tests that
// absolutely require a database connection.
//
// Usage:
//
//	db := RequireTestDB(t)
//	// No need to check for nil - test will fail if DB setup fails
//
// This is an alternative to SetupPostgresTestDB for tests that cannot proceed without
// a database connection.
func RequireTestDB(t *testing.T) *gorm.DB {
	db := SetupPostgresTestDB(t)
	if db == nil {
		t.Fatal("Test database setup failed - cannot proceed with test")
	}
	return db
}
