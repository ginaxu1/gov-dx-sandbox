package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
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

// MongoDBConfig holds MongoDB connection configuration
type MongoDBConfig struct {
	URI            string
	Database       string
	Collection     string
	ConnectTimeout time.Duration
	RetryAttempts  int
	RetryDelay     time.Duration
}

// NewMongoDBConfig creates a new MongoDB configuration from environment variables
func NewMongoDBConfig() *MongoDBConfig {
	connectTimeout := parseDurationOrDefault("MONGODB_CONNECT_TIMEOUT", "10s")
	retryAttempts := parseIntOrDefault("MONGODB_RETRY_ATTEMPTS", 10)
	retryDelay := parseDurationOrDefault("MONGODB_RETRY_DELAY", "2s")

	uri := getEnvOrDefault("MONGODB_URI", getEnvOrDefault("CHOREO_MONGODB_CONNECTION_URI", "mongodb://localhost:27017"))
	database := getEnvOrDefault("MONGODB_DATABASE", "audit")
	collection := getEnvOrDefault("MONGODB_COLLECTION", "audit_logs")

	slog.Info("MongoDB configuration",
		"uri", maskURI(uri),
		"database", database,
		"collection", collection,
		"connect_timeout", connectTimeout,
		"retry_attempts", retryAttempts,
		"retry_delay", retryDelay)

	return &MongoDBConfig{
		URI:            uri,
		Database:       database,
		Collection:     collection,
		ConnectTimeout: connectTimeout,
		RetryAttempts:  retryAttempts,
		RetryDelay:     retryDelay,
	}
}

// maskURI masks sensitive parts of MongoDB URI for logging
func maskURI(uri string) string {
	// Simple masking - replace password part if present
	// mongodb://user:password@host:port/db -> mongodb://user:***@host:port/db
	if len(uri) > 20 {
		return uri[:10] + "***" + uri[len(uri)-10:]
	}
	return "***"
}

// ConnectMongoDB establishes a MongoDB connection
func ConnectMongoDB(config *MongoDBConfig) (*mongo.Client, *mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer cancel()

	clientOptions := options.Client().ApplyURI(config.URI)
	clientOptions.SetWriteConcern(writeconcern.New(writeconcern.WMajority(), writeconcern.J(true)))
	clientOptions.SetReadConcern(readconcern.Majority())

	var client *mongo.Client
	var err error

	// Retry connection attempts
	for attempt := 1; attempt <= config.RetryAttempts; attempt++ {
		slog.Info("Attempting MongoDB connection", "attempt", attempt, "max_attempts", config.RetryAttempts)

		client, err = mongo.Connect(ctx, clientOptions)
		if err != nil {
			slog.Warn("Failed to connect to MongoDB", "attempt", attempt, "error", err)
			if attempt < config.RetryAttempts {
				time.Sleep(config.RetryDelay)
				continue
			}
			return nil, nil, fmt.Errorf("failed to connect to MongoDB after %d attempts: %w", config.RetryAttempts, err)
		}

		// Test connection
		err = client.Ping(ctx, nil)
		if err != nil {
			slog.Warn("Failed to ping MongoDB", "attempt", attempt, "error", err)
			client.Disconnect(ctx)
			if attempt < config.RetryAttempts {
				time.Sleep(config.RetryDelay)
				continue
			}
			return nil, nil, fmt.Errorf("failed to ping MongoDB after %d attempts: %w", config.RetryAttempts, err)
		}

		slog.Info("MongoDB connection established successfully",
			"database", config.Database,
			"collection", config.Collection)
		break
	}

	db := client.Database(config.Database)
	return client, db, nil
}

// CreateMongoDBIndexes creates indexes for the audit_logs collection
func CreateMongoDBIndexes(ctx context.Context, db *mongo.Database, collectionName string) error {
	collection := db.Collection(collectionName)

	indexes := []mongo.IndexModel{
		// Trace ID index (partial, only for non-null trace_id)
		{
			Keys:    bson.D{{Key: "trace_id", Value: 1}, {Key: "timestamp", Value: 1}},
			Options: options.Index().SetPartialFilterExpression(bson.M{"trace_id": bson.M{"$ne": nil}}),
		},
		// Timestamp index
		{
			Keys: bson.D{{Key: "timestamp", Value: -1}},
		},
		// Event name index
		{
			Keys: bson.D{{Key: "event_name", Value: 1}},
		},
		// Status index
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		// Actor service name index (partial, only for SERVICE actor_type)
		{
			Keys:    bson.D{{Key: "actor_service_name", Value: 1}, {Key: "actor_type", Value: 1}},
			Options: options.Index().SetPartialFilterExpression(bson.M{"actor_type": "SERVICE"}),
		},
		// Actor user ID index (partial, only for USER actor_type)
		{
			Keys:    bson.D{{Key: "actor_user_id", Value: 1}, {Key: "actor_type", Value: 1}},
			Options: options.Index().SetPartialFilterExpression(bson.M{"actor_type": "USER"}),
		},
		// Target service name index (partial, only for SERVICE target_type)
		{
			Keys:    bson.D{{Key: "target_service_name", Value: 1}, {Key: "target_type", Value: 1}},
			Options: options.Index().SetPartialFilterExpression(bson.M{"target_type": "SERVICE"}),
		},
		// Target resource index (partial, only for RESOURCE target_type)
		{
			Keys:    bson.D{{Key: "target_resource", Value: 1}, {Key: "target_type", Value: 1}},
			Options: options.Index().SetPartialFilterExpression(bson.M{"target_type": "RESOURCE"}),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	slog.Info("MongoDB indexes created successfully", "collection", collectionName, "count", len(indexes))
	return nil
}
