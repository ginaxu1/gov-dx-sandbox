package main

import (
	"context"
	"testing"
)

// Helper function to get field names from fields map
func getFieldNames(fields map[string]interface{}) []string {
	var names []string
	for name := range fields {
		names = append(names, name)
	}
	return names
}

// TestNewPolicyEvaluator tests the creation of a policy evaluator
func TestNewPolicyEvaluator(t *testing.T) {
	ctx := context.Background()
	evaluator, err := NewMockPolicyEvaluator(ctx)
	if err != nil {
		t.Fatalf("Failed to create policy evaluator: %v", err)
	}

	if evaluator == nil {
		t.Fatal("Expected non-nil policy evaluator")
	}
}

// TestPolicyEvaluator_Authorize tests the main authorization functionality
func TestPolicyEvaluator_Authorize(t *testing.T) {
	ctx := context.Background()
	evaluator, err := NewMockPolicyEvaluator(ctx)
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
			name: "Public field with allow list - authorized consumer",
			request: map[string]interface{}{
				"consumer_id":     "test-app",
				"app_id":          "test-app",
				"request_id":      "req_public",
				"required_fields": []string{"person.fullName"},
			},
			expected:        true,
			consentRequired: false,
		},
		{
			name: "Restricted field with allow list - authorized consumer",
			request: map[string]interface{}{
				"consumer_id":     "test-app",
				"app_id":          "test-app",
				"request_id":      "req_restricted",
				"required_fields": []string{"person.birthDate"},
			},
			expected:        true,
			consentRequired: true,
		},
		{
			name: "Field requiring consent - authorized consumer",
			request: map[string]interface{}{
				"consumer_id":     "test-app",
				"app_id":          "test-app",
				"request_id":      "req_consent",
				"required_fields": []string{"person.nic"},
			},
			expected:        true,
			consentRequired: false,
		},
		{
			name: "Public field with no allow list - any consumer",
			request: map[string]interface{}{
				"consumer_id":     "any-app",
				"app_id":          "any-app",
				"request_id":      "req_public_no_list",
				"required_fields": []string{"public.field"},
			},
			expected:        true,
			consentRequired: false,
		},
		{
			name: "Unauthorized consumer",
			request: map[string]interface{}{
				"consumer_id":     "unauthorized-app",
				"app_id":          "unauthorized-app",
				"request_id":      "req_unauth",
				"required_fields": []string{"person.fullName"},
			},
			expected:        false,
			consentRequired: false,
		},
		{
			name: "Missing consumer_id",
			request: map[string]interface{}{
				"app_id":          "passport-app",
				"request_id":      "req_missing",
				"required_fields": []string{"person.fullName"},
			},
			expected:        false,
			consentRequired: false,
		},
		{
			name: "Empty required_fields",
			request: map[string]interface{}{
				"consumer_id":     "passport-app",
				"app_id":          "passport-app",
				"request_id":      "req_empty",
				"required_fields": []string{},
			},
			expected:        true,
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
		})
	}
}
