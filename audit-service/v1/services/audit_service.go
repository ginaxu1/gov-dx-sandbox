package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/audit-service/v1/models"
	"gorm.io/gorm"
)

// AuditService interface defines methods for handling audit logs
type AuditService interface {
	CreateAuditLog(ctx context.Context, log *models.AuditLog) (*models.AuditLog, error)
	GetAuditLogs(ctx context.Context, traceID uuid.UUID) ([]models.AuditLog, error)
}

// auditService implementation
type auditService struct {
	db *gorm.DB
}

// NewAuditService creates a new instance of AuditService
func NewAuditService(db *gorm.DB) AuditService {
	// Auto-migrate the table
	if err := db.AutoMigrate(&models.AuditLog{}); err != nil {
		// Log migration error but don't fail service creation
		// This allows the service to start even if migration fails
		// The actual database operation will fail later if schema is wrong
		slog.Warn("Failed to auto-migrate audit_logs table", "error", err)
	}
	return &auditService{db: db}
}

// CreateAuditLog creates a new audit log entry
func (s *auditService) CreateAuditLog(ctx context.Context, log *models.AuditLog) (*models.AuditLog, error) {
	// Check if context is already cancelled
	if ctx.Err() != nil {
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	// Validate required fields before database operation
	if log.TraceID == uuid.Nil {
		return nil, errors.New("traceId cannot be nil")
	}
	if log.SourceService == "" {
		return nil, errors.New("sourceService is required")
	}
	if log.EventType == "" {
		return nil, errors.New("eventType is required")
	}
	if log.Status != models.StatusSuccess && log.Status != models.StatusFailure {
		return nil, fmt.Errorf("invalid status: %s (must be SUCCESS or FAILURE)", log.Status)
	}

	result := s.db.WithContext(ctx).Create(log)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create audit log: %w", result.Error)
	}

	return log, nil
}

// GetAuditLogs retrieves audit logs by trace ID
func (s *auditService) GetAuditLogs(ctx context.Context, traceID uuid.UUID) ([]models.AuditLog, error) {
	// Check if context is already cancelled
	if ctx.Err() != nil {
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	// Validate traceID
	if traceID == uuid.Nil {
		return nil, errors.New("traceId cannot be nil")
	}

	var logs []models.AuditLog
	// Order by timestamp to show the flow chronologically
	// GORM will map uuid.UUID to the UUID column correctly
	result := s.db.WithContext(ctx).
		Where("trace_id = ?", traceID).
		Order("timestamp ASC").
		Find(&logs)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to retrieve audit logs: %w", result.Error)
	}

	// Return empty slice instead of nil if no logs found
	if logs == nil {
		logs = []models.AuditLog{}
	}

	return logs, nil
}
