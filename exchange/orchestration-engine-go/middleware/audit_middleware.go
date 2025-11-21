package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"sync"
)

// AuditMiddleware handles audit logging for CUD operations
type AuditMiddleware struct {
	auditServiceURL string
	httpClient      *http.Client
}

// Global audit middleware instance for easy access from handlers
var (
	globalAuditMiddleware *AuditMiddleware
	globalAuditOnce       sync.Once
)

// DataExchangeEventAuditRequest represents the audit service API structure for data exchange events
type DataExchangeEventAuditRequest struct {
	Timestamp         string          `json:"timestamp" validate:"required"`
	Status            string          `json:"status" validate:"required"`
	ApplicationID     string          `json:"applicationId" validate:"required"`
	SchemaID          string          `json:"schemaId" validate:"required"`
	RequestedData     json.RawMessage `json:"requestedData" validate:"required"`
	OnBehalfOfOwnerID *string         `json:"onBehalfOfOwnerId,omitempty"`
	ConsumerID        *string         `json:"consumerId,omitempty"`
	ProviderID        *string         `json:"providerId,omitempty"`
	AdditionalInfo    json.RawMessage `json:"additionalInfo,omitempty"`
}

// NewAuditMiddleware creates a new audit middleware with thread-safe global initialization
// This function should typically only be called once during application startup.
// Subsequent calls will return a new instance but won't update the global instance.
func NewAuditMiddleware(auditServiceURL string) *AuditMiddleware {
	var middleware *AuditMiddleware

	if auditServiceURL == "" {
		middleware = &AuditMiddleware{auditServiceURL: "", httpClient: nil}
	} else {
		middleware = &AuditMiddleware{
			auditServiceURL: auditServiceURL,
			httpClient:      &http.Client{},
		}
	}

	globalAuditOnce.Do(func() {
		globalAuditMiddleware = middleware
	})

	return middleware
}

// LogAudit - function to log audit events directly from federator
func (m *AuditMiddleware) LogAudit(auditRequest *DataExchangeEventAuditRequest) {
	// Skip if audit service is not configured
	if m.auditServiceURL == "" {
		return
	}

	// Log asynchronously (fire-and-forget) using background context
	go m.logDataExchangeEvent(context.Background(), *auditRequest)
}

// logDataExchangeEvent sends the audit event to the audit service
func (m *AuditMiddleware) logDataExchangeEvent(ctx context.Context, event DataExchangeEventAuditRequest) {
	if m.httpClient == nil {
		return
	}

	payloadBytes, err := json.Marshal(event)
	if err != nil {
		slog.Error("Failed to marshal audit request", "error", err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, "POST", m.auditServiceURL+"/data-exchange-events", bytes.NewReader(payloadBytes))
	if err != nil {
		slog.Error("Failed to create audit request", "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		slog.Error("Failed to send audit request", "error", err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			slog.Error("Failed to close audit response body", "error", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		slog.Error("Audit service returned non-201 status", "status", resp.StatusCode, "body", string(bodyBytes))
		return
	}

	slog.Debug("Data exchange audit event logged successfully",
		"applicationId", event.ApplicationID,
		"schemaId", event.SchemaID,
		"status", event.Status)
}

// LogAuditEvent logs a data exchange audit event using global audit middleware instance
func LogAuditEvent(auditRequest *DataExchangeEventAuditRequest) {
	if globalAuditMiddleware != nil {
		globalAuditMiddleware.LogAudit(auditRequest)
	} else {
		slog.Warn("Global AuditMiddleware is not initialized; audit event not logged")
	}
}

// ResetGlobalAuditMiddleware is a helper function for tests to reset the global audit middleware instance
func ResetGlobalAuditMiddleware() {
	globalAuditOnce = sync.Once{}
	globalAuditMiddleware = nil
}
