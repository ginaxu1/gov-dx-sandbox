package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
	retryAttempts := parseIntOrDefault("DB_RETRY_ATTEMPTS", 10)
	retryDelay := parseDurationOrDefault("DB_RETRY_DELAY", "2s")

	slog.Info("Database configuration",
		"host", getEnvOrDefault("CHOREO_OPENDIF_DATABASE_HOSTNAME", getEnvOrDefault("CHOREO_OPENDIF_DB_HOSTNAME", "localhost")),
		"port", getEnvOrDefault("CHOREO_OPENDIF_DATABASE_PORT", getEnvOrDefault("CHOREO_OPENDIF_DB_PORT", "5432")),
		"database", getEnvOrDefault("CHOREO_OPENDIF_DATABASE_DATABASENAME", getEnvOrDefault("CHOREO_OPENDIF_DB_DATABASENAME", "gov_dx_sandbox")),
		"max_open_conns", maxOpenConns,
		"max_idle_conns", maxIdleConns,
		"conn_max_lifetime", connMaxLifetime,
		"conn_max_idle_time", connMaxIdleTime,
		"query_timeout", queryTimeout,
		"connect_timeout", connectTimeout,
		"retry_attempts", retryAttempts,
		"retry_delay", retryDelay,
	)
	return &DatabaseConfig{
		Host:            getEnvOrDefault("CHOREO_OPENDIF_DATABASE_HOSTNAME", getEnvOrDefault("CHOREO_OPENDIF_DB_HOSTNAME", "localhost")),
		Port:            getEnvOrDefault("CHOREO_OPENDIF_DATABASE_PORT", getEnvOrDefault("CHOREO_OPENDIF_DB_PORT", "5432")),
		Username:        getEnvOrDefault("CHOREO_OPENDIF_DATABASE_USERNAME", getEnvOrDefault("CHOREO_OPENDIF_DB_USERNAME", "user")),
		Password:        getEnvOrDefault("CHOREO_OPENDIF_DATABASE_PASSWORD", getEnvOrDefault("CHOREO_OPENDIF_DB_PASSWORD", "password")),
		Database:        getEnvOrDefault("CHOREO_OPENDIF_DATABASE_DATABASENAME", getEnvOrDefault("CHOREO_OPENDIF_DB_DATABASENAME", "gov_dx_sandbox")),
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

// getEnvOrDefault returns the environment variable value or a default
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
			"database", config.Database,
			"max_open_conns", config.MaxOpenConns,
			"max_idle_conns", config.MaxIdleConns,
			"conn_max_lifetime", config.ConnMaxLifetime,
			"conn_max_idle_time", config.ConnMaxIdleTime)

		return db, nil
	}

	return nil, fmt.Errorf("failed to establish database connection after %d attempts", config.RetryAttempts)
}

// ConnectGORM establishes a GORM connection to the PostgreSQL database
func ConnectGORM(config *DatabaseConfig) (*gorm.DB, error) {
	// Build DSN string
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		config.Host, config.Username, config.Password, config.Database, config.Port, config.SSLMode)

	var gormDB *gorm.DB
	var err error

	// Retry connection attempts
	for attempt := 1; attempt <= config.RetryAttempts; attempt++ {
		slog.Info("Attempting GORM database connection", "attempt", attempt, "max_attempts", config.RetryAttempts)

		// Open GORM connection
		gormDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			slog.Warn("Failed to open GORM database connection", "attempt", attempt, "error", err)
			if attempt < config.RetryAttempts {
				time.Sleep(config.RetryDelay)
				continue
			}
			return nil, fmt.Errorf("failed to open GORM database connection after %d attempts: %w", config.RetryAttempts, err)
		}

		// Get underlying sql.DB to configure connection pool
		sqlDB, err := gormDB.DB()
		if err != nil {
			slog.Warn("Failed to get underlying sql.DB", "attempt", attempt, "error", err)
			if attempt < config.RetryAttempts {
				time.Sleep(config.RetryDelay)
				continue
			}
			return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
		}

		// Configure connection pool
		sqlDB.SetMaxOpenConns(config.MaxOpenConns)
		sqlDB.SetMaxIdleConns(config.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
		sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

		// Test connection with timeout
		ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
		err = sqlDB.PingContext(ctx)
		cancel()

		if err != nil {
			slog.Warn("Failed to ping database via GORM", "attempt", attempt, "error", err)
			if attempt < config.RetryAttempts {
				time.Sleep(config.RetryDelay)
				continue
			}
			return nil, fmt.Errorf("failed to ping database via GORM after %d attempts: %w", config.RetryAttempts, err)
		}

		slog.Info("GORM database connection established successfully",
			"host", config.Host,
			"port", config.Port,
			"database", config.Database)
		return gormDB, nil
	}

	return nil, fmt.Errorf("failed to establish GORM database connection after %d attempts", config.RetryAttempts)
}

