package testutil

import (
	"context"

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

// GetAuditLogsByTraceID returns empty results (can be extended for more complex test scenarios)
func (m *MockRepository) GetAuditLogsByTraceID(ctx context.Context, traceID string) ([]v1models.AuditLog, error) {
	return nil, nil
}

// GetAuditLogs returns empty results (can be extended for more complex test scenarios)
func (m *MockRepository) GetAuditLogs(ctx context.Context, filters *database.AuditLogFilters) ([]v1models.AuditLog, int64, error) {
	return nil, 0, nil
}

// GetLogs returns all logs stored in the mock (useful for test assertions)
func (m *MockRepository) GetLogs() []*v1models.AuditLog {
	return m.logs
}

// ClearLogs clears all stored logs (useful for test cleanup)
func (m *MockRepository) ClearLogs() {
	m.logs = make([]*v1models.AuditLog, 0)
}

