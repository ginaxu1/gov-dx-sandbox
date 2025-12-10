package services

import (
	"context"
	"errors"
	"fmt"
	"net/url"
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
func NewConsentService(db *gorm.DB, consentPortalBaseURL string) (*ConsentService, error) {
	parsed, err := url.Parse(consentPortalBaseURL)
	if consentPortalBaseURL == "" || err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid consentPortalBaseURL: must be a non-empty, valid URL with scheme and host")
	}
	return &ConsentService{
		db:                   db,
		consentPortalBaseURL: consentPortalBaseURL,
	}, nil
}

// CreateConsentRecord creates a new consent record in the database
func (s *ConsentService) CreateConsentRecord(ctx context.Context, req models.CreateConsentRequest) ([]models.ConsentResponseInternalView, error) {
	// Validate input
	if err := validateCreateConsentRequest(req); err != nil {
		return nil, fmt.Errorf("%w: %w", models.ErrConsentCreateFailed, err)
	}

	consentRecords := make([]models.ConsentRecord, 0, len(req.ConsentRequirements))
	// Iterate over consent requirements to create consent records
	for _, requirement := range req.ConsentRequirements {
		consentID := uuid.New()
		currentTime := time.Now().UTC()
		consentType := models.TypeRealtime

		// Calculate pending expiry time based on consent type
		pendingExpiresAt := currentTime.Add(parsePendingTimeoutDuration(consentType))

		consentRecord := models.ConsentRecord{
			ConsentID:        consentID,
			OwnerID:          requirement.OwnerID,
			OwnerEmail:       requirement.OwnerEmail,
			AppID:            req.AppID,
			AppName:          req.AppName,
			Status:           string(models.StatusPending),
			Type:             string(consentType),
			CreatedAt:        currentTime,
			UpdatedAt:        currentTime,
			GrantDuration:    string(getGrantDurationOrDefault((*models.GrantDuration)(req.GrantDuration))),
			Fields:           requirement.Fields,
			ConsentPortalURL: fmt.Sprintf("%s?consentId=%s", s.consentPortalBaseURL, consentID.String()),
			PendingExpiresAt: &pendingExpiresAt,
		}
		consentRecords = append(consentRecords, consentRecord)
	}

	// Bulk insert consent records
	if err := s.db.WithContext(ctx).Create(&consentRecords).Error; err != nil {
		return nil, fmt.Errorf("%w: %w", models.ErrConsentCreateFailed, err)
	}

	// Convert to internal view responses
	responses := make([]models.ConsentResponseInternalView, 0, len(consentRecords))
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

// GetConsentInternalView retrieves a consent record by ID or by ((ownerID OR ownerEmail) AND appID) and returns its internal view
func (s *ConsentService) GetConsentInternalView(ctx context.Context, consentID *string, ownerID *string, ownerEmail *string, appID *string) (*models.ConsentResponseInternalView, error) {
	var consentRecord models.ConsentRecord
	query := s.db.WithContext(ctx).Model(&models.ConsentRecord{})

	if consentID != nil {
		parsedConsentID, err := uuid.Parse(*consentID)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid consent ID", models.ErrConsentGetFailed)
		}
		query = query.Where("consent_id = ?", parsedConsentID)
	} else if ownerID != nil && appID != nil {
		// OwnerID gets priority over OwnerEmail if both are provided
		query = query.Where("owner_id = ? AND app_id = ?", *ownerID, *appID)
	} else if ownerEmail != nil && appID != nil {
		// NOTE: This query on (owner_email, app_id) does not benefit from a composite index.
		// If this query pattern is common or the table is large, consider adding a composite index on (owner_email, app_id)
		query = query.Where("owner_email = ? AND app_id = ?", *ownerEmail, *appID)
	} else {
		return nil, fmt.Errorf("%w: either consentID or (ownerID/ownerEmail and appID) must be provided", models.ErrConsentGetFailed)
	}

	if err := query.First(&consentRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %w", models.ErrConsentNotFound, err)
		}
		return nil, fmt.Errorf("%w: %w", models.ErrConsentGetFailed, err)
	}

	internalView := consentRecord.ToConsentResponseInternalView()
	return &internalView, nil
}

// GetConsentPortalView retrieves a consent record by ID and returns its portal view
func (s *ConsentService) GetConsentPortalView(ctx context.Context, consentID string) (*models.ConsentResponsePortalView, error) {
	var consentRecord models.ConsentRecord
	parsedConsentID, err := uuid.Parse(consentID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid consent ID", models.ErrConsentGetFailed)
	}

	if err := s.db.WithContext(ctx).Where("consent_id = ?", parsedConsentID).First(&consentRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %w", models.ErrConsentNotFound, err)
		}
		return nil, fmt.Errorf("%w: %w", models.ErrConsentGetFailed, err)
	}

	portalView := consentRecord.ToConsentResponsePortalView()
	return &portalView, nil
}

