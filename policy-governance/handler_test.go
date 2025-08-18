// handler_test.go
package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"policy-governance/internal/models"
	"testing"
)

// MockPolicyDataFetcher is a mock implementation of the PolicyDataFetcher interface
// for testing purposes. It allows us to control the behavior of GetPolicyFromDB
// without connecting to a real database.
type MockPolicyDataFetcher struct {
	// Policies is now keyed by "consumerID.subgraph.type.field"
	Policies map[string]models.PolicyRecord
	Err      error
}

// GetPolicyFromDB implements the PolicyDataFetcher interface for the mock.
// It now uses consumerID as part of the lookup key.
func (m *MockPolicyDataFetcher) GetPolicyFromDB(consumerID, subgraph, typ, field string) (models.PolicyRecord, error) { // <-- UPDATED SIGNATURE
	if m.Err != nil {
		return models.PolicyRecord{}, m.Err
	}
	key := fmt.Sprintf("%s.%s.%s.%s", consumerID, subgraph, typ, field)
	if policy, ok := m.Policies[key]; ok {
		return policy, nil
	}
	return models.PolicyRecord{Classification: models.DENY}, sql.ErrNoRows // Default to DENY if not found
}

// TestEvaluateAccessPolicy tests the core policy evaluation logic.
func TestEvaluateAccessPolicy(t *testing.T) {
	tests := []struct {
		name                   string
		mockPolicies           map[string]models.PolicyRecord
		mockErr                error
		request                models.PolicyRequest
		expectedAccessScopes   []models.AccessScope
		expectedOverallConsent bool
	}{
		{
			name: "ConsumerA - ALLOW from DB",
			mockPolicies: map[string]models.PolicyRecord{
				"consumerA.public.Product.name": {
					ConsumerID:     "consumerA",
					Classification: models.ALLOW,
				},
			},
			request: models.PolicyRequest{
				ConsumerID: "consumerA",
				RequestedFields: []models.RequestedField{
					{
						SubgraphName:   "public",
						TypeName:       "Product",
						FieldName:      "name",
						Classification: models.ALLOW,
					},
				},
			},
			expectedAccessScopes: []models.AccessScope{
				{
					SubgraphName:           "public",
					TypeName:               "Product",
					FieldName:              "name",
					ResolvedClassification: models.ALLOW,
				},
			},
			expectedOverallConsent: false,
		},
		{
			name: "ConsumerB - ALLOW_PROVIDER_CONSENT from DB",
			mockPolicies: map[string]models.PolicyRecord{
				"consumerB.dmt.VehicleInfo.engineNumber": {
					ConsumerID:     "consumerB",
					Classification: models.ALLOW_PROVIDER_CONSENT,
				},
			},
			request: models.PolicyRequest{
				ConsumerID: "consumerB",
				RequestedFields: []models.RequestedField{
					{
						SubgraphName:   "dmt",
						TypeName:       "VehicleInfo",
						FieldName:      "engineNumber",
						Classification: models.ALLOW,
					},
				},
			},
			expectedAccessScopes: []models.AccessScope{
				{
					SubgraphName:           "dmt",
					TypeName:               "VehicleInfo",
					FieldName:              "engineNumber",
					ResolvedClassification: models.ALLOW_PROVIDER_CONSENT,
				},
			},
			expectedOverallConsent: true,
		},
		{
			name: "ConsumerA - DENY for sensitive field",
			mockPolicies: map[string]models.PolicyRecord{
				"consumerA.sensitive.MedicalRecord.diagnosis": {
					ConsumerID:     "consumerA",
					Classification: models.DENY,
				},
			},
			request: models.PolicyRequest{
				ConsumerID: "consumerA",
				RequestedFields: []models.RequestedField{
					{
						SubgraphName:   "sensitive",
						TypeName:       "MedicalRecord",
						FieldName:      "diagnosis",
						Classification: models.ALLOW,
					},
				},
			},
			expectedAccessScopes: []models.AccessScope{
				{
					SubgraphName:           "sensitive",
					TypeName:               "MedicalRecord",
					FieldName:              "diagnosis",
					ResolvedClassification: models.DENY,
				},
			},
			expectedOverallConsent: false,
		},
		{
			name:         "ConsumerC - Policy not found (defaults to DENY)",
			mockPolicies: map[string]models.PolicyRecord{
				// No policy for consumerC
			},
			request: models.PolicyRequest{
				ConsumerID: "consumerC",
				RequestedFields: []models.RequestedField{
					{
						SubgraphName:   "nonexistent",
						TypeName:       "Data",
						FieldName:      "field",
						Classification: models.ALLOW,
					},
				},
			},
			expectedAccessScopes: []models.AccessScope{
				{
					SubgraphName:           "nonexistent",
					TypeName:               "Data",
					FieldName:              "field",
					ResolvedClassification: models.DENY, // Because not found in mock
				},
			},
			expectedOverallConsent: false,
		},
		{
			name: "Database error during policy fetch for consumerX",
			mockPolicies: map[string]models.PolicyRecord{
				"consumerX.public.Product.name": {
					ConsumerID:     "consumerX",
					Classification: models.ALLOW,
				},
			},
			mockErr: errors.New("database connection lost"),
			request: models.PolicyRequest{
				ConsumerID: "consumerX",
				RequestedFields: []models.RequestedField{
					{
						SubgraphName:   "public",
						TypeName:       "Product",
						FieldName:      "name",
						Classification: models.ALLOW,
					},
				},
			},
			expectedAccessScopes: []models.AccessScope{
				{
					SubgraphName:           "public",
					TypeName:               "Product",
					FieldName:              "name",
					ResolvedClassification: models.DENY, // Because of DB error, defaults to DENY
				},
			},
			expectedOverallConsent: false,
		},
		// --- Additional test cases for more coverage ---
		{
			name: "Mixed classifications for same consumer",
			mockPolicies: map[string]models.PolicyRecord{
				"consumerM.dmt.VehicleInfo.engineNumber": {Classification: models.ALLOW_PROVIDER_CONSENT},
				"consumerM.public.Product.name":          {Classification: models.ALLOW},
			},
			request: models.PolicyRequest{
				ConsumerID: "consumerM",
				RequestedFields: []models.RequestedField{
					{
						SubgraphName:   "dmt",
						TypeName:       "VehicleInfo",
						FieldName:      "engineNumber",
						Classification: models.ALLOW,
					},
					{
						SubgraphName:   "public",
						TypeName:       "Product",
						FieldName:      "name",
						Classification: models.ALLOW,
					},
				},
			},
			expectedAccessScopes: []models.AccessScope{
				{
					SubgraphName:           "dmt",
					TypeName:               "VehicleInfo",
					FieldName:              "engineNumber",
					ResolvedClassification: models.ALLOW_PROVIDER_CONSENT,
				},
				{
					SubgraphName:           "public",
					TypeName:               "Product",
					FieldName:              "name",
					ResolvedClassification: models.ALLOW,
				},
			},
			expectedOverallConsent: true,
		},
		{
			name: "Different consumers, different policies for same field",
			mockPolicies: map[string]models.PolicyRecord{
				"bankA.dmt.VehicleInfo.registrationNumber":    {Classification: models.ALLOW},
				"citizenB.dmt.VehicleInfo.registrationNumber": {Classification: models.DENY},
			},
			request: models.PolicyRequest{
				ConsumerID: "bankA",
				RequestedFields: []models.RequestedField{
					{
						SubgraphName:   "dmt",
						TypeName:       "VehicleInfo",
						FieldName:      "registrationNumber",
						Classification: models.ALLOW,
					},
				},
			},
			expectedAccessScopes: []models.AccessScope{
				{
					SubgraphName:           "dmt",
					TypeName:               "VehicleInfo",
					FieldName:              "registrationNumber",
					ResolvedClassification: models.ALLOW,
				},
			},
			expectedOverallConsent: false,
		},
		{
			name: "Another different consumer, different policies for same field",
			mockPolicies: map[string]models.PolicyRecord{
				"bankA.dmt.VehicleInfo.registrationNumber":    {Classification: models.ALLOW},
				"citizenB.dmt.VehicleInfo.registrationNumber": {Classification: models.DENY},
			},
			request: models.PolicyRequest{
				ConsumerID: "citizenB",
				RequestedFields: []models.RequestedField{
					{
						SubgraphName:   "dmt",
						TypeName:       "VehicleInfo",
						FieldName:      "registrationNumber",
						Classification: models.ALLOW,
					},
				},
			},
			expectedAccessScopes: []models.AccessScope{
				{
					SubgraphName:           "dmt",
					TypeName:               "VehicleInfo",
					FieldName:              "registrationNumber",
					ResolvedClassification: models.DENY,
				},
			},
			expectedOverallConsent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFetcher := &MockPolicyDataFetcher{
				Policies: tt.mockPolicies,
				Err:      tt.mockErr,
			}
			service := &PolicyGovernanceService{Fetcher: mockFetcher}

			actualResponse := service.EvaluateAccessPolicy(tt.request)

			if actualResponse.ConsumerID != tt.request.ConsumerID {
				t.Errorf("ConsumerID mismatch. Expected %s, got %s", tt.request.ConsumerID, actualResponse.ConsumerID)
			}
			if actualResponse.OverallConsentRequired != tt.expectedOverallConsent {
				t.Errorf("OverallConsentRequired mismatch. Expected %t, got %t", tt.expectedOverallConsent, actualResponse.OverallConsentRequired)
			}
			if len(actualResponse.AccessScopes) != len(tt.expectedAccessScopes) {
				t.Fatalf("Expected %d access scopes, got %d", len(tt.expectedAccessScopes), len(actualResponse.AccessScopes))
			}

			for i, actualScope := range actualResponse.AccessScopes {
				expectedScope := tt.expectedAccessScopes[i]
				if actualScope != expectedScope { // Direct comparison now works for AccessScope since no slices/maps
					t.Errorf("Scope %d mismatch.\nExpected: %+v\nGot:      %+v", i, expectedScope, actualScope)
				}
			}
		})
	}
}

