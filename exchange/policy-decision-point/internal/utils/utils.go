package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// ErrorResponse represents a standard error response structure
type ErrorResponse struct {
	Error string `json:"error"`
}

// RespondWithJSON sends a JSON response with the given status code and data
func RespondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("Failed to encode JSON response", "error", err)
	}
}

// RespondWithError sends a JSON error response
func RespondWithError(w http.ResponseWriter, statusCode int, message string) {
	RespondWithJSON(w, statusCode, ErrorResponse{Error: message})
}

// RespondWithSuccess sends a JSON success response
func RespondWithSuccess(w http.ResponseWriter, statusCode int, data interface{}) {
	RespondWithJSON(w, statusCode, data)
}

// HealthHandler creates a health check handler
func HealthHandler(serviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{
			"service": serviceName,
			"status":  "healthy",
		}
		RespondWithJSON(w, http.StatusOK, response)
	}
}

// PanicRecoveryMiddleware provides panic recovery for HTTP handlers
func PanicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("Handler panicked", "error", err, "path", r.URL.Path)
				RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// ServerConfig holds configuration for HTTP servers
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DefaultServerConfig returns a default server configuration
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Port:         GetEnvOrDefault("PORT", "8080"),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// StartServerWithGracefulShutdown starts an HTTP server with graceful shutdown
func StartServerWithGracefulShutdown(server *http.Server, serviceName string) error {
	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		slog.Info("Shutting down server...", "service", serviceName)
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("Server shutdown error", "error", err, "service", serviceName)
		}
	}()

	slog.Info("Server starting", "service", serviceName, "address", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Server failed to start", "error", err, "service", serviceName)
		return err
	}
	return nil
}

// CreateServer creates an HTTP server with the given configuration
func CreateServer(config *ServerConfig, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf(":%s", config.Port),
		Handler:      handler,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}
}

// GetEnvOrDefault returns the environment variable value or a default
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SetupLogging configures logging based on the configuration
func SetupLogging(format, level string) {
	var handler slog.Handler

	switch format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: getLogLevel(level),
		})
	default:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: getLogLevel(level),
		})
	}

	slog.SetDefault(slog.New(handler))
}

// getLogLevel converts string level to slog.Level
func getLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

