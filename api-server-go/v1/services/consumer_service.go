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
	db *gorm.DB
}

// NewConsumerService creates a new consumer service
func NewConsumerService(db *gorm.DB) *ConsumerService {
	return &ConsumerService{db: db}
}

// GetConsumer retrieves a consumer by ID with entity information
func (s *ConsumerService) GetConsumer(consumerID string) (*models.ConsumerResponse, error) {
	var consumer models.Consumer
	var entity models.Entity

	err := s.db.Preload("Entity").First(&consumer, "consumer_id = ?", consumerID).Error
	if err != nil {
		return nil, fmt.Errorf("consumer not found: %w", err)
	}

	// Get entity information
	err = s.db.First(&entity, "entity_id = ?", consumer.EntityID).Error
	if err != nil {
		return nil, fmt.Errorf("entity not found: %w", err)
	}

	return &models.ConsumerResponse{
		ConsumerID:  consumer.ConsumerID,
		EntityID:    consumer.EntityID,
		Name:        entity.Name,
		EntityType:  entity.EntityType,
		Email:       entity.Email,
		PhoneNumber: entity.PhoneNumber,
		CreatedAt:   consumer.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   consumer.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// GetConsumerApplications retrieves applications for a consumer
func (s *ConsumerService) GetConsumerApplications(consumerID string) ([]models.ConsumerApplicationResponse, error) {
	var applications []models.ConsumerApplication

	err := s.db.Where("consumer_id = ?", consumerID).Find(&applications).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch applications: %w", err)
	}

	response := make([]models.ConsumerApplicationResponse, len(applications))
	for i, app := range applications {
		response[i] = models.ConsumerApplicationResponse{
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
func (s *ConsumerService) CreateConsumerApplicationSubmission(consumerID string, req models.CreateConsumerApplicationSubmissionRequest) (*models.ConsumerApplicationSubmissionResponse, error) {
	// Verify consumer exists
	var consumer models.Consumer
	err := s.db.First(&consumer, "consumer_id = ?", consumerID).Error
	if err != nil {
		return nil, fmt.Errorf("consumer not found: %w", err)
	}

	// Create submission
	submission := models.ConsumerApplicationSubmission{
		SubmissionID:           "sub_" + uuid.New().String(),
		PreviousApplicationID:  req.PreviousApplicationID,
		ApplicationName:        req.ApplicationName,
		ApplicationDescription: req.ApplicationDescription,
		SelectedFields:         models.StringArray(req.SelectedFields),
		ConsumerID:             consumerID,
		Status:                 models.StatusPending,
	}

	err = s.db.Create(&submission).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create application submission: %w", err)
	}

	return &models.ConsumerApplicationSubmissionResponse{
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
func (s *ConsumerService) GetConsumerApplicationSubmissions(consumerID string, status string) ([]models.ConsumerApplicationSubmissionResponse, error) {
	var submissions []models.ConsumerApplicationSubmission

	query := s.db.Where("consumer_id = ?", consumerID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Find(&submissions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch application submissions: %w", err)
	}

	response := make([]models.ConsumerApplicationSubmissionResponse, len(submissions))
	for i, submission := range submissions {
		response[i] = models.ConsumerApplicationSubmissionResponse{
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
