package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/models"
)

type ConsumerService struct {
	consumers          map[string]*models.Consumer
	applications       map[string]*models.ConsumerApp
	clientMappings     map[string]*models.ClientMapping     // Maps clientId -> ClientMapping
	consumerMappings   map[string]*models.ClientMapping     // Maps consumerId -> ClientMapping
	credentialMappings map[string]*models.CredentialMapping // Maps apiKey -> CredentialMapping
	mutex              sync.RWMutex
}

func NewConsumerService() *ConsumerService {
	return &ConsumerService{
		consumers:          make(map[string]*models.Consumer),
		applications:       make(map[string]*models.ConsumerApp),
		credentialMappings: make(map[string]*models.CredentialMapping),
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

// UpdateConsumer updates a consumer
func (s *ConsumerService) UpdateConsumer(id string, req models.UpdateConsumerRequest) (*models.Consumer, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	consumer, exists := s.consumers[id]
	if !exists {
		return nil, fmt.Errorf("consumer not found")
	}

	// Update fields if provided
	if req.ConsumerName != nil {
		consumer.ConsumerName = *req.ConsumerName
	}
	if req.ContactEmail != nil {
		consumer.ContactEmail = *req.ContactEmail
	}
	if req.PhoneNumber != nil {
		consumer.PhoneNumber = *req.PhoneNumber
	}

	slog.Info("Updated consumer", "consumerId", id)
	return consumer, nil
}

// DeleteConsumer deletes a consumer
func (s *ConsumerService) DeleteConsumer(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exists := s.consumers[id]
	if !exists {
		return fmt.Errorf("consumer not found")
	}

	delete(s.consumers, id)
	slog.Info("Deleted consumer", "consumerId", id)
	return nil
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

		// Create credential mapping to platform's Asgardeo client
		// Get platform's Asgardeo client credentials (these should be configured)
		platformClientID := s.getPlatformAsgardeoClientID()
		platformClientSecret := s.getPlatformAsgardeoClientSecret()

		// Create credential mapping directly
		mapping := &models.CredentialMapping{
			APIKey:               credentials.APIKey,
			APISecret:            credentials.APISecret,
			AsgardeoClientID:     platformClientID,
			AsgardeoClientSecret: platformClientSecret,
			ConsumerID:           application.ConsumerID,
		}
		s.credentialMappings[credentials.APIKey] = mapping
		slog.Info("Created credential mapping for approved consumer",
			"consumerId", application.ConsumerID,
			"apiKey", credentials.APIKey)

		// For now, assign a default provider ID when approved
		// In a real system, this would be based on business logic
		response.ProviderID = "default_provider"
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

		// Create credential mapping to platform's Asgardeo client
		// Get platform's Asgardeo client credentials (these should be configured)
		platformClientID := s.getPlatformAsgardeoClientID()
		platformClientSecret := s.getPlatformAsgardeoClientSecret()

		// Create credential mapping directly (we already hold the lock)
		mapping := &models.CredentialMapping{
			APIKey:               credentials.APIKey,
			APISecret:            credentials.APISecret,
			AsgardeoClientID:     platformClientID,
			AsgardeoClientSecret: platformClientSecret,
			ConsumerID:           application.ConsumerID,
		}
		s.credentialMappings[credentials.APIKey] = mapping
		slog.Info("Created credential mapping for approved consumer",
			"consumerId", application.ConsumerID,
			"apiKey", credentials.APIKey)

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

// getPlatformAsgardeoClientID returns the platform's Asgardeo client ID
func (s *ConsumerService) getPlatformAsgardeoClientID() string {
	// Get from environment variable
	clientID := os.Getenv("ASGARDEO_CLIENT_ID")
	if clientID == "" {
		slog.Warn("ASGARDEO_CLIENT_ID not set, using default")
		return "platform_asgardeo_client_id"
	}
	return clientID
}

// getPlatformAsgardeoClientSecret returns the platform's Asgardeo client secret
func (s *ConsumerService) getPlatformAsgardeoClientSecret() string {
	// Get from environment variable
	clientSecret := os.Getenv("ASGARDEO_CLIENT_SECRET")
	if clientSecret == "" {
		slog.Warn("ASGARDEO_CLIENT_SECRET not set, using default")
		return "platform_asgardeo_client_secret"
	}
	return clientSecret
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

// validateAndGetMapping validates API credentials and returns the credential mapping
func (s *ConsumerService) ValidateAndGetMapping(apiKey, apiSecret string) (*models.CredentialMapping, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Debug: Log all available credential mappings
	slog.Info("Available credential mappings", "count", len(s.credentialMappings))
	for key, mapping := range s.credentialMappings {
		slog.Info("Credential mapping", "apiKey", key[:8]+"...", "consumerId", mapping.ConsumerID)
	}

	// Find credential mapping
	mapping, exists := s.credentialMappings[apiKey]
	if !exists {
		slog.Warn("Invalid API key provided", "apiKey", apiKey[:8]+"...", "totalMappings", len(s.credentialMappings))
		return nil, fmt.Errorf("invalid API key")
	}

	// Validate API secret
	if mapping.APISecret != apiSecret {
		slog.Warn("Invalid API secret provided", "apiKey", apiKey[:8]+"...")
		return nil, fmt.Errorf("invalid API secret")
	}

	// Validate mapping has required fields
	if mapping.AsgardeoClientID == "" || mapping.AsgardeoClientSecret == "" {
		slog.Error("Credential mapping missing Asgardeo credentials", "consumerId", mapping.ConsumerID)
		return nil, fmt.Errorf("credential mapping is incomplete")
	}

	return mapping, nil
}

// IsCredentialMappingConfigured checks if the credential mapping system is properly configured
func (s *ConsumerService) IsCredentialMappingConfigured() bool {
	clientID := s.getPlatformAsgardeoClientID()
	clientSecret := s.getPlatformAsgardeoClientSecret()

	// Check if we have valid Asgardeo credentials (not the default placeholders)
	return clientID != "platform_asgardeo_client_id" && clientSecret != "platform_asgardeo_client_secret"
}

// validateCredentials validates API credentials against stored applications (legacy)
func (s *ConsumerService) validateCredentials(apiKey, apiSecret string) (*models.ConsumerApp, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Find application with matching credentials
	for _, app := range s.applications {
		if app.Credentials != nil &&
			app.Credentials.APIKey == apiKey &&
			app.Credentials.APISecret == apiSecret {
			return app, nil
		}
	}

	return nil, fmt.Errorf("invalid credentials")
}

// CreateCredentialMapping creates a new credential mapping
func (s *ConsumerService) CreateCredentialMapping(consumerID, asgardeoClientID, asgardeoClientSecret string) (*models.CredentialMapping, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Generate API credentials
	apiKeyBytes := make([]byte, 16)
	if _, err := rand.Read(apiKeyBytes); err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	apiSecretBytes := make([]byte, 32)
	if _, err := rand.Read(apiSecretBytes); err != nil {
		return nil, fmt.Errorf("failed to generate API secret: %w", err)
	}

	apiKey := hex.EncodeToString(apiKeyBytes)
	apiSecret := hex.EncodeToString(apiSecretBytes)

	mapping := &models.CredentialMapping{
		APIKey:               apiKey,
		APISecret:            apiSecret,
		AsgardeoClientID:     asgardeoClientID,
		AsgardeoClientSecret: asgardeoClientSecret,
		ConsumerID:           consumerID,
	}

	s.credentialMappings[apiKey] = mapping

	slog.Info("Created credential mapping",
		"consumerId", consumerID,
		"asgardeoClientId", asgardeoClientID,
		"apiKey", apiKey)

	return mapping, nil
}

// CreateCredentialMappingWithCredentials creates a credential mapping using existing API credentials
func (s *ConsumerService) CreateCredentialMappingWithCredentials(apiKey, apiSecret, consumerID, asgardeoClientID, asgardeoClientSecret string) (*models.CredentialMapping, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	slog.Info("Creating credential mapping",
		"consumerId", consumerID,
		"apiKey", apiKey[:8]+"...",
		"asgardeoClientId", asgardeoClientID)

	mapping := &models.CredentialMapping{
		APIKey:               apiKey,
		APISecret:            apiSecret,
		AsgardeoClientID:     asgardeoClientID,
		AsgardeoClientSecret: asgardeoClientSecret,
		ConsumerID:           consumerID,
	}

	s.credentialMappings[apiKey] = mapping

	slog.Info("Created credential mapping with existing credentials",
		"consumerId", consumerID,
		"asgardeoClientId", asgardeoClientID,
		"apiKey", apiKey,
		"totalMappings", len(s.credentialMappings))

	return mapping, nil
}
