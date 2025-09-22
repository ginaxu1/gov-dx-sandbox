package main

import (
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/lib/pq"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
	SSLMode  string
}

// NewDatabaseConfig creates a new database configuration from environment variables
func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:     getEnvOrDefault("CHOREO_OPENDIF_DB_HOSTNAME", "localhost"),
		Port:     getEnvOrDefault("CHOREO_OPENDIF_DB_PORT", "5432"),
		Username: getEnvOrDefault("CHOREO_OPENDIF_DB_USERNAME", "postgres"),
		Password: getEnvOrDefault("CHOREO_OPENDIF_DB_PASSWORD", "password"),
		Database: getEnvOrDefault("CHOREO_OPENDIF_DB_DATABASENAME", "consent_engine"),
		SSLMode:  getEnvOrDefault("DB_SSLMODE", "require"),
	}
}

// ConnectDB establishes a connection to the PostgreSQL database
func ConnectDB(config *DatabaseConfig) (*sql.DB, error) {
	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)

	// Open connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("Successfully connected to PostgreSQL database",
		"host", config.Host,
		"port", config.Port,
		"database", config.Database)

	return db, nil
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
