package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gov-dx-sandbox/exchange/consent-engine/v1/models"
	"github.com/stretchr/testify/assert"
)

func TestPostgresConsentEngine_FindExistingConsent(t *testing.T) {
	testEngine := SetupPostgresTestEngineWithDB(t)
	engine := testEngine.Engine

	// Create a consent first
	createReq := models.ConsentRequest{
		AppID: "test-app",
		ConsentRequirements: []models.ConsentRequirement{
			{
				Owner:   "CITIZEN",
				OwnerID: "user@example.com",
				Fields: []models.ConsentField{
					{
						FieldName: "personInfo.name",
						SchemaID:  "schema-123",
					},
				},
			},
		},
	}
	record, err := engine.ProcessConsentRequest(createReq)
	assert.NoError(t, err)
	assert.NotNil(t, record)

	// Test FindExistingConsent
	found := engine.FindExistingConsent("test-app", "user@example.com")
	assert.NotNil(t, found)
	assert.Equal(t, record.ConsentID, found.ConsentID)
	assert.Equal(t, "test-app", found.AppID)
	assert.Equal(t, "user@example.com", found.OwnerID)

	// Test with non-existent consent
	notFound := engine.FindExistingConsent("non-existent-app", "user@example.com")
	assert.Nil(t, notFound)
}

func TestPostgresConsentEngine_ProcessConsentPortalRequest(t *testing.T) {
	testEngine := SetupPostgresTestEngineWithDB(t)
	engine := testEngine.Engine

	// Create a consent first
	createReq := models.ConsentRequest{
		AppID: "test-app",
		ConsentRequirements: []models.ConsentRequirement{
			{
				Owner:   "CITIZEN",
				OwnerID: "user@example.com",
				Fields: []models.ConsentField{
					{
						FieldName: "personInfo.name",
						SchemaID:  "schema-123",
					},
				},
			},
		},
	}
	record, err := engine.ProcessConsentRequest(createReq)
	assert.NoError(t, err)
	assert.NotNil(t, record)

	t.Run("Approve action", func(t *testing.T) {
		portalReq := models.ConsentPortalRequest{
			ConsentID: record.ConsentID,
			Action:    "approve",
			DataOwner: "user@example.com",
			Reason:    "Test reason",
		}

		result, err := engine.ProcessConsentPortalRequest(portalReq)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, string(models.StatusApproved), result.Status)
	})

	t.Run("Deny action", func(t *testing.T) {
		// Create a fresh consent for deny test
		createReq := models.ConsentRequest{
			AppID: "test-app-deny",
			ConsentRequirements: []models.ConsentRequirement{
				{
					Owner:   "CITIZEN",
					OwnerID: "user-deny@example.com",
					Fields: []models.ConsentField{
						{
							FieldName: "personInfo.name",
							SchemaID:  "schema-123",
						},
					},
				},
			},
		}
		testRecord, err := engine.ProcessConsentRequest(createReq)
		assert.NoError(t, err)

		portalReq := models.ConsentPortalRequest{
			ConsentID: testRecord.ConsentID,
			Action:    "deny",
			DataOwner: "user-deny@example.com",
			Reason:    "Test reason",
		}

		result, err := engine.ProcessConsentPortalRequest(portalReq)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, string(models.StatusRejected), result.Status)
	})

	t.Run("Revoke action", func(t *testing.T) {
		// Create a fresh consent with unique app ID and approve it first, then revoke
		createReq := models.ConsentRequest{
			AppID: "test-app-revoke-unique",
			ConsentRequirements: []models.ConsentRequirement{
				{
					Owner:   "CITIZEN",
					OwnerID: "user-revoke@example.com",
					Fields: []models.ConsentField{
						{
							FieldName: "personInfo.name",
							SchemaID:  "schema-123",
						},
					},
				},
			},
		}
		testRecord, err := engine.ProcessConsentRequest(createReq)
		assert.NoError(t, err)

		// Approve first
		approveReq := models.ConsentPortalRequest{
			ConsentID: testRecord.ConsentID,
			Action:    "approve",
			DataOwner: "user-revoke@example.com",
			Reason:    "Test approval",
		}
		approved, err := engine.ProcessConsentPortalRequest(approveReq)
		assert.NoError(t, err)
		assert.Equal(t, string(models.StatusApproved), approved.Status)

		// Now revoke (approved can transition to revoked)
		revokeReq := models.ConsentPortalRequest{
			ConsentID: testRecord.ConsentID,
			Action:    "revoke",
			DataOwner: "user-revoke@example.com",
			Reason:    "Test reason",
		}

		result, err := engine.ProcessConsentPortalRequest(revokeReq)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, string(models.StatusRevoked), result.Status)
	})

	t.Run("Invalid action", func(t *testing.T) {
		portalReq := models.ConsentPortalRequest{
			ConsentID: record.ConsentID,
			Action:    "invalid",
			DataOwner: "user@example.com",
			Reason:    "Test reason",
		}

		result, err := engine.ProcessConsentPortalRequest(portalReq)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid action")
	})
}

