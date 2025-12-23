package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/audit-service/v1/database"
	v1models "github.com/gov-dx-sandbox/audit-service/v1/models"
	v1types "github.com/gov-dx-sandbox/audit-service/v1/types"
)

// AuditService handles generalized audit log operations
type AuditService struct {
	repo database.AuditRepository
}

// NewAuditService creates a new audit service instance using the database repository
func NewAuditService(repo database.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

// CreateAuditLog creates a new audit log entry from a request
func (s *AuditService) CreateAuditLog(ctx context.Context, req *v1types.CreateAuditLogRequest) (*v1models.AuditLog, error) {
	// Convert request to model
	auditLog := &v1models.AuditLog{
		EventName:            req.EventName,
		Action:               req.Action,
		Status:               req.Status,
		ActorType:            req.ActorType,
		ActorID:              req.ActorID,
		Metadata:             req.Metadata,
		Changes:              req.Changes,
		SourceService:        req.SourceService,
		TargetService:        req.TargetService,
		ResourceType:         req.ResourceType,
		IPAddress:            req.IPAddress,
		UserAgent:            req.UserAgent,
		RequestID:            req.RequestID,
		SessionID:            req.SessionID,
		MicroappID:           req.MicroappID,
		Platform:             req.Platform,
		ErrorMessage:         req.ErrorMessage,
		ErrorCode:            req.ErrorCode,
		AuthorizationGroups:  req.AuthorizationGroups,
		AuthenticationMethod: req.AuthenticationMethod,
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

	// Handle resource ID
	if req.ResourceID != nil && *req.ResourceID != "" {
		if resourceUUID, err := uuid.Parse(*req.ResourceID); err == nil {
			auditLog.ResourceID = &resourceUUID
		}
	}

	// Validate before creating
	if err := auditLog.Validate(); err != nil {
		return nil, err
	}

	// Create in database using repository
	createdLog, err := s.repo.CreateAuditLog(ctx, auditLog)
	if err != nil {
		return nil, err
	}

	return createdLog, nil
}

// GetAuditLogs retrieves audit logs with optional filtering
func (s *AuditService) GetAuditLogs(traceID *string, eventName *string, limit, offset int) ([]v1models.AuditLog, int64, error) {
	filters := &database.AuditLogFilters{
		TraceID:   traceID,
		EventName: eventName,
		Limit:     limit,
		Offset:    offset,
	}

	return s.repo.GetAuditLogs(context.Background(), filters)
}

// GetAuditLogsByTraceID retrieves audit logs by trace ID (convenience method)
func (s *AuditService) GetAuditLogsByTraceID(traceID string) ([]v1models.AuditLog, error) {
	return s.repo.GetAuditLogsByTraceID(context.Background(), traceID)
}
