package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/gov-dx-sandbox/audit-service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDataExchangeEventService(t *testing.T) {
	db := SetupPostgresTestDB(t)
	if db == nil {
		return // test was skipped
	}
	service := NewDataExchangeEventService(db)
	assert.NotNil(t, service)
}

func TestDataExchangeEventService_CreateDataExchangeEvent(t *testing.T) {
	db := SetupPostgresTestDB(t)
	if db == nil {
		return // test was skipped
	}
	service := NewDataExchangeEventService(db)

	t.Run("Success", func(t *testing.T) {
		consumerID := "consumer-123"
		providerID := "provider-456"
		ownerID := "owner-789"
		requestedData := json.RawMessage(`{"field1": "value1", "field2": "value2"}`)
		additionalInfo := json.RawMessage(`{"extra": "info"}`)

		req := &models.CreateDataExchangeEventRequest{
			Timestamp:         "2024-01-01T00:00:00Z",
			Status:            "success",
			ApplicationID:     "app-123",
			SchemaID:          "schema-456",
			RequestedData:     requestedData,
			OnBehalfOfOwnerID: &ownerID,
			ConsumerID:        &consumerID,
			ProviderID:        &providerID,
			AdditionalInfo:    additionalInfo,
		}

		event, err := service.CreateDataExchangeEvent(context.Background(), req)
		require.NoError(t, err)
		assert.NotEmpty(t, event.ID)
		assert.Equal(t, "success", event.Status)
		assert.Equal(t, "app-123", event.ApplicationID)
		assert.Equal(t, "schema-456", event.SchemaID)
		assert.Equal(t, requestedData, event.RequestedData)
		assert.Equal(t, &ownerID, event.OnBehalfOfOwnerID)
		assert.Equal(t, &consumerID, event.ConsumerID)
		assert.Equal(t, &providerID, event.ProviderID)
		assert.Equal(t, additionalInfo, event.AdditionalInfo)
	})

	t.Run("SuccessWithMinimalFields", func(t *testing.T) {
		requestedData := json.RawMessage(`{"required": "data"}`)

		req := &models.CreateDataExchangeEventRequest{
			Timestamp:     "2024-01-01T12:00:00Z",
			Status:        "failure",
			ApplicationID: "app-456",
			SchemaID:      "schema-789",
			RequestedData: requestedData,
		}

		event, err := service.CreateDataExchangeEvent(context.Background(), req)
		require.NoError(t, err)
		assert.NotEmpty(t, event.ID)
		assert.Equal(t, "failure", event.Status)
		assert.Equal(t, "app-456", event.ApplicationID)
		assert.Equal(t, "schema-789", event.SchemaID)
		assert.Equal(t, requestedData, event.RequestedData)
		assert.Nil(t, event.OnBehalfOfOwnerID)
		assert.Nil(t, event.ConsumerID)
		assert.Nil(t, event.ProviderID)
	})

	t.Run("AutoGenerateEventID", func(t *testing.T) {
		requestedData := json.RawMessage(`{"test": "data"}`)

		req := &models.CreateDataExchangeEventRequest{
			Timestamp:     "2024-01-01T12:30:00Z",
			Status:        "success",
			ApplicationID: "app-789",
			SchemaID:      "schema-012",
			RequestedData: requestedData,
		}

		event, err := service.CreateDataExchangeEvent(context.Background(), req)
		require.NoError(t, err)
		assert.NotEmpty(t, event.ID)
	})

	t.Run("InvalidTimestamp", func(t *testing.T) {
		requestedData := json.RawMessage(`{"test": "data"}`)

		req := &models.CreateDataExchangeEventRequest{
			Timestamp:     "invalid-timestamp",
			Status:        "success",
			ApplicationID: "app-123",
			SchemaID:      "schema-456",
			RequestedData: requestedData,
		}

		_, err := service.CreateDataExchangeEvent(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid timestamp format")
	})

	t.Run("InvalidStatus", func(t *testing.T) {
		requestedData := json.RawMessage(`{"test": "data"}`)

		req := &models.CreateDataExchangeEventRequest{
			Timestamp:     "2024-01-01T00:00:00Z",
			Status:        "invalid_status",
			ApplicationID: "app-123",
			SchemaID:      "schema-456",
			RequestedData: requestedData,
		}

		_, err := service.CreateDataExchangeEvent(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status")
	})

	t.Run("WithComplexRequestedData", func(t *testing.T) {
		complexData := json.RawMessage(`{
			"personalInfo": {
				"name": "John Doe",
				"age": 30
			},
			"addresses": [
				{"type": "home", "street": "123 Main St"},
				{"type": "work", "street": "456 Oak Ave"}
			],
			"preferences": {
				"notifications": true,
				"theme": "dark"
			}
		}`)

		req := &models.CreateDataExchangeEventRequest{
			Timestamp:     "2024-01-01T00:00:00Z",
			Status:        "success",
			ApplicationID: "app-complex",
			SchemaID:      "schema-complex",
			RequestedData: complexData,
		}

		event, err := service.CreateDataExchangeEvent(context.Background(), req)
		require.NoError(t, err)
		assert.NotEmpty(t, event.ID)
		// JSON comparison: PostgreSQL may reformat JSON, so compare as strings
		assert.JSONEq(t, string(complexData), string(event.RequestedData))
	})

	t.Run("WithComplexAdditionalInfo", func(t *testing.T) {
		requestedData := json.RawMessage(`{"field": "value"}`)
		additionalInfo := json.RawMessage(`{
			"processingTime": "150ms",
			"dataSize": "2.5MB",
			"encryption": {
				"algorithm": "AES-256",
				"keyVersion": "v2"
			}
		}`)

		req := &models.CreateDataExchangeEventRequest{
			Timestamp:      "2024-01-01T00:00:00Z",
			Status:         "success",
			ApplicationID:  "app-info",
			SchemaID:       "schema-info",
			RequestedData:  requestedData,
			AdditionalInfo: additionalInfo,
		}

		event, err := service.CreateDataExchangeEvent(context.Background(), req)
		require.NoError(t, err)
		assert.NotEmpty(t, event.ID)
		// JSON comparison: PostgreSQL may reformat JSON, so compare as strings
		assert.JSONEq(t, string(additionalInfo), string(event.AdditionalInfo))
	})
}

func TestDataExchangeEventService_GetDataExchangeEvents(t *testing.T) {
	db := SetupPostgresTestDB(t)
	if db == nil {
		return // test was skipped
	}
	service := NewDataExchangeEventService(db)

	// Create test events
	consumerID1 := "consumer-1"
	consumerID2 := "consumer-2"
	providerID1 := "provider-1"
	providerID2 := "provider-2"
	ownerID := "owner-1"

	events := []*models.CreateDataExchangeEventRequest{
		{
			Timestamp:         "2024-01-01T00:00:00Z",
			Status:            "success",
			ApplicationID:     "app-1",
			SchemaID:          "schema-1",
			RequestedData:     json.RawMessage(`{"data": "test1"}`),
			OnBehalfOfOwnerID: &ownerID,
			ConsumerID:        &consumerID1,
			ProviderID:        &providerID1,
		},
		{
			Timestamp:     "2024-01-01T01:00:00Z",
			Status:        "failure",
			ApplicationID: "app-2",
			SchemaID:      "schema-2",
			RequestedData: json.RawMessage(`{"data": "test2"}`),
			ConsumerID:    &consumerID2,
			ProviderID:    &providerID2,
		},
		{
			Timestamp:     "2024-01-01T02:00:00Z",
			Status:        "success",
			ApplicationID: "app-1",
			SchemaID:      "schema-3",
			RequestedData: json.RawMessage(`{"data": "test3"}`),
			ConsumerID:    &consumerID1,
			ProviderID:    &providerID1,
		},
	}

	for _, req := range events {
		_, err := service.CreateDataExchangeEvent(context.Background(), req)
		require.NoError(t, err)
	}

	t.Run("GetAllEvents", func(t *testing.T) {
		filter := &models.DataExchangeEventFilter{}
		response, err := service.GetDataExchangeEvents(context.Background(), filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(response.Events), 3)
		assert.GreaterOrEqual(t, response.Total, int64(3))
	})

	t.Run("FilterByStatus", func(t *testing.T) {
		status := "success"
		filter := &models.DataExchangeEventFilter{
			Status: &status,
		}
		response, err := service.GetDataExchangeEvents(context.Background(), filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(response.Events), 2)
		for _, event := range response.Events {
			assert.Equal(t, "success", event.Status)
		}
	})

	t.Run("FilterByApplicationID", func(t *testing.T) {
		appID := "app-1"
		filter := &models.DataExchangeEventFilter{
			ApplicationID: &appID,
		}
		response, err := service.GetDataExchangeEvents(context.Background(), filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(response.Events), 2)
		for _, event := range response.Events {
			assert.Equal(t, "app-1", event.ApplicationID)
		}
	})

	t.Run("FilterBySchemaID", func(t *testing.T) {
		schemaID := "schema-1"
		filter := &models.DataExchangeEventFilter{
			SchemaID: &schemaID,
		}
		response, err := service.GetDataExchangeEvents(context.Background(), filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(response.Events), 1)
		for _, event := range response.Events {
			assert.Equal(t, "schema-1", event.SchemaID)
		}
	})

	t.Run("FilterByConsumerID", func(t *testing.T) {
		filter := &models.DataExchangeEventFilter{
			ConsumerID: &consumerID1,
		}
		response, err := service.GetDataExchangeEvents(context.Background(), filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(response.Events), 2)
		for _, event := range response.Events {
			assert.NotNil(t, event.ConsumerID)
			assert.Equal(t, consumerID1, *event.ConsumerID)
		}
	})

	t.Run("FilterByProviderID", func(t *testing.T) {
		filter := &models.DataExchangeEventFilter{
			ProviderID: &providerID1,
		}
		response, err := service.GetDataExchangeEvents(context.Background(), filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(response.Events), 2)
		for _, event := range response.Events {
			assert.NotNil(t, event.ProviderID)
			assert.Equal(t, providerID1, *event.ProviderID)
		}
	})

	t.Run("Pagination", func(t *testing.T) {
		filter := &models.DataExchangeEventFilter{
			Limit:  2,
			Offset: 0,
		}
		response, err := service.GetDataExchangeEvents(context.Background(), filter)
		require.NoError(t, err)
		assert.Equal(t, 2, len(response.Events))
		assert.Equal(t, 2, response.Limit)
		assert.Equal(t, 0, response.Offset)
	})

	t.Run("DateRangeFilter", func(t *testing.T) {
		startDate := time.Now().Add(-24 * time.Hour)
		endDate := time.Now().Add(24 * time.Hour)
		filter := &models.DataExchangeEventFilter{
			StartDate: &startDate,
			EndDate:   &endDate,
		}
		response, err := service.GetDataExchangeEvents(context.Background(), filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(response.Events), 0)
	})

	t.Run("DefaultLimit", func(t *testing.T) {
		filter := &models.DataExchangeEventFilter{
			Limit: 0, // Should use default
		}
		response, err := service.GetDataExchangeEvents(context.Background(), filter)
		require.NoError(t, err)
		assert.Equal(t, 50, response.Limit)
	})

	t.Run("MultipleFilters", func(t *testing.T) {
		status := "success"
		appID := "app-1"
		filter := &models.DataExchangeEventFilter{
			Status:        &status,
			ApplicationID: &appID,
		}
		response, err := service.GetDataExchangeEvents(context.Background(), filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(response.Events), 1)
		for _, event := range response.Events {
			assert.Equal(t, "success", event.Status)
			assert.Equal(t, "app-1", event.ApplicationID)
		}
	})

	t.Run("NoResults", func(t *testing.T) {
		nonExistentApp := "non-existent-app"
		filter := &models.DataExchangeEventFilter{
			ApplicationID: &nonExistentApp,
		}
		response, err := service.GetDataExchangeEvents(context.Background(), filter)
		require.NoError(t, err)
		assert.Equal(t, 0, len(response.Events))
		assert.Equal(t, int64(0), response.Total)
	})
}
