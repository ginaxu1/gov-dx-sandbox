package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/exchange/policy-decision-point/v1/models"
	"gorm.io/gorm"
)

// PolicyMetadataService provides business logic for policy metadata operations
type PolicyMetadataService struct {
	db *gorm.DB
}

// NewPolicyMetadataService creates a new policy metadata service
func NewPolicyMetadataService(db *gorm.DB) *PolicyMetadataService {
	return &PolicyMetadataService{
		db: db,
	}
}

// CreatePolicyMetadata creates new policy metadata records with validation
func (s *PolicyMetadataService) CreatePolicyMetadata(req *models.PolicyMetadataCreateRequest) (*models.PolicyMetadataCreateResponse, error) {
	// Check if there are already records for the given schema ID
	var existingMetadata []models.PolicyMetadata
	if err := s.db.Where("schema_id = ?", req.SchemaID).Find(&existingMetadata).Error; err != nil {
		return nil, fmt.Errorf("failed to check existing policy metadata: %w", err)
	}

	// Create a map for faster lookups of existing records by field name
	existingMap := make(map[string]*models.PolicyMetadata)
	for i := range existingMetadata {
		metadata := existingMetadata[i]
		existingMap[metadata.FieldName] = &metadata
	}

	now := time.Now()
	var newRecords []models.PolicyMetadata
	var updatedRecords []models.PolicyMetadata
	processedFields := make(map[string]bool)

	// Process incoming records
	for _, record := range req.Records {
		processedFields[record.FieldName] = true

		if existing, exists := existingMap[record.FieldName]; exists {
			// Update existing record with all fields
			existing.DisplayName = record.DisplayName
			existing.Description = record.Description
			existing.Source = record.Source
			existing.IsOwner = record.IsOwner
			existing.AccessControlType = record.AccessControlType
			existing.Owner = record.Owner
			existing.UpdatedAt = now

			if err := s.db.Save(existing).Error; err != nil {
				return nil, fmt.Errorf("failed to update existing policy metadata: %w", err)
			}
			updatedRecords = append(updatedRecords, *existing)
		} else {
			// Create new record
			policyMetadata := models.PolicyMetadata{
				ID:                uuid.New(),
				SchemaID:          req.SchemaID,
				FieldName:         record.FieldName,
				DisplayName:       record.DisplayName,
				Description:       record.Description,
				Source:            record.Source,
				IsOwner:           record.IsOwner,
				AccessControlType: record.AccessControlType,
				AllowList:         make(models.AllowList),
				Owner:             record.Owner,
				CreatedAt:         now,
				UpdatedAt:         now,
			}
			newRecords = append(newRecords, policyMetadata)
		}
	}

	// Delete records that weren't in the request (obsolete records)
	var idsToDelete []uuid.UUID
	for fieldName, existing := range existingMap {
		if !processedFields[fieldName] {
			idsToDelete = append(idsToDelete, existing.ID)
		}
	}

	if len(idsToDelete) > 0 {
		if err := s.db.Where("id IN ?", idsToDelete).Delete(&models.PolicyMetadata{}).Error; err != nil {
			return nil, fmt.Errorf("failed to delete obsolete policy metadata records: %w", err)
		}
	}

	// Bulk create new records
	if len(newRecords) > 0 {
		if err := s.db.Create(&newRecords).Error; err != nil {
			return nil, fmt.Errorf("failed to create policy metadata records: %w", err)
		}
	}

	// Prepare response including both new and updated records
	var responseRecords []models.PolicyMetadataResponse

	// Add new records to response
	for _, pm := range newRecords {
		responseRecord := models.PolicyMetadataResponse{
			ID:                pm.ID.String(),
			SchemaID:          pm.SchemaID,
			FieldName:         pm.FieldName,
			DisplayName:       pm.DisplayName,
			Description:       pm.Description,
			Source:            pm.Source,
			IsOwner:           pm.IsOwner,
			AccessControlType: pm.AccessControlType,
			AllowList:         pm.AllowList,
			Owner:             pm.Owner,
			CreatedAt:         pm.CreatedAt.Format(time.RFC3339),
			UpdatedAt:         pm.UpdatedAt.Format(time.RFC3339),
		}
		responseRecords = append(responseRecords, responseRecord)
	}

	// Add updated records to response
	for _, pm := range updatedRecords {
		responseRecord := models.PolicyMetadataResponse{
			ID:                pm.ID.String(),
			SchemaID:          pm.SchemaID,
			FieldName:         pm.FieldName,
			DisplayName:       pm.DisplayName,
			Description:       pm.Description,
			Source:            pm.Source,
			IsOwner:           pm.IsOwner,
			AccessControlType: pm.AccessControlType,
			AllowList:         pm.AllowList,
			Owner:             pm.Owner,
			CreatedAt:         pm.CreatedAt.Format(time.RFC3339),
			UpdatedAt:         pm.UpdatedAt.Format(time.RFC3339),
		}
		responseRecords = append(responseRecords, responseRecord)
	}

	return &models.PolicyMetadataCreateResponse{
		Records: responseRecords,
	}, nil
}