// TestHandlePolicyRequest tests the HTTP handler.
// This section would also need updates to its mockPolicies and expectedResponseBody
// to include consumer-specific policies, similar to TestEvaluateAccessPolicy.
// For brevity, only a placeholder is shown here, as the complexity
// increases significantly with consumer-specific HTTP request mocking.
func TestHandlePolicyRequest(t *testing.T) {
	// Example for a simple HTTP test case, you'd expand this.
	// You'd need to ensure the requestBody.ConsumerID matches the mockPolicies key.
	tests := []struct {
		name                 string
		requestBody          models.PolicyRequest
		mockPolicies         map[string]models.PolicyRecord
		mockErr              error
		expectedStatus       int
		expectedResponseBody models.PolicyResponse
		sendInvalidJSON      bool
	}{
		{
			name: "Successful policy evaluation for specific consumer",
			requestBody: models.PolicyRequest{
				ConsumerID: "bankA",
				RequestedFields: []models.RequestedField{
					{
						SubgraphName:   "public",
						TypeName:       "Product",
						FieldName:      "name",
						Classification: models.ALLOW,
					},
				},
			},
			mockPolicies: map[string]models.PolicyRecord{
				"bankA.public.Product.name": {ConsumerID: "bankA", Classification: models.ALLOW},
			},
			expectedStatus: http.StatusOK,
			expectedResponseBody: models.PolicyResponse{
				ConsumerID: "bankA",
				AccessScopes: []models.AccessScope{
					{
						SubgraphName:           "public",
						TypeName:               "Product",
						FieldName:              "name",
						ResolvedClassification: models.ALLOW,
					},
				},
				OverallConsentRequired: false,
			},
		},
		{
			name: "Consent required for consumerB",
			requestBody: models.PolicyRequest{
				ConsumerID: "consumerB",
				RequestedFields: []models.RequestedField{
					{
						SubgraphName: "dmt",
						TypeName:     "VehicleInfo",
						FieldName:    "engineNumber",
					},
				},
			},
			mockPolicies: map[string]models.PolicyRecord{
				"consumerB.dmt.VehicleInfo.engineNumber": {ConsumerID: "consumerB", Classification: models.ALLOW_PROVIDER_CONSENT},
			},
			expectedStatus: http.StatusOK,
			expectedResponseBody: models.PolicyResponse{
				ConsumerID: "consumerB",
				AccessScopes: []models.AccessScope{
					{
						SubgraphName:           "dmt",
						TypeName:               "VehicleInfo",
						FieldName:              "engineNumber",
						ResolvedClassification: models.ALLOW_PROVIDER_CONSENT,
					},
				},
				OverallConsentRequired: true,
			},
		},
		{
			name: "Access Denied for consumerC on specific field",
			requestBody: models.PolicyRequest{
				ConsumerID: "consumerC",
				RequestedFields: []models.RequestedField{
					{
						SubgraphName: "sensitive",
						TypeName:     "MedicalRecord",
						FieldName:    "diagnosis",
					},
				},
			},
			mockPolicies: map[string]models.PolicyRecord{
				"consumerC.sensitive.MedicalRecord.diagnosis": {ConsumerID: "consumerC", Classification: models.DENY},
			},
			expectedStatus: http.StatusOK,
			expectedResponseBody: models.PolicyResponse{
				ConsumerID: "consumerC",
				AccessScopes: []models.AccessScope{
					{
						SubgraphName:           "sensitive",
						TypeName:               "MedicalRecord",
						FieldName:              "diagnosis",
						ResolvedClassification: models.DENY,
					},
				},
				OverallConsentRequired: false,
			},
		},
		{
			name: "Consumer not found in policies (defaults to DENY)",
			requestBody: models.PolicyRequest{
				ConsumerID: "unknownConsumer",
				RequestedFields: []models.RequestedField{
					{
						SubgraphName: "public",
						TypeName:     "Product",
						FieldName:    "name",
					},
				},
			},
			mockPolicies:   map[string]models.PolicyRecord{}, // No policies for this consumer
			expectedStatus: http.StatusOK,
			expectedResponseBody: models.PolicyResponse{
				ConsumerID: "unknownConsumer",
				AccessScopes: []models.AccessScope{
					{
						SubgraphName:           "public",
						TypeName:               "Product",
						FieldName:              "name",
						ResolvedClassification: models.DENY,
					},
				},
				OverallConsentRequired: false,
			},
		},
		// ... existing invalid JSON, DB fetch error, Method Not Allowed tests ...
		{
			name:            "Invalid JSON request body",
			requestBody:     models.PolicyRequest{}, // This struct is valid, but we'll send bad JSON
			sendInvalidJSON: true,                   // Flag to trigger sending malformed JSON
			mockPolicies:    nil,
			expectedStatus:  http.StatusBadRequest,
			expectedResponseBody: models.PolicyResponse{ // Expect default empty response on error
				ConsumerID:             "",
				AccessScopes:           []models.AccessScope{},
				OverallConsentRequired: false,
			},
		},
		{
			name: "DB fetch error (defaults to DENY by service logic)",
			requestBody: models.PolicyRequest{
				ConsumerID: "someConsumer",
				RequestedFields: []models.RequestedField{
					{
						SubgraphName:   "unknown",
						TypeName:       "Type",
						FieldName:      "field",
						Classification: models.ALLOW,
					},
				},
			},
			mockPolicies:   map[string]models.PolicyRecord{}, // No specific policy here
			mockErr:        errors.New("simulated database error"),
			expectedStatus: http.StatusOK, // Service handles DB error gracefully and returns DENY
			expectedResponseBody: models.PolicyResponse{
				ConsumerID: "someConsumer",
				AccessScopes: []models.AccessScope{
					{
						SubgraphName:           "unknown",
						TypeName:               "Type",
						FieldName:              "field",
						ResolvedClassification: models.DENY,
					},
				},
				OverallConsentRequired: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFetcher := &MockPolicyDataFetcher{
				Policies: tt.mockPolicies,
				Err:      tt.mockErr,
			}
			service := &PolicyGovernanceService{Fetcher: mockFetcher}

			var reqBodyBytes []byte
			var err error
			if tt.sendInvalidJSON {
				reqBodyBytes = []byte("{invalid json,") // Malformed JSON
			} else {
				reqBodyBytes, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/evaluate-policy", bytes.NewBuffer(reqBodyBytes))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := HandlePolicyRequest(service)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Handler returned wrong status code:\nGot:  %v\nWant: %v\nResponse Body: %s",
					status, tt.expectedStatus, rr.Body.String())
			}

			if tt.expectedStatus == http.StatusOK {
				var actualResponse models.PolicyResponse
				err := json.Unmarshal(rr.Body.Bytes(), &actualResponse)
				if err != nil {
					t.Fatalf("Could not unmarshal response: %v\nRaw body: %s", err, rr.Body.String())
				}

				if actualResponse.ConsumerID != tt.expectedResponseBody.ConsumerID {
					t.Errorf("ConsumerID mismatch.\nExpected: %s\nGot:      %s", tt.expectedResponseBody.ConsumerID, actualResponse.ConsumerID)
				}
				if actualResponse.OverallConsentRequired != tt.expectedResponseBody.OverallConsentRequired {
					t.Errorf("OverallConsentRequired mismatch.\nExpected: %t\nGot:      %t", tt.expectedResponseBody.OverallConsentRequired, actualResponse.OverallConsentRequired)
				}

				if len(actualResponse.AccessScopes) != len(tt.expectedResponseBody.AccessScopes) {
					t.Fatalf("Expected %d access scopes, got %d", len(tt.expectedResponseBody.AccessScopes), len(actualResponse.AccessScopes))
				}
				for i, actualScope := range actualResponse.AccessScopes {
					expectedScope := tt.expectedResponseBody.AccessScopes[i]
					if actualScope != expectedScope {
						t.Errorf("Scope %d mismatch.\nExpected: %+v\nGot:      %+v", i, expectedScope, actualScope)
					}
				}
			} else {
				body, _ := ioutil.ReadAll(rr.Body)
				if tt.sendInvalidJSON && !bytes.Contains(body, []byte("Invalid request payload")) {
					t.Errorf("Expected 'Invalid request payload' in error, got: %s", string(body))
				}
			}
		})
	}

	t.Run("Method Not Allowed", func(t *testing.T) {
		mockFetcher := &MockPolicyDataFetcher{}
		service := &PolicyGovernanceService{Fetcher: mockFetcher}

		req := httptest.NewRequest(http.MethodGet, "/evaluate-policy", nil)
		rr := httptest.NewRecorder()
		handler := HandlePolicyRequest(service)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusMethodNotAllowed {
			t.Errorf("Handler returned wrong status code:\nGot:  %v\nWant: %v",
				status, http.StatusMethodNotAllowed)
		}
	})
}
