package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/audit-service/config"
	"github.com/gov-dx-sandbox/audit-service/v1/database"
	v1models "github.com/gov-dx-sandbox/audit-service/v1/models"
	"github.com/stretchr/testify/assert"
)

// mockRepository is a simple mock implementation for testing
type mockRepository struct {
	logs []*v1models.AuditLog
}

func (m *mockRepository) CreateAuditLog(ctx context.Context, log *v1models.AuditLog) (*v1models.AuditLog, error) {
	// Simulate BeforeCreate hook behavior
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	m.logs = append(m.logs, log)
	return log, nil
}

func (m *mockRepository) GetAuditLogsByTraceID(ctx context.Context, traceID string) ([]v1models.AuditLog, error) {
	return nil, nil
}

func (m *mockRepository) GetAuditLogs(ctx context.Context, filters *database.AuditLogFilters) ([]v1models.AuditLog, int64, error) {
	return nil, 0, nil
}

func TestAuditService_CreateAuditLog_Validation(t *testing.T) {
	// Set up enum configuration
	enums := &config.AuditEnums{
		EventTypes:   []string{"POLICY_CHECK", "MANAGEMENT_EVENT"},
		EventActions: []string{"CREATE", "READ", "UPDATE", "DELETE"},
		ActorTypes:   []string{"SERVICE", "ADMIN", "MEMBER", "SYSTEM"},
		TargetTypes:  []string{"SERVICE", "RESOURCE"},
	}
	enums.InitializeMaps()
	v1models.SetEnumConfig(enums)

	mockRepo := &mockRepository{}
	service := NewAuditService(mockRepo)

	tests := []struct {
		name    string
		req     *v1models.CreateAuditLogRequest
		wantErr bool
	}{
		{
			name: "Valid request with SERVICE actor",
			req: &v1models.CreateAuditLogRequest{
				Timestamp:  time.Now().UTC().Format(time.RFC3339),
				Status:     v1models.StatusSuccess,
				ActorType:  "SERVICE",
				ActorID:    "orchestration-engine",
				TargetType: "SERVICE",
				TargetID:   stringPtr("consent-engine"),
				EventType:  stringPtr("POLICY_CHECK"),
			},
			wantErr: false,
		},
		{
			name: "Valid request with ADMIN actor",
			req: &v1models.CreateAuditLogRequest{
				Timestamp:   time.Now().UTC().Format(time.RFC3339),
				Status:      v1models.StatusSuccess,
				ActorType:   "ADMIN",
				ActorID:     "admin@example.com",
				TargetType:  "RESOURCE",
				TargetID:    stringPtr("user-123"),
				EventAction: stringPtr("CREATE"),
			},
			wantErr: false,
		},
		{
			name: "Invalid actor type",
			req: &v1models.CreateAuditLogRequest{
				Timestamp:  time.Now().UTC().Format(time.RFC3339),
				Status:     v1models.StatusSuccess,
				ActorType:  "INVALID",
				ActorID:    "actor-1",
				TargetType: "SERVICE",
				TargetID:   stringPtr("service-1"),
			},
			wantErr: true,
		},
		{
			name: "Missing actor ID",
			req: &v1models.CreateAuditLogRequest{
				Timestamp:  time.Now().UTC().Format(time.RFC3339),
				Status:     v1models.StatusSuccess,
				ActorType:  "SERVICE",
				ActorID:    "",
				TargetType: "SERVICE",
				TargetID:   stringPtr("service-1"),
			},
			wantErr: true,
		},
		{
			name: "Invalid event type",
			req: &v1models.CreateAuditLogRequest{
				Timestamp:  time.Now().UTC().Format(time.RFC3339),
				Status:     v1models.StatusSuccess,
				ActorType:  "SERVICE",
				ActorID:    "service-1",
				TargetType: "SERVICE",
				TargetID:   stringPtr("service-2"),
				EventType:  stringPtr("INVALID_EVENT"),
			},
			wantErr: true,
		},
		{
			name: "Missing timestamp",
			req: &v1models.CreateAuditLogRequest{
				Status:     v1models.StatusSuccess,
				ActorType:  "SERVICE",
				ActorID:    "service-1",
				TargetType: "SERVICE",
				TargetID:   stringPtr("service-1"),
			},
			wantErr: true,
		},
		{
			name: "Invalid timestamp format",
			req: &v1models.CreateAuditLogRequest{
				Timestamp:  "invalid-timestamp",
				Status:     v1models.StatusSuccess,
				ActorType:  "SERVICE",
				ActorID:    "service-1",
				TargetType: "SERVICE",
				TargetID:   stringPtr("service-1"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, err := service.CreateAuditLog(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err, "Expected validation error")
				assert.Nil(t, log)
			} else {
				assert.NoError(t, err, "Expected no validation error")
				assert.NotNil(t, log)
				assert.NotEmpty(t, log.ID)
				assert.Equal(t, tt.req.Status, log.Status)
				assert.Equal(t, tt.req.ActorType, log.ActorType)
				assert.Equal(t, tt.req.ActorID, log.ActorID)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
