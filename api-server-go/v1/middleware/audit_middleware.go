package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	auditclient "github.com/gov-dx-sandbox/audit-service/client"
)

// AuditMiddleware handles audit logging for requests
// It reads information from request context (set by RequestContextMiddleware)
// instead of doing brute force extraction
type AuditMiddleware struct {
	auditClient auditclient.AuditClient
}

// NewAuditMiddleware creates a new audit middleware
func NewAuditMiddleware(auditServiceURL string) *AuditMiddleware {
	auditClient := auditclient.NewAuditClient(auditServiceURL)
	if auditClient != nil {
		slog.Info("Audit middleware initialized", "auditServiceURL", auditServiceURL)
	} else {
		slog.Warn("Audit middleware disabled (audit service URL not configured)")
	}

	return &AuditMiddleware{
		auditClient: auditClient,
	}
}

// AuditLoggingMiddleware wraps an http.Handler with audit logging
// It reads from context instead of extracting from path/body
func (m *AuditMiddleware) AuditLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Track start time
		startTime := time.Now()

		// Create a response writer wrapper to capture status code
		responseWrapper := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call the next handler
		next.ServeHTTP(responseWrapper, r)

		// Only log write operations (POST, PUT, PATCH, DELETE)
		if !isWriteOperation(r.Method) {
			return
		}

		// Determine success/failure based on status code
		status := "SUCCESS"
		if responseWrapper.statusCode >= 400 {
			status = "FAILURE"
		}

		// Read all information from context (set by RequestContextMiddleware)
		ctx := r.Context()
		actorType := GetActorType(ctx)
		actorID := GetActorID(ctx)
		actorRole := GetActorRole(ctx)
		targetResource := GetTargetResource(ctx)

		// Get resource ID from context based on resource type
		var targetResourceID string
		switch targetResource {
		case "SCHEMAS":
			targetResourceID = GetSchemaID(ctx)
		case "APPLICATIONS":
			targetResourceID = GetApplicationID(ctx)
		case "MEMBERS":
			targetResourceID = GetMemberID(ctx)
		case "SCHEMA-SUBMISSIONS":
			targetResourceID = GetSchemaSubmissionID(ctx)
		case "APPLICATION-SUBMISSIONS":
			targetResourceID = GetApplicationSubmissionID(ctx)
		}

		// Only log if we have the required information
		if targetResource == "" || targetResourceID == "" {
			slog.Debug("Skipping audit log - missing resource information",
				"targetResource", targetResource,
				"targetResourceID", targetResourceID,
				"path", r.URL.Path)
			return
		}

		// Determine event type from HTTP method
		eventType := determineEventType(r.Method)

		// Log management event asynchronously (fire-and-forget)
		m.logManagementEvent(ctx, eventType, actorType, actorID, actorRole, targetResource, targetResourceID, status, time.Since(startTime))
	})
}

// logManagementEvent logs a management event using the audit client
func (m *AuditMiddleware) logManagementEvent(
	ctx context.Context,
	eventType, actorType string,
	actorID, actorRole *string,
	targetResource, targetResourceID, status string,
	duration time.Duration,
) {
	if m.auditClient == nil {
		return // Audit client disabled
	}

	// Log asynchronously (fire-and-forget)
	_ = m.auditClient.LogManagementEvent(ctx, auditclient.ManagementEventRequest{
		EventType: eventType,
		Actor: auditclient.Actor{
			Type: actorType,
			ID:   actorID,
			Role: actorRole,
		},
		Target: auditclient.Target{
			Resource:   targetResource,
			ResourceID: targetResourceID,
		},
	})

	slog.Debug("Audit log sent",
		"eventType", eventType,
		"targetResource", targetResource,
		"targetResourceID", targetResourceID,
		"status", status,
		"duration", duration)
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Helper functions

func isWriteOperation(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete
}

func determineEventType(method string) string {
	switch method {
	case http.MethodPost:
		return "CREATE"
	case http.MethodPut, http.MethodPatch:
		return "UPDATE"
	case http.MethodDelete:
		return "DELETE"
	default:
		return ""
	}
}

// GetSchemaSubmissionID retrieves schema submission ID from context
func GetSchemaSubmissionID(ctx context.Context) string {
	if val := ctx.Value(contextKeySchemaSubmissionID); val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// GetApplicationSubmissionID retrieves application submission ID from context
func GetApplicationSubmissionID(ctx context.Context) string {
	if val := ctx.Value(contextKeyApplicationSubmissionID); val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}
