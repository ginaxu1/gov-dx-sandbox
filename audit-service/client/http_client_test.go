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
	"github.com/stretchr/testify/require"
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

func TestHTTPClient_LogManagementEvent(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/api/events", r.URL.Path)

			var req ManagementEventRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			assert.Equal(t, "CREATE", req.EventType)
			assert.Equal(t, "USER", req.Actor.Type)
			assert.NotNil(t, req.Actor.ID)
			assert.NotNil(t, req.Actor.Role)

			w.WriteHeader(http.StatusCreated)
		}))
		defer server.Close()

		client := NewAuditClient(server.URL).(*httpClient)
		actorID := "user-123"
		actorRole := "ADMIN"
		event := ManagementEventRequest{
			EventType: "CREATE",
			Actor: Actor{
				Type: "USER",
				ID:   &actorID,
				Role: &actorRole,
			},
			Target: Target{
				Resource:   "MEMBERS",
				ResourceID: "member-456",
			},
		}

		err := client.LogManagementEvent(context.Background(), event)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
	})

	t.Run("MissingRequiredFields", func(t *testing.T) {
		client := NewAuditClient("http://localhost:3001").(*httpClient)
		event := ManagementEventRequest{
			EventType: "", // Missing
		}

		err := client.LogManagementEvent(context.Background(), event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required fields")
	})

	t.Run("MissingActorIDForUSER", func(t *testing.T) {
		client := NewAuditClient("http://localhost:3001").(*httpClient)
		actorRole := "ADMIN"
		event := ManagementEventRequest{
			EventType: "CREATE",
			Actor: Actor{
				Type: "USER",
				ID:   nil, // Missing
				Role: &actorRole,
			},
			Target: Target{
				Resource:   "MEMBERS",
				ResourceID: "member-456",
			},
		}

		err := client.LogManagementEvent(context.Background(), event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing actor.id")
	})

	t.Run("InvalidActorRoleForUSER", func(t *testing.T) {
		client := NewAuditClient("http://localhost:3001").(*httpClient)
		actorID := "user-123"
		actorRole := "INVALID"
		event := ManagementEventRequest{
			EventType: "CREATE",
			Actor: Actor{
				Type: "USER",
				ID:   &actorID,
				Role: &actorRole,
			},
			Target: Target{
				Resource:   "MEMBERS",
				ResourceID: "member-456",
			},
		}

		err := client.LogManagementEvent(context.Background(), event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid actor.role")
	})

	t.Run("SERVICEActorType", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}))
		defer server.Close()

		client := NewAuditClient(server.URL).(*httpClient)
		event := ManagementEventRequest{
			EventType: "CREATE",
			Actor: Actor{
				Type: "SERVICE",
				ID:   nil, // OK for SERVICE
				Role: nil, // OK for SERVICE
			},
			Target: Target{
				Resource:   "MEMBERS",
				ResourceID: "member-456",
			},
		}

		err := client.LogManagementEvent(context.Background(), event)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
	})

	t.Run("AutoGenerateEventID", func(t *testing.T) {
		var requestBody []byte
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusCreated)
		}))
		defer server.Close()

		client := NewAuditClient(server.URL).(*httpClient)
		actorID := "user-123"
		actorRole := "MEMBER"
		event := ManagementEventRequest{
			EventID:   "", // Empty, should be generated
			EventType: "UPDATE",
			Actor: Actor{
				Type: "USER",
				ID:   &actorID,
				Role: &actorRole,
			},
			Target: Target{
				Resource:   "SCHEMAS",
				ResourceID: "schema-789",
			},
		}

		err := client.LogManagementEvent(context.Background(), event)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		// Check that event ID was generated in the request body
		var sentEvent ManagementEventRequest
		err = json.Unmarshal(requestBody, &sentEvent)
		assert.NoError(t, err)
		assert.NotEmpty(t, sentEvent.EventID)
	})

	t.Run("AutoGenerateTimestamp", func(t *testing.T) {
		var requestBody []byte
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusCreated)
		}))
		defer server.Close()

		client := NewAuditClient(server.URL).(*httpClient)
		actorID := "user-123"
		actorRole := "ADMIN"
		event := ManagementEventRequest{
			EventType: "DELETE",
			Timestamp: nil, // Empty, should be generated
			Actor: Actor{
				Type: "USER",
				ID:   &actorID,
				Role: &actorRole,
			},
			Target: Target{
				Resource:   "APPLICATIONS",
				ResourceID: "app-012",
			},
		}

		err := client.LogManagementEvent(context.Background(), event)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		// Check that timestamp was generated in the request body
		var sentEvent ManagementEventRequest
		err = json.Unmarshal(requestBody, &sentEvent)
		assert.NoError(t, err)
		assert.NotNil(t, sentEvent.Timestamp)
		assert.NotEmpty(t, *sentEvent.Timestamp)
	})

	t.Run("HTTPError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		client := NewAuditClient(server.URL).(*httpClient)
		actorID := "user-123"
		actorRole := "ADMIN"
		event := ManagementEventRequest{
			EventType: "CREATE",
			Actor: Actor{
				Type: "USER",
				ID:   &actorID,
				Role: &actorRole,
			},
			Target: Target{
				Resource:   "MEMBERS",
				ResourceID: "member-456",
			},
		}

		err := client.LogManagementEvent(context.Background(), event)
		assert.NoError(t, err) // Should not return error (async)

		time.Sleep(100 * time.Millisecond)
	})
}

func TestNewHTTPClient(t *testing.T) {
	client := newHTTPClient()
	assert.NotNil(t, client)
	assert.Equal(t, 5*time.Second, client.Timeout)
}
