package main

import (
	"testing"
)

// TestConsentWorkflowIntegration tests the complete consent management workflow
func TestConsentWorkflowIntegration(t *testing.T) {
	th := NewTestHelper(t)

	t.Run("CompleteWorkflow_ApproveAndOTP", func(t *testing.T) {
		// Step 1: Create consent
		consentID := th.CreateTestConsent(t, "199512345678")
		th.VerifyConsentStatus(t, consentID, "pending")

		// Step 2: Approve consent
		th.ApproveConsent(t, consentID, "199512345678")
		th.VerifyConsentStatus(t, consentID, "approved")

		// Step 3: Verify OTP
		code := th.VerifyOTP(t, consentID, "123456")
		if code != 200 {
			t.Errorf("Expected OTP verification to succeed, got status %d", code)
		}
		th.VerifyConsentStatus(t, consentID, "approved")
	})

	t.Run("CompleteWorkflow_RejectConsent", func(t *testing.T) {
		// Step 1: Create consent
		consentID := th.CreateTestConsent(t, "1991111111")
		th.VerifyConsentStatus(t, consentID, "pending")

		// Step 2: Reject consent
		th.RejectConsent(t, consentID, "1991111111")
		th.VerifyConsentStatus(t, consentID, "rejected")
	})

	t.Run("CompleteWorkflow_OTPFailure", func(t *testing.T) {
		// Step 1: Create and approve consent
		consentID := th.CreateTestConsent(t, "1992222222")
		th.ApproveConsent(t, consentID, "1992222222")

		// Step 2: Try wrong OTP 3 times
		for i := 1; i <= 3; i++ {
			code := th.VerifyOTP(t, consentID, "000000")
			if code != 400 {
				t.Errorf("Attempt %d: Expected status 400, got %d", i, code)
			}
		}

		// Step 3: Verify consent is rejected
		th.VerifyConsentStatus(t, consentID, "rejected")
	})
}

// TestEndToEndPassportApplicationWorkflow tests the complete passport application flow
func TestEndToEndPassportApplicationWorkflow(t *testing.T) {
	th := NewTestHelper(t)

	t.Run("CompletePassportApplicationFlow", func(t *testing.T) {
		// Step 1: User clicks "Apply" button in PassportApp
		t.Log("Step 1: User clicks 'Apply' button in PassportApp")

		// Step 2: Orchestration-engine-go calls policy-decision-point
		t.Log("Step 2: Orchestration-engine-go calls policy-decision-point")

		// Simulate policy decision point response
		policyDecision := map[string]interface{}{
			"allow":                   true,
			"consent_required":        true,
			"consent_required_fields": []string{"person.permanentAddress", "person.birthDate"},
		}

		t.Logf("Policy Decision: %+v", policyDecision)

		// Step 3: Orchestration-engine-go calls consent-engine with POST /consents
		t.Log("Step 3: Orchestration-engine-go calls consent-engine with POST /consents")

		// Create consent for first owner
		consentID1 := th.CreateTestConsent(t, "199512345678")
		t.Logf("Created consent ID 1: %s", consentID1)

		// Create consent for second owner
		consentID2 := th.CreateTestConsent(t, "1995000000")
		t.Logf("Created consent ID 2: %s", consentID2)

		// Step 4: PassportApp is redirected to ConsentPortal (popup window)
		t.Log("Step 4: PassportApp opens ConsentPortal popup window")

		// Step 5: User approves first consent (permanentAddress)
		t.Log("Step 5: User approves first consent (permanentAddress)")
		th.ApproveConsent(t, consentID1, "199512345678")

		// Step 6: OTP workflow for first consent
		t.Log("Step 6: OTP workflow for first consent")
		code := th.VerifyOTP(t, consentID1, "123456")
		if code != 200 {
			t.Errorf("Expected OTP verification to succeed, got status %d", code)
		}

		// Step 7: User rejects second consent (birthDate)
		t.Log("Step 7: User rejects second consent (birthDate)")
		th.RejectConsent(t, consentID2, "1995000000")

		// Step 8: Verify final consent statuses
		t.Log("Step 8: Verify final consent statuses")
		th.VerifyConsentStatus(t, consentID1, "approved")
		th.VerifyConsentStatus(t, consentID2, "rejected")

		// Step 9: Orchestration-engine-go processes consent results
		t.Log("Step 9: Orchestration-engine-go processes consent results")

		// Simulate orchestration engine decision
		consentResults := []map[string]interface{}{
			{
				"consent_id":   consentID1,
				"status":       "approved",
				"redirect_url": "https://passport-app.gov.lk/callback",
				"owner_id":     "199512345678",
				"fields":       []string{"personInfo.permanentAddress"},
			},
			{
				"consent_id":   consentID2,
				"status":       "rejected",
				"redirect_url": "https://passport-app.gov.lk/callback",
				"owner_id":     "1995000000",
				"fields":       []string{"personInfo.birthDate"},
			},
		}

		t.Logf("Consent Results: %+v", consentResults)

		// Step 10: PassportApp shows appropriate message to user
		t.Log("Step 10: PassportApp shows appropriate message to user")

		// Simulate PassportApp behavior based on consent results
		approvedFields := []string{}
		rejectedFields := []string{}

		for _, result := range consentResults {
			if result["status"] == "approved" {
				if fields, ok := result["fields"].([]string); ok {
					approvedFields = append(approvedFields, fields...)
				}
			} else {
				if fields, ok := result["fields"].([]string); ok {
					rejectedFields = append(rejectedFields, fields...)
				}
			}
		}

		if len(approvedFields) > 0 {
			t.Logf("✅ Approved fields: %v", approvedFields)
		}

		if len(rejectedFields) > 0 {
			t.Logf("❌ Rejected fields: %v", rejectedFields)
			t.Log("PassportApp shows: 'Request failed. Please manually enter these fields'")
		}

		// Step 11: Final validation
		t.Log("Step 11: Final validation")

		// Verify that we have the expected number of consent records
		expectedConsentCount := 2
		actualConsentCount := len(th.engine.consentRecords)

		if actualConsentCount != expectedConsentCount {
			t.Errorf("Expected %d consent records, got %d", expectedConsentCount, actualConsentCount)
		}

		// Verify that the consent records have the correct statuses
		consent1Record, exists1 := th.engine.consentRecords[consentID1]
		if !exists1 {
			t.Errorf("Consent record %s not found", consentID1)
		} else if consent1Record.Status != string(StatusApproved) {
			t.Errorf("Expected consent %s to be approved, got %s", consentID1, consent1Record.Status)
		}

		consent2Record, exists2 := th.engine.consentRecords[consentID2]
		if !exists2 {
			t.Errorf("Consent record %s not found", consentID2)
		} else if consent2Record.Status != string(StatusRejected) {
			t.Errorf("Expected consent %s to be rejected, got %s", consentID2, consent2Record.Status)
		}

		t.Log("✅ End-to-end workflow completed successfully!")
	})
}

