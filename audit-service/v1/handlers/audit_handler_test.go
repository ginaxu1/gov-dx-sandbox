package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/audit-service/config"
	"github.com/gov-dx-sandbox/audit-service/v1/database"
	v1models "github.com/gov-dx-sandbox/audit-service/v1/models"
	v1services "github.com/gov-dx-sandbox/audit-service/v1/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRepository for handler tests
type mockRepository struct {
	logs []*v1models.AuditLog
}

func (m *mockRepository) CreateAuditLog(ctx context.Context, log *v1models.AuditLog) (*v1models.AuditLog, error) {
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

func TestAuditHandler_CreateAuditLog(t *testing.T) {
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
	service := v1services.NewAuditService(mockRepo)
	handler := NewAuditHandler(service)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name: "Valid request",
			requestBody: map[string]interface{}{
				"timestamp":  time.Now().UTC().Format(time.RFC3339),
				"status":     v1models.StatusSuccess,
				"actorType":  "SERVICE",
				"actorId":    "orchestration-engine",
				"targetType": "SERVICE",
				"targetId":   "consent-engine",
				"eventType":  "POLICY_CHECK",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Missing status",
			requestBody: map[string]interface{}{
				"timestamp":  time.Now().UTC().Format(time.RFC3339),
				"actorType":  "SERVICE",
				"actorId":    "service-1",
				"targetType": "SERVICE",
				"targetId":   "service-2",
			},
			expectedStatus: http.StatusBadRequest, // Validation error from service layer
		},
		{
			name: "Missing actorId",
			requestBody: map[string]interface{}{
				"timestamp":  time.Now().UTC().Format(time.RFC3339),
				"status":     v1models.StatusSuccess,
				"actorType":  "SERVICE",
				"targetType": "SERVICE",
				"targetId":   "service-2",
			},
			expectedStatus: http.StatusBadRequest, // Validation error from service layer
		},
		{
			name: "Invalid actor type",
			requestBody: map[string]interface{}{
				"timestamp":  time.Now().UTC().Format(time.RFC3339),
				"status":     v1models.StatusSuccess,
				"actorType":  "INVALID",
				"actorId":    "actor-1",
				"targetType": "SERVICE",
				"targetId":   "service-1",
			},
			expectedStatus: http.StatusBadRequest, // Validation error from service layer
		},
		{
			name: "Invalid event type",
			requestBody: map[string]interface{}{
				"timestamp":  time.Now().UTC().Format(time.RFC3339),
				"status":     v1models.StatusSuccess,
				"actorType":  "SERVICE",
				"actorId":    "service-1",
				"targetType": "SERVICE",
				"targetId":   "service-2",
				"eventType":  "INVALID_EVENT",
			},
			expectedStatus: http.StatusBadRequest, // Validation error from service layer
		},
		{
			name: "Missing timestamp",
			requestBody: map[string]interface{}{
				"status":     v1models.StatusSuccess,
				"actorType":  "SERVICE",
				"actorId":    "service-1",
				"targetType": "SERVICE",
				"targetId":   "service-2",
			},
			expectedStatus: http.StatusBadRequest, // Validation error - timestamp is required
		},
		{
			name: "Invalid timestamp format",
			requestBody: map[string]interface{}{
				"timestamp":  "invalid-timestamp",
				"status":     v1models.StatusSuccess,
				"actorType":  "SERVICE",
				"actorId":    "service-1",
				"targetType": "SERVICE",
				"targetId":   "service-2",
			},
			expectedStatus: http.StatusBadRequest, // Validation error - invalid timestamp format
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/audit-logs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateAuditLog(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "Expected status %d, got %d", tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response v1models.AuditLog
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.ID)
				assert.Equal(t, tt.requestBody["status"], response.Status)
			}
		})
	}
}
