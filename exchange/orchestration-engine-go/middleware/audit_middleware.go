package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/shared/redis"
)

// AuditMiddleware handles audit logging for requests
type AuditMiddleware struct {
	redisClient *redis.RedisClient
}

// NewAuditMiddleware creates a new audit middleware
func NewAuditMiddleware(schemaService interface{}) *AuditMiddleware {
	// Connect to Redis
	redisClient, err := redis.NewClient(&redis.Config{
		Addr:     getEnvOrDefault("REDIS_ADDR", "localhost:6379"),
		Password: getEnvOrDefault("REDIS_PASSWORD", ""),
		DB:       0,
	})
	if err != nil {
		logger.Log.Warn("Failed to connect to Redis, audit middleware will be disabled", "error", err)
		redisClient = nil
	} else {
		logger.Log.Info("Audit middleware connected to Redis")
	}

	return &AuditMiddleware{
		redisClient: redisClient,
	}
}

// AuditHandler wraps an http.HandlerFunc with audit logging
func (am *AuditMiddleware) AuditHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip audit for health checks and non-POST requests
		if r.URL.Path == "/health" || r.Method != http.MethodPost {
			next(w, r)
			return
		}

		// Read request body
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Log.Error("Failed to read request body", "error", err)
			next(w, r)
			return
		}

		// Restore request body
		r.Body = io.NopCloser(io.Reader(bytes.NewBuffer(bodyBytes)))

		// Create response writer to capture response
		responseWriter := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           &bytes.Buffer{},
		}

		// Track start time
		startTime := time.Now()

		// Call next handler
		next(responseWriter, r)

		// Determine status
		status := "success"
		if responseWriter.statusCode >= 400 {
			status = "failure"
		}

		// Send audit log asynchronously
		if am.redisClient != nil {
			go am.sendAuditLog(r, bodyBytes, responseWriter.body.Bytes(), status, startTime)
		}

		// Write response to original writer
		w.WriteHeader(responseWriter.statusCode)
		w.Write(responseWriter.body.Bytes())
	}
}

// sendAuditLog sends audit log to Redis Stream
func (am *AuditMiddleware) sendAuditLog(r *http.Request, requestBody, responseBody []byte, status string, startTime time.Time) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Extract application ID from Authorization header or use default
	applicationID := "unknown"
	if auth := r.Header.Get("Authorization"); auth != "" {
		// Simple extraction - in real implementation, you'd parse the JWT
		applicationID = "app-from-token"
	}

	// Create audit event
	eventData := map[string]interface{}{
		"event_id":           uuid.New().String(),
		"consumer_id":        applicationID,
		"provider_id":        "orchestration-engine",
		"requested_data":     string(requestBody),
		"response_data":      string(responseBody),
		"transaction_status": status,
		"user_agent":         r.Header.Get("User-Agent"),
		"ip_address":         r.RemoteAddr,
		"timestamp":          time.Now().Unix(),
	}

	// Send to Redis Stream
	msgID, err := am.redisClient.PublishAuditEvent(ctx, "audit-events", eventData)
	if err != nil {
		logger.Log.Error("Failed to send audit log", "error", err)
		return
	}

	logger.Log.Info("Audit log sent", "message_id", msgID, "status", status, "duration", time.Since(startTime))
}

// responseWriter captures response data
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.body.Write(b)
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
