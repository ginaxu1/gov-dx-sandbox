package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/gov-dx-sandbox/audit-service/models"
	"gorm.io/gorm"
)

// TestConstants contains shared test constants
var (
	TestActorID   = "actor-123"
	TestActorRole = "ADMIN"
	TestTargetID  = "target-123"
	TestOwnerID   = "owner-123"
	TestConsumerID = "consumer-123"
	TestProviderID = "provider-123"
	TestAppID     = "app-123"
	TestSchemaID  = "schema-123"
)

// CreateTestDataExchangeEventRequest creates a test data exchange event request
func CreateTestDataExchangeEventRequest(t *testing.T, overrides ...func(*models.CreateDataExchangeEventRequest)) *models.CreateDataExchangeEventRequest {
	req := &models.CreateDataExchangeEventRequest{
		Timestamp:         time.Now().UTC().Format(time.RFC3339),
		Status:            "success",
		ApplicationID:     TestAppID,
		SchemaID:          TestSchemaID,
		RequestedData:     json.RawMessage(`{}`),
		OnBehalfOfOwnerID: &TestOwnerID,
		ConsumerID:        &TestConsumerID,
		ProviderID:        &TestProviderID,
	}
	for _, override := range overrides {
		override(req)
	}
	return req
}

// CreateTestManagementEventRequest creates a test management event request
func CreateTestManagementEventRequest(t *testing.T, overrides ...func(*models.CreateManagementEventRequest)) *models.CreateManagementEventRequest {
	actorID := TestActorID
	actorRole := TestActorRole
	targetID := TestTargetID
	req := &models.CreateManagementEventRequest{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		EventType: "CREATE",
		Status:    "success",
		Actor: models.Actor{
			Type: "USER",
			ID:   &actorID,
			Role: &actorRole,
		},
		Target: models.Target{
			Resource:   "APPLICATIONS",
			ResourceID: &targetID,
		},
	}
	for _, override := range overrides {
		override(req)
	}
	return req
}

// CreateTestEvents creates multiple test events for testing queries
func CreateTestEvents(t *testing.T, service *DataExchangeEventService, count int, baseTime time.Time) {
	for i := 0; i < count; i++ {
		req := CreateTestDataExchangeEventRequest(t, func(r *models.CreateDataExchangeEventRequest) {
			r.Timestamp = baseTime.Add(time.Duration(i) * time.Minute).Format(time.RFC3339)
		})
		_, err := service.CreateDataExchangeEvent(context.Background(), req)
		if err != nil {
			t.Fatalf("Failed to create test event: %v", err)
		}
	}
}

// CreateTestManagementEvents creates multiple test management events for testing queries
func CreateTestManagementEvents(t *testing.T, service *ManagementEventService, count int, baseTime time.Time) {
	actorID := TestActorID
	actorRole := TestActorRole
	targetID := TestTargetID
	for i := 0; i < count; i++ {
		req := &models.CreateManagementEventRequest{
			Timestamp: baseTime.Add(time.Duration(i) * time.Minute).Format(time.RFC3339),
			EventType: "CREATE",
			Status:    "success",
			Actor: models.Actor{
				Type: "USER",
				ID:   &actorID,
				Role: &actorRole,
			},
			Target: models.Target{
				Resource:   "APPLICATIONS",
				ResourceID: &targetID,
			},
		}
		_, err := service.CreateManagementEvent(context.Background(), req)
		if err != nil {
			t.Fatalf("Failed to create test management event: %v", err)
		}
	}
}

// SetupTestService creates a test service with a fresh database
func SetupTestService(t *testing.T, serviceFactory func(*gorm.DB) interface{}) (interface{}, *gorm.DB) {
	db := SetupSQLiteTestDB(t)
	return serviceFactory(db), db
}

