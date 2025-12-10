package federator

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine/middleware"
	"github.com/stretchr/testify/assert"
)

func TestLogAuditHelper(t *testing.T) {
	// Mock the global audit middleware
	middleware.ResetGlobalAuditMiddleware()
	
	var receivedRequest *middleware.CreateAuditLogRequest
	requestReceived := make(chan bool, 1)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var auditReq middleware.CreateAuditLogRequest
		if err := json.NewDecoder(r.Body).Decode(&auditReq); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}
		receivedRequest = &auditReq
		requestReceived <- true
		w.WriteHeader(http.StatusCreated)
	}))
	defer testServer.Close()

	// Initialize audit middleware - this sets the global instance used by logAuditHelper
	_ = middleware.NewAuditMiddleware(testServer.URL)
	
	f := &Federator{}

	ctx := context.Background()
	traceID := "test-trace-id"
	eventType := "TEST_EVENT"
	status := "SUCCESS"
	targetService := "test-target"
	resources := map[string]interface{}{"key": "value"}
	metadata := map[string]interface{}{"meta": "data"}

	f.logAuditHelper(ctx, traceID, eventType, status, targetService, resources, metadata)

	select {
	case <-requestReceived:
		assert.NotNil(t, receivedRequest)
		assert.Equal(t, traceID, receivedRequest.TraceID)
		assert.Equal(t, eventType, receivedRequest.EventType)
		assert.Equal(t, status, receivedRequest.Status)
		assert.Equal(t, targetService, receivedRequest.TargetService)
		assert.Contains(t, string(receivedRequest.Resources), "key")
		assert.Contains(t, string(receivedRequest.Metadata), "meta")
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for audit log")
	}
}
