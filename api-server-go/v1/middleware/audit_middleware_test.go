package middleware

import (
	"context"
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

func TestWithAudit_InjectsAuditInfo(t *testing.T) {
	// Reset global state
	ResetGlobalAuditMiddleware()
	mw := NewAuditMiddleware("http://localhost:3001")

	// Create a handler that checks for AuditInfo
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auditInfo := GetAuditInfo(r.Context())
		if auditInfo == nil {
			t.Error("Expected AuditInfo to be injected into context")
			return
		}
		
		// Verify we can set resource ID
		auditInfo.SetResourceID("test-resource-id")
		
		// Verify it persists
		if auditInfo.ResourceID != "test-resource-id" {
			t.Error("Expected ResourceID to be set")
		}
		
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with middleware
	wrappedHandler := mw.WithAudit("TEST_RESOURCE")(handler)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()

	// Execute
	wrappedHandler.ServeHTTP(w, req)
}

func TestWithAudit_SkipsReadOperations(t *testing.T) {
	// Reset global state
	ResetGlobalAuditMiddleware()
	mw := NewAuditMiddleware("http://localhost:3001")

	// Create a handler that checks for AuditInfo (should NOT be present for GET)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auditInfo := GetAuditInfo(r.Context())
		if auditInfo != nil {
			t.Error("Expected AuditInfo NOT to be injected for GET request")
		}
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with middleware
	wrappedHandler := mw.WithAudit("TEST_RESOURCE")(handler)

	// Create GET request
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	// Execute
	wrappedHandler.ServeHTTP(w, req)
}

func TestGetAuditInfo_NilContext(t *testing.T) {
	info := GetAuditInfo(nil)
	if info != nil {
		t.Error("Expected nil AuditInfo for nil context")
	}
}

func TestGetAuditInfo_NoAuditInfo(t *testing.T) {
	ctx := context.Background()
	info := GetAuditInfo(ctx)
	if info != nil {
		t.Error("Expected nil AuditInfo for context without audit info")
	}
}

// TestCreateRequest_AuditLogging verifies that a CREATE request triggers the audit logic.
// Since we can't easily mock the internal http client of the middleware without refactoring,
// we verify that the middleware runs without panic and the context injection works as expected.
// In a real integration test, we would point the audit URL to a mock server.
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
		auditInfo := GetAuditInfo(r.Context())
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
}
