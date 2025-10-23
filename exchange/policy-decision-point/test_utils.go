package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gov-dx-sandbox/exchange/policy-decision-point/models"
	"github.com/open-policy-agent/opa/rego"
)

// MockDatabaseService is a mock implementation of DatabaseService for testing
type MockDatabaseService struct {
	policyMetadata []models.PolicyMetadata
}

// NewMockDatabaseService creates a new mock database service with test data
func NewMockDatabaseService() *MockDatabaseService {
	// Create test policy metadata
	policyMetadata := []models.PolicyMetadata{
		{
			ID:                "test-id-1",
			FieldName:         "person.fullName",
			DisplayName:       stringPtr("Full Name"),
			Description:       stringPtr("Complete name of the person"),
			Source:            "drp",
			IsOwner:           true,
			Owner:             "CITIZEN",
			AccessControlType: "public",
			AllowList: []models.AllowListEntry{
				{
					ApplicationID: "test-app",
					ExpiresAt:     time.Now().Add(24 * time.Hour).Unix(),
				},
			},
			CreatedAt: time.Now().Format(time.RFC3339),
			UpdatedAt: time.Now().Format(time.RFC3339),
		},
		{
			ID:                "test-id-2",
			FieldName:         "person.nic",
			DisplayName:       stringPtr("NIC Number"),
			Description:       stringPtr("National Identity Card number"),
			Source:            "drp",
			IsOwner:           true,
			Owner:             "CITIZEN",
			AccessControlType: "restricted",
			AllowList: []models.AllowListEntry{
				{
					ApplicationID: "test-app",
					ExpiresAt:     time.Now().Add(24 * time.Hour).Unix(),
				},
			},
			CreatedAt: time.Now().Format(time.RFC3339),
			UpdatedAt: time.Now().Format(time.RFC3339),
		},
		{
			ID:                "test-id-3",
			FieldName:         "person.birthDate",
			DisplayName:       stringPtr("Birth Date"),
			Description:       stringPtr("Date of birth"),
			Source:            "drp",
			IsOwner:           true,
			Owner:             "CITIZEN",
			AccessControlType: "restricted",
			AllowList: []models.AllowListEntry{
				{
					ApplicationID: "test-app",
					ExpiresAt:     time.Now().Add(24 * time.Hour).Unix(),
				},
			},
			CreatedAt: time.Now().Format(time.RFC3339),
			UpdatedAt: time.Now().Format(time.RFC3339),
		},
		{
			ID:                "test-id-4",
			FieldName:         "public.field",
			DisplayName:       stringPtr("Public Field"),
			Description:       stringPtr("A public field"),
			Source:            "external",
			IsOwner:           false,
			Owner:             "CITIZEN",
			AccessControlType: "public",
			AllowList:         []models.AllowListEntry{}, // Empty allow list means public access
			CreatedAt:         time.Now().Format(time.RFC3339),
			UpdatedAt:         time.Now().Format(time.RFC3339),
		},
	}

	return &MockDatabaseService{
		policyMetadata: policyMetadata,
	}
}

// CreatePolicyMetadata creates a new policy metadata record
func (m *MockDatabaseService) CreatePolicyMetadata(req *models.PolicyMetadataCreateRequest) (string, error) {
	id := fmt.Sprintf("test-id-%d", len(m.policyMetadata)+1)

	// Convert allow list from request
	allowList := make([]models.AllowListEntry, len(req.AllowList))
	for i, entry := range req.AllowList {
		allowList[i] = models.AllowListEntry{
			ApplicationID: entry.ApplicationID,
			ExpiresAt:     entry.ExpiresAt,
		}
	}

	newRecord := models.PolicyMetadata{
		ID:                id,
		FieldName:         req.FieldName,
		DisplayName:       stringPtr(req.DisplayName),
		Description:       stringPtr(req.Description),
		Source:            req.Source,
		IsOwner:           req.IsOwner,
		Owner:             "CITIZEN",
		AccessControlType: req.AccessControlType,
		AllowList:         allowList,
		CreatedAt:         time.Now().Format(time.RFC3339),
		UpdatedAt:         time.Now().Format(time.RFC3339),
	}

	m.policyMetadata = append(m.policyMetadata, newRecord)
	return id, nil
}

