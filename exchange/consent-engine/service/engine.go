package service

import (
	"context"
	"time"

	"github.com/gov-dx-sandbox/exchange/consent-engine/v1/models"
)

// ConsentEngine defines the interface for consent management operations
type ConsentEngine interface {
	// CreateConsent creates a new consent record and stores it
	CreateConsent(req models.ConsentRequest) (*models.ConsentRecord, error)
	// FindExistingConsent finds an existing consent record by consumer app ID and owner ID
	FindExistingConsent(consumerAppID, ownerID string) *models.ConsentRecord
	// GetConsentStatus retrieves a consent record by its ID
	GetConsentStatus(id string) (*models.ConsentRecord, error)
	// UpdateConsent updates the status of a consent record
	UpdateConsent(id string, req models.UpdateConsentRequest) (*models.ConsentRecord, error)
	// ProcessConsentPortalRequest handles consent portal interactions
	ProcessConsentPortalRequest(req models.ConsentPortalRequest) (*models.ConsentRecord, error)
	// GetConsentsByDataOwner retrieves all consent records for a data owner
	GetConsentsByDataOwner(dataOwner string) ([]*models.ConsentRecord, error)
	// GetConsentsByConsumer retrieves all consent records for a consumer
	GetConsentsByConsumer(consumer string) ([]*models.ConsentRecord, error)
	// CheckConsentExpiry checks and updates expired consent records
	CheckConsentExpiry() ([]*models.ConsentRecord, error)
	// StartBackgroundExpiryProcess starts the background process for checking consent expiry
	StartBackgroundExpiryProcess(ctx context.Context, interval time.Duration)
	// RevokeConsent revokes a consent record
	RevokeConsent(id string, reason string) (*models.ConsentRecord, error)
	// ProcessConsentRequest processes the new consent workflow request
	ProcessConsentRequest(req models.ConsentRequest) (*models.ConsentRecord, error)
	// StopBackgroundExpiryProcess stops the background expiry process
	StopBackgroundExpiryProcess()
}
