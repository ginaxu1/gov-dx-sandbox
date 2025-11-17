package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/pkg/monitoring"
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
		Host:            getEnvOrDefault("CHOREO_DB_CE_HOSTNAME", "localhost"),
		Port:            getEnvOrDefault("CHOREO_DB_CE_PORT", "5432"),
		Username:        getEnvOrDefault("CHOREO_DB_CE_USERNAME", "postgres"),
		Password:        getEnvOrDefault("CHOREO_DB_CE_PASSWORD", "password"),
		Database:        getEnvOrDefault("CHOREO_DB_CE_DATABASENAME", "consent_engine"),
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
		start := time.Now()
		err = db.PingContext(ctx)
		duration := time.Since(start)
		cancel()
		monitoring.RecordExternalCall(context.Background(), "consent-db", "connect", duration, err)

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
	// Create a timeout context for the query
	queryCtx, cancel := context.WithTimeout(ctx, config.QueryTimeout)
	defer cancel()

	start := time.Now()
	result, err := db.ExecContext(queryCtx, query, args...)
	monitoring.RecordDBLatency(ctx, "consent-db", "exec", time.Since(start))
	return result, err
}

// QueryWithTimeout executes a query with timeout and returns rows
func QueryWithTimeout(ctx context.Context, db *sql.DB, config *DatabaseConfig, query string, args ...interface{}) (*sql.Rows, error) {
	// Create a timeout context for the query
	queryCtx, cancel := context.WithTimeout(ctx, config.QueryTimeout)
	defer cancel()

	start := time.Now()
	rows, err := db.QueryContext(queryCtx, query, args...)
	monitoring.RecordDBLatency(ctx, "consent-db", "query", time.Since(start))
	return rows, err
}

// QueryRowWithTimeout executes a query with timeout and returns a single row and a cleanup function.
// The cleanup function MUST be called after Scan() completes to avoid context leaks.
// Usage:
//
//	row, cleanup := QueryRowWithTimeout(ctx, db, config, "SELECT ...")
//	defer cleanup() // Call cleanup after Scan() completes
//	err := row.Scan(&result)
func QueryRowWithTimeout(ctx context.Context, db *sql.DB, config *DatabaseConfig, query string, args ...interface{}) (*sql.Row, func()) {
	// Create a timeout context for the query
	queryCtx, cancel := context.WithTimeout(ctx, config.QueryTimeout)

	start := time.Now()
	row := db.QueryRowContext(queryCtx, query, args...)
	monitoring.RecordDBLatency(ctx, "consent-db", "queryrow", time.Since(start))
	return row, cancel
}

// InitDatabase creates the necessary tables if they don't exist
func InitDatabase(db *sql.DB) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS consent_records (
		consent_id VARCHAR(255) PRIMARY KEY,
		owner_id VARCHAR(255) NOT NULL,
		owner_email VARCHAR(255) NOT NULL,
		app_id VARCHAR(255) NOT NULL,
		status VARCHAR(50) NOT NULL,
		type VARCHAR(50) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL,
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
		grant_duration VARCHAR(50) NOT NULL,
		fields TEXT[] NOT NULL,
		session_id VARCHAR(255) NOT NULL,
		consent_portal_url TEXT,
		updated_by VARCHAR(255)
	);

	-- Create indexes for better performance
	CREATE INDEX IF NOT EXISTS idx_consent_records_owner_id ON consent_records(owner_id);
	CREATE INDEX IF NOT EXISTS idx_consent_records_owner_email ON consent_records(owner_email);
	CREATE INDEX IF NOT EXISTS idx_consent_records_app_id ON consent_records(app_id);
	CREATE INDEX IF NOT EXISTS idx_consent_records_status ON consent_records(status);
	CREATE INDEX IF NOT EXISTS idx_consent_records_created_at ON consent_records(created_at);
	CREATE INDEX IF NOT EXISTS idx_consent_records_expires_at ON consent_records(expires_at);
	
	-- Create composite index for finding existing pending consents
	CREATE INDEX IF NOT EXISTS idx_consent_records_pending_lookup ON consent_records(owner_id, owner_email, app_id, status) 
		WHERE status = 'pending';
	
	-- Create unique partial index to enforce only one pending record per (owner_id, app_id) tuple
	CREATE UNIQUE INDEX IF NOT EXISTS idx_consent_records_unique_pending ON consent_records(owner_id, app_id) 
		WHERE status = 'pending';
	`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create consent_records table: %w", err)
	}

	slog.Info("Database tables initialized successfully")
	return nil
}
