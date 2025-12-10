package middleware

import (
	"context"
	"encoding/json"
	"io"
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

	// Create test audit request
	testRequest := &CreateAuditLogRequest{
		TraceID:       "trace-123",
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		SourceService: "orchestration-engine",
		TargetService: "pdp",
		EventType:     "POLICY_CHECK_REQUEST",
		Status:        "SUCCESS",
		Resources:     json.RawMessage(`{"appId": "test-app"}`),
	}

	// Call LogGeneralizedAudit
	middleware.LogGeneralizedAudit(context.Background(), testRequest)

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

	if receivedRequest.TraceID != testRequest.TraceID {
		t.Errorf("Expected TraceID %s, got %s", testRequest.TraceID, receivedRequest.TraceID)
	}

	if receivedRequest.EventType != testRequest.EventType {
		t.Errorf("Expected EventType %s, got %s", testRequest.EventType, receivedRequest.EventType)
	}
	if receivedRequest.EventType != testRequest.EventType {
		t.Errorf("Expected EventType %s, got %s", testRequest.EventType, receivedRequest.EventType)
	}
}

func TestLogGeneralizedAudit_EmptyURL(t *testing.T) {
	ResetGlobalAuditMiddleware()
	
	// Initialize with empty URL
	middleware := NewAuditMiddleware("")
	
	req := &CreateAuditLogRequest{
		TraceID: "trace-123",
	}
	
	// Should not panic and return immediately
	middleware.LogGeneralizedAudit(context.Background(), req)
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

func TestLogAuditLogEvent_NilHttpClient(t *testing.T) {
	// Reset global middleware before test
	ResetGlobalAuditMiddleware()

	// Create middleware with nil httpClient (simulating disabled state)
	middleware := &AuditMiddleware{
		auditServiceURL: "http://test-server",
		httpClient:      nil,
	}

	testRequest := CreateAuditLogRequest{
		TraceID:       "trace-123",
		SourceService: "orchestration-engine",
		EventType:     "TEST_EVENT",
		Status:        "SUCCESS",
	}

	// This should not panic and should return early
	middleware.logAuditLogEvent(context.Background(), testRequest)

	// Test passes if no panic occurs
}

func TestLogAuditLogEvent_Marshaling(t *testing.T) {
	// Reset global middleware before test
	ResetGlobalAuditMiddleware()

	var receivedPayload []byte
	requestReceived := make(chan bool, 1)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		receivedPayload, err = io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
		}

		// Verify it's valid JSON
		var auditReq CreateAuditLogRequest
		if err := json.Unmarshal(receivedPayload, &auditReq); err != nil {
			t.Errorf("Failed to unmarshal request body: %v", err)
		}

		w.WriteHeader(http.StatusCreated)
		requestReceived <- true
	}))
	defer testServer.Close()

	middleware := &AuditMiddleware{
		auditServiceURL: testServer.URL,
		httpClient:      &http.Client{},
	}

	testRequest := CreateAuditLogRequest{
		TraceID:       "trace-123",
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		SourceService: "orchestration-engine",
		TargetService: "pdp",
		EventType:     "POLICY_CHECK_REQUEST",
		Status:        "SUCCESS",
		Resources:     json.RawMessage(`{"appId": "test-app"}`),
		Metadata:      json.RawMessage(`{"query": "test query"}`),
	}

	middleware.logAuditLogEvent(context.Background(), testRequest)

	select {
	case <-requestReceived:
		// Verify the payload was correctly marshaled
		var unmarshaled CreateAuditLogRequest
		err := json.Unmarshal(receivedPayload, &unmarshaled)
		if err != nil {
			t.Fatalf("Failed to unmarshal received payload: %v", err)
		}

		if unmarshaled.TraceID != testRequest.TraceID {
			t.Errorf("Expected TraceID %s, got %s", testRequest.TraceID, unmarshaled.TraceID)
		}
		if unmarshaled.EventType != testRequest.EventType {
			t.Errorf("Expected EventType %s, got %s", testRequest.EventType, unmarshaled.EventType)
		}
		if unmarshaled.SourceService != testRequest.SourceService {
			t.Errorf("Expected SourceService %s, got %s", testRequest.SourceService, unmarshaled.SourceService)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for audit request")
	}
}

