package main

import (
	"testing"
	"time"
)

// TestConsentEngine_CreateConsent tests the core CreateConsent functionality
func TestConsentEngine_CreateConsent(t *testing.T) {
	TestWithPostgresEngine(t, "CreateConsent", func(t *testing.T, engine ConsentEngine) {

		req := ConsentRequest{
			AppID: "passport-app",
			ConsentRequirements: []ConsentRequirement{
				{
					Owner:   "CITIZEN",
					OwnerID: "user123@example.com",
					Fields: []ConsentField{
						{
							FieldName: "person.permanentAddress",
							SchemaID:  "schema_123",
						},
					},
				},
			},
		}

		record, err := engine.CreateConsent(req)
		if err != nil {
			t.Fatalf("CreateConsent failed: %v", err)
		}

		if record.ConsentID == "" {
			t.Error("Expected non-empty consent ID")
		}

		if record.Status != string(StatusPending) {
			t.Errorf("Expected status=%s, got %s", string(StatusPending), record.Status)
		}

		if record.AppID != req.AppID {
			t.Errorf("Expected AppID=%s, got %s", req.AppID, record.AppID)
		}

		if record.OwnerID != req.ConsentRequirements[0].OwnerID {
			t.Errorf("Expected OwnerID=%s, got %s", req.ConsentRequirements[0].OwnerID, record.OwnerID)
		}

		if record.OwnerEmail != req.ConsentRequirements[0].OwnerID {
			t.Errorf("Expected OwnerEmail=%s, got %s", req.ConsentRequirements[0].OwnerID, record.OwnerEmail)
		}
	})
}

// TestConsentEngine_GetConsentStatus tests retrieving consent status
func TestConsentEngine_GetConsentStatus(t *testing.T) {
	TestWithPostgresEngine(t, "GetConsentStatus", func(t *testing.T, engine ConsentEngine) {

		// Create a consent record first
		req := ConsentRequest{
			AppID: "passport-app",
			ConsentRequirements: []ConsentRequirement{
				{
					Owner:   "CITIZEN",
					OwnerID: "user123@example.com",
					Fields: []ConsentField{
						{
							FieldName: "person.permanentAddress",
							SchemaID:  "schema_123",
						},
					},
				},
			},
		}

		createdRecord, err := engine.CreateConsent(req)
		if err != nil {
			t.Fatalf("CreateConsent failed: %v", err)
		}

		// Test getting the consent status
		record, err := engine.GetConsentStatus(createdRecord.ConsentID)
		if err != nil {
			t.Fatalf("GetConsentStatus failed: %v", err)
		}

		if record.ConsentID != createdRecord.ConsentID {
			t.Errorf("Expected ConsentID=%s, got %s", createdRecord.ConsentID, record.ConsentID)
		}

		// Test getting non-existent consent
		_, err = engine.GetConsentStatus("non-existent-id")
		if err == nil {
			t.Error("Expected error for non-existent consent ID")
		}
	})
}

// TestConsentEngine_UpdateConsent tests updating consent status
func TestConsentEngine_UpdateConsent(t *testing.T) {
	TestWithPostgresEngine(t, "UpdateConsent", func(t *testing.T, engine ConsentEngine) {

		// Create a consent record first
		req := ConsentRequest{
			AppID: "passport-app",
			ConsentRequirements: []ConsentRequirement{
				{
					Owner:   "CITIZEN",
					OwnerID: "user123@example.com",
					Fields: []ConsentField{
						{
							FieldName: "person.permanentAddress",
							SchemaID:  "schema_123",
						},
					},
				},
			},
		}

		createdRecord, err := engine.CreateConsent(req)
		if err != nil {
			t.Fatalf("CreateConsent failed: %v", err)
		}

		// Test updating consent status
		updateReq := UpdateConsentRequest{
			Status:    StatusApproved,
			UpdatedBy: "user123",
			Reason:    "User approved consent",
		}

		updatedRecord, err := engine.UpdateConsent(createdRecord.ConsentID, updateReq)
		if err != nil {
			t.Fatalf("UpdateConsent failed: %v", err)
		}

		if updatedRecord.Status != string(StatusApproved) {
			t.Errorf("Expected status=%s, got %s", string(StatusApproved), updatedRecord.Status)
		}

		// Test invalid status transition (approved -> pending is not allowed)
		invalidUpdateReq := UpdateConsentRequest{
			Status:    StatusPending,
			UpdatedBy: "user123",
			Reason:    "Invalid transition",
		}

		_, err = engine.UpdateConsent(createdRecord.ConsentID, invalidUpdateReq)
		if err == nil {
			t.Error("Expected error for invalid status transition from approved to pending")
		}
	})
}

