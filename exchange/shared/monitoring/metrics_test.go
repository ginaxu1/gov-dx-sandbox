package monitoring

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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
	if !strings.Contains(body, "# HELP") && !strings.Contains(body, "# TYPE") {
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
	if !strings.Contains(metricsBody, "http_requests_total") {
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
		{"/consents/123", "/consents/:id"},
		{"/consents/abc123def456", "/consents/:id"},
		{"/consents/consent_abc123", "/consents/:id"},
		{"/data-owner/user@example.com", "/data-owner/:id"},
		{"/api/v1/policy/metadata", "/api/v1/policy/metadata"},
		{"/api/v1/policy/decide", "/api/v1/policy/decide"},
		{"/api/v1/policy/123", "/api/v1/policy/:id"},
		{"/test?query=value", "/test"},
		{"/consumer/app-123", "/consumer/:id"},
		{"/admin/check", "/admin/check"},
	}

	for _, tt := range tests {
		result := normalizeRoute(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeRoute(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestRecordExternalCall(t *testing.T) {
	// Record a successful external call
	RecordExternalCall("postgres", "create_consent", 100*time.Millisecond, nil)

	// Record a failed external call
	RecordExternalCall("postgres", "create_consent", 50*time.Millisecond, fmt.Errorf("connection failed"))

	// Verify metrics were recorded (check via /metrics endpoint)
	metricsHandler := Handler()
	metricsReq := httptest.NewRequest("GET", "/metrics", nil)
	metricsW := httptest.NewRecorder()
	metricsHandler.ServeHTTP(metricsW, metricsReq)

	metricsBody := metricsW.Body.String()
	if !strings.Contains(metricsBody, "external_calls_total") {
		t.Error("external_calls_total metric not found")
	}
	if !strings.Contains(metricsBody, "external_call_errors_total") {
		t.Error("external_call_errors_total metric not found")
	}
	if !strings.Contains(metricsBody, "external_call_duration_seconds") {
		t.Error("external_call_duration_seconds metric not found")
	}
}

func TestRecordBusinessEvent(t *testing.T) {
	// Record business events
	RecordBusinessEvent("consent_created", "success")
	RecordBusinessEvent("consent_approved", "success")
	RecordBusinessEvent("policy_decision", "allow")

	// Verify metrics were recorded (check via /metrics endpoint)
	metricsHandler := Handler()
	metricsReq := httptest.NewRequest("GET", "/metrics", nil)
	metricsW := httptest.NewRecorder()
	metricsHandler.ServeHTTP(metricsW, metricsReq)

	metricsBody := metricsW.Body.String()
	if !strings.Contains(metricsBody, "business_events_total") {
		t.Error("business_events_total metric not found")
	}
}
