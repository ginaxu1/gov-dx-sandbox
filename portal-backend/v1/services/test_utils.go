package services

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/gov-dx-sandbox/portal-backend/v1/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// loadEnvFile loads environment variables from .env.test if it exists
func loadEnvFile(filename string) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return // File doesn't exist, skip
	}

	file, err := os.Open(filename)
	if err != nil {
		return // Can't open file, skip
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Only set if not already set
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
	}
}

// getEnvOrDefault returns the environment variable value or a default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SetupSQLiteTestDB creates an in-memory SQLite database for testing
func SetupSQLiteTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("Failed to connect to SQLite test database: %v", err)
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
		t.Fatalf("Failed to migrate test database: %v", err)
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
func RequireTestDB(t *testing.T) *gorm.DB {
	db := SetupSQLiteTestDB(t)
	if db == nil {
		t.Fatal("Test database setup failed - cannot proceed with test")
	}
	return db
}