// TestConsentEngine_FindExistingConsent tests finding existing consents
func TestConsentEngine_FindExistingConsent(t *testing.T) {
	TestWithPostgresEngine(t, "FindExistingConsent", func(t *testing.T, engine ConsentEngine) {

		// Create a consent record first
		req := ConsentRequest{
			AppID: "passport-app",
			ConsentRequirements: []ConsentRequirement{
				{
					Owner:   "CITIZEN",
					OwnerID: "user123@example.com",
					Fields: []ConsentField{
						{
							FieldName: "person.permanentAddress",
							SchemaID:  "schema_123",
						},
					},
				},
			},
		}

		createdRecord, err := engine.CreateConsent(req)
		if err != nil {
			t.Fatalf("CreateConsent failed: %v", err)
		}

		// Test finding existing consent
		foundRecord := engine.FindExistingConsent("passport-app", "user123")
		if foundRecord == nil {
			t.Error("Expected to find existing consent record")
		} else if foundRecord.ConsentID != createdRecord.ConsentID {
			t.Errorf("Expected ConsentID=%s, got %s", createdRecord.ConsentID, foundRecord.ConsentID)
		}

		// Test finding non-existent consent
		notFoundRecord := engine.FindExistingConsent("different-app", "user123")
		if notFoundRecord != nil {
			t.Error("Expected not to find consent record for different app")
		}
	})
}

// TestConsentEngine_StatusTransitions tests status transition validation
func TestConsentEngine_StatusTransitions(t *testing.T) {
	// Test valid transitions
	validTransitions := []struct {
		from  ConsentStatus
		to    ConsentStatus
		valid bool
	}{
		// Pending state transitions
		{StatusPending, StatusApproved, true},
		{StatusPending, StatusRejected, true},
		{StatusPending, StatusExpired, true},

		// Approved state transitions
		{StatusApproved, StatusApproved, true}, // Direct approval
		{StatusApproved, StatusRejected, true}, // Direct rejection
		{StatusApproved, StatusRevoked, true},  // User revocation
		{StatusApproved, StatusExpired, true},  // Expiry

		// Terminal states - can only transition to expired or stay the same
		{StatusRejected, StatusExpired, true}, // Rejected -> Expired
		{StatusExpired, StatusExpired, true},  // Expired -> Expired
		{StatusRevoked, StatusExpired, true},  // Revoked -> Expired

		// Invalid transitions from terminal states
		{StatusRejected, StatusPending, false},  // Terminal state cannot go back to pending
		{StatusRejected, StatusApproved, false}, // Terminal state cannot go back to approved
		{StatusExpired, StatusPending, false},   // Terminal state cannot go back to pending
		{StatusExpired, StatusApproved, false},  // Terminal state cannot go back to approved
		{StatusRevoked, StatusPending, false},   // Terminal state cannot go back to pending
		{StatusRevoked, StatusApproved, false},  // Terminal state cannot go back to approved

		// Invalid transitions from approved
		{StatusApproved, StatusPending, false}, // Cannot go back to pending
	}

	for _, tt := range validTransitions {
		result := isValidStatusTransition(tt.from, tt.to)
		if result != tt.valid {
			t.Errorf("isValidStatusTransition(%v, %v) = %v, want %v", tt.from, tt.to, result, tt.valid)
		}
	}
}

