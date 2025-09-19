package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTrustedAuthMiddleware(t *testing.T) {
	// Create a test handler that returns success
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Create middleware
	middleware := TrustedAuthMiddleware(testHandler)

	tests := []struct {
		name           string
		headers        map[string]string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Valid request with X-Consumer-ID header",
			headers: map[string]string{
				"X-Consumer-ID": "test-consumer-123",
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name: "Missing X-Consumer-ID header",
			headers: map[string]string{
				"Authorization": "Bearer test-token",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"errors":[{"extensions":{"code":"UNAUTHENTICATED"},"message":"X-Consumer-ID header is required - request not properly authenticated by Choreo Gateway"}]}`,
		},
		{
			name:           "No headers",
			headers:        map[string]string{},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"errors":[{"extensions":{"code":"UNAUTHENTICATED"},"message":"X-Consumer-ID header is required - request not properly authenticated by Choreo Gateway"}]}`,
		},
		{
			name: "Empty X-Consumer-ID header",
			headers: map[string]string{
				"X-Consumer-ID": "",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"errors":[{"extensions":{"code":"UNAUTHENTICATED"},"message":"X-Consumer-ID header is required - request not properly authenticated by Choreo Gateway"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("POST", "/", nil)

			// Add headers
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call middleware
			middleware.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check response body (trim newlines for comparison)
			actualBody := strings.TrimSpace(rr.Body.String())
			expectedBody := strings.TrimSpace(tt.expectedBody)
			if actualBody != expectedBody {
				t.Errorf("expected body %q, got %q", expectedBody, actualBody)
			}
		})
	}
}

func TestTrustedAuthMiddlewareHealthCheck(t *testing.T) {
	// Create a test handler that returns success
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("health check"))
	})

	// Create middleware
	middleware := TrustedAuthMiddleware(testHandler)

	// Test health check endpoint (should bypass authentication)
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	// Should succeed without X-Consumer-ID header
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if rr.Body.String() != "health check" {
		t.Errorf("expected body %q, got %q", "health check", rr.Body.String())
	}
}

func TestTrustedAuthMiddlewareOptions(t *testing.T) {
	// Create a test handler that returns success
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("options"))
	})

	// Create middleware
	middleware := TrustedAuthMiddleware(testHandler)

	// Test OPTIONS request (should bypass authentication)
	req := httptest.NewRequest("OPTIONS", "/", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	// Should succeed without X-Consumer-ID header
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if rr.Body.String() != "options" {
		t.Errorf("expected body %q, got %q", "options", rr.Body.String())
	}
}
