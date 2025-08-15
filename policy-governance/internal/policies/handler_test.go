package policies_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"policy-governance/internal/database"
	"policy-governance/internal/models"
	"policy-governance/internal/policies"
)

var mockGetConsumerPolicy func(consumerID string) (*models.Policy, error)

func TestMain(m *testing.M) {
	originalGetConsumerPolicy := database.GetConsumerPolicy
	defer func() {
		database.GetConsumerPolicy = originalGetConsumerPolicy
	}()

	database.GetConsumerPolicy = func(consumerID string) (*models.Policy, error) {
		return mockGetConsumerPolicy(consumerID)
	}

	m.Run()
}

func TestPolicyHandler(t *testing.T) {
	tests := []struct {
		name               string
		consumerIDHeader   string
		requestBody        interface{}
		mockPolicy         *models.Policy
		mockPolicyErr      error
		expectedAuth       bool
		expectedMessage    string
		expectedHTTPStatus int
	}{
		{
			name:             "Missing Consumer ID Header",
			consumerIDHeader: "",
			requestBody: models.RequestBody{
				ConsumerID: "test-consumer",
				RequestedFields: []models.RequestField{
					{SubgraphName: "dmt", TypeName: "VehicleInfo", FieldName: "make", Classification: "ALLOW"},
				},
			},
			mockPolicy:         nil,
			mockPolicyErr:      nil,
			expectedAuth:       false,
			expectedMessage:    "X-Consumer-Id header is missing",
			expectedHTTPStatus: http.StatusOK,
		},
		{
			name:               "Invalid Request Body",
			consumerIDHeader:   "test-consumer",
			requestBody:        "this is not valid json",
			mockPolicy:         nil,
			mockPolicyErr:      nil,
			expectedAuth:       false,
			expectedMessage:    "Invalid request body: invalid character 'h' in literal true (expecting 'r')",
			expectedHTTPStatus: http.StatusOK,
		},
		{
			name:             "Consumer Not Found",
			consumerIDHeader: "non-existent-consumer",
			requestBody: models.RequestBody{
				ConsumerID: "non-existent-consumer",
				RequestedFields: []models.RequestField{
					{SubgraphName: "dmt", TypeName: "VehicleInfo", FieldName: "make", Classification: "ALLOW"},
				},
			},
			mockPolicy:         nil,
			mockPolicyErr:      errors.New("no policy found"),
			expectedAuth:       false,
			expectedMessage:    "Failed to retrieve policy for consumer: non-existent-consumer",
			expectedHTTPStatus: http.StatusOK,
		},
		{
			name:             "Authorized Request - Single Field",
			consumerIDHeader: "consumer1",
			requestBody: models.RequestBody{
				ConsumerID: "consumer1",
				RequestedFields: []models.RequestField{
					{SubgraphName: "dmt", TypeName: "VehicleInfo", FieldName: "make", Classification: "ALLOW"},
				},
			},
			mockPolicy: &models.Policy{
				Subgraphs: map[string]map[string][]string{
					"dmt": {"VehicleInfo": {"make", "model"}},
				},
			},
			mockPolicyErr:      nil,
			expectedAuth:       true,
			expectedMessage:    "Request is authorized",
			expectedHTTPStatus: http.StatusOK,
		},
		{
			name:             "Authorized Request - Multiple Fields",
			consumerIDHeader: "consumer2",
			requestBody: models.RequestBody{
				ConsumerID: "consumer2",
				RequestedFields: []models.RequestField{
					{SubgraphName: "drp", TypeName: "PersonData", FieldName: "fullName", Classification: "ALLOW_CITIZEN_CONSENT"},
					{SubgraphName: "drp", TypeName: "PersonData", FieldName: "email", Classification: "ALLOW_CITIZEN_CONSENT"},
					{SubgraphName: "dmt", TypeName: "VehicleInfo", FieldName: "model", Classification: "ALLOW"},
				},
			},
			mockPolicy: &models.Policy{
				Subgraphs: map[string]map[string][]string{
					"drp": {"PersonData": {"fullName", "email"}},
					"dmt": {"VehicleInfo": {"make", "model"}},
				},
			},
			mockPolicyErr:      nil,
			expectedAuth:       true,
			expectedMessage:    "Request is authorized",
			expectedHTTPStatus: http.StatusOK,
		},
		{
			name:             "Unauthorized Subgraph",
			consumerIDHeader: "consumer3",
			requestBody: models.RequestBody{
				ConsumerID: "consumer3",
				RequestedFields: []models.RequestField{
					{SubgraphName: "nonexistent", TypeName: "SomeType", FieldName: "someField", Classification: "ALLOW"},
				},
			},
			mockPolicy: &models.Policy{
				Subgraphs: map[string]map[string][]string{
					"dmt": {"VehicleInfo": {"make"}},
				},
			},
			mockPolicyErr:      nil,
			expectedAuth:       false,
			expectedMessage:    "Unauthorized access: Subgraph 'nonexistent' is not allowed.",
			expectedHTTPStatus: http.StatusOK,
		},
		{
			name:             "Unauthorized Type",
			consumerIDHeader: "consumer4",
			requestBody: models.RequestBody{
				ConsumerID: "consumer4",
				RequestedFields: []models.RequestField{
					{SubgraphName: "dmt", TypeName: "NonExistentType", FieldName: "someField", Classification: "ALLOW"},
				},
			},
			mockPolicy: &models.Policy{
				Subgraphs: map[string]map[string][]string{
					"dmt": {"VehicleInfo": {"make"}},
				},
			},
			mockPolicyErr:      nil,
			expectedAuth:       false,
			expectedMessage:    "Unauthorized access: Type 'NonExistentType' in subgraph 'dmt' is not allowed.",
			expectedHTTPStatus: http.StatusOK,
		},
		{
			name:             "Unauthorized Field",
			consumerIDHeader: "consumer5",
			requestBody: models.RequestBody{
				ConsumerID: "consumer5",
				RequestedFields: []models.RequestField{
					{SubgraphName: "dmt", TypeName: "VehicleInfo", FieldName: "unauthorizedField", Classification: "ALLOW"},
				},
			},
			mockPolicy: &models.Policy{
				Subgraphs: map[string]map[string][]string{
					"dmt": {"VehicleInfo": {"make", "model"}},
				},
			},
			mockPolicyErr:      nil,
			expectedAuth:       false,
			expectedMessage:    "Unauthorized access: Field 'unauthorizedField' in type 'VehicleInfo' of subgraph 'dmt' is not allowed.",
			expectedHTTPStatus: http.StatusOK,
		},
		{
			name:             "Explicitly Denied Field",
			consumerIDHeader: "consumer6",
			requestBody: models.RequestBody{
				ConsumerID: "consumer6",
				RequestedFields: []models.RequestField{
					{SubgraphName: "dmt", TypeName: "VehicleInfo", FieldName: "make", Classification: "DENY"},
				},
			},
			mockPolicy: &models.Policy{
				Subgraphs: map[string]map[string][]string{
					"dmt": {"VehicleInfo": {"make"}},
				},
			},
			mockPolicyErr:      nil,
			expectedAuth:       false,
			expectedMessage:    "Access to field 'make' is explicitly denied.",
			expectedHTTPStatus: http.StatusOK,
		},
		{
			name:             "Consumer ID Mismatch Header vs Body",
			consumerIDHeader: "consumerA",
			requestBody: models.RequestBody{
				ConsumerID: "consumerB",
				RequestedFields: []models.RequestField{
					{SubgraphName: "dmt", TypeName: "VehicleInfo", FieldName: "make", Classification: "ALLOW"},
				},
			},
			mockPolicy:         nil,
			mockPolicyErr:      nil,
			expectedAuth:       false,
			expectedMessage:    "Consumer ID mismatch between header and body",
			expectedHTTPStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGetConsumerPolicy = func(consumerID string) (*models.Policy, error) {
				if tt.mockPolicyErr != nil {
					return nil, tt.mockPolicyErr
				}
				return tt.mockPolicy, nil
			}

			var reqBodyBytes []byte
			if s, ok := tt.requestBody.(string); ok {
				reqBodyBytes = []byte(s)
			} else {
				reqBodyBytes, _ = json.Marshal(tt.requestBody)
			}

			req, err := http.NewRequest("POST", "/policy", bytes.NewBuffer(reqBodyBytes))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			if tt.consumerIDHeader != "" {
				req.Header.Set("X-Consumer-Id", tt.consumerIDHeader)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(policies.PolicyHandler)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedHTTPStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedHTTPStatus)
			}

			var response models.ResponseBody
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("could not unmarshal response: %v", err)
			}

			if response.Authorized != tt.expectedAuth {
				t.Errorf("handler returned wrong authorization status: got %v want %v", response.Authorized, tt.expectedAuth)
			}
			if response.Message != tt.expectedMessage {
				t.Errorf("handler returned wrong message: got %q want %q", response.Message, tt.expectedMessage)
			}
		})
	}
}
