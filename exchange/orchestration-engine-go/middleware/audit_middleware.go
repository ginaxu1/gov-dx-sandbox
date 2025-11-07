package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/auth"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
)

// AuditMiddleware handles audit logging for requests
type AuditMiddleware struct {
	environment     string
	auditServiceURL string
	httpClient      *http.Client
	schemaService   interface{} // Schema service interface
}

// NewAuditMiddleware creates a new audit middleware
func NewAuditMiddleware(
	environment,
	auditServiceURL string, schemaService interface{}) *AuditMiddleware {
	return &AuditMiddleware{
		environment:     environment,
		auditServiceURL: auditServiceURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		schemaService: schemaService,
	}
}

// AuditLogRequest represents the request structure for audit service
type AuditLogRequest struct {
	Status        string `json:"status"`
	RequestedData string `json:"requestedData"`
	ApplicationID string `json:"applicationId"`
	SchemaID      string `json:"schemaId"`
}

// AuditLogResponse represents the response from audit service
type AuditLogResponse struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
}

// AuditHandler wraps an http.HandlerFunc with audit logging
func (am *AuditMiddleware) AuditHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read the request body
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Log.Error("Failed to read request body", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Restore the request body for the next handler
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Create a response writer wrapper to capture the response
		responseWrapper := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           &bytes.Buffer{},
		}

		// Track start time
		startTime := time.Now()

		// Call the next handler
		next(responseWrapper, r)

		// Determine success/failure based on status code
		status := "success"
		if responseWrapper.statusCode >= 400 {
			status = "failure"
		}

		// Extract application_id and schema_id from the request
		applicationID, schemaID := am.ExtractAuditInfo(r, bodyBytes)

		// Prepare audit log request
		auditRequest := AuditLogRequest{
			Status:        status,
			RequestedData: string(bodyBytes),
			ApplicationID: applicationID,
			SchemaID:      schemaID,
		}

		// Send audit log asynchronously (don't block the response)
		go am.sendAuditLog(auditRequest, startTime)

		// Write the response to the original writer
		w.WriteHeader(responseWrapper.statusCode)
		w.Write(responseWrapper.body.Bytes())
	}
}

// sendAuditLog sends the audit log to the audit service
func (am *AuditMiddleware) sendAuditLog(auditRequest AuditLogRequest, startTime time.Time) {
	// Marshal the audit request
	requestBody, err := json.Marshal(auditRequest)
	if err != nil {
		logger.Log.Error("Failed to marshal audit request", "error", err)
		return
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", am.auditServiceURL+"/api/logs", bytes.NewBuffer(requestBody))
	if err != nil {
		logger.Log.Error("Failed to create audit request", "error", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := am.httpClient.Do(req)
	if err != nil {
		logger.Log.Error("Failed to send audit log", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		logger.Log.Error("Audit service returned error", "status", resp.StatusCode)
		return
	}

	// Log successful audit
	logger.Log.Info("Audit log sent successfully",
		"status", auditRequest.Status,
		"duration", time.Since(startTime),
		"audit_service_status", resp.StatusCode)
}

// responseWriter wraps http.ResponseWriter to capture response details
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

// extractAuditInfo extracts application_id and schema_id from the request
func (am *AuditMiddleware) ExtractAuditInfo(r *http.Request, bodyBytes []byte) (string, string) {
	// Extract application_id from consumer_applications table
	applicationID := am.getApplicationIDFromConsumer(r)

	// Get the currently active schema ID from the orchestration engine
	schemaID := am.getActiveSchemaID()

	return applicationID, schemaID
}

// getActiveSchemaID fetches the currently active schema ID from the schema service
func (am *AuditMiddleware) getActiveSchemaID() string {
	// Try to get the active schema from the schema service
	if am.schemaService != nil {
		// Use reflection to call GetActiveSchema method
		schemaServiceValue := reflect.ValueOf(am.schemaService)
		if schemaServiceValue.IsValid() {
			// Only call IsNil() on pointer types
			if schemaServiceValue.Kind() == reflect.Ptr && schemaServiceValue.IsNil() {
				// It's a nil pointer, skip
				logger.Log.Warn("Schema service is nil, using fallback")
				return "unknown-schema"
			}
			getActiveSchemaMethod := schemaServiceValue.MethodByName("GetActiveSchema")
			if getActiveSchemaMethod.IsValid() {
				results := getActiveSchemaMethod.Call([]reflect.Value{})
				if len(results) >= 2 && !results[1].IsNil() {
					// Error occurred, use default
					logger.Log.Warn("Failed to get active schema from service", "Error", results[1].Interface())
				} else if len(results) >= 1 && !results[0].IsNil() {
					// Got schema from service
					schemaRecord := results[0].Interface()
					// Extract ID from schema record using reflection
					schemaRecordValue := reflect.ValueOf(schemaRecord)
					// If it's a pointer, dereference it
					if schemaRecordValue.Kind() == reflect.Ptr {
						schemaRecordValue = schemaRecordValue.Elem()
					}
					idField := schemaRecordValue.FieldByName("ID")
					if idField.IsValid() && idField.Kind() == reflect.String {
						return idField.String()
					}
				}
			}
		}
	}

	// Default fallback when schema service is not available or fails
	logger.Log.Warn("No active schema found, using fallback")
	return "unknown-schema"
}

// getApplicationIDFromConsumer extracts application_id from consumer_applications table
func (am *AuditMiddleware) getApplicationIDFromConsumer(r *http.Request) string {
	// Try to get consumer information from JWT token
	consumerAssertion, err := auth.GetConsumerJwtFromToken(am.environment, r)
	if err != nil || consumerAssertion == nil {
		logger.Log.Warn("Failed to get consumer assertion from token", "error", err)
		return "unknown-app"
	}

	// For now, we'll use a simple mapping approach
	// In a real implementation, this would query the consumer_applications table
	// based on the consumer information from the JWT token

	// Map known consumer IDs to application IDs
	consumerToAppMap := map[string]string{
		"test-user":    "app-123",
		"passport-app": "app-123",
		"consumer-123": "app-123",
		// Add more mappings as needed
	}

	// Try to get application ID from consumer ID
	if appID, exists := consumerToAppMap[consumerAssertion.Subscriber]; exists {
		return appID
	}

	// Try to get application ID from application ID in token
	if appID, exists := consumerToAppMap[consumerAssertion.ApplicationId]; exists {
		return appID
	}

	// Default fallback
	logger.Log.Warn("No application mapping found for consumer",
		"subscriber", consumerAssertion.Subscriber,
		"applicationId", consumerAssertion.ApplicationId)
	return "unknown-app"
}
