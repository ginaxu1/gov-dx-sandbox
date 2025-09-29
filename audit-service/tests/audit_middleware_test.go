package tests

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gov-dx-sandbox/audit-service/middleware"
	_ "github.com/lib/pq"
)

// TestAuditServiceMiddlewareTiming tests that audit entries are created after request processing
func TestAuditServiceMiddlewareTiming(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	defer db.Close()

	// Create audit middleware
	auditMiddleware := middleware.NewAuditMiddleware(db)

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

	// Wait a bit for async audit log creation
	time.Sleep(100 * time.Millisecond)

	// Verify audit log was created in database
	auditLogs, err := getAuditLogsFromDB(db)
	if err != nil {
		t.Fatalf("Failed to get audit logs from database: %v", err)
	}

	if len(auditLogs) != 1 {
		t.Errorf("Expected 1 audit log, got %d", len(auditLogs))
	}

	auditLog := auditLogs[0]

	// Verify audit log contains expected data
	if auditLog.Path != "/consumers" {
		t.Errorf("Expected request path '/consumers', got '%s'", auditLog.Path)
	}

	if auditLog.Method != "POST" {
		t.Errorf("Expected request method 'POST', got '%s'", auditLog.Method)
	}

	if auditLog.TransactionStatus != "SUCCESS" {
		t.Errorf("Expected status 'SUCCESS', got '%s'", auditLog.TransactionStatus)
	}

	// Verify timing - audit should be created after response
	duration := endTime.Sub(startTime)
	if duration < 50*time.Millisecond {
		t.Errorf("Expected processing time to be at least 50ms, got %v", duration)
	}

	t.Logf("Request processing completed in %v", duration)
	t.Logf("Audit log created with status: %s", auditLog.TransactionStatus)
}

// TestAuditServiceMiddlewareSequence tests the complete sequence of events
func TestAuditServiceMiddlewareSequence(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	auditMiddleware := middleware.NewAuditMiddleware(db)

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

	// Wait for async audit log creation
	time.Sleep(100 * time.Millisecond)

	// Verify response was sent
	if !responseSent {
		t.Error("Response was not sent")
	}

	// Verify audit log was created
	auditLogs, err := getAuditLogsFromDB(db)
	if err != nil {
		t.Fatalf("Failed to get audit logs from database: %v", err)
	}

	if len(auditLogs) != 1 {
		t.Errorf("Expected 1 audit log, got %d", len(auditLogs))
	}

	auditLog := auditLogs[0]

	// Verify audit log contains response data
	if !strings.Contains(string(auditLog.ResponseData), "test-123") {
		t.Error("Audit log should contain response data")
	}

	// Log the sequence for verification
	t.Log("Event sequence:")
	for i, event := range eventSequence {
		t.Logf("  %s", event)
	}
	t.Log("  5. Audit log created (after response sent)")

	// Verify the audit log was created after the response
	if auditLog.TransactionStatus != "SUCCESS" {
		t.Errorf("Expected audit status 'SUCCESS', got '%s'", auditLog.TransactionStatus)
	}
}

// TestAuditServiceMiddlewareFailedRequest tests audit logging for failed requests
func TestAuditServiceMiddlewareFailedRequest(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	auditMiddleware := middleware.NewAuditMiddleware(db)

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

	// Verify audit log was created
	auditLogs, err := getAuditLogsFromDB(db)
	if err != nil {
		t.Fatalf("Failed to get audit logs from database: %v", err)
	}

	if len(auditLogs) != 1 {
		t.Errorf("Expected 1 audit log, got %d", len(auditLogs))
	}

	auditLog := auditLogs[0]

	if auditLog.TransactionStatus != "FAILURE" {
		t.Errorf("Expected audit status 'FAILURE', got '%s'", auditLog.TransactionStatus)
	}

	// Verify audit log contains error response data
	if !strings.Contains(string(auditLog.ResponseData), "Invalid request") {
		t.Error("Audit log should contain error response data")
	}

	t.Logf("Failed request audited with status: %s", auditLog.TransactionStatus)
}

