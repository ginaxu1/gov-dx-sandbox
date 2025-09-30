package middleware

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/audit-service/services"
)

// AuditMiddleware creates audit log entries for API requests
type AuditMiddleware struct {
	db                     *sql.DB
	piiService             *services.PIIRedactionService
	httpClient             *http.Client
	orchestrationEngineURL string
	consentEngineURL       string
}

// NewAuditMiddleware creates a new audit middleware
func NewAuditMiddleware(db *sql.DB) *AuditMiddleware {
	return &AuditMiddleware{
		db:                     db,
		piiService:             services.NewPIIRedactionService(),
		httpClient:             &http.Client{Timeout: 30 * time.Second},
		orchestrationEngineURL: getEnvOrDefault("ORCHESTRATION_ENGINE_URL", "https://41200aa1-4106-4e6c-babf-311dce37c04a-prod.e1-us-east-azure.choreoapis.dev/opendif-ndx/orchestration-engine/v2"),
		consentEngineURL:       getEnvOrDefault("CONSENT_ENGINE_URL", "http://localhost:8081"),
	}
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// AuditLoggingMiddleware wraps HTTP handlers to automatically create audit logs
func (m *AuditMiddleware) AuditLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip audit logging for audit endpoints themselves to avoid recursion
		if strings.HasPrefix(r.URL.Path, "/audit") {
			next.ServeHTTP(w, r)
			return
		}

		// Skip health checks
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		// Create a response writer that captures the response
		responseWriter := &ResponseCapture{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           &bytes.Buffer{},
		}

		// Read request body
		requestBody, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(requestBody))

		startTime := time.Now()
		next.ServeHTTP(responseWriter, r)
		duration := time.Since(startTime)

		// Extract audit information from the request
		auditInfo := m.extractAuditInfo(r, requestBody, responseWriter.body.Bytes(), responseWriter.statusCode, duration)

		// Only create audit log if we have the required information
		if auditInfo != nil {
			go func() {
				if err := m.createAuditLog(auditInfo); err != nil {
					slog.Error("Failed to create audit log", "error", err, "path", r.URL.Path)
				}
			}()
		}

		// Handle external service calls for specific endpoints
		if m.shouldMakeExternalCall(r) {
			go func() {
				m.handleExternalServiceCall(r, responseWriter.body.Bytes(), responseWriter.statusCode)
			}()
		}
	})
}

// ResponseCapture captures the response for audit logging
type ResponseCapture struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

// Write captures the response body
func (rc *ResponseCapture) Write(b []byte) (int, error) {
	rc.body.Write(b)
	return rc.ResponseWriter.Write(b)
}

// WriteHeader captures the status code
func (rc *ResponseCapture) WriteHeader(statusCode int) {
	rc.statusCode = statusCode
	rc.ResponseWriter.WriteHeader(statusCode)
}

// AuditInfo represents audit information extracted from a request
type AuditInfo struct {
	ConsumerID        string
	ProviderID        string
	RequestedData     []byte
	ResponseData      []byte
	TransactionStatus string
	CitizenHash       string
	Path              string
	Method            string
	Duration          time.Duration
	UserAgent         string
	IPAddress         string
}

// extractAuditInfo extracts audit information from the request and response
func (m *AuditMiddleware) extractAuditInfo(r *http.Request, requestBody, responseBody []byte, statusCode int, duration time.Duration) *AuditInfo {
	// Extract consumer ID
	consumerID := m.extractConsumerID(r)
	if consumerID == "" {
		consumerID = "unknown-consumer"
	}

	// Extract provider ID
	providerID := m.extractProviderID(r)
	if providerID == "" {
		providerID = "api-server"
	}

	// Extract citizen hash from request
	citizenHash := m.extractCitizenHashFromRequest(r)

	// Determine transaction status
	transactionStatus := m.getTransactionStatus(statusCode)

	// Redact PII from request and response data
	redactedRequestData := m.redactPIIFromData(requestBody)
	redactedResponseData := m.redactPIIFromData(responseBody)

	// Extract user agent and IP address
	userAgent := r.Header.Get("User-Agent")
	ipAddress := m.extractIPAddress(r)

	return &AuditInfo{
		ConsumerID:        consumerID,
		ProviderID:        providerID,
		RequestedData:     redactedRequestData,
		ResponseData:      redactedResponseData,
		TransactionStatus: transactionStatus,
		CitizenHash:       citizenHash,
		Path:              r.URL.Path,
		Method:            r.Method,
		Duration:          duration,
		UserAgent:         userAgent,
		IPAddress:         ipAddress,
	}
}

