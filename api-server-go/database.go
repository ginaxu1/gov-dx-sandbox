package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strconv"
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

	return &DatabaseConfig{
		Host:            getEnvOrDefault("CHOREO_OPENDIF_DB_HOSTNAME", "localhost"),
		Port:            getEnvOrDefault("CHOREO_OPENDIF_DB_PORT", "5432"),
		Username:        getEnvOrDefault("CHOREO_OPENDIF_DB_USERNAME", "postgres"),
		Password:        getEnvOrDefault("CHOREO_OPENDIF_DB_PASSWORD", "password"),
		Database:        getEnvOrDefault("CHOREO_OPENDIF_DB_DATABASENAME", "consent_engine"),
		SSLMode:         getEnvOrDefault("DB_SSLMODE", "require"),
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
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
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
	return time.Hour // Ultimate fallback
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ConnectDB establishes a connection to the PostgreSQL database
func ConnectDB(config *DatabaseConfig) (*sql.DB, error) {
	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)

	slog.Info("Connecting to PostgreSQL database", "host", config.Host, "port", config.Port, "database", config.Database)

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
			"database", config.Database,
			"max_open_conns", config.MaxOpenConns,
			"max_idle_conns", config.MaxIdleConns,
			"conn_max_lifetime", config.ConnMaxLifetime,
			"conn_max_idle_time", config.ConnMaxIdleTime)

		return db, nil
	}

	return nil, fmt.Errorf("unexpected error: should not reach here")
}

// ExecuteWithTimeout executes a query with timeout using the provided context
func ExecuteWithTimeout(ctx context.Context, db *sql.DB, config *DatabaseConfig, query string, args ...interface{}) (sql.Result, error) {
	// Create a context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, config.QueryTimeout)
	defer cancel()

	return db.ExecContext(timeoutCtx, query, args...)
}

// QueryWithTimeout executes a query with timeout using the provided context
func QueryWithTimeout(ctx context.Context, db *sql.DB, config *DatabaseConfig, query string, args ...interface{}) (*sql.Rows, error) {
	// Create a context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, config.QueryTimeout)
	defer cancel()

	return db.QueryContext(timeoutCtx, query, args...)
}

// QueryRowWithTimeout executes a query with timeout using the provided context
func QueryRowWithTimeout(ctx context.Context, db *sql.DB, config *DatabaseConfig, query string, args ...interface{}) *sql.Row {
	// Create a context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, config.QueryTimeout)
	defer cancel()

	return db.QueryRowContext(timeoutCtx, query, args...)
}

// InitDatabase initializes the database tables for api-server-go
func InitDatabase(db *sql.DB) error {
	slog.Info("Initializing database tables for api-server-go")

	// Create consumers table
	createConsumersTable := `
	CREATE TABLE IF NOT EXISTS consumers (
		consumer_id VARCHAR(255) PRIMARY KEY,
		consumer_name VARCHAR(255) NOT NULL,
		contact_email VARCHAR(255) NOT NULL,
		phone_number VARCHAR(50),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	// Create consumer_apps table
	createConsumerAppsTable := `
	CREATE TABLE IF NOT EXISTS consumer_apps (
		submission_id VARCHAR(255) PRIMARY KEY,
		consumer_id VARCHAR(255) NOT NULL REFERENCES consumers(consumer_id) ON DELETE CASCADE,
		status VARCHAR(50) NOT NULL DEFAULT 'pending',
		required_fields JSONB,
		credentials JSONB,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	// Create provider_submissions table
	createProviderSubmissionsTable := `
	CREATE TABLE IF NOT EXISTS provider_submissions (
		submission_id VARCHAR(255) PRIMARY KEY,
		provider_name VARCHAR(255) NOT NULL,
		contact_email VARCHAR(255) NOT NULL,
		phone_number VARCHAR(50) NOT NULL,
		provider_type VARCHAR(100) NOT NULL,
		status VARCHAR(50) NOT NULL DEFAULT 'pending',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	// Create provider_profiles table
	createProviderProfilesTable := `
	CREATE TABLE IF NOT EXISTS provider_profiles (
		provider_id VARCHAR(255) PRIMARY KEY,
		provider_name VARCHAR(255) NOT NULL,
		contact_email VARCHAR(255) NOT NULL,
		phone_number VARCHAR(50) NOT NULL,
		provider_type VARCHAR(100) NOT NULL,
		approved_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	// Create provider_schemas table
	createProviderSchemasTable := `
	CREATE TABLE IF NOT EXISTS provider_schemas (
		submission_id VARCHAR(255) PRIMARY KEY,
		provider_id VARCHAR(255) NOT NULL REFERENCES provider_profiles(provider_id) ON DELETE CASCADE,
		schema_id VARCHAR(255),
		status VARCHAR(50) NOT NULL DEFAULT 'pending',
		schema_input JSONB,
		sdl TEXT,
		field_configurations JSONB,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	// Create consumer_grants table
	createConsumerGrantsTable := `
	CREATE TABLE IF NOT EXISTS consumer_grants (
		consumer_id VARCHAR(255) PRIMARY KEY,
		approved_fields JSONB NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	// Create provider_metadata table
	createProviderMetadataTable := `
	CREATE TABLE IF NOT EXISTS provider_metadata (
		field_name VARCHAR(255) PRIMARY KEY,
		owner VARCHAR(255) NOT NULL,
		provider VARCHAR(255) NOT NULL,
		consent_required BOOLEAN NOT NULL DEFAULT false,
		access_control_type VARCHAR(100) NOT NULL DEFAULT 'public',
		allow_list JSONB,
		description TEXT,
		expiry_time VARCHAR(50),
		metadata JSONB,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	// Create indexes for better performance
	createIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_consumer_apps_consumer_id ON consumer_apps(consumer_id);",
		"CREATE INDEX IF NOT EXISTS idx_consumer_apps_status ON consumer_apps(status);",
		"CREATE INDEX IF NOT EXISTS idx_provider_submissions_status ON provider_submissions(status);",
		"CREATE INDEX IF NOT EXISTS idx_provider_schemas_provider_id ON provider_schemas(provider_id);",
		"CREATE INDEX IF NOT EXISTS idx_provider_schemas_status ON provider_schemas(status);",
		"CREATE INDEX IF NOT EXISTS idx_provider_metadata_owner ON provider_metadata(owner);",
		"CREATE INDEX IF NOT EXISTS idx_provider_metadata_provider ON provider_metadata(provider);",
	}

	// Execute table creation queries
	tables := []string{
		createConsumersTable,
		createConsumerAppsTable,
		createProviderSubmissionsTable,
		createProviderProfilesTable,
		createProviderSchemasTable,
		createConsumerGrantsTable,
		createProviderMetadataTable,
	}

	for _, query := range tables {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
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
