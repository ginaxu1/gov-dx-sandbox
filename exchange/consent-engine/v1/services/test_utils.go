package service

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gov-dx-sandbox/exchange/shared/utils"
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
		return SetupPostgresTestEngine(t)
	}

	// Default to PostgreSQL engine for tests
	return SetupPostgresTestEngine(t)
}

// SetupPostgresTestEngine creates a PostgreSQL test engine
func SetupPostgresTestEngine(t *testing.T) ConsentEngine {
	return SetupPostgresTestEngineWithDB(t).Engine
}

// PostgresTestEngineWithDB holds both the engine and database connection for tests
type PostgresTestEngineWithDB struct {
	Engine ConsentEngine
	DB     *sql.DB
}

// SetupPostgresTestEngineWithDB creates a PostgreSQL test engine and returns both engine and DB
func SetupPostgresTestEngineWithDB(t *testing.T) *PostgresTestEngineWithDB {
	// Use test database configuration
	config := &DatabaseConfig{
		Host:            utils.GetEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:            utils.GetEnvOrDefault("TEST_DB_PORT", "5432"),
		Username:        utils.GetEnvOrDefault("TEST_DB_USERNAME", "postgres"),
		Password:        utils.GetEnvOrDefault("TEST_DB_PASSWORD", "password"),
		Database:        utils.GetEnvOrDefault("TEST_DB_DATABASE", "consent_engine_test"),
		SSLMode:         utils.GetEnvOrDefault("TEST_DB_SSLMODE", "disable"),
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
	CleanupTestData(t, db)

	return &PostgresTestEngineWithDB{
		Engine: NewPostgresConsentEngine(db, "http://localhost:5173"),
		DB:     db,
	}
}

// CleanupTestData removes all test data from the database
func CleanupTestData(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM consent_records")
	if err != nil {
		t.Logf("Warning: failed to cleanup test data: %v", err)
	}
}

// UpdateConsentExpiry directly updates the expires_at timestamp in the database
// This is useful for tests to simulate expired consents without waiting
func UpdateConsentExpiry(t *testing.T, db *sql.DB, consentID string, expiresAt time.Time) error {
	_, err := db.Exec("UPDATE consent_records SET expires_at = $1 WHERE consent_id = $2", expiresAt, consentID)
	if err != nil {
		return fmt.Errorf("failed to update consent expiry: %w", err)
	}
	return nil
}

// TestWithPostgresEngine runs a test function with PostgreSQL engine
func TestWithPostgresEngine(t *testing.T, testName string, testFunc func(t *testing.T, engine ConsentEngine)) {
	// Only run PostgreSQL tests if explicitly enabled
	if os.Getenv("TEST_USE_POSTGRES") == "true" {
		t.Run("PostgreSQL_"+testName, func(t *testing.T) {
			engine := SetupPostgresTestEngine(t)
			testFunc(t, engine)
		})
	} else {
		// Skip test if PostgreSQL is not enabled
		t.Skip("PostgreSQL tests not enabled (set TEST_USE_POSTGRES=true)")
	}
}
