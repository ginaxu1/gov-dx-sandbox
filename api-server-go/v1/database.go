package v1

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/v1/models"
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
		Database:        getEnvOrDefault("CHOREO_OPENDIF_DATABASE_DATABASENAME_V1", "testdb"),
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
			&models.Entity{},
			&models.Provider{},
			&models.Consumer{},
			&models.Schema{},
			&models.SchemaSubmission{},
			&models.Application{},
			&models.ApplicationSubmission{},
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
		// Entity table indexes
		"CREATE INDEX IF NOT EXISTS idx_entities_email ON entities(email)",
		"CREATE INDEX IF NOT EXISTS idx_entities_idp_user_id ON entities(idp_user_id)",
		"CREATE INDEX IF NOT EXISTS idx_entities_created_at ON entities(created_at)",

		// Provider table indexes
		"CREATE INDEX IF NOT EXISTS idx_providers_entity_id ON providers(entity_id)",
		"CREATE INDEX IF NOT EXISTS idx_providers_created_at ON providers(created_at)",

		// Consumer table indexes
		"CREATE INDEX IF NOT EXISTS idx_consumers_entity_id ON consumers(entity_id)",
		"CREATE INDEX IF NOT EXISTS idx_consumers_created_at ON consumers(created_at)",

		// Schema table indexes
		"CREATE INDEX IF NOT EXISTS idx_provider_schemas_provider_id ON provider_schemas(provider_id)",
		"CREATE INDEX IF NOT EXISTS idx_provider_schemas_version ON provider_schemas(version)",
		"CREATE INDEX IF NOT EXISTS idx_provider_schemas_created_at ON provider_schemas(created_at)",

		// Schema Submission table indexes
		"CREATE INDEX IF NOT EXISTS idx_provider_schema_submissions_provider_id ON provider_schema_submissions(provider_id)",
		"CREATE INDEX IF NOT EXISTS idx_provider_schema_submissions_status ON provider_schema_submissions(status)",
		"CREATE INDEX IF NOT EXISTS idx_provider_schema_submissions_previous_schema_id ON provider_schema_submissions(previous_schema_id)",
		"CREATE INDEX IF NOT EXISTS idx_provider_schema_submissions_created_at ON provider_schema_submissions(created_at)",

		// Application table indexes
		"CREATE INDEX IF NOT EXISTS idx_consumer_applications_consumer_id ON consumer_applications(consumer_id)",
		"CREATE INDEX IF NOT EXISTS idx_consumer_applications_version ON consumer_applications(version)",
		"CREATE INDEX IF NOT EXISTS idx_consumer_applications_created_at ON consumer_applications(created_at)",

		// Application Submission table indexes
		"CREATE INDEX IF NOT EXISTS idx_consumer_application_submissions_consumer_id ON consumer_application_submissions(consumer_id)",
		"CREATE INDEX IF NOT EXISTS idx_consumer_application_submissions_status ON consumer_application_submissions(status)",
		"CREATE INDEX IF NOT EXISTS idx_consumer_application_submissions_previous_application_id ON consumer_application_submissions(previous_application_id)",
		"CREATE INDEX IF NOT EXISTS idx_consumer_application_submissions_created_at ON consumer_application_submissions(created_at)",

		// Composite indexes for common query patterns
		"CREATE INDEX IF NOT EXISTS idx_consumers_entity_created ON consumers(entity_id, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_providers_entity_created ON providers(entity_id, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_applications_consumer_created ON consumer_applications(consumer_id, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_schemas_provider_created ON provider_schemas(provider_id, created_at)",
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			return fmt.Errorf("failed to create index: %s, error: %w", indexSQL, err)
		}
	}

	return nil
}
