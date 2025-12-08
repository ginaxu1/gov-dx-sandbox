package services

import (
	"time"

	"github.com/google/uuid"
	v1models "github.com/gov-dx-sandbox/audit-service/v1/models"
	v1types "github.com/gov-dx-sandbox/audit-service/v1/types"
	"gorm.io/gorm"
)

// AuditService handles generalized audit log operations
type AuditService struct {
	db *gorm.DB
}

// NewAuditService creates a new audit service instance
func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

// CreateAuditLog creates a new audit log entry from a request
func (s *AuditService) CreateAuditLog(req *v1types.CreateAuditLogRequest) (*v1models.AuditLog, error) {
	// Convert request to model
	auditLog := &v1models.AuditLog{
		EventName:        req.EventName,
		EventType:        req.EventType,
		Status:           req.Status,
		ActorType:        req.ActorType,
		TargetType:       req.TargetType,
		RequestedData:    req.RequestedData,
		ResponseMetadata: req.ResponseMetadata,
		EventMetadata:    req.EventMetadata,
		ActorMetadata:    req.ActorMetadata,
		TargetMetadata:   req.TargetMetadata,
	}

	// Handle timestamp
	if req.Timestamp != nil && *req.Timestamp != "" {
		if t, err := time.Parse(time.RFC3339, *req.Timestamp); err == nil {
			auditLog.Timestamp = t.UTC()
		}
	}
	// If timestamp is zero, BeforeCreate hook will set it

	// Handle trace ID
	if req.TraceID != nil && *req.TraceID != "" {
		if traceUUID, err := uuid.Parse(*req.TraceID); err == nil {
			auditLog.TraceID = &traceUUID
		}
	}

	// Handle actor fields
	if req.ActorServiceName != nil {
		auditLog.ActorServiceName = req.ActorServiceName
	}
	if req.ActorUserID != nil {
		if userUUID, err := uuid.Parse(*req.ActorUserID); err == nil {
			auditLog.ActorUserID = &userUUID
		}
	}
	if req.ActorUserType != nil {
		auditLog.ActorUserType = req.ActorUserType
	}

	// Handle target fields
	if req.TargetServiceName != nil {
		auditLog.TargetServiceName = req.TargetServiceName
	}
	if req.TargetResource != nil {
		auditLog.TargetResource = req.TargetResource
	}
	if req.TargetResourceID != nil {
		if resourceUUID, err := uuid.Parse(*req.TargetResourceID); err == nil {
			auditLog.TargetResourceID = &resourceUUID
		}
	}

	// Validate before creating
	if err := auditLog.Validate(); err != nil {
		return nil, err
	}

	// Create in database
	if err := s.db.Create(auditLog).Error; err != nil {
		return nil, err
	}

	return auditLog, nil
}

// GetAuditLogs retrieves audit logs with optional filtering
func (s *AuditService) GetAuditLogs(traceID *string, eventName *string, limit, offset int) ([]v1models.AuditLog, int64, error) {
	var logs []v1models.AuditLog
	var total int64

	query := s.db.Model(&v1models.AuditLog{})

	// Apply filters
	if traceID != nil && *traceID != "" {
		if traceUUID, err := uuid.Parse(*traceID); err == nil {
			query = query.Where("trace_id = ?", traceUUID)
		}
	}
	if eventName != nil && *eventName != "" {
		query = query.Where("event_name = ?", *eventName)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and ordering
	if err := query.Order("timestamp DESC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