// UpdateAllowList updates the allow list for a specific field
func (m *MockDatabaseService) UpdateAllowList(req *models.AllowListUpdateRequest) error {
	// Parse expires_at timestamp
	expiresAt, err := time.Parse(time.RFC3339, req.ExpiresAt)
	if err != nil {
		return fmt.Errorf("invalid expires_at format, expected RFC3339: %w", err)
	}

	// Find the field and update its allow list
	for i, record := range m.policyMetadata {
		if record.FieldName == req.FieldName {
			// Check if application already exists in allow list
			found := false
			for j, entry := range record.AllowList {
				if entry.ApplicationID == req.ApplicationID {
					// Update existing entry
					record.AllowList[j] = models.AllowListEntry{
						ApplicationID: req.ApplicationID,
						ExpiresAt:     expiresAt.Unix(),
					}
					found = true
					break
				}
			}

			// Add new entry if not found
			if !found {
				newEntry := models.AllowListEntry{
					ApplicationID: req.ApplicationID,
					ExpiresAt:     expiresAt.Unix(),
				}
				record.AllowList = append(record.AllowList, newEntry)
			}

			record.UpdatedAt = time.Now().Format(time.RFC3339)
			m.policyMetadata[i] = record
			return nil
		}
	}

	return fmt.Errorf("field %s not found", req.FieldName)
}

// GetAllPolicyMetadata returns mock policy metadata for testing
func (m *MockDatabaseService) GetAllPolicyMetadata() (map[string]interface{}, error) {
	// Convert mock data to the expected format
	fields := make(map[string]interface{})
	for _, record := range m.policyMetadata {
		fieldMetadata := map[string]interface{}{
			"owner":               record.Owner,
			"provider":            record.Source,
			"is_owner":            record.IsOwner,
			"access_control_type": record.AccessControlType,
			"allow_list":          record.AllowList,
		}

		if record.DisplayName != nil {
			fieldMetadata["display_name"] = *record.DisplayName
		}
		if record.Description != nil {
			fieldMetadata["description"] = *record.Description
		}

		fields[record.FieldName] = fieldMetadata
	}

	return map[string]interface{}{
		"fields": fields,
	}, nil
}

// Close is a no-op for testing
func (m *MockDatabaseService) Close() error {
	return nil
}

// NewMockPolicyEvaluator creates a policy evaluator with mock database service
func NewMockPolicyEvaluator(ctx context.Context) (*PolicyEvaluator, error) {
	// Create mock database service
	dbService := NewMockDatabaseService()

	// Create policy evaluator with mock data
	return createPolicyEvaluatorWithData(ctx, dbService)
}

// createPolicyEvaluatorWithData is a helper function to create a policy evaluator with specific data
func createPolicyEvaluatorWithData(ctx context.Context, dbService DatabaseServiceInterface) (*PolicyEvaluator, error) {
	// Create test data structure that matches the new policy metadata format
	dataModule := `
		package opendif.authz

		policy_metadata = {
			"fields": {
				"person.fullName": {
					"owner": "CITIZEN",
					"provider": "primary",
					"is_owner": true,
					"access_control_type": "public",
					"allow_list": [
						{
							"application_id": "test-app",
							"expires_at": 1704067199
						}
					]
				},
				"person.nic": {
					"owner": "CITIZEN", 
					"provider": "primary",
					"is_owner": true,
					"access_control_type": "restricted",
					"allow_list": [
						{
							"application_id": "test-app",
							"expires_at": 1704067199
						}
					]
				},
				"person.birthDate": {
					"owner": "CITIZEN",
					"provider": "primary", 
					"is_owner": false,
					"access_control_type": "restricted",
					"allow_list": [
						{
							"application_id": "test-app",
							"expires_at": 1704067199
						}
					]
				},
				"public.field": {
					"owner": "CITIZEN",
					"provider": "primary",
					"is_owner": false,
					"access_control_type": "public",
					"allow_list": []
				}
			}
		}
		`

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

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
