package main

import (
	"context"
	"encoding/json"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gov-dx-sandbox/audit-service/config"
	"github.com/gov-dx-sandbox/audit-service/middleware"
	v1database "github.com/gov-dx-sandbox/audit-service/v1/database"
	v1handlers "github.com/gov-dx-sandbox/audit-service/v1/handlers"
	v1models "github.com/gov-dx-sandbox/audit-service/v1/models"
	v1services "github.com/gov-dx-sandbox/audit-service/v1/services"
)

// Build information - set during build
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

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

	// Load enum configuration from YAML file
	configPath := getEnvOrDefault("AUDIT_ENUMS_CONFIG", "config/enums.yaml")
	enums, err := config.LoadEnums(configPath)
	if err != nil {
		slog.Warn("Failed to load enum configuration, using defaults", "error", err, "path", configPath)
		enums = config.GetDefaultEnums()
	}
	slog.Info("Loaded enum configuration", "path", configPath,
		"eventTypes", len(enums.EventTypes),
		"eventActions", len(enums.EventActions),
		"actorTypes", len(enums.ActorTypes),
		"targetTypes", len(enums.TargetTypes))

	// Initialize enum configuration in models package
	// Pass the AuditEnums instance to leverage O(1) validation methods
	v1models.SetEnumConfig(enums)

	// Initialize database connection
	dbConfig := NewDatabaseConfig()
	slog.Info("Connecting to database",
		"database_path", dbConfig.DatabasePath)

	// Initialize GORM connection
	gormDB, err := ConnectGORM(dbConfig)
	if err != nil {
		slog.Error("Failed to connect to database via GORM", "error", err)
		os.Exit(1)
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

	// Initialize v1 API with database-agnostic repository
	v1Repository := v1database.NewGormRepository(gormDB)
	v1AuditService := v1services.NewAuditService(v1Repository)
	v1AuditHandler := v1handlers.NewAuditHandler(v1AuditService)

	// API endpoint for generalized audit logs (V1)
	mux.HandleFunc("/api/audit-logs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			v1AuditHandler.CreateAuditLog(w, r)
		case http.MethodGet:
			v1AuditHandler.GetAuditLogs(w, r)
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
		"database_path", dbConfig.DatabasePath)

	// Setup CORS middleware
	corsMiddleware := middleware.NewCORSMiddleware()

	// Apply middleware chain: CORS -> main handler
	handler := corsMiddleware(mux)

	server := &http.Server{
		Addr:         ":" + serverPort,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	slog.Info("Starting HTTP server", "address", server.Addr)

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

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

	slog.Info("Audit Service exited")
}
