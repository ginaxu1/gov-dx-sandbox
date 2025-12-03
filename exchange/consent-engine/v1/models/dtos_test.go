package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestConsentRecord_ToConsentResponse(t *testing.T) {
	t.Run("Pending status with portal URL", func(t *testing.T) {
		portalURL := "https://portal.example.com/consent/123"
		cr := &ConsentRecord{
			ConsentID:        uuid.New(),
			Status:           string(StatusPending),
			ConsentPortalURL: portalURL,
		}

		response := cr.ToConsentResponse()

		assert.Equal(t, cr.ConsentID, response.ConsentID)
		assert.Equal(t, string(StatusPending), response.Status)
		assert.NotNil(t, response.ConsentPortalURL)
		assert.Equal(t, portalURL, *response.ConsentPortalURL)
	})

	t.Run("Pending status without portal URL", func(t *testing.T) {
		cr := &ConsentRecord{
			ConsentID:        uuid.New(),
			Status:           string(StatusPending),
			ConsentPortalURL: "",
		}

		response := cr.ToConsentResponse()

		assert.Equal(t, cr.ConsentID, response.ConsentID)
		assert.Equal(t, string(StatusPending), response.Status)
		assert.Nil(t, response.ConsentPortalURL)
	})

	t.Run("Approved status with portal URL", func(t *testing.T) {
		portalURL := "https://portal.example.com/consent/123"
		cr := &ConsentRecord{
			ConsentID:        uuid.New(),
			Status:           string(StatusApproved),
			ConsentPortalURL: portalURL,
		}

		response := cr.ToConsentResponse()

		assert.Equal(t, cr.ConsentID, response.ConsentID)
		assert.Equal(t, string(StatusApproved), response.Status)
		assert.Nil(t, response.ConsentPortalURL) // Should not include URL for non-pending status
	})

	t.Run("Rejected status", func(t *testing.T) {
		cr := &ConsentRecord{
			ConsentID: uuid.New(),
			Status:    string(StatusRejected),
		}

		response := cr.ToConsentResponse()

		assert.Equal(t, cr.ConsentID, response.ConsentID)
		assert.Equal(t, string(StatusRejected), response.Status)
		assert.Nil(t, response.ConsentPortalURL)
	})
}

func TestConsentRecord_ToConsentPortalView(t *testing.T) {
	t.Run("Basic conversion", func(t *testing.T) {
		now := time.Now()
		cr := &ConsentRecord{
			ConsentID:   uuid.New(),
			AppID:       "my-test-app",
			OwnerID:     "john-doe",
			OwnerEmail:  "john@example.com",
			Status:      string(StatusPending),
			Type:        "realtime",
			CreatedAt:   now,
			Fields: []ConsentField{
				{
					FieldName:   "person.fullName",
					SchemaID:    "schema-123",
					DisplayName: stringPtr("Full Name"),
					Description: stringPtr("Person's full name"),
				},
			},
		}

		view := cr.ToConsentPortalView()

		assert.NotNil(t, view)
		assert.Equal(t, "My Test App", view.AppDisplayName)
		assert.Equal(t, "John Doe", view.OwnerName)
		assert.Equal(t, "john@example.com", view.OwnerEmail)
		assert.Equal(t, string(StatusPending), view.Status)
		assert.Equal(t, "realtime", view.Type)
		assert.Equal(t, now, view.CreatedAt)
		assert.Len(t, view.Fields, 1)
		assert.Equal(t, "person.fullName", view.Fields[0].FieldName)
	})

	t.Run("App ID with underscores", func(t *testing.T) {
		cr := &ConsentRecord{
			AppID:      "my_test_app",
			OwnerID:    "user_123",
			OwnerEmail: "user@example.com",
			Status:     string(StatusApproved),
			Type:       "offline",
			CreatedAt:  time.Now(),
			Fields:     []ConsentField{},
		}

		view := cr.ToConsentPortalView()

		assert.Equal(t, "My_test App", view.AppDisplayName) // Underscores preserved
		assert.Equal(t, "User_123", view.OwnerName)
	})

	t.Run("Multiple fields", func(t *testing.T) {
		cr := &ConsentRecord{
			AppID:      "test-app",
			OwnerID:    "owner-1",
			OwnerEmail: "owner@example.com",
			Status:     string(StatusPending),
			Type:       "realtime",
			CreatedAt:  time.Now(),
			Fields: []ConsentField{
				{FieldName: "field1", SchemaID: "schema-1"},
				{FieldName: "field2", SchemaID: "schema-2"},
				{FieldName: "field3", SchemaID: "schema-3"},
			},
		}

		view := cr.ToConsentPortalView()

		assert.Len(t, view.Fields, 3)
		assert.Equal(t, "field1", view.Fields[0].FieldName)
		assert.Equal(t, "field2", view.Fields[1].FieldName)
		assert.Equal(t, "field3", view.Fields[2].FieldName)
	})
}

func stringPtr(s string) *string {
	return &s
}

