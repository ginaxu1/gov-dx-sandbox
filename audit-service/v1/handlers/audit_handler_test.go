package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/audit-service/services"
	"github.com/gov-dx-sandbox/audit-service/v1/models"
	v1services "github.com/gov-dx-sandbox/audit-service/v1/services"
	"github.com/stretchr/testify/assert"
)

func TestCreateAuditLog(t *testing.T) {
	db := services.SetupSQLiteTestDB(t)
	service := v1services.NewAuditService(db)
	handler := NewAuditHandler(service)

	t.Run("Success", func(t *testing.T) {
		traceID := uuid.New().String()
		reqBody := models.CreateAuditLogRequest{
			TraceID:       traceID,
			Timestamp:     time.Now().Format(time.RFC3339),
			SourceService: "oe",
			TargetService: "pdp",
			EventType:     "CHECK",
			Status:        "SUCCESS",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/audit-logs", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.CreateAuditLog(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp models.AuditLog
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, uuid.MustParse(reqBody.TraceID), resp.TraceID)
	})

	t.Run("MissingRequiredFields", func(t *testing.T) {
		reqBody := models.CreateAuditLogRequest{
			TraceID: uuid.New().String(),
			// Missing SourceService, EventType, Status
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/audit-logs", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.CreateAuditLog(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("InvalidTraceID", func(t *testing.T) {
		reqBody := models.CreateAuditLogRequest{
			TraceID:       "invalid-uuid",
			Timestamp:     time.Now().Format(time.RFC3339),
			SourceService: "oe",
			TargetService: "pdp",
			EventType:     "CHECK",
			Status:        "SUCCESS",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/audit-logs", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.CreateAuditLog(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetAuditLogs(t *testing.T) {
	db := services.SetupSQLiteTestDB(t)
	service := v1services.NewAuditService(db)
	handler := NewAuditHandler(service)

	traceID := uuid.New()
	
	// Seed data
	_, _ = service.CreateAuditLog(context.Background(), &models.AuditLog{
		TraceID:       traceID,
		Timestamp:     time.Now(),
		SourceService: "oe",
		EventType:     "TEST",
		Status:        "SUCCESS",
	})

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/audit-logs?traceId="+traceID.String(), nil)
		w := httptest.NewRecorder()

		handler.GetAuditLogs(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp []models.AuditLog
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, traceID, resp[0].TraceID)
	})

	t.Run("MissingTraceID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/audit-logs", nil)
		w := httptest.NewRecorder()

		handler.GetAuditLogs(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("InvalidTraceID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/audit-logs?traceId=invalid", nil)
		w := httptest.NewRecorder()

		handler.GetAuditLogs(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
