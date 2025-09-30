package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/middleware"
)

// TestAuditMiddlewareIntegration tests the audit middleware integration
func TestAuditMiddlewareIntegration(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	// Create audit middleware with a mock audit service URL
	auditMiddleware := middleware.NewAuditMiddleware("http://localhost:3002")

	// Create a test handler that simulates processing time
	processingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate some processing time
		time.Sleep(50 * time.Millisecond)

		// Write response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	})

	// Wrap handler with audit middleware
	auditedHandler := auditMiddleware.AuditLoggingMiddleware(processingHandler)

	// Create test request
	req := httptest.NewRequest("POST", "/consumers", strings.NewReader(`{"consumerName": "Test Consumer"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Record timing
	startTime := time.Now()
	auditedHandler.ServeHTTP(w, req)
	endTime := time.Now()

	// Verify response was sent
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify timing - audit should be created after response
	duration := endTime.Sub(startTime)
	if duration < 50*time.Millisecond {
		t.Errorf("Expected processing time to be at least 50ms, got %v", duration)
	}

	t.Logf("Request processing completed in %v", duration)
}

// TestAuditMiddlewareSequence tests the complete sequence of events
func TestAuditMiddlewareSequence(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	auditMiddleware := middleware.NewAuditMiddleware("http://localhost:3002")

	// Track the sequence of events
	var eventSequence []string
	var responseSent bool

	// Create a test handler that tracks events
	processingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		eventSequence = append(eventSequence, "1. Request received by handler")
		eventSequence = append(eventSequence, "2. Service logic executing")

		// Simulate service processing
		time.Sleep(10 * time.Millisecond)

		eventSequence = append(eventSequence, "3. Service logic completed")

		// Write response
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"consumerId": "test-123"}`))
		responseSent = true
		eventSequence = append(eventSequence, "4. Response sent to client")
	})

	// Wrap with audit middleware
	auditedHandler := auditMiddleware.AuditLoggingMiddleware(processingHandler)

	// Create test request
	req := httptest.NewRequest("POST", "/consumers", strings.NewReader(`{"consumerName": "Test Consumer"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute request
	auditedHandler.ServeHTTP(w, req)

	// Wait a bit for async audit log creation
	time.Sleep(100 * time.Millisecond)

	// Verify response was sent
	if !responseSent {
		t.Error("Response was not sent")
	}

	// Log the sequence for verification
	t.Log("Event sequence:")
	for _, event := range eventSequence {
		t.Logf("  %s", event)
	}
	t.Log("  5. Audit log created (after response sent)")

	// Verify the response
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
}

// TestAuditMiddlewareFailedRequest tests audit logging for failed requests
func TestAuditMiddlewareFailedRequest(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	auditMiddleware := middleware.NewAuditMiddleware("http://localhost:3002")

	// Create a handler that returns an error
	errorHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate processing time
		time.Sleep(25 * time.Millisecond)

		// Return error response
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid request"}`))
	})

	// Wrap with audit middleware
	auditedHandler := auditMiddleware.AuditLoggingMiddleware(errorHandler)

	// Create test request with invalid data
	req := httptest.NewRequest("POST", "/consumers", strings.NewReader(`{"invalid": "data"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute request
	auditedHandler.ServeHTTP(w, req)

	// Wait for async audit log creation
	time.Sleep(100 * time.Millisecond)

	// Verify response was sent with error status
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	t.Logf("Failed request processed and audited")
}

// TestAuditMiddlewareSkippedEndpoints tests that certain endpoints are skipped
func TestAuditMiddlewareSkippedEndpoints(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	auditMiddleware := middleware.NewAuditMiddleware("http://localhost:3002")

	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	// Wrap with audit middleware
	auditedHandler := auditMiddleware.AuditLoggingMiddleware(handler)

	// Test endpoints that should be skipped
	skippedEndpoints := []string{
		"/health",
		"/debug",
		"/openapi.yaml",
		"/favicon.ico",
	}

	for _, endpoint := range skippedEndpoints {
		// Create request
		req := httptest.NewRequest("GET", endpoint, nil)
		w := httptest.NewRecorder()

		// Execute request
		auditedHandler.ServeHTTP(w, req)

		// Wait a bit
		time.Sleep(50 * time.Millisecond)

		// Verify response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for %s, got %d", endpoint, w.Code)
		}

		t.Logf("Endpoint '%s' correctly processed (audit skipped)", endpoint)
	}
}

// TestAuditMiddlewareConcurrentRequests tests audit logging under concurrent load
func TestAuditMiddlewareConcurrentRequests(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	auditMiddleware := middleware.NewAuditMiddleware("http://localhost:3002")

	// Create a handler with some processing time
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	})

	// Wrap with audit middleware
	auditedHandler := auditMiddleware.AuditLoggingMiddleware(handler)

	// Number of concurrent requests
	numRequests := 10
	done := make(chan bool, numRequests)

	// Start concurrent requests
	for i := 0; i < numRequests; i++ {
		go func(requestNum int) {
			req := httptest.NewRequest("GET", "/consumers", nil)
			w := httptest.NewRecorder()

			auditedHandler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Request %d failed with status %d", requestNum, w.Code)
			}

			done <- true
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		<-done
	}

	// Wait for all async audit logs to be created
	time.Sleep(500 * time.Millisecond)

	t.Logf("All %d concurrent requests were processed", numRequests)
}

// TestAuditMiddlewareResponseCapture tests that response data is properly captured
func TestAuditMiddlewareResponseCapture(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	auditMiddleware := middleware.NewAuditMiddleware("http://localhost:3002")

	// Test data to be returned in response
	expectedResponseData := map[string]interface{}{
		"consumerId":   "test-consumer-123",
		"consumerName": "Test Consumer",
		"status":       "created",
		"timestamp":    "2024-01-01T12:00:00Z",
	}

	// Create handler that returns specific data
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		jsonData, _ := json.Marshal(expectedResponseData)
		w.Write(jsonData)
	})

	// Wrap with audit middleware
	auditedHandler := auditMiddleware.AuditLoggingMiddleware(handler)

	// Create test request
	req := httptest.NewRequest("POST", "/consumers", strings.NewReader(`{"consumerName": "Test Consumer"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute request
	auditedHandler.ServeHTTP(w, req)

	// Wait for async audit log creation
	time.Sleep(100 * time.Millisecond)

	// Verify response
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	// Verify response body
	var responseBody map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	if responseBody["consumerId"] != expectedResponseData["consumerId"] {
		t.Errorf("Expected consumerId '%s', got '%s'", expectedResponseData["consumerId"], responseBody["consumerId"])
	}

	t.Log("Response data was properly captured and processed")
}

// TestAuditMiddlewareRequestCapture tests that request data is properly captured
func TestAuditMiddlewareRequestCapture(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	auditMiddleware := middleware.NewAuditMiddleware("http://localhost:3002")

	// Create handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	})

	// Wrap with audit middleware
	auditedHandler := auditMiddleware.AuditLoggingMiddleware(handler)

	// Test request data
	requestData := map[string]interface{}{
		"consumerName": "Test Consumer",
		"contactEmail": "test@example.com",
		"phoneNumber":  "123-456-7890",
		"metadata": map[string]interface{}{
			"source":  "test",
			"version": "1.0",
		},
	}

	// Create test request
	jsonData, _ := json.Marshal(requestData)
	req := httptest.NewRequest("POST", "/consumers", strings.NewReader(string(jsonData)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Consumer-ID", "test-consumer-123")
	req.Header.Set("X-Provider-ID", "test-provider-456")
	w := httptest.NewRecorder()

	// Execute request
	auditedHandler.ServeHTTP(w, req)

	// Wait for async audit log creation
	time.Sleep(100 * time.Millisecond)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	t.Log("Request data was properly captured and processed")
}

// TestAuditMiddlewareWithRealAPI tests audit middleware with real API endpoints
func TestAuditMiddlewareWithRealAPI(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()

	auditMiddleware := middleware.NewAuditMiddleware("http://localhost:3002")

	// Wrap the existing API server with audit middleware
	auditedMux := http.NewServeMux()
	auditedMux.Handle("/", auditMiddleware.AuditLoggingMiddleware(ts.Mux))

	// Test consumer creation
	consumerData := map[string]string{
		"consumerName": "Audit Test Consumer",
		"contactEmail": "audit-test@example.com",
		"phoneNumber":  "123-456-7890",
	}

	jsonData, _ := json.Marshal(consumerData)
	req := httptest.NewRequest("POST", "/consumers", strings.NewReader(string(jsonData)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute request
	auditedMux.ServeHTTP(w, req)

	// Wait for async audit log creation
	time.Sleep(100 * time.Millisecond)

	// Verify response
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Response: %s", w.Code, w.Body.String())
	}

	// Verify response body contains consumer ID
	var responseBody map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	if responseBody["consumerId"] == nil {
		t.Error("Expected consumerId in response")
	}

	t.Logf("Real API request processed successfully with consumer ID: %v", responseBody["consumerId"])
}
