package tests

import (
	"database/sql"
	"os"
	"testing"

	"github.com/gov-dx-sandbox/audit-service/handlers"
	"github.com/gov-dx-sandbox/audit-service/services"
	_ "github.com/lib/pq"
)

// TestEngineType represents the type of engine to use for testing
type TestEngineType int

const (
	InMemoryEngine TestEngineType = iota
	PostgresEngine
)

// TestServer represents a test server with all dependencies
type TestServer struct {
	DB           *sql.DB
	AuditService *services.AuditService
	Handler      *handlers.AuditHandler
}

// SetupTestServer creates a test server using PostgreSQL (only option currently supported)
func SetupTestServer(t *testing.T) *TestServer {
	// Check if we should use PostgreSQL for testing
	usePostgres := os.Getenv("TEST_USE_POSTGRES") == "true"

	if usePostgres {
		return setupPostgresTestServer(t)
	}

	// Default to PostgreSQL server for tests
	return setupPostgresTestServer(t)
}

// setupPostgresTestServer creates a PostgreSQL test server
func setupPostgresTestServer(t *testing.T) *TestServer {
	// Use test database configuration
	host := getEnvOrDefault("TEST_DB_HOST", "localhost")
	port := getEnvOrDefault("TEST_DB_PORT", "5432")
	username := getEnvOrDefault("TEST_DB_USERNAME", "postgres")
	password := getEnvOrDefault("TEST_DB_PASSWORD", "password")
	database := getEnvOrDefault("TEST_DB_DATABASE", "audit_service_test")
	sslmode := getEnvOrDefault("TEST_DB_SSLMODE", "disable")

	connectionString := "host=" + host + " port=" + port + " user=" + username +
		" password=" + password + " dbname=" + database + " sslmode=" + sslmode

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		t.Skipf("Skipping PostgreSQL test: failed to connect to database: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		t.Skipf("Skipping PostgreSQL test: failed to ping database: %v", err)
	}

	// Initialize database tables
	if err := initTestDatabase(db); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	// Clean up test data before each test
	cleanupTestData(t, db)

	// Create services
	auditService := services.NewAuditService(db)
	handler := handlers.NewAuditHandler(auditService)

	return &TestServer{
		DB:           db,
		AuditService: auditService,
		Handler:      handler,
	}
}

// initTestDatabase initializes the test database with required tables
func initTestDatabase(db *sql.DB) error {
	// Create audit_logs table for testing (matching the new schema)
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS audit_logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			status VARCHAR(10) NOT NULL CHECK (status IN ('success', 'failure')),
			requested_data TEXT NOT NULL,
			application_id VARCHAR(255) NOT NULL,
			schema_id VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`

	if _, err := db.Exec(createTableQuery); err != nil {
		return err
	}

	// Create a simple view for testing (without joins since related tables may not exist)
	createViewQuery := `
		CREATE OR REPLACE VIEW audit_logs_with_provider_consumer AS
		SELECT id,
			   timestamp,
			   status,
			   requested_data,
			   application_id,
			   schema_id,
			   'test-consumer' as consumer_id,
			   'test-provider' as provider_id
		FROM audit_logs;
	`

	_, err := db.Exec(createViewQuery)
	return err
}

// cleanupTestData removes all test data from the database
func cleanupTestData(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM audit_logs")
	if err != nil {
		t.Logf("Warning: failed to cleanup test data: %v", err)
	}
}

// TestWithPostgresServer runs a test function with PostgreSQL server
func TestWithPostgresServer(t *testing.T, testName string, testFunc func(t *testing.T, server *TestServer)) {
	// Only run PostgreSQL tests if explicitly enabled
	if os.Getenv("TEST_USE_POSTGRES") == "true" {
		t.Run("PostgreSQL_"+testName, func(t *testing.T) {
			server := setupPostgresTestServer(t)
			testFunc(t, server)
		})
	} else {
		// Skip test if PostgreSQL is not enabled
		t.Skip("PostgreSQL tests not enabled (set TEST_USE_POSTGRES=true)")
	}
}

// getEnvOrDefault gets an environment variable with a fallback default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Close closes the test server and cleans up resources
func (ts *TestServer) Close() error {
	return ts.DB.Close()
}
