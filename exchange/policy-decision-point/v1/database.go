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

		// Create enum types if they don't exist
		// Note: These enum values must match the constants defined in models/policy.go
		enumQueries := []string{
			// Matches models.SourcePrimary and models.SourceFallback
			`DO $$ BEGIN
				CREATE TYPE source_enum AS ENUM ('primary', 'fallback');
			EXCEPTION
				WHEN duplicate_object THEN null;
			END $$;`,
			// Matches models.AccessControlPublic and models.AccessControlRestricted
			`DO $$ BEGIN
				CREATE TYPE access_control_type_enum AS ENUM ('public', 'restricted');
			EXCEPTION
				WHEN duplicate_object THEN null;
			END $$;`,
			// Matches models.OwnerCitizen
			`DO $$ BEGIN
				CREATE TYPE owner_enum AS ENUM ('citizen');
			EXCEPTION
				WHEN duplicate_object THEN null;
			END $$;`,
		}

		for _, query := range enumQueries {
			if err := db.Exec(query).Error; err != nil {
				return nil, fmt.Errorf("failed to create enum type: %w", err)
			}
		}

		err = db.AutoMigrate(
			&models.PolicyMetadata{},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to run auto-migration: %w", err)
		}
		slog.Info("GORM auto-migration completed successfully")
	} else {
		slog.Info("Database connected (migration skipped)")
	}

	return db, nil
}
