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

// CreateConsumerApplication creates a new consumer application
func (s *ConsumerService) CreateConsumerApplication(req models.CreateApplicationRequest) (*models.ApplicationResponse, error) {
	// Verify consumer exists
	var consumer models.Consumer
	err := s.db.First(&consumer, "consumer_id = ?", req.ConsumerID).Error
	if err != nil {
		return nil, fmt.Errorf("consumer not found: %w", err)
	}

	// Create application
	application := models.Application{
		ApplicationID:          "app_" + uuid.New().String(),
		ApplicationName:        req.ApplicationName,
		ApplicationDescription: req.ApplicationDescription,
		SelectedFields:         models.StringArray(req.SelectedFields),
		ConsumerID:             req.ConsumerID,
		Version:                models.ActiveVersion,
	}

	err = s.db.Create(&application).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	return &models.ApplicationResponse{
		ApplicationID:          application.ApplicationID,
		ApplicationName:        application.ApplicationName,
		ApplicationDescription: application.ApplicationDescription,
		SelectedFields:         application.SelectedFields,
		ConsumerID:             application.ConsumerID,
		Version:                application.Version,
		CreatedAt:              application.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              application.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// UpdateConsumerApplication updates an existing consumer application
func (s *ConsumerService) UpdateConsumerApplication(applicationID string, req models.UpdateApplicationRequest) (*models.ApplicationResponse, error) {
	var application models.Application
	err := s.db.First(&application, "application_id = ?", applicationID).Error
	if err != nil {
		return nil, fmt.Errorf("application not found: %w", err)
	}

	// Update fields if provided
	if req.ApplicationName != nil {
		application.ApplicationName = *req.ApplicationName
	}
	if req.ApplicationDescription != nil {
		application.ApplicationDescription = req.ApplicationDescription
	}
	if req.SelectedFields != nil {
		application.SelectedFields = *req.SelectedFields
	}
	if req.Version != nil {
		application.Version = *req.Version
	}

	if err := s.db.Save(&application).Error; err != nil {
		return nil, fmt.Errorf("failed to update application: %w", err)
	}

	return &models.ApplicationResponse{
		ApplicationID:          application.ApplicationID,
		ApplicationName:        application.ApplicationName,
		ApplicationDescription: application.ApplicationDescription,
		SelectedFields:         []string(application.SelectedFields),
		ConsumerID:             application.ConsumerID,
		Version:                application.Version,
		CreatedAt:              application.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              application.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// GetConsumerApplications retrieves applications for a consumer
func (s *ConsumerService) GetConsumerApplications(consumerID string) ([]models.ApplicationResponse, error) {
	var applications []models.Application

	err := s.db.Preload("Consumer").Preload("Consumer.Entity").Where("consumer_id = ?", consumerID).Find(&applications).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch applications: %w", err)
	}

	response := make([]models.ApplicationResponse, len(applications))
	for i, app := range applications {
		response[i] = models.ApplicationResponse{
			ApplicationID:          app.ApplicationID,
			ApplicationName:        app.ApplicationName,
			ApplicationDescription: app.ApplicationDescription,
			SelectedFields:         []string(app.SelectedFields),
			ConsumerID:             app.ConsumerID,
			Version:                app.Version,
			CreatedAt:              app.CreatedAt.Format(time.RFC3339),
			UpdatedAt:              app.UpdatedAt.Format(time.RFC3339),
		}
	}

	return response, nil
}

// CreateConsumerApplicationSubmission creates a new application submission
func (s *ConsumerService) CreateConsumerApplicationSubmission(req models.CreateApplicationSubmissionRequest) (*models.ApplicationSubmissionResponse, error) {
	// Verify consumer exists
	var consumer models.Consumer
	err := s.db.First(&consumer, "consumer_id = ?", req.ConsumerID).Error
	if err != nil {
		return nil, fmt.Errorf("consumer not found: %w", err)
	}

	// Create submission
	submission := models.ApplicationSubmission{
		SubmissionID:           "sub_" + uuid.New().String(),
		PreviousApplicationID:  req.PreviousApplicationID,
		ApplicationName:        req.ApplicationName,
		ApplicationDescription: req.ApplicationDescription,
		SelectedFields:         models.StringArray(req.SelectedFields),
		ConsumerID:             req.ConsumerID,
		Status:                 models.StatusPending,
	}

	err = s.db.Create(&submission).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create application submission: %w", err)
	}

	return &models.ApplicationSubmissionResponse{
		SubmissionID:           submission.SubmissionID,
		PreviousApplicationID:  submission.PreviousApplicationID,
		ApplicationName:        submission.ApplicationName,
		ApplicationDescription: submission.ApplicationDescription,
		SelectedFields:         []string(submission.SelectedFields),
		ConsumerID:             submission.ConsumerID,
		Status:                 submission.Status,
		CreatedAt:              submission.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              submission.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// UpdateConsumerApplicationSubmission updates an existing application submission
func (s *ConsumerService) UpdateConsumerApplicationSubmission(submissionID string, req models.UpdateApplicationSubmissionRequest) (*models.ApplicationSubmissionResponse, error) {
	var submission models.ApplicationSubmission
	err := s.db.First(&submission, "submission_id = ?", submissionID).Error
	if err != nil {
		return nil, fmt.Errorf("application submission not found: %w", err)
	}

	// Update fields if provided
	if req.ApplicationName != nil {
		submission.ApplicationName = *req.ApplicationName
	}
	if req.ApplicationDescription != nil {
		submission.ApplicationDescription = req.ApplicationDescription
	}
	if req.SelectedFields != nil {
		submission.SelectedFields = *req.SelectedFields
	}
	if req.Status != nil {
		submission.Status = *req.Status
	}

	if err := s.db.Save(&submission).Error; err != nil {
		return nil, fmt.Errorf("failed to update application submission: %w", err)
	}

	return &models.ApplicationSubmissionResponse{
		SubmissionID:           submission.SubmissionID,
		PreviousApplicationID:  submission.PreviousApplicationID,
		ApplicationName:        submission.ApplicationName,
		ApplicationDescription: submission.ApplicationDescription,
		SelectedFields:         []string(submission.SelectedFields),
		ConsumerID:             submission.ConsumerID,
		Status:                 submission.Status,
		CreatedAt:              submission.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              submission.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// GetConsumerApplicationSubmissions retrieves application submissions for a consumer
func (s *ConsumerService) GetConsumerApplicationSubmissions(consumerID string, status string) ([]models.ApplicationSubmissionResponse, error) {
	var submissions []models.ApplicationSubmission

	query := s.db.Preload("Consumer").Preload("Consumer.Entity").Preload("PreviousApplication").Where("consumer_id = ?", consumerID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Find(&submissions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch application submissions: %w", err)
	}

	response := make([]models.ApplicationSubmissionResponse, len(submissions))
	for i, submission := range submissions {
		response[i] = models.ApplicationSubmissionResponse{
			SubmissionID:           submission.SubmissionID,
			PreviousApplicationID:  submission.PreviousApplicationID,
			ApplicationName:        submission.ApplicationName,
			ApplicationDescription: submission.ApplicationDescription,
			SelectedFields:         []string(submission.SelectedFields),
			ConsumerID:             submission.ConsumerID,
			Status:                 submission.Status,
			CreatedAt:              submission.CreatedAt.Format(time.RFC3339),
			UpdatedAt:              submission.UpdatedAt.Format(time.RFC3339),
		}
	}

	return response, nil
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
