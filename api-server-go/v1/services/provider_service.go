package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"gorm.io/gorm"
)

// ProviderService handles provider-related operations
type ProviderService struct {
	db            *gorm.DB
	entityService *EntityService
}

// NewProviderService creates a new provider service
func NewProviderService(db *gorm.DB, entityService *EntityService) *ProviderService {
	return &ProviderService{db: db, entityService: entityService}
}

// CreateProvider creates a new provider
func (s *ProviderService) CreateProvider(req *models.CreateProviderRequest) (*models.ProviderResponse, error) {
	var entity models.Entity
	if req.EntityID != nil {
		// Verify entity exists
		err := s.db.First(&entity, "entity_id = ?", req.EntityID).Error
		if err != nil {
			return nil, fmt.Errorf("entity not found: %w", err)
		}
	} else {
		// Use shared entityService instance
		newEntity, err := s.entityService.CreateEntity(&models.CreateEntityRequest{
			Name:        req.Name,
			EntityType:  req.EntityType,
			Email:       req.Email,
			PhoneNumber: req.PhoneNumber,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create entity: %w", err)
		}
		entity = newEntity.ToEntity()
	}
	// Create provider
	provider := models.Provider{
		ProviderID: "prov_" + uuid.New().String(),
		EntityID:   entity.EntityID,
	}

	if err := s.db.Create(&provider).Error; err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	response := &models.ProviderResponse{
		ProviderID:  provider.ProviderID,
		EntityID:    provider.EntityID,
		Name:        entity.Name,
		EntityType:  entity.EntityType,
		Email:       entity.Email,
		PhoneNumber: entity.PhoneNumber,
		CreatedAt:   provider.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   provider.UpdatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// GetProvider retrieves a provider by ID with entity information
func (s *ProviderService) GetProvider(providerID string) (*models.ProviderResponse, error) {
	var provider models.Provider

	err := s.db.Preload("Entity").First(&provider, "provider_id = ?", providerID).Error
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	return &models.ProviderResponse{
		ProviderID:  provider.ProviderID,
		EntityID:    provider.EntityID,
		Name:        provider.Entity.Name,
		EntityType:  provider.Entity.EntityType,
		Email:       provider.Entity.Email,
		PhoneNumber: provider.Entity.PhoneNumber,
		CreatedAt:   provider.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   provider.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// UpdateProvider updates an existing provider and its associated entity
func (s *ProviderService) UpdateProvider(providerID string, req *models.UpdateProviderRequest) (*models.ProviderResponse, error) {
	var provider models.Provider
	err := s.db.Preload("Entity").First(&provider, "provider_id = ?", providerID).Error
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	// Update associated entity fields if provided
	if req.Name != nil {
		provider.Entity.Name = *req.Name
	}
	if req.EntityType != nil {
		provider.Entity.EntityType = *req.EntityType
	}
	if req.Email != nil {
		provider.Entity.Email = *req.Email
	}
	if req.PhoneNumber != nil {
		provider.Entity.PhoneNumber = *req.PhoneNumber
	}

	if err := s.db.Save(&provider.Entity).Error; err != nil {
		return nil, fmt.Errorf("failed to update entity: %w", err)
	}

	// Save updated provider (if there were any provider-specific fields to update)
	if err := s.db.Save(&provider).Error; err != nil {
		return nil, fmt.Errorf("failed to update provider: %w", err)
	}

	response := &models.ProviderResponse{
		ProviderID:  provider.ProviderID,
		EntityID:    provider.EntityID,
		Name:        provider.Entity.Name,
		EntityType:  provider.Entity.EntityType,
		Email:       provider.Entity.Email,
		PhoneNumber: provider.Entity.PhoneNumber,
		CreatedAt:   provider.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   provider.UpdatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// CreateProviderSchema creates a new approved schema for a provider
func (s *ProviderService) CreateProviderSchema(providerID string, req *models.CreateProviderSchemaRequest) (*models.ProviderSchemaResponse, error) {
	// Verify provider exists
	var provider models.Provider
	err := s.db.First(&provider, "provider_id = ?", providerID).Error
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	// Create schema
	schema := models.ProviderSchema{
		SchemaID:          "schema_" + uuid.New().String(),
		SchemaName:        req.SchemaName,
		SDL:               req.SDL,
		Endpoint:          req.Endpoint,
		Version:           models.ActiveVersion,
		SchemaDescription: req.SchemaDescription,
		ProviderID:        providerID,
	}

	err = s.db.Create(&schema).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return &models.ProviderSchemaResponse{
		SchemaID:          schema.SchemaID,
		ProviderID:        schema.ProviderID,
		SchemaName:        schema.SchemaName,
		SDL:               schema.SDL,
		Endpoint:          schema.Endpoint,
		Version:           schema.Version,
		SchemaDescription: schema.SchemaDescription,
		CreatedAt:         schema.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         schema.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// UpdateProviderSchema updates an existing approved schema for a provider
func (s *ProviderService) UpdateProviderSchema(providerID, schemaID string, req *models.UpdateProviderSchemaRequest) (*models.ProviderSchemaResponse, error) {
	var schema models.ProviderSchema
	err := s.db.First(&schema, "schema_id = ?", schemaID).Error
	if err != nil {
		return nil, fmt.Errorf("schema not found: %w", err)
	}

	// Update fields if provided
	if req.SchemaName != nil {
		schema.SchemaName = *req.SchemaName
	}
	if req.SDL != nil {
		schema.SDL = *req.SDL
	}
	if req.Endpoint != nil {
		schema.Endpoint = *req.Endpoint
	}
	if req.Version != nil {
		schema.Version = *req.Version
	}
	if req.SchemaDescription != nil {
		schema.SchemaDescription = req.SchemaDescription
	}

	if err := s.db.Save(&schema).Error; err != nil {
		return nil, fmt.Errorf("failed to update schema: %w", err)
	}

	return &models.ProviderSchemaResponse{
		SchemaID:          schema.SchemaID,
		ProviderID:        schema.ProviderID,
		SchemaName:        schema.SchemaName,
		SDL:               schema.SDL,
		Endpoint:          schema.Endpoint,
		Version:           schema.Version,
		SchemaDescription: schema.SchemaDescription,
		CreatedAt:         schema.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         schema.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// GetProviderSchemas retrieves approved schemas for a provider
func (s *ProviderService) GetProviderSchemas(providerID string) ([]models.ProviderSchemaResponse, error) {
	var schemas []models.ProviderSchema

	err := s.db.Preload("Provider").Preload("Provider.Entity").Where("provider_id = ?", providerID).Find(&schemas).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schemas: %w", err)
	}

	response := make([]models.ProviderSchemaResponse, len(schemas))
	for i, schema := range schemas {
		response[i] = models.ProviderSchemaResponse{
			SchemaID:          schema.SchemaID,
			ProviderID:        schema.ProviderID,
			SchemaName:        schema.SchemaName,
			SDL:               schema.SDL,
			Endpoint:          schema.Endpoint,
			Version:           schema.Version,
			SchemaDescription: schema.SchemaDescription,
			CreatedAt:         schema.CreatedAt.Format(time.RFC3339),
			UpdatedAt:         schema.UpdatedAt.Format(time.RFC3339),
		}
	}

	return response, nil
}

// CreateProviderSchemaSubmission creates a new schema submission
func (s *ProviderService) CreateProviderSchemaSubmission(providerID string, req models.CreateProviderSchemaSubmissionRequest) (*models.ProviderSchemaSubmissionResponse, error) {
	// Verify provider exists
	var provider models.Provider
	err := s.db.First(&provider, "provider_id = ?", providerID).Error
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	// Create submission
	submission := models.ProviderSchemaSubmission{
		SubmissionID:      "sub_" + uuid.New().String(),
		PreviousSchemaID:  req.PreviousSchemaID,
		SchemaName:        req.SchemaName,
		SchemaDescription: req.SchemaDescription,
		SDL:               req.SDL,
		SchemaEndpoint:    req.SchemaEndpoint,
		Status:            models.StatusPending,
		ProviderID:        providerID,
	}

	err = s.db.Create(&submission).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create schema submission: %w", err)
	}

	return &models.ProviderSchemaSubmissionResponse{
		SubmissionID:      submission.SubmissionID,
		PreviousSchemaID:  submission.PreviousSchemaID,
		SchemaName:        submission.SchemaName,
		SchemaDescription: submission.SchemaDescription,
		SDL:               submission.SDL,
		SchemaEndpoint:    submission.SchemaEndpoint,
		Status:            submission.Status,
		ProviderID:        submission.ProviderID,
		CreatedAt:         submission.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         submission.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// UpdateProviderSchemaSubmission updates an existing schema submission
func (s *ProviderService) UpdateProviderSchemaSubmission(providerID, submissionID string, req *models.UpdateProviderSchemaSubmissionRequest) (*models.ProviderSchemaSubmissionResponse, error) {
	var submission models.ProviderSchemaSubmission
	err := s.db.First(&submission, "submission_id = ?", submissionID).Error
	if err != nil {
		return nil, fmt.Errorf("schema submission not found: %w", err)
	}

	// Update fields if provided
	if req.PreviousSchemaID != nil {
		submission.PreviousSchemaID = req.PreviousSchemaID
	}
	if req.SchemaName != nil {
		submission.SchemaName = *req.SchemaName
	}
	if req.SchemaDescription != nil {
		submission.SchemaDescription = req.SchemaDescription
	}
	if req.SDL != nil {
		submission.SDL = *req.SDL
	}
	if req.SchemaEndpoint != nil {
		submission.SchemaEndpoint = *req.SchemaEndpoint
	}
	if req.Status != nil {
		submission.Status = *req.Status
	}

	if err := s.db.Save(&submission).Error; err != nil {
		return nil, fmt.Errorf("failed to update schema submission: %w", err)
	}

	return &models.ProviderSchemaSubmissionResponse{
		SubmissionID:      submission.SubmissionID,
		PreviousSchemaID:  submission.PreviousSchemaID,
		SchemaName:        submission.SchemaName,
		SchemaDescription: submission.SchemaDescription,
		SDL:               submission.SDL,
		SchemaEndpoint:    submission.SchemaEndpoint,
		Status:            submission.Status,
		ProviderID:        submission.ProviderID,
		CreatedAt:         submission.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         submission.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// GetProviderSchemaSubmissions retrieves schema submissions for a provider
func (s *ProviderService) GetProviderSchemaSubmissions(providerID string, status string) ([]models.ProviderSchemaSubmissionResponse, error) {
	var submissions []models.ProviderSchemaSubmission

	query := s.db.Preload("Provider").Preload("Provider.Entity").Preload("PreviousSchema").Where("provider_id = ?", providerID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Find(&submissions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schema submissions: %w", err)
	}

	response := make([]models.ProviderSchemaSubmissionResponse, len(submissions))
	for i, submission := range submissions {
		response[i] = models.ProviderSchemaSubmissionResponse{
			SubmissionID:      submission.SubmissionID,
			PreviousSchemaID:  submission.PreviousSchemaID,
			SchemaName:        submission.SchemaName,
			SchemaDescription: submission.SchemaDescription,
			SDL:               submission.SDL,
			SchemaEndpoint:    submission.SchemaEndpoint,
			Status:            submission.Status,
			ProviderID:        submission.ProviderID,
			CreatedAt:         submission.CreatedAt.Format(time.RFC3339),
			UpdatedAt:         submission.UpdatedAt.Format(time.RFC3339),
		}
	}

	return response, nil
}

// GetAllProviders retrieves all providers with entity information
func (s *ProviderService) GetAllProviders() ([]models.ProviderResponse, error) {
	var providers []models.Provider

	err := s.db.Preload("Entity").Find(&providers).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch providers: %w", err)
	}

	response := make([]models.ProviderResponse, len(providers))
	for i, provider := range providers {
		response[i] = models.ProviderResponse{
			ProviderID:  provider.ProviderID,
			EntityID:    provider.EntityID,
			Name:        provider.Entity.Name,
			EntityType:  provider.Entity.EntityType,
			Email:       provider.Entity.Email,
			PhoneNumber: provider.Entity.PhoneNumber,
			CreatedAt:   provider.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   provider.UpdatedAt.Format(time.RFC3339),
		}
	}

	return response, nil
}
