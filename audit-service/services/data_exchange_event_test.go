package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/gov-dx-sandbox/audit-service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDataExchangeEventService(t *testing.T) {
	db := SetupSQLiteTestDB(t)
	service := NewDataExchangeEventService(db)
	assert.NotNil(t, service)
	assert.Equal(t, db, service.db)
}

func TestDataExchangeEventService_CreateDataExchangeEvent(t *testing.T) {
	db := SetupSQLiteTestDB(t)
	service := NewDataExchangeEventService(db)

	t.Run("Create valid event", func(t *testing.T) {
		req := CreateTestDataExchangeEventRequest(t, func(r *models.CreateDataExchangeEventRequest) {
			r.RequestedData = json.RawMessage(`["person.fullName"]`)
		})

		resp, err := service.CreateDataExchangeEvent(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "success", resp.Status)
		assert.Equal(t, "app-123", resp.ApplicationID)
	})

	t.Run("Invalid timestamp format", func(t *testing.T) {
		req := CreateTestDataExchangeEventRequest(t, func(r *models.CreateDataExchangeEventRequest) {
			r.Timestamp = "invalid-timestamp"
		})

		_, err := service.CreateDataExchangeEvent(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid timestamp format")
	})

	t.Run("Invalid status", func(t *testing.T) {
		req := CreateTestDataExchangeEventRequest(t, func(r *models.CreateDataExchangeEventRequest) {
			r.Status = "invalid-status"
		})

		_, err := service.CreateDataExchangeEvent(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status")
	})

	t.Run("Failure status", func(t *testing.T) {
		req := CreateTestDataExchangeEventRequest(t, func(r *models.CreateDataExchangeEventRequest) {
			r.Status = "failure"
			r.ApplicationID = "app-456"
			r.SchemaID = "schema-456"
		})

		resp, err := service.CreateDataExchangeEvent(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "failure", resp.Status)
	})
}

func TestDataExchangeEventService_GetDataExchangeEvents(t *testing.T) {
	db := SetupSQLiteTestDB(t)
	service := NewDataExchangeEventService(db)

	// Create test events
	now := time.Now().UTC()
	CreateTestEvents(t, service, 5, now)

	t.Run("Get all events", func(t *testing.T) {
		filter := &models.DataExchangeEventFilter{
			Limit: 10,
		}

		resp, err := service.GetDataExchangeEvents(context.Background(), filter)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.GreaterOrEqual(t, resp.Total, int64(5))
		assert.GreaterOrEqual(t, len(resp.Events), 5)
	})

	t.Run("Filter by status", func(t *testing.T) {
		status := "success"
		filter := &models.DataExchangeEventFilter{
			Status: &status,
			Limit:  10,
		}

		resp, err := service.GetDataExchangeEvents(context.Background(), filter)
		require.NoError(t, err)
		for _, event := range resp.Events {
			assert.Equal(t, "success", event.Status)
		}
	})

	t.Run("Filter by application ID", func(t *testing.T) {
		appID := TestAppID
		filter := &models.DataExchangeEventFilter{
			ApplicationID: &appID,
			Limit:         10,
		}

		resp, err := service.GetDataExchangeEvents(context.Background(), filter)
		require.NoError(t, err)
		for _, event := range resp.Events {
			assert.Equal(t, "app-123", event.ApplicationID)
		}
	})

	t.Run("Pagination", func(t *testing.T) {
		filter := &models.DataExchangeEventFilter{
			Limit:  2,
			Offset: 0,
		}

		resp, err := service.GetDataExchangeEvents(context.Background(), filter)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(resp.Events), 2)
	})

	t.Run("Date range filter", func(t *testing.T) {
		startDate := now.Add(-1 * time.Hour)
		endDate := now.Add(1 * time.Hour)
		filter := &models.DataExchangeEventFilter{
			StartDate: &startDate,
			EndDate:   &endDate,
			Limit:     10,
		}

		resp, err := service.GetDataExchangeEvents(context.Background(), filter)
		require.NoError(t, err)
		for _, event := range resp.Events {
			// Parse timestamp string to time.Time for comparison
			eventTime, err := time.Parse(time.RFC3339, event.Timestamp)
			require.NoError(t, err)
			assert.True(t, eventTime.After(startDate) || eventTime.Equal(startDate))
			assert.True(t, eventTime.Before(endDate) || eventTime.Equal(endDate))
		}
	})
}
