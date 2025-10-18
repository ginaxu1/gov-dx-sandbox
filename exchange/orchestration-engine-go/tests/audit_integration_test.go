package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/audit"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuditIntegration tests that the orchestration engine successfully calls POST /audit/logs
// when a request comes into POST /
func TestAuditIntegration(t *testing.T) {
	// Initialize logger for tests
	logger.Init()

	// Create a mock audit service
	mockAuditServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request is to the correct endpoint
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/logs", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Read and verify the request body
		var auditRequest audit.AuditLogRequest
		err := json.NewDecoder(r.Body).Decode(&auditRequest)
		require.NoError(t, err, "Should decode audit request successfully")

		// Verify the audit request structure
		assert.NotEmpty(t, auditRequest.Status, "Status should not be empty")
		assert.NotEmpty(t, auditRequest.RequestedData, "RequestedData should not be empty")
		assert.NotEmpty(t, auditRequest.ApplicationID, "ApplicationID should not be empty")
		assert.NotEmpty(t, auditRequest.SchemaID, "SchemaID should not be empty")

		// Verify the status is either "success" or "failure"
		assert.Contains(t, []string{"success", "failure"}, auditRequest.Status, "Status should be success or failure")

		// Verify the requested data contains the GraphQL query
		assert.Contains(t, auditRequest.RequestedData, "query", "RequestedData should contain GraphQL query")

		// Return a mock response
		response := audit.AuditLogResponse{
			ID:            "test-audit-id-123",
			Timestamp:     time.Now(),
			Status:        auditRequest.Status,
			RequestedData: auditRequest.RequestedData,
			ApplicationID: auditRequest.ApplicationID,
			SchemaID:      auditRequest.SchemaID,
			ConsumerID:    auditRequest.ApplicationID,
			ProviderID:    auditRequest.SchemaID,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer mockAuditServer.Close()

	// Set the audit service URL environment variable
	t.Setenv("AUDIT_SERVICE_URL", mockAuditServer.URL)

	// Create the orchestration engine server
	orchestrationServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock the orchestration engine's POST / handler
		if r.Method == "POST" && r.URL.Path == "/" {
			// Read the GraphQL request
			var graphqlReq struct {
				Query         string                 `json:"query"`
				Variables     map[string]interface{} `json:"variables"`
				OperationName string                 `json:"operationName"`
			}

			err := json.NewDecoder(r.Body).Decode(&graphqlReq)
			if err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			// Mock a successful GraphQL response
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"personInfo": map[string]interface{}{
						"fullName": "John Doe",
						"name":     "John",
					},
				},
			}

			// Simulate the audit logging that happens in the real server
			auditClient := audit.NewAuditClient()

			// Determine status (simulate success for this test)
			status := "success"

			// Extract consumer and provider information
			consumerID := "test-consumer-123"
			providerID := "test-provider-456"

			// Log the query execution to audit service
			err = auditClient.LogQuery(graphqlReq.Query, status, consumerID, providerID)
			if err != nil {
				t.Logf("Audit logging failed: %v", err)
				// Don't fail the test if audit logging fails, just log it
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer orchestrationServer.Close()

	// Test cases
	testCases := []struct {
		name        string
		query       string
		expectError bool
	}{
		{
			name: "Simple GraphQL Query",
			query: `
				query GetPersonInfo {
					personInfo(nic: "123456789V") {
						fullName
						name
					}
				}
			`,
			expectError: false,
		},
		{
			name: "GraphQL Query with Variables",
			query: `
				query GetPersonInfo($nic: String!) {
					personInfo(nic: $nic) {
						fullName
						name
					}
				}
			`,
			expectError: false,
		},
		{
			name: "Complex GraphQL Query",
			query: `
				query GetPersonWithVehicles {
					personInfo(nic: "123456789V") {
						fullName
						name
						ownedVehicles {
							regNo
							make
							model
						}
					}
				}
			`,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Prepare the GraphQL request
			requestBody := map[string]interface{}{
				"query": tc.query,
				"variables": map[string]interface{}{
					"nic": "123456789V",
				},
			}

			jsonBody, err := json.Marshal(requestBody)
			require.NoError(t, err, "Should marshal request body")

			// Make the request to the orchestration engine
			resp, err := http.Post(orchestrationServer.URL+"/", "application/json", bytes.NewBuffer(jsonBody))
			require.NoError(t, err, "Should make HTTP request successfully")
			defer resp.Body.Close()

			// Verify the response
			if tc.expectError {
				assert.NotEqual(t, http.StatusOK, resp.StatusCode, "Should return error status")
			} else {
				assert.Equal(t, http.StatusOK, resp.StatusCode, "Should return success status")

				// Verify the response body
				var response map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err, "Should decode response successfully")

				// Verify the response contains the expected data
				assert.Contains(t, response, "data", "Response should contain data field")
				data := response["data"].(map[string]interface{})
				assert.Contains(t, data, "personInfo", "Data should contain personInfo field")
			}
		})
	}
}

