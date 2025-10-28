package v1

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gov-dx-sandbox/exchange/policy-decision-point/v1/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseConfig holds GORM database connection configuration
type DatabaseConfig struct {
	Host            string
	Port            string
	Username        string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// NewDatabaseConfig creates a new GORM database configuration for V1
func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:            getEnvOrDefault("CHOREO_OPENDIF_DATABASE_HOSTNAME", "localhost"),
		Port:            getEnvOrDefault("CHOREO_OPENDIF_DATABASE_PORT", "5432"),
		Username:        getEnvOrDefault("CHOREO_OPENDIF_DATABASE_USERNAME", "postgres"),
		Password:        getEnvOrDefault("CHOREO_OPENDIF_DATABASE_PASSWORD", "password"),
		Database:        getEnvOrDefault("CHOREO_OPENDIF_DATABASE_DATABASENAME", "testdb"),
		SSLMode:         getEnvOrDefault("DB_SSLMODE", "require"),
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,
	}
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ConnectGormDB establishes a GORM connection to PostgreSQL
func ConnectGormDB(config *DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)

	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Warn)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("Successfully connected to PostgreSQL database with GORM (V1)",
		"host", config.Host,
		"port", config.Port,
		"database", config.Database)

	// Only run migration if environment variable is set
	if os.Getenv("RUN_MIGRATION") == "true" {
		slog.Info("Running GORM auto-migration for V1 models")
		err = db.AutoMigrate(
			&models.PolicyMetadata{},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to run auto-migration: %w", err)
		}
		slog.Info("GORM auto-migration completed successfully")

		// Create performance indexes after migration
		if err := createPerformanceIndexes(db); err != nil {
			slog.Warn("Failed to create performance indexes", "error", err)
		} else {
			slog.Info("Performance indexes created successfully")
		}
	} else {
		slog.Info("Database connected (migration skipped)")
	}

	return db, nil
}

// createPerformanceIndexes creates database indexes for performance optimization
func createPerformanceIndexes(db *gorm.DB) error {
	indexes := []string{
		// Policy Metadata table indexes (additional to existing unique index)
		"CREATE INDEX IF NOT EXISTS idx_policy_metadata_schema_id ON policy_metadata(schema_id)",
		"CREATE INDEX IF NOT EXISTS idx_policy_metadata_field_name ON policy_metadata(field_name)",
		"CREATE INDEX IF NOT EXISTS idx_policy_metadata_source ON policy_metadata(source)",
		"CREATE INDEX IF NOT EXISTS idx_policy_metadata_access_control_type ON policy_metadata(access_control_type)",
		"CREATE INDEX IF NOT EXISTS idx_policy_metadata_is_owner ON policy_metadata(is_owner)",
		"CREATE INDEX IF NOT EXISTS idx_policy_metadata_owner ON policy_metadata(owner)",
		"CREATE INDEX IF NOT EXISTS idx_policy_metadata_created_at ON policy_metadata(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_policy_metadata_updated_at ON policy_metadata(updated_at)",

		// Composite indexes for common query patterns
		"CREATE INDEX IF NOT EXISTS idx_policy_metadata_schema_field ON policy_metadata(schema_id, field_name)",
		"CREATE INDEX IF NOT EXISTS idx_policy_metadata_schema_created ON policy_metadata(schema_id, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_policy_metadata_owner_created ON policy_metadata(owner, created_at)",

		// GIN index for JSONB allow_list column for efficient JSON queries
		"CREATE INDEX IF NOT EXISTS idx_policy_metadata_allow_list_gin ON policy_metadata USING GIN (allow_list)",
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			return fmt.Errorf("failed to create index: %s, error: %w", indexSQL, err)
		}
	}

	return nil
}
