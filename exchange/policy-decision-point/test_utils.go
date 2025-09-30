package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/models"
	"github.com/open-policy-agent/opa/rego"
)

// MockDatabaseService is a mock implementation of DatabaseService for testing
type MockDatabaseService struct {
	metadata *models.ProviderMetadata
}

// NewMockDatabaseService creates a new mock database service with test data
func NewMockDatabaseService() *MockDatabaseService {
	// Create test provider metadata
	metadata := &models.ProviderMetadata{
		Fields: map[string]models.ProviderMetadataField{
			"person.fullName": {
				Owner:             "citizen",
				Provider:          "drp",
				ConsentRequired:   false,
				AccessControlType: "public",
				AllowList: []models.PDPAllowListEntry{
					{
						ConsumerID:    "test-app",
						ExpiresAt:     time.Now().Add(24 * time.Hour).Unix(),
						GrantDuration: "30d",
					},
				},
			},
			"person.nic": {
				Owner:             "citizen",
				Provider:          "drp",
				ConsentRequired:   true,
				AccessControlType: "public",
				AllowList: []models.PDPAllowListEntry{
					{
						ConsumerID:    "test-app",
						ExpiresAt:     time.Now().Add(24 * time.Hour).Unix(),
						GrantDuration: "30d",
					},
				},
			},
			"person.birthDate": {
				Owner:             "citizen",
				Provider:          "drp",
				ConsentRequired:   false,
				AccessControlType: "restricted",
				AllowList: []models.PDPAllowListEntry{
					{
						ConsumerID:    "test-app",
						ExpiresAt:     time.Now().Add(24 * time.Hour).Unix(),
						GrantDuration: "30d",
					},
				},
			},
			"public.field": {
				Owner:             "external",
				Provider:          "external-provider",
				ConsentRequired:   false,
				AccessControlType: "public",
				AllowList:         []models.PDPAllowListEntry{}, // Empty allow list means public access
			},
		},
	}

	return &MockDatabaseService{
		metadata: metadata,
	}
}

// GetAllProviderMetadata returns the mock provider metadata
func (m *MockDatabaseService) GetAllProviderMetadata() (*models.ProviderMetadata, error) {
	return m.metadata, nil
}

// UpdateProviderField is a no-op for testing
func (m *MockDatabaseService) UpdateProviderField(fieldName string, field models.ProviderMetadataField) error {
	if m.metadata.Fields == nil {
		m.metadata.Fields = make(map[string]models.ProviderMetadataField)
	}
	m.metadata.Fields[fieldName] = field
	return nil
}

// UpdateProviderMetadata is a no-op for testing
func (m *MockDatabaseService) UpdateProviderMetadata(metadata *models.ProviderMetadata) error {
	m.metadata = metadata
	return nil
}

// Close is a no-op for testing
func (m *MockDatabaseService) Close() error {
	return nil
}

// NewMockPolicyEvaluator creates a policy evaluator with mock database service
func NewMockPolicyEvaluator(ctx context.Context) (*PolicyEvaluator, error) {
	// Create mock database service
	dbService := NewMockDatabaseService()

	// Load provider metadata from mock service
	providerMetadata, err := dbService.GetAllProviderMetadata()
	if err != nil {
		return nil, err
	}

	// Create policy evaluator with mock data
	return createPolicyEvaluatorWithData(ctx, providerMetadata, dbService)
}

// createPolicyEvaluatorWithData is a helper function to create a policy evaluator with specific data
func createPolicyEvaluatorWithData(ctx context.Context, metadata *models.ProviderMetadata, dbService DatabaseServiceInterface) (*PolicyEvaluator, error) {
	// Convert data to JSON string for embedding in policy
	providerMetadataJSON, _ := json.Marshal(metadata)

	// Create a module with the data embedded as JSON values
	dataModule := fmt.Sprintf(`
		package opendif.authz

		provider_metadata = %s
		`, string(providerMetadataJSON))

	query := "data.opendif.authz.decision"
	r := rego.New(
		rego.Query(query),
		rego.Load([]string{"./policies"}, nil), // Load policy files
		rego.Module("data.rego", dataModule),   // Add data as module
	)

	pq, err := r.PrepareForEval(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare OPA query: %w", err)
	}

	return &PolicyEvaluator{preparedQuery: pq, dbService: dbService}, nil
}
