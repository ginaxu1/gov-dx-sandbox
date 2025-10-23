package main

import (
	"context"
	"testing"
	"time"

	"github.com/gov-dx-sandbox/exchange/policy-decision-point/models"
)

func TestConsentLogic(t *testing.T) {
	ctx := context.Background()
	evaluator, err := NewMockPolicyEvaluator(ctx)
	if err != nil {
		t.Fatalf("Failed to create mock policy evaluator: %v", err)
	}

	testCases := []struct {
		name                  string
		request               models.PolicyDecisionRequest
		expectedAllow         bool
		expectedConsent       bool
		expectedConsentFields []string
	}{
		{
			name: "Owner with public field - no consent required",
			request: models.PolicyDecisionRequest{
				ApplicationID:  "test-app",
				AppID:          "test-app",
				RequestID:      "req-1",
				RequiredFields: []string{"person.fullName"},
			},
			expectedAllow:         true,
			expectedConsent:       false,
			expectedConsentFields: []string{},
		},
		{
			name: "Owner with restricted field - no consent required",
			request: models.PolicyDecisionRequest{
				ApplicationID:  "test-app",
				AppID:          "test-app",
				RequestID:      "req-2",
				RequiredFields: []string{"person.nic"},
			},
			expectedAllow:         true,
			expectedConsent:       false,
			expectedConsentFields: []string{},
		},
		{
			name: "Non-owner with restricted field - consent required",
			request: models.PolicyDecisionRequest{
				ApplicationID:  "test-app",
				AppID:          "test-app",
				RequestID:      "req-3",
				RequiredFields: []string{"person.birthDate"},
			},
			expectedAllow:         true,
			expectedConsent:       true,
			expectedConsentFields: []string{"person.birthDate"},
		},
		{
			name: "Non-owner with public field - no consent required",
			request: models.PolicyDecisionRequest{
				ApplicationID:  "test-app",
				AppID:          "test-app",
				RequestID:      "req-4",
				RequiredFields: []string{"public.field"},
			},
			expectedAllow:         true,
			expectedConsent:       false,
			expectedConsentFields: []string{},
		},
		{
			name: "Mixed fields - some require consent",
			request: models.PolicyDecisionRequest{
				ApplicationID:  "test-app",
				AppID:          "test-app",
				RequestID:      "req-5",
				RequiredFields: []string{"person.fullName", "person.birthDate", "public.field"},
			},
			expectedAllow:         true,
			expectedConsent:       true,
			expectedConsentFields: []string{"person.birthDate"},
		},
		{
			name: "Unauthorized app - access denied",
			request: models.PolicyDecisionRequest{
				ApplicationID:  "unauthorized-app",
				AppID:          "unauthorized-app",
				RequestID:      "req-6",
				RequiredFields: []string{"person.fullName"},
			},
			expectedAllow:         false,
			expectedConsent:       false,
			expectedConsentFields: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decision, err := evaluator.Authorize(ctx, tc.request)
			if err != nil {
				t.Fatalf("Policy evaluation failed: %v", err)
			}

			if decision.Allow != tc.expectedAllow {
				t.Errorf("Expected allow=%v, got %v", tc.expectedAllow, decision.Allow)
			}

			if decision.ConsentRequired != tc.expectedConsent {
				t.Errorf("Expected consent_required=%v, got %v", tc.expectedConsent, decision.ConsentRequired)
			}

			if len(decision.ConsentRequiredFields) != len(tc.expectedConsentFields) {
				t.Errorf("Expected %d consent fields, got %d", len(tc.expectedConsentFields), len(decision.ConsentRequiredFields))
			}

			// Check if all expected consent fields are present
			for _, expectedField := range tc.expectedConsentFields {
				found := false
				for _, actualField := range decision.ConsentRequiredFields {
					if actualField == expectedField {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected consent field %s not found in %v", expectedField, decision.ConsentRequiredFields)
				}
			}
		})
	}
}

