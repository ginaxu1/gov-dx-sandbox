package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToConsentPortalView(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name     string
		record   *ConsentRecord
		validate func(t *testing.T, view *ConsentPortalView)
	}{
		{
			name: "BasicConversion",
			record: &ConsentRecord{
				ConsentID:  "consent_123",
				OwnerID:   "user-example-com",
				OwnerEmail: "user@example.com",
				AppID:      "test-app",
				Status:     StatusPending,
				Type:       "realtime",
				CreatedAt:  now,
				Fields:     []string{"personInfo.name"},
			},
			validate: func(t *testing.T, view *ConsentPortalView) {
				assert.Equal(t, "Test App", view.AppDisplayName)
				assert.Equal(t, "User Example Com", view.OwnerName)
				assert.Equal(t, "user@example.com", view.OwnerEmail)
				assert.Equal(t, "pending", view.Status)
				assert.Equal(t, "realtime", view.Type)
				assert.Equal(t, []string{"personInfo.name"}, view.Fields)
				assert.Equal(t, now, view.CreatedAt)
			},
		},
		{
			name: "ApprovedStatus",
			record: &ConsentRecord{
				ConsentID:  "consent_456",
				OwnerID:   "john-doe",
				OwnerEmail: "john@example.com",
				AppID:      "my-application",
				Status:     StatusApproved,
				Type:       "offline",
				CreatedAt:  now,
				Fields:     []string{"personInfo.name", "personInfo.address"},
			},
			validate: func(t *testing.T, view *ConsentPortalView) {
				assert.Equal(t, "My Application", view.AppDisplayName)
				assert.Equal(t, "John Doe", view.OwnerName)
				assert.Equal(t, "approved", view.Status)
				assert.Equal(t, "offline", view.Type)
				assert.Len(t, view.Fields, 2)
			},
		},
		{
			name: "EmptyFields",
			record: &ConsentRecord{
				ConsentID:  "consent_789",
				OwnerID:   "test-user",
				OwnerEmail: "test@example.com",
				AppID:      "app",
				Status:     StatusPending,
				Type:       "realtime",
				CreatedAt:  now,
				Fields:     []string{},
			},
			validate: func(t *testing.T, view *ConsentPortalView) {
				assert.NotNil(t, view.Fields)
				assert.Len(t, view.Fields, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := tt.record.ToConsentPortalView()
			assert.NotNil(t, view)
			tt.validate(t, view)
		})
	}
}

func TestToConsentResponse(t *testing.T) {
	portalURL := "http://localhost:3000/?consent_id=consent_123"
	
	tests := []struct {
		name     string
		record   *ConsentRecord
		validate func(t *testing.T, response ConsentResponse)
	}{
		{
			name: "PendingWithPortalURL",
			record: &ConsentRecord{
				ConsentID:        "consent_123",
				Status:           StatusPending,
				ConsentPortalURL: portalURL,
			},
			validate: func(t *testing.T, response ConsentResponse) {
				assert.Equal(t, "consent_123", response.ConsentID)
				assert.Equal(t, "pending", response.Status)
				assert.NotNil(t, response.ConsentPortalURL)
				assert.Equal(t, portalURL, *response.ConsentPortalURL)
			},
		},
		{
			name: "PendingWithoutPortalURL",
			record: &ConsentRecord{
				ConsentID:        "consent_456",
				Status:           StatusPending,
				ConsentPortalURL: "",
			},
			validate: func(t *testing.T, response ConsentResponse) {
				assert.Equal(t, "consent_456", response.ConsentID)
				assert.Equal(t, "pending", response.Status)
				assert.Nil(t, response.ConsentPortalURL)
			},
		},
		{
			name: "ApprovedStatus",
			record: &ConsentRecord{
				ConsentID:        "consent_789",
				Status:           StatusApproved,
				ConsentPortalURL: portalURL,
			},
			validate: func(t *testing.T, response ConsentResponse) {
				assert.Equal(t, "consent_789", response.ConsentID)
				assert.Equal(t, "approved", response.Status)
				assert.Nil(t, response.ConsentPortalURL)
			},
		},
		{
			name: "RejectedStatus",
			record: &ConsentRecord{
				ConsentID:        "consent_abc",
				Status:           StatusRejected,
				ConsentPortalURL: portalURL,
			},
			validate: func(t *testing.T, response ConsentResponse) {
				assert.Equal(t, "consent_abc", response.ConsentID)
				assert.Equal(t, "rejected", response.Status)
				assert.Nil(t, response.ConsentPortalURL)
			},
		},
		{
			name: "RevokedStatus",
			record: &ConsentRecord{
				ConsentID:        "consent_def",
				Status:           StatusRevoked,
				ConsentPortalURL: portalURL,
			},
			validate: func(t *testing.T, response ConsentResponse) {
				assert.Equal(t, "consent_def", response.ConsentID)
				assert.Equal(t, "revoked", response.Status)
				assert.Nil(t, response.ConsentPortalURL)
			},
		},
		{
			name: "ExpiredStatus",
			record: &ConsentRecord{
				ConsentID:        "consent_ghi",
				Status:           StatusExpired,
				ConsentPortalURL: portalURL,
			},
			validate: func(t *testing.T, response ConsentResponse) {
				assert.Equal(t, "consent_ghi", response.ConsentID)
				assert.Equal(t, "expired", response.Status)
				assert.Nil(t, response.ConsentPortalURL)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := tt.record.ToConsentResponse()
			tt.validate(t, response)
		})
	}
}

