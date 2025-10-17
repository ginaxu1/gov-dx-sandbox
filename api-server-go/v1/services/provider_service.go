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
