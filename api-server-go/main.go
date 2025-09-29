package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/handlers"
	"github.com/gov-dx-sandbox/api-server-go/middleware"
	"github.com/gov-dx-sandbox/api-server-go/shared/utils"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	slog.SetDefault(logger)

	slog.Info("Starting API Server initialization")

	// Initialize database connection
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

	// Initialize database tables
	if err := InitDatabase(db); err != nil {
		slog.Error("Failed to initialize database tables", "error", err)
		os.Exit(1)
	}

	// Initialize API server with database
	apiServer := handlers.NewAPIServerWithDB(db)

	// Setup routes
	mux := http.NewServeMux()
	apiServer.SetupRoutes(mux)

	// Health check endpoint (matching consent-engine approach)
	mux.Handle("/health", utils.PanicRecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simple health check - just verify database connection
		if db == nil {
			utils.RespondWithJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status":  "unhealthy",
				"service": "api-server",
				"error":   "database connection is nil",
			})
			return
		}

		// Test database connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			utils.RespondWithJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status":  "unhealthy",
				"service": "api-server",
				"error":   err.Error(),
			})
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, map[string]string{
			"status":  "healthy",
			"service": "api-server",
		})
	})))

	// Debug endpoint
	mux.Handle("/debug", utils.PanicRecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utils.RespondWithJSON(w, http.StatusOK, map[string]string{"path": r.URL.Path, "method": r.Method})
	})))

	// Database debug endpoint
	mux.Handle("/debug/db", utils.PanicRecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			utils.RespondWithJSON(w, http.StatusServiceUnavailable, map[string]string{
				"error": "database connection is nil",
			})
			return
		}

		// Test database connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			utils.RespondWithJSON(w, http.StatusServiceUnavailable, map[string]string{
				"error": fmt.Sprintf("database ping failed: %v", err),
			})
			return
		}

		// Check if consumers table exists
		var tableExists bool
		checkTableQuery := `SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'consumers'
		)`

		if err := db.QueryRowContext(ctx, checkTableQuery).Scan(&tableExists); err != nil {
			utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("failed to check table existence: %v", err),
			})
			return
		}

		// Get table count if it exists
		var count int
		if tableExists {
			countQuery := `SELECT COUNT(*) FROM consumers`
			if err := db.QueryRowContext(ctx, countQuery).Scan(&count); err != nil {
				utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{
					"error": fmt.Sprintf("failed to count consumers: %v", err),
				})
				return
			}
		}

		// Check actual table structure
		var tableStructure string
		if tableExists {
			structureQuery := `SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'consumers' ORDER BY ordinal_position`
			rows, err := db.QueryContext(ctx, structureQuery)
			if err != nil {
				tableStructure = fmt.Sprintf("Failed to get table structure: %v", err)
			} else {
				defer rows.Close()
				var columns []string
				for rows.Next() {
					var colName, dataType string
					if err := rows.Scan(&colName, &dataType); err == nil {
						columns = append(columns, fmt.Sprintf("%s (%s)", colName, dataType))
					}
				}
				tableStructure = fmt.Sprintf("Columns: %s", strings.Join(columns, ", "))
			}
		}

		// Test the actual SELECT query that's failing
		var testQueryError string
		if tableExists {
			testQuery := `SELECT c.consumer_id, e.entity_name, e.contact_email, e.phone_number, c.created_at, c.updated_at FROM consumers c JOIN entities e ON c.entity_id = e.entity_id ORDER BY c.created_at DESC`
			rows, err := db.QueryContext(ctx, testQuery)
			if err != nil {
				testQueryError = fmt.Sprintf("SELECT query failed: %v", err)
			} else {
				rows.Close()
				testQueryError = "SELECT query succeeded"
			}
		}

		utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"database_connected":     true,
			"consumers_table_exists": tableExists,
			"consumers_count":        count,
			"table_structure":        tableStructure,
			"select_query_test":      testQueryError,
		})
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
