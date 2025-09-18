package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/handlers"
	"github.com/gov-dx-sandbox/api-server-go/middleware"
	"github.com/gov-dx-sandbox/exchange/shared/utils"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	slog.SetDefault(logger)

	// Initialize API server
	apiServer := handlers.NewAPIServer()

	// Setup routes
	mux := http.NewServeMux()
	apiServer.SetupRoutes(mux)

	// Health check
	mux.Handle("/health", utils.PanicRecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utils.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "healthy", "service": "api-server"})
	})))

	// Debug endpoint
	mux.Handle("/debug", utils.PanicRecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utils.RespondWithJSON(w, http.StatusOK, map[string]string{"path": r.URL.Path, "method": r.Method})
	})))

	// Apply security middleware
	securityHandler := middleware.SecurityHeaders(
		middleware.InputValidation(
			middleware.SecurityLogging(
				middleware.RateLimitMiddleware(100, time.Minute)(mux),
			),
		),
	)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	addr := ":" + port
	slog.Info("API Server starting", "port", port)
	if err := http.ListenAndServe(addr, securityHandler); err != nil {
		slog.Error("Failed to start API server", "error", err)
		os.Exit(1)
	}
}
