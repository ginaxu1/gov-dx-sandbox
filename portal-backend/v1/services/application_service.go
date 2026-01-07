package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/portal-backend/idp"
	"github.com/gov-dx-sandbox/portal-backend/v1/models"
	"gorm.io/gorm"
)

// ApplicationService handles application-related operations
type ApplicationService struct {
	db            *gorm.DB
	policyService *PDPService
	idp           idp.IdentityProviderAPI
}

// NewApplicationService creates a new application service
func NewApplicationService(db *gorm.DB, pdpService *PDPService, idp idp.IdentityProviderAPI) *ApplicationService {
	return &ApplicationService{db: db, policyService: pdpService, idp: idp}
}

// CreateApplication creates a new application
func (s *ApplicationService) CreateApplication(ctx context.Context, req *models.CreateApplicationRequest) (*models.ApplicationResponse, error) {
	// Step 1: Create Application in the IDP
	description := ""
	if req.ApplicationDescription != nil {
		description = *req.ApplicationDescription
	}

	applicationInstance := &idp.Application{
		Name:        req.ApplicationName,
		Description: description,
		TemplateId:  models.TemplateIDM2M,
	}
	idpApplicationID, err := s.idp.CreateApplication(ctx, applicationInstance)
	if err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}
	appOIDCInfo, err := s.idp.GetApplicationOIDC(ctx, *idpApplicationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get application OIDC: %w", err)
	}

	// Step 2: Create application in database
	application := models.Application{
		ApplicationID:          uuid.New().String(),
		ApplicationName:        req.ApplicationName,
		ApplicationDescription: req.ApplicationDescription,
		SelectedFields:         models.SelectedFieldRecords(req.SelectedFields),
		IdpApplicationID:       idpApplicationID,
		IdpClientID:            &appOIDCInfo.ClientId,
		MemberID:               req.MemberID,
		Version:                string(models.ActiveVersion),
	}

	if err := s.db.WithContext(ctx).Create(&application).Error; err != nil {
		// Compensation: Delete the application we just created
		if deleteErr := s.idp.DeleteApplication(ctx, *idpApplicationID); deleteErr != nil {
			// Log the compensation failure - this needs monitoring
			slog.Error("Failed to compensate application creation",
				"applicationID", application.ApplicationID,
				"originalError", err,
				"compensationError", deleteErr)
			// Return both errors for visibility
			return nil, fmt.Errorf("failed to create application: %w, and failed to compensate: %w", err, deleteErr)
		}
		slog.Info("Successfully compensated application creation", "applicationID", application.ApplicationID)
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	// Step 3: Update allow list in PDP (Saga Pattern)
	policyReq := models.AllowListUpdateRequest{
		ApplicationID: application.ApplicationID,
		Records:       application.SelectedFields,
		GrantDuration: models.GrantDurationTypeOneMonth, // Default duration
	}

	_, err = s.policyService.UpdateAllowList(policyReq)
	if err != nil {
		// Compensation: Delete the application we just created
		if deleteErr := s.db.Delete(&application).Error; deleteErr != nil {
			// Log the compensation failure - this needs monitoring
			slog.Error("Failed to compensate application creation",
				"applicationID", application.ApplicationID,
				"originalError", err,
				"compensationError", deleteErr)
			// Return both errors for visibility
			return nil, fmt.Errorf("failed to update allow list: %w, and failed to compensate: %w", err, deleteErr)
		}
		// Compensation: Delete the application in IDP as well
		if idpDeleteErr := s.idp.DeleteApplication(ctx, *application.IdpApplicationID); idpDeleteErr != nil {
			slog.Error("Failed to compensate application creation in IDP",
				"applicationID", application.ApplicationID,
				"originalError", err,
				"compensationError", idpDeleteErr)
			return nil, fmt.Errorf("failed to update allow list: %w, and failed to compensate in IDP: %w", err, idpDeleteErr)
		}
		slog.Info("Successfully compensated application creation", "applicationID", application.ApplicationID)
		return nil, fmt.Errorf("failed to update allow list: %w", err)
	}

	response := &models.ApplicationResponse{
		ApplicationID:          application.ApplicationID,
		ApplicationName:        application.ApplicationName,
		ApplicationDescription: application.ApplicationDescription,
		SelectedFields:         application.SelectedFields,
		MemberID:               application.MemberID,
		Version:                application.Version,
		IdpApplicationID:       application.IdpApplicationID,
		IdpClientID:            application.IdpClientID,
		CreatedAt:              application.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              application.UpdatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// UpdateApplication updates an existing application
func (s *ApplicationService) UpdateApplication(ctx context.Context, applicationID string, req *models.UpdateApplicationRequest) (*models.ApplicationResponse, error) {
	var application models.Application
	err := s.db.First(&application, "application_id = ?", applicationID).Error
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	// Note: SelectedFields updates are intentionally not supported for approved applications
	// to maintain data integrity and audit trail. Field changes require resubmission process.
	if req.ApplicationName != nil {
		application.ApplicationName = *req.ApplicationName
	}
	if req.ApplicationDescription != nil {
		application.ApplicationDescription = req.ApplicationDescription
	}
	if req.Version != nil {
		application.Version = *req.Version
	}

	if err := s.db.WithContext(ctx).Save(&application).Error; err != nil {
		return nil, err
	}

	response := &models.ApplicationResponse{
		ApplicationID:          application.ApplicationID,
		ApplicationName:        application.ApplicationName,
		ApplicationDescription: application.ApplicationDescription,
		SelectedFields:         application.SelectedFields,
		MemberID:               application.MemberID,
		Version:                application.Version,
		IdpApplicationID:       application.IdpApplicationID,
		IdpClientID:            application.IdpClientID,
		CreatedAt:              application.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              application.UpdatedAt.Format(time.RFC3339),
	}
	if application.ApplicationDescription != nil && *application.ApplicationDescription != "" {
		response.ApplicationDescription = application.ApplicationDescription
	}

	return response, nil
}

// GetApplication retrieves an application by ID
func (s *ApplicationService) GetApplication(ctx context.Context, applicationID string) (*models.ApplicationResponse, error) {
	var application models.Application
	err := s.db.WithContext(ctx).Preload("Member").First(&application, "application_id = ?", applicationID).Error
	if err != nil {
		return nil, err
	}

	response := &models.ApplicationResponse{
		ApplicationID:          application.ApplicationID,
		ApplicationName:        application.ApplicationName,
		ApplicationDescription: application.ApplicationDescription,
		SelectedFields:         application.SelectedFields,
		MemberID:               application.MemberID,
		Version:                application.Version,
		IdpApplicationID:       application.IdpApplicationID,
		IdpClientID:            application.IdpClientID,
		CreatedAt:              application.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              application.UpdatedAt.Format(time.RFC3339),
	}
	if application.ApplicationDescription != nil && *application.ApplicationDescription != "" {
		response.ApplicationDescription = application.ApplicationDescription
	}

	return response, nil
}

// GetApplicationIdByIdpClientId retrives applicationId by idpClientId
func (s *ApplicationService) GetApplicationIdByIdpClientId(ctx context.Context, idpClientId string) (*models.ApplicationIDResponse, error) {
	var application models.Application
	err := s.db.WithContext(ctx).First(&application, "idp_client_id = ?", idpClientId).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("application not found for idpClientId: %s", idpClientId)
		}
		return nil, fmt.Errorf("failed to retrieve application: %w", err)
	}
	return &models.ApplicationIDResponse{
		ApplicationID: application.ApplicationID,
	}, nil
}

