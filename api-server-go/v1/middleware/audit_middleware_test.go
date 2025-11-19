package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestAuditMiddleware_Initialization(t *testing.T) {
	// Reset global state for this test
	ResetGlobalAuditMiddleware()

	// Test with audit enabled
	auditMiddleware := NewAuditMiddleware("http://localhost:8080")
	if auditMiddleware.auditServiceURL == "" {
		t.Error("Expected audit middleware to have service URL when URL is provided")
	}
	if auditMiddleware.httpClient == nil {
		t.Error("Expected audit middleware to have HTTP client when URL is provided")
	}

	// Test with audit disabled (create new instance, but global should already be set)
	auditMiddleware2 := NewAuditMiddleware("")
	if auditMiddleware2.auditServiceURL != "" {
		t.Error("Expected audit middleware to have empty service URL when URL is empty")
	}
	if auditMiddleware2.httpClient != nil {
		t.Error("Expected audit middleware to have nil HTTP client when URL is empty")
	}

	// Global should still be the first instance (due to sync.Once)
	if globalAuditMiddleware != auditMiddleware {
		t.Error("Expected global instance to be the first initialized middleware")
	}
}

func TestLogAuditEvent_GlobalFunction(t *testing.T) {
	// Reset global state for this test
	ResetGlobalAuditMiddleware()

	// Initialize global audit middleware
	_ = NewAuditMiddleware("http://localhost:3001")

	// Test that LogAuditEvent doesn't panic when called
	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	req.Header.Set("X-User-ID", "test-user")
	req.Header.Set("X-User-Role", "ADMIN")

	// This should not panic even if the audit service is not available
	LogAuditEvent(req, "TEST_RESOURCE", "test-id-123")
}

func TestLogAudit_SkipsReadOperations(t *testing.T) {
	_ = NewAuditMiddleware("http://localhost:3001")

	// GET request should be skipped (LogAuditEvent will just set ResourceID if middleware is active)
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	LogAuditEvent(req, "TEST_RESOURCE", "test-id")

	// This test passes if no panic occurs - we can't easily test HTTP calls without a mock server
}

func TestLogAudit_ProcessesWriteOperations(t *testing.T) {
	auditMiddleware := NewAuditMiddleware("http://localhost:3001")
	_ = auditMiddleware // Use the middleware to initialize it

	// POST request should be processed (though it may fail to send)
	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	req.Header.Set("X-User-ID", "test-user")
	req.Header.Set("X-User-Role", "MEMBER")

	LogAuditEvent(req, "TEST_RESOURCE", "test-id")

	// This test passes if no panic occurs - we can't easily test HTTP calls without a mock server
}

// TestCreateRequest_AuditLogging verifies that a CREATE request triggers the audit logic.
func TestCreateRequest_AuditLogging(t *testing.T) {
	// Reset global state
	ResetGlobalAuditMiddleware()

	// Setup a mock audit service server to verify the request is sent
	auditServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request to audit service, got %s", r.Method)
		}
		if r.URL.Path != "/api/events" {
			t.Errorf("Expected path /api/events, got %s", r.URL.Path)
		}

		// Decode body
		var event ManagementEventRequest
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			t.Errorf("Failed to decode audit event: %v", err)
			return
		}

		// Verify event details
		if event.EventType != "CREATE" {
			t.Errorf("Expected EventType CREATE, got %s", event.EventType)
		}
		if event.Status != "SUCCESS" {
			t.Errorf("Expected Status SUCCESS, got %s", event.Status)
		}
		if event.Target.Resource != "TEST_RESOURCE" {
			t.Errorf("Expected Resource TEST_RESOURCE, got %s", event.Target.Resource)
		}
		if event.Target.ResourceID != "created-id-123" {
			t.Errorf("Expected ResourceID created-id-123, got %s", event.Target.ResourceID)
		}
		if event.Actor.ID == nil || *event.Actor.ID != "test-user" {
			t.Errorf("Expected Actor ID test-user, got %v", event.Actor.ID)
		}

		w.WriteHeader(http.StatusCreated)
	}))
	defer auditServer.Close()

	// Initialize middleware with mock server URL
	mw := NewAuditMiddleware(auditServer.URL)

	// Create a handler that simulates a CREATE operation
	createHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Retrieve AuditInfo
		auditInfo, err := GetAuditInfoFromContext(r.Context())
		if err != nil {
			t.Errorf("Failed to get audit info: %v", err)
			return
		}
		if auditInfo == nil {
			t.Error("AuditInfo not found in context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// 2. Simulate creating a resource
		newID := "created-id-123"

		// 3. Set the resource ID in AuditInfo
		auditInfo.SetResourceID(newID)

		w.WriteHeader(http.StatusCreated)
	})

	// Wrap with middleware
	wrappedHandler := mw.WithAudit("TEST_RESOURCE")(createHandler)

	// Create POST request
	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	req.Header.Set("X-User-ID", "test-user")
	req.Header.Set("X-User-Role", "ADMIN")
	w := httptest.NewRecorder()

	// Execute
	wrappedHandler.ServeHTTP(w, req)

	// Wait a bit for the goroutine to finish (since logManagementEvent is async)
	time.Sleep(100 * time.Millisecond)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
}

