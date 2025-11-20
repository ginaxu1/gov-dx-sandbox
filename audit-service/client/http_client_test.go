package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAuditClient(t *testing.T) {
	t.Run("WithURL", func(t *testing.T) {
		client := NewAuditClient("http://localhost:3001")
		assert.NotNil(t, client)
	})

	t.Run("WithEmptyURL", func(t *testing.T) {
		client := NewAuditClient("")
		assert.Nil(t, client)
	})
}

func TestHTTPClient_LogDataExchange(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/v1/audit/exchange", r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewAuditClient(server.URL).(*httpClient)
		event := DataExchangeEvent{
			ConsumerAppID:    "app-123",
			ProviderSchemaID: "schema-456",
			ConsumerID:       "consumer-789",
			ProviderID:       "provider-012",
			Status:           "SUCCESS",
		}

		err := client.LogDataExchange(context.Background(), event)
		assert.NoError(t, err)

		// Wait for async goroutine
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("MissingRequiredFields", func(t *testing.T) {
		client := NewAuditClient("http://localhost:3001").(*httpClient)
		event := DataExchangeEvent{
			ConsumerAppID: "app-123",
			// Missing other required fields
		}

		err := client.LogDataExchange(context.Background(), event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required fields")
	})

	t.Run("AutoGenerateEventID", func(t *testing.T) {
		var requestBody []byte
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewAuditClient(server.URL).(*httpClient)
		event := DataExchangeEvent{
			ConsumerAppID:    "app-123",
			ProviderSchemaID: "schema-456",
			ConsumerID:       "consumer-789",
			ProviderID:       "provider-012",
			Status:           "SUCCESS",
			EventID:          "", // Empty, should be generated
		}

		err := client.LogDataExchange(context.Background(), event)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		// Check that event ID was generated in the request body
		var sentEvent DataExchangeEvent
		err = json.Unmarshal(requestBody, &sentEvent)
		assert.NoError(t, err)
		assert.NotEmpty(t, sentEvent.EventID)
	})

	t.Run("AutoGenerateTimestamp", func(t *testing.T) {
		var requestBody []byte
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewAuditClient(server.URL).(*httpClient)
		event := DataExchangeEvent{
			ConsumerAppID:    "app-123",
			ProviderSchemaID: "schema-456",
			ConsumerID:       "consumer-789",
			ProviderID:       "provider-012",
			Status:           "SUCCESS",
			Timestamp:        "", // Empty, should be generated
		}

		err := client.LogDataExchange(context.Background(), event)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		// Check that timestamp was generated in the request body
		var sentEvent DataExchangeEvent
		err = json.Unmarshal(requestBody, &sentEvent)
		assert.NoError(t, err)
		assert.NotEmpty(t, sentEvent.Timestamp)
	})

	t.Run("HTTPError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewAuditClient(server.URL).(*httpClient)
		event := DataExchangeEvent{
			ConsumerAppID:    "app-123",
			ProviderSchemaID: "schema-456",
			ConsumerID:       "consumer-789",
			ProviderID:       "provider-012",
			Status:           "SUCCESS",
		}

		err := client.LogDataExchange(context.Background(), event)
		assert.NoError(t, err) // Should not return error (async)

		time.Sleep(100 * time.Millisecond)
	})
}

func TestNewHTTPClient(t *testing.T) {
	client := newHTTPClient()
	assert.NotNil(t, client)
	assert.Equal(t, 5*time.Second, client.Timeout)
}
