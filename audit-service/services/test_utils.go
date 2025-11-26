package services

import (
	"fmt"
	"log"
	"net/url"
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

// getEnvOrDefault returns the environment variable value or a default
func getEnvOrDefault(key, defaultValue string) string {
	loadEnvOnce() // Ensure .env.local is loaded
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SetupPostgresTestDB creates a PostgreSQL test database connection
// Uses environment variables for configuration (TEST_DB_*)
// Automatically loads configuration from .env.local file if available
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
// Environment variables (can be set in .env.local file):
//   - TEST_DB_HOST (optional, default: "localhost"): Database host
//   - TEST_DB_PORT (optional, default: "5432"): Database port
//   - TEST_DB_USERNAME (optional, default: "postgres"): Database username
//   - TEST_DB_PASSWORD (optional, default: "password"): Database password
//   - TEST_DB_USER (alternative to TEST_DB_USERNAME): Database username
//   - TEST_DB_PASS (alternative to TEST_DB_PASSWORD): Database password
//   - TEST_DB_DATABASE (optional, default: "audit_service_test"): Database name
//   - TEST_DB_SSLMODE (optional, default: "disable"): SSL mode
func SetupPostgresTestDB(t *testing.T) *gorm.DB {
	// Load environment variables from .env.local if available before any env var reads
	loadEnvOnce()

	host := getEnvOrDefault("TEST_DB_HOST", "localhost")
	port := getEnvOrDefault("TEST_DB_PORT", "5432")
	testDB := getEnvOrDefault("TEST_DB_DATABASE", "audit_service_test")
	sslmode := getEnvOrDefault("TEST_DB_SSLMODE", "disable")

	// Try to get credentials from environment first (now .env.local is loaded)
	envUser := os.Getenv("TEST_DB_USERNAME")
	envPassword := os.Getenv("TEST_DB_PASSWORD")

	var db *gorm.DB
	var err error

	// If both username and password are set via environment variable, use them
	if envUser != "" && envPassword != "" {
		db, err = tryConnection(host, port, envUser, envPassword, testDB, sslmode)
		if err == nil {
			t.Logf("Connected to PostgreSQL with environment credentials")
			return setupDatabase(t, db)
		}
	}

	// Try credential combinations from environment and common defaults
	// Build a map to track unique credentials and avoid duplicates
	credMap := make(map[string]bool)
	var credentials []struct {
		user string
		pass string
	}

	// Helper function to add unique credentials
	addCredential := func(user, pass string) {
		if user == "" {
			return
		}
		key := user + ":" + pass
		if !credMap[key] {
			credMap[key] = true
			credentials = append(credentials, struct{ user, pass string }{user, pass})
		}
	}

	// Add credentials in priority order:
	// 1. CI default
	addCredential("postgres", "password")

	// 2. Credentials from .env.local if specified
	localEnvUser := getEnvOrDefault("TEST_DB_USER", "")
	localEnvPass := getEnvOrDefault("TEST_DB_PASS", "")
	addCredential(localEnvUser, localEnvPass)

	// 3. postgres with no password
	addCredential("postgres", "")

	// 4. Current user with no password
	addCredential(os.Getenv("USER"), "")

	for _, cred := range credentials {
		if cred.user == "" {
			continue
		}

		// First try connecting to the test database directly
		db, err = tryConnection(host, port, cred.user, cred.pass, testDB, sslmode)
		if err == nil {
			t.Logf("Connected to PostgreSQL with user=%s", cred.user)
			return setupDatabase(t, db)
		}

		// If test database doesn't exist, try to connect to default database and create it
		if isDBNotExistError(err) {
			defaultDB := "postgres"
			if adminDB, adminErr := tryConnection(host, port, cred.user, cred.pass, defaultDB, sslmode); adminErr == nil {
				t.Logf("Connected to admin database, creating test database")
				// Create test database with properly quoted identifiers to prevent SQL injection
				createSQL := fmt.Sprintf("CREATE DATABASE %s WITH OWNER = %s",
					quoteIdentifier(testDB),
					quoteIdentifier(cred.user))
				if createErr := adminDB.Exec(createSQL).Error; createErr != nil {
					// Database might already exist, ignore error
					t.Logf("Note: Could not create test database (might already exist): %v", createErr)
				}

				// Close the admin database connection properly
				if sqlDB, err := adminDB.DB(); err == nil {
					sqlDB.Close()
				}

				// Now try connecting to the test database again
				db, err = tryConnection(host, port, cred.user, cred.pass, testDB, sslmode)
				if err == nil {
					t.Logf("Successfully created and connected to test database with user=%s", cred.user)
					return setupDatabase(t, db)
				}
			}
		}
	}

	// If we reach here, all connection attempts failed
	t.Skipf("Skipping test: could not connect to test database with any credentials: %v", err)
	return nil
}

// tryConnection attempts to connect to PostgreSQL with given credentials
func tryConnection(host, port, user, password, database, sslmode string) (*gorm.DB, error) {
	// Use url.Values to properly escape special characters in credentials
	params := url.Values{}
	params.Set("host", host)
	params.Set("port", port)
	params.Set("user", user)
	params.Set("password", password)
	params.Set("dbname", database)
	params.Set("sslmode", sslmode)

	// Build DSN with proper escaping
	dsn := params.Encode()
	// Replace & with spaces for PostgreSQL connection string format
	dsn = strings.ReplaceAll(dsn, "&", " ")

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
}