// InitDatabase creates the necessary tables if they don't exist
func InitDatabase(db *sql.DB) error {
	// Check if audit_logs table exists and has the new schema
	var tableExists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'audit_logs'
		)
	`).Scan(&tableExists)

	if err != nil {
		return fmt.Errorf("failed to check if audit_logs table exists: %w", err)
	}

	if tableExists {
		// Check if table has consumer_id and provider_id columns
		var hasConsumerID bool
		var hasProviderID bool
		err := db.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.columns 
				WHERE table_schema = 'public' 
				AND table_name = 'audit_logs' 
				AND column_name = 'consumer_id'
			)
		`).Scan(&hasConsumerID)
		if err != nil {
			return fmt.Errorf("failed to check for consumer_id column: %w", err)
		}

		err = db.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.columns 
				WHERE table_schema = 'public' 
				AND table_name = 'audit_logs' 
				AND column_name = 'provider_id'
			)
		`).Scan(&hasProviderID)
		if err != nil {
			return fmt.Errorf("failed to check for provider_id column: %w", err)
		}

		if !hasConsumerID || !hasProviderID {
			// Add missing columns
			slog.Info("Adding consumer_id and provider_id columns to audit_logs table")
			if !hasConsumerID {
				if _, err := db.Exec("ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS consumer_id VARCHAR(255)"); err != nil {
					return fmt.Errorf("failed to add consumer_id column: %w", err)
				}
				// Create index for consumer_id
				if _, err := db.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_consumer_id ON audit_logs(consumer_id)"); err != nil {
					slog.Warn("Failed to create index for consumer_id", "error", err)
				}
			}
			if !hasProviderID {
				if _, err := db.Exec("ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS provider_id VARCHAR(255)"); err != nil {
					return fmt.Errorf("failed to add provider_id column: %w", err)
				}
				// Create index for provider_id
				if _, err := db.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_provider_id ON audit_logs(provider_id)"); err != nil {
					slog.Warn("Failed to create index for provider_id", "error", err)
				}
			}
			// Create composite index for member-to-member queries
			if _, err := db.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_consumer_provider_timestamp ON audit_logs(consumer_id, provider_id, timestamp DESC)"); err != nil {
				slog.Warn("Failed to create composite index", "error", err)
			}
		} else {
			slog.Info("audit_logs table already exists with consumer_id and provider_id columns")
		}
		// Update view to select directly from table
		return updateAuditLogsView(db)
	}

	// Create the new schema table
	createTableSQL := `
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

	-- Create indexes for better performance
	CREATE INDEX IF NOT EXISTS idx_audit_logs_application_id ON audit_logs(application_id);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_schema_id ON audit_logs(schema_id);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_status ON audit_logs(status);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_consumer_id ON audit_logs(consumer_id);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_provider_id ON audit_logs(provider_id);
	
	-- Create composite indexes for common query patterns
	CREATE INDEX IF NOT EXISTS idx_audit_logs_application_timestamp ON audit_logs(application_id, timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_schema_timestamp ON audit_logs(schema_id, timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_consumer_provider_timestamp ON audit_logs(consumer_id, provider_id, timestamp DESC);

	-- Note: View creation is handled separately after table creation
	`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create audit_logs table: %w", err)
	}

	// Create view that selects consumer_id and provider_id directly from audit_logs table
	if err := updateAuditLogsView(db); err != nil {
		return err
	}

	// Create management_events table for Admin/Member Portal events
	createManagementEventsTableSQL := `
	CREATE TABLE IF NOT EXISTS management_events (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		event_type VARCHAR(10) NOT NULL CHECK (event_type IN ('CREATE', 'UPDATE', 'DELETE')),
		status VARCHAR(10) NOT NULL CHECK (status IN ('SUCCESS', 'FAILURE')),
		timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
		
		-- Actor information
		actor_type VARCHAR(10) NOT NULL CHECK (actor_type IN ('USER', 'SERVICE')),
		actor_id VARCHAR(255), -- NULL if SERVICE type
		actor_role VARCHAR(10) CHECK (actor_role IN ('MEMBER', 'ADMIN')), -- NULL if SERVICE type
		
		-- Target resource
		target_resource VARCHAR(50) NOT NULL CHECK (target_resource IN (
			'MEMBERS', 'SCHEMAS', 'SCHEMA-SUBMISSIONS', 
			'APPLICATIONS', 'APPLICATION-SUBMISSIONS', 'POLICY-METADATA'
		)),
		target_resource_id VARCHAR(255), -- NULL allowed for CREATE failures
		
		-- Additional metadata
		metadata JSONB,
		
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	-- Create indexes for better performance
	CREATE INDEX IF NOT EXISTS idx_management_events_event_id ON management_events(event_id);
	CREATE INDEX IF NOT EXISTS idx_management_events_timestamp ON management_events(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_management_events_actor ON management_events(actor_type, actor_id);
	CREATE INDEX IF NOT EXISTS idx_management_events_target ON management_events(target_resource, target_resource_id);
	CREATE INDEX IF NOT EXISTS idx_management_events_status ON management_events(status);
	CREATE INDEX IF NOT EXISTS idx_management_events_actor_target_time ON management_events(actor_type, actor_id, timestamp DESC);
	`

	_, err = db.Exec(createManagementEventsTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create management_events table: %w", err)
	}

	slog.Info("Database tables and view initialized successfully")
	return nil
}

// updateAuditLogsView creates or updates the view to select consumer_id and provider_id directly from audit_logs
func updateAuditLogsView(db *sql.DB) error {
	// Create view that selects consumer_id and provider_id directly from audit_logs table
	// (no longer using joins since these columns are now stored directly in the table)
	viewSQL := `
	CREATE OR REPLACE VIEW audit_logs_with_provider_consumer AS
	SELECT id,
		   "timestamp",
		   status,
		   requested_data,
		   application_id,
		   schema_id,
		   COALESCE(consumer_id, 'unknown') as consumer_id,
		   COALESCE(provider_id, 'unknown') as provider_id
	FROM audit_logs;
	`

	_, err := db.Exec(viewSQL)
	if err != nil {
		return fmt.Errorf("failed to create/update audit_logs_with_provider_consumer view: %w", err)
	}

	slog.Info("Created/updated audit_logs_with_provider_consumer view (selecting directly from audit_logs table)")
	return nil
}

// GracefulShutdown gracefully closes the database connection
func GracefulShutdown(db *sql.DB) error {
	if db == nil {
		return nil
	}

	slog.Info("Starting database graceful shutdown")

	if err := db.Close(); err != nil {
		slog.Error("Error during database shutdown", "error", err)
		return fmt.Errorf("failed to close database: %w", err)
	}

	slog.Info("Database connection closed successfully")
	return nil
}
