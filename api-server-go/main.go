package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/handlers"
	"github.com/gov-dx-sandbox/api-server-go/middleware"
	"github.com/gov-dx-sandbox/api-server-go/shared/utils"
	v1 "github.com/gov-dx-sandbox/api-server-go/v1"
	v1handlers "github.com/gov-dx-sandbox/api-server-go/v1/handlers"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists (optional - fails silently if not found)
	_ = godotenv.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	slog.SetDefault(logger)

	slog.Info("Starting API Server initialization")

	// Initialize database connection (legacy)
	dbConfig := NewDatabaseConfig()
	db, err := ConnectDB(dbConfig)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := GracefulShutdown(db); err != nil {
			slog.Error("Error during database graceful shutdown", "error", err)
		}
	}()

	// Initialize database tables (legacy)
	if err := InitDatabase(db); err != nil {
		slog.Error("Failed to initialize database tables", "error", err)
		os.Exit(1)
	}

	// Initialize GORM database connection for V1
	v1DbConfig := v1.NewDatabaseConfig()
	gormDB, err := v1.ConnectGormDB(v1DbConfig)
	if err != nil {
		slog.Error("Failed to connect to GORM database", "error", err)
		os.Exit(1)
	}

	// GORM database connection handles migration based on RUN_MIGRATION env var

	// Initialize API server with database (legacy)
	apiServer := handlers.NewAPIServerWithDB(db)

	// Initialize V1 handlers
	v1Handler := v1handlers.NewV1Handler(gormDB)

	// Setup routes
	mux := http.NewServeMux()
	apiServer.SetupRoutes(mux)   // Legacy routes
	v1Handler.SetupV1Routes(mux) // V1 routes

	// Health check endpoint (matching consent-engine approach)
	// --- Structured types for health status ---
	type DBHealth struct {
		Status   string `json:"status"`
		Error    string `json:"error,omitempty"`
		Database string `json:"database,omitempty"`
	}
	type HealthStatus struct {
		Status    string              `json:"status"`
		Service   string              `json:"service"`
		Databases map[string]DBHealth `json:"databases"`
	}

	mux.Handle("/health", utils.PanicRecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := HealthStatus{
			Status:  "healthy",
			Service: "api-server",
			Databases: map[string]DBHealth{
				"legacy": {Status: "unknown"},
				"v1":     {Status: "unknown"},
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		allHealthy := true

		// Test legacy database connection
		if db == nil {
			status.Databases["legacy"] = DBHealth{Status: "unhealthy", Error: "connection is nil"}
			allHealthy = false
		} else if err := db.PingContext(ctx); err != nil {
			status.Databases["legacy"] = DBHealth{Status: "unhealthy", Error: err.Error()}
			allHealthy = false
		} else {
			status.Databases["legacy"] = DBHealth{Status: "healthy", Database: dbConfig.Database}
		}

		// Test V1 GORM database connection
		if gormDB == nil {
			status.Databases["v1"] = DBHealth{Status: "unhealthy", Error: "GORM connection is nil"}
			allHealthy = false
		} else {
			sqlDB, err := gormDB.DB()
			if err != nil {
				status.Databases["v1"] = DBHealth{Status: "unhealthy", Error: fmt.Sprintf("failed to get sql.DB: %v", err)}
				allHealthy = false
			} else if err := sqlDB.PingContext(ctx); err != nil {
				status.Databases["v1"] = DBHealth{Status: "unhealthy", Error: err.Error()}
				allHealthy = false
			} else {
				status.Databases["v1"] = DBHealth{Status: "healthy", Database: v1DbConfig.Database}
			}
		}

		if !allHealthy {
			status.Status = "unhealthy"
			utils.RespondWithJSON(w, http.StatusServiceUnavailable, status)
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, status)
	})))

	// Debug endpoint
	mux.Handle("/debug", utils.PanicRecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utils.RespondWithJSON(w, http.StatusOK, map[string]string{"path": r.URL.Path, "method": r.Method})
	})))

	// Database debug endpoint
	mux.Handle("/debug/db", utils.PanicRecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		debugInfo := map[string]interface{}{
			"legacy": map[string]interface{}{},
			"v1":     map[string]interface{}{},
		}

		// Test legacy database connection
		if db == nil {
			debugInfo["legacy"] = map[string]interface{}{
				"error": "database connection is nil",
			}
		} else if err := db.PingContext(ctx); err != nil {
			debugInfo["legacy"] = map[string]interface{}{
				"error": fmt.Sprintf("database ping failed: %v", err),
			}
		} else {
			// Check if consumers table exists in legacy DB
			var tableExists bool
			checkTableQuery := `SELECT EXISTS (
			       SELECT FROM information_schema.tables 
			       WHERE table_schema = 'public' 
			       AND table_name = 'consumers'
		       )`

			legacyInfo := map[string]interface{}{
				"status":   "connected",
				"database": dbConfig.Database,
			}

			if err := db.QueryRowContext(ctx, checkTableQuery).Scan(&tableExists); err != nil {
				legacyInfo["table_check_error"] = fmt.Sprintf("failed to check table existence: %v", err)
			} else {
				legacyInfo["consumers_table_exists"] = tableExists
				if tableExists {
					var count int
					countQuery := `SELECT COUNT(*) FROM consumers`
					if err := db.QueryRowContext(ctx, countQuery).Scan(&count); err != nil {
						legacyInfo["count_error"] = fmt.Sprintf("failed to count consumers: %v", err)
					} else {
						legacyInfo["consumers_count"] = count
					}
				}
			}
			debugInfo["legacy"] = legacyInfo
		}

		// Test V1 GORM database connection
		if gormDB == nil {
			debugInfo["v1"] = map[string]interface{}{
				"error": "GORM connection is nil",
			}
		} else {
			sqlDB, err := gormDB.DB()
			if err != nil {
				debugInfo["v1"] = map[string]interface{}{
					"error": fmt.Sprintf("failed to get sql.DB: %v", err),
				}
			} else if err := sqlDB.PingContext(ctx); err != nil {
				debugInfo["v1"] = map[string]interface{}{
					"error": fmt.Sprintf("V1 database ping failed: %v", err),
				}
			} else {
				v1Info := map[string]interface{}{
					"status":   "connected",
					"database": v1DbConfig.Database,
				}

				// Check if entities table exists in V1 DB
				var entitiesExists bool
				checkEntitiesQuery := `SELECT EXISTS (
				       SELECT FROM information_schema.tables 
				       WHERE table_schema = 'public' 
				       AND table_name = 'entities'
			       )`

				if err := sqlDB.QueryRowContext(ctx, checkEntitiesQuery).Scan(&entitiesExists); err != nil {
					v1Info["table_check_error"] = fmt.Sprintf("failed to check entities table: %v", err)
				} else {
					v1Info["entities_table_exists"] = entitiesExists
					if entitiesExists {
						var entityCount int
						countEntitiesQuery := `SELECT COUNT(*) FROM entities`
						if err := sqlDB.QueryRowContext(ctx, countEntitiesQuery).Scan(&entityCount); err != nil {
							v1Info["count_error"] = fmt.Sprintf("failed to count entities: %v", err)
						} else {
							v1Info["entities_count"] = entityCount
						}
					}
				}
				debugInfo["v1"] = v1Info
			}
		}

		utils.RespondWithJSON(w, http.StatusOK, debugInfo)
	})))

	// Setup audit middleware
	auditServiceURL := getEnvOrDefault("AUDIT_SERVICE_URL", "http://localhost:3001")
	auditMiddleware := middleware.NewAuditMiddleware(auditServiceURL)
	handler := auditMiddleware.AuditLoggingMiddleware(mux)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	addr := ":" + port
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		slog.Info("API Server starting", "port", port, "addr", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start API server", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down API Server...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("API Server exited")
}
