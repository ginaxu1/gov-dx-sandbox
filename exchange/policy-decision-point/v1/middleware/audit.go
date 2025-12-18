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
	"time"
)

// AuditMiddleware handles audit logging for PDP operations
type AuditMiddleware struct {
	auditServiceURL string
	httpClient      *http.Client
}

// Global audit middleware instance for easy access from handlers
var (
	globalAuditMiddleware *AuditMiddleware
	globalAuditOnce       sync.Once
)

// CreateAuditLogRequest represents the request payload for creating a generalized audit log
// This matches the audit-service v1 API structure
type CreateAuditLogRequest struct {
	// Trace & Correlation
	TraceID *string `json:"traceId,omitempty"` // UUID string, nullable for standalone events

	// Temporal
	Timestamp *string `json:"timestamp,omitempty"` // ISO 8601 format, optional (defaults to now)

	// Event Classification
	EventName string  `json:"eventName" validate:"required"` // POLICY_CHECK, CONSENT_CHECK, DATA_FETCH, MANAGEMENT_EVENT
	EventType *string `json:"eventType,omitempty"`           // CREATE, READ, UPDATE, DELETE (nullable for non-CRUD)
	Status    string  `json:"status" validate:"required"`    // SUCCESS or FAILURE

	// Actor (Flattened from ActorMetadata)
	ActorType        string          `json:"actorType" validate:"required"` // USER or SERVICE
	ActorServiceName *string         `json:"actorServiceName,omitempty"`    // Required for SERVICE, NULL for USER
	ActorUserID      *string         `json:"actorUserId,omitempty"`         // Required for USER, NULL for SERVICE (UUID string)
	ActorUserType    *string         `json:"actorUserType,omitempty"`       // ADMIN or MEMBER (for USER)
	ActorMetadata    json.RawMessage `json:"actorMetadata,omitempty"`       // Additional actor context

	// Target (Flattened from TargetMetadata)
	TargetType        string          `json:"targetType" validate:"required"` // RESOURCE or SERVICE
	TargetServiceName *string         `json:"targetServiceName,omitempty"`    // Required for SERVICE, NULL for RESOURCE
	TargetResource    *string         `json:"targetResource,omitempty"`       // Required for RESOURCE, NULL for SERVICE
	TargetResourceID  *string         `json:"targetResourceId,omitempty"`     // Optional UUID string
	TargetMetadata    json.RawMessage `json:"targetMetadata,omitempty"`       // Additional target context

	// Request/Response (PIA-free)
	RequestedData    json.RawMessage `json:"requestedData,omitempty"`    // Request payload
	ResponseMetadata json.RawMessage `json:"responseMetadata,omitempty"` // Response or error

	// Additional Context
	EventMetadata json.RawMessage `json:"eventMetadata,omitempty"` // Additional event-specific metadata
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
			httpClient:      &http.Client{Timeout: 5 * time.Second},
		}
	}

	globalAuditOnce.Do(func() {
		globalAuditMiddleware = middleware
	})

	return middleware
}

// LogGeneralizedAudit logs a generalized audit event
func (m *AuditMiddleware) LogGeneralizedAudit(ctx context.Context, auditRequest *CreateAuditLogRequest) {
	// Skip if audit service is not configured
	if m.auditServiceURL == "" {
		return
	}

	// If TraceID is missing in request but present in context, use it
	if auditRequest.TraceID == nil || *auditRequest.TraceID == "" {
		if val := ctx.Value(TraceIDKey{}); val != nil {
			if traceID, ok := val.(string); ok && traceID != "" {
				auditRequest.TraceID = &traceID
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

	traceIDStr := ""
	if event.TraceID != nil {
		traceIDStr = *event.TraceID
	}
	slog.Debug("Generalized audit event logged successfully",
		"traceId", traceIDStr,
		"eventName", event.EventName)
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
