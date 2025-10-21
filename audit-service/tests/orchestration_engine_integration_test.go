package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gov-dx-sandbox/audit-service/models"
)

// TestOrchestrationEngineIntegration tests the integration between orchestration engine and audit service
func TestOrchestrationEngineIntegration(t *testing.T) {
	server := SetupTestServer(t)
	defer server.Close()

	t.Run("POST_Logs_From_Orchestration_Engine", func(t *testing.T) {
		// Test data that would come from orchestration engine
		testCases := []struct {
			name           string
			request        models.LogRequest
			expectedStatus int
			description    string
		}{
			{
				name: "Success_Response_From_Subgraph",
				request: models.LogRequest{
					Status:        "success",
					RequestedData: `query { user { id name email } }`,
					ApplicationID: "app-123",
					SchemaID:      "schema-456",
					ConsumerID:    "consumer-123",
					ProviderID:    "provider-456",
				},
				expectedStatus: http.StatusOK,
				description:    "Should successfully log a successful subgraph query response",
			},
			{
				name: "Failure_Response_From_Subgraph",
				request: models.LogRequest{
					Status:        "failure",
					RequestedData: `query { invalidField { id } }`,
					ApplicationID: "app-123",
					SchemaID:      "schema-456",
					ConsumerID:    "consumer-123",
					ProviderID:    "provider-456",
				},
				expectedStatus: http.StatusOK,
				description:    "Should successfully log a failed subgraph query response",
			},
			{
				name: "Complex_GraphQL_Query",
				request: models.LogRequest{
					Status:        "success",
					RequestedData: `query GetUserData($userId: ID!) { user(id: $userId) { id name email posts { id title content } } }`,
					ApplicationID: "app-123",
					SchemaID:      "schema-456",
					ConsumerID:    "consumer-123",
					ProviderID:    "provider-456",
				},
				expectedStatus: http.StatusOK,
				description:    "Should handle complex GraphQL queries with variables",
			},
			{
				name: "Mutation_Query",
				request: models.LogRequest{
					Status:        "success",
					RequestedData: `mutation CreateUser($input: UserInput!) { createUser(input: $input) { id name email } }`,
					ApplicationID: "app-123",
					SchemaID:      "schema-456",
					ConsumerID:    "consumer-123",
					ProviderID:    "provider-456",
				},
				expectedStatus: http.StatusOK,
				description:    "Should handle GraphQL mutation queries",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Create request body
				reqBody, err := json.Marshal(tc.request)
				if err != nil {
					t.Fatalf("Failed to marshal request: %v", err)
				}

				// Create HTTP request
				req := httptest.NewRequest("POST", "/api/logs", bytes.NewBuffer(reqBody))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				// Call the handler
				server.Handler.CreateLog(w, req)

				// Verify response status
				if w.Code != tc.expectedStatus {
					t.Errorf("Expected status %d, got %d. %s", tc.expectedStatus, w.Code, tc.description)
				}

				// Verify response body for successful requests
				if tc.expectedStatus == http.StatusOK {
					var response models.Log
					if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
						t.Errorf("Failed to unmarshal response: %v", err)
					}

					// Verify response fields
					if response.Status != tc.request.Status {
						t.Errorf("Expected status %s, got %s", tc.request.Status, response.Status)
					}
					if response.RequestedData != tc.request.RequestedData {
						t.Errorf("Expected requested data %s, got %s", tc.request.RequestedData, response.RequestedData)
					}
					if response.ApplicationID != tc.request.ApplicationID {
						t.Errorf("Expected application ID %s, got %s", tc.request.ApplicationID, response.ApplicationID)
					}
					if response.SchemaID != tc.request.SchemaID {
						t.Errorf("Expected schema ID %s, got %s", tc.request.SchemaID, response.SchemaID)
					}
					if response.ConsumerID != tc.request.ConsumerID {
						t.Errorf("Expected consumer ID %s, got %s", tc.request.ConsumerID, response.ConsumerID)
					}
					if response.ProviderID != tc.request.ProviderID {
						t.Errorf("Expected provider ID %s, got %s", tc.request.ProviderID, response.ProviderID)
					}
					if response.ID == "" {
						t.Error("Expected non-empty ID")
					}
					if response.Timestamp.IsZero() {
						t.Error("Expected non-zero timestamp")
					}
				}
			})
		}
	})

	t.Run("GET_Logs_Using_View", func(t *testing.T) {
		// First, create some test data
		testLogs := []models.LogRequest{
			{
				Status:        "success",
				RequestedData: `query { user { id name } }`,
				ApplicationID: "app-123",
				SchemaID:      "schema-456",
				ConsumerID:    "consumer-123",
				ProviderID:    "provider-456",
			},
			{
				Status:        "failure",
				RequestedData: `query { invalidField { id } }`,
				ApplicationID: "app-123",
				SchemaID:      "schema-456",
				ConsumerID:    "consumer-123",
				ProviderID:    "provider-456",
			},
			{
				Status:        "success",
				RequestedData: `query { posts { id title } }`,
				ApplicationID: "app-789",
				SchemaID:      "schema-789",
				ConsumerID:    "consumer-789",
				ProviderID:    "provider-789",
			},
		}

		// Create test logs
		for _, logReq := range testLogs {
			reqBody, _ := json.Marshal(logReq)
			req := httptest.NewRequest("POST", "/api/logs", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			server.Handler.CreateLog(w, req)
		}

		// Test cases for GET requests using the view
		testCases := []struct {
			name          string
			queryParams   string
			expectedCount int
			description   string
		}{
			{
				name:          "Get_All_Logs",
				queryParams:   "",
				expectedCount: 3,
				description:   "Should return all logs using the view",
			},
			{
				name:          "Filter_By_Consumer_ID",
				queryParams:   "?consumerId=consumer-123",
				expectedCount: 2,
				description:   "Should filter logs by consumer ID using the view",
			},
			{
				name:          "Filter_By_Provider_ID",
				queryParams:   "?providerId=provider-456",
				expectedCount: 2,
				description:   "Should filter logs by provider ID using the view",
			},
			{
				name:          "Filter_By_Status",
				queryParams:   "?status=success",
				expectedCount: 2,
				description:   "Should filter logs by status using the view",
			},
			{
				name:          "Filter_By_Status_Failure",
				queryParams:   "?status=failure",
				expectedCount: 1,
				description:   "Should filter logs by failure status using the view",
			},
			{
				name:          "Combined_Filters",
				queryParams:   "?consumerId=consumer-123&status=success",
				expectedCount: 1,
				description:   "Should apply multiple filters using the view",
			},
			{
				name:          "Pagination",
				queryParams:   "?limit=2&offset=0",
				expectedCount: 2,
				description:   "Should support pagination using the view",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Create GET request
				req := httptest.NewRequest("GET", "/api/logs"+tc.queryParams, nil)
				w := httptest.NewRecorder()

				// Call the handler
				server.Handler.GetLogs(w, req)

				// Verify response status
				if w.Code != http.StatusOK {
					t.Errorf("Expected status %d, got %d. %s", http.StatusOK, w.Code, tc.description)
				}

				// Verify response body
				var response models.LogResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}

				// Verify count
				if len(response.Logs) != tc.expectedCount {
					t.Errorf("Expected %d logs, got %d. %s", tc.expectedCount, len(response.Logs), tc.description)
				}

				// Verify that all returned logs have the expected fields from the view
				for _, log := range response.Logs {
					if log.ConsumerID == "" {
						t.Error("Expected non-empty consumer ID from view")
					}
					if log.ProviderID == "" {
						t.Error("Expected non-empty provider ID from view")
					}
					if log.Status == "" {
						t.Error("Expected non-empty status from view")
					}
					if log.RequestedData == "" {
						t.Error("Expected non-empty requested data from view")
					}
				}
			})
		}
	})

	t.Run("Error_Handling", func(t *testing.T) {
		// Test invalid request body
		t.Run("Invalid_JSON", func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/logs", bytes.NewBufferString("invalid json"))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.Handler.CreateLog(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
		})

		// Test missing required fields
		t.Run("Missing_Required_Fields", func(t *testing.T) {
			invalidRequest := models.LogRequest{
				Status: "success",
				// Missing other required fields
			}

			reqBody, _ := json.Marshal(invalidRequest)
			req := httptest.NewRequest("POST", "/api/logs", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.Handler.CreateLog(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
		})

		// Test invalid status
		t.Run("Invalid_Status", func(t *testing.T) {
			invalidRequest := models.LogRequest{
				Status:        "invalid",
				RequestedData: "query { user { id } }",
				ApplicationID: "app-123",
				SchemaID:      "schema-456",
				ConsumerID:    "consumer-123",
				ProviderID:    "provider-456",
			}

			reqBody, _ := json.Marshal(invalidRequest)
			req := httptest.NewRequest("POST", "/api/logs", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.Handler.CreateLog(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	})

	t.Run("Concurrent_Requests", func(t *testing.T) {
		// Test concurrent POST requests to simulate multiple orchestration engine instances
		concurrency := 10
		done := make(chan bool, concurrency)

		for i := 0; i < concurrency; i++ {
			go func(index int) {
				defer func() { done <- true }()

				request := models.LogRequest{
					Status:        "success",
					RequestedData: fmt.Sprintf("query { user%d { id name } }", index),
					ApplicationID: "app-123",
					SchemaID:      "schema-456",
					ConsumerID:    "consumer-123",
					ProviderID:    "provider-456",
				}

				reqBody, _ := json.Marshal(request)
				req := httptest.NewRequest("POST", "/api/logs", bytes.NewBuffer(reqBody))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				server.Handler.CreateLog(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("Concurrent request %d failed with status %d", index, w.Code)
				}
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < concurrency; i++ {
			<-done
		}

		// Verify all logs were created
		req := httptest.NewRequest("GET", "/api/logs", nil)
		w := httptest.NewRecorder()
		server.Handler.GetLogs(w, req)

		var response models.LogResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		if len(response.Logs) < concurrency {
			t.Errorf("Expected at least %d logs, got %d", concurrency, len(response.Logs))
		}
	})
}

// TestOrchestrationEngineAuditClient tests the audit client functionality
func TestOrchestrationEngineAuditClient(t *testing.T) {
	server := SetupTestServer(t)
	defer server.Close()

	// Create a test server to simulate the audit service
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/logs" && r.Method == "POST" {
			// Simulate the audit service handler
			var req models.LogRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			// Validate required fields
			if req.Status == "" || req.RequestedData == "" || req.ApplicationID == "" || req.SchemaID == "" {
				http.Error(w, "Missing required fields", http.StatusBadRequest)
				return
			}

			// Simulate successful response
			response := models.Log{
				ID:            "test-id-123",
				Timestamp:     time.Now(),
				Status:        req.Status,
				RequestedData: req.RequestedData,
				ApplicationID: req.ApplicationID,
				SchemaID:      req.SchemaID,
				ConsumerID:    req.ConsumerID,
				ProviderID:    req.ProviderID,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer testServer.Close()

	t.Run("Audit_Client_Success", func(t *testing.T) {
		// This would be the actual audit client test
		// For now, we'll test the HTTP endpoint directly
		request := models.LogRequest{
			Status:        "success",
			RequestedData: "query { user { id name } }",
			ApplicationID: "app-123",
			SchemaID:      "schema-456",
			ConsumerID:    "consumer-123",
			ProviderID:    "provider-456",
		}

		reqBody, _ := json.Marshal(request)
		req := httptest.NewRequest("POST", "/api/logs", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.Handler.CreateLog(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response models.Log
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response.Status != request.Status {
			t.Errorf("Expected status %s, got %s", request.Status, response.Status)
		}
	})
}

// TestDatabaseViewIntegration tests the database view functionality
func TestDatabaseViewIntegration(t *testing.T) {
	server := SetupTestServer(t)
	defer server.Close()

	// This test would require the actual database view to be created
	// For now, we'll test the basic functionality
	t.Run("View_Data_Retrieval", func(t *testing.T) {
		// Create test data
		testLogs := []models.LogRequest{
			{
				Status:        "success",
				RequestedData: "query { user { id name } }",
				ApplicationID: "app-123",
				SchemaID:      "schema-456",
				ConsumerID:    "consumer-123",
				ProviderID:    "provider-456",
			},
		}

		// Create logs
		for _, logReq := range testLogs {
			reqBody, _ := json.Marshal(logReq)
			req := httptest.NewRequest("POST", "/api/logs", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			server.Handler.CreateLog(w, req)
		}

		// Test GET request (which should use the view)
		req := httptest.NewRequest("GET", "/api/logs", nil)
		w := httptest.NewRecorder()
		server.Handler.GetLogs(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response models.LogResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if len(response.Logs) == 0 {
			t.Error("Expected at least one log from the view")
		}

		// Verify that the view provides the expected fields
		for _, log := range response.Logs {
			if log.ConsumerID == "" {
				t.Error("Expected consumer ID from view")
			}
			if log.ProviderID == "" {
				t.Error("Expected provider ID from view")
			}
		}
	})
}
