package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gov-dx-sandbox/audit-service/consumer"
	"github.com/gov-dx-sandbox/audit-service/handlers"
	"github.com/gov-dx-sandbox/audit-service/redis"
	"github.com/gov-dx-sandbox/audit-service/services"
)

// Build information - set during build
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// simpleCORS is a simple middleware function that adds CORS headers to all responses
func simpleCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}

// Database-related functions are now in database.go

func main() {
	// Parse command line flags
	var (
		env  = flag.String("env", getEnvOrDefault("ENVIRONMENT", "production"), "Environment (development, production)")
		port = flag.String("port", getEnvOrDefault("PORT", "3001"), "Port to listen on")
	)
	flag.Parse()

	// Server configuration
	serverPort := *port

	// Initialize database connection using GORM
	dbConfig := &GormDatabaseConfig{
		Host:            getEnvOrDefault("CHOREO_DB_AUDIT_HOSTNAME", "localhost"),
		Port:            getEnvOrDefault("CHOREO_DB_AUDIT_PORT", "5432"),
		Username:        getEnvOrDefault("CHOREO_DB_AUDIT_USERNAME", "postgres"),
		Password:        getEnvOrDefault("CHOREO_DB_AUDIT_PASSWORD", "password"),
		Database:        getEnvOrDefault("CHOREO_DB_AUDIT_DATABASENAME", "gov_dx_sandbox"),
		SSLMode:         getEnvOrDefault("DB_SSLMODE", "require"),
		MaxOpenConns:    parseIntOrDefault("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    parseIntOrDefault("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: parseDurationOrDefault("DB_CONN_MAX_LIFETIME", "1h"),
		ConnMaxIdleTime: parseDurationOrDefault("DB_CONN_MAX_IDLE_TIME", "30m"),
	}

	slog.Info("Connecting to database",
		"host", dbConfig.Host,
		"port", dbConfig.Port,
		"database", dbConfig.Database)

	db, err := NewGormDatabase(dbConfig)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Initialize database tables using GORM AutoMigrate
	slog.Info("Initializing database tables")
	if err := AutoMigrate(db); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		slog.Warn("Continuing with database initialization failure - some operations may not work")
	}

	// Initialize services
	auditService := services.NewAuditService(db)

	// Initialize handlers
	auditHandler := handlers.NewAuditHandler(auditService)

	// Initialize Redis client for stream consumer (optional)
	var redisClient *redis.RedisClient
	var streamConsumer *consumer.StreamConsumer

	redisHost := getEnvOrDefault("CHOREO_REDIS_AUDIT_HOSTNAME", getEnvOrDefault("REDIS_HOST", ""))
	redisPort := getEnvOrDefault("CHOREO_REDIS_AUDIT_PORT", getEnvOrDefault("REDIS_PORT", ""))
	redisUsername := getEnvOrDefault("CHOREO_REDIS_AUDIT_USERNAME", getEnvOrDefault("REDIS_USERNAME", ""))
	redisPassword := getEnvOrDefault("CHOREO_REDIS_AUDIT_PASSWORD", getEnvOrDefault("REDIS_PASSWORD", ""))

	if redisHost != "" && redisPort != "" {
		// Combine host and port
		redisAddr := redisHost + ":" + redisPort

		slog.Info("Initializing Redis connection for stream consumer", "address", redisAddr, "username", redisUsername)
		redisConfig := &redis.Config{
			Addr:     redisAddr,
			Username: redisUsername,
			Password: redisPassword,
			DB:       0,
		}

		var err error
		redisClient, err = redis.NewClient(redisConfig)
		if err != nil {
			slog.Warn("Failed to connect to Redis, continuing without stream consumer", "error", err)
		} else {
			slog.Info("Successfully connected to Redis")

			// Create processor
			processor := consumer.NewDatabaseEventProcessor(auditService)

			// Create consumer
			streamConsumer, err = consumer.NewStreamConsumer(redisClient, processor, "audit-consumer-1")
			if err != nil {
				slog.Warn("Failed to create stream consumer", "error", err)
				redisClient.Close()
				redisClient = nil
			} else {
				slog.Info("Stream consumer created successfully")
			}
		}
	} else {
		slog.Info("Redis not configured, running in API-only mode")
	}

	// Setup routes
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Simple health check - just return healthy if service is running
		// Database connectivity is checked during startup, not in health check
		w.WriteHeader(http.StatusOK)
		response := map[string]string{
			"service": "audit-service",
			"status":  "healthy",
		}

		json.NewEncoder(w).Encode(response)
	})

	// Version endpoint
	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := map[string]string{
			"version":   Version,
			"buildTime": BuildTime,
			"gitCommit": GitCommit,
			"service":   "audit-service",
		}

		json.NewEncoder(w).Encode(response)
	})

	// API endpoints for log access
	mux.HandleFunc("/api/logs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			auditHandler.HandleAuditLogs(w, r)
		case http.MethodPost:
			auditHandler.HandleCreateLog(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Start server
	slog.Info("Audit Service starting",
		"environment", *env,
		"port", serverPort,
		"version", Version,
		"buildTime", BuildTime,
		"gitCommit", GitCommit)
	slog.Info("Database configuration",
		"host", dbConfig.Host,
		"port", dbConfig.Port,
		"database", dbConfig.Database,
		"choreoHost", os.Getenv("CHOREO_DB_AUDIT_HOSTNAME"),
		"fallbackHost", os.Getenv("DB_HOST"))

	server := &http.Server{
		Addr:         ":" + serverPort,
		Handler:      simpleCORS(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		slog.Info("Audit Service starting",
			"environment", *env,
			"port", serverPort,
			"version", Version,
			"buildTime", BuildTime,
			"gitCommit", GitCommit)
		slog.Info("Database configuration",
			"host", dbConfig.Host,
			"port", dbConfig.Port,
			"database", dbConfig.Database,
			"choreoHost", os.Getenv("CHOREO_DB_AUDIT_HOSTNAME"),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Start Redis stream consumer if configured
	if streamConsumer != nil {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			log.Printf("Starting Redis stream consumer...")
			streamConsumer.Start(ctx)
		}()

		slog.Info("Redis stream consumer started")
	}

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down Audit Service...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	// Gracefully close database connection
	slog.Info("Closing database connection")
	if sqlDB, err := db.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			slog.Error("Error during database shutdown", "error", err)
		}
	}

	// Gracefully close Redis connection
	if redisClient != nil {
		slog.Info("Closing Redis connection")
		if err := redisClient.Close(); err != nil {
			slog.Error("Error during Redis shutdown", "error", err)
		}
	}

	slog.Info("Audit Service exited")
}
