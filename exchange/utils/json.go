// Package utils provides common utility functions for the project
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

// ErrorResponse defines the structure for a standard JSON error message
type ErrorResponse struct {
	Error string `json:"error"`
}

// RespondWithJSON is a utility function to write a JSON response
// It sets the Content-Type header, writes the HTTP status code, and encodes the payload
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
	}
}

// RespondWithError is a utility function to write a JSON error response
func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, ErrorResponse{Error: message})
}

// RespondWithSuccess is a utility function to write a JSON success response
func RespondWithSuccess(w http.ResponseWriter, code int, data interface{}) {
	RespondWithJSON(w, code, data)
}

// ExtractIDFromPath extracts the ID from a URL path by taking the last segment
func ExtractIDFromPath(path string) string {
	path = strings.TrimSuffix(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// ParseJSONRequest parses a JSON request body into the target struct
func ParseJSONRequest(r *http.Request, target interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

// CreateCollectionResponse creates a standardized collection response with count
func CreateCollectionResponse(items interface{}, count int) map[string]interface{} {
	return map[string]interface{}{
		"items": items,
		"count": count,
	}
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

// HealthHandler creates a standard health check handler
func HealthHandler(serviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			RespondWithJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
			return
		}
		RespondWithJSON(w, http.StatusOK, map[string]string{
			"status":  "healthy",
			"service": serviceName,
		})
	}
}

// GetEnvOrDefault returns the environment variable value or a default
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ParseExpiryTime parses expiry time strings like "30d", "1h", "7d"
func ParseExpiryTime(expiryStr string) (time.Duration, error) {
	if len(expiryStr) < 2 {
		return 0, fmt.Errorf("invalid expiry time format")
	}

	unit := expiryStr[len(expiryStr)-1:]
	value := expiryStr[:len(expiryStr)-1]

	var duration time.Duration
	switch unit {
	case "d":
		duration = 24 * time.Hour
	case "h":
		duration = time.Hour
	case "m":
		duration = time.Minute
	case "s":
		duration = time.Second
	default:
		return 0, fmt.Errorf("unsupported time unit: %s", unit)
	}

	// Parse the numeric value
	var multiplier int
	if _, err := fmt.Sscanf(value, "%d", &multiplier); err != nil {
		return 0, fmt.Errorf("invalid numeric value: %s", value)
	}

	return time.Duration(multiplier) * duration, nil
}