func TestConsentLogicEdgeCases(t *testing.T) {
	ctx := context.Background()
	evaluator, err := NewMockPolicyEvaluator(ctx)
	if err != nil {
		t.Fatalf("Failed to create mock policy evaluator: %v", err)
	}

	// Test with empty fields
	t.Run("Empty required fields", func(t *testing.T) {
		request := models.PolicyDecisionRequest{
			ApplicationID:  "test-app",
			AppID:          "test-app",
			RequestID:      "req-empty",
			RequiredFields: []string{},
		}

		decision, err := evaluator.Authorize(ctx, request)
		if err != nil {
			t.Fatalf("Policy evaluation failed: %v", err)
		}

		// Should allow access to empty field list
		if !decision.Allow {
			t.Errorf("Expected allow=true for empty fields, got %v", decision.Allow)
		}
	})

	// Test with non-existent field
	t.Run("Non-existent field", func(t *testing.T) {
		request := models.PolicyDecisionRequest{
			ApplicationID:  "test-app",
			AppID:          "test-app",
			RequestID:      "req-nonexistent",
			RequiredFields: []string{"nonexistent.field"},
		}

		decision, err := evaluator.Authorize(ctx, request)
		if err != nil {
			t.Fatalf("Policy evaluation failed: %v", err)
		}

		// Should deny access to non-existent field
		if decision.Allow {
			t.Errorf("Expected allow=false for non-existent field, got %v", decision.Allow)
		}
	})
}

func TestConsentLogicWithRealData(t *testing.T) {
	ctx := context.Background()

	// Create a real database service for integration testing
	dbService := NewMockDatabaseService()

	// Add some test policy metadata
	_, err := dbService.CreatePolicyMetadata(&models.PolicyMetadataCreateRequest{
		FieldName:         "test.ownerField",
		DisplayName:       "Owner Field",
		Description:       "A field owned by the data owner",
		Source:            "test_system",
		IsOwner:           true,
		AccessControlType: "restricted",
		AllowList:         []models.AllowListEntry{},
	})
	if err != nil {
		t.Fatalf("Failed to create test policy metadata: %v", err)
	}

	_, err = dbService.CreatePolicyMetadata(&models.PolicyMetadataCreateRequest{
		FieldName:         "test.nonOwnerField",
		DisplayName:       "Non-Owner Field",
		Description:       "A field not owned by the data owner",
		Source:            "test_system",
		IsOwner:           false,
		AccessControlType: "restricted",
		AllowList:         []models.AllowListEntry{},
	})
	if err != nil {
		t.Fatalf("Failed to create test policy metadata: %v", err)
	}

	// Update allow list for both fields
	err = dbService.UpdateAllowList(&models.AllowListUpdateRequest{
		FieldName:     "test.ownerField",
		ApplicationID: "test-app",
		ExpiresAt:     time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("Failed to update allow list: %v", err)
	}

	err = dbService.UpdateAllowList(&models.AllowListUpdateRequest{
		FieldName:     "test.nonOwnerField",
		ApplicationID: "test-app",
		ExpiresAt:     time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("Failed to update allow list: %v", err)
	}

	// Create policy evaluator with real data
	evaluator, err := createPolicyEvaluatorWithData(ctx, dbService)
	if err != nil {
		t.Fatalf("Failed to create policy evaluator: %v", err)
	}

	// Test owner field - should not require consent
	t.Run("Owner field - no consent required", func(t *testing.T) {
		request := models.PolicyDecisionRequest{
			ApplicationID:  "test-app",
			AppID:          "test-app",
			RequestID:      "req-owner",
			RequiredFields: []string{"test.ownerField"},
		}

		decision, err := evaluator.Authorize(ctx, request)
		if err != nil {
			t.Fatalf("Policy evaluation failed: %v", err)
		}

		if !decision.Allow {
			t.Errorf("Expected allow=true, got %v", decision.Allow)
		}

		if decision.ConsentRequired {
			t.Errorf("Expected consent_required=false for owner field, got %v", decision.ConsentRequired)
		}
	})

	// Test non-owner field - should require consent
	t.Run("Non-owner field - consent required", func(t *testing.T) {
		request := models.PolicyDecisionRequest{
			ApplicationID:  "test-app",
			AppID:          "test-app",
			RequestID:      "req-nonowner",
			RequiredFields: []string{"test.nonOwnerField"},
		}

		decision, err := evaluator.Authorize(ctx, request)
		if err != nil {
			t.Fatalf("Policy evaluation failed: %v", err)
		}

		if !decision.Allow {
			t.Errorf("Expected allow=true, got %v", decision.Allow)
		}

		if !decision.ConsentRequired {
			t.Errorf("Expected consent_required=true for non-owner field, got %v", decision.ConsentRequired)
		}

		if len(decision.ConsentRequiredFields) != 1 || decision.ConsentRequiredFields[0] != "test.nonOwnerField" {
			t.Errorf("Expected consent field 'test.nonOwnerField', got %v", decision.ConsentRequiredFields)
		}
	})
}
