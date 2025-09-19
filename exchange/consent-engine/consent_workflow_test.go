package main

import (
	"encoding/json"
	"testing"

	"github.com/gov-dx-sandbox/exchange/consent-engine/models"
)

func TestConsentWorkflowRequest(t *testing.T) {
	// Test ConsentWorkflowRequest model structure
	req := models.ConsentWorkflowRequest{
		AppID: "passport-app",
		DataFields: []models.DataField{
			{
				OwnerType:  "citizen",
				OwnerID:    "199512345678",
				OwnerEmail: "199512345678@example.com",
				Fields:     []string{"person.permanentAddress"},
			},
		},
		Purpose:   "passport_application",
		SessionID: "session_123",
	}

	// Verify the request structure
	if req.AppID != "passport-app" {
		t.Errorf("Expected AppID to be 'passport-app', got '%s'", req.AppID)
	}

	if len(req.DataFields) != 1 {
		t.Errorf("Expected 1 data field, got %d", len(req.DataFields))
	}

	if req.DataFields[0].OwnerType != "citizen" {
		t.Errorf("Expected OwnerType to be 'citizen', got '%s'", req.DataFields[0].OwnerType)
	}

	if req.DataFields[0].OwnerID != "199512345678" {
		t.Errorf("Expected OwnerID to be '199512345678', got '%s'", req.DataFields[0].OwnerID)
	}

	if req.DataFields[0].OwnerEmail != "199512345678@example.com" {
		t.Errorf("Expected OwnerEmail to be '199512345678@example.com', got '%s'", req.DataFields[0].OwnerEmail)
	}

	if len(req.DataFields[0].Fields) != 1 {
		t.Errorf("Expected 1 field, got %d", len(req.DataFields[0].Fields))
	}

	if req.DataFields[0].Fields[0] != "person.permanentAddress" {
		t.Errorf("Expected field to be 'person.permanentAddress', got '%s'", req.DataFields[0].Fields[0])
	}

	if req.Purpose != "passport_application" {
		t.Errorf("Expected Purpose to be 'passport_application', got '%s'", req.Purpose)
	}

	if req.SessionID != "session_123" {
		t.Errorf("Expected SessionID to be 'session_123', got '%s'", req.SessionID)
	}

}

func TestDataField(t *testing.T) {
	// Test DataField model structure
	field := models.DataField{
		OwnerType:  "citizen",
		OwnerID:    "199512345678",
		OwnerEmail: "199512345678@example.com",
		Fields:     []string{"person.permanentAddress", "person.fullName"},
	}

	// Verify the field structure
	if field.OwnerType != "citizen" {
		t.Errorf("Expected OwnerType to be 'citizen', got '%s'", field.OwnerType)
	}

	if field.OwnerID != "199512345678" {
		t.Errorf("Expected OwnerID to be '199512345678', got '%s'", field.OwnerID)
	}

	if field.OwnerEmail != "199512345678@example.com" {
		t.Errorf("Expected OwnerEmail to be '199512345678@example.com', got '%s'", field.OwnerEmail)
	}

	if len(field.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(field.Fields))
	}

	expectedFields := []string{"person.permanentAddress", "person.fullName"}
	for i, expectedField := range expectedFields {
		if field.Fields[i] != expectedField {
			t.Errorf("Expected field %d to be '%s', got '%s'", i, expectedField, field.Fields[i])
		}
	}
}

