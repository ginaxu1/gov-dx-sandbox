package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	logger.Init()
}

func TestNewAuditMiddleware(t *testing.T) {
	middleware := NewAuditMiddleware("test-env", "http://audit-service.com", nil)
	assert.NotNil(t, middleware)
	assert.Equal(t, "test-env", middleware.environment)
	assert.Equal(t, "http://audit-service.com", middleware.auditServiceURL)
	assert.NotNil(t, middleware.httpClient)
	assert.Equal(t, 10*time.Second, middleware.httpClient.Timeout)
}

func TestAuditMiddleware_AuditHandler_Success(t *testing.T) {
	// Create a test server to capture audit requests
	var capturedRequest *AuditLogRequest
	auditReceived := make(chan bool, 1)
	auditServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req AuditLogRequest
		json.NewDecoder(r.Body).Decode(&req)
		capturedRequest = &req
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(AuditLogResponse{
			ID:        "audit-123",
			Timestamp: time.Now(),
			Status:    "success",
		})
		auditReceived <- true
	}))
	defer auditServer.Close()

	middleware := NewAuditMiddleware("test-env", auditServer.URL, nil)

	handler := middleware.AuditHandler(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(`{"query": "test"}`))
	w := httptest.NewRecorder()

	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())

	// Wait for async audit log to be received
	select {
	case <-auditReceived:
		// Verify audit request was sent
		require.NotNil(t, capturedRequest, "Audit request should have been captured")
		assert.Equal(t, "success", capturedRequest.Status)
		assert.Contains(t, capturedRequest.RequestedData, "query")
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for audit log to be sent")
	}
}

func TestAuditMiddleware_AuditHandler_ErrorResponse(t *testing.T) {
	auditReceived := make(chan bool, 1)
	auditServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		auditReceived <- true
	}))
	defer auditServer.Close()

	middleware := NewAuditMiddleware("test-env", auditServer.URL, nil)

	handler := middleware.AuditHandler(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error"))
	})

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(`{"query": "test"}`))
	w := httptest.NewRecorder()

	handler(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "Error", w.Body.String())

	// Wait for async audit log to be received
	select {
	case <-auditReceived:
		// Audit log was sent successfully
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for audit log to be sent")
	}
}

func TestAuditMiddleware_ExtractAuditInfo(t *testing.T) {
	middleware := NewAuditMiddleware("test-env", "http://audit-service.com", nil)

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(`{"query": "test"}`))

	appID, schemaID := middleware.ExtractAuditInfo(req, []byte(`{"query": "test"}`))

	// Should return fallback values when JWT extraction fails
	assert.NotEmpty(t, appID)
	assert.NotEmpty(t, schemaID)
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	rw := &responseWriter{
		ResponseWriter: httptest.NewRecorder(),
		statusCode:     http.StatusOK,
		body:           &bytes.Buffer{},
	}

	rw.WriteHeader(http.StatusNotFound)
	assert.Equal(t, http.StatusNotFound, rw.statusCode)
}

func TestResponseWriter_Write(t *testing.T) {
	rw := &responseWriter{
		ResponseWriter: httptest.NewRecorder(),
		statusCode:     http.StatusOK,
		body:           &bytes.Buffer{},
	}

	n, err := rw.Write([]byte("test data"))
	assert.NoError(t, err)
	assert.Equal(t, 9, n)
	assert.Equal(t, "test data", rw.body.String())
}
