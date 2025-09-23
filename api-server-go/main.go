package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/handlers"
	"github.com/gov-dx-sandbox/api-server-go/pkg/database"
	"github.com/gov-dx-sandbox/api-server-go/shared/utils"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	slog.SetDefault(logger)

	slog.Info("Starting API Server initialization")

	// Initialize database connection
	dbConfig := database.NewDatabaseConfig()
	db, err := database.ConnectDB(dbConfig)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := database.GracefulShutdown(db); err != nil {
			slog.Error("Error during database graceful shutdown", "error", err)
		}
	}()

	// Initialize database tables
	if err := database.InitDatabase(db); err != nil {
		slog.Error("Failed to initialize database tables", "error", err)
		os.Exit(1)
	}

	// Initialize API server with database
	apiServer := handlers.NewAPIServerWithDB(db)

	// Setup routes
	mux := http.NewServeMux()
	apiServer.SetupRoutes(mux)

	// Health check with database status
	mux.Handle("/health", utils.PanicRecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Perform database health check
		if err := database.HealthCheck(db); err != nil {
			slog.Error("Health check failed", "error", err)
			utils.RespondWithJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
				"status":   "unhealthy",
				"service":  "api-server",
				"database": "unavailable",
				"error":    err.Error(),
			})
			return
		}

		// Get connection pool stats
		poolStats := database.GetConnectionPoolStats(db)

		utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"status":          "healthy",
			"service":         "api-server",
			"database":        "available",
			"connection_pool": poolStats,
		})
	})))

	// Debug endpoint
	mux.Handle("/debug", utils.PanicRecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utils.RespondWithJSON(w, http.StatusOK, map[string]string{"path": r.URL.Path, "method": r.Method})
	})))

	// Connection pool monitoring endpoint
	mux.Handle("/metrics/db", utils.PanicRecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		poolStats := database.GetConnectionPoolStats(db)
		utils.RespondWithJSON(w, http.StatusOK, poolStats)
	})))

	// Cloud database health endpoint (optimized for Choreo/Aiven)
	mux.Handle("/health/db", utils.PanicRecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		if err := database.HealthCheck(db); err != nil {
			duration := time.Since(start)
			slog.Error("Database health check failed", "error", err, "duration", duration)
			utils.RespondWithJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
				"status":    "unhealthy",
				"database":  "choreo_postgresql",
				"error":     err.Error(),
				"duration":  duration.String(),
				"timestamp": time.Now().Unix(),
			})
			return
		}

		poolStats := database.GetConnectionPoolStats(db)
		utilization := float64(poolStats.InUse) / float64(poolStats.MaxOpenConns) * 100
		duration := time.Since(start)

		status := "healthy"
		if utilization > 90 {
			status = "critical"
		} else if utilization > 70 {
			status = "warning"
		}

		utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"status":              status,
			"database":            "choreo_postgresql",
			"duration":            duration.String(),
			"timestamp":           time.Now().Unix(),
			"connection_pool":     poolStats,
			"utilization_percent": utilization,
		})
	})))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	addr := ":" + port
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start periodic connection pool monitoring
	go func() {
		ticker := time.NewTicker(5 * time.Minute) // Log every 5 minutes
		defer ticker.Stop()

		for range ticker.C {
			database.LogConnectionPoolStats(db)
		}
	}()

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
