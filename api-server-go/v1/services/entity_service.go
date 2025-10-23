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
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		IdpUserID:   req.IdpUserID,
	}

	if err := s.db.Create(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}

	response := &models.EntityResponse{
		EntityID:    entity.EntityID,
		IdpUserID:   entity.IdpUserID,
		Name:        entity.Name,
		Email:       entity.Email,
		PhoneNumber: entity.PhoneNumber,
		CreatedAt:   entity.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   entity.UpdatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// UpdateEntity updates an existing entity
func (s *EntityService) UpdateEntity(entityID string, req *models.UpdateEntityRequest) (*models.EntityResponse, error) {
	var entity models.Entity
	err := s.db.First(&entity, "entity_id = ?", entityID).Error
	if err != nil {
		return nil, fmt.Errorf("entity not found: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		entity.Name = *req.Name
	}
	if req.IdpUserID != nil {
		entity.IdpUserID = *req.IdpUserID
	}
	if req.Email != nil {
		entity.Email = *req.Email
	}
	if req.PhoneNumber != nil {
		entity.PhoneNumber = *req.PhoneNumber
	}

	if err := s.db.Save(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to update entity: %w", err)
	}

	response := &models.EntityResponse{
		EntityID:    entity.EntityID,
		IdpUserID:   entity.IdpUserID,
		Name:        entity.Name,
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
	}
	err := s.db.Table("entities").
		Where("entities.entity_id = ?", entityID).
		First(&result).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch entity: %w", err)
	}

	response := &models.EntityResponse{
		EntityID:    result.Entity.EntityID,
		IdpUserID:   result.Entity.IdpUserID,
		Name:        result.Entity.Name,
		Email:       result.Entity.Email,
		PhoneNumber: result.Entity.PhoneNumber,
		CreatedAt:   result.Entity.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   result.Entity.UpdatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// GetAllEntities retrieves all entities with their associated provider/consumer information
func (s *EntityService) GetAllEntities() ([]models.EntityResponse, error) {
	var results []struct {
		models.Entity
	}

	err := s.db.Table("entities").Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch entities: %w", err)
	}

	response := make([]models.EntityResponse, len(results))
	for i, result := range results {
		response[i] = models.EntityResponse{
			EntityID:    result.Entity.EntityID,
			IdpUserID:   result.Entity.IdpUserID,
			Name:        result.Entity.Name,
			Email:       result.Entity.Email,
			PhoneNumber: result.Entity.PhoneNumber,
			CreatedAt:   result.Entity.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   result.Entity.UpdatedAt.Format(time.RFC3339),
		}
	}

	return response, nil
}
