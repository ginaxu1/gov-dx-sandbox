package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/api-server-go/idp"
	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"gorm.io/gorm"
)

// EntityService handles entity-related operations
type EntityService struct {
	db  *gorm.DB
	idp *idp.IdentityProviderAPI
}

// NewEntityService creates a new entity service
func NewEntityService(db *gorm.DB, idp *idp.IdentityProviderAPI) *EntityService {
	return &EntityService{db: db, idp: idp}
}

// CreateEntity creates a new entity
func (s *EntityService) CreateEntity(req *models.CreateEntityRequest) (*models.EntityResponse, error) {
	// Create user in the IDP using the factory-created client
	ctx := context.Background()

	userInstance := &idp.User{
		Email:       req.Email,
		FirstName:   req.Name,
		LastName:    "",
		PhoneNumber: req.PhoneNumber,
	}

	createdUser, err := (*s.idp).CreateUser(ctx, userInstance)
	if err != nil {
		return nil, fmt.Errorf("failed to create user in IDP: %w", err)
	}

	if createdUser.Email != userInstance.Email {
		return nil, fmt.Errorf("IDP user email mismatch")
	}

	slog.Info("Created user in IDP", "userID", createdUser.Id)

	// Create entity in the database
	entity := models.Entity{
		EntityID:    "ent_" + uuid.New().String(),
		Name:        req.Name,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		IdpUserID:   createdUser.Id,
	}

	if err := s.db.Create(&entity).Error; err != nil {
		// Rollback IDP user creation if DB operation fails (not implemented here)
		err := (*s.idp).DeleteUser(ctx, createdUser.Id)
		if err != nil {
			return nil, err
		}
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

	// Check if we need to update the IDP user
	needsIdpUpdate := false
	beforeUpdateName := entity.Name
	beforeUpdatePhoneNumber := entity.PhoneNumber

	// Update fields if provided
	if req.Name != nil {
		entity.Name = *req.Name
		needsIdpUpdate = true
	}
	if req.PhoneNumber != nil {
		entity.PhoneNumber = *req.PhoneNumber
		needsIdpUpdate = true
	}

	// Update user in IDP if necessary
	if needsIdpUpdate {
		ctx := context.Background()
		userInstance := &idp.User{
			Email:       entity.Email,
			FirstName:   entity.Name,
			LastName:    "",
			PhoneNumber: entity.PhoneNumber,
		}

		_, err := (*s.idp).UpdateUser(ctx, entity.IdpUserID, userInstance)
		if err != nil {
			return nil, fmt.Errorf("failed to update user in IDP: %w", err)
		}

		slog.Info("Updated user in IDP", "userID", entity.IdpUserID)
	}

	if err := s.db.Save(&entity).Error; err != nil {
		// Rollback IDP user update if DB operation fails (not implemented here)
		_, err := (*s.idp).UpdateUser(context.Background(), entity.IdpUserID, &idp.User{
			Email:       entity.Email,
			FirstName:   beforeUpdateName,
			LastName:    "",
			PhoneNumber: beforeUpdatePhoneNumber,
		})
		if err != nil {
			return nil, err
		}
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
