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

func TestNewManagementEventService(t *testing.T) {
	db := setupTestDB(t)
	service := NewManagementEventService(db)
	assert.NotNil(t, service)
	assert.Equal(t, db, service.db)
}

func TestManagementEventService_CreateManagementEvent(t *testing.T) {
	db := setupTestDB(t)
	service := NewManagementEventService(db)

	t.Run("Create valid event", func(t *testing.T) {
		req := &models.CreateManagementEventRequest{
			Timestamp:     time.Now().UTC().Format(time.RFC3339),
			EventType:     "user_created",
			ActorID:       "actor-123",
			ActorType:     "user",
			TargetID:      "target-123",
			TargetType:    "application",
			Action:        "create",
			Details:       map[string]interface{}{"key": "value"},
		}

		resp, err := service.CreateManagementEvent(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "user_created", resp.EventType)
		assert.Equal(t, "actor-123", resp.ActorID)
	})

	t.Run("Invalid timestamp format", func(t *testing.T) {
		req := &models.CreateManagementEventRequest{
			Timestamp: "invalid-timestamp",
			EventType: "user_created",
		}

		_, err := service.CreateManagementEvent(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid timestamp format")
	})
}

func TestManagementEventService_GetManagementEvents(t *testing.T) {
	db := setupTestDB(t)
	service := NewManagementEventService(db)

	// Create test events
	now := time.Now().UTC()
	for i := 0; i < 3; i++ {
		req := &models.CreateManagementEventRequest{
			Timestamp:  now.Add(time.Duration(i) * time.Minute).Format(time.RFC3339),
			EventType:  "user_created",
			ActorID:    "actor-123",
			ActorType:  "user",
			TargetID:   "target-123",
			TargetType: "application",
			Action:     "create",
		}
		_, err := service.CreateManagementEvent(context.Background(), req)
		require.NoError(t, err)
	}

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
		eventType := "user_created"
		filter := &models.ManagementEventFilter{
			EventType: &eventType,
			Limit:     10,
		}

		resp, err := service.GetManagementEvents(context.Background(), filter)
		require.NoError(t, err)
		for _, event := range resp.Events {
			assert.Equal(t, "user_created", event.EventType)
		}
	})

	t.Run("Filter by actor ID", func(t *testing.T) {
		actorID := "actor-123"
		filter := &models.ManagementEventFilter{
			ActorID: &actorID,
			Limit:   10,
		}

		resp, err := service.GetManagementEvents(context.Background(), filter)
		require.NoError(t, err)
		for _, event := range resp.Events {
			assert.Equal(t, "actor-123", event.ActorID)
		}
	})
}
