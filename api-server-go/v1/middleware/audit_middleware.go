package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/v1/models"
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

// ManagementEventRequest represents the audit service API structure
type ManagementEventRequest struct {
	Timestamp *string `json:"timestamp,omitempty"` // ISO 8601 timestamp
	EventType string  `json:"eventType"`           // "CREATE", "UPDATE", "DELETE"
	Status    string  `json:"status"`              // "SUCCESS", "FAILURE"
	Actor     Actor   `json:"actor"`
	Target    Target  `json:"target"`
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

// AuditInfo holds information to be logged by the audit middleware
type AuditInfo struct {
	ResourceID string                 // The ID of the resource being acted upon
	Metadata   map[string]interface{} // Additional metadata for the audit log
}

// SetResourceID sets the resource ID in the AuditInfo
func (a *AuditInfo) SetResourceID(resourceID string) {
	a.ResourceID = resourceID
}

// auditInfoKey is a custom type for context keys to avoid collisions
type auditInfoKeyType string

const auditInfoKey auditInfoKeyType = "auditInfo"

// NewAuditMiddleware creates a new audit middleware with thread-safe global initialization
// This function should typically only be called once during application startup.
// Subsequent calls will return a new instance but won't update the global instance.
func NewAuditMiddleware(auditServiceURL string) *AuditMiddleware {
	var middleware *AuditMiddleware

	if auditServiceURL == "" {
		slog.Warn("Audit middleware disabled (audit service URL not configured)")
		middleware = &AuditMiddleware{auditServiceURL: "", httpClient: nil}
	} else {
		httpClient := &http.Client{
			Timeout: 10 * time.Second,
		}

		slog.Info("Audit middleware initialized", "auditServiceURL", auditServiceURL)
		middleware = &AuditMiddleware{
			auditServiceURL: auditServiceURL,
			httpClient:      httpClient,
		}
	}

	// Set global instance for easy access from handlers (thread-safe, only once)
	globalAuditOnce.Do(func() {
		globalAuditMiddleware = middleware
		slog.Debug("Global audit middleware instance set")
	})

	return middleware
}

// statusRecorder is a wrapper around http.ResponseWriter to capture the status code
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// WithAudit is a middleware that enables audit logging for a handler
func (m *AuditMiddleware) WithAudit(resource string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip if audit service is not configured
			if m.auditServiceURL == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Only log write operations (POST, PUT, PATCH, DELETE)
			if !isWriteOperation(r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			// Create AuditInfo and inject into context
			auditInfo := &AuditInfo{
				Metadata: make(map[string]interface{}),
			}
			ctx := context.WithValue(r.Context(), auditInfoKey, auditInfo)
			r = r.WithContext(ctx)

			// Wrap ResponseWriter to capture status code
			recorder := &statusRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // Default to 200 if WriteHeader is not called
			}

			// Call the next handler
			next.ServeHTTP(recorder, r)

			// After handler returns, log the audit event
			// IMPORTANT: We always log events, even for failures where ResourceID may be empty.
			// This ensures we capture all failure scenarios, including CREATE failures that occur
			// before a resource ID is generated (e.g., validation errors, authorization failures).
			// The audit service accepts empty ResourceID for CREATE failures (eventType=CREATE, status=FAILURE).

			// Extract actor info directly from request
			actorType, actorID, actorRole := extractActorInfoFromRequest(r)

			// Determine event type from HTTP method
			eventType := determineEventType(r.Method)
			if eventType == "" {
				return
			}

			// Determine status based on status code
			// Any status code >= 400 indicates a failure
			status := "SUCCESS"
			if recorder.statusCode >= 400 {
				status = "FAILURE"
			}

			// Get current timestamp in ISO 8601 format
			now := time.Now().Format(time.RFC3339)

			// Create audit event
			// Note: ResourceID may be empty for CREATE failures (when handler fails before creating resource)
			// This is acceptable and will be validated by the audit service
			auditEvent := ManagementEventRequest{
				Timestamp: &now,
				EventType: eventType,
				Status:    status,
				Actor: Actor{
					Type: actorType,
					ID:   actorID,
					Role: actorRole,
				},
				Target: Target{
					Resource:   resource,
					ResourceID: auditInfo.ResourceID, // May be empty for CREATE failures
				},
			}

			// Log asynchronously (fire-and-forget) using background context
			// This always executes, regardless of whether ResourceID is set
			go m.logManagementEvent(context.Background(), auditEvent)
		})
	}
}

// GetAuditInfoFromContext retrieves the AuditInfo from the request context
func GetAuditInfoFromContext(ctx context.Context) (*AuditInfo, error) {
	auditInfo, ok := ctx.Value(auditInfoKey).(*AuditInfo)
	if !ok {
		return nil, fmt.Errorf("audit info not found in context")
	}
	return auditInfo, nil
}

// SetResourceID sets the resource ID in the AuditInfo from the request context
func SetResourceID(r *http.Request, resourceID string) error {
	auditInfo, err := GetAuditInfoFromContext(r.Context())
	if err != nil {
		return err
	}
	auditInfo.ResourceID = resourceID
	return nil
}

// LogAuditEvent is a helper function for backward compatibility
// It sets the ResourceID in the AuditInfo context, which will be logged by the middleware
func LogAuditEvent(r *http.Request, resource string, resourceID string) {
	// Try to set resource ID in context (if audit middleware is active)
	if err := SetResourceID(r, resourceID); err != nil {
		// If audit middleware is not active, just log a debug message
		slog.Debug("Audit middleware not active, skipping audit log", "resource", resource, "resourceId", resourceID)
		return
	}
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

// ResetGlobalAuditMiddleware resets the global audit middleware instance for testing purposes
// This should only be used in tests to reset state between test cases
func ResetGlobalAuditMiddleware() {
	globalAuditMiddleware = nil
	globalAuditOnce = sync.Once{}
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
