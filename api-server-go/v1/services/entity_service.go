package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
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

// CreateEntity creates a new entity
func (s *EntityService) CreateEntity(req *models.CreateEntityRequest) (*models.EntityResponse, error) {
	entity := models.Entity{
		EntityID:    "ent_" + uuid.New().String(),
		Name:        req.Name,
		EntityType:  req.EntityType,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
	}

	if err := s.db.Create(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
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

	return response, nil
}

// GetEntity retrieves an entity by ID with associated provider/consumer information
func (s *EntityService) GetEntity(entityID string) (*models.EntityResponse, error) {
	var result struct {
		models.Entity
		ProviderID *string `gorm:"column:provider_id"`
		ConsumerID *string `gorm:"column:consumer_id"`
	}

	err := s.db.Table("entities").
		Select("entities.*, providers.provider_id, consumers.consumer_id").
		Joins("INNER JOIN providers ON entities.entity_id = providers.entity_id").
		Joins("INNER JOIN consumers ON entities.entity_id = consumers.entity_id").
		Where("entities.entity_id = ?", entityID).
		First(&result).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch entity: %w", err)
	}

	response := &models.EntityResponse{
		EntityID:    result.Entity.EntityID,
		Name:        result.Entity.Name,
		EntityType:  result.Entity.EntityType,
		Email:       result.Entity.Email,
		PhoneNumber: result.Entity.PhoneNumber,
		CreatedAt:   result.Entity.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   result.Entity.UpdatedAt.Format(time.RFC3339),
		ProviderID:  result.ProviderID,
		ConsumerID:  result.ConsumerID,
	}

	return response, nil
}

// GetAllEntities retrieves all entities with their associated provider/consumer information
func (s *EntityService) GetAllEntities() ([]models.EntityResponse, error) {
	var results []struct {
		models.Entity
		ProviderID *string `gorm:"column:provider_id"`
		ConsumerID *string `gorm:"column:consumer_id"`
	}

	err := s.db.Table("entities").
		Select("entities.*, providers.provider_id, consumers.consumer_id").
		Joins("INNER JOIN providers ON entities.entity_id = providers.entity_id").
		Joins("INNER JOIN consumers ON entities.entity_id = consumers.entity_id").
		Find(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch entities: %w", err)
	}

	response := make([]models.EntityResponse, len(results))
	for i, result := range results {
		response[i] = models.EntityResponse{
			EntityID:    result.Entity.EntityID,
			Name:        result.Entity.Name,
			EntityType:  result.Entity.EntityType,
			Email:       result.Entity.Email,
			PhoneNumber: result.Entity.PhoneNumber,
			CreatedAt:   result.Entity.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   result.Entity.UpdatedAt.Format(time.RFC3339),
			ProviderID:  result.ProviderID,
			ConsumerID:  result.ConsumerID,
		}
	}

	return response, nil
}