// TestConsentEngine_CheckConsentExpiry tests the CheckConsentExpiry functionality
func TestConsentEngine_CheckConsentExpiry(t *testing.T) {
	engine := setupPostgresTestEngine(t)

	t.Run("NoExpiredRecords", func(t *testing.T) {
		// Create a consent that won't expire soon
		req := ConsentRequest{
			AppID: "passport-app",
			ConsentRequirements: []ConsentRequirement{
				{
					Owner:   "CITIZEN",
					OwnerID: "user123@example.com",
					Fields: []ConsentField{
						{
							FieldName: "person.permanentAddress",
							SchemaID:  "schema_123",
						},
					},
				},
			},
		}

		record, err := engine.CreateConsent(req)
		if err != nil {
			t.Fatalf("CreateConsent failed: %v", err)
		}

		// Approve the consent
		updateReq := UpdateConsentRequest{
			Status:    StatusApproved,
			UpdatedBy: "citizen_123",
			Reason:    "User approved",
		}
		_, err = engine.UpdateConsent(record.ConsentID, updateReq)
		if err != nil {
			t.Fatalf("UpdateConsent failed: %v", err)
		}

		// Check expiry - should return empty list (no records to delete)
		deletedRecords, err := engine.CheckConsentExpiry()
		if err != nil {
			t.Fatalf("CheckConsentExpiry failed: %v", err)
		}

		if len(deletedRecords) != 0 {
			t.Errorf("Expected 0 deleted records, got %d", len(deletedRecords))
		}
	})

	t.Run("HasExpiredRecords", func(t *testing.T) {
		// Create a new engine to avoid interference
		engine := setupPostgresTestEngine(t)

		// Create a consent
		req := ConsentRequest{
			AppID: "passport-app",
			ConsentRequirements: []ConsentRequirement{
				{
					Owner:   "CITIZEN",
					OwnerID: "user456@example.com",
					Fields: []ConsentField{
						{
							FieldName: "person.permanentAddress",
							SchemaID:  "schema_123",
						},
					},
				},
			},
		}

		record, err := engine.CreateConsent(req)
		if err != nil {
			t.Fatalf("CreateConsent failed: %v", err)
		}

		// Approve the consent
		updateReq := UpdateConsentRequest{
			Status:    StatusApproved,
			UpdatedBy: "citizen_456",
			Reason:    "User approved",
		}
		_, err = engine.UpdateConsent(record.ConsentID, updateReq)
		if err != nil {
			t.Fatalf("UpdateConsent failed: %v", err)
		}

		// For testing expiry, we'll create a record with a very short grant duration
		// and then wait for it to expire
		expiredRecord := ConsentRequest{
			AppID: "test-app-2",
			ConsentRequirements: []ConsentRequirement{
				{
					Owner:   "CITIZEN",
					OwnerID: "test2@example.com",
					Fields: []ConsentField{
						{
							FieldName: "email",
							SchemaID:  "schema_123",
						},
					},
				},
			},
			GrantDuration: "1s", // Very short duration to ensure expiry
		}

		expiredConsent, err := engine.CreateConsent(expiredRecord)
		if err != nil {
			t.Fatalf("CreateConsent failed: %v", err)
		}

		// Update the expired record to be approved
		updateExpiredReq := UpdateConsentRequest{
			Status: StatusApproved,
		}
		_, err = engine.UpdateConsent(expiredConsent.ConsentID, updateExpiredReq)
		if err != nil {
			t.Fatalf("Failed to update expired consent: %v", err)
		}

		// Wait a bit to ensure the record is expired
		time.Sleep(2 * time.Second)

		// Check expiry - should delete the expired record
		deletedRecords, err := engine.CheckConsentExpiry()
		if err != nil {
			t.Fatalf("CheckConsentExpiry failed: %v", err)
		}

		if len(deletedRecords) != 1 {
			t.Errorf("Expected 1 deleted record, got %d", len(deletedRecords))
		}

		if deletedRecords[0].ConsentID != expiredConsent.ConsentID {
			t.Errorf("Expected deleted record ID %s, got %s", expiredConsent.ConsentID, deletedRecords[0].ConsentID)
		}

		// Verify the record was actually deleted from the store
		_, err = engine.GetConsentStatus(expiredConsent.ConsentID)
		if err == nil {
			t.Errorf("Expected record to be deleted, but it still exists")
		}
	})

	t.Run("OnlyApprovedRecordsExpire", func(t *testing.T) {
		// Create a new engine to avoid interference
		engine := setupPostgresTestEngine(t)

		// Create a consent
		req := ConsentRequest{
			AppID: "passport-app",
			ConsentRequirements: []ConsentRequirement{
				{
					Owner:   "CITIZEN",
					OwnerID: "user789@example.com",
					Fields: []ConsentField{
						{
							FieldName: "person.permanentAddress",
							SchemaID:  "schema_123",
						},
					},
				},
			},
		}

		record, err := engine.CreateConsent(req)
		if err != nil {
			t.Fatalf("CreateConsent failed: %v", err)
		}

		// Reject the consent (not approved)
		updateReq := UpdateConsentRequest{
			Status:    StatusRejected,
			UpdatedBy: "citizen_789",
			Reason:    "User rejected",
		}
		_, err = engine.UpdateConsent(record.ConsentID, updateReq)
		if err != nil {
			t.Fatalf("UpdateConsent failed: %v", err)
		}

		// For testing expiry, we'll create a new expired record instead
		// since we can't directly modify the database timestamp
		expiredRecord2 := ConsentRequest{
			AppID: "test-app-3",
			ConsentRequirements: []ConsentRequirement{
				{
					Owner:   "CITIZEN",
					OwnerID: "test3@example.com",
					Fields: []ConsentField{
						{
							FieldName: "phone",
							SchemaID:  "schema_123",
						},
					},
				},
			},
			GrantDuration: "1h",
		}

		expiredConsent2, err := engine.CreateConsent(expiredRecord2)
		if err != nil {
			t.Fatalf("CreateConsent failed: %v", err)
		}

		// Update the expired record to be approved
		updateExpiredReq2 := UpdateConsentRequest{
			Status: StatusApproved,
		}
		_, err = engine.UpdateConsent(expiredConsent2.ConsentID, updateExpiredReq2)
		if err != nil {
			t.Fatalf("Failed to update expired consent: %v", err)
		}

		// Check expiry - should return empty list (rejected records don't get deleted)
		deletedRecords, err := engine.CheckConsentExpiry()
		if err != nil {
			t.Fatalf("CheckConsentExpiry failed: %v", err)
		}

		if len(deletedRecords) != 0 {
			t.Errorf("Expected 0 deleted records (rejected records don't get deleted), got %d", len(deletedRecords))
		}
	})

	t.Run("MultipleExpiredRecords", func(t *testing.T) {
		// Create a new engine to avoid interference
		engine := setupPostgresTestEngine(t)

		// Create multiple consents with short grant durations
		consentIDs := make([]string, 3)
		for i := 0; i < 3; i++ {
			req := ConsentRequest{
				AppID: "passport-app",
				ConsentRequirements: []ConsentRequirement{
					{
						Owner:   "CITIZEN",
						OwnerID: "user" + string(rune('0'+i)) + "@example.com",
						Fields: []ConsentField{
							{
								FieldName: "person.permanentAddress",
								SchemaID:  "schema_123",
							},
						},
					},
				},
				GrantDuration: "1s", // Very short duration to ensure expiry
			}

			record, err := engine.CreateConsent(req)
			if err != nil {
				t.Fatalf("CreateConsent failed: %v", err)
			}

			// Approve the consent
			updateReq := UpdateConsentRequest{
				Status:    StatusApproved,
				UpdatedBy: "citizen_" + string(rune('0'+i)),
				Reason:    "User approved",
			}
			_, err = engine.UpdateConsent(record.ConsentID, updateReq)
			if err != nil {
				t.Fatalf("UpdateConsent failed: %v", err)
			}

			consentIDs[i] = record.ConsentID
		}

		// Wait for all records to expire
		time.Sleep(2 * time.Second)

		// Check expiry - should delete all 3 expired records
		deletedRecords, err := engine.CheckConsentExpiry()
		if err != nil {
			t.Fatalf("CheckConsentExpiry failed: %v", err)
		}

		if len(deletedRecords) != 3 {
			t.Errorf("Expected 3 deleted records, got %d", len(deletedRecords))
		}

		// Verify all records were actually deleted from the store
		for _, consentID := range consentIDs {
			_, err = engine.GetConsentStatus(consentID)
			if err == nil {
				t.Errorf("Expected record %s to be deleted, but it still exists", consentID)
			}
		}
	})
}

