package tests

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/gov-dx-sandbox/audit-service/handlers"
	"github.com/gov-dx-sandbox/audit-service/services"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

// TestServerWithGORM represents a test server with GORM support for management events
type TestServerWithGORM struct {
	DB                     *sql.DB
	GormDB                 *gorm.DB
	AuditService           *services.AuditService
	ManagementEventService *services.ManagementEventService
	Handler                *handlers.AuditHandler
	DataExchangeHandler    *handlers.AuditHandler
	ManagementEventHandler *handlers.ManagementEventHandler
	Context                context.Context
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
	// Drop and recreate audit_logs table to ensure correct schema
	// This ensures the table always has the latest schema including consumer_id and provider_id
	dropTableQuery := `DROP TABLE IF EXISTS audit_logs CASCADE;`
	if _, err := db.Exec(dropTableQuery); err != nil {
		return err
	}

	// Create audit_logs table for testing (matching the new schema)
	createTableQuery := `
		CREATE TABLE audit_logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			status VARCHAR(10) NOT NULL CHECK (status IN ('success', 'failure')),
			requested_data TEXT NOT NULL,
			application_id VARCHAR(255) NOT NULL,
			schema_id VARCHAR(255) NOT NULL,
			consumer_id VARCHAR(255),
			provider_id VARCHAR(255),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`

	if _, err := db.Exec(createTableQuery); err != nil {
		return err
	}

	// Create a simple view for testing (uses actual columns from audit_logs table)
	createViewQuery := `
		CREATE OR REPLACE VIEW audit_logs_with_provider_consumer AS
		SELECT id,
			   timestamp,
			   status,
			   requested_data,
			   application_id,
			   schema_id,
			   COALESCE(consumer_id, 'test-consumer') as consumer_id,
			   COALESCE(provider_id, 'test-provider') as provider_id
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

// SetupTestServerWithGORM creates a test server with GORM support for both CASE 1 and CASE 2
func SetupTestServerWithGORM(t *testing.T) *TestServerWithGORM {
	// Use test database configuration
	host := getEnvOrDefault("TEST_DB_HOST", "localhost")
	port := getEnvOrDefault("TEST_DB_PORT", "5432")
	username := getEnvOrDefault("TEST_DB_USERNAME", "postgres")
	password := getEnvOrDefault("TEST_DB_PASSWORD", "password")
	database := getEnvOrDefault("TEST_DB_DATABASE", "audit_service_test")
	sslmode := getEnvOrDefault("TEST_DB_SSLMODE", "disable")

	connectionString := "host=" + host + " port=" + port + " user=" + username +
		" password=" + password + " dbname=" + database + " sslmode=" + sslmode

	// Create SQL connection
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		t.Skipf("Skipping test: failed to connect to database: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		t.Skipf("Skipping test: failed to ping database: %v", err)
	}

	// Create GORM connection
	dsn := connectionString + " TimeZone=UTC"
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping test: failed to connect via GORM: %v", err)
	}

	// Initialize database tables
	if err := initTestDatabaseWithGORM(db, gormDB); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	// Clean up test data before each test
	cleanupTestDataWithGORM(t, db, gormDB)

	// Create services
	auditService := services.NewAuditService(db)
	managementEventService := services.NewManagementEventService(gormDB)
	handler := handlers.NewAuditHandler(auditService)
	managementEventHandler := handlers.NewManagementEventHandler(managementEventService)

	return &TestServerWithGORM{
		DB:                     db,
		GormDB:                 gormDB,
		AuditService:           auditService,
		ManagementEventService: managementEventService,
		Handler:                handler,
		DataExchangeHandler:    handler, // Uses same handler for data exchange
		ManagementEventHandler: managementEventHandler,
		Context:                context.Background(),
	}
}

// initTestDatabaseWithGORM initializes the test database with required tables for both cases
func initTestDatabaseWithGORM(db *sql.DB, gormDB *gorm.DB) error {
	// Create audit_logs table for CASE 1
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS audit_logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			status VARCHAR(10) NOT NULL CHECK (status IN ('success', 'failure')),
			requested_data TEXT NOT NULL,
			application_id VARCHAR(255) NOT NULL,
			schema_id VARCHAR(255) NOT NULL,
			consumer_id VARCHAR(255),
			provider_id VARCHAR(255),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`

	if _, err := db.Exec(createTableQuery); err != nil {
		return err
	}

	// Create management_events table for CASE 2 using GORM AutoMigrate
	if err := gormDB.Exec(`
		CREATE TABLE IF NOT EXISTS management_events (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			event_id UUID NOT NULL UNIQUE,
			event_type VARCHAR(10) NOT NULL CHECK (event_type IN ('CREATE', 'READ', 'UPDATE', 'DELETE')),
			timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			actor_type VARCHAR(10) NOT NULL CHECK (actor_type IN ('USER', 'SERVICE')),
			actor_id VARCHAR(255),
			actor_role VARCHAR(10) CHECK (actor_role IN ('MEMBER', 'ADMIN')),
			target_resource VARCHAR(50) NOT NULL CHECK (target_resource IN (
				'MEMBERS', 'SCHEMAS', 'SCHEMA-SUBMISSIONS', 
				'APPLICATIONS', 'APPLICATION-SUBMISSIONS', 'POLICY-METADATA'
			)),
			target_resource_id VARCHAR(255) NOT NULL,
			metadata JSONB,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`).Error; err != nil {
		return err
	}

	// Create indexes for management_events
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_management_events_event_id ON management_events(event_id)",
		"CREATE INDEX IF NOT EXISTS idx_management_events_timestamp ON management_events(timestamp DESC)",
		"CREATE INDEX IF NOT EXISTS idx_management_events_actor ON management_events(actor_type, actor_id)",
		"CREATE INDEX IF NOT EXISTS idx_management_events_target ON management_events(target_resource, target_resource_id)",
	}

	for _, indexSQL := range indexes {
		if err := gormDB.Exec(indexSQL).Error; err != nil {
			return err
		}
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

// cleanupTestDataWithGORM removes all test data from the database
func cleanupTestDataWithGORM(t *testing.T, db *sql.DB, gormDB *gorm.DB) {
	if _, err := db.Exec("DELETE FROM audit_logs"); err != nil {
		t.Logf("Warning: failed to cleanup audit_logs: %v", err)
	}
	if err := gormDB.Exec("DELETE FROM management_events").Error; err != nil {
		t.Logf("Warning: failed to cleanup management_events: %v", err)
	}
}

// Close closes the test server and cleans up resources
func (ts *TestServerWithGORM) Close() error {
	sqlDB, err := ts.GormDB.DB()
	if err == nil {
		sqlDB.Close()
	}
	return ts.DB.Close()
}
