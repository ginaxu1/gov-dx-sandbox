package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gov-dx-sandbox/audit-service/models"
	"github.com/gov-dx-sandbox/audit-service/services"
	"github.com/stretchr/testify/assert"
)

func TestCreateManagementEvent(t *testing.T) {
	// Setup
	db := services.SetupSQLiteTestDB(t)
	service := services.NewManagementEventService(db)
	handler := NewManagementEventHandler(service)

	t.Run("Success", func(t *testing.T) {
		reqBody := models.CreateManagementEventRequest{
			EventType: "CREATE",
			Status:    "success",
			Timestamp: time.Now().Format(time.RFC3339),
			Actor: models.Actor{
				Type: "USER",
				ID:   strPtr("user-123"),
				Role: strPtr("ADMIN"),
			},
			Target: models.Target{
				Resource: "MEMBERS",
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/management-events", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.CreateManagementEvent(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp models.ManagementEvent
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, reqBody.Actor.ID, resp.ActorID)
	})

	t.Run("InvalidMethod", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/management-events", nil)
		w := httptest.NewRecorder()

		handler.CreateManagementEvent(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/management-events", bytes.NewBufferString("invalid json"))
		w := httptest.NewRecorder()

		handler.CreateManagementEvent(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetManagementEvents(t *testing.T) {
	// Setup
	db := services.SetupSQLiteTestDB(t)
	service := services.NewManagementEventService(db)
	handler := NewManagementEventHandler(service)

	// Seed data
	ctx := context.Background()
	_, err := service.CreateManagementEvent(ctx, &models.CreateManagementEventRequest{
		EventType: "CREATE",
		Status:    "success",
		Timestamp: time.Now().Format(time.RFC3339),
		Actor: models.Actor{
			Type: "USER",
			ID:   strPtr("user-1"),
			Role: strPtr("ADMIN"),
		},
		Target: models.Target{
			Resource: "MEMBERS",
		},
	})
	assert.NoError(t, err)

	t.Run("GetAll", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/management-events", nil)
		w := httptest.NewRecorder()

		handler.GetManagementEvents(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp models.ManagementEventResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(resp.Events))
	})

	t.Run("FilterByActorID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/management-events?actorId=user-1", nil)
		w := httptest.NewRecorder()

		handler.GetManagementEvents(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp models.ManagementEventResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(resp.Events))
		assert.Equal(t, "user-1", *resp.Events[0].ActorID)
	})

	t.Run("InvalidMethod", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/management-events", nil)
		w := httptest.NewRecorder()

		handler.GetManagementEvents(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}
