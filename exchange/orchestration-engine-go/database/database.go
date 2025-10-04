package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host            string
	Port            string
	Username        string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int           // Maximum number of open connections
	MaxIdleConns    int           // Maximum number of idle connections
	ConnMaxLifetime time.Duration // Maximum lifetime of a connection
	ConnMaxIdleTime time.Duration // Maximum idle time of a connection
	QueryTimeout    time.Duration // Timeout for individual queries
	ConnectTimeout  time.Duration // Timeout for initial connection
	RetryAttempts   int           // Number of retry attempts for connection
	RetryDelay      time.Duration // Delay between retry attempts
}

// NewDatabaseConfig creates a new database configuration from environment variables
func NewDatabaseConfig() *DatabaseConfig {
	// Parse durations from environment variables with defaults
	maxOpenConns := parseIntOrDefault("DB_MAX_OPEN_CONNS", 25)
	maxIdleConns := parseIntOrDefault("DB_MAX_IDLE_CONNS", 5)
	connMaxLifetime := parseDurationOrDefault("DB_CONN_MAX_LIFETIME", "1h")
	connMaxIdleTime := parseDurationOrDefault("DB_CONN_MAX_IDLE_TIME", "30m")
	queryTimeout := parseDurationOrDefault("DB_QUERY_TIMEOUT", "30s")
	connectTimeout := parseDurationOrDefault("DB_CONNECT_TIMEOUT", "10s")
	retryAttempts := parseIntOrDefault("DB_RETRY_ATTEMPTS", 3)
	retryDelay := parseDurationOrDefault("DB_RETRY_DELAY", "1s")

	// Use Choreo connection variables with fallback to legacy variables
	host := getEnvOrDefault("CHOREO_DB_OE_HOSTNAME", getEnvOrDefault("CHOREO_OPENDIF_DATABASE_HOSTNAME", "localhost"))
	port := getEnvOrDefault("CHOREO_DB_OE_PORT", getEnvOrDefault("CHOREO_OPENDIF_DATABASE_PORT", "5432"))
	username := getEnvOrDefault("CHOREO_DB_OE_USERNAME", getEnvOrDefault("CHOREO_OPENDIF_DATABASE_USERNAME", "postgres"))
	password := getEnvOrDefault("CHOREO_DB_OE_PASSWORD", getEnvOrDefault("CHOREO_OPENDIF_DATABASE_PASSWORD", "password"))
	database := getEnvOrDefault("CHOREO_DB_OE_DATABASENAME", getEnvOrDefault("CHOREO_OPENDIF_DATABASE_DATABASENAME", "orchestration_engine"))
	sslMode := getEnvOrDefault("DB_SSLMODE", "require")

	// Debug logging
	slog.Info("Environment variables",
		"host", host,
		"port", port,
		"username", username,
		"database", database,
		"sslMode", sslMode)

	return &DatabaseConfig{
		Host:            host,
		Port:            port,
		Username:        username,
		Password:        password,
		Database:        database,
		SSLMode:         sslMode,
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: connMaxLifetime,
		ConnMaxIdleTime: connMaxIdleTime,
		QueryTimeout:    queryTimeout,
		ConnectTimeout:  connectTimeout,
		RetryAttempts:   retryAttempts,
		RetryDelay:      retryDelay,
	}
}

// parseIntOrDefault parses an integer from environment variable or returns default
func parseIntOrDefault(key string, defaultValue int) int {
	if value := getEnvOrDefault(key, ""); value != "" {
		if parsed, err := fmt.Sscanf(value, "%d", &defaultValue); err == nil && parsed == 1 {
			return defaultValue
		}
	}
	return defaultValue
}

// parseDurationOrDefault parses a duration from environment variable or returns default
func parseDurationOrDefault(key, defaultValue string) time.Duration {
	if value := getEnvOrDefault(key, defaultValue); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	// Fallback to parsing the default value
	if parsed, err := time.ParseDuration(defaultValue); err == nil {
		return parsed
	}
	// Ultimate fallback
	return time.Hour
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ConnectDB establishes a connection to the PostgreSQL database with retry logic
func ConnectDB(config *DatabaseConfig) (*sql.DB, error) {
	// Build connection string with timeout
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode, int(config.ConnectTimeout.Seconds()))

	var db *sql.DB
	var err error

	// Retry connection attempts
	for attempt := 1; attempt <= config.RetryAttempts; attempt++ {
		slog.Info("Attempting database connection", "attempt", attempt, "max_attempts", config.RetryAttempts)

		// Open connection
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			slog.Warn("Failed to open database connection", "attempt", attempt, "error", err)
			if attempt < config.RetryAttempts {
				time.Sleep(config.RetryDelay)
				continue
			}
			return nil, fmt.Errorf("failed to open database connection after %d attempts: %w", config.RetryAttempts, err)
		}

		// Configure connection pool
		db.SetMaxOpenConns(config.MaxOpenConns)
		db.SetMaxIdleConns(config.MaxIdleConns)
		db.SetConnMaxLifetime(config.ConnMaxLifetime)
		db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

		// Test connection with timeout
		ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
		err = db.PingContext(ctx)
		cancel()

		if err != nil {
			slog.Warn("Failed to ping database", "attempt", attempt, "error", err)
			db.Close()
			if attempt < config.RetryAttempts {
				time.Sleep(config.RetryDelay)
				continue
			}
			return nil, fmt.Errorf("failed to ping database after %d attempts: %w", config.RetryAttempts, err)
		}

		// Connection successful
		slog.Info("Successfully connected to PostgreSQL database",
			"host", config.Host,
			"port", config.Port,
			"database", config.Database)

		return db, nil
	}

	return nil, fmt.Errorf("unexpected error: should not reach here")
}