// extractConsumerID extracts consumer ID from the request
func (m *AuditMiddleware) extractConsumerID(r *http.Request) string {
	// Try headers first
	if consumerID := r.Header.Get("X-Consumer-ID"); consumerID != "" {
		return consumerID
	}

	// Try to extract from path
	if strings.Contains(r.URL.Path, "/consumer-applications") {
		return "api-server-consumer"
	}

	return "unknown-consumer"
}

// extractProviderID extracts provider ID from the request
func (m *AuditMiddleware) extractProviderID(r *http.Request) string {
	// Try headers first
	if providerID := r.Header.Get("X-Provider-ID"); providerID != "" {
		return providerID
	}

	// Try to extract from path
	if strings.Contains(r.URL.Path, "/provider-submissions") {
		return "api-server-provider"
	}

	return "api-server"
}

// extractIPAddress extracts the client IP address from the request
func (m *AuditMiddleware) extractIPAddress(r *http.Request) string {
	// Check for forwarded headers first (for load balancers/proxies)
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := strings.Index(ip, ","); idx != -1 {
			ip = ip[:idx]
		}
		return strings.TrimSpace(ip)
	}

	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx] // Remove port number
	}
	return ip
}

// extractCitizenHashFromRequest extracts citizen hash from the request
func (m *AuditMiddleware) extractCitizenHashFromRequest(r *http.Request) string {
	citizenID := r.Header.Get("X-Citizen-ID")
	if citizenID == "" {
		// Try to extract from request body
		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		if len(body) > 0 {
			var data interface{}
			if err := json.Unmarshal(body, &data); err == nil {
				// Use the PII service to extract citizen ID
				citizenID = m.extractCitizenIDFromData(data)
			}
		}
	}

	if citizenID != "" {
		return m.piiService.HashCitizenID(citizenID)
	}

	return "no-citizen-id"
}

// extractCitizenIDFromData extracts citizen ID from data using the PII service
func (m *AuditMiddleware) extractCitizenIDFromData(data interface{}) string {
	// Use reflection to access the private method
	// Since we can't access the private method directly, we'll implement a simple extraction here
	switch v := data.(type) {
	case map[string]interface{}:
		// Look for common citizen ID field names
		citizenIDFields := []string{
			"citizenId", "citizen_id", "provider_id", "consumer_id", "consent_id",
			"providerId", "consumerId", "consentId",
			"userId", "user_id", "ownerId", "owner_id",
			"nationalId", "national_id", "nic", "nicNumber", "nic_number",
		}

		for _, field := range citizenIDFields {
			if value, exists := v[field]; exists {
				if str, ok := value.(string); ok && str != "" {
					return str
				}
			}
		}

		// If not found in common fields, recursively search
		for _, value := range v {
			if id := m.extractCitizenIDFromData(value); id != "" {
				return id
			}
		}

	case []interface{}:
		// Search in array elements
		for _, item := range v {
			if id := m.extractCitizenIDFromData(item); id != "" {
				return id
			}
		}
	}

	return ""
}

// redactPIIFromData redacts PII from JSON data
func (m *AuditMiddleware) redactPIIFromData(data []byte) []byte {
	if len(data) == 0 {
		return data
	}

	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return data
	}

	redactedData, err := m.piiService.RedactData(jsonData)
	if err != nil {
		slog.Warn("Failed to redact PII", "error", err)
		return data
	}

	redactedBytes, err := json.Marshal(redactedData)
	if err != nil {
		slog.Warn("Failed to marshal redacted data", "error", err)
		return data
	}

	return redactedBytes
}

