package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/models"
)

type ProviderService struct {
	submissions map[string]*models.ProviderSubmission
	profiles    map[string]*models.ProviderProfile
	schemas     map[string]*models.ProviderSchema
	mutex       sync.RWMutex
}

func NewProviderService() *ProviderService {
	return &ProviderService{
		submissions: make(map[string]*models.ProviderSubmission),
		profiles:    make(map[string]*models.ProviderProfile),
		schemas:     make(map[string]*models.ProviderSchema),
	}
}

// Provider Submission methods
func (s *ProviderService) GetAllProviderSubmissions() ([]*models.ProviderSubmission, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	submissions := make([]*models.ProviderSubmission, 0, len(s.submissions))
	for _, sub := range s.submissions {
		submissions = append(submissions, sub)
	}

	return submissions, nil
}

func (s *ProviderService) CreateProviderSubmission(req models.CreateProviderSubmissionRequest) (*models.ProviderSubmission, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Validate required fields
	if req.ProviderName == "" {
		return nil, fmt.Errorf("providerName is required")
	}
	if req.ContactEmail == "" {
		return nil, fmt.Errorf("contactEmail is required")
	}
	if req.PhoneNumber == "" {
		return nil, fmt.Errorf("phoneNumber is required")
	}
	if req.ProviderType == "" {
		return nil, fmt.Errorf("providerType is required")
	}

	// Check for duplicate pending submission
	for _, sub := range s.submissions {
		if sub.ProviderName == req.ProviderName && sub.Status == models.SubmissionStatusPending {
			return nil, fmt.Errorf("a pending submission for '%s' already exists", req.ProviderName)
		}
	}

	// Generate unique submission ID
	submissionID, err := s.generateSubmissionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate submission ID: %w", err)
	}

	submission := &models.ProviderSubmission{
		SubmissionID: submissionID,
		ProviderName: req.ProviderName,
		ContactEmail: req.ContactEmail,
		PhoneNumber:  req.PhoneNumber,
		ProviderType: req.ProviderType,
		Status:       models.SubmissionStatusPending,
		CreatedAt:    time.Now(),
	}

	s.submissions[submissionID] = submission

	slog.Info("Created new provider submission", "submissionId", submissionID)
	return submission, nil
}

func (s *ProviderService) GetProviderSubmission(id string) (*models.ProviderSubmission, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	submission, exists := s.submissions[id]
	if !exists {
		return nil, fmt.Errorf("provider submission not found")
	}

	return submission, nil
}

func (s *ProviderService) UpdateProviderSubmission(id string, req models.UpdateProviderSubmissionRequest) (*models.UpdateProviderSubmissionResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	submission, exists := s.submissions[id]
	if !exists {
		return nil, fmt.Errorf("provider submission not found")
	}

	response := &models.UpdateProviderSubmissionResponse{
		ProviderSubmission: submission,
	}

	// Update status if provided
	if req.Status != nil {
		submission.Status = *req.Status

		// If approved, create provider profile
		if *req.Status == models.SubmissionStatusApproved {
			profile, err := s.createProviderProfile(submission)
			if err != nil {
				return nil, fmt.Errorf("failed to create provider profile: %w", err)
			}
			s.profiles[profile.ProviderID] = profile
			response.ProviderID = profile.ProviderID
		}
	}

	slog.Info("Updated provider submission", "submissionId", id, "status", submission.Status)
	return response, nil
}

// Provider Profile methods
func (s *ProviderService) GetAllProviderProfiles() ([]*models.ProviderProfile, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	profiles := make([]*models.ProviderProfile, 0, len(s.profiles))
	for _, profile := range s.profiles {
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

func (s *ProviderService) GetProviderProfile(id string) (*models.ProviderProfile, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	profile, exists := s.profiles[id]
	if !exists {
		return nil, fmt.Errorf("provider profile not found")
	}

	return profile, nil
}

// Provider Schema methods
func (s *ProviderService) GetAllProviderSchemas() ([]*models.ProviderSchema, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	schemas := make([]*models.ProviderSchema, 0, len(s.schemas))
	for _, schema := range s.schemas {
		schemas = append(schemas, schema)
	}

	return schemas, nil
}

func (s *ProviderService) CreateProviderSchema(req models.CreateProviderSchemaRequest) (*models.ProviderSchema, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Verify provider exists
	_, exists := s.profiles[req.ProviderID]
	if !exists {
		return nil, fmt.Errorf("provider profile not found")
	}

	// Generate unique schema submission ID
	schemaID, err := s.generateSchemaID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate schema ID: %w", err)
	}

	schema := &models.ProviderSchema{
		SubmissionID:        schemaID,
		ProviderID:          req.ProviderID,
		Status:              models.SchemaStatusPending,
		SchemaInput:         req.SchemaInput,
		FieldConfigurations: req.FieldConfigurations,
	}

	s.schemas[schemaID] = schema

	slog.Info("Created new provider schema", "submissionId", schemaID)
	return schema, nil
}

func (s *ProviderService) GetProviderSchema(id string) (*models.ProviderSchema, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	schema, exists := s.schemas[id]
	if !exists {
		return nil, fmt.Errorf("provider schema not found")
	}

	return schema, nil
}

func (s *ProviderService) UpdateProviderSchema(id string, req models.UpdateProviderSchemaRequest) (*models.ProviderSchema, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	schema, exists := s.schemas[id]
	if !exists {
		return nil, fmt.Errorf("provider schema not found")
	}

	// Update fields if provided
	if req.Status != nil {
		schema.Status = *req.Status
	}
	if req.FieldConfigurations != nil {
		schema.FieldConfigurations = req.FieldConfigurations
	}

	slog.Info("Updated provider schema", "submissionId", id, "status", schema.Status)
	return schema, nil
}

// Helper methods
func (s *ProviderService) generateSubmissionID() (string, error) {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "sub_prov_" + hex.EncodeToString(bytes), nil
}

func (s *ProviderService) generateSchemaID() (string, error) {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "schema_" + hex.EncodeToString(bytes), nil
}

func (s *ProviderService) generateProviderID() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "prov_" + hex.EncodeToString(bytes), nil
}

func (s *ProviderService) createProviderProfile(submission *models.ProviderSubmission) (*models.ProviderProfile, error) {
	providerID, err := s.generateProviderID()
	if err != nil {
		return nil, err
	}

	profile := &models.ProviderProfile{
		ProviderID:   providerID,
		ProviderName: submission.ProviderName,
		ContactEmail: submission.ContactEmail,
		PhoneNumber:  submission.PhoneNumber,
		ProviderType: submission.ProviderType,
		ApprovedAt:   time.Now(),
	}

	return profile, nil
}