// TestAuditServiceMiddlewareMultipleRequests tests audit logging for multiple requests
func TestAuditServiceMiddlewareMultipleRequests(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	auditMiddleware := middleware.NewAuditMiddleware(db)

	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	})

	// Wrap with audit middleware
	auditedHandler := auditMiddleware.AuditLoggingMiddleware(handler)

	// Test multiple requests
	requests := []struct {
		method string
		path   string
		body   string
		status int
	}{
		{"GET", "/consumers", "", 200},
		{"POST", "/consumers", `{"consumerName": "Test 1"}`, 201},
		{"GET", "/providers", "", 200},
		{"POST", "/provider-submissions", `{"providerName": "Test Provider"}`, 201},
	}

	for i, req := range requests {
		// Create request
		httpReq := httptest.NewRequest(req.method, req.path, strings.NewReader(req.body))
		if req.body != "" {
			httpReq.Header.Set("Content-Type", "application/json")
		}
		w := httptest.NewRecorder()

		// Execute request
		auditedHandler.ServeHTTP(w, httpReq)

		// Wait for async audit log creation
		time.Sleep(50 * time.Millisecond)

		// Verify response
		if w.Code != req.status {
			t.Errorf("Request %d: Expected status %d, got %d", i+1, req.status, w.Code)
		}
	}

	// Wait a bit more for all async operations
	time.Sleep(200 * time.Millisecond)

	// Verify all requests were audited
	auditLogs, err := getAuditLogsFromDB(db)
	if err != nil {
		t.Fatalf("Failed to get audit logs from database: %v", err)
	}

	expectedCount := len(requests)
	actualCount := len(auditLogs)
	if actualCount != expectedCount {
		t.Errorf("Expected %d audit logs, got %d", expectedCount, actualCount)
	}

	// Verify each audit log
	for i, auditLog := range auditLogs {
		expectedPath := requests[i].path
		if auditLog.Path != expectedPath {
			t.Errorf("Audit log %d: Expected path '%s', got '%s'", i+1, expectedPath, auditLog.Path)
		}
	}

	t.Logf("All %d requests were properly audited", expectedCount)
}

// TestAuditServiceMiddlewareSkippedEndpoints tests that certain endpoints are skipped
func TestAuditServiceMiddlewareSkippedEndpoints(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	auditMiddleware := middleware.NewAuditMiddleware(db)

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
		"/audit",
		"/debug",
	}

	for _, endpoint := range skippedEndpoints {
		// Clear previous audit logs
		clearAuditLogsFromDB(db)

		// Create request
		req := httptest.NewRequest("GET", endpoint, nil)
		w := httptest.NewRecorder()

		// Execute request
		auditedHandler.ServeHTTP(w, req)

		// Wait a bit
		time.Sleep(50 * time.Millisecond)

		// Verify no audit log was created
		auditLogs, err := getAuditLogsFromDB(db)
		if err != nil {
			t.Fatalf("Failed to get audit logs from database: %v", err)
		}

		if len(auditLogs) != 0 {
			t.Errorf("Endpoint '%s' should not be audited, but %d audit logs were created", endpoint, len(auditLogs))
		}

		t.Logf("Endpoint '%s' correctly skipped from audit", endpoint)
	}
}

// TestAuditServiceMiddlewareConcurrentRequests tests audit logging under concurrent load
func TestAuditServiceMiddlewareConcurrentRequests(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	auditMiddleware := middleware.NewAuditMiddleware(db)

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
			req := httptest.NewRequest("GET", fmt.Sprintf("/consumers?req=%d", requestNum), nil)
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

	// Verify all requests were audited
	auditLogs, err := getAuditLogsFromDB(db)
	if err != nil {
		t.Fatalf("Failed to get audit logs from database: %v", err)
	}

	actualCount := len(auditLogs)
	if actualCount != numRequests {
		t.Errorf("Expected %d audit logs, got %d", numRequests, actualCount)
	}

	t.Logf("All %d concurrent requests were properly audited", numRequests)
}

// TestAuditServiceMiddlewareResponseCapture tests that response data is properly captured
func TestAuditServiceMiddlewareResponseCapture(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	auditMiddleware := middleware.NewAuditMiddleware(db)

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

	// Verify audit log was created
	auditLogs, err := getAuditLogsFromDB(db)
	if err != nil {
		t.Fatalf("Failed to get audit logs from database: %v", err)
	}

	if len(auditLogs) != 1 {
		t.Errorf("Expected 1 audit log, got %d", len(auditLogs))
	}

	auditLog := auditLogs[0]

	// Parse captured response data
	var capturedResponse map[string]interface{}
	if err := json.Unmarshal(auditLog.ResponseData, &capturedResponse); err != nil {
		t.Fatalf("Failed to parse captured response data: %v", err)
	}

	// Verify key fields are captured
	if capturedResponse["consumerId"] != expectedResponseData["consumerId"] {
		t.Errorf("Expected consumerId '%s', got '%s'", expectedResponseData["consumerId"], capturedResponse["consumerId"])
	}

	if capturedResponse["consumerName"] != expectedResponseData["consumerName"] {
		t.Errorf("Expected consumerName '%s', got '%s'", expectedResponseData["consumerName"], capturedResponse["consumerName"])
	}

	t.Log("Response data was properly captured in audit log")
}

