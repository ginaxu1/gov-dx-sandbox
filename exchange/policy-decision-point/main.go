package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	v1 "github.com/gov-dx-sandbox/exchange/policy-decision-point/v1"
	"github.com/gov-dx-sandbox/exchange/shared/utils"
	"github.com/joho/godotenv"
)

// Build information - set during build
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Load .env file if it exists (optional - fails silently if not found)
	_ = godotenv.Load()

	// Setup logging
	utils.SetupLogging("json", getEnvOrDefault("LOG_LEVEL", "info"))

	slog.Info("Starting policy decision point (V1)",
		"version", Version,
		"build_time", BuildTime,
		"git_commit", GitCommit)

	// Log database configuration being used
	slog.Info("Database configuration",
		"choreo_host", os.Getenv("CHOREO_OPENDIF_DATABASE_HOSTNAME"),
		"choreo_port", os.Getenv("CHOREO_OPENDIF_DATABASE_PORT"),
		"choreo_user", os.Getenv("CHOREO_OPENDIF_DATABASE_USERNAME"),
		"choreo_database", os.Getenv("CHOREO_OPENDIF_DATABASE_DATABASENAME"),
		"sslmode", os.Getenv("DB_SSLMODE"),
	)

	// Initialize V1 GORM database connection
	v1DbConfig := v1.NewDatabaseConfig()
	gormDB, err := v1.ConnectGormDB(v1DbConfig)
	if err != nil {
		slog.Error("Failed to connect to GORM database", "error", err)
		os.Exit(1)
	}

	// Initialize V1 handlers
	v1Handler := v1.NewHandler(gormDB)

	// Setup routes
	mux := http.NewServeMux()
	v1Handler.SetupRoutes(mux) // V1 routes with /api/v1/policy/ prefix

	// Health check endpoint
	mux.Handle("/health", utils.PanicRecoveryMiddleware(utils.HealthHandler("policy-decision-point")))

	// Debug endpoint
	mux.Handle("/debug", utils.PanicRecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utils.RespondWithJSON(w, http.StatusOK, map[string]string{
			"service": "policy-decision-point",
			"version": Version,
			"path":    r.URL.Path,
			"method":  r.Method,
		})
	})))

	// Database debug endpoint
	mux.Handle("/debug/db", utils.PanicRecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		debugInfo := map[string]interface{}{
			"service": "policy-decision-point",
			"v1":      map[string]interface{}{},
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
					"error": "failed to get sql.DB: " + err.Error(),
				}
			} else if err := sqlDB.PingContext(ctx); err != nil {
				debugInfo["v1"] = map[string]interface{}{
					"error": "database ping failed: " + err.Error(),
				}
			} else {
				v1Info := map[string]interface{}{
					"status":   "connected",
					"database": v1DbConfig.Database,
				}

				// Check if policy_metadata table exists
				var tableExists bool
				checkTableQuery := `SELECT EXISTS (
				       SELECT FROM information_schema.tables 
				       WHERE table_schema = 'public' 
				       AND table_name = 'policy_metadata'
			       )`

				if err := sqlDB.QueryRowContext(ctx, checkTableQuery).Scan(&tableExists); err != nil {
					v1Info["table_check_error"] = "failed to check policy_metadata table: " + err.Error()
				} else {
					v1Info["policy_metadata_table_exists"] = tableExists
					if tableExists {
						var count int
						countQuery := `SELECT COUNT(*) FROM policy_metadata`
						if err := sqlDB.QueryRowContext(ctx, countQuery).Scan(&count); err != nil {
							v1Info["count_error"] = "failed to count policy_metadata: " + err.Error()
						} else {
							v1Info["policy_metadata_count"] = count
						}
					}
				}
				debugInfo["v1"] = v1Info
			}
		}

		utils.RespondWithJSON(w, http.StatusOK, debugInfo)
	})))

	// Create server using utils
	port := getEnvOrDefault("PORT", "8082")
	serverConfig := &utils.ServerConfig{
		Port:         port,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	server := utils.CreateServer(serverConfig, mux)

	// Start server with graceful shutdown
	if err := utils.StartServerWithGracefulShutdown(server, "policy-decision-point"); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}

	// Cleanup database connection on shutdown
	defer func() {
		if gormDB != nil {
			if sqlDB, err := gormDB.DB(); err == nil {
				if err := sqlDB.Close(); err != nil {
					slog.Error("Failed to close database connection", "error", err)
				}
			}
		}
	}()
}

// getEnvOrDefault gets an environment variable with a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
