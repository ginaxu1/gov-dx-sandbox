package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

// SuccessResponse represents a standard success response structure
type SuccessResponse struct {
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
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

// JSONHandler handles JSON request/response with error handling
func JSONHandler(w http.ResponseWriter, r *http.Request, req interface{}, handler func() (interface{}, int, error)) {
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		slog.Error("Failed to decode request body", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid JSON input")
		return
	}

	data, statusCode, err := handler()
	if err != nil {
		slog.Error("Handler failed", "error", err)
		RespondWithError(w, statusCode, err.Error())
		return
	}

	RespondWithJSON(w, statusCode, data)
}

// PathHandler handles path-based requests with parameter extraction
func PathHandler(w http.ResponseWriter, r *http.Request, prefix string, handler func(string) (interface{}, int, error)) {
	param := r.URL.Path[len(prefix):]
	if param == "" {
		RespondWithError(w, http.StatusBadRequest, "Parameter is required")
		return
	}

	data, statusCode, err := handler(param)
	if err != nil {
		slog.Error("Handler failed", "error", err)
		RespondWithError(w, statusCode, err.Error())
		return
	}

	RespondWithJSON(w, statusCode, data)
}

// GenericHandler handles generic requests without specific parameter extraction
func GenericHandler(w http.ResponseWriter, r *http.Request, handler func() (interface{}, int, error)) {
	data, statusCode, err := handler()
	if err != nil {
		slog.Error("Handler failed", "error", err)
		RespondWithError(w, statusCode, err.Error())
		return
	}

	RespondWithJSON(w, statusCode, data)
}

// Helper functions for common patterns
func HandleError(w http.ResponseWriter, err error, statusCode int, operation string) {
	slog.Error("Operation failed", "operation", operation, "error", err)
	RespondWithError(w, statusCode, fmt.Sprintf("failed to %s: %v", operation, err))
}

func HandleSuccess(w http.ResponseWriter, data interface{}, statusCode int, operation string, logData map[string]interface{}) {
	// Convert map to key-value pairs for slog
	args := make([]interface{}, 0, len(logData)*2+2)
	args = append(args, "operation", operation)
	for k, v := range logData {
		args = append(args, k, v)
	}
	slog.Info("Operation successful", args...)
	RespondWithSuccess(w, statusCode, data)
}

func ValidateMethod(w http.ResponseWriter, r *http.Request, allowedMethod string) bool {
	if r.Method != allowedMethod {
		w.Header().Set("Allow", allowedMethod)
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return false
	}
	return true
}

func ExtractIDFromPath(r *http.Request, prefix string) (string, error) {
	id := r.URL.Path[len(prefix):]
	if id == "" {
		return "", fmt.Errorf("ID is required")
	}
	return id, nil
}

func ExtractQueryParam(r *http.Request, param string) (string, error) {
	value := r.URL.Query().Get(param)
	if value == "" {
		return "", fmt.Errorf("%s is required", param)
	}
	return value, nil
}

// Additional utility functions from individual utils packages

// ExtractIDFromPathString extracts the ID from a URL path by taking the last segment
func ExtractIDFromPathString(path string) string {
	path = strings.TrimSuffix(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// ExtractProviderIDFromPath extracts the provider ID from a URL path like /providers/{id}/schemas
func ExtractProviderIDFromPath(path string) string {
	path = strings.TrimSuffix(path, "/")
	parts := strings.Split(path, "/")

	// Look for "providers" and return the next segment
	for i, part := range parts {
		if part == "providers" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// ExtractConsumerIDFromPath extracts the consumer ID from a URL path like /consumers/{id}/applications
func ExtractConsumerIDFromPath(path string) string {
	path = strings.TrimSuffix(path, "/")
	parts := strings.Split(path, "/")

	// Look for "consumers" and return the next segment
	for i, part := range parts {
		if part == "consumers" && i+1 < len(parts) {
			return parts[i+1]
		}
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

// ReadRequestBody reads the entire request body as bytes
func ReadRequestBody(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}
