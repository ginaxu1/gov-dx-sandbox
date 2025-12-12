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

func TestCreateAuditLog(t *testing.T) {
	db := services.SetupSQLiteTestDB(t)
	service := services.NewAuditService(db)
	handler := NewAuditHandler(service)

	t.Run("Success", func(t *testing.T) {
		reqBody := models.CreateAuditLogRequest{
			TraceID:       "trace-123",
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
		assert.Equal(t, reqBody.TraceID, resp.TraceID)
	})

	t.Run("MissingRequiredFields", func(t *testing.T) {
		reqBody := models.CreateAuditLogRequest{
			TraceID: "trace-123",
			// Missing SourceService, EventType, Status
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
	service := services.NewAuditService(db)
	handler := NewAuditHandler(service)

	// Seed data
	_, _ = service.CreateAuditLog(context.Background(), &models.AuditLog{
		TraceID:       "trace-test",
		Timestamp:     time.Now(),
		SourceService: "oe",
		EventType:     "TEST",
		Status:        "SUCCESS",
	})

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/audit-logs?traceId=trace-test", nil)
		w := httptest.NewRecorder()

		handler.GetAuditLogs(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp []models.AuditLog
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, "trace-test", resp[0].TraceID)
	})

	t.Run("MissingTraceID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/audit-logs", nil)
		w := httptest.NewRecorder()

		handler.GetAuditLogs(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
