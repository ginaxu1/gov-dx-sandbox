package services

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/gov-dx-sandbox/audit-service/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var envLoadOnce sync.Once

// quoteIdentifier safely quotes a PostgreSQL identifier (database name, user name, etc.)
// to prevent SQL injection. Identifiers are double-quoted and any internal double-quotes
// are escaped by doubling them.
func quoteIdentifier(identifier string) string {
	// Replace any double-quotes in the identifier with doubled double-quotes
	escaped := strings.ReplaceAll(identifier, `"`, `""`)
	// Wrap the identifier in double-quotes
	return `"` + escaped + `"`
}

// loadEnvOnce loads environment variables from .env.local file (once)
func loadEnvOnce() {
	envLoadOnce.Do(func() {
		// Try to load .env.local file from current directory and parent directories
		envFiles := []string{
			".env.local",
			"../.env.local",
			"../../.env.local",
		}

		for _, envFile := range envFiles {
			if absPath, err := filepath.Abs(envFile); err == nil {
				if _, err := os.Stat(absPath); err == nil {
					if err := godotenv.Load(absPath); err == nil {
						log.Printf("Loaded test environment from: %s", absPath)
						return
					}
				}
			}
		}
		// If no .env.local found, that's okay - we'll use system env vars
	})
}

// getEnvVar returns the environment variable value
func getEnvVar(key string) string {
	loadEnvOnce() // Ensure .env.local is loaded
	return os.Getenv(key)
}

// SetupPostgresTestDB creates a PostgreSQL test database connection
// Uses environment variables for configuration (TEST_DB_*)
// Automatically loads configuration from .env.local file
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
// Required environment variables (must be set in .env.local file):
//   - TEST_DB_HOST: Database host
//   - TEST_DB_PORT: Database port
//   - TEST_DB_USERNAME: Database username
//   - TEST_DB_PASSWORD: Database password
//   - TEST_DB_DATABASE: Database name
//   - TEST_DB_SSLMODE: SSL mode
func SetupPostgresTestDB(t *testing.T) *gorm.DB {
	// Load environment variables from .env.local
	loadEnvOnce()

	host := getEnvVar("TEST_DB_HOST")
	port := getEnvVar("TEST_DB_PORT")
	testDB := getEnvVar("TEST_DB_DATABASE")
	sslmode := getEnvVar("TEST_DB_SSLMODE")
	user := getEnvVar("TEST_DB_USERNAME")
	password := getEnvVar("TEST_DB_PASSWORD")

	// Validate required environment variables
	if host == "" || port == "" || testDB == "" || sslmode == "" || user == "" || password == "" {
		t.Skipf("Skipping test: missing required environment variables. Please check .env.local file.")
		return nil
	}

	// Try connecting to the test database directly
	db, err := tryConnection(host, port, user, password, testDB, sslmode)
	if err == nil {
		t.Logf("Connected to PostgreSQL test database")
		return setupDatabase(t, db)
	}

	// If test database doesn't exist, try to connect to default database and create it
	if isDBNotExistError(err) {
		defaultDB := "postgres"
		if adminDB, adminErr := tryConnection(host, port, user, password, defaultDB, sslmode); adminErr == nil {
			t.Logf("Connected to admin database, creating test database")
			// Create test database with properly quoted identifiers to prevent SQL injection
			createSQL := fmt.Sprintf("CREATE DATABASE %s WITH OWNER = %s",
				quoteIdentifier(testDB),
				quoteIdentifier(user))
			if createErr := adminDB.Exec(createSQL).Error; createErr != nil {
				// Database might already exist, ignore error
				t.Logf("Note: Could not create test database (might already exist): %v", createErr)
			}

			// Close the admin database connection properly
			if sqlDB, err := adminDB.DB(); err == nil {
				sqlDB.Close()
			}

			// Now try connecting to the test database again
			db, err = tryConnection(host, port, user, password, testDB, sslmode)
			if err == nil {
				t.Logf("Successfully created and connected to test database")
				return setupDatabase(t, db)
			}
		}
	}

	// Connection failed
	t.Skipf("Skipping test: could not connect to test database: %v", err)
	return nil
}

// tryConnection attempts to connect to PostgreSQL with given credentials
func tryConnection(host, port, user, password, database, sslmode string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, database, sslmode)

	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
}

// isDBNotExistError checks if the error is due to database not existing
func isDBNotExistError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "3D000")
}

// setupDatabase performs migration and cleanup for the test database
func setupDatabase(t *testing.T, db *gorm.DB) *gorm.DB {
	// Auto-migrate all models
	err := db.AutoMigrate(
		&models.ManagementEvent{},
		&models.DataExchangeEvent{},
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
func CleanupTestData(t *testing.T, db *gorm.DB) {
	if db == nil {
		return
	}

	// Delete all management events
	if err := db.Exec("DELETE FROM management_events").Error; err != nil {
		t.Logf("Warning: could not cleanup management_events: %v", err)
	}

	// Delete all data exchange events
	if err := db.Exec("DELETE FROM data_exchange_events").Error; err != nil {
		t.Logf("Warning: could not cleanup data_exchange_events: %v", err)
	}
}
