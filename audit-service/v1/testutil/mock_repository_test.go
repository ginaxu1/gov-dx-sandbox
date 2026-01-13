package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gov-dx-sandbox/audit-service/v1/database"
	v1models "github.com/gov-dx-sandbox/audit-service/v1/models"
)

func TestMockRepository_GetAuditLogsByTraceID(t *testing.T) {
	mockRepo := NewMockRepository()
	ctx := context.Background()

	// Create test trace IDs
	traceID1 := uuid.New()
	traceID2 := uuid.New()

	// Create audit logs with different trace IDs
	eventType1 := "POLICY_CHECK"
	eventType2 := "CONSENT_CHECK"
	eventType3 := "PROVIDER_FETCH"

	// Logs for traceID1 (should be returned in chronological order)
	log1 := &v1models.AuditLog{
		ID:         uuid.New(),
		Timestamp:  time.Now().Add(-2 * time.Hour),
		TraceID:    &traceID1,
		Status:     v1models.StatusSuccess,
		EventType:  &eventType1,
		ActorType:  "SERVICE",
		ActorID:    "policy-decision-point",
		TargetType: "SERVICE",
	}

	log2 := &v1models.AuditLog{
		ID:         uuid.New(),
		Timestamp:  time.Now().Add(-1 * time.Hour),
		TraceID:    &traceID1,
		Status:     v1models.StatusSuccess,
		EventType:  &eventType2,
		ActorType:  "SERVICE",
		ActorID:    "consent-engine",
		TargetType: "SERVICE",
	}

	log3 := &v1models.AuditLog{
		ID:         uuid.New(),
		Timestamp:  time.Now(),
		TraceID:    &traceID1,
		Status:     v1models.StatusSuccess,
		EventType:  &eventType3,
		ActorType:  "SERVICE",
		ActorID:    "orchestration-engine",
		TargetType: "SERVICE",
	}

	// Log for traceID2 (should not be returned)
	log4 := &v1models.AuditLog{
		ID:         uuid.New(),
		Timestamp:  time.Now(),
		TraceID:    &traceID2,
		Status:     v1models.StatusSuccess,
		EventType:  &eventType1,
		ActorType:  "SERVICE",
		ActorID:    "policy-decision-point",
		TargetType: "SERVICE",
	}

	// Create logs in the repository
	_, err := mockRepo.CreateAuditLog(ctx, log1)
	require.NoError(t, err)
	_, err = mockRepo.CreateAuditLog(ctx, log2)
	require.NoError(t, err)
	_, err = mockRepo.CreateAuditLog(ctx, log3)
	require.NoError(t, err)
	_, err = mockRepo.CreateAuditLog(ctx, log4)
	require.NoError(t, err)

	// Test: Get logs by traceID1
	logs, err := mockRepo.GetAuditLogsByTraceID(ctx, traceID1.String())
	require.NoError(t, err)
	assert.Len(t, logs, 3, "Should return 3 logs for traceID1")

	// Verify chronological ordering (ASC)
	assert.Equal(t, log1.ID, logs[0].ID, "First log should be the oldest")
	assert.Equal(t, log2.ID, logs[1].ID, "Second log should be the middle one")
	assert.Equal(t, log3.ID, logs[2].ID, "Third log should be the newest")

	// Verify all logs have the correct traceID
	for _, log := range logs {
		assert.NotNil(t, log.TraceID)
		assert.Equal(t, traceID1, *log.TraceID)
	}

	// Test: Get logs by traceID2
	logs2, err := mockRepo.GetAuditLogsByTraceID(ctx, traceID2.String())
	require.NoError(t, err)
	assert.Len(t, logs2, 1, "Should return 1 log for traceID2")
	assert.Equal(t, log4.ID, logs2[0].ID)

	// Test: Get logs by non-existent traceID
	logs3, err := mockRepo.GetAuditLogsByTraceID(ctx, uuid.New().String())
	require.NoError(t, err)
	assert.Len(t, logs3, 0, "Should return empty slice for non-existent traceID")
	assert.NotNil(t, logs3, "Should return empty slice, not nil")
}

