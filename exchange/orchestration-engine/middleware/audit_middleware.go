package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sync"

	"github.com/google/uuid"
)

// TraceIDKey is the context key for Trace ID
type TraceIDKey struct{}

const (
	// TraceIDHeader is the HTTP header name for trace ID
	TraceIDHeader = "X-Trace-ID"
)

// GetTraceIDFromContext retrieves the trace ID from the context
func GetTraceIDFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey{}).(string); ok {
		return traceID
	}
	return ""
}

// TraceIDMiddleware extracts or generates a trace ID and adds it to the request context
func TraceIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract trace ID from header or generate new one
		traceID := r.Header.Get(TraceIDHeader)
		if traceID == "" {
			// Generate new trace ID if not provided
			traceID = generateTraceID()
		}

		// Add trace ID to context
		ctx := context.WithValue(r.Context(), TraceIDKey{}, traceID)

		// Add trace ID to response header for client visibility
		w.Header().Set(TraceIDHeader, traceID)

		// Continue with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// generateTraceID generates a UUID trace ID
func generateTraceID() string {
	return uuid.New().String()
}

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

	auditURL, err := url.JoinPath(m.auditServiceURL, "data-exchange-events")
	if err != nil {
		slog.Error("Failed to construct audit URL", "error", err)
		return
	}
	req, err := http.NewRequestWithContext(ctx, "POST", auditURL, bytes.NewReader(payloadBytes))
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

// CreateAuditLogRequest represents the request payload for creating a generalized audit log
type CreateAuditLogRequest struct {
	TraceID       string          `json:"traceId" validate:"required,uuid"`
	Timestamp     string          `json:"timestamp"` // Optional, defaults to now
	SourceService string          `json:"sourceService" validate:"required"`
	TargetService string          `json:"targetService,omitempty"`
	EventType     string          `json:"eventType" validate:"required"`
	Status        string          `json:"status" validate:"required"`
	ActorID       *string         `json:"actorId,omitempty"`
	Resources     json.RawMessage `json:"resources,omitempty"`
	Metadata      json.RawMessage `json:"metadata,omitempty"`
}

// LogGeneralizedAudit logs a generalized audit event
func (m *AuditMiddleware) LogGeneralizedAudit(ctx context.Context, auditRequest *CreateAuditLogRequest) {
	// Skip if audit service is not configured
	if m.auditServiceURL == "" {
		return
	}

	// If TraceID is missing in request but present in context, use it
	if auditRequest.TraceID == "" {
		if val := ctx.Value(TraceIDKey{}); val != nil {
			if traceID, ok := val.(string); ok {
				auditRequest.TraceID = traceID
			}
		}
	}

	// Log asynchronously (fire-and-forget) using background context
	go m.logGeneralizedAuditEvent(context.Background(), *auditRequest)
}

// logGeneralizedAuditEvent sends the audit log to the audit service
func (m *AuditMiddleware) logGeneralizedAuditEvent(ctx context.Context, event CreateAuditLogRequest) {
	if m.httpClient == nil {
		return
	}

	payloadBytes, err := json.Marshal(event)
	if err != nil {
		slog.Error("Failed to marshal audit request", "error", err)
		return
	}

	auditURL, err := url.JoinPath(m.auditServiceURL, "api", "audit-logs")
	if err != nil {
		slog.Error("Failed to construct audit URL", "error", err)
		return
	}
	req, err := http.NewRequestWithContext(ctx, "POST", auditURL, bytes.NewReader(payloadBytes))
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

	slog.Debug("Generalized audit event logged successfully",
		"traceId", event.TraceID,
		"eventType", event.EventType)
}

// LogGeneralizedAuditEvent helper for global access
func LogGeneralizedAuditEvent(ctx context.Context, auditRequest *CreateAuditLogRequest) {
	if globalAuditMiddleware != nil {
		globalAuditMiddleware.LogGeneralizedAudit(ctx, auditRequest)
	} else {
		slog.Warn("Global AuditMiddleware is not initialized; audit event not logged")
	}
}

// ResetGlobalAuditMiddleware is a helper function for tests to reset the global audit middleware instance
func ResetGlobalAuditMiddleware() {
	globalAuditOnce = sync.Once{}
	globalAuditMiddleware = nil
}
