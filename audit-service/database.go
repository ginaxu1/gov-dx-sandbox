package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DatabaseConfig holds database connection configuration for SQLite
type DatabaseConfig struct {
	DatabasePath    string        // Path to SQLite database file
	MaxOpenConns    int           // Maximum number of open connections
	MaxIdleConns    int           // Maximum number of idle connections
	ConnMaxLifetime time.Duration // Maximum amount of time a connection may be reused
	ConnMaxIdleTime time.Duration // Maximum amount of time a connection may be idle before being closed
}

// NewDatabaseConfig creates a new database configuration from environment variables
func NewDatabaseConfig() *DatabaseConfig {
	// SQLite best practice: Use MaxOpenConns=1 to serialize database access through a single connection.
	// This prevents "database is locked" errors that can occur with concurrent write operations,
	// even with WAL mode enabled. While WAL allows concurrent readers, writes are serialized.
	maxOpenConns := parseIntOrDefault("DB_MAX_OPEN_CONNS", 1)
	maxIdleConns := parseIntOrDefault("DB_MAX_IDLE_CONNS", 1)

	// Connection lifetime settings for SQLite
	// ConnMaxLifetime: 1 hour - ensures connections don't stay open indefinitely
	// ConnMaxIdleTime: 15 minutes - closes idle connections after period of inactivity
	connMaxLifetime := parseDurationOrDefault("DB_CONN_MAX_LIFETIME", time.Hour)
	connMaxIdleTime := parseDurationOrDefault("DB_CONN_MAX_IDLE_TIME", 15*time.Minute)

	// Default to ./data/audit.db if not specified
	dbPath := getEnvOrDefault("DB_PATH", "./data/audit.db")

	// Ensure directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		slog.Warn("Failed to create database directory", "path", dbDir, "error", err)
	}

	slog.Info("Database configuration",
		"database_path", dbPath,
		"max_open_conns", maxOpenConns,
		"max_idle_conns", maxIdleConns,
		"conn_max_lifetime", connMaxLifetime,
		"conn_max_idle_time", connMaxIdleTime,
	)
	return &DatabaseConfig{
		DatabasePath:    dbPath,
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: connMaxLifetime,
		ConnMaxIdleTime: connMaxIdleTime,
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
// Accepts formats like "1h", "30m", "15s", etc.
func parseDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := getEnvOrDefault(key, ""); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
		slog.Warn("Invalid duration format, using default", "key", key, "value", value, "default", defaultValue)
	}
	return defaultValue
}

// getEnvOrDefault returns the environment variable value or a default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ConnectGORM establishes a GORM connection to the SQLite database
func ConnectGORM(config *DatabaseConfig) (*gorm.DB, error) {
	slog.Info("Attempting GORM SQLite database connection", "path", config.DatabasePath)

	// Open GORM connection to SQLite
	gormDB, err := gorm.Open(sqlite.Open(config.DatabasePath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open GORM SQLite database connection: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	err = sqlDB.PingContext(ctx)
	cancel()

	if err != nil {
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	slog.Info("GORM SQLite database connection established successfully", "path", config.DatabasePath)
	return gormDB, nil
}
