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
	Policies map[string]models.PolicyRecord // Map key: "subgraph.type.field"
	Err      error
}

// GetPolicyFromDB implements the PolicyDataFetcher interface for the mock.
func (m *MockPolicyDataFetcher) GetPolicyFromDB(subgraph, typ, field string) (models.PolicyRecord, error) {
	if m.Err != nil {
		return models.PolicyRecord{}, m.Err
	}
	key := fmt.Sprintf("%s.%s.%s", subgraph, typ, field)
	if policy, ok := m.Policies[key]; ok {
		return policy, nil
	}
	return models.PolicyRecord{Classification: models.DENIED}, sql.ErrNoRows // Default to DENIED if not found
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
			name: "All ALLOW from DB",
			mockPolicies: map[string]models.PolicyRecord{
				"public.Product.name": {
					Classification: models.ALLOW,
				},
			},
			request: models.PolicyRequest{
				ConsumerID: "test-consumer",
				RequestedFields: []models.RequestedField{
					{
						SubgraphName:   "public",
						TypeName:       "Product",
						FieldName:      "name",
						Classification: models.ALLOW, // Request hint
					},
				},
			},
			expectedAccessScopes: []models.AccessScope{
				{
					SubgraphName:           "public",
					TypeName:               "Product",
					FieldName:              "name",
					ResolvedClassification: models.ALLOW,
					ConsentRequired:        false,
					ConsentType:            "",
				},
			},
			expectedOverallConsent: false,
		},
		{
			name: "ALLOW_PROVIDER_CONSENT from DB",
			mockPolicies: map[string]models.PolicyRecord{
				"dmt.VehicleInfo.engineNumber": {
					Classification: models.ALLOW_PROVIDER_CONSENT,
				},
			},
			request: models.PolicyRequest{
				ConsumerID: "test-consumer",
				RequestedFields: []models.RequestedField{
					{
						SubgraphName:   "dmt",
						TypeName:       "VehicleInfo",
						FieldName:      "engineNumber",
						Classification: models.ALLOW, // Request hint
					},
				},
			},
			expectedAccessScopes: []models.AccessScope{
				{
					SubgraphName:           "dmt",
					TypeName:               "VehicleInfo",
					FieldName:              "engineNumber",
					ResolvedClassification: models.ALLOW_PROVIDER_CONSENT,
					ConsentRequired:        true,
					ConsentType:            "provider",
				},
			},
			expectedOverallConsent: true,
		},
		{
			name: "ALLOW_CITIZEN_CONSENT from DB with context",
			mockPolicies: map[string]models.PolicyRecord{
				"drp.PersonData.photo": {
					Classification: models.ALLOW_CITIZEN_CONSENT,
				},
			},
			request: models.PolicyRequest{
				ConsumerID: "test-consumer",
				RequestedFields: []models.RequestedField{
					{
						SubgraphName:   "drp",
						TypeName:       "PersonData",
						FieldName:      "photo",
						Classification: models.ALLOW, // Request hint
						Context:        models.Context{"citizenId": "citizen-123"},
					},
				},
			},
			expectedAccessScopes: []models.AccessScope{
				{
					SubgraphName:           "drp",
					TypeName:               "PersonData",
					FieldName:              "photo",
					ResolvedClassification: models.ALLOW_CITIZEN_CONSENT,
					ConsentRequired:        true,
					ConsentType:            "citizen",
				},
			},
			expectedOverallConsent: true,
		},
		{
			name: "ALLOW_CONSENT with citizen context",
			mockPolicies: map[string]models.PolicyRecord{
				"finance.Account.balance": {
					Classification: models.ALLOW_CONSENT,
				},
			},
			request: models.PolicyRequest{
				ConsumerID: "test-consumer",
				RequestedFields: []models.RequestedField{
					{
						SubgraphName:   "finance",
						TypeName:       "Account",
						FieldName:      "balance",
						Classification: models.ALLOW,
						Context:        models.Context{"citizenId": "some-id"},
					},
				},
			},
			expectedAccessScopes: []models.AccessScope{
				{
					SubgraphName:           "finance",
					TypeName:               "Account",
					FieldName:              "balance",
					ResolvedClassification: models.ALLOW_CONSENT,
					ConsentRequired:        true,
					ConsentType:            "citizen", // Determined by context
				},
			},
			expectedOverallConsent: true,
		},
		{
			name: "ALLOW_CONSENT without citizen context (defaults to provider)",
			mockPolicies: map[string]models.PolicyRecord{
				"finance.Account.balance": {
					Classification: models.ALLOW_CONSENT,
				},
			},
			request: models.PolicyRequest{
				ConsumerID: "test-consumer",
				RequestedFields: []models.RequestedField{
					{
						SubgraphName:   "finance",
						TypeName:       "Account",
						FieldName:      "balance",
						Classification: models.ALLOW,
						Context:        models.Context{},
					},
				},
			},
			expectedAccessScopes: []models.AccessScope{
				{
					SubgraphName:           "finance",
					TypeName:               "Account",
					FieldName:              "balance",
					ResolvedClassification: models.ALLOW_CONSENT,
					ConsentRequired:        true,
					ConsentType:            "provider", // Default if no citizenId in context
				},
			},
			expectedOverallConsent: true,
		},
		{
			name: "DENIED from DB",
			mockPolicies: map[string]models.PolicyRecord{
				"sensitive.MedicalRecord.diagnosis": {
					Classification: models.DENIED,
				},
			},
			request: models.PolicyRequest{
				ConsumerID: "test-consumer",
				RequestedFields: []models.RequestedField{
					{
						SubgraphName:   "sensitive",
						TypeName:       "MedicalRecord",
						FieldName:      "diagnosis",
						Classification: models.ALLOW, // Request hint
					},
				},
			},
			expectedAccessScopes: []models.AccessScope{
				{
					SubgraphName:           "sensitive",
					TypeName:               "MedicalRecord",
					FieldName:              "diagnosis",
					ResolvedClassification: models.DENIED,
					ConsentRequired:        false,
					ConsentType:            "",
				},
			},
			expectedOverallConsent: false,
		},
		{
			name:         "Policy not found in DB (defaults to DENIED)",
			mockPolicies: map[string]models.PolicyRecord{
				// No policy for "nonexistent.Data.field"
			},
			request: models.PolicyRequest{
				ConsumerID: "test-consumer",
				RequestedFields: []models.RequestedField{
					{
						SubgraphName:   "nonexistent",
						TypeName:       "Data",
						FieldName:      "field",
						Classification: models.ALLOW, // Request hint
					},
				},
			},
			expectedAccessScopes: []models.AccessScope{
				{
					SubgraphName:           "nonexistent",
					TypeName:               "Data",
					FieldName:              "field",
					ResolvedClassification: models.DENIED, // Because not found and GetPolicyFromDB returns DENIED
					ConsentRequired:        false,
					ConsentType:            "",
				},
			},
			expectedOverallConsent: false,
		},
		{
			name: "Database error during policy fetch",
			mockPolicies: map[string]models.PolicyRecord{
				"public.Product.name": {
					Classification: models.ALLOW,
				},
			},
			mockErr: errors.New("database connection lost"),
			request: models.PolicyRequest{
				ConsumerID: "test-consumer",
				RequestedFields: []models.RequestedField{
					{
						SubgraphName:   "public",
						TypeName:       "Product",
						FieldName:      "name",
						Classification: models.ALLOW, // Request hint
					},
				},
			},
			expectedAccessScopes: []models.AccessScope{
				{
					SubgraphName:           "public",
					TypeName:               "Product",
					FieldName:              "name",
					ResolvedClassification: models.DENIED, // Because of DB error, defaults to DENIED
					ConsentRequired:        false,
					ConsentType:            "",
				},
			},
			expectedOverallConsent: false,
		},
		{
			name: "Mixed classifications and consent requirements",
			mockPolicies: map[string]models.PolicyRecord{
				"dmt.VehicleInfo.engineNumber":      {Classification: models.ALLOW_PROVIDER_CONSENT},
				"drp.PersonData.photo":              {Classification: models.ALLOW_CITIZEN_CONSENT},
				"public.Product.name":               {Classification: models.ALLOW},
				"sensitive.MedicalRecord.diagnosis": {Classification: models.DENIED},
				"finance.Account.balance":           {Classification: models.ALLOW_CONSENT},
			},
			request: models.PolicyRequest{
				ConsumerID: "test-consumer-mixed",
				RequestedFields: []models.RequestedField{
					{
						SubgraphName:   "dmt",
						TypeName:       "VehicleInfo",
						FieldName:      "engineNumber",
						Classification: models.ALLOW,
					},
					{
						SubgraphName:   "drp",
						TypeName:       "PersonData",
						FieldName:      "photo",
						Classification: models.ALLOW,
						Context:        models.Context{"citizenId": "citizen-999"},
					},
					{
						SubgraphName:   "public",
						TypeName:       "Product",
						FieldName:      "name",
						Classification: models.ALLOW,
					},
					{
						SubgraphName:   "sensitive",
						TypeName:       "MedicalRecord",
						FieldName:      "diagnosis",
						Classification: models.ALLOW,
					},
					{
						SubgraphName:   "finance",
						TypeName:       "Account",
						FieldName:      "balance",
						Classification: models.ALLOW,
						Context:        models.Context{}, // This should default to provider consent
					},
				},
			},
			expectedAccessScopes: []models.AccessScope{
				{
					SubgraphName:           "dmt",
					TypeName:               "VehicleInfo",
					FieldName:              "engineNumber",
					ResolvedClassification: models.ALLOW_PROVIDER_CONSENT,
					ConsentRequired:        true,
					ConsentType:            "provider",
				},
				{
					SubgraphName:           "drp",
					TypeName:               "PersonData",
					FieldName:              "photo",
					ResolvedClassification: models.ALLOW_CITIZEN_CONSENT,
					ConsentRequired:        true,
					ConsentType:            "citizen",
				},
				{
					SubgraphName:           "public",
					TypeName:               "Product",
					FieldName:              "name",
					ResolvedClassification: models.ALLOW,
					ConsentRequired:        false,
					ConsentType:            "",
				},
				{
					SubgraphName:           "sensitive",
					TypeName:               "MedicalRecord",
					FieldName:              "diagnosis",
					ResolvedClassification: models.DENIED,
					ConsentRequired:        false,
					ConsentType:            "",
				},
				{
					SubgraphName:           "finance",
					TypeName:               "Account",
					FieldName:              "balance",
					ResolvedClassification: models.ALLOW_CONSENT,
					ConsentRequired:        true,
					ConsentType:            "provider",
				},
			},
			expectedOverallConsent: true,
		},
		{
			name: "Empty requested fields",
			request: models.PolicyRequest{
				ConsumerID:      "test-consumer-empty",
				RequestedFields: []models.RequestedField{},
			},
			expectedAccessScopes:   []models.AccessScope{},
			expectedOverallConsent: false,
		},
		{
			name:         "Request with undefined Classification (should default to DENIED from DB if not found)",
			mockPolicies: map[string]models.PolicyRecord{
				// No policy for this specific field in mock
			},
			request: models.PolicyRequest{
				ConsumerID: "test-consumer",
				RequestedFields: []models.RequestedField{
					{
						SubgraphName: "unknown-subgraph",
						TypeName:     "UnknownType",
						FieldName:    "unknownField",
						// Classification field is intentionally missing or empty
					},
				},
			},
			expectedAccessScopes: []models.AccessScope{
				{
					SubgraphName:           "unknown-subgraph",
					TypeName:               "UnknownType",
					FieldName:              "unknownField",
					ResolvedClassification: models.DENIED, // Default behavior when not found in mock
					ConsentRequired:        false,
					ConsentType:            "",
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

			// Compare ConsumerID
			if actualResponse.ConsumerID != tt.request.ConsumerID {
				t.Errorf("ConsumerID mismatch. Expected %s, got %s", tt.request.ConsumerID, actualResponse.ConsumerID)
			}

			// Compare OverallConsentRequired
			if actualResponse.OverallConsentRequired != tt.expectedOverallConsent {
				t.Errorf("OverallConsentRequired mismatch. Expected %t, got %t", tt.expectedOverallConsent, actualResponse.OverallConsentRequired)
			}

			// Compare AccessScopes (order might matter based on iteration, so compare elements)
			if len(actualResponse.AccessScopes) != len(tt.expectedAccessScopes) {
				t.Fatalf("Expected %d access scopes, got %d", len(tt.expectedAccessScopes), len(actualResponse.AccessScopes))
			}

			// For robust comparison of slices of structs, it's better to sort them
			// or iterate and compare each field of each struct if order isn't guaranteed.
			// Given the current processing order is preserved, a direct index comparison is used.
			for i, actualScope := range actualResponse.AccessScopes {
				expectedScope := tt.expectedAccessScopes[i]
				if actualScope != expectedScope {
					t.Errorf("Scope %d mismatch.\nExpected: %+v\nGot:      %+v", i, expectedScope, actualScope)
				}
			}
		})
	}
}

