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
	// Prepare policy metadata records
	var policyMetadataList []models.PolicyMetadata
	now := time.Now()
	for _, record := range req.Records {
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
		policyMetadataList = append(policyMetadataList, policyMetadata)
	}

	if err := s.db.Create(&policyMetadataList).Error; err != nil {
		return nil, fmt.Errorf("failed to create policy metadata records: %w", err)
	}

	// Prepare response
	var responseRecords []models.PolicyMetadataResponse
	for _, pm := range policyMetadataList {
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
		var expiresAt time.Time
		switch req.GrantDuration {
		case models.GrantDurationTypeOneMonth:
			expiresAt = time.Now().AddDate(0, 1, 0)
		case models.GrantDurationTypeOneYear:
			expiresAt = time.Now().AddDate(1, 0, 0)
		default:
			return nil, fmt.Errorf("invalid grant duration: %s", req.GrantDuration)
		}

		// Update allow list
		if pm.AllowList == nil {
			pm.AllowList = make(models.AllowList)
		}
		pm.AllowList[req.ApplicationID] = models.AllowListEntry{
			ExpiresAt: expiresAt,
			UpdatedAt: time.Now(),
		}

		if err := s.db.Save(&pm).Error; err != nil {
			return nil, fmt.Errorf("failed to update allow list for schema_id %s and field_name %s: %w", record.SchemaID, record.FieldName, err)
		}

		// Prepare response record
		responseRecord := models.AllowListUpdateResponseRecord{
			FieldName: record.FieldName,
			SchemaID:  record.SchemaID,
			ExpiresAt: expiresAt.Format(time.RFC3339),
			UpdatedAt: time.Now().Format(time.RFC3339),
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
