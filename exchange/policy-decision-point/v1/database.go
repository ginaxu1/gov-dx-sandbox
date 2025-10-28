package v1

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

// NewDatabaseConfig creates a new database configuration from environment variables
func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:     getEnvOrDefault("CHOREO_OPENDIF_DATABASE_HOSTNAME", "localhost"),
		Port:     getEnvOrDefault("CHOREO_OPENDIF_DATABASE_PORT", "5432"),
		User:     getEnvOrDefault("CHOREO_OPENDIF_DATABASE_USERNAME", "postgres"),
		Password: getEnvOrDefault("CHOREO_OPENDIF_DATABASE_PASSWORD", "password"),
		Database: getEnvOrDefault("CHOREO_OPENDIF_DATABASE_DATABASENAME", "policy_decision_point"),
		SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
	}
}

// ConnectGormDB establishes a GORM database connection
func ConnectGormDB(config *DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode)

	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Info)
	if os.Getenv("LOG_LEVEL") == "debug" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Warn)
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// AutoMigrate runs database migrations
func AutoMigrate(db *gorm.DB) error {
	// Import models to ensure they're registered with GORM
	// This will be done by importing the models package in main.go

	// Run migrations
	if err := db.AutoMigrate(
	// Models will be imported and registered here
	); err != nil {
		return fmt.Errorf("failed to run auto migrations: %w", err)
	}

	return nil
}

// getEnvOrDefault gets an environment variable with a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
