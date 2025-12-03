package services

import (
	"context"
	"testing"
	"time"

	"github.com/gov-dx-sandbox/audit-service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManagementEventService(t *testing.T) {
	db := SetupSQLiteTestDB(t)
	service := NewManagementEventService(db)
	assert.NotNil(t, service)
	assert.Equal(t, db, service.db)
}

func TestManagementEventService_CreateManagementEvent(t *testing.T) {
	db := SetupSQLiteTestDB(t)
	service := NewManagementEventService(db)

	t.Run("Create valid event", func(t *testing.T) {
		req := CreateTestManagementEventRequest(t)

		resp, err := service.CreateManagementEvent(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "CREATE", resp.EventType)
		assert.Equal(t, "actor-123", *resp.ActorID)
	})

	t.Run("Invalid timestamp format", func(t *testing.T) {
		req := CreateTestManagementEventRequest(t, func(r *models.CreateManagementEventRequest) {
			r.Timestamp = "invalid-timestamp"
		})

		_, err := service.CreateManagementEvent(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid timestamp format")
	})
}

func TestManagementEventService_GetManagementEvents(t *testing.T) {
	db := SetupSQLiteTestDB(t)
	service := NewManagementEventService(db)

	// Create test events
	now := time.Now().UTC()
	CreateTestManagementEvents(t, service, 3, now)

	t.Run("Get all events", func(t *testing.T) {
		filter := &models.ManagementEventFilter{
			Limit: 10,
		}

		resp, err := service.GetManagementEvents(context.Background(), filter)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.GreaterOrEqual(t, resp.Total, int64(3))
	})

	t.Run("Filter by event type", func(t *testing.T) {
		eventType := "CREATE"
		filter := &models.ManagementEventFilter{
			EventType: &eventType,
			Limit:     10,
		}

		resp, err := service.GetManagementEvents(context.Background(), filter)
		require.NoError(t, err)
		for _, event := range resp.Events {
			assert.Equal(t, "CREATE", event.EventType)
		}
	})

	t.Run("Filter by actor ID", func(t *testing.T) {
		actorID := TestActorID
		filter := &models.ManagementEventFilter{
			ActorID: &actorID,
			Limit:   10,
		}

		resp, err := service.GetManagementEvents(context.Background(), filter)
		require.NoError(t, err)
		for _, event := range resp.Events {
			assert.Equal(t, TestActorID, *event.ActorID)
		}
	})
}
