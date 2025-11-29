package monitoring

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	handler := Handler()
	if handler == nil {
		t.Fatal("Handler() returned nil")
	}

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Metrics handler returned empty body")
	}

	// Check for Prometheus format
	if !contains(body, "# HELP") && !contains(body, "# TYPE") {
		t.Error("Response doesn't appear to be in Prometheus format")
	}
}

func TestHTTPMetricsMiddleware(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Wrap with metrics middleware
	wrapped := HTTPMetricsMiddleware(testHandler)

	// Make a request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify metrics were recorded (check via /metrics endpoint)
	metricsHandler := Handler()
	metricsReq := httptest.NewRequest("GET", "/metrics", nil)
	metricsW := httptest.NewRecorder()
	metricsHandler.ServeHTTP(metricsW, metricsReq)

	metricsBody := metricsW.Body.String()
	if !contains(metricsBody, "http_requests_total") {
		t.Error("http_requests_total metric not found after request")
	}
}

func TestNormalizeRoute(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/", "/"},
		{"/health", "/health"},
		{"/consents", "/consents"},
		{"/consents/123", "/consents"},
		{"/data-owner/user@example.com", "/data-owner"},
		{"/api/v1/policy/metadata", "/api"},
		{"/test?query=value", "/test"},
	}

	for _, tt := range tests {
		result := normalizeRoute(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeRoute(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