func TestFailureRequest_AuditLogging(t *testing.T) {
	// Reset global state
	ResetGlobalAuditMiddleware()

	// Setup a mock audit service server
	auditServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var event ManagementEventRequest
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			t.Errorf("Failed to decode audit event: %v", err)
			return
		}

		// Verify status is FAILURE
		if event.Status != "FAILURE" {
			t.Errorf("Expected Status FAILURE, got %s", event.Status)
		}
		// Resource ID might be empty or partial depending on when failure occurred,
		// but here we expect it to be empty as handler failed before setting it
		if event.Target.ResourceID != "" {
			t.Errorf("Expected empty ResourceID for failure, got %s", event.Target.ResourceID)
		}

		w.WriteHeader(http.StatusCreated)
	}))
	defer auditServer.Close()

	mw := NewAuditMiddleware(auditServer.URL)

	// Create a handler that fails
	failHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	wrappedHandler := mw.WithAudit("TEST_RESOURCE")(failHandler)

	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	time.Sleep(100 * time.Millisecond)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestCreateFailure_AuditLoggingWithoutResourceID verifies that CREATE failures
// are logged even when ResourceID is empty (e.g., validation errors before resource creation)
func TestCreateFailure_AuditLoggingWithoutResourceID(t *testing.T) {
	// Reset global state
	ResetGlobalAuditMiddleware()

	eventReceived := make(chan ManagementEventRequest, 1)

	// Setup a mock audit service server
	auditServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var event ManagementEventRequest
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			t.Errorf("Failed to decode audit event: %v", err)
			return
		}

		// Verify this is a CREATE failure
		if event.EventType != "CREATE" {
			t.Errorf("Expected EventType CREATE, got %s", event.EventType)
		}
		if event.Status != "FAILURE" {
			t.Errorf("Expected Status FAILURE, got %s", event.Status)
		}
		// ResourceID should be empty for CREATE failures that occur before resource creation
		if event.Target.ResourceID != "" {
			t.Errorf("Expected empty ResourceID for CREATE failure, got %s", event.Target.ResourceID)
		}
		if event.Target.Resource != "TEST_RESOURCE" {
			t.Errorf("Expected Resource TEST_RESOURCE, got %s", event.Target.Resource)
		}

		eventReceived <- event
		w.WriteHeader(http.StatusCreated)
	}))
	defer auditServer.Close()

	mw := NewAuditMiddleware(auditServer.URL)

	// Create a handler that fails before setting ResourceID (simulating validation error)
	failHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handler fails immediately (e.g., validation error) before creating resource
		// ResourceID is never set, but audit event should still be logged
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "validation failed"}`))
	})

	wrappedHandler := mw.WithAudit("TEST_RESOURCE")(failHandler)

	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	req.Header.Set("X-User-ID", "test-user")
	req.Header.Set("X-User-Role", "MEMBER")
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	// Wait for async audit log
	select {
	case event := <-eventReceived:
		// Verify event was received and logged
		if event.EventType != "CREATE" || event.Status != "FAILURE" {
			t.Errorf("Expected CREATE FAILURE event, got %s %s", event.EventType, event.Status)
		}
	case <-time.After(1 * time.Second):
		t.Error("Audit event was not received - CREATE failure was not logged!")
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestAuditMiddleware_ThreadSafety(t *testing.T) {
	// Reset global state for this test
	ResetGlobalAuditMiddleware()

	const numGoroutines = 10
	var wg sync.WaitGroup
	var instances []*AuditMiddleware
	var mu sync.Mutex

	// Start multiple goroutines trying to initialize audit middleware concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			url := "http://localhost:3001"
			if id%2 == 0 {
				url = "" // Mix enabled and disabled instances
			}

			instance := NewAuditMiddleware(url)

			mu.Lock()
			instances = append(instances, instance)
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Verify we have the expected number of instances
	if len(instances) != numGoroutines {
		t.Errorf("Expected %d instances, got %d", numGoroutines, len(instances))
	}

	// Verify that the global instance was set (should be one of the instances)
	if globalAuditMiddleware == nil {
		t.Error("Expected global audit middleware to be set")
	}

	// Test that LogAuditEvent works with the global instance
	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	req.Header.Set("X-User-ID", "test-user")
	req.Header.Set("X-User-Role", "ADMIN")

	// This should not panic
	LogAuditEvent(req, "TEST_RESOURCE", "test-id-concurrent")
}

func TestLogAuditEvent_WithoutInitialization(t *testing.T) {
	// Reset global state to ensure no global instance
	ResetGlobalAuditMiddleware()

	// Test LogAuditEvent when global middleware is not initialized
	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	req.Header.Set("X-User-ID", "test-user")
	req.Header.Set("X-User-Role", "ADMIN")

	// This should not panic and should log a warning
	LogAuditEvent(req, "TEST_RESOURCE", "test-id-no-init")
}
