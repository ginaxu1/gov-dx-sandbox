package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sync"

	"github.com/gov-dx-sandbox/api-server-go/models"
)

type ConsumerService struct {
	applications map[string]*models.Application
	mutex        sync.RWMutex
}

func NewConsumerService() *ConsumerService {
	return &ConsumerService{
		applications: make(map[string]*models.Application),
	}
}

// GetAllApplications retrieves all consumer applications
func (s *ConsumerService) GetAllApplications() ([]*models.Application, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	applications := make([]*models.Application, 0, len(s.applications))
	for _, app := range s.applications {
		applications = append(applications, app)
	}

	return applications, nil
}

// CreateApplication creates a new consumer application
func (s *ConsumerService) CreateApplication(req models.CreateApplicationRequest) (*models.Application, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Generate unique app ID
	appID, err := s.generateAppID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate app ID: %w", err)
	}

	application := &models.Application{
		AppID:          appID,
		Status:         models.StatusPending,
		RequiredFields: req.RequiredFields,
	}

	s.applications[appID] = application

	slog.Info("Created new application", "appId", appID)
	return application, nil
}

// GetApplication retrieves a specific consumer application
func (s *ConsumerService) GetApplication(id string) (*models.Application, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	application, exists := s.applications[id]
	if !exists {
		return nil, fmt.Errorf("application not found")
	}

	return application, nil
}

// UpdateApplication updates a consumer application
func (s *ConsumerService) UpdateApplication(id string, req models.UpdateApplicationRequest) (*models.UpdateApplicationResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	application, exists := s.applications[id]
	if !exists {
		return nil, fmt.Errorf("application not found")
	}

	response := &models.UpdateApplicationResponse{
		Application: application,
	}

	// Update fields if provided
	if req.Status != nil {
		application.Status = *req.Status
	}
	if req.RequiredFields != nil {
		application.RequiredFields = req.RequiredFields
	}

	// Generate credentials if status is approved and credentials don't exist
	if application.Status == models.StatusApproved && application.Credentials == nil {
		credentials, err := s.generateCredentials()
		if err != nil {
			return nil, fmt.Errorf("failed to generate credentials: %w", err)
		}
		application.Credentials = credentials

		// For now, assign a default provider ID when approved
		// In a real system, this would be based on business logic
		response.ProviderID = "default_provider"
	}

	slog.Info("Updated application", "appId", id, "status", application.Status)
	return response, nil
}

// DeleteApplication deletes a consumer application
func (s *ConsumerService) DeleteApplication(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exists := s.applications[id]
	if !exists {
		return fmt.Errorf("application not found")
	}

	delete(s.applications, id)
	slog.Info("Deleted application", "appId", id)
	return nil
}

// generateAppID generates a unique application ID
func (s *ConsumerService) generateAppID() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "app_" + hex.EncodeToString(bytes), nil
}

// generateCredentials generates API credentials
func (s *ConsumerService) generateCredentials() (*models.Credentials, error) {
	apiKeyBytes := make([]byte, 16)
	if _, err := rand.Read(apiKeyBytes); err != nil {
		return nil, err
	}

	apiSecretBytes := make([]byte, 32)
	if _, err := rand.Read(apiSecretBytes); err != nil {
		return nil, err
	}

	return &models.Credentials{
		APIKey:    hex.EncodeToString(apiKeyBytes),
		APISecret: hex.EncodeToString(apiSecretBytes),
	}, nil
}
