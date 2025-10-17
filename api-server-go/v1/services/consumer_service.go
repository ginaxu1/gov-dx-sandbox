package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"gorm.io/gorm"
)

// ConsumerService handles consumer-related operations
type ConsumerService struct {
	db            *gorm.DB
	entityService *EntityService
}

// NewConsumerService creates a new consumer service
func NewConsumerService(db *gorm.DB, entityService *EntityService) *ConsumerService {
	return &ConsumerService{db: db, entityService: entityService}
}

// CreateConsumer creates a new consumer
func (s *ConsumerService) CreateConsumer(req *models.CreateConsumerRequest) (*models.ConsumerResponse, error) {
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

	// Create consumer
	consumer := models.Consumer{
		ConsumerID: "cons_" + uuid.New().String(),
		EntityID:   entity.EntityID,
	}

	if err := s.db.Create(&consumer).Error; err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	response := &models.ConsumerResponse{
		ConsumerID:  consumer.ConsumerID,
		EntityID:    consumer.EntityID,
		Name:        entity.Name,
		EntityType:  entity.EntityType,
		Email:       entity.Email,
		PhoneNumber: entity.PhoneNumber,
		CreatedAt:   consumer.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   consumer.UpdatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// UpdateConsumer updates an existing consumer and its associated entity
func (s *ConsumerService) UpdateConsumer(consumerID string, req *models.UpdateConsumerRequest) (*models.ConsumerResponse, error) {
	var consumer models.Consumer
	err := s.db.Preload("Entity").First(&consumer, "consumer_id = ?", consumerID).Error
	if err != nil {
		return nil, fmt.Errorf("consumer not found: %w", err)
	}

	// Update associated entity fields if provided
	if req.Name != nil {
		consumer.Entity.Name = *req.Name
	}
	if req.EntityType != nil {
		consumer.Entity.EntityType = *req.EntityType
	}
	if req.Email != nil {
		consumer.Entity.Email = *req.Email
	}
	if req.PhoneNumber != nil {
		consumer.Entity.PhoneNumber = *req.PhoneNumber
	}

	// Save updated entity
	if err := s.db.Save(&consumer.Entity).Error; err != nil {
		return nil, fmt.Errorf("failed to update entity: %w", err)
	}

	// Save updated consumer (if there were any consumer-specific fields to update)
	if err := s.db.Save(&consumer).Error; err != nil {
		return nil, fmt.Errorf("failed to update consumer: %w", err)
	}

	response := &models.ConsumerResponse{
		ConsumerID:  consumer.ConsumerID,
		EntityID:    consumer.EntityID,
		Name:        consumer.Entity.Name,
		EntityType:  consumer.Entity.EntityType,
		Email:       consumer.Entity.Email,
		PhoneNumber: consumer.Entity.PhoneNumber,
		CreatedAt:   consumer.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   consumer.UpdatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// GetConsumer retrieves a consumer by ID with entity information
func (s *ConsumerService) GetConsumer(consumerID string) (*models.ConsumerResponse, error) {
	var consumer models.Consumer

	err := s.db.Preload("Entity").First(&consumer, "consumer_id = ?", consumerID).Error
	if err != nil {
		return nil, fmt.Errorf("consumer not found: %w", err)
	}

	return &models.ConsumerResponse{
		ConsumerID:  consumer.ConsumerID,
		EntityID:    consumer.EntityID,
		Name:        consumer.Entity.Name,
		EntityType:  consumer.Entity.EntityType,
		Email:       consumer.Entity.Email,
		PhoneNumber: consumer.Entity.PhoneNumber,
		CreatedAt:   consumer.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   consumer.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// GetAllConsumers retrieves all consumers with entity information
func (s *ConsumerService) GetAllConsumers() ([]models.ConsumerResponse, error) {
	var consumers []models.Consumer

	err := s.db.Preload("Entity").Find(&consumers).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch consumers: %w", err)
	}

	response := make([]models.ConsumerResponse, len(consumers))
	for i, consumer := range consumers {
		response[i] = models.ConsumerResponse{
			ConsumerID:  consumer.ConsumerID,
			EntityID:    consumer.EntityID,
			Name:        consumer.Entity.Name,
			EntityType:  consumer.Entity.EntityType,
			Email:       consumer.Entity.Email,
			PhoneNumber: consumer.Entity.PhoneNumber,
			CreatedAt:   consumer.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   consumer.UpdatedAt.Format(time.RFC3339),
		}
	}

	return response, nil
}