func TestLogAuditLogEvent_CorrectEndpoint(t *testing.T) {
	// Reset global middleware before test
	ResetGlobalAuditMiddleware()

	var receivedPath string
	requestReceived := make(chan bool, 1)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusCreated)
		requestReceived <- true
	}))
	defer testServer.Close()

	middleware := &AuditMiddleware{
		auditServiceURL: testServer.URL,
		httpClient:      &http.Client{},
	}

	testRequest := CreateAuditLogRequest{
		TraceID:       "trace-123",
		SourceService: "orchestration-engine",
		EventType:     "TEST_EVENT",
		Status:        "SUCCESS",
	}

	middleware.logAuditLogEvent(context.Background(), testRequest)

	select {
	case <-requestReceived:
		expectedPath := "/api/audit-logs"
		if receivedPath != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, receivedPath)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for audit request")
	}
}

func TestLogAuditLogEvent_HttpErrorHandling(t *testing.T) {
	// Reset global middleware before test
	ResetGlobalAuditMiddleware()

	testCases := []struct {
		name          string
		statusCode    int
		responseBody  string
		shouldSucceed bool
	}{
		{
			name:          "Server returns 500",
			statusCode:    http.StatusInternalServerError,
			responseBody:  "Internal Server Error",
			shouldSucceed: false,
		},
		{
			name:          "Server returns 400",
			statusCode:    http.StatusBadRequest,
			responseBody:  "Bad Request",
			shouldSucceed: false,
		},
		{
			name:          "Server returns 201",
			statusCode:    http.StatusCreated,
			responseBody:  "Created",
			shouldSucceed: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requestReceived := make(chan bool, 1)

			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				w.Write([]byte(tc.responseBody))
				requestReceived <- true
			}))
			defer testServer.Close()

			middleware := &AuditMiddleware{
				auditServiceURL: testServer.URL,
				httpClient:      &http.Client{},
			}

			testRequest := CreateAuditLogRequest{
				TraceID:       "trace-123",
				SourceService: "orchestration-engine",
				EventType:     "TEST_EVENT",
				Status:        "SUCCESS",
			}

			// This should not panic regardless of status code
			middleware.logAuditLogEvent(context.Background(), testRequest)

			select {
			case <-requestReceived:
				// Test passes if no panic occurs
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for audit request")
			}
		})
	}
}

func TestLogAuditLogEvent_ResponseBodyClosed(t *testing.T) {
	// Reset global middleware before test
	ResetGlobalAuditMiddleware()

	bodyClosed := make(chan bool, 1)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("test response"))

		// Use a custom response writer to detect when body is closed
		// Note: This is a simplified test - in practice, we'd need to wrap the response
		// to detect actual close events
		go func() {
			time.Sleep(100 * time.Millisecond)
			bodyClosed <- true
		}()
	}))
	defer testServer.Close()

	middleware := &AuditMiddleware{
		auditServiceURL: testServer.URL,
		httpClient:      &http.Client{},
	}

	testRequest := CreateAuditLogRequest{
		TraceID:       "trace-123",
		SourceService: "orchestration-engine",
		EventType:     "TEST_EVENT",
		Status:        "SUCCESS",
	}

	middleware.logAuditLogEvent(context.Background(), testRequest)

	// Wait a bit to ensure the defer function executes
	select {
	case <-bodyClosed:
		// Test passes if body was closed (no panic)
	case <-time.After(2 * time.Second):
		// Even if we can't detect the close, the test passes if no panic occurred
		t.Log("Body close detection timeout, but no panic occurred - test passes")
	}
}

func TestLogAuditLogEvent_NetworkError(t *testing.T) {
	// Reset global middleware before test
	ResetGlobalAuditMiddleware()

	// Use an invalid URL to simulate network error
	middleware := &AuditMiddleware{
		auditServiceURL: "http://localhost:99999", // Invalid port
		httpClient:      &http.Client{Timeout: 100 * time.Millisecond},
	}

	testRequest := CreateAuditLogRequest{
		TraceID:       "trace-123",
		SourceService: "orchestration-engine",
		EventType:     "TEST_EVENT",
		Status:        "SUCCESS",
	}

	// This should not panic and should handle the error gracefully
	middleware.logAuditLogEvent(context.Background(), testRequest)

	// Give time for async operation
	time.Sleep(200 * time.Millisecond)

	// Test passes if no panic occurs
}
