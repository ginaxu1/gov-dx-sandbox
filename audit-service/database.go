package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

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