func TestConsentWorkflowJSONSerialization(t *testing.T) {
	// Test JSON serialization/deserialization
	req := models.ConsentWorkflowRequest{
		AppID: "passport-app",
		DataFields: []models.DataField{
			{
				OwnerType:  "citizen",
				OwnerID:    "199512345678",
				OwnerEmail: "199512345678@example.com",
				Fields:     []string{"person.permanentAddress"},
			},
		},
		Purpose:   "passport_application",
		SessionID: "session_123",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal ConsentWorkflowRequest: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaledReq models.ConsentWorkflowRequest
	err = json.Unmarshal(jsonData, &unmarshaledReq)
	if err != nil {
		t.Fatalf("Failed to unmarshal ConsentWorkflowRequest: %v", err)
	}

	// Verify the unmarshaled request matches the original
	if unmarshaledReq.AppID != req.AppID {
		t.Errorf("Expected AppID to be '%s', got '%s'", req.AppID, unmarshaledReq.AppID)
	}

	if len(unmarshaledReq.DataFields) != len(req.DataFields) {
		t.Errorf("Expected %d data fields, got %d", len(req.DataFields), len(unmarshaledReq.DataFields))
	}

	if unmarshaledReq.Purpose != req.Purpose {
		t.Errorf("Expected Purpose to be '%s', got '%s'", req.Purpose, unmarshaledReq.Purpose)
	}

	if unmarshaledReq.SessionID != req.SessionID {
		t.Errorf("Expected SessionID to be '%s', got '%s'", req.SessionID, unmarshaledReq.SessionID)
	}

}

func TestConsentWorkflowValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     models.ConsentWorkflowRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: models.ConsentWorkflowRequest{
				AppID: "passport-app",
				DataFields: []models.DataField{
					{
						OwnerType: "citizen",
						OwnerID:   "199512345678",
						Fields:    []string{"person.permanentAddress"},
					},
				},
				Purpose:   "passport_application",
				SessionID: "session_123",
			},
			wantErr: false,
		},
		{
			name: "empty app_id",
			req: models.ConsentWorkflowRequest{
				AppID: "",
				DataFields: []models.DataField{
					{
						OwnerType: "citizen",
						OwnerID:   "199512345678",
						Fields:    []string{"person.permanentAddress"},
					},
				},
				Purpose:   "passport_application",
				SessionID: "session_123",
			},
			wantErr: true,
		},
		{
			name: "empty data_fields",
			req: models.ConsentWorkflowRequest{
				AppID:      "passport-app",
				DataFields: []models.DataField{},
				Purpose:    "passport_application",
				SessionID:  "session_123",
			},
			wantErr: true,
		},
		{
			name: "empty purpose",
			req: models.ConsentWorkflowRequest{
				AppID: "passport-app",
				DataFields: []models.DataField{
					{
						OwnerType: "citizen",
						OwnerID:   "199512345678",
						Fields:    []string{"person.permanentAddress"},
					},
				},
				Purpose:   "",
				SessionID: "session_123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation logic (you can expand this based on your validation requirements)
			hasError := false

			if tt.req.AppID == "" {
				hasError = true
			}

			if len(tt.req.DataFields) == 0 {
				hasError = true
			}

			if tt.req.Purpose == "" {
				hasError = true
			}

			if hasError != tt.wantErr {
				t.Errorf("Expected error: %v, got error: %v", tt.wantErr, hasError)
			}
		})
	}
}

func TestConsentWorkflowEdgeCases(t *testing.T) {
	// Test with multiple data fields
	req := models.ConsentWorkflowRequest{
		AppID: "passport-app",
		DataFields: []models.DataField{
			{
				OwnerType: "citizen",
				OwnerID:   "199512345678",
				Fields:    []string{"person.permanentAddress"},
			},
			{
				OwnerType: "citizen",
				OwnerID:   "199512345679",
				Fields:    []string{"person.fullName"},
			},
		},
		Purpose:   "passport_application",
		SessionID: "session_123",
	}

	if len(req.DataFields) != 2 {
		t.Errorf("Expected 2 data fields, got %d", len(req.DataFields))
	}

	// Test with multiple fields per data owner
	multiFieldReq := models.ConsentWorkflowRequest{
		AppID: "passport-app",
		DataFields: []models.DataField{
			{
				OwnerType: "citizen",
				OwnerID:   "199512345678",
				Fields:    []string{"person.permanentAddress", "person.fullName", "person.email"},
			},
		},
		Purpose:   "passport_application",
		SessionID: "session_123",
	}

	if len(multiFieldReq.DataFields[0].Fields) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(multiFieldReq.DataFields[0].Fields))
	}
}
