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

func TestPolicyEvaluator_Authorize(t *testing.T) {
	ctx := context.Background()
	evaluator, err := NewPolicyEvaluator(ctx)
	if err != nil {
		t.Fatalf("Failed to create policy evaluator: %v", err)
	}

	tests := []struct {
		name     string
		request  map[string]interface{}
		expected bool
	}{
		{
			name: "Valid request - no consent required",
			request: map[string]interface{}{
				"consumer": map[string]interface{}{
					"id": "passport-app",
				},
				"request": map[string]interface{}{
					"resource":    "person_data",
					"action":      "read",
					"data_fields": []string{"person.fullName", "person.nic"},
				},
			},
			expected: true,
		},
		{
			name: "Request for unapproved field - should be denied",
			request: map[string]interface{}{
				"consumer": map[string]interface{}{
					"id": "passport-app",
				},
				"request": map[string]interface{}{
					"resource":    "person_data",
					"action":      "read",
					"data_fields": []string{"person.permanentAddress"},
				},
			},
			expected: false,
		},
		{
			name: "Invalid consumer",
			request: map[string]interface{}{
				"consumer": map[string]interface{}{
					"id": "unknown-app",
				},
				"request": map[string]interface{}{
					"resource":    "person_data",
					"action":      "read",
					"data_fields": []string{"person.fullName"},
				},
			},
			expected: false,
		},
		{
			name: "Unauthorized field access",
			request: map[string]interface{}{
				"consumer": map[string]interface{}{
					"id": "passport-app",
				},
				"request": map[string]interface{}{
					"resource":    "person_data",
					"action":      "read",
					"data_fields": []string{"person.ssn"},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, err := evaluator.Authorize(ctx, tt.request)

			if err != nil {
				t.Fatalf("Authorize failed: %v", err)
			}

			if decision.Allow != tt.expected {
				t.Errorf("Expected allow=%v, got %v", tt.expected, decision.Allow)
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

	// Test that requesting unapproved fields results in denial (not consent requirement)
	request := map[string]interface{}{
		"consumer": map[string]interface{}{
			"id": "passport-app",
		},
		"request": map[string]interface{}{
			"resource":    "person_data",
			"action":      "read",
			"data_fields": []string{"person.fullName", "person.permanentAddress"},
		},
	}

	decision, err := evaluator.Authorize(ctx, request)
	if err != nil {
		t.Fatalf("Authorize failed: %v", err)
	}

	// Since person.permanentAddress is not in approved fields, the request should be denied
	// and consent_required should be false (because it's denied outright)
	if decision.Allow {
		t.Error("Expected allow=false for unapproved field")
	}

	if decision.ConsentRequired {
		t.Error("Expected consent_required=false for unapproved field (should be denied outright)")
	}

	if len(decision.ConsentRequiredFields) != 0 {
		t.Errorf("Expected empty consent_required_fields for unapproved field, got %v", decision.ConsentRequiredFields)
	}
}

func TestPolicyEvaluator_ConsentFlow(t *testing.T) {
	ctx := context.Background()
	evaluator, err := NewPolicyEvaluator(ctx)
	if err != nil {
		t.Fatalf("Failed to create policy evaluator: %v", err)
	}

	// This test demonstrates the consent flow logic:
	// 1. Consumer requests approved fields that don't require consent -> allow=true, consent_required=false
	// 2. Consumer requests approved fields that require consent -> allow=false, consent_required=true
	// 3. Consumer requests unapproved fields -> allow=false, consent_required=false

	tests := []struct {
		name                    string
		request                 map[string]interface{}
		expectedAllow           bool
		expectedConsentRequired bool
		expectedConsentFields   []string
	}{
		{
			name: "Approved fields, no consent required",
			request: map[string]interface{}{
				"consumer": map[string]interface{}{
					"id": "passport-app",
				},
				"request": map[string]interface{}{
					"resource":    "person_data",
					"action":      "read",
					"data_fields": []string{"person.fullName", "person.nic"},
				},
			},
			expectedAllow:           true,
			expectedConsentRequired: false,
			expectedConsentFields:   []string{},
		},
		{
			name: "Unapproved fields (should be denied outright)",
			request: map[string]interface{}{
				"consumer": map[string]interface{}{
					"id": "passport-app",
				},
				"request": map[string]interface{}{
					"resource":    "person_data",
					"action":      "read",
					"data_fields": []string{"person.permanentAddress"},
				},
			},
			expectedAllow:           false,
			expectedConsentRequired: false,
			expectedConsentFields:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, err := evaluator.Authorize(ctx, tt.request)
			if err != nil {
				t.Fatalf("Authorize failed: %v", err)
			}

			if decision.Allow != tt.expectedAllow {
				t.Errorf("Expected allow=%v, got %v", tt.expectedAllow, decision.Allow)
			}

			if decision.ConsentRequired != tt.expectedConsentRequired {
				t.Errorf("Expected consent_required=%v, got %v", tt.expectedConsentRequired, decision.ConsentRequired)
			}

			if len(decision.ConsentRequiredFields) != len(tt.expectedConsentFields) {
				t.Errorf("Expected consent_required_fields length=%d, got %d", len(tt.expectedConsentFields), len(decision.ConsentRequiredFields))
			}
		})
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