func TestMockRepository_GetAuditLogs_Filtering(t *testing.T) {
	mockRepo := NewMockRepository()
	ctx := context.Background()

	// Create test data
	traceID1 := uuid.New()
	traceID2 := uuid.New()
	eventType1 := "POLICY_CHECK"
	eventType2 := "CONSENT_CHECK"
	eventAction1 := "READ"
	eventAction2 := "CREATE"

	// Create various audit logs
	logs := []*v1models.AuditLog{
		{
			ID:          uuid.New(),
			Timestamp:   time.Now().Add(-3 * time.Hour),
			TraceID:     &traceID1,
			Status:      v1models.StatusSuccess,
			EventType:   &eventType1,
			EventAction: &eventAction1,
			ActorType:   "SERVICE",
			ActorID:     "policy-decision-point",
			TargetType:  "SERVICE",
		},
		{
			ID:          uuid.New(),
			Timestamp:   time.Now().Add(-2 * time.Hour),
			TraceID:     &traceID1,
			Status:      v1models.StatusSuccess,
			EventType:   &eventType2,
			EventAction: &eventAction1,
			ActorType:   "SERVICE",
			ActorID:     "consent-engine",
			TargetType:  "SERVICE",
		},
		{
			ID:          uuid.New(),
			Timestamp:   time.Now().Add(-1 * time.Hour),
			TraceID:     &traceID2,
			Status:      v1models.StatusFailure,
			EventType:   &eventType1,
			EventAction: &eventAction2,
			ActorType:   "SERVICE",
			ActorID:     "policy-decision-point",
			TargetType:  "SERVICE",
		},
		{
			ID:          uuid.New(),
			Timestamp:   time.Now(),
			TraceID:     &traceID2,
			Status:      v1models.StatusSuccess,
			EventType:   &eventType2,
			EventAction: &eventAction1,
			ActorType:   "SERVICE",
			ActorID:     "consent-engine",
			TargetType:  "SERVICE",
		},
	}

	// Seed the repository
	for _, log := range logs {
		_, err := mockRepo.CreateAuditLog(ctx, log)
		require.NoError(t, err)
	}

	// Test: Filter by TraceID
	t.Run("FilterByTraceID", func(t *testing.T) {
		filters := &database.AuditLogFilters{
			TraceID: stringPtr(traceID1.String()),
		}
		result, total, err := mockRepo.GetAuditLogs(ctx, filters)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total, "Should find 2 logs with traceID1")
		assert.Len(t, result, 2, "Should return 2 logs")
		for _, log := range result {
			assert.NotNil(t, log.TraceID)
			assert.Equal(t, traceID1, *log.TraceID)
		}
	})

	// Test: Filter by EventType
	t.Run("FilterByEventType", func(t *testing.T) {
		filters := &database.AuditLogFilters{
			EventType: &eventType1,
		}
		result, total, err := mockRepo.GetAuditLogs(ctx, filters)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total, "Should find 2 logs with eventType1")
		assert.Len(t, result, 2, "Should return 2 logs")
		for _, log := range result {
			assert.NotNil(t, log.EventType)
			assert.Equal(t, eventType1, *log.EventType)
		}
	})

	// Test: Filter by Status
	t.Run("FilterByStatus", func(t *testing.T) {
		filters := &database.AuditLogFilters{
			Status: stringPtr(v1models.StatusFailure),
		}
		result, total, err := mockRepo.GetAuditLogs(ctx, filters)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total, "Should find 1 log with FAILURE status")
		assert.Len(t, result, 1, "Should return 1 log")
		assert.Equal(t, v1models.StatusFailure, result[0].Status)
	})

	// Test: Filter by EventAction
	t.Run("FilterByEventAction", func(t *testing.T) {
		filters := &database.AuditLogFilters{
			EventAction: &eventAction1,
		}
		result, total, err := mockRepo.GetAuditLogs(ctx, filters)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total, "Should find 3 logs with eventAction1")
		assert.Len(t, result, 3, "Should return 3 logs")
		for _, log := range result {
			assert.NotNil(t, log.EventAction)
			assert.Equal(t, eventAction1, *log.EventAction)
		}
	})

	// Test: Multiple filters combined
	t.Run("MultipleFilters", func(t *testing.T) {
		filters := &database.AuditLogFilters{
			TraceID:   stringPtr(traceID1.String()),
			EventType: &eventType1,
		}
		result, total, err := mockRepo.GetAuditLogs(ctx, filters)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total, "Should find 1 log matching both filters")
		assert.Len(t, result, 1, "Should return 1 log")
		assert.Equal(t, traceID1, *result[0].TraceID)
		assert.Equal(t, eventType1, *result[0].EventType)
	})

	// Test: No filters (should return all logs)
	t.Run("NoFilters", func(t *testing.T) {
		filters := &database.AuditLogFilters{}
		result, total, err := mockRepo.GetAuditLogs(ctx, filters)
		require.NoError(t, err)
		assert.Equal(t, int64(4), total, "Should find all 4 logs")
		assert.Len(t, result, 4, "Should return all 4 logs")
	})
}