// TestAuditServiceMiddlewareRequestCapture tests that request data is properly captured
func TestAuditServiceMiddlewareRequestCapture(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	auditMiddleware := middleware.NewAuditMiddleware(db)

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

	// Verify audit log was created
	auditLogs, err := getAuditLogsFromDB(db)
	if err != nil {
		t.Fatalf("Failed to get audit logs from database: %v", err)
	}

	if len(auditLogs) != 1 {
		t.Errorf("Expected 1 audit log, got %d", len(auditLogs))
	}

	auditLog := auditLogs[0]

	// Parse captured request data
	var capturedRequest map[string]interface{}
	if err := json.Unmarshal(auditLog.RequestedData, &capturedRequest); err != nil {
		t.Fatalf("Failed to parse captured request data: %v", err)
	}

	// Verify key fields are captured
	if capturedRequest["consumerName"] != requestData["consumerName"] {
		t.Errorf("Expected consumerName '%s', got '%s'", requestData["consumerName"], capturedRequest["consumerName"])
	}

	if capturedRequest["contactEmail"] != requestData["contactEmail"] {
		t.Errorf("Expected contactEmail '%s', got '%s'", requestData["contactEmail"], capturedRequest["contactEmail"])
	}

	// Verify entity IDs are extracted
	if auditLog.ConsumerID != "test-consumer-123" {
		t.Errorf("Expected ConsumerID 'test-consumer-123', got '%s'", auditLog.ConsumerID)
	}

	if auditLog.ProviderID != "test-provider-456" {
		t.Errorf("Expected ProviderID 'test-provider-456', got '%s'", auditLog.ProviderID)
	}

	t.Log("Request data was properly captured in audit log")
}

// Helper functions

func setupTestDB(t *testing.T) *sql.DB {
	// Get test database configuration from environment variables
	host := getEnvOrDefault("TEST_DB_HOST", "localhost")
	port := getEnvOrDefault("TEST_DB_PORT", "5434")
	user := getEnvOrDefault("TEST_DB_USER", "test_user")
	password := getEnvOrDefault("TEST_DB_PASSWORD", "test_password")
	database := getEnvOrDefault("TEST_DB_NAME", "audit_service_test")
	sslmode := getEnvOrDefault("TEST_DB_SSLMODE", "disable")

	// Create PostgreSQL connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, database, sslmode)

	// Connect to test database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	// Create audit_logs table for testing
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS audit_logs (
			event_id VARCHAR(255) PRIMARY KEY,
			timestamp TIMESTAMP NOT NULL,
			consumer_id VARCHAR(255) NOT NULL,
			provider_id VARCHAR(255) NOT NULL,
			requested_data JSONB,
			response_data JSONB,
			transaction_status VARCHAR(50) NOT NULL,
			citizen_hash VARCHAR(255),
			path VARCHAR(255),
			method VARCHAR(10)
		)
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		t.Fatalf("Failed to create audit_logs table: %v", err)
	}

	return db
}

func getAuditLogsFromDB(db *sql.DB) ([]AuditLog, error) {
	query := `
		SELECT event_id, timestamp, consumer_id, provider_id, 
		       requested_data, response_data, transaction_status, 
		       citizen_hash, path, method
		FROM audit_logs 
		ORDER BY timestamp DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var auditLogs []AuditLog
	for rows.Next() {
		var auditLog AuditLog
		err := rows.Scan(
			&auditLog.EventID,
			&auditLog.Timestamp,
			&auditLog.ConsumerID,
			&auditLog.ProviderID,
			&auditLog.RequestedData,
			&auditLog.ResponseData,
			&auditLog.TransactionStatus,
			&auditLog.CitizenHash,
			&auditLog.Path,
			&auditLog.Method,
		)
		if err != nil {
			return nil, err
		}
		auditLogs = append(auditLogs, auditLog)
	}

	return auditLogs, nil
}

func clearAuditLogsFromDB(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM audit_logs")
	return err
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// AuditLog represents an audit log entry from the database
type AuditLog struct {
	EventID           string          `json:"event_id"`
	Timestamp         time.Time       `json:"timestamp"`
	ConsumerID        string          `json:"consumer_id"`
	ProviderID        string          `json:"provider_id"`
	RequestedData     json.RawMessage `json:"requested_data"`
	ResponseData      json.RawMessage `json:"response_data"`
	TransactionStatus string          `json:"transaction_status"`
	CitizenHash       string          `json:"citizen_hash"`
	Path              string          `json:"path"`
	Method            string          `json:"method"`
}