// getTransactionStatus converts HTTP status code to transaction status
func (m *AuditMiddleware) getTransactionStatus(statusCode int) string {
	if statusCode >= 200 && statusCode < 300 {
		return "SUCCESS"
	}
	return "FAILURE"
}

// createAuditLog creates an audit log entry
func (m *AuditMiddleware) createAuditLog(auditInfo *AuditInfo) error {
	// Ensure we have valid JSON data
	var requestedData json.RawMessage
	var responseData json.RawMessage

	// Validate and clean requested data
	if len(auditInfo.RequestedData) > 0 {
		// Try to parse as JSON to validate
		var temp interface{}
		if err := json.Unmarshal(auditInfo.RequestedData, &temp); err != nil {
			// If not valid JSON, wrap it in a simple object
			requestedData = json.RawMessage(fmt.Sprintf(`{"raw_data": %q}`, string(auditInfo.RequestedData)))
		} else {
			requestedData = auditInfo.RequestedData
		}
	} else {
		requestedData = json.RawMessage(`{}`)
	}

	// Validate and clean response data
	if len(auditInfo.ResponseData) > 0 {
		// Try to parse as JSON to validate
		var temp interface{}
		if err := json.Unmarshal(auditInfo.ResponseData, &temp); err != nil {
			// If not valid JSON, wrap it in a simple object
			responseData = json.RawMessage(fmt.Sprintf(`{"raw_data": %q}`, string(auditInfo.ResponseData)))
		} else {
			responseData = auditInfo.ResponseData
		}
	} else {
		responseData = json.RawMessage(`{}`)
	}

	// Create audit log directly in database
	query := `
		INSERT INTO audit_logs (
			event_id, timestamp, consumer_id, provider_id,
			requested_data, response_data, transaction_status, citizen_hash,
			user_agent, ip_address
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`

	// Convert empty IP address to nil for database NULL value
	var ipAddress interface{}
	if auditInfo.IPAddress != "" {
		ipAddress = auditInfo.IPAddress
	}

	_, err := m.db.Exec(
		query,
		uuid.New(),
		time.Now(),
		auditInfo.ConsumerID,
		auditInfo.ProviderID,
		requestedData,
		responseData,
		auditInfo.TransactionStatus,
		auditInfo.CitizenHash,
		auditInfo.UserAgent,
		ipAddress,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	slog.Info("Audit log created",
		"consumer_id", auditInfo.ConsumerID,
		"provider_id", auditInfo.ProviderID,
		"path", auditInfo.Path,
		"method", auditInfo.Method,
		"status", auditInfo.TransactionStatus,
		"duration_ms", auditInfo.Duration.Milliseconds())

	return nil
}

// shouldMakeExternalCall determines if an external service call should be made
func (m *AuditMiddleware) shouldMakeExternalCall(r *http.Request) bool {
	// Only make external calls for successful requests
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		return false
	}

	// Check for specific endpoints that should trigger external calls
	path := r.URL.Path
	return strings.Contains(path, "/consumer-applications") ||
		strings.Contains(path, "/provider-submissions") ||
		strings.Contains(path, "/consents")
}

// handleExternalServiceCall makes calls to external services and creates audit logs
func (m *AuditMiddleware) handleExternalServiceCall(r *http.Request, responseBody []byte, statusCode int) {
	if statusCode < 200 || statusCode >= 300 {
		return // Only make external calls for successful requests
	}

	// Determine which external service to call based on the endpoint
	switch {
	case strings.Contains(r.URL.Path, "/consumer-applications"):
		m.callOrchestrationEngine(r, responseBody)
	case strings.Contains(r.URL.Path, "/provider-submissions"):
		m.callOrchestrationEngine(r, responseBody)
	case strings.Contains(r.URL.Path, "/consents"):
		m.callConsentEngine(r, responseBody)
	}
}

