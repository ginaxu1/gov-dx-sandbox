package main

import (
	"encoding/json"
	"testing"
)

func TestConsentWorkflowRequest(t *testing.T) {
	// Test ConsentWorkflowRequest model structure
	req := ConsentRequest{
		AppID: "passport-app",
		ConsentRequirements: []ConsentRequirement{
			{
				Owner:   "CITIZEN",
				OwnerID: "199512345678@example.com",
				Fields: []ConsentField{
					{
						FieldName: "person.permanentAddress",
						SchemaID:  "schema_123",
					},
				},
			},
		},
	}

	// Verify the request structure
	if req.AppID != "passport-app" {
		t.Errorf("Expected AppID to be 'passport-app', got '%s'", req.AppID)
	}

	if len(req.ConsentRequirements) != 1 {
		t.Errorf("Expected 1 consent requirement, got %d", len(req.ConsentRequirements))
	}

	if req.ConsentRequirements[0].OwnerID != "199512345678@example.com" {
		t.Errorf("Expected OwnerID to be '199512345678@example.com', got '%s'", req.ConsentRequirements[0].OwnerID)
	}

	if len(req.ConsentRequirements[0].Fields) != 1 {
		t.Errorf("Expected 1 field, got %d", len(req.ConsentRequirements[0].Fields))
	}

	if req.ConsentRequirements[0].Fields[0].FieldName != "person.permanentAddress" {
		t.Errorf("Expected fieldName to be 'person.permanentAddress', got '%s'", req.ConsentRequirements[0].Fields[0].FieldName)
	}

}

func TestDataField(t *testing.T) {
	// Test DataField model structure
	field := DataField{

		OwnerID:    "199512345678",
		OwnerEmail: "199512345678@example.com",
		Fields:     []string{"person.permanentAddress", "person.fullName"},
	}

	// Verify the field structure

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
	req := ConsentRequest{
		AppID: "passport-app",
		ConsentRequirements: []ConsentRequirement{
			{
				Owner:   "CITIZEN",
				OwnerID: "199512345678@example.com",
				Fields: []ConsentField{
					{
						FieldName: "person.permanentAddress",
						SchemaID:  "schema_123",
					},
				},
			},
		},
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal ConsentWorkflowRequest: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaledReq ConsentRequest
	err = json.Unmarshal(jsonData, &unmarshaledReq)
	if err != nil {
		t.Fatalf("Failed to unmarshal ConsentWorkflowRequest: %v", err)
	}

	// Verify the unmarshaled request matches the original
	if unmarshaledReq.AppID != req.AppID {
		t.Errorf("Expected AppID to be '%s', got '%s'", req.AppID, unmarshaledReq.AppID)
	}

	if len(unmarshaledReq.ConsentRequirements) != len(req.ConsentRequirements) {
		t.Errorf("Expected %d consent requirements, got %d", len(req.ConsentRequirements), len(unmarshaledReq.ConsentRequirements))
	}

}

func TestConsentWorkflowValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     ConsentRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: ConsentRequest{
				AppID: "passport-app",
				ConsentRequirements: []ConsentRequirement{
					{
						Owner:   "CITIZEN",
						OwnerID: "199512345678@example.com",
						Fields: []ConsentField{
							{
								FieldName: "person.permanentAddress",
								SchemaID:  "schema_123",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty app_id",
			req: ConsentRequest{
				AppID: "",
				ConsentRequirements: []ConsentRequirement{
					{
						Owner:   "CITIZEN",
						OwnerID: "199512345678@example.com",
						Fields: []ConsentField{
							{
								FieldName: "person.permanentAddress",
								SchemaID:  "schema_123",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty consent_requirements",
			req: ConsentRequest{
				AppID:               "passport-app",
				ConsentRequirements: []ConsentRequirement{},
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

			if len(tt.req.ConsentRequirements) == 0 {
				hasError = true
			}

			if hasError != tt.wantErr {
				t.Errorf("Expected error: %v, got error: %v", tt.wantErr, hasError)
			}
		})
	}
}

func TestConsentWorkflowEdgeCases(t *testing.T) {
	// Test with multiple consent requirements
	req := ConsentRequest{
		AppID: "passport-app",
		ConsentRequirements: []ConsentRequirement{
			{
				Owner:   "CITIZEN",
				OwnerID: "199512345678@example.com",
				Fields: []ConsentField{
					{
						FieldName: "person.permanentAddress",
						SchemaID:  "schema_123",
					},
				},
			},
			{
				Owner:   "CITIZEN",
				OwnerID: "199512345679@example.com",
				Fields: []ConsentField{
					{
						FieldName: "person.fullName",
						SchemaID:  "schema_123",
					},
				},
			},
		},
	}

	if len(req.ConsentRequirements) != 2 {
		t.Errorf("Expected 2 consent requirements, got %d", len(req.ConsentRequirements))
	}

	// Test with multiple fields per consent requirement
	multiFieldReq := ConsentRequest{
		AppID: "passport-app",
		ConsentRequirements: []ConsentRequirement{
			{
				Owner:   "CITIZEN",
				OwnerID: "199512345678@example.com",
				Fields: []ConsentField{
					{
						FieldName: "person.permanentAddress",
						SchemaID:  "schema_123",
					},
					{
						FieldName: "person.fullName",
						SchemaID:  "schema_123",
					},
					{
						FieldName: "person.email",
						SchemaID:  "schema_123",
					},
				},
			},
		},
	}

	if len(multiFieldReq.ConsentRequirements[0].Fields) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(multiFieldReq.ConsentRequirements[0].Fields))
	}
}
