package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/v1/models"
)

// AuditMiddleware handles audit logging for CUD operations
type AuditMiddleware struct {
	auditServiceURL string
	httpClient      *http.Client
}

// Global audit middleware instance for easy access from handlers
var globalAuditMiddleware *AuditMiddleware

// ManagementEventRequest represents the audit service API structure
type ManagementEventRequest struct {
	EventType string `json:"eventType"`
	Actor     Actor  `json:"actor"`
	Target    Target `json:"target"`
}

// Actor represents who performed the action
type Actor struct {
	Type string  `json:"type"`           // USER or SERVICE
	ID   *string `json:"id,omitempty"`   // User ID (required for USER type)
	Role *string `json:"role,omitempty"` // MEMBER or ADMIN (required for USER type)
}

// Target represents what resource was acted upon
type Target struct {
	Resource   string `json:"resource"`   // MEMBERS, SCHEMAS, etc.
	ResourceID string `json:"resourceId"` // Actual resource ID
}

// NewAuditMiddleware creates a new audit middleware
func NewAuditMiddleware(auditServiceURL string) *AuditMiddleware {
	if auditServiceURL == "" {
		slog.Warn("Audit middleware disabled (audit service URL not configured)")
		return &AuditMiddleware{auditServiceURL: "", httpClient: nil}
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	slog.Info("Audit middleware initialized", "auditServiceURL", auditServiceURL)
	middleware := &AuditMiddleware{
		auditServiceURL: auditServiceURL,
		httpClient:      httpClient,
	}

	// Set global instance for easy access from handlers
	globalAuditMiddleware = middleware
	return middleware
}

// AuditLoggingMiddleware wraps an http.Handler with audit logging (simplified)
func (m *AuditMiddleware) AuditLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Just pass through - actual audit logging is now done directly by handlers
		next.ServeHTTP(w, r)
	})
}

// LogAudit - simplified function to log audit events directly from handlers
func (m *AuditMiddleware) LogAudit(r *http.Request, resource, resourceID string) {
	// Skip if audit service is not configured
	if m.auditServiceURL == "" {
		return
	}

	// Only log write operations (POST, PUT, PATCH, DELETE)
	if !isWriteOperation(r.Method) {
		return
	}

	// Extract actor info directly from request
	actorType, actorID, actorRole := extractActorInfoFromRequest(r)

	// Determine event type from HTTP method
	eventType := determineEventType(r.Method)
	if eventType == "" {
		return
	}

	// Create audit event
	auditEvent := ManagementEventRequest{
		EventType: eventType,
		Actor: Actor{
			Type: actorType,
			ID:   actorID,
			Role: actorRole,
		},
		Target: Target{
			Resource:   resource,
			ResourceID: resourceID,
		},
	}

	// Log asynchronously (fire-and-forget) using background context
	go m.logManagementEvent(context.Background(), auditEvent)
}

// logManagementEvent sends the audit event to the audit service
func (m *AuditMiddleware) logManagementEvent(ctx context.Context, event ManagementEventRequest) {
	if m.httpClient == nil {
		return
	}

	// Marshal the event
	jsonData, err := json.Marshal(event)
	if err != nil {
		slog.Error("Failed to marshal audit event", "error", err)
		return
	}

	// Create request to audit service
	url := fmt.Sprintf("%s/api/events", m.auditServiceURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		slog.Error("Failed to create audit request", "error", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := m.httpClient.Do(req)
	if err != nil {
		slog.Error("Failed to send audit event", "error", err, "url", url)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		slog.Error("Audit service returned error", "status", resp.StatusCode, "url", url)
		return
	}

	slog.Debug("Audit event logged successfully",
		"eventType", event.EventType,
		"resource", event.Target.Resource,
		"resourceId", event.Target.ResourceID)
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
func LogAuditEvent(r *http.Request, resource, resourceID string) {
	if globalAuditMiddleware != nil {
		globalAuditMiddleware.LogAudit(r, resource, resourceID)
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
