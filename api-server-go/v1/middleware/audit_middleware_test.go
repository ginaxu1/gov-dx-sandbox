package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuditMiddleware_Basic(t *testing.T) {
	// Test with audit enabled
	auditMiddleware := NewAuditMiddleware("http://localhost:8080")
	if auditMiddleware.auditServiceURL == "" {
		t.Error("Expected audit middleware to have service URL when URL is provided")
	}

	// Test with audit disabled
	auditMiddleware = NewAuditMiddleware("")
	if auditMiddleware.auditServiceURL != "" {
		t.Error("Expected audit middleware to have empty service URL when URL is empty")
	}

	// Test middleware functionality
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Apply audit middleware
	middlewareHandler := auditMiddleware.AuditLoggingMiddleware(testHandler)
	middlewareHandler.ServeHTTP(w, req)

	// Should complete without error
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "test response" {
		t.Errorf("Expected 'test response', got %s", w.Body.String())
	}
}

func TestAuditMiddleware_Passthrough(t *testing.T) {
	// Create audit middleware
	auditMiddleware := NewAuditMiddleware("http://localhost:8080")

	// Test that it passes through requests without modification
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test-Header", "test-value")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created"))
	})

	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	w := httptest.NewRecorder()

	// Apply audit middleware
	middlewareHandler := auditMiddleware.AuditLoggingMiddleware(testHandler)
	middlewareHandler.ServeHTTP(w, req)

	// Verify response is unchanged
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	if w.Body.String() != "created" {
		t.Errorf("Expected 'created', got %s", w.Body.String())
	}

	if w.Header().Get("X-Test-Header") != "test-value" {
		t.Errorf("Expected header 'test-value', got %s", w.Header().Get("X-Test-Header"))
	}
}
