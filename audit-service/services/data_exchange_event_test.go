package services

import (
	"context"
	"testing"
	"time"

	"github.com/gov-dx-sandbox/audit-service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	return SetupSQLiteTestDB(t)
}

func TestNewDataExchangeEventService(t *testing.T) {
	db := setupTestDB(t)
	service := NewDataExchangeEventService(db)
	assert.NotNil(t, service)
	assert.Equal(t, db, service.db)
}

func TestDataExchangeEventService_CreateDataExchangeEvent(t *testing.T) {
	db := setupTestDB(t)
	service := NewDataExchangeEventService(db)

	t.Run("Create valid event", func(t *testing.T) {
		req := &models.CreateDataExchangeEventRequest{
			Timestamp:         time.Now().UTC().Format(time.RFC3339),
			Status:            "success",
			ApplicationID:     "app-123",
			SchemaID:          "schema-123",
			RequestedData:     []string{"person.fullName"},
			OnBehalfOfOwnerID: "owner-123",
			ConsumerID:        "consumer-123",
			ProviderID:        "provider-123",
		}

		resp, err := service.CreateDataExchangeEvent(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "success", resp.Status)
		assert.Equal(t, "app-123", resp.ApplicationID)
	})

	t.Run("Invalid timestamp format", func(t *testing.T) {
		req := &models.CreateDataExchangeEventRequest{
			Timestamp: "invalid-timestamp",
			Status:    "success",
		}

		_, err := service.CreateDataExchangeEvent(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid timestamp format")
	})

	t.Run("Invalid status", func(t *testing.T) {
		req := &models.CreateDataExchangeEventRequest{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Status:    "invalid-status",
		}

		_, err := service.CreateDataExchangeEvent(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status")
	})

	t.Run("Failure status", func(t *testing.T) {
		req := &models.CreateDataExchangeEventRequest{
			Timestamp:     time.Now().UTC().Format(time.RFC3339),
			Status:        "failure",
			ApplicationID: "app-456",
		}

		resp, err := service.CreateDataExchangeEvent(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "failure", resp.Status)
	})
}

func TestDataExchangeEventService_GetDataExchangeEvents(t *testing.T) {
	db := setupTestDB(t)
	service := NewDataExchangeEventService(db)

	// Create test events
	now := time.Now().UTC()
	for i := 0; i < 5; i++ {
		req := &models.CreateDataExchangeEventRequest{
			Timestamp:     now.Add(time.Duration(i) * time.Minute).Format(time.RFC3339),
			Status:        "success",
			ApplicationID: "app-123",
			SchemaID:      "schema-123",
		}
		_, err := service.CreateDataExchangeEvent(context.Background(), req)
		require.NoError(t, err)
	}

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
		appID := "app-123"
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
			assert.True(t, event.Timestamp.After(startDate) || event.Timestamp.Equal(startDate))
			assert.True(t, event.Timestamp.Before(endDate) || event.Timestamp.Equal(endDate))
		}
	})
}
