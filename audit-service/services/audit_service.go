package services

import (
	"context"

	"github.com/gov-dx-sandbox/audit-service/models"
	"gorm.io/gorm"
)

// AuditService interface defines methods for handling audit logs
type AuditService interface {
	CreateAuditLog(ctx context.Context, log *models.AuditLog) (*models.AuditLog, error)
	GetAuditLogs(ctx context.Context, traceID string) ([]models.AuditLog, error)
}

// auditService implementation
type auditService struct {
	db *gorm.DB
}

// NewAuditService creates a new instance of AuditService
func NewAuditService(db *gorm.DB) AuditService {
	// Auto-migrate the table
	db.AutoMigrate(&models.AuditLog{})
	return &auditService{db: db}
}

// CreateAuditLog creates a new audit log entry
func (s *auditService) CreateAuditLog(ctx context.Context, log *models.AuditLog) (*models.AuditLog, error) {
	result := s.db.WithContext(ctx).Create(log)
	if result.Error != nil {
		return nil, result.Error
	}
	return log, nil
}

// GetAuditLogs retrieves audit logs by trace ID
func (s *auditService) GetAuditLogs(ctx context.Context, traceID string) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	// Order by timestamp to show the flow
	result := s.db.WithContext(ctx).Where("trace_id = ?", traceID).Order("timestamp asc").Find(&logs)
	if result.Error != nil {
		return nil, result.Error
	}
	return logs, nil
}