// TestHandlePolicyRequest tests the HTTP handler.
func TestHandlePolicyRequest(t *testing.T) {
	tests := []struct {
		name                 string
		requestBody          models.PolicyRequest
		mockPolicies         map[string]models.PolicyRecord
		mockErr              error
		expectedStatus       int
		expectedResponseBody models.PolicyResponse // Use models.PolicyResponse for expected body
		sendInvalidJSON      bool                  // Flag to explicitly send invalid JSON
	}{
		{
			name: "Successful policy evaluation",
			requestBody: models.PolicyRequest{
				ConsumerID: "http-consumer",
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
				"public.Product.name": {Classification: models.ALLOW},
			},
			expectedStatus: http.StatusOK,
			expectedResponseBody: models.PolicyResponse{
				ConsumerID: "http-consumer",
				AccessScopes: []models.AccessScope{
					{
						SubgraphName:           "public",
						TypeName:               "Product",
						FieldName:              "name",
						ResolvedClassification: models.ALLOW,
						ConsentRequired:        false,
					},
				},
				OverallConsentRequired: false,
			},
		},
		{
			name: "Consent required scenario",
			requestBody: models.PolicyRequest{
				ConsumerID: "http-consumer",
				RequestedFields: []models.RequestedField{
					{
						SubgraphName:   "dmt",
						TypeName:       "VehicleInfo",
						FieldName:      "engineNumber",
						Classification: models.ALLOW,
					},
				},
			},
			mockPolicies: map[string]models.PolicyRecord{
				"dmt.VehicleInfo.engineNumber": {Classification: models.ALLOW_PROVIDER_CONSENT},
			},
			expectedStatus: http.StatusOK,
			expectedResponseBody: models.PolicyResponse{
				ConsumerID: "http-consumer",
				AccessScopes: []models.AccessScope{
					{
						SubgraphName:           "dmt",
						TypeName:               "VehicleInfo",
						FieldName:              "engineNumber",
						ResolvedClassification: models.ALLOW_PROVIDER_CONSENT,
						ConsentRequired:        true,
						ConsentType:            "provider",
					},
				},
				OverallConsentRequired: true,
			},
		},
		{
			name: "Empty requested fields in HTTP request",
			requestBody: models.PolicyRequest{
				ConsumerID:      "http-consumer-empty",
				RequestedFields: []models.RequestedField{},
			},
			mockPolicies:   nil, // Not used for this scenario
			expectedStatus: http.StatusOK,
			expectedResponseBody: models.PolicyResponse{
				ConsumerID:             "http-consumer-empty",
				AccessScopes:           []models.AccessScope{},
				OverallConsentRequired: false,
			},
		},
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
			name: "DB fetch error (defaults to DENIED by service logic)",
			requestBody: models.PolicyRequest{
				ConsumerID: "http-consumer",
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
			expectedStatus: http.StatusOK, // Service handles DB error gracefully and returns DENIED
			expectedResponseBody: models.PolicyResponse{
				ConsumerID: "http-consumer",
				AccessScopes: []models.AccessScope{
					{
						SubgraphName:           "unknown",
						TypeName:               "Type",
						FieldName:              "field",
						ResolvedClassification: models.DENIED,
						ConsentRequired:        false,
					},
				},
				OverallConsentRequired: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up mock data fetcher
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

			// Create a mock HTTP request
			req := httptest.NewRequest(http.MethodPost, "/evaluate-policy", bytes.NewBuffer(reqBodyBytes))
			req.Header.Set("Content-Type", "application/json")

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Call the handler function
			handler := HandlePolicyRequest(service) // Get the http.HandlerFunc
			handler.ServeHTTP(rr, req)

			// Check the status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Handler returned wrong status code:\nGot:  %v\nWant: %v\nResponse Body: %s",
					status, tt.expectedStatus, rr.Body.String())
			}

			// Check the response body for successful requests
			if tt.expectedStatus == http.StatusOK {
				var actualResponse models.PolicyResponse
				err := json.Unmarshal(rr.Body.Bytes(), &actualResponse)
				if err != nil {
					t.Fatalf("Could not unmarshal response: %v\nRaw body: %s", err, rr.Body.String())
				}

				// Basic comparison of fields
				if actualResponse.ConsumerID != tt.expectedResponseBody.ConsumerID {
					t.Errorf("ConsumerID mismatch.\nExpected: %s\nGot:      %s", tt.expectedResponseBody.ConsumerID, actualResponse.ConsumerID)
				}
				if actualResponse.OverallConsentRequired != tt.expectedResponseBody.OverallConsentRequired {
					t.Errorf("OverallConsentRequired mismatch.\nExpected: %t\nGot:      %t", tt.expectedResponseBody.OverallConsentRequired, actualResponse.OverallConsentRequired)
				}

				// Compare AccessScopes (order might matter, so compare elements individually)
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
				// For error cases, we might check error messages if the handler returns them
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

		req := httptest.NewRequest(http.MethodGet, "/evaluate-policy", nil) // Use GET method
		rr := httptest.NewRecorder()
		handler := HandlePolicyRequest(service)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusMethodNotAllowed {
			t.Errorf("Handler returned wrong status code:\nGot:  %v\nWant: %v",
				status, http.StatusMethodNotAllowed)
		}
	})
}
