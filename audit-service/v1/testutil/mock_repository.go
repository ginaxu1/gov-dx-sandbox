package testutil

import (
	"context"
	"sort"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/audit-service/v1/database"
	v1models "github.com/gov-dx-sandbox/audit-service/v1/models"
)

// MockRepository is a simple mock implementation of database.AuditRepository for testing
type MockRepository struct {
	logs []*v1models.AuditLog
}

// NewMockRepository creates a new MockRepository instance
func NewMockRepository() *MockRepository {
	return &MockRepository{
		logs: make([]*v1models.AuditLog, 0),
	}
}

// CreateAuditLog simulates creating an audit log
// It automatically generates an ID if not provided (simulating BeforeCreate hook behavior)
func (m *MockRepository) CreateAuditLog(ctx context.Context, log *v1models.AuditLog) (*v1models.AuditLog, error) {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	m.logs = append(m.logs, log)
	return log, nil
}

// GetAuditLogsByTraceID retrieves all audit logs for a given trace ID
// Results are ordered by timestamp ASC (chronological order)
func (m *MockRepository) GetAuditLogsByTraceID(ctx context.Context, traceID string) ([]v1models.AuditLog, error) {
	// Parse traceID string to UUID for comparison
	traceUUID, err := uuid.Parse(traceID)
	if err != nil {
		return []v1models.AuditLog{}, nil // Return empty slice for invalid UUID
	}

	filteredLogs := []v1models.AuditLog{}
	for _, log := range m.logs {
		if log.TraceID != nil && *log.TraceID == traceUUID {
			filteredLogs = append(filteredLogs, *log)
		}
	}

	// Sort by timestamp ASC (chronological order)
	sort.Slice(filteredLogs, func(i, j int) bool {
		return filteredLogs[i].Timestamp.Before(filteredLogs[j].Timestamp)
	})

	return filteredLogs, nil
}

// GetAuditLogs retrieves audit logs with optional filtering
// Results are ordered by timestamp DESC (newest first) and paginated
func (m *MockRepository) GetAuditLogs(ctx context.Context, filters *database.AuditLogFilters) ([]v1models.AuditLog, int64, error) {
	if filters == nil {
		filters = &database.AuditLogFilters{}
	}

	// Filter logs based on provided criteria
	filteredLogs := []v1models.AuditLog{}
	for _, log := range m.logs {
		matches := true

		// Filter by TraceID
		if filters.TraceID != nil && *filters.TraceID != "" {
			traceUUID, err := uuid.Parse(*filters.TraceID)
			if err != nil {
				continue // Skip if traceID is invalid
			}
			if log.TraceID == nil || *log.TraceID != traceUUID {
				matches = false
			}
		}

		// Filter by EventType
		if matches && filters.EventType != nil && *filters.EventType != "" {
			if log.EventType == nil || *log.EventType != *filters.EventType {
				matches = false
			}
		}

		// Filter by EventAction
		if matches && filters.EventAction != nil && *filters.EventAction != "" {
			if log.EventAction == nil || *log.EventAction != *filters.EventAction {
				matches = false
			}
		}

		// Filter by Status
		if matches && filters.Status != nil && *filters.Status != "" {
			if log.Status != *filters.Status {
				matches = false
			}
		}

		if matches {
			filteredLogs = append(filteredLogs, *log)
		}
	}

	// Get total count before pagination
	total := int64(len(filteredLogs))

	// Sort by timestamp DESC (newest first)
	sort.Slice(filteredLogs, func(i, j int) bool {
		return filteredLogs[i].Timestamp.After(filteredLogs[j].Timestamp)
	})

	// Apply pagination
	limit := filters.Limit
	if limit <= 0 {
		limit = 100 // default
	}
	if limit > 1000 {
		limit = 1000 // max
	}

	offset := filters.Offset
	if offset < 0 {
		offset = 0
	}

	// Apply offset and limit
	start := offset
	end := offset + limit
	if start > len(filteredLogs) {
		start = len(filteredLogs)
	}
	if end > len(filteredLogs) {
		end = len(filteredLogs)
	}

	if start >= end {
		return []v1models.AuditLog{}, total, nil
	}

	paginatedLogs := filteredLogs[start:end]
	return paginatedLogs, total, nil
}

// GetLogs returns all logs stored in the mock (useful for test assertions)
func (m *MockRepository) GetLogs() []*v1models.AuditLog {
	return m.logs
}

// ClearLogs clears all stored logs (useful for test cleanup)
func (m *MockRepository) ClearLogs() {
	m.logs = make([]*v1models.AuditLog, 0)
}
