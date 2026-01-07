package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gov-dx-sandbox/portal-backend/v1/models"
	auditpkg "github.com/gov-dx-sandbox/shared/audit"
)

// Re-export types and functions from shared/audit for convenience
type (
	AuditClient     = auditpkg.AuditClient
	AuditMiddleware = auditpkg.AuditMiddleware
	AuditLogRequest = auditpkg.AuditLogRequest
)

var (
	NewAuditMiddleware         = auditpkg.NewAuditMiddleware
	ResetGlobalAuditMiddleware = auditpkg.ResetGlobalAuditMiddleware
)

// LogAudit logs an audit event for portal-backend operations by extracting request info and creating an audit log
func LogAudit(client AuditClient, r *http.Request, resource string, resourceID *string, status string) {
	// Skip if audit client is not enabled
	if client == nil || !client.IsEnabled() {
		return
	}

	// Only log write operations (POST, PUT, PATCH, DELETE)
	if !isWriteOperation(r.Method) {
		return
	}

	// Extract actor info directly from request
	actorType, actorIDPtr, _ := extractActorInfoFromRequest(r)
	if actorIDPtr == nil {
		// If no actor ID, we can't log the event (required field)
		slog.Warn("Cannot log audit event: no actor ID found")
		return
	}
	actorID := *actorIDPtr

	// Determine event action from HTTP method (CREATE, UPDATE, DELETE)
	eventAction := determineEventType(r.Method)
	if eventAction == "" {
		return
	}

	// Set event type to MANAGEMENT_EVENT for portal operations
	managementEventType := "MANAGEMENT_EVENT"
	eventType := &managementEventType

	// Set target type and ID
	targetType := "RESOURCE"
	var targetID *string
	if resourceID != nil {
		targetID = resourceID
	}

	// Create audit event using local DTO
	timestamp := time.Now().UTC().Format(time.RFC3339)
	additionalMetadata := func() json.RawMessage {
		meta := map[string]interface{}{
			"resource": resource,
		}
		if bytes, err := json.Marshal(meta); err == nil {
			return bytes
		}
		return nil
	}()

	auditRequest := &AuditLogRequest{
		TraceID:            nil, // No trace ID for standalone management events
		Timestamp:          timestamp,
		EventType:          eventType,
		EventAction:        &eventAction,
		Status:             status,
		ActorType:          actorType,
		ActorID:            actorID,
		TargetType:         targetType,
		TargetID:           targetID,
		AdditionalMetadata: additionalMetadata,
	}

	// Log asynchronously (fire-and-forget) using background context
	// If r.Context() is used, it may be cancelled before the audit log is sent
	client.LogEvent(context.Background(), auditRequest)
}

// extractActorInfoFromRequest extracts actor information from the request
func extractActorInfoFromRequest(r *http.Request) (actorType string, actorID *string, actorRole *string) {
	// Try to get authenticated user first
	user, err := GetUserFromRequest(r)
	if err == nil && user != nil {
		actorType = "USER"
		userID := user.IdpUserID
		actorID = &userID

		// Get user role (simplified)
		role := "MEMBER" // Default role
		if user.HasPermission(models.PermissionCreateMember) {
			role = "ADMIN"
		}
		actorRole = &role
		return
	}

	// Fallback to headers
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = r.Header.Get("X-Auth-User-ID")
	}

	role := r.Header.Get("X-User-Role")
	if role == "" {
		role = r.Header.Get("X-Auth-Role")
	}

	if userID != "" {
		actorType = "USER"
		actorID = &userID
		if role != "" && (role == "MEMBER" || role == "ADMIN") {
			actorRole = &role
		} else {
			defaultRole := "MEMBER"
			actorRole = &defaultRole
		}
	} else {
		actorType = "SERVICE"
		actorID = nil
		actorRole = nil
	}

	return
}

// LogAuditEvent - global function for easy access from handlers
func LogAuditEvent(r *http.Request, resource string, resourceID *string, status string) {
	globalMiddleware := auditpkg.GetGlobalAuditMiddleware()
	if globalMiddleware != nil {
		LogAudit(globalMiddleware.Client(), r, resource, resourceID, status)
	} else {
		slog.Warn("Audit logging skipped: globalAuditMiddleware is not initialized")
	}
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
