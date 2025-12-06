package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/exchange/consent-engine/v1/models"
	"gorm.io/gorm"
)

// ConsentService provides business logic for consent operations
type ConsentService struct {
	db                   *gorm.DB
	consentPortalBaseURL string
}

// NewConsentService creates a new consent service
func NewConsentService(db *gorm.DB, consentPortalBaseURL string) *ConsentService {
	return &ConsentService{
		db:                   db,
		consentPortalBaseURL: consentPortalBaseURL,
	}
}

// CreateConsentRecord creates a new consent record in the database
func (s *ConsentService) CreateConsentRecord(req models.CreateConsentRequest) ([]models.ConsentResponseInternalView, error) {
	var consentRecords []models.ConsentRecord
	// Iterate over consent requirements to create consent records
	for _, requirement := range req.ConsentRequirements {
		consentID := uuid.New()
		currentTime := time.Now().UTC()
		consentRecord := models.ConsentRecord{
			ConsentID:        consentID,
			OwnerID:          requirement.OwnerID,
			OwnerEmail:       requirement.OwnerEmail,
			AppID:            req.AppID,
			AppName:          req.AppName,
			Status:           string(models.StatusPending),
			Type:             string(models.TypeRealtime),
			CreatedAt:        currentTime,
			UpdatedAt:        currentTime,
			GrantDuration:    string(getGrantDurationOrDefault((*models.GrantDuration)(req.GrantDuration))),
			Fields:           requirement.Fields,
			ConsentPortalURL: fmt.Sprintf("%s?%s", s.consentPortalBaseURL, consentID.String()),
		}
		consentRecords = append(consentRecords, consentRecord)
	}

	// Bulk insert consent records
	if err := s.db.Create(&consentRecords).Error; err != nil {
		return nil, fmt.Errorf("failed to create consent records: %w", err)
	}

	// Convert to internal view responses
	var responses []models.ConsentResponseInternalView
	for _, record := range consentRecords {
		responses = append(responses, record.ToConsentResponseInternalView())
	}

	return responses, nil
}

// getGrantDurationOrDefault returns the provided grant duration or the default if empty
func getGrantDurationOrDefault(grantDuration *models.GrantDuration) models.GrantDuration {
	if grantDuration == nil || *grantDuration == "" {
		return models.DurationDefault
	}
	return *grantDuration
}

// GetConsentInternalView retrieves a consent record by ID or by (ownerID, appID) and returns its internal view
func (s *ConsentService) GetConsentInternalView(consentID *string, ownerID *string, appID *string) (*models.ConsentResponseInternalView, error) {
	var consentRecord models.ConsentRecord
	query := s.db.Model(&models.ConsentRecord{})

	if consentID != nil {
		parsedConsentID, err := uuid.Parse(*consentID)
		if err != nil {
			return nil, fmt.Errorf("invalid consent ID format: %w", err)
		}
		query = query.Where("consent_id = ?", parsedConsentID)
	} else if ownerID != nil && appID != nil {
		query = query.Where("owner_id = ? AND app_id = ?", *ownerID, *appID)
	} else {
		return nil, errors.New("either consentID or both ownerID and appID must be provided")
	}

	if err := query.First(&consentRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Consent record not found
		}
		return nil, fmt.Errorf("failed to retrieve consent record: %w", err)
	}

	internalView := consentRecord.ToConsentResponseInternalView()
	return &internalView, nil
}

// GetConsentPortalView retrieves a consent record by ID and returns its portal view
func (s *ConsentService) GetConsentPortalView(consentID string) (*models.ConsentResponsePortalView, error) {
	var consentRecord models.ConsentRecord
	parsedConsentID, err := uuid.Parse(consentID)
	if err != nil {
		return nil, fmt.Errorf("invalid consent ID format: %w", err)
	}

	if err := s.db.Where("consent_id = ?", parsedConsentID).First(&consentRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Consent record not found
		}
		return nil, fmt.Errorf("failed to retrieve consent record: %w", err)
	}

	portalView := consentRecord.ToConsentResponsePortalView()
	return &portalView, nil
}