// callOrchestrationEngine makes a call to the orchestration engine
func (m *AuditMiddleware) callOrchestrationEngine(r *http.Request, responseBody []byte) {
	// Extract GraphQL query from the request
	query, variables := m.extractGraphQLFromRequest(r)
	if query == "" {
		return
	}

	// Create GraphQL request for orchestration engine
	graphqlReq := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	reqBody, err := json.Marshal(graphqlReq)
	if err != nil {
		slog.Error("Failed to marshal GraphQL request", "error", err)
		return
	}

	// Make HTTP request to orchestration engine
	req, err := http.NewRequest("POST", m.orchestrationEngineURL, bytes.NewBuffer(reqBody))
	if err != nil {
		slog.Error("Failed to create orchestration engine request", "error", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", r.Header.Get("Authorization"))

	startTime := time.Now()
	resp, err := m.httpClient.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		slog.Error("Failed to call orchestration engine", "error", err)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read orchestration engine response", "error", err)
		return
	}

	// Create audit log for orchestration engine call
	auditInfo := &AuditInfo{
		ConsumerID:        m.extractConsumerID(r),
		ProviderID:        "orchestration-engine",
		RequestedData:     reqBody,
		ResponseData:      respBody,
		TransactionStatus: m.getTransactionStatus(resp.StatusCode),
		CitizenHash:       m.extractCitizenHashFromRequest(r),
		Path:              m.orchestrationEngineURL,
		Method:            "POST",
		Duration:          duration,
		UserAgent:         r.Header.Get("User-Agent"),
		IPAddress:         m.extractIPAddress(r),
	}

	if err := m.createAuditLog(auditInfo); err != nil {
		slog.Error("Failed to create orchestration engine audit log", "error", err)
	}
}

// callConsentEngine makes a call to the consent engine
func (m *AuditMiddleware) callConsentEngine(r *http.Request, responseBody []byte) {
	// Extract consent data from the request
	consentData := m.extractConsentDataFromRequest(r)
	if consentData == nil {
		return
	}

	// Determine the consent engine endpoint
	var endpoint string
	if strings.Contains(r.URL.Path, "/consents/") && r.Method == "PUT" {
		// Extract consent ID from path
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) > 2 {
			consentID := pathParts[2]
			endpoint = fmt.Sprintf("%s/consents/%s", m.consentEngineURL, consentID)
		}
	} else if strings.Contains(r.URL.Path, "/consents") && r.Method == "POST" {
		endpoint = fmt.Sprintf("%s/consents", m.consentEngineURL)
	}

	if endpoint == "" {
		return
	}

	// Make HTTP request to consent engine
	req, err := http.NewRequest(r.Method, endpoint, bytes.NewBuffer(consentData))
	if err != nil {
		slog.Error("Failed to create consent engine request", "error", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", r.Header.Get("Authorization"))

	startTime := time.Now()
	resp, err := m.httpClient.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		slog.Error("Failed to call consent engine", "error", err)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read consent engine response", "error", err)
		return
	}

	// Create audit log for consent engine call
	auditInfo := &AuditInfo{
		ConsumerID:        m.extractConsumerID(r),
		ProviderID:        "consent-engine",
		RequestedData:     consentData,
		ResponseData:      respBody,
		TransactionStatus: m.getTransactionStatus(resp.StatusCode),
		CitizenHash:       m.extractCitizenHashFromRequest(r),
		Path:              endpoint,
		Method:            r.Method,
		Duration:          duration,
		UserAgent:         r.Header.Get("User-Agent"),
		IPAddress:         m.extractIPAddress(r),
	}

	if err := m.createAuditLog(auditInfo); err != nil {
		slog.Error("Failed to create consent engine audit log", "error", err)
	}
}

// extractGraphQLFromRequest extracts GraphQL query and variables from the request
func (m *AuditMiddleware) extractGraphQLFromRequest(r *http.Request) (string, map[string]interface{}) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(r.Header.Get("X-GraphQL-Query")), &data); err != nil {
		// Try to extract from request body
		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(body))
		json.Unmarshal(body, &data)
	}

	query, _ := data["query"].(string)
	variables, _ := data["variables"].(map[string]interface{})

	return query, variables
}

// extractConsentDataFromRequest extracts consent data from the request
func (m *AuditMiddleware) extractConsentDataFromRequest(r *http.Request) []byte {
	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	return body
}
