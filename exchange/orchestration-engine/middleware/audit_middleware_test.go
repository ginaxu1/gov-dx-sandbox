package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLogAuditEvent(t *testing.T) {
	// Reset global middleware before test
	ResetGlobalAuditMiddleware()

	// Create a test server to mock the audit service
	var receivedRequest *DataExchangeEventAuditRequest
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/data-exchange-events" {
			t.Errorf("Expected path /data-exchange-events, got %s", r.URL.Path)
		}

		var auditReq DataExchangeEventAuditRequest
		if err := json.NewDecoder(r.Body).Decode(&auditReq); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}
		receivedRequest = &auditReq

		w.WriteHeader(http.StatusCreated)
	}))
	defer testServer.Close()

	// Initialize audit middleware with test server URL
	middleware := NewAuditMiddleware(testServer.URL)
	if middleware == nil {
		t.Fatal("Failed to create audit middleware")
	}

	// Create test audit request
	testRequest := &DataExchangeEventAuditRequest{
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		Status:         "success",
		ApplicationID:  "test-app-123",
		SchemaID:       "schema-456",
		RequestedData:  json.RawMessage(`{"fields": ["field1", "field2"]}`),
		AdditionalInfo: json.RawMessage(`{"serviceKey": "test-service"}`),
	}

	// Call LogAuditEvent
	LogAuditEvent(testRequest)

	// Give some time for the async operation to complete
	time.Sleep(100 * time.Millisecond)

	// Verify the request was received
	if receivedRequest == nil {
		t.Fatal("No request received by test server")
	}

	if receivedRequest.ApplicationID != testRequest.ApplicationID {
		t.Errorf("Expected ApplicationID %s, got %s", testRequest.ApplicationID, receivedRequest.ApplicationID)
	}

	if receivedRequest.SchemaID != testRequest.SchemaID {
		t.Errorf("Expected SchemaID %s, got %s", testRequest.SchemaID, receivedRequest.SchemaID)
	}

	if receivedRequest.Status != testRequest.Status {
		t.Errorf("Expected Status %s, got %s", testRequest.Status, receivedRequest.Status)
	}
}

func TestLogAuditEventWhenNotConfigured(t *testing.T) {
	// Reset global middleware before test
	ResetGlobalAuditMiddleware()

	// Initialize audit middleware without URL (disabled)
	middleware := NewAuditMiddleware("")
	if middleware == nil {
		t.Fatal("Failed to create audit middleware")
	}

	// Create test audit request
	testRequest := &DataExchangeEventAuditRequest{
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Status:        "success",
		ApplicationID: "test-app-123",
		SchemaID:      "schema-456",
		RequestedData: json.RawMessage(`{"fields": ["field1", "field2"]}`),
	}

	// This should not panic or cause errors
	LogAuditEvent(testRequest)

	// Give some time for any potential async operation
	time.Sleep(50 * time.Millisecond)

	// Test passes if no panic occurs
}

func TestLogAuditEventWhenGlobalMiddlewareNotInitialized(t *testing.T) {
	// Reset global middleware to simulate uninitialized state
	ResetGlobalAuditMiddleware()

	// Create test audit request
	testRequest := &DataExchangeEventAuditRequest{
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Status:        "success",
		ApplicationID: "test-app-123",
		SchemaID:      "schema-456",
		RequestedData: json.RawMessage(`{"fields": ["field1", "field2"]}`),
	}

	// This should not panic or cause errors
	LogAuditEvent(testRequest)

	// Test passes if no panic occurs
}
