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

	"github.com/gov-dx-sandbox/audit-service/handlers"
	"github.com/gov-dx-sandbox/audit-service/middleware"
	"github.com/gov-dx-sandbox/audit-service/services"
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

	// Initialize database connection
	dbConfig := NewDatabaseConfig()
	slog.Info("Connecting to database",
		"host", dbConfig.Host,
		"port", dbConfig.Port,
		"database", dbConfig.Database)

	db, err := ConnectDB(dbConfig)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	// Note: We'll close the database connection in graceful shutdown, not with defer

	// Initialize database tables
	slog.Info("Initializing database tables and views")
	if err := InitDatabase(db); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		// Don't exit immediately - log the error and continue
		// The service can still start and handle requests, database operations will fail gracefully
		slog.Warn("Continuing with database initialization failure - some operations may not work")
	}

	// Initialize GORM connection for management events
	gormDB, err := ConnectGORM(dbConfig)
	if err != nil {
		slog.Error("Failed to connect to database via GORM", "error", err)
		os.Exit(1)
	}

	// Initialize services
	auditService := services.NewAuditService(db)
	managementEventService := services.NewManagementEventService(gormDB)

	// Initialize handlers
	auditHandler := handlers.NewAuditHandler(auditService)
	managementEventHandler := handlers.NewManagementEventHandler(managementEventService)

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

	// API endpoint for log access (GET only - for Admin Portal and Entity Portals)
	mux.HandleFunc("/api/logs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		auditHandler.GetLogs(w, r)
	})

	// API endpoint for data exchange events from Orchestration Engine
	mux.HandleFunc("/v1/audit/exchange", auditHandler.CreateDataExchangeEvent)

	// API endpoints for management events from API Server
	mux.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			managementEventHandler.GetEvents(w, r)
		case http.MethodPost:
			managementEventHandler.CreateEvent(w, r)
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
		"choreoHost", os.Getenv("CHOREO_OPENDIF_DB_HOSTNAME"),
		"fallbackHost", os.Getenv("DB_HOST"))

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

	// Gracefully close database connection
	if err := GracefulShutdown(db); err != nil {
		slog.Error("Error during database graceful shutdown", "error", err)
	}

	slog.Info("Audit Service exited")
}