// TestConsentEngine_UpdateConsentWithGrantDuration tests updating consent with grant_duration
func TestConsentEngine_UpdateConsentWithGrantDuration(t *testing.T) {
	engine := setupPostgresTestEngine(t)

	// Create a consent
	req := ConsentRequest{
		AppID: "passport-app",
		ConsentRequirements: []ConsentRequirement{
			{
				Owner:   "CITIZEN",
				OwnerID: "user123@example.com",
				Fields: []ConsentField{
					{
						FieldName: "person.permanentAddress",
						SchemaID:  "schema_123",
					},
				},
			},
		},
	}

	record, err := engine.CreateConsent(req)
	if err != nil {
		t.Fatalf("CreateConsent failed: %v", err)
	}

	// Update consent with grant_duration
	updateReq := UpdateConsentRequest{
		Status:        StatusApproved,
		GrantDuration: "1m",
		UpdatedBy:     "citizen_123",
		Reason:        "User approved with custom duration",
	}

	updatedRecord, err := engine.UpdateConsent(record.ConsentID, updateReq)
	if err != nil {
		t.Fatalf("UpdateConsent failed: %v", err)
	}

	// Verify grant_duration was updated
	if updatedRecord.GrantDuration != "1m" {
		t.Errorf("Expected grant_duration '1m', got %s", updatedRecord.GrantDuration)
	}

	// Verify expires_at was recalculated (should be approximately 1 minute from now)
	expectedExpiry := time.Now().Add(1 * time.Minute)
	timeDiff := updatedRecord.ExpiresAt.Sub(expectedExpiry)
	if timeDiff < -5*time.Second || timeDiff > 5*time.Second {
		t.Errorf("Expected expires_at to be approximately 1 minute from now, got %v", updatedRecord.ExpiresAt)
	}

	// Verify status was updated
	if updatedRecord.Status != string(StatusApproved) {
		t.Errorf("Expected status %s, got %s", string(StatusApproved), updatedRecord.Status)
	}
}

