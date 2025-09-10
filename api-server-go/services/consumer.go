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

type ConsumerService struct {
	consumers    map[string]*models.Consumer
	applications map[string]*models.ConsumerApp
	mutex        sync.RWMutex
}

func NewConsumerService() *ConsumerService {
	return &ConsumerService{
		consumers:    make(map[string]*models.Consumer),
		applications: make(map[string]*models.ConsumerApp),
	}
}

// Consumer management methods

// CreateConsumer creates a new consumer
func (s *ConsumerService) CreateConsumer(req models.CreateConsumerRequest) (*models.Consumer, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Generate unique consumer ID
	consumerID, err := s.generateConsumerID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate consumer ID: %w", err)
	}

	consumer := &models.Consumer{
		ConsumerID:   consumerID,
		ConsumerName: req.ConsumerName,
		ContactEmail: req.ContactEmail,
		PhoneNumber:  req.PhoneNumber,
		CreatedAt:    time.Now(),
	}

	s.consumers[consumerID] = consumer

	slog.Info("Created new consumer", "consumerId", consumerID)
	return consumer, nil
}

// GetConsumer retrieves a specific consumer
func (s *ConsumerService) GetConsumer(id string) (*models.Consumer, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	consumer, exists := s.consumers[id]
	if !exists {
		return nil, fmt.Errorf("consumer not found")
	}

	return consumer, nil
}

// GetAllConsumers retrieves all consumers
func (s *ConsumerService) GetAllConsumers() ([]*models.Consumer, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	consumers := make([]*models.Consumer, 0, len(s.consumers))
	for _, consumer := range s.consumers {
		consumers = append(consumers, consumer)
	}

	return consumers, nil
}

// ConsumerApp management methods

// CreateConsumerApp creates a new consumer application
func (s *ConsumerService) CreateConsumerApp(req models.CreateConsumerAppRequest) (*models.ConsumerApp, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Verify consumer exists
	_, exists := s.consumers[req.ConsumerID]
	if !exists {
		return nil, fmt.Errorf("consumer not found")
	}

	// Generate unique submission ID
	submissionID, err := s.generateSubmissionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate submission ID: %w", err)
	}

	application := &models.ConsumerApp{
		SubmissionID:   submissionID,
		ConsumerID:     req.ConsumerID,
		Status:         models.StatusPending,
		RequiredFields: req.RequiredFields,
		CreatedAt:      time.Now(),
	}

	s.applications[submissionID] = application

	slog.Info("Created new consumer application", "submissionId", submissionID, "consumerId", req.ConsumerID)
	return application, nil
}

// GetConsumerApp retrieves a specific consumer application
func (s *ConsumerService) GetConsumerApp(id string) (*models.ConsumerApp, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	application, exists := s.applications[id]
	if !exists {
		return nil, fmt.Errorf("application not found")
	}

	return application, nil
}

// UpdateConsumerApp updates a consumer application
func (s *ConsumerService) UpdateConsumerApp(id string, req models.UpdateConsumerAppRequest) (*models.UpdateConsumerAppResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	application, exists := s.applications[id]
	if !exists {
		return nil, fmt.Errorf("application not found")
	}

	response := &models.UpdateConsumerAppResponse{
		ConsumerApp: application,
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
	}

	slog.Info("Updated consumer application", "submissionId", id, "status", application.Status)
	return response, nil
}

// GetAllConsumerApps retrieves all consumer applications
func (s *ConsumerService) GetAllConsumerApps() ([]*models.ConsumerApp, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	applications := make([]*models.ConsumerApp, 0, len(s.applications))
	for _, app := range s.applications {
		applications = append(applications, app)
	}

	return applications, nil
}

// GetConsumerAppsByConsumerID retrieves all applications for a specific consumer
func (s *ConsumerService) GetConsumerAppsByConsumerID(consumerID string) ([]*models.ConsumerApp, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Verify consumer exists
	_, exists := s.consumers[consumerID]
	if !exists {
		return nil, fmt.Errorf("consumer not found")
	}

	applications := make([]*models.ConsumerApp, 0)
	for _, app := range s.applications {
		if app.ConsumerID == consumerID {
			applications = append(applications, app)
		}
	}

	return applications, nil
}

// Legacy methods for backward compatibility

// GetAllApplications retrieves all consumer applications (legacy)
func (s *ConsumerService) GetAllApplications() ([]*models.Application, error) {
	apps, err := s.GetAllConsumerApps()
	if err != nil {
		return nil, err
	}

	// Convert ConsumerApp to Application (they're the same type now)
	applications := make([]*models.Application, len(apps))
	for i, app := range apps {
		applications[i] = (*models.Application)(app)
	}

	return applications, nil
}

// CreateApplication creates a new consumer application (legacy)
func (s *ConsumerService) CreateApplication(req models.CreateApplicationRequest) (*models.Application, error) {
	// Convert to new format - this is a legacy method that creates a consumer app without a consumer
	// This should not be used in new code

	// Create a default consumer for legacy compatibility
	consumerReq := models.CreateConsumerRequest{
		ConsumerName: "Legacy Consumer",
		ContactEmail: "legacy@example.com",
		PhoneNumber:  "000-000-0000",
	}

	consumer, err := s.CreateConsumer(consumerReq)
	if err != nil {
		return nil, err
	}

	consumerAppReq := models.CreateConsumerAppRequest{
		ConsumerID:     consumer.ConsumerID,
		RequiredFields: req.RequiredFields,
	}

	consumerApp, err := s.CreateConsumerApp(consumerAppReq)
	if err != nil {
		return nil, err
	}

	return (*models.Application)(consumerApp), nil
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
		ConsumerApp: application,
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

// ID generation methods

// generateConsumerID generates a unique consumer ID
func (s *ConsumerService) generateConsumerID() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "consumer_" + hex.EncodeToString(bytes), nil
}

// generateSubmissionID generates a unique submission ID
func (s *ConsumerService) generateSubmissionID() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "sub_" + hex.EncodeToString(bytes), nil
}

// generateAppID generates a unique application ID (legacy)
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
