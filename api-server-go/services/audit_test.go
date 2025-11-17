package services

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/api-server-go/models"
	"github.com/stretchr/testify/assert"
)

func TestNewAuditService(t *testing.T) {
	service := NewAuditService("http://localhost:8080")
	assert.NotNil(t, service)
	assert.Equal(t, "http://localhost:8080", service.auditServiceURL)
	assert.NotNil(t, service.httpClient)
	assert.Equal(t, 5*time.Second, service.httpClient.Timeout)
}

func TestAuditService_DetermineTransactionStatus(t *testing.T) {
	service := NewAuditService("http://localhost:8080")

	t.Run("SuccessStatus", func(t *testing.T) {
		assert.Equal(t, "SUCCESS", service.DetermineTransactionStatus(200))
		assert.Equal(t, "SUCCESS", service.DetermineTransactionStatus(201))
		assert.Equal(t, "SUCCESS", service.DetermineTransactionStatus(299))
	})

	t.Run("FailureStatus", func(t *testing.T) {
		assert.Equal(t, "FAILURE", service.DetermineTransactionStatus(400))
		assert.Equal(t, "FAILURE", service.DetermineTransactionStatus(404))
		assert.Equal(t, "FAILURE", service.DetermineTransactionStatus(500))
	})
}

func TestAuditService_mapTransactionStatus(t *testing.T) {
	service := NewAuditService("http://localhost:8080")

	t.Run("MapSuccess", func(t *testing.T) {
		assert.Equal(t, "success", service.mapTransactionStatus("SUCCESS"))
	})

	t.Run("MapFailure", func(t *testing.T) {
		assert.Equal(t, "failure", service.mapTransactionStatus("FAILURE"))
		assert.Equal(t, "failure", service.mapTransactionStatus("OTHER"))
	})
}

func TestAuditService_ExtractConsumerIDFromRequest(t *testing.T) {
	service := NewAuditService("http://localhost:8080")

	t.Run("FromPath", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/consumers/consumer-123", nil)
		consumerID := service.ExtractConsumerIDFromRequest(req)
		assert.Equal(t, "consumer-123", consumerID)
	})

	t.Run("FromPath_ConsumerApplications", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/consumer-applications/app-456", nil)
		consumerID := service.ExtractConsumerIDFromRequest(req)
		assert.Equal(t, "app-456", consumerID)
	})

	t.Run("FromHeader_XConsumerID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Consumer-ID", "header-consumer-123")
		consumerID := service.ExtractConsumerIDFromRequest(req)
		assert.Equal(t, "header-consumer-123", consumerID)
	})

	t.Run("FromHeader_XUserID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-User-ID", "user-123")
		consumerID := service.ExtractConsumerIDFromRequest(req)
		assert.Equal(t, "user-123", consumerID)
	})

	t.Run("FromQueryParam", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test?consumerId=query-consumer-123", nil)
		consumerID := service.ExtractConsumerIDFromRequest(req)
		assert.Equal(t, "query-consumer-123", consumerID)
	})

	t.Run("FromQueryParam_Underscore", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test?consumer_id=query-consumer-456", nil)
		consumerID := service.ExtractConsumerIDFromRequest(req)
		assert.Equal(t, "query-consumer-456", consumerID)
	})

	t.Run("FromBody", func(t *testing.T) {
		body := map[string]string{"consumerId": "body-consumer-123"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(bodyBytes))
		consumerID := service.ExtractConsumerIDFromRequest(req)
		assert.Equal(t, "body-consumer-123", consumerID)
	})

	t.Run("NotFound", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		consumerID := service.ExtractConsumerIDFromRequest(req)
		assert.Equal(t, "", consumerID)
	})
}

func TestAuditService_ExtractProviderIDFromRequest(t *testing.T) {
	service := NewAuditService("http://localhost:8080")

	t.Run("FromPath", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/providers/provider-123", nil)
		providerID := service.ExtractProviderIDFromRequest(req)
		assert.Equal(t, "provider-123", providerID)
	})

	t.Run("FromHeader", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Provider-ID", "header-provider-123")
		providerID := service.ExtractProviderIDFromRequest(req)
		assert.Equal(t, "header-provider-123", providerID)
	})

	t.Run("FromQueryParam", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test?providerId=query-provider-123", nil)
		providerID := service.ExtractProviderIDFromRequest(req)
		assert.Equal(t, "query-provider-123", providerID)
	})

	t.Run("FromBody", func(t *testing.T) {
		body := map[string]string{"providerId": "body-provider-123"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(bodyBytes))
		providerID := service.ExtractProviderIDFromRequest(req)
		assert.Equal(t, "body-provider-123", providerID)
	})

	t.Run("NotFound", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		providerID := service.ExtractProviderIDFromRequest(req)
		assert.Equal(t, "", providerID)
	})
}