// TestRejectedConsentReuseIssue tests that rejected consents are reused and updated
func TestRejectedConsentReuseIssue(t *testing.T) {
	// Create a new consent engine
	engine := setupPostgresTestEngine(t)

	// Create initial consent request
	req1 := ConsentRequest{
		AppID:         "test-app",
		GrantDuration: "1h",
		ConsentRequirements: []ConsentRequirement{
			{
				Owner:   "CITIZEN",
				OwnerID: "user@example.com",
				Fields: []ConsentField{
					{
						FieldName: "name",
						SchemaID:  "schema_test",
					},
					{
						FieldName: "email",
						SchemaID:  "schema_test",
					},
				},
			},
		},
	}

	// Create the first consent record
	record1, err := engine.ProcessConsentRequest(req1)
	if err != nil {
		t.Fatalf("Failed to create first consent: %v", err)
	}

	t.Logf("First consent created: ID=%s, Status=%s", record1.ConsentID, record1.Status)

	// Reject the consent
	updateReq := UpdateConsentRequest{
		Status:    StatusRejected,
		UpdatedBy: "test-user",
		Reason:    "testing rejection",
	}

	rejectedRecord, err := engine.UpdateConsent(record1.ConsentID, updateReq)
	if err != nil {
		t.Fatalf("Failed to reject consent: %v", err)
	}

	if rejectedRecord.Status != string(StatusRejected) {
		t.Errorf("Expected status to be rejected, got %s", rejectedRecord.Status)
	}

	t.Logf("Consent rejected: ID=%s, Status=%s", rejectedRecord.ConsentID, rejectedRecord.Status)

	// Now send the same request but with different number of fields
	req2 := ConsentRequest{
		AppID:         "test-app",
		GrantDuration: "1h",
		ConsentRequirements: []ConsentRequirement{
			{
				Owner:   "CITIZEN",
				OwnerID: "user@example.com",
				Fields: []ConsentField{
					{
						FieldName: "name",
						SchemaID:  "schema_test",
					},
					{
						FieldName: "email",
						SchemaID:  "schema_test",
					},
					{
						FieldName: "phone",
						SchemaID:  "schema_test",
					},
				},
			},
		},
	}

	// This should reuse the existing rejected record, not create a new one
	record2, err := engine.ProcessConsentRequest(req2)
	if err != nil {
		t.Fatalf("Failed to process second consent request: %v", err)
	}

	t.Logf("Second consent processed: ID=%s, Status=%s", record2.ConsentID, record2.Status)

	// When a consent is rejected, we should reuse that record and update the fields
	if record2.ConsentID != record1.ConsentID {
		t.Errorf("Expected to reuse existing consent ID %s, but got new ID %s", record1.ConsentID, record2.ConsentID)
		t.Logf("ISSUE: Created new consent record when it should have reused the rejected one")
	} else {
		t.Logf("SUCCESS: Reused existing rejected consent record as expected")
	}

	// Verify the record remains rejected (not changed to pending)
	if record2.Status != string(StatusRejected) {
		t.Errorf("Expected status to remain rejected, got %s", record2.Status)
	}

	// Verify the fields were updated
	expectedFields := []string{"name", "email", "phone"}
	if len(record2.Fields) != len(expectedFields) {
		t.Errorf("Expected %d fields, got %d", len(expectedFields), len(record2.Fields))
	}

	// Stop the background process
	engine.StopBackgroundExpiryProcess()
}

