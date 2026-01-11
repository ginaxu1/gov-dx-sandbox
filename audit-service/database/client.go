package database

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	configpkg "github.com/gov-dx-sandbox/audit-service/config"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DatabaseType represents the type of database to use
type DatabaseType string

const (
	DatabaseTypeSQLite   DatabaseType = "sqlite"
	DatabaseTypePostgres DatabaseType = "postgres"
)

// Config holds database connection configuration
type Config struct {
	// Database type (sqlite or postgres)
	Type DatabaseType

	// SQLite configuration
	DatabasePath string // Path to SQLite database file

	// PostgreSQL configuration
	Host     string
	Port     string
	Username string
	Password string
	Database string
	SSLMode  string

	// Connection pool settings (applies to both database types)
	MaxOpenConns    int           // Maximum number of open connections
	MaxIdleConns    int           // Maximum number of idle connections
	ConnMaxLifetime time.Duration // Maximum amount of time a connection may be reused
	ConnMaxIdleTime time.Duration // Maximum amount of time a connection may be idle before being closed
}

// NewDatabaseConfig creates a new database configuration from environment variables
// Supports both SQLite (default) and PostgreSQL databases
// Configuration priority:
//  1. If DB_TYPE=postgres → PostgreSQL (DB_HOST, DB_PASSWORD, etc. required)
//  2. If DB_TYPE=sqlite OR DB_PATH is set → File-based SQLite (default: ./data/audit.db)
//     Note: Unknown DB_TYPE values also default to file-based SQLite
//  3. If no database configuration (no DB_TYPE, no DB_PATH) → In-memory SQLite (:memory:)
//     Note: DB_HOST is only relevant for PostgreSQL and does not count as "database configuration"
func NewDatabaseConfig() *Config {
	// Determine database type from environment variable
	dbTypeStr := strings.ToLower(configpkg.GetEnvOrDefault("DB_TYPE", ""))
	var dbType DatabaseType

	// Check SQLite-specific configuration (DB_HOST is only relevant for PostgreSQL)
	dbTypeSet := os.Getenv("DB_TYPE") != ""
	dbPathSet := os.Getenv("DB_PATH") != ""
	dbHostSet := os.Getenv("DB_HOST") != ""

	// For SQLite: only DB_TYPE=sqlite or DB_PATH count as configuration
	// DB_HOST is only relevant when DB_TYPE=postgres
	hasSQLiteConfig := dbPathSet || (dbTypeSet && dbTypeStr != "postgres" && dbTypeStr != "postgresql")

	switch dbTypeStr {
	case "postgres", "postgresql":
		dbType = DatabaseTypePostgres
	case "sqlite":
		dbType = DatabaseTypeSQLite
	case "":
		// No DB_TYPE specified - use SQLite (in-memory or file-based depending on other config)
		dbType = DatabaseTypeSQLite
	default:
		slog.Warn("Unknown DB_TYPE, defaulting to sqlite", "db_type", dbTypeStr)
		dbType = DatabaseTypeSQLite
	}

	// Warn if DB_HOST is set but not using PostgreSQL
	if dbHostSet && dbType != DatabaseTypePostgres {
		slog.Warn("DB_HOST is set but DB_TYPE is not 'postgres'. DB_HOST will be ignored. Use DB_TYPE=postgres to connect to PostgreSQL.",
			"db_type", dbTypeStr,
			"db_host", os.Getenv("DB_HOST"))
	}

	config := &Config{
		Type: dbType,
	}

	if dbType == DatabaseTypeSQLite {
		// SQLite configuration
		// SQLite best practice: Use MaxOpenConns=1 to serialize database access through a single connection.
		// This prevents "database is locked" errors that can occur with concurrent write operations,
		// even with WAL mode enabled. While WAL allows concurrent readers, writes are serialized.
		config.MaxOpenConns = parseIntOrDefault("DB_MAX_OPEN_CONNS", 1)
		config.MaxIdleConns = parseIntOrDefault("DB_MAX_IDLE_CONNS", 1)

		// Determine database path based on configuration
		if !hasSQLiteConfig {
			// No SQLite configuration at all → in-memory database for quick testing
			config.DatabasePath = ":memory:"
			slog.Info("No database configuration found, using in-memory SQLite")
		} else {
			// SQLite configuration provided → file-based SQLite
			config.DatabasePath = configpkg.GetEnvOrDefault("DB_PATH", "./data/audit.db")
		}

		// Ensure directory exists if not in-memory
		if config.DatabasePath != ":memory:" {
			dbDir := filepath.Dir(config.DatabasePath)
			if err := os.MkdirAll(dbDir, 0o755); err != nil {
				slog.Warn("Failed to create database directory", "path", dbDir, "error", err)
			}
		}

		slog.Info("Database configuration (SQLite)",
			"database_path", config.DatabasePath,
			"max_open_conns", config.MaxOpenConns,
			"max_idle_conns", config.MaxIdleConns,
		)
	} else {
		// PostgreSQL configuration
		config.Host = configpkg.GetEnvOrDefault("DB_HOST", "localhost")
		config.Port = configpkg.GetEnvOrDefault("DB_PORT", "5432")
		config.Username = configpkg.GetEnvOrDefault("DB_USERNAME", "postgres")
		config.Password = configpkg.GetEnvOrDefault("DB_PASSWORD", "")
		config.Database = configpkg.GetEnvOrDefault("DB_NAME", "audit_db")
		config.SSLMode = configpkg.GetEnvOrDefault("DB_SSLMODE", "disable")

		// PostgreSQL connection pool settings (higher defaults than SQLite)
		config.MaxOpenConns = parseIntOrDefault("DB_MAX_OPEN_CONNS", 25)
		config.MaxIdleConns = parseIntOrDefault("DB_MAX_IDLE_CONNS", 5)

		slog.Info("Database configuration (PostgreSQL)",
			"host", config.Host,
			"port", config.Port,
			"database", config.Database,
			"username", config.Username,
			"sslmode", config.SSLMode,
			"max_open_conns", config.MaxOpenConns,
			"max_idle_conns", config.MaxIdleConns,
		)
	}

	// Connection lifetime settings (applies to both database types)
	config.ConnMaxLifetime = parseDurationOrDefault("DB_CONN_MAX_LIFETIME", time.Hour)
	config.ConnMaxIdleTime = parseDurationOrDefault("DB_CONN_MAX_IDLE_TIME", 15*time.Minute)

	return config
}

