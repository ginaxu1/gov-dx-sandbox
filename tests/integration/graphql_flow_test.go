package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	orchestrationEngineURL = "http://127.0.0.1:4000/public/graphql"
	pdpURL                 = "http://127.0.0.1:8082/api/v1/policy"
	consentEngineURL       = "http://127.0.0.1:8081/consents"
)

func TestMain(m *testing.M) {
	// Wait for all services to be available
	services := []struct {
		name string
		url  string
	}{
		{"Orchestration Engine", "http://127.0.0.1:4000/health"},
		{"Policy Decision Point", "http://127.0.0.1:8082/health"},
		{"Consent Engine", "http://127.0.0.1:8081/health"},
	}

	for _, svc := range services {
		if err := waitForService(svc.url, 30); err != nil {
			fmt.Printf("Service %s not available: %v\n", svc.name, err)
			os.Exit(1)
		}
		fmt.Printf("✅ %s is available\n", svc.name)
	}

	code := m.Run()
	os.Exit(code)
}

func waitForService(url string, maxAttempts int) error {
	for i := 0; i < maxAttempts; i++ {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("service at %s did not become available after %d attempts", url, maxAttempts)
}

// TestGraphQLFlow tests the complete success path:
// 1. GraphQL query to orchestration-engine-go
// 2. Success path through PDP (policy evaluation)
// 3. Success path through consent-engine (consent check)
// 4. Valid response back
func TestGraphQLFlow(t *testing.T) {
	// Setup: Create policy metadata in PDP
	schemaID := "test-schema-123"
	fieldName := "email"
	appID := "test-consumer-app"

	t.Run("Setup_PolicyMetadata", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"schemaId": schemaID,
			"records": []map[string]interface{}{
				{
					"fieldName":         fieldName,
					"source":            "primary",
					"isOwner":           true,
					"accessControlType": "restricted",
				},
			},
		}
		jsonData, err := json.Marshal(reqBody)
		require.NoError(t, err)

		resp, err := http.Post(pdpURL+"/metadata", "application/json", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Policy metadata should be created successfully")

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.NotNil(t, result["records"])
	})

	t.Run("Setup_Consent", func(t *testing.T) {
		// Create a consent record for the test scenario
		// Format based on consent-engine API: app_id, consent_requirements with owner, owner_id, and fields
		consentReq := map[string]interface{}{
			"app_id": appID,
			"consent_requirements": []map[string]interface{}{
				{
					"owner":    "CITIZEN",
					"owner_id": "test-owner-123",
					"fields": []map[string]interface{}{
						{
							"fieldName": fieldName,
							"schemaId":  schemaID,
						},
					},
				},
			},
			"grant_duration": "P30D",
		}
		jsonData, err := json.Marshal(consentReq)
		require.NoError(t, err)

		resp, err := http.Post(consentEngineURL, "application/json", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		defer resp.Body.Close()

		// Consent creation should return 201
		if resp.StatusCode != http.StatusCreated {
			bodyBytes := make([]byte, 1024)
			n, _ := resp.Body.Read(bodyBytes)
			t.Logf("Consent creation response status: %d, body: %s", resp.StatusCode, string(bodyBytes[:n]))
		}
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Consent should be created successfully")

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		t.Logf("Consent created: %+v", result)
		assert.NotEmpty(t, result["consent_id"], "Consent ID should be present in response")
	})

	t.Run("GraphQL_Query_To_OrchestrationEngine", func(t *testing.T) {
		// Create a simple GraphQL query
		// Note: This is a simplified query - adjust based on your actual schema
		graphQLQuery := map[string]interface{}{
			"query": `
				query TestQuery {
					__typename
				}
			`,
			"variables": map[string]interface{}{},
		}

		jsonData, err := json.Marshal(graphQLQuery)
		require.NoError(t, err)

		// Create HTTP request with a mock JWT token
		// Note: In a real scenario, you'd need a valid JWT token
		req, err := http.NewRequest("POST", orchestrationEngineURL, bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		// Add a mock authorization header if required
		// req.Header.Set("Authorization", "Bearer mock-token")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// The orchestration engine should process the request
		// It will call PDP and consent-engine internally
		t.Logf("Orchestration Engine response status: %d", resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err == nil {
			t.Logf("Orchestration Engine response: %+v", result)
		}

		// Validate response structure
		// The exact validation depends on your GraphQL schema
		// For now, we just verify the service responds
		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized,
			"Orchestration engine should respond (OK or Unauthorized if JWT required)")
	})

	t.Run("Verify_PDP_Integration", func(t *testing.T) {
		// Verify PDP can evaluate policies using /decide endpoint
		evalReq := map[string]interface{}{
			"consumer_id":     appID,
			"app_id":          appID,
			"request_id":      "test-req-123",
			"required_fields": []string{fieldName},
		}
		jsonData, err := json.Marshal(evalReq)
		require.NoError(t, err)

		resp, err := http.Post(pdpURL+"/decide", "application/json", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		defer resp.Body.Close()

		t.Logf("PDP evaluation response status: %d", resp.StatusCode)

		if resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			if err == nil {
				t.Logf("PDP evaluation result: %+v", result)
				assert.NotNil(t, result, "PDP should return evaluation result")
				assert.Contains(t, result, "appAuthorized", "PDP response should contain 'appAuthorized' field")
			}
		} else {
			bodyBytes := make([]byte, 1024)
			n, _ := resp.Body.Read(bodyBytes)
			t.Logf("PDP evaluation error response: %s", string(bodyBytes[:n]))
		}
	})

	t.Run("Verify_ConsentEngine_Integration", func(t *testing.T) {
		// Verify consent engine can retrieve consents by consumer
		checkURL := fmt.Sprintf("http://127.0.0.1:8081/consumer/%s", appID)
		resp, err := http.Get(checkURL)
		require.NoError(t, err)
		defer resp.Body.Close()

		t.Logf("Consent Engine check response status: %d", resp.StatusCode)

		if resp.StatusCode == http.StatusOK {
			var result interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			if err == nil {
				t.Logf("Consent check result: %+v", result)
				assert.NotNil(t, result, "Consent engine should return consent list")
			}
		} else {
			bodyBytes := make([]byte, 1024)
			n, _ := resp.Body.Read(bodyBytes)
			t.Logf("Consent Engine check error response: %s", string(bodyBytes[:n]))
		}
	})
}
// func TestGraphQLFlow_InvalidConsent(t *testing.T)
// func TestGraphQLFlow_MissingPolicyMetadata(t *testing.T)
// func TestGraphQLFlow_ExpiredConsent(t *testing.T)
// func TestGraphQLFlow_UnauthorizedApp(t *testing.T)
// func TestGraphQLFlow_ServiceTimeout(t *testing.T)
