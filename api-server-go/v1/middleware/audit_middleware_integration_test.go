package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	auditclient "github.com/gov-dx-sandbox/audit-service/client"
)

// mockAuditClient is a mock implementation of AuditClient for testing
type mockAuditClient struct {
	loggedEvents []auditclient.ManagementEventRequest
}

func (m *mockAuditClient) LogDataExchange(ctx context.Context, event auditclient.DataExchangeEvent) error {
	return nil
}

func (m *mockAuditClient) LogManagementEvent(ctx context.Context, event auditclient.ManagementEventRequest) error {
	m.loggedEvents = append(m.loggedEvents, event)
	return nil
}

func TestAuditMiddleware_Integration(t *testing.T) {
	// Create a mock audit client
	mockClient := &mockAuditClient{
		loggedEvents: make([]auditclient.ManagementEventRequest, 0),
	}

	// Create audit middleware with mock client
	auditMiddleware := &AuditMiddleware{
		auditClient: mockClient,
	}

	tests := []struct {
		name               string
		method             string
		path               string
		contextSetup       func(context.Context) context.Context
		expectedEvent      string
		expectedResource   string
		expectedResourceID string
	}{
		{
			name:   "POST /api/v1/members - CREATE event",
			method: http.MethodPost,
			path:   "/api/v1/members",
			contextSetup: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, contextKeyActorType, "USER")
				ctx = context.WithValue(ctx, contextKeyActorID, "user-123")
				ctx = context.WithValue(ctx, contextKeyActorRole, "ADMIN")
				ctx = context.WithValue(ctx, contextKeyTargetResource, "MEMBERS")
				ctx = context.WithValue(ctx, contextKeyMemberID, "member-456")
				return ctx
			},
			expectedEvent:      "CREATE",
			expectedResource:   "MEMBERS",
			expectedResourceID: "member-456",
		},
		{
			name:   "PUT /api/v1/schemas/{id} - UPDATE event",
			method: http.MethodPut,
			path:   "/api/v1/schemas/schema-789",
			contextSetup: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, contextKeyActorType, "USER")
				ctx = context.WithValue(ctx, contextKeyActorID, "user-123")
				ctx = context.WithValue(ctx, contextKeyActorRole, "MEMBER")
				ctx = context.WithValue(ctx, contextKeyTargetResource, "SCHEMAS")
				ctx = context.WithValue(ctx, contextKeySchemaID, "schema-789")
				return ctx
			},
			expectedEvent:      "UPDATE",
			expectedResource:   "SCHEMAS",
			expectedResourceID: "schema-789",
		},
		{
			name:   "DELETE /api/v1/applications/{id} - DELETE event",
			method: http.MethodDelete,
			path:   "/api/v1/applications/app-123",
			contextSetup: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, contextKeyActorType, "USER")
				ctx = context.WithValue(ctx, contextKeyActorID, "user-123")
				ctx = context.WithValue(ctx, contextKeyActorRole, "ADMIN")
				ctx = context.WithValue(ctx, contextKeyTargetResource, "APPLICATIONS")
				ctx = context.WithValue(ctx, contextKeyApplicationID, "app-123")
				return ctx
			},
			expectedEvent:      "DELETE",
			expectedResource:   "APPLICATIONS",
			expectedResourceID: "app-123",
		},
		{
			name:   "GET /api/v1/members - No audit log (read operation)",
			method: http.MethodGet,
			path:   "/api/v1/members",
			contextSetup: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, contextKeyTargetResource, "MEMBERS")
				ctx = context.WithValue(ctx, contextKeyMemberID, "member-456")
				return ctx
			},
			expectedEvent:      "",
			expectedResource:   "",
			expectedResourceID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock client
			mockClient.loggedEvents = make([]auditclient.ManagementEventRequest, 0)

			// Create a simple test handler
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			})

			// Create request
			req := httptest.NewRequest(tt.method, tt.path, nil)
			// Setup context BEFORE middleware runs (as RequestContextMiddleware would)
			ctx := tt.contextSetup(req.Context())
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			// Apply audit middleware
			middlewareHandler := auditMiddleware.AuditLoggingMiddleware(testHandler)
			middlewareHandler.ServeHTTP(w, req)

			// Verify results
			if tt.expectedEvent == "" {
				// Read operations should not log
				if len(mockClient.loggedEvents) != 0 {
					t.Errorf("Expected no audit events for read operation, got %d", len(mockClient.loggedEvents))
				}
			} else {
				// Write operations should log
				if len(mockClient.loggedEvents) != 1 {
					t.Errorf("Expected 1 audit event, got %d", len(mockClient.loggedEvents))
					return
				}

				event := mockClient.loggedEvents[0]
				if event.EventType != tt.expectedEvent {
					t.Errorf("Expected event type %s, got %s", tt.expectedEvent, event.EventType)
				}
				if event.Target.Resource != tt.expectedResource {
					t.Errorf("Expected resource %s, got %s", tt.expectedResource, event.Target.Resource)
				}
				if event.Target.ResourceID != tt.expectedResourceID {
					t.Errorf("Expected resource ID %s, got %s", tt.expectedResourceID, event.Target.ResourceID)
				}
			}
		})
	}
}

func TestAuditMiddleware_NilClient(t *testing.T) {
	// Create audit middleware with nil client (audit service not configured)
	auditMiddleware := &AuditMiddleware{
		auditClient: nil,
	}

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), contextKeyTargetResource, "MEMBERS")
		ctx = context.WithValue(ctx, contextKeyMemberID, "member-456")
		r = r.WithContext(ctx)
		w.WriteHeader(http.StatusOK)
	})

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/members", nil)
	w := httptest.NewRecorder()

	// Apply audit middleware - should not panic
	middlewareHandler := auditMiddleware.AuditLoggingMiddleware(testHandler)
	middlewareHandler.ServeHTTP(w, req)

	// Should complete without error
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestAuditMiddleware_MissingResourceInfo(t *testing.T) {
	// Create a mock audit client
	mockClient := &mockAuditClient{
		loggedEvents: make([]auditclient.ManagementEventRequest, 0),
	}

	// Create audit middleware with mock client
	auditMiddleware := &AuditMiddleware{
		auditClient: mockClient,
	}

	// Create a test handler without resource info in context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/members", nil)
	w := httptest.NewRecorder()

	// Apply audit middleware
	middlewareHandler := auditMiddleware.AuditLoggingMiddleware(testHandler)
	middlewareHandler.ServeHTTP(w, req)

	// Should not log anything because resource info is missing
	if len(mockClient.loggedEvents) != 0 {
		t.Errorf("Expected no audit events when resource info is missing, got %d", len(mockClient.loggedEvents))
	}
}