// UpdateConsentStatusByPortalAction updates the consent status based on user action from the consent portal
func (s *ConsentService) UpdateConsentStatusByPortalAction(ctx context.Context, req models.ConsentPortalActionRequest) error {
	// Validate action
	if !isValidConsentPortalAction(req.Action) {
		return fmt.Errorf("%w: invalid action: %s", models.ErrPortalRequestFailed, req.Action)
	}

	var consentRecord models.ConsentRecord
	parsedConsentID, err := uuid.Parse(req.ConsentID)
	if err != nil {
		return fmt.Errorf("%w: invalid consent ID", models.ErrPortalRequestFailed)
	}

	if err := s.db.WithContext(ctx).Where("consent_id = ?", parsedConsentID).First(&consentRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: %w", models.ErrConsentNotFound, err)
		}
		return fmt.Errorf("%w: %w", models.ErrConsentUpdateFailed, err)
	}

	currentTime := time.Now().UTC()
	consentRecord.UpdatedAt = currentTime
	consentRecord.UpdatedBy = &req.UpdatedBy

	switch req.Action {
	case models.ActionApprove:
		consentRecord.Status = string(models.StatusApproved)
		grantExpiresAt := currentTime.Add(parseGrantDuration((models.GrantDuration)(consentRecord.GrantDuration)))
		consentRecord.GrantExpiresAt = &grantExpiresAt
		consentRecord.PendingExpiresAt = nil
	case models.ActionReject:
		consentRecord.Status = string(models.StatusRejected)
		// Do not set GrantExpiresAt on rejection - only approval gets a grant expiry
		consentRecord.PendingExpiresAt = nil
	default:
		return fmt.Errorf("%w: invalid action: %s", models.ErrPortalRequestFailed, req.Action)
	}

	if err := s.db.WithContext(ctx).Save(&consentRecord).Error; err != nil {
		return fmt.Errorf("%w: %w", models.ErrConsentUpdateFailed, err)
	}

	return nil
}

// parseGrantDuration parses the grant duration string into a time.Duration
func parseGrantDuration(grantDuration models.GrantDuration) time.Duration {
	switch grantDuration {
	case models.DurationOneHour:
		return time.Hour
	case models.DurationSixHours:
		return 6 * time.Hour
	case models.DurationTwelveHours:
		return 12 * time.Hour
	case models.DurationOneDay:
		return 24 * time.Hour
	case models.DurationSevenDays:
		return 7 * 24 * time.Hour
	case models.DurationThirtyDays:
		return 30 * 24 * time.Hour
	default:
		return time.Hour // Default to 1 hour if unrecognized
	}
}

// parsePendingTimeoutDuration returns the pending timeout duration based on consent type
func parsePendingTimeoutDuration(consentType models.ConsentType) time.Duration {
	switch consentType {
	case models.TypeRealtime:
		return time.Hour // 1 hour for realtime
	case models.TypeOffline:
		return 24 * time.Hour // 1 day for offline
	default:
		return time.Hour // Default to 1 hour if unrecognized
	}
}

// validateCreateConsentRequest validates the create consent request input
func validateCreateConsentRequest(req models.CreateConsentRequest) error {
	if req.AppID == "" {
		return errors.New("appId is required")
	}

	if len(req.ConsentRequirements) == 0 {
		return errors.New("consentRequirements cannot be empty")
	}

	// Validate grant duration if provided
	if req.GrantDuration != nil && *req.GrantDuration != "" {
		if !isValidGrantDuration(models.GrantDuration(*req.GrantDuration)) {
			return fmt.Errorf("invalid grantDuration: %s", *req.GrantDuration)
		}
	}

	for i, requirement := range req.ConsentRequirements {
		if requirement.OwnerID == "" {
			return fmt.Errorf("consentRequirements[%d].ownerId is required", i)
		}
		if requirement.OwnerEmail == "" {
			return fmt.Errorf("consentRequirements[%d].ownerEmail is required", i)
		}
		if len(requirement.Fields) == 0 {
			return fmt.Errorf("consentRequirements[%d].fields cannot be empty", i)
		}

		// Validate each field
		for j, field := range requirement.Fields {
			if field.FieldName == "" {
				return fmt.Errorf("consentRequirements[%d].fields[%d].fieldName is required", i, j)
			}
			if field.SchemaID == "" {
				return fmt.Errorf("consentRequirements[%d].fields[%d].schemaId is required", i, j)
			}
		}
	}

	return nil
}

// isValidGrantDuration checks if a grant duration is valid
func isValidGrantDuration(grantDuration models.GrantDuration) bool {
	switch grantDuration {
	case models.DurationOneHour,
		models.DurationSixHours,
		models.DurationTwelveHours,
		models.DurationOneDay,
		models.DurationSevenDays,
		models.DurationThirtyDays:
		return true
	default:
		return false
	}
}

// isValidConsentPortalAction checks if a consent portal action is valid
func isValidConsentPortalAction(action models.ConsentPortalAction) bool {
	switch action {
	case models.ActionApprove, models.ActionReject:
		return true
	default:
		return false
	}
}
