package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
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
	auditMiddleware := NewAuditMiddleware("http://localhost:3001")

	// GET request should be skipped
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	auditMiddleware.LogAudit(req, "TEST_RESOURCE", "test-id")

	// This test passes if no panic occurs - we can't easily test HTTP calls without a mock server
}

func TestLogAudit_ProcessesWriteOperations(t *testing.T) {
	auditMiddleware := NewAuditMiddleware("http://localhost:3001")

	// POST request should be processed (though it may fail to send)
	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	req.Header.Set("X-User-ID", "test-user")
	req.Header.Set("X-User-Role", "MEMBER")

	auditMiddleware.LogAudit(req, "TEST_RESOURCE", "test-id")

	// This test passes if no panic occurs - we can't easily test HTTP calls without a mock server
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
