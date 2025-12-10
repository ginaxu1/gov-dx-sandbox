package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestLogAuditEvent(t *testing.T) {
	// Reset global middleware before test
	ResetGlobalAuditMiddleware()

	// Create a test server to mock the audit service
	var receivedRequest *DataExchangeEventAuditRequest
	var mu sync.Mutex
	requestReceived := make(chan bool, 1)

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

		mu.Lock()
		receivedRequest = &auditReq
		mu.Unlock()

		w.WriteHeader(http.StatusCreated)
		requestReceived <- true
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

	// Wait for the async operation to complete with timeout
	select {
	case <-requestReceived:
		// Request received successfully
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for audit request")
	}

	// Verify the request was received
	mu.Lock()
	defer mu.Unlock()

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

func TestLogGeneralizedAudit(t *testing.T) {
	// Reset global middleware before test
	ResetGlobalAuditMiddleware()

	// Create a test server to mock the audit service
	var receivedRequest *CreateAuditLogRequest
	var mu sync.Mutex
	requestReceived := make(chan bool, 1)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/audit-logs" {
			t.Errorf("Expected path /api/audit-logs, got %s", r.URL.Path)
		}

		var auditReq CreateAuditLogRequest
		if err := json.NewDecoder(r.Body).Decode(&auditReq); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		mu.Lock()
		receivedRequest = &auditReq
		mu.Unlock()

		w.WriteHeader(http.StatusCreated)
		requestReceived <- true
	}))
	defer testServer.Close()

	// Initialize audit middleware with test server URL
	middleware := NewAuditMiddleware(testServer.URL)
	if middleware == nil {
		t.Fatal("Failed to create audit middleware")
	}

	// Test Case 1: TraceID provided in request
	t.Run("ExplicitTraceID", func(t *testing.T) {
		traceID := uuid.New().String()
		req := &CreateAuditLogRequest{
			TraceID:       traceID,
			SourceService: "test-service",
			EventType:     "TEST_EVENT",
			Status:        "SUCCESS",
		}

		LogGeneralizedAuditEvent(context.Background(), req)

		select {
		case <-requestReceived:
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for audit request")
		}

		mu.Lock()
		defer mu.Unlock()
		if receivedRequest.TraceID != traceID {
			t.Errorf("Expected TraceID %s, got %s", traceID, receivedRequest.TraceID)
		}
	})

	// Test Case 2: TraceID from Context
	t.Run("ContextTraceID", func(t *testing.T) {
		traceID := uuid.New().String()
		ctx := context.WithValue(context.Background(), TraceIDKey{}, traceID)
		
		req := &CreateAuditLogRequest{
			SourceService: "test-service",
			EventType:     "TEST_EVENT_CTX",
			Status:        "SUCCESS",
		}

		LogGeneralizedAuditEvent(ctx, req)

		select {
		case <-requestReceived:
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for audit request")
		}

		mu.Lock()
		defer mu.Unlock()
		if receivedRequest.TraceID != traceID {
			t.Errorf("Expected TraceID %s, got %s", traceID, receivedRequest.TraceID)
		}
	})
}

func TestGetTraceIDFromContext(t *testing.T) {
	traceID := "test-trace-id"
	ctx := context.WithValue(context.Background(), TraceIDKey{}, traceID)
	
	if got := GetTraceIDFromContext(ctx); got != traceID {
		t.Errorf("GetTraceIDFromContext() = %v, want %v", got, traceID)
	}

	emptyCtx := context.Background()
	if got := GetTraceIDFromContext(emptyCtx); got != "" {
		t.Errorf("GetTraceIDFromContext() = %v, want empty string", got)
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

func TestLogGeneralizedAuditWhenNotConfigured(t *testing.T) {
	ResetGlobalAuditMiddleware()
	middleware := NewAuditMiddleware("")
	
	req := &CreateAuditLogRequest{
		TraceID: "test",
		SourceService: "test",
	}
	
	middleware.LogGeneralizedAudit(context.Background(), req)
	// Should not panic
}

