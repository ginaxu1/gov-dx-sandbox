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
