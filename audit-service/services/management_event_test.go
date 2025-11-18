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
	db := SetupPostgresTestDB(t)
	if db == nil {
		return // test was skipped
	}
	service := NewManagementEventService(db)
	assert.NotNil(t, service)
}

func TestManagementEventService_CreateManagementEvent(t *testing.T) {
	db := SetupPostgresTestDB(t)
	if db == nil {
		return // test was skipped
	}
	service := NewManagementEventService(db)

	t.Run("Success", func(t *testing.T) {
		actorID := "user-123"
		actorRole := "ADMIN"
		req := &models.ManagementEventRequest{
			EventType: "CREATE",
			Actor: models.Actor{
				Type: "USER",
				ID:   &actorID,
				Role: &actorRole,
			},
			Target: models.Target{
				Resource:   "MEMBERS",
				ResourceID: "member-456",
			},
		}

		event, err := service.CreateManagementEvent(context.Background(), req)
		require.NoError(t, err)
		assert.NotEmpty(t, event.ID)
		assert.Equal(t, "CREATE", event.EventType)
		assert.Equal(t, "USER", event.ActorType)
		assert.Equal(t, &actorID, event.ActorID)
		assert.Equal(t, &actorRole, event.ActorRole)
		assert.Equal(t, "MEMBERS", event.TargetResource)
		assert.Equal(t, "member-456", event.TargetResourceID)
	})

	t.Run("AutoGenerateEventID", func(t *testing.T) {
		actorID := "user-123"
		actorRole := "MEMBER"
		req := &models.ManagementEventRequest{
			EventID:   "", // Empty, should be generated
			EventType: "UPDATE",
			Actor: models.Actor{
				Type: "USER",
				ID:   &actorID,
				Role: &actorRole,
			},
			Target: models.Target{
				Resource:   "SCHEMAS",
				ResourceID: "schema-789",
			},
		}

		event, err := service.CreateManagementEvent(context.Background(), req)
		require.NoError(t, err)
		assert.NotEmpty(t, event.EventID)
	})

	t.Run("CustomTimestamp", func(t *testing.T) {
		actorID := "user-123"
		actorRole := "ADMIN"
		timestamp := "2024-01-01T00:00:00Z"
		req := &models.ManagementEventRequest{
			EventType: "DELETE",
			Timestamp: &timestamp,
			Actor: models.Actor{
				Type: "USER",
				ID:   &actorID,
				Role: &actorRole,
			},
			Target: models.Target{
				Resource:   "APPLICATIONS",
				ResourceID: "app-012",
			},
		}

		event, err := service.CreateManagementEvent(context.Background(), req)
		require.NoError(t, err)
		assert.NotZero(t, event.Timestamp)
	})

	t.Run("InvalidEventType", func(t *testing.T) {
		actorID := "user-123"
		actorRole := "ADMIN"
		req := &models.ManagementEventRequest{
			EventType: "INVALID",
			Actor: models.Actor{
				Type: "USER",
				ID:   &actorID,
				Role: &actorRole,
			},
			Target: models.Target{
				Resource:   "MEMBERS",
				ResourceID: "member-456",
			},
		}

		_, err := service.CreateManagementEvent(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid event type")
	})

	t.Run("InvalidActorType", func(t *testing.T) {
		actorID := "user-123"
		actorRole := "ADMIN"
		req := &models.ManagementEventRequest{
			EventType: "CREATE",
			Actor: models.Actor{
				Type: "INVALID",
				ID:   &actorID,
				Role: &actorRole,
			},
			Target: models.Target{
				Resource:   "MEMBERS",
				ResourceID: "member-456",
			},
		}

		_, err := service.CreateManagementEvent(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid actor type")
	})

	t.Run("MissingActorRoleForUSER", func(t *testing.T) {
		actorID := "user-123"
		req := &models.ManagementEventRequest{
			EventType: "CREATE",
			Actor: models.Actor{
				Type: "USER",
				ID:   &actorID,
				Role: nil, // Missing
			},
			Target: models.Target{
				Resource:   "MEMBERS",
				ResourceID: "member-456",
			},
		}

		_, err := service.CreateManagementEvent(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "actor role is required")
	})

	t.Run("InvalidActorRole", func(t *testing.T) {
		actorID := "user-123"
		invalidRole := "INVALID"
		req := &models.ManagementEventRequest{
			EventType: "CREATE",
			Actor: models.Actor{
				Type: "USER",
				ID:   &actorID,
				Role: &invalidRole,
			},
			Target: models.Target{
				Resource:   "MEMBERS",
				ResourceID: "member-456",
			},
		}

		_, err := service.CreateManagementEvent(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid actor role")
	})

	t.Run("SERVICEActorType", func(t *testing.T) {
		req := &models.ManagementEventRequest{
			EventType: "CREATE",
			Actor: models.Actor{
				Type: "SERVICE",
				ID:   nil, // OK for SERVICE
				Role: nil, // OK for SERVICE
			},
			Target: models.Target{
				Resource:   "MEMBERS",
				ResourceID: "member-456",
			},
		}

		event, err := service.CreateManagementEvent(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "SERVICE", event.ActorType)
		assert.Nil(t, event.ActorID)
		assert.Nil(t, event.ActorRole)
	})

	t.Run("InvalidTargetResource", func(t *testing.T) {
		actorID := "user-123"
		actorRole := "ADMIN"
		req := &models.ManagementEventRequest{
			EventType: "CREATE",
			Actor: models.Actor{
				Type: "USER",
				ID:   &actorID,
				Role: &actorRole,
			},
			Target: models.Target{
				Resource:   "INVALID",
				ResourceID: "resource-123",
			},
		}

		_, err := service.CreateManagementEvent(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid target resource")
	})

	t.Run("WithMetadata", func(t *testing.T) {
		actorID := "user-123"
		actorRole := "ADMIN"
		metadata := map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		}
		req := &models.ManagementEventRequest{
			EventType: "CREATE",
			Actor: models.Actor{
				Type: "USER",
				ID:   &actorID,
				Role: &actorRole,
			},
			Target: models.Target{
				Resource:   "MEMBERS",
				ResourceID: "member-456",
			},
			Metadata: &metadata,
		}

		event, err := service.CreateManagementEvent(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, event.Metadata)
	})
}

func TestManagementEventService_GetManagementEvents(t *testing.T) {
	db := SetupPostgresTestDB(t)
	if db == nil {
		return // test was skipped
	}
	service := NewManagementEventService(db)

	// Create test events
	actorID1 := "user-1"
	actorRole1 := "ADMIN"
	actorID2 := "user-2"
	actorRole2 := "MEMBER"

	events := []*models.ManagementEventRequest{
		{
			EventType: "CREATE",
			Actor: models.Actor{
				Type: "USER",
				ID:   &actorID1,
				Role: &actorRole1,
			},
			Target: models.Target{
				Resource:   "MEMBERS",
				ResourceID: "member-1",
			},
		},
		{
			EventType: "UPDATE",
			Actor: models.Actor{
				Type: "USER",
				ID:   &actorID2,
				Role: &actorRole2,
			},
			Target: models.Target{
				Resource:   "SCHEMAS",
				ResourceID: "schema-1",
			},
		},
		{
			EventType: "CREATE",
			Actor: models.Actor{
				Type: "USER",
				ID:   &actorID1,
				Role: &actorRole1,
			},
			Target: models.Target{
				Resource:   "APPLICATIONS",
				ResourceID: "app-1",
			},
		},
	}

	for _, req := range events {
		_, err := service.CreateManagementEvent(context.Background(), req)
		require.NoError(t, err)
	}

	t.Run("GetAllEvents", func(t *testing.T) {
		filter := &models.ManagementEventFilter{}
		response, err := service.GetManagementEvents(context.Background(), filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(response.Events), 3)
		assert.GreaterOrEqual(t, response.Total, int64(3))
	})

	t.Run("FilterByEventType", func(t *testing.T) {
		eventType := "CREATE"
		filter := &models.ManagementEventFilter{
			EventType: &eventType,
		}
		response, err := service.GetManagementEvents(context.Background(), filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(response.Events), 2)
		for _, event := range response.Events {
			assert.Equal(t, "CREATE", event.EventType)
		}
	})

	t.Run("FilterByActorID", func(t *testing.T) {
		filter := &models.ManagementEventFilter{
			ActorID: &actorID1,
		}
		response, err := service.GetManagementEvents(context.Background(), filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(response.Events), 2)
		for _, event := range response.Events {
			assert.Equal(t, &actorID1, event.ActorID)
		}
	})

	t.Run("FilterByTargetResource", func(t *testing.T) {
		targetResource := "MEMBERS"
		filter := &models.ManagementEventFilter{
			TargetResource: &targetResource,
		}
		response, err := service.GetManagementEvents(context.Background(), filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(response.Events), 1)
		for _, event := range response.Events {
			assert.Equal(t, "MEMBERS", event.TargetResource)
		}
	})

	t.Run("Pagination", func(t *testing.T) {
		filter := &models.ManagementEventFilter{
			Limit:  2,
			Offset: 0,
		}
		response, err := service.GetManagementEvents(context.Background(), filter)
		require.NoError(t, err)
		assert.Equal(t, 2, len(response.Events))
		assert.Equal(t, 2, response.Limit)
		assert.Equal(t, 0, response.Offset)
	})

	t.Run("DateRangeFilter", func(t *testing.T) {
		startDate := time.Now().Add(-24 * time.Hour)
		endDate := time.Now().Add(24 * time.Hour)
		filter := &models.ManagementEventFilter{
			StartDate: &startDate,
			EndDate:   &endDate,
		}
		response, err := service.GetManagementEvents(context.Background(), filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(response.Events), 0)
	})

	t.Run("DefaultLimit", func(t *testing.T) {
		filter := &models.ManagementEventFilter{
			Limit: 0, // Should use default
		}
		response, err := service.GetManagementEvents(context.Background(), filter)
		require.NoError(t, err)
		assert.Equal(t, 50, response.Limit)
	})
}
