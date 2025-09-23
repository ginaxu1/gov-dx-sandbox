package database

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
	Host                string
	Port                string
	Username            string
	Password            string
	Database            string
	SSLMode             string
	MaxOpenConns        int           // Maximum number of open connections
	MaxIdleConns        int           // Maximum number of idle connections
	ConnMaxLifetime     time.Duration // Maximum lifetime of a connection
	ConnMaxIdleTime     time.Duration // Maximum idle time of a connection
	QueryTimeout        time.Duration // Timeout for individual queries
	ConnectTimeout      time.Duration // Timeout for initial connection
	RetryAttempts       int           // Number of retry attempts for connection
	RetryDelay          time.Duration // Delay between retry attempts
	TransactionTimeout  time.Duration // Timeout for transactions
	EnableMonitoring    bool          // Enable connection pool monitoring
	HealthCheckInterval time.Duration // Interval for health checks
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
	transactionTimeout := parseDurationOrDefault("DB_TRANSACTION_TIMEOUT", "60s")
	healthCheckInterval := parseDurationOrDefault("DB_HEALTH_CHECK_INTERVAL", "30s")
	enableMonitoring := getEnvOrDefault("DB_ENABLE_MONITORING", "true") == "true"

	return &DatabaseConfig{
		Host:                getEnvOrDefault("CHOREO_OPENDIF_DB_HOSTNAME", "localhost"),
		Port:                getEnvOrDefault("CHOREO_OPENDIF_DB_PORT", "5432"),
		Username:            getEnvOrDefault("CHOREO_OPENDIF_DB_USERNAME", "postgres"),
		Password:            getEnvOrDefault("CHOREO_OPENDIF_DB_PASSWORD", "password"),
		Database:            getEnvOrDefault("CHOREO_OPENDIF_DB_DATABASENAME", "api_server"),
		SSLMode:             getEnvOrDefault("DB_SSLMODE", "require"),
		MaxOpenConns:        maxOpenConns,
		MaxIdleConns:        maxIdleConns,
		ConnMaxLifetime:     connMaxLifetime,
		ConnMaxIdleTime:     connMaxIdleTime,
		QueryTimeout:        queryTimeout,
		ConnectTimeout:      connectTimeout,
		RetryAttempts:       retryAttempts,
		RetryDelay:          retryDelay,
		TransactionTimeout:  transactionTimeout,
		EnableMonitoring:    enableMonitoring,
		HealthCheckInterval: healthCheckInterval,
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

		// Log connection pool configuration for cloud databases
		slog.Info("Database connection pool configured for cloud database",
			"max_open_conns", config.MaxOpenConns,
			"max_idle_conns", config.MaxIdleConns,
			"conn_max_lifetime", config.ConnMaxLifetime,
			"conn_max_idle_time", config.ConnMaxIdleTime,
			"host", config.Host)

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

// DatabaseStats holds database connection pool statistics
type DatabaseStats struct {
	OpenConnections   int           `json:"open_connections"`
	InUse             int           `json:"in_use"`
	Idle              int           `json:"idle"`
	WaitCount         int64         `json:"wait_count"`
	WaitDuration      time.Duration `json:"wait_duration"`
	MaxIdleClosed     int64         `json:"max_idle_closed"`
	MaxIdleTimeClosed int64         `json:"max_idle_time_closed"`
	MaxLifetimeClosed int64         `json:"max_lifetime_closed"`
	LastChecked       time.Time     `json:"last_checked"`
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

// ConnectionPoolStats represents database connection pool statistics
type ConnectionPoolStats struct {
	OpenConnections int           `json:"open_connections"`
	InUse           int           `json:"in_use"`
	Idle            int           `json:"idle"`
	WaitCount       int64         `json:"wait_count"`
	WaitDuration    time.Duration `json:"wait_duration"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`
}

// GetConnectionPoolStats returns current database connection pool statistics
func GetConnectionPoolStats(db *sql.DB) *ConnectionPoolStats {
	stats := db.Stats()
	return &ConnectionPoolStats{
		OpenConnections: stats.OpenConnections,
		InUse:           stats.InUse,
		Idle:            stats.Idle,
		WaitCount:       stats.WaitCount,
		WaitDuration:    stats.WaitDuration,
		// Note: MaxOpenConns, MaxIdleConns, ConnMaxLifetime, ConnMaxIdleTime are configuration values,
		// not available in sql.DBStats. These would need to be stored separately if needed.
		MaxOpenConns:    0, // Not available in stats
		MaxIdleConns:    0, // Not available in stats
		ConnMaxLifetime: 0, // Not available in stats
		ConnMaxIdleTime: 0, // Not available in stats
	}
}

// LogConnectionPoolStats logs current connection pool statistics
func LogConnectionPoolStats(db *sql.DB) {
	stats := GetConnectionPoolStats(db)
	slog.Info("Database connection pool statistics",
		"open_connections", stats.OpenConnections,
		"in_use", stats.InUse,
		"idle", stats.Idle,
		"wait_count", stats.WaitCount,
		"wait_duration", stats.WaitDuration,
		"max_open_conns", stats.MaxOpenConns,
		"max_idle_conns", stats.MaxIdleConns,
		"conn_max_lifetime", stats.ConnMaxLifetime,
		"conn_max_idle_time", stats.ConnMaxIdleTime)
}

// HealthCheck performs a comprehensive database health check optimized for cloud databases
func HealthCheck(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	// Check if database is reachable with shorter timeout for cloud databases
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	start := time.Now()
	if err := db.PingContext(ctx); err != nil {
		duration := time.Since(start)
		slog.Error("Database ping failed", "error", err, "duration", duration)
		return fmt.Errorf("database ping failed after %v: %w", duration, err)
	}
	duration := time.Since(start)

	// Check connection pool health
	stats := GetConnectionPoolStats(db)

	// Calculate utilization percentage
	utilization := float64(stats.InUse) / float64(stats.MaxOpenConns) * 100

	// More aggressive warnings for cloud databases
	if utilization > 70 {
		slog.Warn("Database connection pool utilization is high (cloud database)",
			"utilization_percent", utilization,
			"in_use", stats.InUse,
			"max_open_conns", stats.MaxOpenConns,
			"ping_duration", duration)
	}

	// Critical warning if approaching limits
	if utilization > 90 {
		slog.Error("Database connection pool utilization is critical (cloud database)",
			"utilization_percent", utilization,
			"in_use", stats.InUse,
			"max_open_conns", stats.MaxOpenConns)
	}

	// Warn if there are waiters (indicates connection pool exhaustion)
	if stats.WaitCount > 0 {
		slog.Warn("Database connection pool has waiters (potential connection limit reached)",
			"wait_count", stats.WaitCount,
			"wait_duration", stats.WaitDuration,
			"utilization_percent", utilization)
	}

	// Log slow ping times (common with cloud databases)
	if duration > 1*time.Second {
		slog.Warn("Database ping is slow (cloud database latency)",
			"ping_duration", duration,
			"utilization_percent", utilization)
	}

	slog.Debug("Database health check passed",
		"open_connections", stats.OpenConnections,
		"in_use", stats.InUse,
		"idle", stats.Idle,
		"utilization_percent", utilization,
		"ping_duration", duration)

	return nil
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

// Transaction represents a database transaction with timeout support
type Transaction struct {
	tx     *sql.Tx
	config *DatabaseConfig
	db     *sql.DB
}

// BeginTransaction starts a new database transaction with timeout
func BeginTransaction(db *sql.DB, config *DatabaseConfig) (*Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.TransactionTimeout)
	defer cancel()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &Transaction{
		tx:     tx,
		config: config,
		db:     db,
	}, nil
}

// BeginTransactionWithIsolation starts a new database transaction with specific isolation level
func BeginTransactionWithIsolation(db *sql.DB, config *DatabaseConfig, isolation sql.IsolationLevel) (*Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.TransactionTimeout)
	defer cancel()

	opts := &sql.TxOptions{
		Isolation: isolation,
		ReadOnly:  false,
	}

	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction with isolation level %d: %w", isolation, err)
	}

	return &Transaction{
		tx:     tx,
		config: config,
		db:     db,
	}, nil
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
	slog.Debug("Committing database transaction")
	if err := t.tx.Commit(); err != nil {
		slog.Error("Failed to commit transaction", "error", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	slog.Debug("Transaction committed successfully")
	return nil
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
	slog.Debug("Rolling back database transaction")
	if err := t.tx.Rollback(); err != nil {
		slog.Error("Failed to rollback transaction", "error", err)
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}
	slog.Debug("Transaction rolled back successfully")
	return nil
}

// Query executes a query that returns rows within the transaction
func (t *Transaction) Query(query string, args ...interface{}) (*sql.Rows, error) {
	ctx, cancel := context.WithTimeout(context.Background(), t.config.QueryTimeout)
	defer cancel()

	slog.Debug("Querying in transaction", "query", query)
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("Failed to query in transaction", "error", err, "query", query)
		return nil, fmt.Errorf("failed to query in transaction: %w", err)
	}
	return rows, nil
}

// QueryRow executes a query that returns a single row within the transaction
func (t *Transaction) QueryRow(query string, args ...interface{}) *sql.Row {
	ctx, cancel := context.WithTimeout(context.Background(), t.config.QueryTimeout)
	defer cancel()

	slog.Debug("Querying single row in transaction", "query", query)
	return t.tx.QueryRowContext(ctx, query, args...)
}

// WithTransaction executes a function within a transaction
func WithTransaction(db *sql.DB, config *DatabaseConfig, fn func(*Transaction) error) error {
	tx, err := BeginTransaction(db, config)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			// Rollback on panic
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				slog.Error("Failed to rollback transaction after panic", "error", rollbackErr)
			}
			panic(p) // Re-throw the panic
		}
	}()

	if err := fn(tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			slog.Error("Failed to rollback transaction", "error", rollbackErr)
			return fmt.Errorf("transaction failed and rollback failed: %w (rollback error: %v)", err, rollbackErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Exec executes a query within the transaction
func (t *Transaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), t.config.QueryTimeout)
	defer cancel()

	return t.tx.ExecContext(ctx, query, args...)
}

// GetDatabaseStats returns current database connection pool statistics
func GetDatabaseStats(db *sql.DB) *DatabaseStats {
	stats := db.Stats()
	return &DatabaseStats{
		OpenConnections:   stats.OpenConnections,
		InUse:             stats.InUse,
		Idle:              stats.Idle,
		WaitCount:         stats.WaitCount,
		WaitDuration:      stats.WaitDuration,
		MaxIdleClosed:     stats.MaxIdleClosed,
		MaxIdleTimeClosed: stats.MaxIdleTimeClosed,
		MaxLifetimeClosed: stats.MaxLifetimeClosed,
		LastChecked:       time.Now(),
	}
}

// StartHealthCheckMonitor starts a background health check monitor
func StartHealthCheckMonitor(db *sql.DB, config *DatabaseConfig) {
	if !config.EnableMonitoring {
		return
	}

	go func() {
		ticker := time.NewTicker(config.HealthCheckInterval)
		defer ticker.Stop()

		for range ticker.C {
			if err := HealthCheck(db); err != nil {
				slog.Error("Database health check failed", "error", err)
			} else {
				slog.Debug("Database health check passed")
			}
		}
	}()
}
