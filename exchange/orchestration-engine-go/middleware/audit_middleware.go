package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/shared/redis"
)

// RequestInfo contains analyzed request information for audit logging
type RequestInfo struct {
	ApplicationID string
	SchemaID      string
	RequestType   string // "m2m", "user", "system", "batch"
	AuthMethod    string // "jwt", "apikey", "oauth", "service_account", "none"
	UserID        string
	SessionID     string
}

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

	// Analyze request to determine type and authentication method
	requestInfo := am.analyzeRequest(r)

	// Create audit event with enhanced classification
	eventData := map[string]interface{}{
		"event_id":           uuid.New().String(),
		"consumer_id":        requestInfo.ApplicationID,
		"provider_id":        "orchestration-engine",
		"requested_data":     string(requestBody),
		"response_data":      string(responseBody),
		"transaction_status": status,
		"user_agent":         r.Header.Get("User-Agent"),
		"ip_address":         r.RemoteAddr,
		"timestamp":          time.Now().Unix(),
		// New fields for M2M vs User differentiation
		"request_type":   requestInfo.RequestType,
		"auth_method":    requestInfo.AuthMethod,
		"user_id":        requestInfo.UserID,
		"session_id":     requestInfo.SessionID,
		"application_id": requestInfo.ApplicationID,
		"schema_id":      requestInfo.SchemaID,
	}

	// Send to Redis Stream
	msgID, err := am.redisClient.PublishAuditEvent(ctx, "audit-events", eventData)
	if err != nil {
		logger.Log.Error("Failed to send audit log", "error", err)
		return
	}

	logger.Log.Info("Audit log sent",
		"message_id", msgID,
		"status", status,
		"request_type", requestInfo.RequestType,
		"auth_method", requestInfo.AuthMethod,
		"duration", time.Since(startTime))
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

// analyzeRequest analyzes the HTTP request to determine request type and authentication method
func (am *AuditMiddleware) analyzeRequest(r *http.Request) RequestInfo {
	info := RequestInfo{
		ApplicationID: "unknown",
		SchemaID:      "unknown-schema",
		RequestType:   "unknown",
		AuthMethod:    "none",
		UserID:        "",
		SessionID:     "",
	}

	// Analyze Authorization header
	auth := r.Header.Get("Authorization")
	if auth != "" {
		info.AuthMethod = am.determineAuthMethod(auth)
		info.ApplicationID = am.extractApplicationID(auth)
	}

	// Analyze User-Agent to help determine request type
	userAgent := r.Header.Get("User-Agent")
	info.RequestType = am.determineRequestType(userAgent, auth, r)

	// Extract user context if available
	info.UserID = r.Header.Get("X-User-ID")
	info.SessionID = r.Header.Get("X-Session-ID")

	// Extract schema ID from headers or request
	info.SchemaID = r.Header.Get("X-Schema-ID")
	if info.SchemaID == "" {
		info.SchemaID = "unknown-schema"
	}

	return info
}

// determineAuthMethod analyzes the authorization header to determine authentication method
func (am *AuditMiddleware) determineAuthMethod(auth string) string {
	auth = strings.TrimSpace(auth)

	if strings.HasPrefix(auth, "Bearer ") {
		// JWT token - likely user session
		return "jwt"
	} else if strings.HasPrefix(auth, "ApiKey ") {
		// API key - likely M2M
		return "apikey"
	} else if strings.HasPrefix(auth, "Basic ") {
		// Basic auth - could be M2M or service account
		return "service_account"
	} else if strings.HasPrefix(auth, "OAuth ") {
		// OAuth token
		return "oauth"
	}

	return "unknown"
}

// extractApplicationID extracts application ID from authorization header
func (am *AuditMiddleware) extractApplicationID(auth string) string {
	// Simple extraction - in production, you'd parse JWT claims or API key metadata
	if strings.HasPrefix(auth, "Bearer ") {
		// For JWT, you'd decode and extract from claims
		return "app-from-jwt"
	} else if strings.HasPrefix(auth, "ApiKey ") {
		// For API key, you'd look up the key in a database
		return "app-from-apikey"
	}

	return "app-from-token"
}

// determineRequestType determines if request is M2M, user-initiated, system, or batch
func (am *AuditMiddleware) determineRequestType(userAgent, auth string, r *http.Request) string {
	userAgent = strings.ToLower(userAgent)

	// Check for M2M indicators
	if am.isM2MRequest(userAgent, auth, r) {
		return "m2m"
	}

	// Check for system/batch indicators
	if am.isSystemRequest(userAgent, r) {
		return "system"
	}

	// Check for batch job indicators
	if am.isBatchRequest(userAgent, r) {
		return "batch"
	}

	// Default to user-initiated
	return "user"
}

// isM2MRequest checks if request is machine-to-machine
func (am *AuditMiddleware) isM2MRequest(userAgent, auth string, r *http.Request) bool {
	// M2M indicators:
	// 1. API key authentication
	if strings.HasPrefix(auth, "ApiKey ") {
		return true
	}

	// 2. Service account authentication
	if strings.HasPrefix(auth, "Basic ") {
		return true
	}

	// 3. M2M user agents
	m2mAgents := []string{
		"curl", "wget", "postman", "insomnia", "httpie",
		"python-requests", "go-http-client", "java-http-client",
		"node-fetch", "axios", "okhttp", "apache-httpclient",
	}

	for _, agent := range m2mAgents {
		if strings.Contains(userAgent, agent) {
			return true
		}
	}

	// 4. No user agent (common in M2M)
	if userAgent == "" {
		return true
	}

	// 5. Custom headers indicating M2M
	if r.Header.Get("X-Client-Type") == "system" {
		return true
	}

	return false
}

// isSystemRequest checks if request is from system processes
func (am *AuditMiddleware) isSystemRequest(userAgent string, r *http.Request) bool {
	systemAgents := []string{
		"cron", "systemd", "daemon", "service",
		"health-check", "monitor", "agent",
	}

	for _, agent := range systemAgents {
		if strings.Contains(userAgent, agent) {
			return true
		}
	}

	// Check for system headers
	if r.Header.Get("X-System-Request") == "true" {
		return true
	}

	return false
}

// isBatchRequest checks if request is from batch jobs
func (am *AuditMiddleware) isBatchRequest(userAgent string, r *http.Request) bool {
	batchAgents := []string{
		"batch", "job", "scheduler", "etl",
		"import", "export", "sync",
	}

	for _, agent := range batchAgents {
		if strings.Contains(userAgent, agent) {
			return true
		}
	}

	// Check for batch headers
	if r.Header.Get("X-Batch-Job") == "true" {
		return true
	}

	return false
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
