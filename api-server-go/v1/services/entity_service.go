package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/api-server-go/idp"
	"github.com/gov-dx-sandbox/api-server-go/idp/idpfactory"
	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"gorm.io/gorm"
)

// EntityService handles entity-related operations
type EntityService struct {
	db  *gorm.DB
	idp *idp.IdentityProviderAPI
}

// NewEntityService creates a new entity service
func NewEntityService(db *gorm.DB) *EntityService {
	// Get scopes from environment variable, fallback to default if not set
	scopesEnv := os.Getenv("ASGARDEO_SCOPES")
	var scopes []string
	if scopesEnv != "" {
		// Split by space to handle multiple scopes
		scopes = strings.Fields(scopesEnv)
	}

	// Create the NewIdpProvider
	idpProvider, err := idpfactory.NewIdpAPIProvider(idpfactory.FactoryConfig{
		ProviderType: idp.ProviderAsgardeo,
		BaseURL:      os.Getenv("ASGARDEO_BASE_URL"),
		ClientID:     os.Getenv("ASGARDEO_CLIENT_ID"),
		ClientSecret: os.Getenv("ASGARDEO_CLIENT_SECRET"),
		Scopes:       scopes,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create IDP provider: %v", err))
	}
	return &EntityService{db: db, idp: &idpProvider}
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
	updatedName := entity.Name
	updatedPhoneNumber := entity.PhoneNumber

	// Update fields if provided
	if req.Name != nil {
		entity.Name = *req.Name
		updatedName = *req.Name
		needsIdpUpdate = true
	}
	if req.PhoneNumber != nil {
		entity.PhoneNumber = *req.PhoneNumber
		updatedPhoneNumber = *req.PhoneNumber
		needsIdpUpdate = true
	}

	// Update user in IDP if necessary
	if needsIdpUpdate {
		ctx := context.Background()
		userInstance := &idp.User{
			Email:       entity.Email,
			FirstName:   updatedName,
			LastName:    "",
			PhoneNumber: updatedPhoneNumber,
		}

		_, err := (*s.idp).UpdateUser(ctx, entity.IdpUserID, userInstance)
		if err != nil {
			return nil, fmt.Errorf("failed to update user in IDP: %w", err)
		}

		slog.Info("Updated user in IDP", "userID", entity.IdpUserID)
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
