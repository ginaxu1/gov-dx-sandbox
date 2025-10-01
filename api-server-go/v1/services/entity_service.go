package services

import (
	"fmt"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"gorm.io/gorm"
)

// EntityService handles entity-related operations
type EntityService struct {
	db *gorm.DB
}

// NewEntityService creates a new entity service
func NewEntityService(db *gorm.DB) *EntityService {
	return &EntityService{db: db}
}

// GetEntity retrieves an entity by ID with associated provider/consumer information
func (s *EntityService) GetEntity(entityID string) (*models.EntityResponse, error) {
	var entity models.Entity

	err := s.db.First(&entity, "entity_id = ?", entityID).Error
	if err != nil {
		return nil, fmt.Errorf("entity not found: %w", err)
	}

	response := &models.EntityResponse{
		EntityID:    entity.EntityID,
		Name:        entity.Name,
		EntityType:  entity.EntityType,
		Email:       entity.Email,
		PhoneNumber: entity.PhoneNumber,
		CreatedAt:   entity.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   entity.UpdatedAt.Format(time.RFC3339),
	}

	// Check if this entity is a provider
	var provider models.Provider
	if err := s.db.First(&provider, "entity_id = ?", entityID).Error; err == nil {
		response.ProviderID = &provider.ProviderID
	}

	// Check if this entity is a consumer
	var consumer models.Consumer
	if err := s.db.First(&consumer, "entity_id = ?", entityID).Error; err == nil {
		response.ConsumerID = &consumer.ConsumerID
	}

	return response, nil
}

// GetAllEntities retrieves all entities with their associated provider/consumer information
func (s *EntityService) GetAllEntities() ([]models.EntityResponse, error) {
	var entities []models.Entity

	err := s.db.Find(&entities).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch entities: %w", err)
	}

	response := make([]models.EntityResponse, len(entities))
	for i, entity := range entities {
		entityResponse := models.EntityResponse{
			EntityID:    entity.EntityID,
			Name:        entity.Name,
			EntityType:  entity.EntityType,
			Email:       entity.Email,
			PhoneNumber: entity.PhoneNumber,
			CreatedAt:   entity.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   entity.UpdatedAt.Format(time.RFC3339),
		}

		// Check if this entity is a provider
		var provider models.Provider
		if err := s.db.First(&provider, "entity_id = ?", entity.EntityID).Error; err == nil {
			entityResponse.ProviderID = &provider.ProviderID
		}

		// Check if this entity is a consumer
		var consumer models.Consumer
		if err := s.db.First(&consumer, "entity_id = ?", entity.EntityID).Error; err == nil {
			entityResponse.ConsumerID = &consumer.ConsumerID
		}

		response[i] = entityResponse
	}

	return response, nil
}
