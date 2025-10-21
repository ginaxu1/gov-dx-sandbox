package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/configs"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/middleware"
)

func TestAuditMiddleware(t *testing.T) {
	// Initialize config for testing
	configs.AppConfig = &configs.Config{
		Environment: "local",
	}

	// Create a test server to mock the audit service
	auditServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/logs" {
			t.Errorf("Expected POST /api/logs, got %s %s", r.Method, r.URL.Path)
		}

		// Verify the request body
		var auditRequest middleware.AuditLogRequest
		if err := json.NewDecoder(r.Body).Decode(&auditRequest); err != nil {
			t.Errorf("Failed to decode audit request: %v", err)
		}

		// Verify required fields
		if auditRequest.Status == "" {
			t.Error("Status field is required")
		}
		if auditRequest.RequestedData == "" {
			t.Error("RequestedData field is required")
		}
		if auditRequest.ApplicationID == "" {
			t.Error("ApplicationID field is required")
		}
		if auditRequest.SchemaID == "" {
			t.Error("SchemaID field is required")
		}

		// Return success response
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": "test-audit-id"})
	}))
	defer auditServer.Close()

	// Create audit middleware
	auditMiddleware := middleware.NewAuditMiddleware(auditServer.URL, nil)

	// Test successful request
	t.Run("SuccessfulRequest", func(t *testing.T) {
		// Create a test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		// Wrap with audit middleware
		auditedHandler := auditMiddleware.AuditHandler(testHandler)

		// Create test request
		reqBody := `{"query": "test query"}`
		req := httptest.NewRequest("POST", "/", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		rr := httptest.NewRecorder()

		// Execute request
		auditedHandler(rr, req)

		// Verify response
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}
	})

	// Test failed request
	t.Run("FailedRequest", func(t *testing.T) {
		// Create a test handler that returns an error
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		})

		// Wrap with audit middleware
		auditedHandler := auditMiddleware.AuditHandler(testHandler)

		// Create test request
		reqBody := `{"query": "test query"}`
		req := httptest.NewRequest("POST", "/", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		rr := httptest.NewRecorder()

		// Execute request
		auditedHandler(rr, req)

		// Verify response
		if rr.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", rr.Code)
		}
	})
}
