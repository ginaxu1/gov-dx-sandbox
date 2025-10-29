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
	var provider models.Provider
	var entity models.Entity

	// Use transaction to ensure atomicity between entity and provider creation
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if req.EntityID != nil {
			// Verify entity exists
			if err := tx.First(&entity, "entity_id = ?", req.EntityID).Error; err != nil {
				return fmt.Errorf("entity not found: %w", err)
			}
		} else {
			// Create entity within same transaction
			entity = models.Entity{
				EntityID:    "ent_" + uuid.New().String(),
				Name:        req.Name,
				Email:       req.Email,
				PhoneNumber: req.PhoneNumber,
				IdpUserID:   req.IdpUserID,
			}
			if err := tx.Create(&entity).Error; err != nil {
				return fmt.Errorf("failed to create entity: %w", err)
			}
		}

		// Create provider
		provider = models.Provider{
			ProviderID: "prov_" + uuid.New().String(),
			EntityID:   entity.EntityID,
		}

		if err := tx.Create(&provider).Error; err != nil {
			return fmt.Errorf("failed to create provider: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	response := &models.ProviderResponse{
		ProviderID:  provider.ProviderID,
		EntityID:    provider.EntityID,
		IdpUserID:   entity.IdpUserID,
		Name:        entity.Name,
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
		Email:       provider.Entity.Email,
		PhoneNumber: provider.Entity.PhoneNumber,
		CreatedAt:   provider.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   provider.UpdatedAt.Format(time.RFC3339),
		IdpUserID:   provider.Entity.IdpUserID,
	}, nil
}

// UpdateProvider updates an existing provider and its associated entity
func (s *ProviderService) UpdateProvider(providerID string, req *models.UpdateProviderRequest) (*models.ProviderResponse, error) {
	var provider models.Provider

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Fetch provider with entity in single query within transaction
		if err := tx.Preload("Entity").First(&provider, "provider_id = ?", providerID).Error; err != nil {
			return fmt.Errorf("provider not found: %w", err)
		}

		// Update associated entity fields if provided
		if req.Name != nil {
			provider.Entity.Name = *req.Name
		}
		if req.IdpUserID != nil {
			provider.Entity.IdpUserID = *req.IdpUserID
		}
		if req.Email != nil {
			provider.Entity.Email = *req.Email
		}
		if req.PhoneNumber != nil {
			provider.Entity.PhoneNumber = *req.PhoneNumber
		}

		// Batch save both entity and provider in single transaction
		if err := tx.Save(&provider.Entity).Error; err != nil {
			return fmt.Errorf("failed to update entity: %w", err)
		}

		if err := tx.Save(&provider).Error; err != nil {
			return fmt.Errorf("failed to update provider: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	response := &models.ProviderResponse{
		ProviderID:  provider.ProviderID,
		EntityID:    provider.EntityID,
		Name:        provider.Entity.Name,
		Email:       provider.Entity.Email,
		PhoneNumber: provider.Entity.PhoneNumber,
		CreatedAt:   provider.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   provider.UpdatedAt.Format(time.RFC3339),
		IdpUserID:   provider.Entity.IdpUserID,
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
			IdpUserID:   provider.Entity.IdpUserID,
			Name:        provider.Entity.Name,
			Email:       provider.Entity.Email,
			PhoneNumber: provider.Entity.PhoneNumber,
			CreatedAt:   provider.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   provider.UpdatedAt.Format(time.RFC3339),
		}
	}

	return response, nil
}
