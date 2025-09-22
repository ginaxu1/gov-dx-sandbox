package main

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

// TestEngineType represents the type of engine to use for testing
type TestEngineType int

const (
	InMemoryEngine TestEngineType = iota
	PostgresEngine
)

// SetupTestEngine creates a test engine based on environment configuration
func SetupTestEngine(t *testing.T) ConsentEngine {
	// Check if we should use PostgreSQL for testing
	usePostgres := os.Getenv("TEST_USE_POSTGRES") == "true"

	if usePostgres {
		return setupPostgresTestEngine(t)
	}

	// Default to in-memory engine for faster tests
	consentPortalURL := getEnvOrDefault("TEST_CONSENT_PORTAL_URL", "http://localhost:5173")
	return NewConsentEngine(consentPortalURL)
}

// setupPostgresTestEngine creates a PostgreSQL test engine
func setupPostgresTestEngine(t *testing.T) ConsentEngine {
	// Use test database configuration
	config := &DatabaseConfig{
		Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:     getEnvOrDefault("TEST_DB_PORT", "5432"),
		Username: getEnvOrDefault("TEST_DB_USERNAME", "postgres"),
		Password: getEnvOrDefault("TEST_DB_PASSWORD", "password"),
		Database: getEnvOrDefault("TEST_DB_DATABASE", "consent_engine_test"),
		SSLMode:  getEnvOrDefault("TEST_DB_SSLMODE", "disable"),
	}

	db, err := ConnectDB(config)
	if err != nil {
		t.Skipf("Skipping PostgreSQL test: failed to connect to database: %v", err)
	}

	// Initialize database tables
	if err := InitDatabase(db); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	// Clean up test data before each test
	cleanupTestData(t, db)

	return NewPostgresConsentEngine(db)
}

// cleanupTestData removes all test data from the database
func cleanupTestData(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM consent_records")
	if err != nil {
		t.Logf("Warning: failed to cleanup test data: %v", err)
	}
}

// TestWithBothEngines runs a test function with both in-memory and PostgreSQL engines
func TestWithBothEngines(t *testing.T, testName string, testFunc func(t *testing.T, engine ConsentEngine)) {
	t.Run("InMemory_"+testName, func(t *testing.T) {
		engine := NewConsentEngine("http://localhost:5173")
		testFunc(t, engine)
	})

	// Only run PostgreSQL tests if explicitly enabled
	if os.Getenv("TEST_USE_POSTGRES") == "true" {
		t.Run("PostgreSQL_"+testName, func(t *testing.T) {
			engine := setupPostgresTestEngine(t)
			testFunc(t, engine)
		})
	}
}
