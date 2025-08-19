// handler_test.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"policy-governance/internal/models" // Import your models package

	"github.com/stretchr/testify/assert" // For assertion library
)

// MockPolicyFetcher implements PolicyDataFetcher for testing purposes.
// It allows us to define expected policy records without hitting a real database.
type MockPolicyFetcher struct {
	Policies map[string]models.PolicyRecord // Key: "consumerID-subgraph-type-field"
	Err      error                          // Simulate a database error
}

// GetPolicyFromDB implements the PolicyDataFetcher interface for the mock.
func (m *MockPolicyFetcher) GetPolicyFromDB(consumerID, subgraph, typ, field string) (models.PolicyRecord, error) {
	if m.Err != nil {
		return models.PolicyRecord{}, m.Err // Return a simulated error
	}
	key := fmt.Sprintf("%s-%s-%s-%s", consumerID, subgraph, typ, field)
	if policy, ok := m.Policies[key]; ok {
		return policy, nil // Return the predefined mock policy
	}
	// If a specific policy is not found in the mock, default to DENY,
	// mimicking the behavior in DatabasePolicyFetcher when sql.ErrNoRows.
	return models.PolicyRecord{Classification: models.DENY}, nil
}

// TestHandlePolicyRequest tests the HTTP handler for policy requests.
func TestHandlePolicyRequest(t *testing.T) {
	// Define a slice of test cases
	tests := []struct {
		name           string                         // Name of the test case
		consumerID     string                         // Value for the x-consumer-id header
		requestMethod  string                         // HTTP method for the request
		requestBody    string                         // JSON body of the GraphQL request
		mockPolicies   map[string]models.PolicyRecord // Policies to be returned by the mock fetcher
		mockErr        error                          // Error to be returned by the mock fetcher (if any)
		expectedStatus int                            // Expected HTTP status code in the response
		expectedResp   models.PolicyResponse          // Expected PolicyResponse JSON
	}{
		{
			name:          "Citizen_AllowedAndProviderConsentVehicle",
			consumerID:    "citizen",
			requestMethod: http.MethodPost,
			requestBody: `{
				"query": "query GetVehicleDetails { vehicle { vehicleInfoById(vehicleId: \"v-123\") { id make engineNumber } } }"
			}`,
			mockPolicies: map[string]models.PolicyRecord{
				// Policies for fields extracted by parseGraphQLQuery
				"citizen-dmt-Query-vehicle":            {Classification: models.ALLOW},
				"citizen-dmt-Vehicle-vehicleInfoById":  {Classification: models.ALLOW},
				"citizen-dmt-VehicleInfo-id":           {Classification: models.ALLOW},
				"citizen-dmt-VehicleInfo-make":         {Classification: models.ALLOW},
				"citizen-dmt-VehicleInfo-engineNumber": {Classification: models.ALLOW_PROVIDER_CONSENT},
			},
			expectedStatus: http.StatusOK,
			expectedResp: models.PolicyResponse{
				ConsumerID: "citizen",
				AccessScopes: []models.AccessScope{
					{SubgraphName: "dmt", TypeName: "Query", FieldName: "vehicle", ResolvedClassification: models.ALLOW},
					{SubgraphName: "dmt", TypeName: "Vehicle", FieldName: "vehicleInfoById", ResolvedClassification: models.ALLOW},
					{SubgraphName: "dmt", TypeName: "VehicleInfo", FieldName: "id", ResolvedClassification: models.ALLOW},
					{SubgraphName: "dmt", TypeName: "VehicleInfo", FieldName: "make", ResolvedClassification: models.ALLOW},
					{SubgraphName: "dmt", TypeName: "VehicleInfo", FieldName: "engineNumber", ResolvedClassification: models.ALLOW_PROVIDER_CONSENT},
				},
				OverallConsentRequired: true,
			},
		},
		{
			name:          "Bank_ConsentAndDeniedPerson",
			consumerID:    "bank",
			requestMethod: http.MethodPost,
			requestBody: `{
				"query": "query GetPersonDetails { person { getPersonByNic(nic: \"123456789V\") { photo email } } }"
			}`,
			mockPolicies: map[string]models.PolicyRecord{
				"bank-drp-Query-person":          {Classification: models.ALLOW},
				"bank-drp-person-getPersonByNic": {Classification: models.ALLOW},
				"bank-drp-PersonData-photo":      {Classification: models.ALLOW_CITIZEN_CONSENT},
				"bank-drp-PersonData-email":      {Classification: models.DENY}, // Explicitly denied
			},
			expectedStatus: http.StatusOK,
			expectedResp: models.PolicyResponse{
				ConsumerID: "bank",
				AccessScopes: []models.AccessScope{
					{SubgraphName: "drp", TypeName: "Query", FieldName: "person", ResolvedClassification: models.ALLOW},
					{SubgraphName: "drp", TypeName: "person", FieldName: "getPersonByNic", ResolvedClassification: models.ALLOW},
					{SubgraphName: "drp", TypeName: "PersonData", FieldName: "photo", ResolvedClassification: models.ALLOW_CITIZEN_CONSENT},
					{SubgraphName: "drp", TypeName: "PersonData", FieldName: "email", ResolvedClassification: models.DENY},
				},
				OverallConsentRequired: true,
			},
		},
		{
			name:          "InvalidGraphQLQuery",
			consumerID:    "citizen",
			requestMethod: http.MethodPost,
			requestBody: `{
				"query": "invalid { query }"
			}`,
			mockPolicies:   nil, // No policies needed for this test
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "NonPOSTRequest",
			consumerID:     "citizen",
			requestMethod:  http.MethodGet, // Use GET method
			requestBody:    "",             // Body not relevant for GET
			mockPolicies:   nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:          "MissingConsumerID",
			consumerID:    "", // Simulate missing header
			requestMethod: http.MethodPost,
			requestBody: `{
				"query": "query GetVehicleDetails { vehicle { vehicleInfoById(vehicleId: \"v-123\") { id } } }"
			}`,
			mockPolicies: map[string]models.PolicyRecord{
				"anonymous-consumer-dmt-Query-vehicle":           {Classification: models.ALLOW},
				"anonymous-consumer-dmt-Vehicle-vehicleInfoById": {Classification: models.ALLOW},
				"anonymous-consumer-dmt-VehicleInfo-id":          {Classification: models.ALLOW},
			},
			expectedStatus: http.StatusOK,
			expectedResp: models.PolicyResponse{
				ConsumerID: "anonymous-consumer", // Should default to this
				AccessScopes: []models.AccessScope{
					{SubgraphName: "dmt", TypeName: "Query", FieldName: "vehicle", ResolvedClassification: models.ALLOW},
					{SubgraphName: "dmt", TypeName: "Vehicle", FieldName: "vehicleInfoById", ResolvedClassification: models.ALLOW},
					{SubgraphName: "dmt", TypeName: "VehicleInfo", FieldName: "id", ResolvedClassification: models.ALLOW},
				},
				OverallConsentRequired: false,
			},
		},
		{
			name:          "DBErrorDuringPolicyFetch",
			consumerID:    "citizen",
			requestMethod: http.MethodPost,
			requestBody: `{
				"query": "query GetVehicleDetails { vehicle { vehicleInfoById(vehicleId: \"v-123\") { id } } }"
			}`,
			mockPolicies:   nil,                                    // No specific policies for this case
			mockErr:        fmt.Errorf("simulated database error"), // Simulate a DB error
			expectedStatus: http.StatusOK,                          // Still OK, but fields default to DENY
			expectedResp: models.PolicyResponse{
				ConsumerID: "citizen",
				AccessScopes: []models.AccessScope{
					{SubgraphName: "dmt", TypeName: "Query", FieldName: "vehicle", ResolvedClassification: models.DENY},
					{SubgraphName: "dmt", TypeName: "Vehicle", FieldName: "vehicleInfoById", ResolvedClassification: models.DENY},
					{SubgraphName: "dmt", TypeName: "VehicleInfo", FieldName: "id", ResolvedClassification: models.DENY},
				},
				OverallConsentRequired: false, // No consent required if all denied by default
			},
		},
	}

	// Iterate over each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize the mock fetcher and service for the current test case
			mockFetcher := &MockPolicyFetcher{
				Policies: tt.mockPolicies,
				Err:      tt.mockErr,
			}
			service := &PolicyGovernanceService{Fetcher: mockFetcher}
			handler := HandlePolicyRequest(service) // Get the HTTP handler function

			// Create a new HTTP request
			var req *http.Request
			if tt.requestBody != "" {
				req = httptest.NewRequest(tt.requestMethod, "/evaluate-policy", bytes.NewBufferString(tt.requestBody))
			} else {
				req = httptest.NewRequest(tt.requestMethod, "/evaluate-policy", nil)
			}

			// Set headers
			if tt.consumerID != "" {
				req.Header.Set("x-consumer-id", tt.consumerID)
			}
			req.Header.Set("Content-Type", "application/json") // Always set Content-Type for POST requests

			// Record the HTTP response
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// Assert HTTP status code
			assert.Equal(t, tt.expectedStatus, rr.Code, "handler returned wrong status code")

			// If the status is OK, decode and assert the response body
			if tt.expectedStatus == http.StatusOK {
				var actualResp models.PolicyResponse
				err := json.NewDecoder(rr.Body).Decode(&actualResp)
				assert.NoError(t, err, "failed to decode response body")

				// Assert the consumer ID
				assert.Equal(t, tt.expectedResp.ConsumerID, actualResp.ConsumerID, "consumerID mismatch")
				// Assert the overall consent required flag
				assert.Equal(t, tt.expectedResp.OverallConsentRequired, actualResp.OverallConsentRequired, "OverallConsentRequired mismatch")
				// Assert AccessScopes (order-independent comparison)
				assertAccessScopes(t, tt.expectedResp.AccessScopes, actualResp.AccessScopes)
			}
		})
	}
}

// assertAccessScopes is a helper function to assert that two slices of AccessScope are equal,
// ignoring the order of elements. This is necessary because map iteration order is not guaranteed.
func assertAccessScopes(t *testing.T, expected, actual []models.AccessScope) {
	assert.Len(t, actual, len(expected), "AccessScopes length mismatch")

	// Create maps for easier comparison
	expectedMap := make(map[string]models.AccessScope)
	for _, scope := range expected {
		key := fmt.Sprintf("%s.%s.%s", scope.SubgraphName, scope.TypeName, scope.FieldName)
		expectedMap[key] = scope
	}

	for _, scope := range actual {
		key := fmt.Sprintf("%s.%s.%s", scope.SubgraphName, scope.TypeName, scope.FieldName)
		expectedScope, ok := expectedMap[key]
		assert.True(t, ok, "Unexpected AccessScope found: %v", scope)
		if ok {
			assert.Equal(t, expectedScope.ResolvedClassification, scope.ResolvedClassification, "Classification mismatch for %s", key)
			// Add more assertions here if other fields in AccessScope need to be compared
		}
	}
}
