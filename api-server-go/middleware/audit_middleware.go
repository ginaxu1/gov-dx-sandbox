package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/models"
	"github.com/gov-dx-sandbox/api-server-go/services"
	"github.com/gov-dx-sandbox/api-server-go/shared/utils"
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

		// Extract all information from request with single body read
		consumerID, providerID, graphqlQuery, requestBody, err := m.auditService.ExtractAllFromRequestWithBody(r)
		if err != nil {
			// If body reading fails, fall back to individual extraction methods
			auditCtx.ConsumerID = m.auditService.ExtractConsumerIDFromRequest(r)
			auditCtx.ProviderID = m.auditService.ExtractProviderIDFromRequest(r)
		} else {
			auditCtx.ConsumerID = consumerID
			auditCtx.ProviderID = providerID
		}

		// Extract applicationId and schemaId from request path (required by audit service)
		auditCtx.ApplicationID = m.extractApplicationIDFromPath(r.URL.Path)
		auditCtx.SchemaID = m.extractSchemaIDFromPath(r.URL.Path)

		// If not found in path, try to extract from request body
		if auditCtx.ApplicationID == "" && len(requestBody) > 0 {
			auditCtx.ApplicationID = m.extractApplicationIDFromBody(requestBody)
		}
		if auditCtx.SchemaID == "" && len(requestBody) > 0 {
			auditCtx.SchemaID = m.extractSchemaIDFromBody(requestBody)
		}

		// If no specific entity ID found, use a default based on the endpoint
		if auditCtx.ConsumerID == "" && auditCtx.ProviderID == "" {
			auditCtx.ConsumerID = m.determineDefaultEntityID(r.URL.Path)
		}

		// Set request data
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

		// Capture response data and ensure it's valid JSON
		responseBytes := responseWrapper.body.Bytes()
		auditCtx.ResponseData = m.ensureValidJSON(responseBytes)
		auditCtx.Status = m.auditService.DetermineTransactionStatus(responseWrapper.statusCode)
		auditCtx.EndTime = time.Now()

		// Use pre-extracted GraphQL query if available
		if graphqlQuery != "" {
			// If we found a GraphQL query, use it as the requested data
			auditCtx.RequestData = []byte(graphqlQuery)
		} else {
			// Ensure we have valid JSON for request and response data
			if len(auditCtx.RequestData) == 0 {
				auditCtx.RequestData = []byte("{}")
			} else {
				auditCtx.RequestData = m.ensureValidJSON(auditCtx.RequestData)
			}
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

// ensureValidJSON ensures the data is valid JSON, converting non-JSON to a JSON string
func (m *AuditMiddleware) ensureValidJSON(data []byte) []byte {
	if len(data) == 0 {
		return []byte("{}")
	}

	// Check if it's already valid JSON
	var js interface{}
	if json.Unmarshal(data, &js) == nil {
		return data
	}

	// If not valid JSON, wrap it as a string in a JSON object
	wrappedData := map[string]string{
		"raw_data": string(data),
	}

	jsonData, err := json.Marshal(wrappedData)
	if err != nil {
		// Fallback to empty JSON object if marshaling fails
		return []byte("{}")
	}

	return jsonData
}

// extractApplicationIDFromPath extracts applicationId from request path
func (m *AuditMiddleware) extractApplicationIDFromPath(path string) string {
	// Extract application ID from paths like /api/v1/applications/{applicationId}
	if strings.Contains(path, "/api/v1/applications/") {
		parts := strings.Split(path, "/")
		for i, part := range parts {
			if part == "applications" && i+1 < len(parts) {
				appID := parts[i+1]
				// Remove any trailing path segments (e.g., /api/v1/applications/{id}/submissions)
				if idx := strings.Index(appID, "/"); idx != -1 {
					appID = appID[:idx]
				}
				return appID
			}
		}
	}
	return ""
}

// extractSchemaIDFromPath extracts schemaId from request path
func (m *AuditMiddleware) extractSchemaIDFromPath(path string) string {
	// Extract schema ID from paths like /api/v1/schemas/{schemaId}
	if strings.Contains(path, "/api/v1/schemas/") {
		parts := strings.Split(path, "/")
		for i, part := range parts {
			if part == "schemas" && i+1 < len(parts) {
				schemaID := parts[i+1]
				// Remove any trailing path segments
				if idx := strings.Index(schemaID, "/"); idx != -1 {
					schemaID = schemaID[:idx]
				}
				return schemaID
			}
		}
	}
	return ""
}

// extractApplicationIDFromBody extracts applicationId from request body
func (m *AuditMiddleware) extractApplicationIDFromBody(body []byte) string {
	return utils.ExtractApplicationIDFromJSON(body)
}

// extractSchemaIDFromBody extracts schemaId from request body
func (m *AuditMiddleware) extractSchemaIDFromBody(body []byte) string {
	return utils.ExtractSchemaIDFromJSON(body)
}