func TestMockRepository_GetAuditLogs_Pagination(t *testing.T) {
	mockRepo := NewMockRepository()
	ctx := context.Background()

	// Create 10 audit logs
	traceID := uuid.New()
	eventType := "POLICY_CHECK"
	for i := 0; i < 10; i++ {
		log := &v1models.AuditLog{
			ID:         uuid.New(),
			Timestamp:  time.Now().Add(time.Duration(i) * time.Minute),
			TraceID:    &traceID,
			Status:     v1models.StatusSuccess,
			EventType:  &eventType,
			ActorType:  "SERVICE",
			ActorID:    "test-service",
			TargetType: "SERVICE",
		}
		_, err := mockRepo.CreateAuditLog(ctx, log)
		require.NoError(t, err)
	}

	// Test: Pagination with limit
	t.Run("PaginationWithLimit", func(t *testing.T) {
		filters := &database.AuditLogFilters{
			Limit: 5,
		}
		result, total, err := mockRepo.GetAuditLogs(ctx, filters)
		require.NoError(t, err)
		assert.Equal(t, int64(10), total, "Total should be 10")
		assert.Len(t, result, 5, "Should return 5 logs (limit)")
	})

	// Test: Pagination with offset
	t.Run("PaginationWithOffset", func(t *testing.T) {
		filters := &database.AuditLogFilters{
			Limit:  5,
			Offset: 5,
		}
		result, total, err := mockRepo.GetAuditLogs(ctx, filters)
		require.NoError(t, err)
		assert.Equal(t, int64(10), total, "Total should be 10")
		assert.Len(t, result, 5, "Should return 5 logs (offset 5, limit 5)")
	})

	// Test: Pagination beyond available logs
	t.Run("PaginationBeyondAvailable", func(t *testing.T) {
		filters := &database.AuditLogFilters{
			Limit:  5,
			Offset: 10,
		}
		result, total, err := mockRepo.GetAuditLogs(ctx, filters)
		require.NoError(t, err)
		assert.Equal(t, int64(10), total, "Total should be 10")
		assert.Len(t, result, 0, "Should return empty slice when offset exceeds available logs")
		assert.NotNil(t, result, "Should return empty slice, not nil")
	})

	// Test: Default limit
	t.Run("DefaultLimit", func(t *testing.T) {
		filters := &database.AuditLogFilters{
			Limit: 0, // Should default to 100
		}
		result, total, err := mockRepo.GetAuditLogs(ctx, filters)
		require.NoError(t, err)
		assert.Equal(t, int64(10), total, "Total should be 10")
		assert.Len(t, result, 10, "Should return all 10 logs (default limit is 100, but we only have 10)")
	})

	// Test: Max limit
	t.Run("MaxLimit", func(t *testing.T) {
		filters := &database.AuditLogFilters{
			Limit: 2000, // Should be capped at 1000
		}
		result, total, err := mockRepo.GetAuditLogs(ctx, filters)
		require.NoError(t, err)
		assert.Equal(t, int64(10), total, "Total should be 10")
		assert.Len(t, result, 10, "Should return all 10 logs (limit capped at 1000, but we only have 10)")
	})
}

func TestMockRepository_GetAuditLogs_Ordering(t *testing.T) {
	mockRepo := NewMockRepository()
	ctx := context.Background()

	// Create logs with different timestamps
	traceID := uuid.New()
	eventType := "POLICY_CHECK"
	now := time.Now()

	logs := []*v1models.AuditLog{
		{
			ID:         uuid.New(),
			Timestamp:  now.Add(-2 * time.Hour),
			TraceID:    &traceID,
			Status:     v1models.StatusSuccess,
			EventType:  &eventType,
			ActorType:  "SERVICE",
			ActorID:    "test-service",
			TargetType: "SERVICE",
		},
		{
			ID:         uuid.New(),
			Timestamp:  now,
			TraceID:    &traceID,
			Status:     v1models.StatusSuccess,
			EventType:  &eventType,
			ActorType:  "SERVICE",
			ActorID:    "test-service",
			TargetType: "SERVICE",
		},
		{
			ID:         uuid.New(),
			Timestamp:  now.Add(-1 * time.Hour),
			TraceID:    &traceID,
			Status:     v1models.StatusSuccess,
			EventType:  &eventType,
			ActorType:  "SERVICE",
			ActorID:    "test-service",
			TargetType: "SERVICE",
		},
	}

	// Seed in non-chronological order
	for _, log := range logs {
		_, err := mockRepo.CreateAuditLog(ctx, log)
		require.NoError(t, err)
	}

	// Test: Results should be ordered DESC (newest first)
	filters := &database.AuditLogFilters{}
	result, total, err := mockRepo.GetAuditLogs(ctx, filters)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)

	// Verify DESC ordering (newest first)
	assert.Equal(t, logs[1].ID, result[0].ID, "First result should be the newest")
	assert.Equal(t, logs[2].ID, result[1].ID, "Second result should be the middle one")
	assert.Equal(t, logs[0].ID, result[2].ID, "Third result should be the oldest")
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