func TestPostgresConsentEngine_StartBackgroundExpiryProcess(t *testing.T) {
	testEngine := SetupPostgresTestEngineWithDB(t)
	engine := testEngine.Engine

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start background process with short interval
	// This should not panic - the function call itself is synchronous
	engine.StartBackgroundExpiryProcess(ctx, 100*time.Millisecond)

	// The background process runs in a goroutine, so we can't easily synchronize with it
	// without modifying the implementation. This test verifies that:
	// 1. Starting the process doesn't panic
	// 2. Cancelling the context doesn't panic
	// The actual background process execution is tested in integration tests

	// Cancel context to stop the process
	cancel()

	// Test passes if no panic occurs during start and stop
}

func TestPostgresConsentEngine_StopBackgroundExpiryProcess(t *testing.T) {
	testEngine := SetupPostgresTestEngineWithDB(t)
	engine := testEngine.Engine

	// This should not panic
	engine.StopBackgroundExpiryProcess()
}

// TestPostgresConsentEngine_CreateConsent tests the CreateConsent wrapper method
func TestPostgresConsentEngine_CreateConsent(t *testing.T) {
	testEngine := SetupPostgresTestEngineWithDB(t)
	engine := testEngine.Engine

	req := models.ConsentRequest{
		AppID: "test-app-create",
		ConsentRequirements: []models.ConsentRequirement{
			{
				Owner:   "CITIZEN",
				OwnerID: "user-create@example.com",
				Fields: []models.ConsentField{
					{
						FieldName: "personInfo.name",
						SchemaID:  "schema-123",
					},
				},
			},
		},
	}

	record, err := engine.CreateConsent(req)
	assert.NoError(t, err)
	assert.NotNil(t, record)
	assert.Equal(t, "test-app-create", record.AppID)
	assert.Equal(t, "user-create@example.com", record.OwnerID)
	assert.Equal(t, string(models.StatusPending), record.Status)
}

// TestPostgresConsentEngine_ProcessConsentRequest_ValidationErrors tests validation error cases
func TestPostgresConsentEngine_ProcessConsentRequest_ValidationErrors(t *testing.T) {
	testEngine := SetupPostgresTestEngineWithDB(t)
	engine := testEngine.Engine

	t.Run("Empty AppID", func(t *testing.T) {
		req := models.ConsentRequest{
			AppID: "",
			ConsentRequirements: []models.ConsentRequirement{
				{
					Owner:   "CITIZEN",
					OwnerID: "user@example.com",
					Fields: []models.ConsentField{
						{
							FieldName: "personInfo.name",
							SchemaID:  "schema-123",
						},
					},
				},
			},
		}
		record, err := engine.ProcessConsentRequest(req)
		assert.Error(t, err)
		assert.Nil(t, record)
		assert.Contains(t, err.Error(), "app_id is required")
	})

	t.Run("Empty ConsentRequirements", func(t *testing.T) {
		req := models.ConsentRequest{
			AppID:               "test-app",
			ConsentRequirements: []models.ConsentRequirement{},
		}
		record, err := engine.ProcessConsentRequest(req)
		assert.Error(t, err)
		assert.Nil(t, record)
		assert.Contains(t, err.Error(), "consent_requirements is required")
	})
}

// TestPostgresConsentEngine_GetConsentStatus_NotFound tests GetConsentStatus with non-existent ID
func TestPostgresConsentEngine_GetConsentStatus_NotFound(t *testing.T) {
	testEngine := SetupPostgresTestEngineWithDB(t)
	engine := testEngine.Engine

	record, err := engine.GetConsentStatus("non-existent-consent-id")
	assert.Error(t, err)
	assert.Nil(t, record)
	assert.Contains(t, err.Error(), "not found")
}