// TestConsentReuseLogic tests the correct behavior for consent record reuse based on status
func TestConsentReuseLogic(t *testing.T) {
	// Create a new consent engine
	engine := setupPostgresTestEngine(t)

	baseReq := ConsentRequest{
		AppID:         "test-app",
		GrantDuration: "1h",
		ConsentRequirements: []ConsentRequirement{
			{
				Owner:   "CITIZEN",
				OwnerID: "user@example.com",
				Fields: []ConsentField{
					{
						FieldName: "name",
						SchemaID:  "schema_test",
					},
					{
						FieldName: "email",
						SchemaID:  "schema_test",
					},
				},
			},
		},
	}

	// Test 1: Pending records should be reused
	t.Log("=== Test 1: Pending records should be reused ===")
	record1, err := engine.ProcessConsentRequest(baseReq)
	if err != nil {
		t.Fatalf("Failed to create first consent: %v", err)
	}
	t.Logf("First consent: ID=%s, Status=%s", record1.ConsentID, record1.Status)

	// Send same request - should reuse pending record
	record2, err := engine.ProcessConsentRequest(baseReq)
	if err != nil {
		t.Fatalf("Failed to process second request: %v", err)
	}
	t.Logf("Second request: ID=%s, Status=%s", record2.ConsentID, record2.Status)

	if record2.ConsentID != record1.ConsentID {
		t.Errorf("Expected to reuse pending consent ID %s, but got new ID %s", record1.ConsentID, record2.ConsentID)
	} else {
		t.Logf("✓ SUCCESS: Reused pending consent record")
	}

	// Test 2: Rejected records should be reused and updated
	t.Log("\n=== Test 2: Rejected records should be reused and updated ===")
	updateReq := UpdateConsentRequest{
		Status:    StatusRejected,
		UpdatedBy: "test-user",
		Reason:    "testing rejection",
	}

	rejectedRecord, err := engine.UpdateConsent(record1.ConsentID, updateReq)
	if err != nil {
		t.Fatalf("Failed to reject consent: %v", err)
	}
	t.Logf("Consent rejected: ID=%s, Status=%s", rejectedRecord.ConsentID, rejectedRecord.Status)

	// Send same request after rejection - should reuse and update the rejected record
	record3, err := engine.ProcessConsentRequest(baseReq)
	if err != nil {
		t.Fatalf("Failed to process third request: %v", err)
	}
	t.Logf("Third request: ID=%s, Status=%s", record3.ConsentID, record3.Status)

	if record3.ConsentID != record1.ConsentID {
		t.Errorf("Expected to reuse rejected consent ID %s, but got new ID %s", record1.ConsentID, record3.ConsentID)
	} else {
		t.Logf("✓ SUCCESS: Reused rejected consent record and updated fields")
	}

	// Verify the record remains rejected
	if record3.Status != string(StatusRejected) {
		t.Errorf("Expected status to remain rejected, got %s", record3.Status)
	}

	// Test 3: Approved records should be reused and updated
	t.Log("\n=== Test 3: Approved records should be reused and updated ===")
	// Create a new consent record with different owner to avoid conflicts
	approveReq := ConsentRequest{
		AppID: "test-app",
		ConsentRequirements: []ConsentRequirement{
			{
				Owner:   "CITIZEN",
				OwnerID: "user456@example.com", // Different owner to create new record
				Fields: []ConsentField{
					{
						FieldName: "name",
						SchemaID:  "schema_456",
					},
					{
						FieldName: "email",
						SchemaID:  "schema_456",
					},
				},
			},
		},
		GrantDuration: "1h",
	}

	record4, err := engine.ProcessConsentRequest(approveReq)
	if err != nil {
		t.Fatalf("Failed to process fourth request: %v", err)
	}
	t.Logf("Fourth request: ID=%s, Status=%s", record4.ConsentID, record4.Status)

	// Approve the new record
	updateReq2 := UpdateConsentRequest{
		Status:    StatusApproved,
		UpdatedBy: "test-user",
		Reason:    "testing approval",
	}

	approvedRecord, err := engine.UpdateConsent(record4.ConsentID, updateReq2)
	if err != nil {
		t.Fatalf("Failed to approve consent: %v", err)
	}
	t.Logf("Consent approved: ID=%s, Status=%s", approvedRecord.ConsentID, approvedRecord.Status)

	// Send same request after approval - should reuse and update the approved record
	record5, err := engine.ProcessConsentRequest(approveReq)
	if err != nil {
		t.Fatalf("Failed to process fifth request: %v", err)
	}
	t.Logf("Fifth request: ID=%s, Status=%s", record5.ConsentID, record5.Status)

	if record5.ConsentID != record4.ConsentID {
		t.Errorf("Expected to reuse approved consent ID %s, but got new ID %s", record4.ConsentID, record5.ConsentID)
	} else {
		t.Logf("✓ SUCCESS: Reused approved consent record and updated fields")
	}

	// Verify the record remains approved
	if record5.Status != string(StatusApproved) {
		t.Errorf("Expected status to remain approved, got %s", record5.Status)
	}

	// Test 4: Only ONE pending record per (appId, ownerId, ownerEmail) tuple
	t.Log("\n=== Test 4: Only ONE pending record per tuple ===")
	// Send another request - should reuse the approved record from Test 3
	record6, err := engine.ProcessConsentRequest(approveReq)
	if err != nil {
		t.Fatalf("Failed to process sixth request: %v", err)
	}
	t.Logf("Sixth request: ID=%s, Status=%s", record6.ConsentID, record6.Status)

	if record6.ConsentID != record5.ConsentID {
		t.Errorf("Expected to reuse approved consent ID %s, but got new ID %s", record5.ConsentID, record6.ConsentID)
	} else {
		t.Logf("✓ SUCCESS: Reused approved consent record")
	}

	// Stop the background process
	engine.StopBackgroundExpiryProcess()
}
