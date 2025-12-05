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

func TestCreateDataExchangeEvent(t *testing.T) {
	// Setup
	db := services.SetupSQLiteTestDB(t)
	service := services.NewDataExchangeEventService(db)
	handler := NewDataExchangeEventHandler(service)

	t.Run("Success", func(t *testing.T) {
		reqBody := models.CreateDataExchangeEventRequest{
			Timestamp:     time.Now().Format(time.RFC3339),
			Status:        "success",
			ApplicationID: "app-123",
			SchemaID:      "schema-123",
			RequestedData: json.RawMessage(`"some data"`),
			ConsumerID:    strPtr("consumer-123"),
			ProviderID:    strPtr("provider-123"),
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/data-exchange-events", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.CreateDataExchangeEvent(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp models.DataExchangeEventResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, reqBody.ApplicationID, resp.ApplicationID)
	})

	t.Run("InvalidMethod", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/data-exchange-events", nil)
		w := httptest.NewRecorder()

		handler.CreateDataExchangeEvent(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/data-exchange-events", bytes.NewBufferString("invalid json"))
		w := httptest.NewRecorder()

		handler.CreateDataExchangeEvent(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetDataExchangeEvents(t *testing.T) {
	// Setup
	db := services.SetupSQLiteTestDB(t)
	service := services.NewDataExchangeEventService(db)
	handler := NewDataExchangeEventHandler(service)

	// Seed data
	ctx := context.Background()
	_, err := service.CreateDataExchangeEvent(ctx, &models.CreateDataExchangeEventRequest{
		Timestamp:     time.Now().Format(time.RFC3339),
		Status:        "success",
		ApplicationID: "app-1",
		SchemaID:      "schema-1",
		RequestedData: json.RawMessage(`"data1"`),
	})
	assert.NoError(t, err)

	t.Run("GetAll", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/data-exchange-events", nil)
		w := httptest.NewRecorder()

		handler.GetDataExchangeEvents(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp models.DataExchangeEventListResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(resp.Events))
	})

	t.Run("FilterByApplicationID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/data-exchange-events?applicationId=app-1", nil)
		w := httptest.NewRecorder()

		handler.GetDataExchangeEvents(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp models.DataExchangeEventListResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(resp.Events))
		assert.Equal(t, "app-1", resp.Events[0].ApplicationID)
	})

	t.Run("FilterByApplicationID_NoMatch", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/data-exchange-events?applicationId=non-existent", nil)
		w := httptest.NewRecorder()

		handler.GetDataExchangeEvents(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp models.DataExchangeEventListResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(resp.Events))
	})

	t.Run("InvalidMethod", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/data-exchange-events", nil)
		w := httptest.NewRecorder()

		handler.GetDataExchangeEvents(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}
