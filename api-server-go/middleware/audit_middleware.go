package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/models"
	"github.com/gov-dx-sandbox/api-server-go/services"
)

// AuditMiddleware wraps HTTP handlers to capture and send audit logs
type AuditMiddleware struct {
	auditService *services.AuditService
}

// NewAuditMiddleware creates a new audit middleware
func NewAuditMiddleware(auditServiceURL string) *AuditMiddleware {
	return &AuditMiddleware{
		auditService: services.NewAuditService(auditServiceURL),
	}
}

// AuditLoggingMiddleware returns a middleware function that logs all requests
func (m *AuditMiddleware) AuditLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip audit logging for health checks and debug endpoints
		if m.shouldSkipAudit(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Create audit context
		auditCtx := models.NewAuditContext()
		auditCtx.UserAgent = r.Header.Get("User-Agent")
		auditCtx.IPAddress = m.getClientIP(r)

		// Extract entity IDs from path
		auditCtx.ConsumerID = m.auditService.ExtractConsumerIDFromPath(r.URL.Path)
		auditCtx.ProviderID = m.auditService.ExtractProviderIDFromPath(r.URL.Path)

		// If no specific entity ID found, use a default based on the endpoint
		if auditCtx.ConsumerID == "" && auditCtx.ProviderID == "" {
			auditCtx.ConsumerID = m.determineDefaultEntityID(r.URL.Path)
		}

		// Capture request body
		var requestBody []byte
		if r.Body != nil {
			requestBody, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(requestBody)) // Restore body for next handler
		}
		auditCtx.RequestData = requestBody

		// Create response writer wrapper to capture response
		responseWrapper := &responseWriter{
			ResponseWriter: w,
			body:           &bytes.Buffer{},
			statusCode:     http.StatusOK, // Default to 200 OK
		}

		// Process the request
		startTime := time.Now()
		next.ServeHTTP(responseWrapper, r)
		duration := time.Since(startTime)

		// Capture response data
		auditCtx.ResponseData = responseWrapper.body.Bytes()
		auditCtx.Status = m.auditService.DetermineTransactionStatus(responseWrapper.statusCode)
		auditCtx.EndTime = time.Now()

		// Ensure we have valid JSON for request and response data
		if len(auditCtx.RequestData) == 0 {
			auditCtx.RequestData = []byte("{}")
		}
		if len(auditCtx.ResponseData) == 0 {
			auditCtx.ResponseData = []byte("{}")
		}

		// Log the request
		slog.Info("Request processed",
			"method", r.Method,
			"path", r.URL.Path,
			"status_code", responseWrapper.statusCode,
			"duration_ms", duration.Milliseconds(),
			"event_id", auditCtx.EventID,
			"consumer_id", auditCtx.ConsumerID,
			"provider_id", auditCtx.ProviderID)

		// Send audit log asynchronously
		auditReq := auditCtx.ToAuditLogRequest()
		m.auditService.SendAuditLogAsync(auditReq)
	})
}

// shouldSkipAudit determines if audit logging should be skipped for this path
func (m *AuditMiddleware) shouldSkipAudit(path string) bool {
	// Skip audit logging in test environment
	if os.Getenv("GO_ENV") == "test" || os.Getenv("TESTING") == "true" {
		return true
	}

	skipPaths := []string{
		"/health",
		"/debug",
		"/openapi.yaml",
		"/favicon.ico",
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// getClientIP extracts the client IP address from the request
func (m *AuditMiddleware) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for load balancers/proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fallback to RemoteAddr
	if r.RemoteAddr != "" {
		// RemoteAddr is in format "IP:port", extract just the IP
		if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
			return r.RemoteAddr[:idx]
		}
		return r.RemoteAddr
	}

	return "unknown"
}

// determineDefaultEntityID determines a default entity ID based on the endpoint
func (m *AuditMiddleware) determineDefaultEntityID(path string) string {
	if strings.Contains(path, "/consumers") || strings.Contains(path, "/consumer-applications") {
		return "system_consumer"
	}
	if strings.Contains(path, "/providers") || strings.Contains(path, "/provider-submissions") {
		return "system_provider"
	}
	return "system_admin"
}

// responseWriter wraps http.ResponseWriter to capture response data
type responseWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

// Write captures the response body
func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
