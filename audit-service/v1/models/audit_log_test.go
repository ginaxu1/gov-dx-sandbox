package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/audit-service/config"
)

func TestAuditLog_Validate_WithConfig(t *testing.T) {
	// Set up enum configuration using AuditEnums type
	enums := &config.AuditEnums{
		EventTypes:   []string{"POLICY_CHECK", "MANAGEMENT_EVENT"},
		EventActions: []string{"CREATE", "READ", "UPDATE", "DELETE"},
		ActorTypes:   []string{"SERVICE", "ADMIN", "MEMBER", "SYSTEM"},
		TargetTypes:  []string{"SERVICE", "RESOURCE"},
	}
	// Initialize maps for O(1) validation
	enums.InitializeMaps()
	SetEnumConfig(enums)

	tests := []struct {
		name    string
		log     AuditLog
		wantErr bool
	}{
		{
			name: "Valid audit log with SERVICE actor",
			log: AuditLog{
				Status:     StatusSuccess,
				ActorType:  "SERVICE",
				ActorID:    "orchestration-engine",
				TargetType: "SERVICE",
				TargetID:   stringPtr("consent-engine"),
			},
			wantErr: false,
		},
		{
			name: "Valid audit log with ADMIN actor",
			log: AuditLog{
				Status:     StatusSuccess,
				ActorType:  "ADMIN",
				ActorID:    "admin@example.com",
				TargetType: "RESOURCE",
				TargetID:   stringPtr("user-123"),
			},
			wantErr: false,
		},
		{
			name: "Valid audit log with SYSTEM actor",
			log: AuditLog{
				Status:     StatusSuccess,
				ActorType:  "SYSTEM",
				ActorID:    "system",
				TargetType: "SERVICE",
				TargetID:   stringPtr("audit-service"),
			},
			wantErr: false,
		},
		{
			name: "Invalid status",
			log: AuditLog{
				Status:     "INVALID",
				ActorType:  "SERVICE",
				ActorID:    "service-1",
				TargetType: "SERVICE",
				TargetID:   stringPtr("service-2"),
			},
			wantErr: true,
		},
		{
			name: "Invalid actor type",
			log: AuditLog{
				Status:     StatusSuccess,
				ActorType:  "INVALID",
				ActorID:    "actor-1",
				TargetType: "SERVICE",
				TargetID:   stringPtr("service-1"),
			},
			wantErr: true,
		},
		{
			name: "Missing actor ID",
			log: AuditLog{
				Status:     StatusSuccess,
				ActorType:  "SERVICE",
				ActorID:    "", // Required
				TargetType: "SERVICE",
				TargetID:   stringPtr("service-1"),
			},
			wantErr: true,
		},
		{
			name: "Invalid target type",
			log: AuditLog{
				Status:     StatusSuccess,
				ActorType:  "SERVICE",
				ActorID:    "service-1",
				TargetType: "INVALID",
				TargetID:   stringPtr("target-1"),
			},
			wantErr: true,
		},
		{
			name: "Valid event type from config",
			log: AuditLog{
				Status:     StatusSuccess,
				EventType:  stringPtr("POLICY_CHECK"),
				ActorType:  "SERVICE",
				ActorID:    "service-1",
				TargetType: "SERVICE",
				TargetID:   stringPtr("service-2"),
			},
			wantErr: false,
		},
		{
			name: "Invalid event type (not in config)",
			log: AuditLog{
				Status:     StatusSuccess,
				EventType:  stringPtr("INVALID_EVENT"),
				ActorType:  "SERVICE",
				ActorID:    "service-1",
				TargetType: "SERVICE",
				TargetID:   stringPtr("service-2"),
			},
			wantErr: true,
		},
		{
			name: "Valid event action from config",
			log: AuditLog{
				Status:      StatusSuccess,
				EventAction: stringPtr("CREATE"),
				ActorType:   "SERVICE",
				ActorID:     "service-1",
				TargetType:  "SERVICE",
				TargetID:    stringPtr("service-2"),
			},
			wantErr: false,
		},
		{
			name: "Invalid event action (not in config)",
			log: AuditLog{
				Status:      StatusSuccess,
				EventAction: stringPtr("INVALID_ACTION"),
				ActorType:   "SERVICE",
				ActorID:     "service-1",
				TargetType:  "SERVICE",
				TargetID:    stringPtr("service-2"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.log.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuditLog_Validate_WithoutConfig(t *testing.T) {
	// Reset enum config to nil to test fallback behavior
	enumConfig = nil

	log := AuditLog{
		Status:     StatusSuccess,
		ActorType:  "SERVICE",
		ActorID:    "service-1",
		TargetType: "SERVICE",
		TargetID:   stringPtr("service-2"),
	}

	// Should still validate successfully using default enums from config
	if err := log.Validate(); err != nil {
		t.Errorf("Validate() should work with default constants, got error: %v", err)
	}
}

func TestAuditLog_BeforeCreate_AutoGeneratesTraceID(t *testing.T) {
	// Set up enum configuration using AuditEnums type
	enums := &config.AuditEnums{
		EventTypes:   []string{"POLICY_CHECK", "MANAGEMENT_EVENT"},
		EventActions: []string{"CREATE", "READ", "UPDATE", "DELETE"},
		ActorTypes:   []string{"SERVICE", "ADMIN", "MEMBER", "SYSTEM"},
		TargetTypes:  []string{"SERVICE", "RESOURCE"},
	}
	// Initialize maps for O(1) validation
	enums.InitializeMaps()
	SetEnumConfig(enums)

	log := AuditLog{
		Status:     StatusSuccess,
		ActorType:  "SERVICE",
		ActorID:    "service-1",
		TargetType: "SERVICE",
		TargetID:   stringPtr("service-2"),
		// TraceID is nil - trace_id is not auto-generated, can remain nil
	}

	// Simulate BeforeCreate hook
	if err := log.BeforeCreate(nil); err != nil {
		t.Fatalf("BeforeCreate() should not return error, got: %v", err)
	}

	// Verify trace_id remains nil when not provided (trace_id is optional)
	if log.TraceID != nil {
		t.Error("TraceID should remain nil when not provided")
	}

	// Verify ID was generated
	if log.ID == uuid.Nil {
		t.Error("ID should be generated by BeforeCreate hook")
	}

	// Verify timestamp was set
	if log.Timestamp.IsZero() {
		t.Error("Timestamp should be set by BeforeCreate hook")
	}

	// Verify BaseModel timestamp was set
	if log.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set by BaseModel BeforeCreate hook")
	}
}

func stringPtr(s string) *string {
	return &s
}