// InitDatabase initializes the database tables for orchestration engine
func InitDatabase(db *sql.DB) error {
	slog.Info("Initializing database tables for orchestration engine")

	// Create entities table (if not exists from api-server)
	createEntitiesTable := `
	CREATE TABLE IF NOT EXISTS entities (
		entity_id VARCHAR(255) PRIMARY KEY,
		entity_name VARCHAR(255) NOT NULL,
		contact_email VARCHAR(255) NOT NULL,
		phone_number VARCHAR(50),
		entity_type VARCHAR(100) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	// Create unified_schemas table
	createUnifiedSchemasTable := `
	CREATE TABLE IF NOT EXISTS unified_schemas (
		id SERIAL PRIMARY KEY,
		version VARCHAR(50) NOT NULL UNIQUE,
		sdl TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		created_by VARCHAR(255) NOT NULL REFERENCES entities(entity_id) ON DELETE RESTRICT,
		status schema_status NOT NULL DEFAULT 'inactive',
		change_type version_change_type NOT NULL,
		notes TEXT,
		previous_version_id INTEGER REFERENCES unified_schemas(id) ON DELETE SET NULL
	);`

	// Create contract_tests table
	createContractTestsTable := `
	CREATE TABLE IF NOT EXISTS contract_tests (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		query TEXT NOT NULL,
		variables JSONB DEFAULT '{}',
		expected JSONB NOT NULL,
		description TEXT,
		priority INTEGER DEFAULT 0,
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		created_by VARCHAR(255) NOT NULL,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);`

	// Create ENUM types
	createSchemaStatusEnum := `
	DO $$ BEGIN
		CREATE TYPE schema_status AS ENUM ('active', 'inactive', 'deprecated');
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;`

	createVersionChangeTypeEnum := `
	DO $$ BEGIN
		CREATE TYPE version_change_type AS ENUM ('major', 'minor', 'patch');
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;`

	// Create indexes for better performance
	createIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_unified_schemas_version ON unified_schemas(version);",
		"CREATE INDEX IF NOT EXISTS idx_unified_schemas_status ON unified_schemas(status);",
		"CREATE INDEX IF NOT EXISTS idx_unified_schemas_created_by ON unified_schemas(created_by);",
		"CREATE INDEX IF NOT EXISTS idx_unified_schemas_created_at ON unified_schemas(created_at);",
		"CREATE INDEX IF NOT EXISTS idx_unified_schemas_previous_version_id ON unified_schemas(previous_version_id);",
		"CREATE INDEX IF NOT EXISTS idx_contract_tests_active ON contract_tests(is_active);",
		"CREATE INDEX IF NOT EXISTS idx_contract_tests_priority ON contract_tests(priority);",
	}

	// Execute table creation queries
	tables := []string{
		createEntitiesTable,
		createSchemaStatusEnum,
		createVersionChangeTypeEnum,
		createUnifiedSchemasTable,
		createContractTestsTable,
	}

	for _, query := range tables {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to create table/enum: %w", err)
		}
	}

	// Execute index creation queries
	for _, query := range createIndexes {
		if _, err := db.Exec(query); err != nil {
			slog.Warn("Failed to create index", "error", err, "query", query)
			// Don't fail on index creation errors, just log them
		}
	}

	slog.Info("Database tables initialized successfully")
	return nil
}

// GracefulShutdown gracefully closes the database connection
func GracefulShutdown(db *sql.DB) error {
	if db == nil {
		return nil
	}

	slog.Info("Starting database graceful shutdown")

	// Close all idle connections
	if err := db.Close(); err != nil {
		slog.Error("Error during database shutdown", "error", err)
		return fmt.Errorf("failed to close database: %w", err)
	}

	slog.Info("Database connection closed successfully")
	return nil
}