// TestEndToEndPassportApplicationWorkflowWithOTPFailure tests the complete flow with OTP failure
func TestEndToEndPassportApplicationWorkflowWithOTPFailure(t *testing.T) {
	th := NewTestHelper(t)

	t.Run("CompletePassportApplicationFlowWithOTPFailure", func(t *testing.T) {
		// Step 1-4: Create consent and redirect to portal
		t.Log("Step 1-4: Create consent and redirect to portal")

		consentID := th.CreateTestConsent(t, "199512345678")

		// Step 5: User approves consent
		t.Log("Step 5: User approves consent")
		th.ApproveConsent(t, consentID, "199512345678")

		// Step 6: OTP workflow with 3 failed attempts
		t.Log("Step 6: OTP workflow with 3 failed attempts")

		// Try wrong OTP 3 times
		for i := 1; i <= 3; i++ {
			t.Logf("OTP Attempt %d/3", i)

			code := th.VerifyOTP(t, consentID, "000000")
			if code != 400 {
				t.Errorf("Attempt %d: Expected status 400, got %d", i, code)
			}
		}

		// Step 7: Verify consent is now rejected due to OTP failure
		t.Log("Step 7: Verify consent is now rejected due to OTP failure")
		th.VerifyConsentStatus(t, consentID, "rejected")

		// Step 8: PassportApp shows failure message
		t.Log("Step 8: PassportApp shows failure message")
		t.Log("PassportApp shows: 'Request failed. Please manually enter these fields'")

		t.Log("✅ End-to-end workflow with OTP failure completed successfully!")
	})
}

// TestEndToEndPassportApplicationWorkflowWithExistingConsent tests the flow when consent already exists
func TestEndToEndPassportApplicationWorkflowWithExistingConsent(t *testing.T) {
	th := NewTestHelper(t)

	t.Run("CompletePassportApplicationFlowWithExistingConsent", func(t *testing.T) {
		// Step 1: Create initial consent
		t.Log("Step 1: Create initial consent")

		consentID1 := th.CreateTestConsent(t, "199512345678")

		// Step 2: Try to create consent again with same (app_id, owner_id) pair
		t.Log("Step 2: Try to create consent again with same (app_id, owner_id) pair")

		// This should return the existing consent ID (same owner, different session)
		consentID2 := th.CreateTestConsent(t, "199512345678")

		// Step 3: Verify that the same consent ID is returned
		t.Log("Step 3: Verify that the same consent ID is returned")

		if consentID1 != consentID2 {
			t.Errorf("Expected same consent ID for existing (app_id, owner_id) pair, got %s and %s", consentID1, consentID2)
		}

		// Step 4: Verify that only one consent record exists
		t.Log("Step 4: Verify that only one consent record exists")

		consentCount := len(th.engine.consentRecords)
		if consentCount != 1 {
			t.Errorf("Expected 1 consent record, got %d", consentCount)
		}

		t.Log("✅ End-to-end workflow with existing consent completed successfully!")
	})
}