// TestPostgresConsentEngine_UpdateConsent_EdgeCases tests UpdateConsent edge cases
func TestPostgresConsentEngine_UpdateConsent_EdgeCases(t *testing.T) {
	testEngine := SetupPostgresTestEngineWithDB(t)
	engine := testEngine.Engine

	t.Run("Non-existent consent", func(t *testing.T) {
		updateReq := models.UpdateConsentRequest{
			Status: models.StatusApproved,
		}
		record, err := engine.UpdateConsent("non-existent-id", updateReq)
		assert.Error(t, err)
		assert.Nil(t, record)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("Invalid status transition", func(t *testing.T) {
		// Create a consent
		createReq := models.ConsentRequest{
			AppID: "test-app-update",
			ConsentRequirements: []models.ConsentRequirement{
				{
					Owner:   "CITIZEN",
					OwnerID: "user-update@example.com",
					Fields: []models.ConsentField{
						{
							FieldName: "personInfo.name",
							SchemaID:  "schema-123",
						},
					},
				},
			},
		}
		record, err := engine.ProcessConsentRequest(createReq)
		assert.NoError(t, err)

		// Try invalid transition: pending -> revoked (not allowed)
		updateReq := models.UpdateConsentRequest{
			Status: models.StatusRevoked,
		}
		updated, err := engine.UpdateConsent(record.ConsentID, updateReq)
		assert.Error(t, err)
		assert.Nil(t, updated)
		assert.Contains(t, err.Error(), "invalid status transition")
	})

	t.Run("Update with empty grant duration uses existing", func(t *testing.T) {
		// Create a consent with specific grant duration
		createReq := models.ConsentRequest{
			AppID:         "test-app-grant",
			GrantDuration: "P2D",
			ConsentRequirements: []models.ConsentRequirement{
				{
					Owner:   "CITIZEN",
					OwnerID: "user-grant@example.com",
					Fields: []models.ConsentField{
						{
							FieldName: "personInfo.name",
							SchemaID:  "schema-123",
						},
					},
				},
			},
		}
		record, err := engine.ProcessConsentRequest(createReq)
		assert.NoError(t, err)

		// Update with empty grant duration - should use existing
		updateReq := models.UpdateConsentRequest{
			Status:        models.StatusApproved,
			GrantDuration: "", // Empty - should use existing
		}
		updated, err := engine.UpdateConsent(record.ConsentID, updateReq)
		assert.NoError(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, "P2D", updated.GrantDuration)
	})

	t.Run("Update with empty fields uses existing", func(t *testing.T) {
		// Create a consent with fields
		createReq := models.ConsentRequest{
			AppID: "test-app-fields",
			ConsentRequirements: []models.ConsentRequirement{
				{
					Owner:   "CITIZEN",
					OwnerID: "user-fields@example.com",
					Fields: []models.ConsentField{
						{
							FieldName: "personInfo.name",
							SchemaID:  "schema-123",
						},
					},
				},
			},
		}
		record, err := engine.ProcessConsentRequest(createReq)
		assert.NoError(t, err)
		originalFields := record.Fields

		// Update with empty fields - should use existing
		updateReq := models.UpdateConsentRequest{
			Status: models.StatusApproved,
			Fields: []string{}, // Empty - should use existing
		}
		updated, err := engine.UpdateConsent(record.ConsentID, updateReq)
		assert.NoError(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, originalFields, updated.Fields)
	})
}

// TestPostgresConsentEngine_GetConsentsByDataOwner tests GetConsentsByDataOwner
func TestPostgresConsentEngine_GetConsentsByDataOwner(t *testing.T) {
	testEngine := SetupPostgresTestEngineWithDB(t)
	engine := testEngine.Engine

	// Create multiple consents for the same owner
	ownerID := "owner@example.com"
	for i := 0; i < 3; i++ {
		createReq := models.ConsentRequest{
			AppID: fmt.Sprintf("test-app-%d", i),
			ConsentRequirements: []models.ConsentRequirement{
				{
					Owner:   "CITIZEN",
					OwnerID: ownerID,
					Fields: []models.ConsentField{
						{
							FieldName: "personInfo.name",
							SchemaID:  "schema-123",
						},
					},
				},
			},
		}
		_, err := engine.ProcessConsentRequest(createReq)
		assert.NoError(t, err)
	}

	// Get all consents for this owner
	records, err := engine.GetConsentsByDataOwner(ownerID)
	assert.NoError(t, err)
	assert.Len(t, records, 3)
	for _, record := range records {
		assert.Equal(t, ownerID, record.OwnerID)
	}

	// Test with non-existent owner
	emptyRecords, err := engine.GetConsentsByDataOwner("non-existent@example.com")
	assert.NoError(t, err)
	assert.Len(t, emptyRecords, 0)
}

// TestPostgresConsentEngine_GetConsentsByConsumer tests GetConsentsByConsumer
func TestPostgresConsentEngine_GetConsentsByConsumer(t *testing.T) {
	testEngine := SetupPostgresTestEngineWithDB(t)
	engine := testEngine.Engine

	// Create multiple consents for the same consumer app
	appID := "consumer-app"
	for i := 0; i < 3; i++ {
		createReq := models.ConsentRequest{
			AppID: appID,
			ConsentRequirements: []models.ConsentRequirement{
				{
					Owner:   "CITIZEN",
					OwnerID: fmt.Sprintf("owner-%d@example.com", i),
					Fields: []models.ConsentField{
						{
							FieldName: "personInfo.name",
							SchemaID:  "schema-123",
						},
					},
				},
			},
		}
		_, err := engine.ProcessConsentRequest(createReq)
		assert.NoError(t, err)
	}

	// Get all consents for this consumer
	records, err := engine.GetConsentsByConsumer(appID)
	assert.NoError(t, err)
	assert.Len(t, records, 3)
	for _, record := range records {
		assert.Equal(t, appID, record.AppID)
	}

	// Test with non-existent consumer
	emptyRecords, err := engine.GetConsentsByConsumer("non-existent-app")
	assert.NoError(t, err)
	assert.Len(t, emptyRecords, 0)
}

// TestPostgresConsentEngine_ProcessConsentRequest_UpdateExistingPending tests updating existing pending consent
func TestPostgresConsentEngine_ProcessConsentRequest_UpdateExistingPending(t *testing.T) {
	testEngine := SetupPostgresTestEngineWithDB(t)
	engine := testEngine.Engine

	// Create initial pending consent
	createReq := models.ConsentRequest{
		AppID: "test-app-pending",
		ConsentRequirements: []models.ConsentRequirement{
			{
				Owner:   "CITIZEN",
				OwnerID: "user-pending@example.com",
				Fields: []models.ConsentField{
					{
						FieldName: "personInfo.name",
						SchemaID:  "schema-123",
					},
				},
			},
		},
	}
	firstRecord, err := engine.ProcessConsentRequest(createReq)
	assert.NoError(t, err)
	firstConsentID := firstRecord.ConsentID

	// Process same request again - should update existing pending consent
	updateReq := models.ConsentRequest{
		AppID: "test-app-pending",
		ConsentRequirements: []models.ConsentRequirement{
			{
				Owner:   "CITIZEN",
				OwnerID: "user-pending@example.com",
				Fields: []models.ConsentField{
					{
						FieldName: "personInfo.email",
						SchemaID:  "schema-456",
					},
				},
			},
		},
	}
	updatedRecord, err := engine.ProcessConsentRequest(updateReq)
	assert.NoError(t, err)
	// Should return the same consent ID (updated)
	assert.Equal(t, firstConsentID, updatedRecord.ConsentID)
	// Fields should be updated
	assert.Contains(t, updatedRecord.Fields, "personInfo.email")
}

// TestPostgresConsentEngine_ProcessConsentRequest_UpdateExistingNonPending tests updating existing non-pending consent
func TestPostgresConsentEngine_ProcessConsentRequest_UpdateExistingNonPending(t *testing.T) {
	testEngine := SetupPostgresTestEngineWithDB(t)
	engine := testEngine.Engine

	// Create and approve a consent
	createReq := models.ConsentRequest{
		AppID: "test-app-approved",
		ConsentRequirements: []models.ConsentRequirement{
			{
				Owner:   "CITIZEN",
				OwnerID: "user-approved@example.com",
				Fields: []models.ConsentField{
					{
						FieldName: "personInfo.name",
						SchemaID:  "schema-123",
					},
				},
			},
		},
	}
	record, err := engine.ProcessConsentRequest(createReq)
	assert.NoError(t, err)

	// Approve it
	approveReq := models.ConsentPortalRequest{
		ConsentID: record.ConsentID,
		Action:    "approve",
		DataOwner: "user-approved@example.com",
	}
	_, err = engine.ProcessConsentPortalRequest(approveReq)
	assert.NoError(t, err)

	// Process same request again - should update existing approved consent
	updateReq := models.ConsentRequest{
		AppID: "test-app-approved",
		ConsentRequirements: []models.ConsentRequirement{
			{
				Owner:   "CITIZEN",
				OwnerID: "user-approved@example.com",
				Fields: []models.ConsentField{
					{
						FieldName: "personInfo.email",
						SchemaID:  "schema-456",
					},
				},
			},
		},
	}
	updatedRecord, err := engine.ProcessConsentRequest(updateReq)
	assert.NoError(t, err)
	// Should return the same consent ID (updated)
	assert.Equal(t, record.ConsentID, updatedRecord.ConsentID)
	// Fields should be updated
	assert.Contains(t, updatedRecord.Fields, "personInfo.email")
}
