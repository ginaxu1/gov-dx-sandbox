package tests

import (
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/models"
)

func TestGrantsService_GetAllConsumerGrants(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	service := ts.APIServer.GetGrantsService()

	// Create some test data
	consumerService := ts.APIServer.GetConsumerService()

	// Create consumers
	consumer1, err := consumerService.CreateConsumer(models.CreateConsumerRequest{
		ConsumerName: "Grant Consumer 1",
		ContactEmail: "grant1@example.com",
		PhoneNumber:  "1111111111",
	})
	if err != nil {
		t.Fatalf("Failed to create consumer: %v", err)
	}

	consumer2, err := consumerService.CreateConsumer(models.CreateConsumerRequest{
		ConsumerName: "Grant Consumer 2",
		ContactEmail: "grant2@example.com",
		PhoneNumber:  "2222222222",
	})
	if err != nil {
		t.Fatalf("Failed to create consumer: %v", err)
	}

	// Create consumer grants directly in the database
	// Note: In a real scenario, grants would be created through a proper API
	// For testing, we'll insert them directly
	_, err = ts.DB.Exec(`
		INSERT INTO consumer_grants (consumer_id, approved_fields, created_at, updated_at) 
		VALUES ($1, $2, NOW(), NOW())
	`, consumer1.ConsumerID, `["person.fullName", "person.nic"]`)
	if err != nil {
		t.Fatalf("Failed to create consumer grant: %v", err)
	}

	_, err = ts.DB.Exec(`
		INSERT INTO consumer_grants (consumer_id, approved_fields, created_at, updated_at) 
		VALUES ($1, $2, NOW(), NOW())
	`, consumer2.ConsumerID, `["person.address"]`)
	if err != nil {
		t.Fatalf("Failed to create consumer grant: %v", err)
	}

	// Get all consumer grants
	grants, err := service.GetAllConsumerGrants()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if grants == nil {
		t.Error("Expected grants data to be returned")
	}

	if len(grants.ConsumerGrants) != 2 {
		t.Errorf("Expected 2 consumer grants, got %d", len(grants.ConsumerGrants))
	}
}

func TestGrantsService_GetConsumerGrant(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	service := ts.APIServer.GetGrantsService()

	// Create a consumer
	consumerService := ts.APIServer.GetConsumerService()
	consumer, err := consumerService.CreateConsumer(models.CreateConsumerRequest{
		ConsumerName: "Grant Consumer",
		ContactEmail: "grant@example.com",
		PhoneNumber:  "1234567890",
	})
	if err != nil {
		t.Fatalf("Failed to create consumer: %v", err)
	}

	// Create a consumer grant
	_, err = ts.DB.Exec(`
		INSERT INTO consumer_grants (consumer_id, approved_fields, created_at, updated_at) 
		VALUES ($1, $2, NOW(), NOW())
	`, consumer.ConsumerID, `["person.fullName", "person.nic"]`)
	if err != nil {
		t.Fatalf("Failed to create consumer grant: %v", err)
	}

	// Get the consumer grant
	grant, err := service.GetConsumerGrant(consumer.ConsumerID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if grant.ConsumerID != consumer.ConsumerID {
		t.Errorf("Expected ConsumerID %s, got %s", consumer.ConsumerID, grant.ConsumerID)
	}

	if len(grant.ApprovedFields) == 0 {
		t.Error("Expected approved fields to be populated")
	}
}

func TestGrantsService_GetConsumerGrant_NotFound(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	service := ts.APIServer.GetGrantsService()

	_, err := service.GetConsumerGrant("non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent consumer grant")
	}
}

func TestGrantsService_GetAllProviderFields(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	service := ts.APIServer.GetGrantsService()

	// Create some test provider metadata
	_, err := ts.DB.Exec(`
		INSERT INTO provider_metadata (field_name, owner, provider, consent_required, access_control_type, allow_list, description, metadata, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	`, "person.fullName", "government", "passport-service", true, "restricted", `[{"type": "government", "value": "passport"}]`, "Full name field", `{"category": "personal"}`)
	if err != nil {
		t.Fatalf("Failed to create provider metadata: %v", err)
	}

	_, err = ts.DB.Exec(`
		INSERT INTO provider_metadata (field_name, owner, provider, consent_required, access_control_type, allow_list, description, metadata, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	`, "person.nic", "government", "passport-service", true, "restricted", `[{"type": "government", "value": "passport"}]`, "NIC field", `{"category": "personal"}`)
	if err != nil {
		t.Fatalf("Failed to create provider metadata: %v", err)
	}

	// Get all provider fields
	fields, err := service.GetAllProviderFields()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if fields == nil {
		t.Error("Expected fields data to be returned")
	}

	if len(fields.Fields) != 2 {
		t.Errorf("Expected 2 provider fields, got %d", len(fields.Fields))
	}

	// Check specific field
	if field, exists := fields.Fields["person.fullName"]; exists {
		if field.Owner != "government" {
			t.Errorf("Expected owner 'government', got %s", field.Owner)
		}
		if !field.ConsentRequired {
			t.Error("Expected consent to be required")
		}
	} else {
		t.Error("Expected 'person.fullName' field to exist")
	}
}

func TestGrantsService_GetProviderField(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	service := ts.APIServer.GetGrantsService()

	// Create test provider metadata
	_, err := ts.DB.Exec(`
		INSERT INTO provider_metadata (field_name, owner, provider, consent_required, access_control_type, allow_list, description, metadata, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	`, "person.fullName", "government", "passport-service", true, "restricted", `[{"type": "government", "value": "passport"}]`, "Full name field", `{"category": "personal"}`)
	if err != nil {
		t.Fatalf("Failed to create provider metadata: %v", err)
	}

	// Get the provider field
	field, err := service.GetProviderField("person.fullName")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if field.FieldName != "person.fullName" {
		t.Errorf("Expected FieldName 'person.fullName', got %s", field.FieldName)
	}
	if field.Owner != "government" {
		t.Errorf("Expected Owner 'government', got %s", field.Owner)
	}
}

func TestGrantsService_GetProviderField_NotFound(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	service := ts.APIServer.GetGrantsService()

	_, err := service.GetProviderField("non-existent-field")
	if err == nil {
		t.Error("Expected error for non-existent provider field")
	}
}