func TestAuditService_ExtractGraphQLQueryFromRequest(t *testing.T) {
	service := NewAuditService("http://localhost:8080")

	t.Run("FromHeader", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/graphql", nil)
		req.Header.Set("X-GraphQL-Query", "query { test }")
		query := service.ExtractGraphQLQueryFromRequest(req)
		assert.Equal(t, "query { test }", query)
	})

	t.Run("FromQueryParam", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/graphql?query=query+%7B+test+%7D", nil)
		query := service.ExtractGraphQLQueryFromRequest(req)
		assert.Contains(t, query, "test")
	})

	t.Run("FromBody_JSON", func(t *testing.T) {
		body := map[string]string{"query": "query { test }"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/graphql", bytes.NewBuffer(bodyBytes))
		query := service.ExtractGraphQLQueryFromRequest(req)
		assert.Equal(t, "query { test }", query)
	})

	t.Run("FromBody_RawGraphQL", func(t *testing.T) {
		body := "query { test }"
		req := httptest.NewRequest("POST", "/graphql", bytes.NewBufferString(body))
		query := service.ExtractGraphQLQueryFromRequest(req)
		assert.Equal(t, body, query)
	})

	t.Run("NotFound", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		query := service.ExtractGraphQLQueryFromRequest(req)
		assert.Equal(t, "", query)
	})
}

func TestAuditService_ExtractAllFromRequestWithBody(t *testing.T) {
	service := NewAuditService("http://localhost:8080")

	t.Run("ExtractAll", func(t *testing.T) {
		body := map[string]interface{}{
			"consumerId": "consumer-123",
			"providerId": "provider-456",
			"query":      "query { test }",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(bodyBytes))

		consumerID, providerID, graphqlQuery, bodyData, err := service.ExtractAllFromRequestWithBody(req)

		assert.NoError(t, err)
		assert.Equal(t, "consumer-123", consumerID)
		assert.Equal(t, "provider-456", providerID)
		assert.Equal(t, "query { test }", graphqlQuery)
		assert.Equal(t, bodyBytes, bodyData)
	})

	t.Run("ReadBodyError", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", &errorReader{})
		consumerID, providerID, graphqlQuery, bodyData, err := service.ExtractAllFromRequestWithBody(req)

		assert.Error(t, err)
		assert.Equal(t, "", consumerID)
		assert.Equal(t, "", providerID)
		assert.Equal(t, "", graphqlQuery)
		assert.Nil(t, bodyData)
	})
}

func TestAuditService_SendAuditLog(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/api/logs", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(models.Log{ID: "log-123", Status: "success"})
		}))
		defer server.Close()

		service := NewAuditService(server.URL)
		auditReq := &models.AuditLogRequest{
			EventID:           uuid.New(),
			ConsumerID:        "consumer-123",
			ProviderID:        "provider-456",
			TransactionStatus: "SUCCESS",
			RequestedData:     json.RawMessage(`{"query": "test"}`),
		}

		err := service.SendAuditLog(context.Background(), auditReq)
		assert.NoError(t, err)
	})

	t.Run("HTTPError", func(t *testing.T) {
		service := NewAuditService("http://invalid-url:9999")
		auditReq := &models.AuditLogRequest{
			EventID:           uuid.New(),
			TransactionStatus: "SUCCESS",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err := service.SendAuditLog(ctx, auditReq)
		assert.Error(t, err)
	})

	t.Run("Non200Status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		service := NewAuditService(server.URL)
		auditReq := &models.AuditLogRequest{
			EventID:           uuid.New(),
			TransactionStatus: "SUCCESS",
		}

		err := service.SendAuditLog(context.Background(), auditReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "status 500")
	})
}

func TestAuditService_SendAuditLogAsync(t *testing.T) {
	auditReceived := make(chan bool, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.Log{ID: "log-123", Status: "success"})
		auditReceived <- true
	}))
	defer server.Close()

	service := NewAuditService(server.URL)
	auditReq := &models.AuditLogRequest{
		EventID:           uuid.New(),
		TransactionStatus: "SUCCESS",
	}

	// Should not block
	service.SendAuditLogAsync(auditReq)

	// Wait for async audit log to be received
	select {
	case <-auditReceived:
		// Audit log was sent successfully
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for audit log to be sent")
	}
}

// Helper type for testing read errors
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, assert.AnError
}