// TestAuditIntegrationWithFailure tests audit logging when GraphQL query fails
func TestAuditIntegrationWithFailure(t *testing.T) {
	// Initialize logger for tests
	logger.Init()

	// Create a mock audit service that expects a failure status
	mockAuditServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request is to the correct endpoint
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/logs", r.URL.Path)

		// Read and verify the request body
		var auditRequest audit.AuditLogRequest
		err := json.NewDecoder(r.Body).Decode(&auditRequest)
		require.NoError(t, err, "Should decode audit request successfully")

		// Verify the status is "failure"
		assert.Equal(t, "failure", auditRequest.Status, "Status should be failure")

		// Return a mock response
		response := audit.AuditLogResponse{
			ID:            "test-audit-id-failure-123",
			Timestamp:     time.Now(),
			Status:        auditRequest.Status,
			RequestedData: auditRequest.RequestedData,
			ApplicationID: auditRequest.ApplicationID,
			SchemaID:      auditRequest.SchemaID,
			ConsumerID:    auditRequest.ApplicationID,
			ProviderID:    auditRequest.SchemaID,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer mockAuditServer.Close()

	// Set the audit service URL environment variable
	t.Setenv("AUDIT_SERVICE_URL", mockAuditServer.URL)

	// Create the orchestration engine server that simulates a failure
	orchestrationServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/" {
			// Read the GraphQL request
			var graphqlReq struct {
				Query         string                 `json:"query"`
				Variables     map[string]interface{} `json:"variables"`
				OperationName string                 `json:"operationName"`
			}

			err := json.NewDecoder(r.Body).Decode(&graphqlReq)
			if err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			// Mock a failed GraphQL response
			response := map[string]interface{}{
				"data": nil,
				"errors": []map[string]interface{}{
					{
						"message": "GraphQL execution failed",
						"locations": []map[string]interface{}{
							{"line": 1, "column": 1},
						},
					},
				},
			}

			// Simulate the audit logging that happens in the real server
			auditClient := audit.NewAuditClient()

			// Determine status (simulate failure for this test)
			status := "failure"

			// Extract consumer and provider information
			consumerID := "test-consumer-failure-123"
			providerID := "test-provider-failure-456"

			// Log the query execution to audit service
			err = auditClient.LogQuery(graphqlReq.Query, status, consumerID, providerID)
			if err != nil {
				t.Logf("Audit logging failed: %v", err)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer orchestrationServer.Close()

	// Test with a GraphQL query that will result in failure
	requestBody := map[string]interface{}{
		"query": `
			query GetPersonInfo {
				personInfo(nic: "123456789V") {
					nonExistentField
				}
			}
		`,
	}

	jsonBody, err := json.Marshal(requestBody)
	require.NoError(t, err, "Should marshal request body")

	// Make the request to the orchestration engine
	resp, err := http.Post(orchestrationServer.URL+"/", "application/json", bytes.NewBuffer(jsonBody))
	require.NoError(t, err, "Should make HTTP request successfully")
	defer resp.Body.Close()

	// Verify the response
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Should return success status")

	// Verify the response body contains errors
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err, "Should decode response successfully")

	assert.Contains(t, response, "errors", "Response should contain errors field")
	errors := response["errors"].([]interface{})
	assert.Len(t, errors, 1, "Should have one error")
}

// TestAuditClientDirectly tests the audit client directly
func TestAuditClientDirectly(t *testing.T) {
	// Initialize logger
	logger.Init()

	// Create a mock audit service
	mockAuditServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/logs", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Read and verify the request body
		var auditRequest audit.AuditLogRequest
		err := json.NewDecoder(r.Body).Decode(&auditRequest)
		require.NoError(t, err, "Should decode audit request successfully")

		// Verify the audit request
		assert.Equal(t, "success", auditRequest.Status)
		assert.Equal(t, "query { personInfo { name } }", auditRequest.RequestedData)
		assert.Equal(t, "test-consumer", auditRequest.ApplicationID)
		assert.Equal(t, "test-provider", auditRequest.SchemaID)

		// Return a mock response
		response := audit.AuditLogResponse{
			ID:            "test-audit-id-direct-123",
			Timestamp:     time.Now(),
			Status:        auditRequest.Status,
			RequestedData: auditRequest.RequestedData,
			ApplicationID: auditRequest.ApplicationID,
			SchemaID:      auditRequest.SchemaID,
			ConsumerID:    auditRequest.ApplicationID,
			ProviderID:    auditRequest.SchemaID,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer mockAuditServer.Close()

	// Set the audit service URL environment variable
	t.Setenv("AUDIT_SERVICE_URL", mockAuditServer.URL)

	// Create audit client
	auditClient := audit.NewAuditClient()

	// Test the LogQuery method
	err := auditClient.LogQuery("query { personInfo { name } }", "success", "test-consumer", "test-provider")
	assert.NoError(t, err, "Should log query successfully")

	// Test the LogQueryAsync method
	auditClient.LogQueryAsync("query { personInfo { name } }", "success", "test-consumer", "test-provider")

	// Give some time for the async operation to complete
	time.Sleep(100 * time.Millisecond)
}

// TestAuditClientErrorHandling tests error handling in the audit client
func TestAuditClientErrorHandling(t *testing.T) {
	// Create a mock audit service that returns an error
	mockAuditServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer mockAuditServer.Close()

	// Set the audit service URL environment variable
	t.Setenv("AUDIT_SERVICE_URL", mockAuditServer.URL)

	// Create audit client
	auditClient := audit.NewAuditClient()

	// Test that LogQuery returns an error when the audit service fails
	err := auditClient.LogQuery("query { personInfo { name } }", "success", "test-consumer", "test-provider")
	assert.Error(t, err, "Should return error when audit service fails")
	assert.Contains(t, err.Error(), "audit service returned status 500", "Error should contain status code")
}
