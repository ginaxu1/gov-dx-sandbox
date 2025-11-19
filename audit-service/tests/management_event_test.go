package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gov-dx-sandbox/audit-service/models"
)

// stringPtr is a helper function to convert string to *string
func stringPtr(s string) *string {
	return &s
}

// TestManagementEventEndpoint tests the POST /api/events and GET /api/events endpoints
func TestManagementEventEndpoint(t *testing.T) {
	server := SetupTestServerWithGORM(t)
	defer server.Close()

	t.Run("CreateEvent_Success_UserActor", func(t *testing.T) {
		actorID := "user-123"
		actorRole := "ADMIN"
		reqBody := models.ManagementEventRequest{
			EventID:   "550e8400-e29b-41d4-a716-446655440010",
			EventType: "CREATE",
			Status:    "SUCCESS",
			Actor: models.Actor{
				Type: "USER",
				ID:   &actorID,
				Role: &actorRole,
			},
			Target: models.Target{
				Resource:   "SCHEMAS",
				ResourceID: stringPtr("schema-456"),
			},
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/events", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.ManagementEventHandler.CreateEvent(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
		}

		var response models.ManagementEvent
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.EventType != "CREATE" {
			t.Errorf("Expected eventType 'CREATE', got %s", response.EventType)
		}

		if response.ActorType != "USER" {
			t.Errorf("Expected actorType 'USER', got %s", response.ActorType)
		}

		if response.ActorID == nil || *response.ActorID != actorID {
			t.Errorf("Expected actorId %s, got %v", actorID, response.ActorID)
		}

		if response.ActorRole == nil || *response.ActorRole != actorRole {
			t.Errorf("Expected actorRole %s, got %v", actorRole, response.ActorRole)
		}

		if response.TargetResource != "SCHEMAS" {
			t.Errorf("Expected targetResource 'SCHEMAS', got %s", response.TargetResource)
		}

		if response.TargetResourceID == nil || *response.TargetResourceID != "schema-456" {
			t.Errorf("Expected targetResourceId 'schema-456', got %v", response.TargetResourceID)
		}
	})

	t.Run("CreateEvent_Success_ServiceActor", func(t *testing.T) {
		reqBody := models.ManagementEventRequest{
			EventID:   "550e8400-e29b-41d4-a716-446655440011",
			EventType: "UPDATE",
			Status:    "SUCCESS",
			Actor: models.Actor{
				Type: "SERVICE",
			},
			Target: models.Target{
				Resource:   "POLICY-METADATA",
				ResourceID: stringPtr("schema-789"),
			},
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/events", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.ManagementEventHandler.CreateEvent(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
		}

		var response models.ManagementEvent
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.ActorType != "SERVICE" {
			t.Errorf("Expected actorType 'SERVICE', got %s", response.ActorType)
		}

		if response.ActorID != nil {
			t.Errorf("Expected actorId to be nil for SERVICE type, got %v", response.ActorID)
		}

		if response.ActorRole != nil {
			t.Errorf("Expected actorRole to be nil for SERVICE type, got %v", response.ActorRole)
		}
	})

	t.Run("CreateEvent_WithMetadata", func(t *testing.T) {
		actorID := "user-456"
		actorRole := "MEMBER"
		metadata := map[string]interface{}{
			"oldValue": "old-schema-name",
			"newValue": "new-schema-name",
		}

		reqBody := models.ManagementEventRequest{
			EventID:   "550e8400-e29b-41d4-a716-446655440012",
			EventType: "UPDATE",
			Status:    "SUCCESS",
			Actor: models.Actor{
				Type: "USER",
				ID:   &actorID,
				Role: &actorRole,
			},
			Target: models.Target{
				Resource:   "SCHEMAS",
				ResourceID: stringPtr("schema-789"),
			},
			Metadata: &metadata,
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/events", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.ManagementEventHandler.CreateEvent(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
		}

		var response models.ManagementEvent
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Metadata == nil {
			t.Error("Expected metadata to be present")
		}
	})

	t.Run("CreateEvent_MissingEventType", func(t *testing.T) {
		reqBody := models.ManagementEventRequest{
			Actor:  models.Actor{Type: "USER"},
			Target: models.Target{Resource: "SCHEMAS", ResourceID: stringPtr("schema-123")},
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/events", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.ManagementEventHandler.CreateEvent(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})

	t.Run("CreateEvent_MissingActorType", func(t *testing.T) {
		reqBody := models.ManagementEventRequest{
			EventType: "CREATE",
			Target:    models.Target{Resource: "SCHEMAS", ResourceID: stringPtr("schema-123")},
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/events", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.ManagementEventHandler.CreateEvent(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})

	t.Run("CreateEvent_UserActorMissingRole", func(t *testing.T) {
		actorID := "user-123"
		reqBody := models.ManagementEventRequest{
			EventType: "CREATE",
			Actor: models.Actor{
				Type: "USER",
				ID:   &actorID,
				// Role is missing
			},
			Target: models.Target{Resource: "SCHEMAS", ResourceID: stringPtr("schema-123")},
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/events", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.ManagementEventHandler.CreateEvent(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})

	t.Run("CreateEvent_InvalidEventType", func(t *testing.T) {
		actorID := "user-123"
		actorRole := "ADMIN"
		reqBody := models.ManagementEventRequest{
			EventType: "INVALID",
			Actor: models.Actor{
				Type: "USER",
				ID:   &actorID,
				Role: &actorRole,
			},
			Target: models.Target{Resource: "SCHEMAS", ResourceID: stringPtr("schema-123")},
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/events", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.ManagementEventHandler.CreateEvent(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})

	t.Run("CreateEvent_InvalidActorType", func(t *testing.T) {
		reqBody := models.ManagementEventRequest{
			EventType: "CREATE",
			Actor: models.Actor{
				Type: "INVALID",
			},
			Target: models.Target{Resource: "SCHEMAS", ResourceID: stringPtr("schema-123")},
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/events", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.ManagementEventHandler.CreateEvent(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})

	t.Run("CreateEvent_InvalidTargetResource", func(t *testing.T) {
		actorID := "user-123"
		actorRole := "ADMIN"
		reqBody := models.ManagementEventRequest{
			EventType: "CREATE",
			Actor: models.Actor{
				Type: "USER",
				ID:   &actorID,
				Role: &actorRole,
			},
			Target: models.Target{
				Resource:   "INVALID-RESOURCE",
				ResourceID: stringPtr("schema-123"),
			},
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/events", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.ManagementEventHandler.CreateEvent(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})

	t.Run("GetEvents_AllEvents", func(t *testing.T) {
		// Create some test events first
		createTestEvents(t, server)

		req := httptest.NewRequest("GET", "/api/events", nil)
		w := httptest.NewRecorder()

		server.ManagementEventHandler.GetEvents(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response models.ManagementEventResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Total < 3 {
			t.Errorf("Expected at least 3 events, got %d", response.Total)
		}

		if len(response.Events) == 0 {
			t.Error("Expected at least one event in response")
		}
	})

	t.Run("GetEvents_FilterByEventType", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/events?eventType=CREATE", nil)
		w := httptest.NewRecorder()

		server.ManagementEventHandler.GetEvents(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response models.ManagementEventResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		for _, event := range response.Events {
			if event.EventType != "CREATE" {
				t.Errorf("Expected all events to have eventType 'CREATE', got %s", event.EventType)
			}
		}
	})

	t.Run("GetEvents_FilterByActorType", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/events?actorType=USER", nil)
		w := httptest.NewRecorder()

		server.ManagementEventHandler.GetEvents(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response models.ManagementEventResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		for _, event := range response.Events {
			if event.ActorType != "USER" {
				t.Errorf("Expected all events to have actorType 'USER', got %s", event.ActorType)
			}
		}
	})

	t.Run("GetEvents_FilterByTargetResource", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/events?targetResource=SCHEMAS", nil)
		w := httptest.NewRecorder()

		server.ManagementEventHandler.GetEvents(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response models.ManagementEventResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		for _, event := range response.Events {
			if event.TargetResource != "SCHEMAS" {
				t.Errorf("Expected all events to have targetResource 'SCHEMAS', got %s", event.TargetResource)
			}
		}
	})

	t.Run("GetEvents_WithPagination", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/events?limit=2&offset=0", nil)
		w := httptest.NewRecorder()

		server.ManagementEventHandler.GetEvents(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response models.ManagementEventResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Limit != 2 {
			t.Errorf("Expected limit 2, got %d", response.Limit)
		}

		if len(response.Events) > 2 {
			t.Errorf("Expected at most 2 events, got %d", len(response.Events))
		}
	})

	t.Run("GetEvents_CombinedFilters", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/events?eventType=CREATE&actorType=USER&targetResource=SCHEMAS", nil)
		w := httptest.NewRecorder()

		server.ManagementEventHandler.GetEvents(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response models.ManagementEventResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		for _, event := range response.Events {
			if event.EventType != "CREATE" {
				t.Errorf("Expected eventType 'CREATE', got %s", event.EventType)
			}
			if event.ActorType != "USER" {
				t.Errorf("Expected actorType 'USER', got %s", event.ActorType)
			}
			if event.TargetResource != "SCHEMAS" {
				t.Errorf("Expected targetResource 'SCHEMAS', got %s", event.TargetResource)
			}
		}
	})

	t.Run("CreateEvent_CREATE_FAILURE_WithoutResourceID", func(t *testing.T) {
		// Test that CREATE failures can be logged without ResourceID
		actorID := "user-123"
		actorRole := "ADMIN"
		reqBody := models.ManagementEventRequest{
			EventID:   "550e8400-e29b-41d4-a716-446655440013",
			EventType: "CREATE",
			Status:    "FAILURE",
			Actor: models.Actor{
				Type: "USER",
				ID:   &actorID,
				Role: &actorRole,
			},
			Target: models.Target{
				Resource:   "SCHEMAS",
				ResourceID: nil, // Empty ResourceID allowed for CREATE failures
			},
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/events", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.ManagementEventHandler.CreateEvent(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
		}

		var response models.ManagementEvent
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Status != "FAILURE" {
			t.Errorf("Expected status 'FAILURE', got %s", response.Status)
		}

		if response.TargetResourceID != nil {
			t.Errorf("Expected targetResourceId to be nil for CREATE failure, got %v", response.TargetResourceID)
		}
	})

	t.Run("CreateEvent_UPDATE_FAILURE_RequiresResourceID", func(t *testing.T) {
		// Test that UPDATE failures still require ResourceID
		actorID := "user-123"
		actorRole := "ADMIN"
		reqBody := models.ManagementEventRequest{
			EventID:   "550e8400-e29b-41d4-a716-446655440014",
			EventType: "UPDATE",
			Status:    "FAILURE",
			Actor: models.Actor{
				Type: "USER",
				ID:   &actorID,
				Role: &actorRole,
			},
			Target: models.Target{
				Resource:   "SCHEMAS",
				ResourceID: nil, // Empty ResourceID should be rejected for UPDATE
			},
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/events", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.ManagementEventHandler.CreateEvent(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})
}

// createTestEvents creates some test events for filtering tests
func createTestEvents(t *testing.T, server *TestServerWithGORM) {
	actorID1 := "user-123"
	actorRole1 := "ADMIN"
	actorID2 := "user-456"
	actorRole2 := "MEMBER"

	events := []models.ManagementEventRequest{
		{
			EventType: "CREATE",
			Status:    "SUCCESS",
			Actor: models.Actor{
				Type: "USER",
				ID:   &actorID1,
				Role: &actorRole1,
			},
			Target: models.Target{
				Resource:   "SCHEMAS",
				ResourceID: stringPtr("schema-1"),
			},
		},
		{
			EventType: "UPDATE",
			Status:    "SUCCESS",
			Actor: models.Actor{
				Type: "USER",
				ID:   &actorID2,
				Role: &actorRole2,
			},
			Target: models.Target{
				Resource:   "APPLICATIONS",
				ResourceID: stringPtr("app-1"),
			},
		},
		{
			EventType: "DELETE",
			Status:    "SUCCESS",
			Actor: models.Actor{
				Type: "SERVICE",
			},
			Target: models.Target{
				Resource:   "POLICY-METADATA",
				ResourceID: stringPtr("schema-2"),
			},
		},
	}

	for _, eventReq := range events {
		_, err := server.ManagementEventService.CreateManagementEvent(server.Context, &eventReq)
		if err != nil {
			t.Fatalf("Failed to create test event: %v", err)
		}
	}
}