// UpdateAllowList updates the allow list for multiple fields with validation
func (s *PolicyMetadataService) UpdateAllowList(req *models.AllowListUpdateRequest) (*models.AllowListUpdateResponse, error) {
	var responseRecords []models.AllowListUpdateResponseRecord

	for _, record := range req.Records {
		var pm models.PolicyMetadata
		if err := s.db.Where("schema_id = ? AND field_name = ?", record.SchemaID, record.FieldName).First(&pm).Error; err != nil {
			return nil, fmt.Errorf("policy metadata not found for schema_id %s and field_name %s: %w", record.SchemaID, record.FieldName, err)
		}

		// Calculate expiration time based on grant duration
		var currentTime = time.Now()
		var expiresAt time.Time
		switch req.GrantDuration {
		case models.GrantDurationTypeOneMonth:
			expiresAt = currentTime.AddDate(0, 1, 0)
		case models.GrantDurationTypeOneYear:
			expiresAt = currentTime.AddDate(1, 0, 0)
		default:
			return nil, fmt.Errorf("invalid grant duration: %s", req.GrantDuration)
		}

		// Update allow list
		if pm.AllowList == nil {
			pm.AllowList = make(models.AllowList)
		}
		pm.AllowList[req.ApplicationID] = models.AllowListEntry{
			ExpiresAt: expiresAt,
			UpdatedAt: currentTime,
		}

		if err := s.db.Save(&pm).Error; err != nil {
			return nil, fmt.Errorf("failed to update allow list for schema_id %s and field_name %s: %w", record.SchemaID, record.FieldName, err)
		}

		// Prepare response record
		responseRecord := models.AllowListUpdateResponseRecord{
			FieldName: record.FieldName,
			SchemaID:  record.SchemaID,
			ExpiresAt: expiresAt.Format(time.RFC3339),
			UpdatedAt: currentTime.Format(time.RFC3339),
		}
		responseRecords = append(responseRecords, responseRecord)
	}

	return &models.AllowListUpdateResponse{
		Records: responseRecords,
	}, nil
}

// GetPolicyDecision evaluates policy decision based on policy metadata
func (s *PolicyMetadataService) GetPolicyDecision(req *models.PolicyDecisionRequest) (*models.PolicyDecisionResponse, error) {
	var consentRequiredFields []models.PolicyDecisionResponseFieldRecord
	var unAuthorizedFields []models.PolicyDecisionResponseFieldRecord
	var expiredFields []models.PolicyDecisionResponseFieldRecord

	for _, record := range req.RequiredFields {
		var pm models.PolicyMetadata
		if err := s.db.Where("schema_id = ? AND field_name = ?", record.SchemaID, record.FieldName).First(&pm).Error; err != nil {
			return nil, fmt.Errorf("policy metadata not found for schema_id %s and field_name %s: %w", record.SchemaID, record.FieldName, err)
		}

		// Check if application is authorized
		if _, exists := pm.AllowList[req.ApplicationID]; !exists {
			unAuthorizedFields = append(unAuthorizedFields, models.PolicyDecisionResponseFieldRecord{
				FieldName:   pm.FieldName,
				SchemaID:    pm.SchemaID,
				DisplayName: pm.DisplayName,
				Description: pm.Description,
				Owner:       pm.Owner,
			})
			continue
		}
		// Check if access has expired
		allowListEntry := pm.AllowList[req.ApplicationID]
		if time.Now().After(allowListEntry.ExpiresAt) {
			expiredFields = append(expiredFields, models.PolicyDecisionResponseFieldRecord{
				FieldName:   pm.FieldName,
				SchemaID:    pm.SchemaID,
				DisplayName: pm.DisplayName,
				Description: pm.Description,
				Owner:       pm.Owner,
			})
			continue
		}
		// Check if owner consent is required
		if !pm.IsOwner && pm.AccessControlType == models.AccessControlTypeRestricted {
			consentRequiredFields = append(consentRequiredFields, models.PolicyDecisionResponseFieldRecord{
				FieldName:   pm.FieldName,
				SchemaID:    pm.SchemaID,
				DisplayName: pm.DisplayName,
				Description: pm.Description,
				Owner:       pm.Owner,
			})
		}
	}

	response := &models.PolicyDecisionResponse{
		ConsentRequiredFields:   consentRequiredFields,
		UnAuthorizedFields:      unAuthorizedFields,
		ExpiredFields:           expiredFields,
		AppNotAuthorized:        len(unAuthorizedFields) > 0,
		AppAccessExpired:        len(expiredFields) > 0,
		AppRequiresOwnerConsent: len(consentRequiredFields) > 0,
	}

	return response, nil
}
