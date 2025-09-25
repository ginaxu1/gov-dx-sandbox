package main

import (
	"database/sql"
	"os"
	"sync"
	"testing"
	"time"

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

	// Default to PostgreSQL engine for tests
	return setupPostgresTestEngine(t)
}

// setupPostgresTestEngine creates a PostgreSQL test engine
func setupPostgresTestEngine(t *testing.T) ConsentEngine {
	// Use test database configuration
	config := &DatabaseConfig{
		Host:            getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:            getEnvOrDefault("TEST_DB_PORT", "5432"),
		Username:        getEnvOrDefault("TEST_DB_USERNAME", "postgres"),
		Password:        getEnvOrDefault("TEST_DB_PASSWORD", "password"),
		Database:        getEnvOrDefault("TEST_DB_DATABASE", "consent_engine_test"),
		SSLMode:         getEnvOrDefault("TEST_DB_SSLMODE", "disable"),
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
		QueryTimeout:    30 * time.Second,
		ConnectTimeout:  10 * time.Second,
		RetryAttempts:   3,
		RetryDelay:      1 * time.Second,
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

	return NewPostgresConsentEngine(db, "http://localhost:5173")
}

// cleanupTestData removes all test data from the database
func cleanupTestData(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM consent_records")
	if err != nil {
		t.Logf("Warning: failed to cleanup test data: %v", err)
	}
}

// ResetGlobalState resets all global state for testing
func ResetGlobalState() {
	// Reset SCIM client and sync.Once to allow reinitialization
	scimClient = nil
	scimOnce = sync.Once{}
}

// SetupTestWithCleanup sets up a test with proper cleanup of global state
func SetupTestWithCleanup(t *testing.T) func() {
	// Reset global state before test
	ResetGlobalState()

	// Return cleanup function
	return func() {
		// Reset global state after test
		ResetGlobalState()
	}
}

// TestWithPostgresEngine runs a test function with PostgreSQL engine
// Note: In-memory engine has been deprecated and removed from the codebase
func TestWithPostgresEngine(t *testing.T, testName string, testFunc func(t *testing.T, engine ConsentEngine)) {
	// Only run PostgreSQL tests if explicitly enabled
	if os.Getenv("TEST_USE_POSTGRES") == "true" {
		t.Run("PostgreSQL_"+testName, func(t *testing.T) {
			cleanup := SetupTestWithCleanup(t)
			defer cleanup()

			engine := setupPostgresTestEngine(t)
			testFunc(t, engine)
		})
	} else {
		// Skip test if PostgreSQL is not enabled
		t.Skip("PostgreSQL tests not enabled (set TEST_USE_POSTGRES=true)")
	}
}
