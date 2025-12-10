package services

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gov-dx-sandbox/audit-service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuditService(t *testing.T) {
	db := SetupSQLiteTestDB(t)
	service := NewAuditService(db)
	assert.NotNil(t, service)
}

func TestAuditService_CreateAuditLog(t *testing.T) {
	db := SetupSQLiteTestDB(t)
	service := NewAuditService(db)

	t.Run("Create valid audit log", func(t *testing.T) {
		req := &models.AuditLog{
			TraceID:       "trace-123",
			Timestamp:     time.Now().UTC(),
			SourceService: "orchestration-engine",
			TargetService: "pdp",
			EventType:     "POLICY_CHECK_REQUEST",
			Status:        "SUCCESS",
			ActorID:       strPtr("user-123"),
			Resources:     json.RawMessage(`{"appId": "app-123"}`),
			Metadata:      json.RawMessage(`{"query": "some-query"}`),
		}

		resp, err := service.CreateAuditLog(req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "trace-123", resp.TraceID)
		assert.NotEmpty(t, resp.ID)
	})
}

func TestAuditService_GetAuditLogs(t *testing.T) {
	db := SetupSQLiteTestDB(t)
	service := NewAuditService(db)

	traceID := "trace-456"
	
	// Create multiple logs for same trace
	logs := []*models.AuditLog{
		{
			TraceID:       traceID,
			Timestamp:     time.Now().UTC().Add(-2 * time.Minute),
			SourceService: "oe",
			TargetService: "pdp",
			EventType:     "REQ_1",
			Status:        "SUCCESS",
		},
		{
			TraceID:       traceID,
			Timestamp:     time.Now().UTC().Add(-1 * time.Minute),
			SourceService: "oe",
			TargetService: "ce",
			EventType:     "REQ_2",
			Status:        "SUCCESS",
		},
		{
			TraceID:       "other-trace",
			Timestamp:     time.Now().UTC(),
			SourceService: "oe",
			EventType:     "REQ_OTHER",
			Status:        "SUCCESS",
		},
	}

	for _, l := range logs {
		_, err := service.CreateAuditLog(l)
		require.NoError(t, err)
	}

	t.Run("Get logs by trace ID", func(t *testing.T) {
		resp, err := service.GetAuditLogs(traceID)
		require.NoError(t, err)
		assert.Len(t, resp, 2)
		assert.Equal(t, "REQ_1", resp[0].EventType)
		assert.Equal(t, "REQ_2", resp[1].EventType)
	})

	t.Run("Get logs for non-existent trace", func(t *testing.T) {
		resp, err := service.GetAuditLogs("non-existent")
		require.NoError(t, err)
		assert.Empty(t, resp)
	})
}

func strPtr(s string) *string {
	return &s
}
