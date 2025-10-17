package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"gorm.io/gorm"
)

// ApplicationService handles application-related operations
type ApplicationService struct {
	db *gorm.DB
}

// NewApplicationService creates a new application service
func NewApplicationService(db *gorm.DB) *ApplicationService {
	return &ApplicationService{db: db}
}

// CreateApplication creates a new application
func (s *ApplicationService) CreateApplication(req *models.CreateApplicationRequest) (*models.ApplicationResponse, error) {
	// Create application
	application := models.Application{
		ApplicationID:          "app_" + uuid.New().String(),
		ApplicationName:        req.ApplicationName,
		ApplicationDescription: req.ApplicationDescription,
		SelectedFields:         models.StringArray(req.SelectedFields),
		ConsumerID:             req.ConsumerID,
		Version:                models.ActiveVersion,
	}
	if err := s.db.Create(&application).Error; err != nil {
		return nil, err
	}

	response := &models.ApplicationResponse{
		ApplicationID:          application.ApplicationID,
		ApplicationName:        application.ApplicationName,
		ApplicationDescription: application.ApplicationDescription,
		SelectedFields:         []string(application.SelectedFields),
		ConsumerID:             application.ConsumerID,
		Version:                application.Version,
		CreatedAt:              application.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              application.UpdatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// UpdateApplication updates an existing application
func (s *ApplicationService) UpdateApplication(applicationID string, req *models.UpdateApplicationRequest) (*models.ApplicationResponse, error) {
	var application models.Application
	err := s.db.First(&application, "application_id = ?", applicationID).Error
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.ApplicationName != nil {
		application.ApplicationName = *req.ApplicationName
	}
	if req.ApplicationDescription != nil {
		application.ApplicationDescription = req.ApplicationDescription
	}
	if req.SelectedFields != nil && len(*req.SelectedFields) > 0 {
		application.SelectedFields = *req.SelectedFields
	}
	if req.Version != nil {
		application.Version = *req.Version
	}

	if err := s.db.Save(&application).Error; err != nil {
		return nil, err
	}

	response := &models.ApplicationResponse{
		ApplicationID:          application.ApplicationID,
		ApplicationName:        application.ApplicationName,
		ApplicationDescription: application.ApplicationDescription,
		SelectedFields:         []string(application.SelectedFields),
		ConsumerID:             application.ConsumerID,
		Version:                application.Version,
		CreatedAt:              application.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              application.UpdatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// GetApplication retrieves an application by ID
func (s *ApplicationService) GetApplication(applicationID string) (*models.ApplicationResponse, error) {
	var application models.Application
	err := s.db.Preload("Consumer").First(&application, "application_id = ?", applicationID).Error
	if err != nil {
		return nil, err
	}

	response := &models.ApplicationResponse{
		ApplicationID:          application.ApplicationID,
		ApplicationName:        application.ApplicationName,
		ApplicationDescription: application.ApplicationDescription,
		SelectedFields:         []string(application.SelectedFields),
		ConsumerID:             application.ConsumerID,
		Version:                application.Version,
		CreatedAt:              application.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              application.UpdatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// GetApplications retrieves all applications and filters by consumer ID if provided
func (s *ApplicationService) GetApplications(consumerID *string) ([]models.ApplicationResponse, error) {
	var applications []models.Application
	query := s.db.Preload("Consumer")
	if consumerID != nil && *consumerID != "" {
		query = query.Where("consumer_id = ?", *consumerID)
	}

	// Order by created_at descending
	query = query.Order("created_at DESC")

	err := query.Find(&applications).Error
	if err != nil {
		return nil, err
	}

	var responses []models.ApplicationResponse
	for _, application := range applications {
		responses = append(responses, models.ApplicationResponse{
			ApplicationID:          application.ApplicationID,
			ApplicationName:        application.ApplicationName,
			ApplicationDescription: application.ApplicationDescription,
			SelectedFields:         []string(application.SelectedFields),
			ConsumerID:             application.ConsumerID,
			Version:                application.Version,
			CreatedAt:              application.CreatedAt.Format(time.RFC3339),
			UpdatedAt:              application.UpdatedAt.Format(time.RFC3339),
		})
	}

	return responses, nil
}

// CreateApplicationSubmission creates a new application submission
func (s *ApplicationService) CreateApplicationSubmission(req *models.CreateApplicationSubmissionRequest) (*models.ApplicationSubmissionResponse, error) {
	// Validate previous application ID if provided
	if req.PreviousApplicationID != nil {
		var prevApp models.Application
		err := s.db.First(&prevApp, "application_id = ?", *req.PreviousApplicationID).Error
		if err != nil {
			return nil, err
		}
	}

	// Validate consumer ID
	var consumer models.Consumer
	err := s.db.First(&consumer, "consumer_id = ?", req.ConsumerID).Error
	if err != nil {
		return nil, err
	}

	// Create application submission
	submission := models.ApplicationSubmission{
		SubmissionID:           "sub_" + uuid.New().String(),
		PreviousApplicationID:  req.PreviousApplicationID,
		ApplicationName:        req.ApplicationName,
		ApplicationDescription: req.ApplicationDescription,
		SelectedFields:         models.StringArray(req.SelectedFields),
		Status:                 models.StatusPending,
		ConsumerID:             req.ConsumerID,
	}
	if err := s.db.Create(&submission).Error; err != nil {
		return nil, err
	}

	response := &models.ApplicationSubmissionResponse{
		SubmissionID:           submission.SubmissionID,
		PreviousApplicationID:  submission.PreviousApplicationID,
		ApplicationName:        submission.ApplicationName,
		ApplicationDescription: submission.ApplicationDescription,
		SelectedFields:         []string(submission.SelectedFields),
		Status:                 submission.Status,
		ConsumerID:             submission.ConsumerID,
		CreatedAt:              submission.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              submission.UpdatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// UpdateApplicationSubmission updates an existing application submission
func (s *ApplicationService) UpdateApplicationSubmission(submissionID string, req *models.UpdateApplicationSubmissionRequest) (*models.ApplicationSubmissionResponse, error) {
	var submission models.ApplicationSubmission

	// Start a transaction
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&submission, "submission_id = ?", submissionID).Error; err != nil {
			return fmt.Errorf("application submission not found: %w", err)
		}

		// Update fields if provided
		if req.ApplicationName != nil {
			submission.ApplicationName = *req.ApplicationName
		}
		if req.ApplicationDescription != nil {
			submission.ApplicationDescription = req.ApplicationDescription
		}
		if req.SelectedFields != nil && len(*req.SelectedFields) > 0 {
			submission.SelectedFields = *req.SelectedFields
		}
		if req.Status != nil {
			submission.Status = *req.Status
			// If status is approved, create a new application
			if *req.Status == models.StatusApproved {
				application := models.Application{
					ApplicationID:          "app_" + uuid.New().String(),
					ApplicationName:        submission.ApplicationName,
					ApplicationDescription: submission.ApplicationDescription,
					SelectedFields:         submission.SelectedFields,
					ConsumerID:             submission.ConsumerID,
					Version:                models.ActiveVersion,
				}
				if err := tx.Create(&application).Error; err != nil {
					return fmt.Errorf("failed to create application: %w", err)
				}
			}
		}
		if req.PreviousApplicationID != nil {
			// Validate previous application ID
			var prevApp models.Application
			if err := tx.First(&prevApp, "application_id = ?", *req.PreviousApplicationID).Error; err != nil {
				return fmt.Errorf("previous application not found: %w", err)
			}
			submission.PreviousApplicationID = req.PreviousApplicationID
		}
		if err := tx.Save(&submission).Error; err != nil {
			return fmt.Errorf("failed to update application submission: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	response := &models.ApplicationSubmissionResponse{
		SubmissionID:           submission.SubmissionID,
		PreviousApplicationID:  submission.PreviousApplicationID,
		ApplicationName:        submission.ApplicationName,
		ApplicationDescription: submission.ApplicationDescription,
		SelectedFields:         []string(submission.SelectedFields),
		Status:                 submission.Status,
		ConsumerID:             submission.ConsumerID,
		CreatedAt:              submission.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              submission.UpdatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// GetApplicationSubmission retrieves an application submission by ID
func (s *ApplicationService) GetApplicationSubmission(submissionID string) (*models.ApplicationSubmissionResponse, error) {
	var submission models.ApplicationSubmission
	err := s.db.Preload("Consumer").Preload("PreviousApplication").First(&submission, "submission_id = ?", submissionID).Error
	if err != nil {
		return nil, err
	}

	response := &models.ApplicationSubmissionResponse{
		SubmissionID:           submission.SubmissionID,
		PreviousApplicationID:  submission.PreviousApplicationID,
		ApplicationName:        submission.ApplicationName,
		ApplicationDescription: submission.ApplicationDescription,
		SelectedFields:         []string(submission.SelectedFields),
		Status:                 submission.Status,
		ConsumerID:             submission.ConsumerID,
		CreatedAt:              submission.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              submission.UpdatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// GetApplicationSubmissions retrieves all application submissions and filters by consumer ID if provided
func (s *ApplicationService) GetApplicationSubmissions(consumerID *string, statusFilter *[]string) ([]models.ApplicationSubmissionResponse, error) {
	var submissions []models.ApplicationSubmission
	query := s.db.Preload("Consumer").Preload("PreviousApplication")
	if consumerID != nil && *consumerID != "" {
		query = query.Where("consumer_id = ?", *consumerID)
	}
	if statusFilter != nil && len(*statusFilter) > 0 {
		query = query.Where("status IN ?", *statusFilter)
	}

	// Order by created_at descending
	query = query.Order("created_at DESC")

	err := query.Find(&submissions).Error
	if err != nil {
		return nil, err
	}

	var responses []models.ApplicationSubmissionResponse
	for _, submission := range submissions {
		responses = append(responses, models.ApplicationSubmissionResponse{
			SubmissionID:           submission.SubmissionID,
			PreviousApplicationID:  submission.PreviousApplicationID,
			ApplicationName:        submission.ApplicationName,
			ApplicationDescription: submission.ApplicationDescription,
			SelectedFields:         []string(submission.SelectedFields),
			Status:                 submission.Status,
			ConsumerID:             submission.ConsumerID,
			CreatedAt:              submission.CreatedAt.Format(time.RFC3339),
			UpdatedAt:              submission.UpdatedAt.Format(time.RFC3339),
		})
	}

	return responses, nil
}