// ConnectGormDB establishes a GORM connection to the database (SQLite or PostgreSQL)
func ConnectGormDB(config *Config) (*gorm.DB, error) {
	var gormDB *gorm.DB
	var err error

	if config.Type == DatabaseTypeSQLite {
		slog.Info("Attempting GORM SQLite database connection", "path", config.DatabasePath)

		// Open GORM connection to SQLite
		gormDB, err = gorm.Open(sqlite.Open(config.DatabasePath), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to open GORM SQLite database connection: %w", err)
		}
	} else {
		// PostgreSQL connection
		// Use net/url to properly encode credentials (handles special characters in passwords)
		dsnURL := url.URL{
			Scheme: "postgres",
			User:   url.UserPassword(config.Username, config.Password),
			Host:   fmt.Sprintf("%s:%s", config.Host, config.Port),
			Path:   config.Database,
		}
		q := dsnURL.Query()
		q.Set("sslmode", config.SSLMode)
		dsnURL.RawQuery = q.Encode()
		dsn := dsnURL.String()

		slog.Info("Attempting GORM PostgreSQL database connection",
			"host", config.Host,
			"port", config.Port,
			"database", config.Database)

		gormDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to open GORM PostgreSQL database connection: %w", err)
		}
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
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	dbTypeStr := "SQLite"
	if config.Type == DatabaseTypePostgres {
		dbTypeStr = "PostgreSQL"
	}
	slog.Info("GORM database connection established successfully", "type", dbTypeStr)
	return gormDB, nil
}

// parseIntOrDefault parses an integer from environment variable or returns default
func parseIntOrDefault(key string, defaultValue int) int {
	if value := configpkg.GetEnvOrDefault(key, ""); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// parseDurationOrDefault parses a duration from environment variable or returns default
// Accepts formats like "1h", "30m", "15s", etc.
func parseDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := configpkg.GetEnvOrDefault(key, ""); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
		slog.Warn("Invalid duration format, using default", "key", key, "value", value, "default", defaultValue)
	}
	return defaultValue
}
