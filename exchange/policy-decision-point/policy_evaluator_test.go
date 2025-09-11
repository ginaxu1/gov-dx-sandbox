package main

import (
	"context"
	"testing"
)

func TestNewPolicyEvaluator(t *testing.T) {
	ctx := context.Background()
	evaluator, err := NewPolicyEvaluator(ctx)
	if err != nil {
		t.Fatalf("Failed to create policy evaluator: %v", err)
	}

	if evaluator == nil {
		t.Fatal("Expected non-nil policy evaluator")
	}
}

func TestPolicyEvaluator_Authorize_NewFormat(t *testing.T) {
	ctx := context.Background()
	evaluator, err := NewPolicyEvaluator(ctx)
	if err != nil {
		t.Fatalf("Failed to create policy evaluator: %v", err)
	}

	tests := []struct {
		name            string
		request         map[string]interface{}
		expected        bool
		consentRequired bool
	}{
		{
			name: "New format - passport-app with fullName and permanentAddress",
			request: map[string]interface{}{
				"consumer_id":     "passport-app",
				"required_fields": []string{"person.fullName", "person.permanentAddress"},
			},
			expected:        true,
			consentRequired: false,
		},
		{
			name: "New format - driver-app with fullName and birthDate",
			request: map[string]interface{}{
				"consumer_id":     "driver-app",
				"required_fields": []string{"person.fullName", "person.birthDate"},
			},
			expected:        true,
			consentRequired: false,
		},
		{
			name: "New format - passport-app with photo (requires consent)",
			request: map[string]interface{}{
				"consumer_id":     "passport-app",
				"required_fields": []string{"person.photo"},
			},
			expected:        true,
			consentRequired: true,
		},
		{
			name: "New format - unauthorized consumer",
			request: map[string]interface{}{
				"consumer_id":     "unauthorized-app",
				"required_fields": []string{"person.permanentAddress"},
			},
			expected:        false,
			consentRequired: false,
		},
		{
			name: "New format - missing consumer_id",
			request: map[string]interface{}{
				"required_fields": []string{"person.fullName"},
			},
			expected:        false,
			consentRequired: false,
		},
		{
			name: "New format - empty required_fields",
			request: map[string]interface{}{
				"consumer_id":     "passport-app",
				"required_fields": []string{},
			},
			expected:        false,
			consentRequired: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, err := evaluator.Authorize(ctx, tt.request)
			if err != nil {
				t.Fatalf("Authorize() error = %v", err)
			}

			if decision.Allow != tt.expected {
				t.Errorf("Authorize() Allow = %v, expected %v", decision.Allow, tt.expected)
			}

			if decision.ConsentRequired != tt.consentRequired {
				t.Errorf("Authorize() ConsentRequired = %v, expected %v", decision.ConsentRequired, tt.consentRequired)
			}

			if !tt.expected && decision.DenyReason == "" {
				t.Error("Expected DenyReason to be set when access is denied")
			}
		})
	}
}

func TestPolicyEvaluator_ConsentRequired(t *testing.T) {
	ctx := context.Background()
	evaluator, err := NewPolicyEvaluator(ctx)
	if err != nil {
		t.Fatalf("Failed to create policy evaluator: %v", err)
	}

	// Test that requesting unauthorized fields results in denial (not consent requirement)
	request := map[string]interface{}{
		"consumer": map[string]interface{}{
			"id": "passport-app",
		},
		"request": map[string]interface{}{
			"resource":    "person_data",
			"action":      "read",
			"data_fields": []string{"person.birthDate"},
		},
	}

	decision, err := evaluator.Authorize(ctx, request)
	if err != nil {
		t.Fatalf("Authorize failed: %v", err)
	}

	// Since person.birthDate is not in passport-app's allow list, the request should be denied
	// and consent_required should be false (because it's denied outright)
	if decision.Allow {
		t.Error("Expected allow=false for unauthorized field")
	}

	if decision.ConsentRequired {
		t.Error("Expected consent_required=false for unauthorized field (should be denied outright)")
	}

	if len(decision.ConsentRequiredFields) != 0 {
		t.Errorf("Expected empty consent_required_fields for unapproved field, got %v", decision.ConsentRequiredFields)
	}
}

func TestPolicyEvaluator_DebugData(t *testing.T) {
	ctx := context.Background()
	evaluator, err := NewPolicyEvaluator(ctx)
	if err != nil {
		t.Fatalf("Failed to create policy evaluator: %v", err)
	}

	debugData, err := evaluator.DebugData(ctx)
	if err != nil {
		t.Fatalf("DebugData failed: %v", err)
	}

	if debugData == nil {
		t.Error("Expected non-nil debug data")
	}
}