// GetApplications retrieves all applications and filters by member ID if provided
func (s *ApplicationService) GetApplications(ctx context.Context, MemberID *string) ([]models.ApplicationResponse, error) {
	var applications []models.Application
	query := s.db.WithContext(ctx).Preload("Member")
	if MemberID != nil && *MemberID != "" {
		query = query.Where("member_id = ?", *MemberID)
	}

	// Order by created_at descending
	query = query.Order("created_at DESC")

	err := query.Find(&applications).Error
	if err != nil {
		return nil, err
	}

	// Pre-allocate slice with known capacity for better performance
	responses := make([]models.ApplicationResponse, 0, len(applications))
	for _, application := range applications {
		resp := models.ApplicationResponse{
			ApplicationID:    application.ApplicationID,
			ApplicationName:  application.ApplicationName,
			SelectedFields:   application.SelectedFields,
			MemberID:         application.MemberID,
			IdpApplicationID: application.IdpApplicationID,
			IdpClientID:      application.IdpClientID,
			Version:          application.Version,
			CreatedAt:        application.CreatedAt.Format(time.RFC3339),
			UpdatedAt:        application.UpdatedAt.Format(time.RFC3339),
		}
		if application.ApplicationDescription != nil && *application.ApplicationDescription != "" {
			resp.ApplicationDescription = application.ApplicationDescription
		}
		responses = append(responses, resp)
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

	// Validate member ID
	var member models.Member
	err := s.db.First(&member, "member_id = ?", req.MemberID).Error
	if err != nil {
		return nil, err
	}

	// Create application submission
	submission := models.ApplicationSubmission{
		SubmissionID:           "sub_" + uuid.New().String(),
		PreviousApplicationID:  req.PreviousApplicationID,
		ApplicationName:        req.ApplicationName,
		ApplicationDescription: req.ApplicationDescription,
		SelectedFields:         models.SelectedFieldRecords(req.SelectedFields),
		Status:                 string(models.StatusPending),
		MemberID:               req.MemberID,
	}
	if err := s.db.Create(&submission).Error; err != nil {
		return nil, err
	}

	response := &models.ApplicationSubmissionResponse{
		SubmissionID:           submission.SubmissionID,
		PreviousApplicationID:  submission.PreviousApplicationID,
		ApplicationName:        submission.ApplicationName,
		ApplicationDescription: submission.ApplicationDescription,
		SelectedFields:         submission.SelectedFields,
		Status:                 submission.Status,
		MemberID:               submission.MemberID,
		CreatedAt:              submission.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              submission.UpdatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// UpdateApplicationSubmission updates an existing application submission
func (s *ApplicationService) UpdateApplicationSubmission(ctx context.Context, submissionID string, req *models.UpdateApplicationSubmissionRequest) (*models.ApplicationSubmissionResponse, error) {
	var submission models.ApplicationSubmission

	// Find the submission
	if err := s.db.First(&submission, "submission_id = ?", submissionID).Error; err != nil {
		return nil, fmt.Errorf("application submission not found: %w", err)
	}

	// Validate PreviousApplicationID first before making any updates
	if req.PreviousApplicationID != nil {
		// Validate previous application ID
		var prevApp models.Application
		if err := s.db.First(&prevApp, "application_id = ?", *req.PreviousApplicationID).Error; err != nil {
			return nil, fmt.Errorf("previous application not found: %w", err)
		}
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

	if req.PreviousApplicationID != nil {
		submission.PreviousApplicationID = req.PreviousApplicationID
	}

	var shouldCreateApplication bool
	if req.Status != nil {
		submission.Status = *req.Status
		// Mark that we need to create an application after saving
		if *req.Status == string(models.StatusApproved) {
			shouldCreateApplication = true
		}
	}

	if req.Review != nil {
		submission.Review = req.Review
	}

	// Save the updated submission
	if err := s.db.Save(&submission).Error; err != nil {
		return nil, fmt.Errorf("failed to update application submission: %w", err)
	}

	// Create application outside of transaction if approval was successful
	if shouldCreateApplication {
		var createApplicationRequest models.CreateApplicationRequest
		createApplicationRequest.ApplicationName = submission.ApplicationName
		createApplicationRequest.ApplicationDescription = submission.ApplicationDescription
		createApplicationRequest.SelectedFields = models.SelectedFieldRecords(submission.SelectedFields)
		createApplicationRequest.MemberID = submission.MemberID

		_, err := s.CreateApplication(ctx, &createApplicationRequest)
		if err != nil {
			// Compensation: Update submission status back to pending
			submission.Status = string(models.StatusPending)
			if updateErr := s.db.Save(&submission).Error; updateErr != nil {
				slog.Error("Failed to compensate submission status after application creation failure",
					"submissionID", submission.SubmissionID,
					"originalError", err,
					"compensationError", updateErr)
				return nil, fmt.Errorf("failed to create application from approved submission: %w, and failed to compensate submission status: %w", err, updateErr)
			}
			slog.Info("Successfully compensated submission status after application creation failure", "submissionID", submission.SubmissionID)
			return nil, fmt.Errorf("failed to create application from approved submission: %w", err)
		}
	}

	response := &models.ApplicationSubmissionResponse{
		SubmissionID:           submission.SubmissionID,
		PreviousApplicationID:  submission.PreviousApplicationID,
		ApplicationName:        submission.ApplicationName,
		ApplicationDescription: submission.ApplicationDescription,
		SelectedFields:         submission.SelectedFields,
		Status:                 submission.Status,
		MemberID:               submission.MemberID,
		CreatedAt:              submission.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              submission.UpdatedAt.Format(time.RFC3339),
		Review:                 submission.Review,
	}

	return response, nil
}

// GetApplicationSubmission retrieves an application submission by ID
func (s *ApplicationService) GetApplicationSubmission(submissionID string) (*models.ApplicationSubmissionResponse, error) {
	var submission models.ApplicationSubmission
	err := s.db.Preload("Member").Preload("PreviousApplication").First(&submission, "submission_id = ?", submissionID).Error
	if err != nil {
		return nil, err
	}

	response := &models.ApplicationSubmissionResponse{
		SubmissionID:           submission.SubmissionID,
		PreviousApplicationID:  submission.PreviousApplicationID,
		ApplicationName:        submission.ApplicationName,
		ApplicationDescription: submission.ApplicationDescription,
		SelectedFields:         submission.SelectedFields,
		Status:                 submission.Status,
		MemberID:               submission.MemberID,
		CreatedAt:              submission.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              submission.UpdatedAt.Format(time.RFC3339),
		Review:                 submission.Review,
	}

	return response, nil
}

// GetApplicationSubmissions retrieves all application submissions and filters by member ID if provided
func (s *ApplicationService) GetApplicationSubmissions(MemberID *string, statusFilter *[]string) ([]models.ApplicationSubmissionResponse, error) {
	var submissions []models.ApplicationSubmission
	query := s.db.Preload("Member").Preload("PreviousApplication")
	if MemberID != nil && *MemberID != "" {
		query = query.Where("member_id = ?", *MemberID)
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
			SelectedFields:         submission.SelectedFields,
			Status:                 submission.Status,
			MemberID:               submission.MemberID,
			CreatedAt:              submission.CreatedAt.Format(time.RFC3339),
			UpdatedAt:              submission.UpdatedAt.Format(time.RFC3339),
			Review:                 submission.Review,
		})
	}

	return responses, nil
}
