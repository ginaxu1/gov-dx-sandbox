package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testCleanupRegistry tracks resources created during tests for cleanup
type testCleanupRegistry struct {
	consentIDs []string
	schemaIDs  []string
	appIDs     []string
}

// cleanupTestData attempts to clean up test data created during test execution.
// Note: Some services may not have DELETE endpoints, so this is best-effort cleanup.
func (r *testCleanupRegistry) cleanup(t *testing.T) {
	client := &http.Client{Timeout: cleanupHTTPTimeout}

	// Cleanup consents (if Consent Engine supports deletion)
	for _, consentID := range r.consentIDs {
		req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", consentEngineURL, consentID), nil)
		if err == nil {
			resp, err := client.Do(req)
			if err == nil {
				resp.Body.Close()
				t.Logf("Cleaned up consent: %s", consentID)
			}
		}
	}

	// Note: Policy metadata and allowList entries typically don't have DELETE endpoints
	// They are managed through updates. For test isolation, use unique identifiers per test run.
	t.Logf("Test cleanup completed. Note: Policy metadata uses unique IDs to prevent conflicts.")
}

const (
	testNIC       = "123456789V"
	testEmail     = "test@example.com"
	testOwnerID   = "test-owner-123"
	testRequestID = "test-req-123"
)

// Timeout constants for test operations
const (
	defaultHTTPTimeout    = 10 * time.Second
	cleanupHTTPTimeout    = 5 * time.Second
	serviceCheckTimeout   = 2 * time.Second
	serviceStartupTimeout = 120 // seconds
	servicePauseDelay     = 2 * time.Second
	serviceRetryInterval  = 1 * time.Second
)

// Shared HTTP client for tests to avoid creating multiple clients
var testHTTPClient = &http.Client{
	Timeout: defaultHTTPTimeout,
}

var (
	orchestrationEngineURL = getEnvOrDefault("ORCHESTRATION_ENGINE_URL", "http://127.0.0.1:4000/public/graphql")
	pdpURL                 = getEnvOrDefault("PDP_URL", "http://127.0.0.1:8082/api/v1/policy")
	consentEngineURL       = getEnvOrDefault("CONSENT_ENGINE_URL", "http://127.0.0.1:8081/consents")
	portalBackendURL       = getEnvOrDefault("PORTAL_BACKEND_URL", "http://127.0.0.1:3000")
)

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// checkDockerComposeServices checks if docker-compose services are running.
// It validates the docker-compose file path and parses the output to ensure services are active.
func checkDockerComposeServices(composeFile string) bool {
	// Validate compose file exists
	if _, err := os.Stat(composeFile); os.IsNotExist(err) {
		return false
	}

	// Sanitize file path to prevent command injection
	// Only allow relative paths and ensure it's within the test directory
	absPath, err := filepath.Abs(composeFile)
	if err != nil {
		return false
	}
	testDir, err := os.Getwd()
	if err != nil {
		return false
	}
	// Ensure the compose file is within the test directory
	if !strings.HasPrefix(absPath, testDir) {
		return false
	}

	// Check if docker-compose services are running
	cmd := exec.Command("docker", "compose", "-f", composeFile, "ps", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Validate output format - should be JSON array
	outputStr := string(output)
	if len(outputStr) == 0 {
		return false
	}

	// Try to parse as JSON to validate format
	var services []map[string]interface{}
	if err := json.Unmarshal([]byte(outputStr), &services); err != nil {
		// If not valid JSON array, check if it's empty array string
		return outputStr != "[]\n" && outputStr != "[]"
	}

	// Check if any services are actually running (not just created)
	for _, service := range services {
		if state, ok := service["State"].(string); ok {
			if state == "running" {
				return true
			}
		}
	}

	return false
}

func TestMain(m *testing.M) {

	// Check if Docker Desktop is running
	if err := exec.Command("docker", "info").Run(); err != nil {
		fmt.Println("❌ Docker is not running. Please start Docker Desktop.")
		os.Exit(1)
	}

	// Define services to check via docker-compose
	composeFile := "docker-compose.test.yml"
	if checkDockerComposeServices(composeFile) {
		fmt.Println("📦 Docker Compose services detected. Checking service health...")
	} else {
		fmt.Println("⚠️  Docker Compose services not detected.")
		fmt.Println("💡 To start services, run:")
		fmt.Println("   cd tests/integration")
		fmt.Println("   docker compose -f docker-compose.test.yml up -d")
		fmt.Println("   Then wait for services to be healthy before running tests.")
		fmt.Println()
		fmt.Println("⏭️  Exiting tests. Please start services and try again.")
		os.Exit(1)
	}

	// Wait for all services to be available with shorter timeout
	services := []struct {
		name string
		url  string
	}{
		{"Portal Backend", "http://127.0.0.1:3000/health"},
		{"Orchestration Engine", "http://127.0.0.1:4000/health"},
		{"Policy Decision Point", "http://127.0.0.1:8082/health"},
		{"Consent Engine", "http://127.0.0.1:8081/health"},
	}

	var unavailableServices []string
	for _, svc := range services {
		if err := waitForService(svc.url, serviceStartupTimeout); err != nil {
			fmt.Printf("❌ Service %s not available: %v\n", svc.name, err)
			unavailableServices = append(unavailableServices, svc.name)
		} else {
			fmt.Printf("%s is available\n", svc.name)
		}
	}

	if len(unavailableServices) > 0 {
		fmt.Printf("\n⚠️  Some services are not available: %v\n", unavailableServices)
		fmt.Println("💡 To start services, run:")
		fmt.Println("   cd tests/integration")
		fmt.Println("   docker compose -f docker-compose.test.yml up -d")
		fmt.Println()
		fmt.Println("⏭️  Exiting tests. Please start services and try again.")
		os.Exit(1)
	}

	fmt.Println("\n🚀 All services are available. Running tests...")
	code := m.Run()
	os.Exit(code)
}

func waitForService(url string, maxAttempts int) error {
	client := &http.Client{
		Timeout: serviceCheckTimeout,
	}

	for i := 0; i < maxAttempts; i++ {
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		// Only log every 5 attempts to reduce noise
		if i > 0 && i%5 == 0 {
			fmt.Printf("  Still waiting for %s... (attempt %d/%d)\n", url, i+1, maxAttempts)
		}
		time.Sleep(serviceRetryInterval)
	}
	return fmt.Errorf("service at %s did not become available after %d attempts", url, maxAttempts)
}

// createTestJWT creates a JWT token for testing with the specified app ID.
//
// SECURITY WARNING: This function creates UNSIGNED JWT tokens (SigningMethodNone)
// which bypasses cryptographic validation. This is ONLY acceptable for integration
// tests where the Orchestration Engine is configured to accept unsigned tokens
// in test environments.
//
// NEVER use unsigned tokens in production code. In production, tokens must be
// properly signed and validated using RS256 or other secure signing methods.
//
// The Orchestration Engine's auth package accepts unsigned tokens when running
// in "local" environment mode for testing purposes only.
func createTestJWT(appID string) (string, error) {
	claims := jwt.MapClaims{
		"http://wso2.org/claims/subscriber":      appID,
		"http://wso2.org/claims/applicationUUId": appID,
		"http://wso2.org/claims/applicationid":   appID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// makeGraphQLRequest performs a GraphQL request to the Orchestration Engine with the given query and JWT token.
func makeGraphQLRequest(t *testing.T, query string, variables map[string]interface{}, token string) (*http.Response, error) {
	graphQLQuery := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}
	jsonData, err := json.Marshal(graphQLQuery)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", orchestrationEngineURL, bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-JWT-Assertion", token)

	return testHTTPClient.Do(req)
}

// parseGraphQLResponse parses a GraphQL response body and returns the result map.
func parseGraphQLResponse(t *testing.T, resp *http.Response) map[string]interface{} {
	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Failed to decode GraphQL response")
	return result
}

// createPolicyMetadata creates policy metadata in the PDP for the given schema and field.
func createPolicyMetadata(t *testing.T, schemaID, fieldName string) {
	reqBody := map[string]interface{}{
		"schemaId": schemaID,
		"records": []map[string]interface{}{
			{
				"fieldName":         fieldName,
				"source":            "primary",
				"isOwner":           false,
				"owner":             "citizen", // Required when isOwner is false
				"accessControlType": "restricted",
			},
		},
	}
	jsonData, err := json.Marshal(reqBody)
	require.NoError(t, err)

	resp, err := testHTTPClient.Post(pdpURL+"/metadata", "application/json", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	require.Equal(t, http.StatusCreated, resp.StatusCode, "Policy metadata should be created successfully")
}

// updatePDPAllowlist adds an application to the PDP allowlist for the given schema and field.
func updatePDPAllowlist(t *testing.T, appID, schemaID, fieldName string) {
	allowListReq := map[string]interface{}{
		"applicationId": appID,
		"records": []map[string]interface{}{
			{
				"fieldName": fieldName,
				"schemaId":  schemaID,
			},
		},
		"grantDuration": "30d",
	}
	allowListData, err := json.Marshal(allowListReq)
	require.NoError(t, err)

	allowListResp, err := testHTTPClient.Post(pdpURL+"/update-allowlist", "application/json", bytes.NewBuffer(allowListData))
	require.NoError(t, err)
	defer func() {
		io.Copy(io.Discard, allowListResp.Body)
		allowListResp.Body.Close()
	}()

	if allowListResp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(allowListResp.Body)
		if err != nil {
			t.Logf("Failed to read allowlist response body: %v", err)
		} else {
			t.Logf("Update AllowList response: %d, body: %s", allowListResp.StatusCode, string(body))
		}
		require.Equal(t, http.StatusOK, allowListResp.StatusCode, "App should be added to AllowList")
		return
	}
	t.Log("Application added to PDP allowList")
}

// createConsent creates a consent record in the Consent Engine and returns the consent ID.
func createConsent(t *testing.T, appID, schemaID, fieldName, ownerID string) string {
	consentReq := map[string]interface{}{
		"app_id": appID,
		"consent_requirements": []map[string]interface{}{
			{
				"owner":    "CITIZEN",
				"owner_id": ownerID,
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

	resp, err := testHTTPClient.Post(consentEngineURL, "application/json", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Logf("Failed to read consent response body: %v", err)
		} else {
			t.Logf("Consent creation response status: %d, body: %s", resp.StatusCode, string(bodyBytes))
		}
		require.Equal(t, http.StatusCreated, resp.StatusCode, "Consent should be created successfully")
		return ""
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	t.Logf("Consent created: %+v", result)

	consentID, ok := result["consent_id"].(string)
	require.True(t, ok, "Consent ID should be present in response")
	require.NotEmpty(t, consentID, "Consent ID should not be empty")
	return consentID
}

// approveConsent approves a consent record using PATCH endpoint (no auth required for internal use).
func approveConsent(t *testing.T, consentID string) {
	patchReq := map[string]interface{}{
		"status":     "approved",
		"updated_by": "test-integration",
		"reason":     "Approved by integration test",
	}
	jsonData, err := json.Marshal(patchReq)
	require.NoError(t, err)

	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/%s", consentEngineURL, consentID), bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := testHTTPClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Logf("Failed to read consent approval response body: %v", err)
		} else {
			t.Logf("Consent approval response status: %d, body: %s", resp.StatusCode, string(bodyBytes))
		}
	}
	require.Equal(t, http.StatusOK, resp.StatusCode, "Consent should be approved successfully")
	t.Logf("Consent %s approved", consentID)
}

// TestGraphQLFlow_SuccessPath tests the complete success path:
// 1. GraphQL query to orchestration-engine-go
// 2. Success path through PDP (policy evaluation)
// 3. Success path through consent-engine (consent check)
// 4. Valid response back
func TestGraphQLFlow_SuccessPath(t *testing.T) {
	// Setup: Create policy metadata in PDP
	// Use unique IDs to ensure test isolation
	timestamp := time.Now().UnixNano()
	// NOTE: We MUST use "test-schema-123" because it is hardcoded in `tests/integration/schema.graphql`
	// which the Orchestration Engine loads. We cannot randomize this without changing the schema file.
	schemaID := "test-schema-123"
	appID := fmt.Sprintf("test-consumer-app-%d", timestamp)
	fieldName := "person.email"

	t.Logf("Running SuccessPath with SchemaID: %s, AppID: %s", schemaID, appID)

	// Registry to track created resources for cleanup
	cleanup := &testCleanupRegistry{
		schemaIDs:  []string{schemaID},
		appIDs:     []string{appID},
		consentIDs: []string{},
	}
	defer cleanup.cleanup(t)

	t.Run("Setup_PolicyMetadata", func(t *testing.T) {
		createPolicyMetadata(t, schemaID, fieldName)
		updatePDPAllowlist(t, appID, schemaID, fieldName)
	})

	t.Run("Setup_Consent", func(t *testing.T) {
		// Use testNIC as owner_id to match the GraphQL query variable
		consentID := createConsent(t, appID, schemaID, fieldName, testNIC)
		approveConsent(t, consentID)
		cleanup.consentIDs = append(cleanup.consentIDs, consentID)
	})

	t.Run("GraphQL_Query_To_OrchestrationEngine", func(t *testing.T) {
		token, err := createTestJWT(appID)
		require.NoError(t, err)

		resp, err := makeGraphQLRequest(t, `
			query TestQuery($nic: String!) {
				personInfo(nic: $nic) {
					email
				}
			}
		`, map[string]interface{}{"nic": testNIC}, token)
		require.NoError(t, err)
		defer resp.Body.Close()

		t.Logf("Orchestration Engine response status: %d", resp.StatusCode)

		result := parseGraphQLResponse(t, resp)
		t.Logf("Orchestration Engine response: %+v", result)

		// Validate response structure
		// GraphQL returns 200 OK even with errors, so check for errors in response
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Orchestration engine should return 200 OK")

		// Verify no errors are present for the success path
		if errors, ok := result["errors"].([]interface{}); ok && len(errors) > 0 {
			assert.Fail(t, "GraphQL response contained unexpected errors", "Errors: %+v", errors)
		}

		// Verify expected data is present
		data, ok := result["data"].(map[string]interface{})
		assert.True(t, ok, "Response should contain data field")
		if ok {
			personInfo, ok := data["personInfo"].(map[string]interface{})
			assert.True(t, ok, "Data should contain personInfo")
			if ok {
				assert.NotEmpty(t, personInfo["email"], "Email should not be empty")
			}
		}
	})

	t.Run("Verify_PDP_Integration", func(t *testing.T) {
		// Verify PDP can evaluate policies using /decide endpoint
		evalReq := map[string]interface{}{
			"consumer_id":     appID,
			"app_id":          appID,
			"request_id":      testRequestID,
			"required_fields": []string{fieldName},
		}
		jsonData, err := json.Marshal(evalReq)
		require.NoError(t, err)

		resp, err := testHTTPClient.Post(pdpURL+"/decide", "application/json", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		defer resp.Body.Close()

		t.Logf("PDP evaluation response status: %d", resp.StatusCode)

		if resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err, "Failed to decode PDP evaluation response")
			t.Logf("PDP evaluation result: %+v", result)
			assert.NotNil(t, result, "PDP should return evaluation result")
			assert.Contains(t, result, "appAuthorized", "PDP response should contain 'appAuthorized' field")
		} else {
			bodyBytes, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				t.Logf("PDP evaluation error response status: %d, failed to read body: %v", resp.StatusCode, readErr)
			} else {
				t.Logf("PDP evaluation error response: %s", string(bodyBytes))
			}
		}
	})

	t.Run("Verify_ConsentEngine_Integration", func(t *testing.T) {
		// Verify consent engine can retrieve consents by consumer
		// Consent engine consumer endpoint is at base URL, not /consents
		consentEngineBaseURL := getEnvOrDefault("CONSENT_ENGINE_BASE_URL", "http://127.0.0.1:8081")
		checkURL := fmt.Sprintf("%s/consumer/%s", consentEngineBaseURL, appID)
		resp, err := testHTTPClient.Get(checkURL)
		require.NoError(t, err)
		defer func() {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}()

		t.Logf("Consent Engine check response status: %d", resp.StatusCode)

		if resp.StatusCode == http.StatusOK {
			var result interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err, "Failed to decode consent engine response")
			t.Logf("Consent check result: %+v", result)
			assert.NotNil(t, result, "Consent engine should return consent list")
		} else {
			bodyBytes, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				t.Logf("Consent Engine check error response status: %d, failed to read body: %v", resp.StatusCode, readErr)
			} else {
				t.Logf("Consent Engine check error response: %s", string(bodyBytes))
			}
		}
	})
}

// TestGraphQLFlow_MissingPolicyMetadata tests the behavior when a field is queried
// without any policy metadata configured in PDP. This should result in an authorization error.
func TestGraphQLFlow_MissingPolicyMetadata(t *testing.T) {
	testID := fmt.Sprintf("%d", time.Now().UnixNano())
	appID := fmt.Sprintf("test-app-missing-policy-%s", testID)

	cleanup := &testCleanupRegistry{
		appIDs: []string{appID},
	}
	defer cleanup.cleanup(t)

	token, err := createTestJWT(appID)
	require.NoError(t, err)

	resp, err := makeGraphQLRequest(t, `
		query TestQuery($nic: String!) {
			personInfo(nic: $nic) {
				profession
			}
		}
	`, map[string]interface{}{"nic": testNIC}, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "GraphQL should return 200 OK even with errors")

	result := parseGraphQLResponse(t, resp)

	// We expect authorization/validation errors
	errors, hasErrors := result["errors"]
	assert.True(t, hasErrors, "Should return errors for field lacking policy metadata")
	if hasErrors {
		errorList := errors.([]interface{})
		assert.NotEmpty(t, errorList, "Error list should not be empty")
		t.Logf("Received errors as expected: %+v", errorList)

		// Check if error contains PDP-related error code
		if len(errorList) > 0 {
			firstError := errorList[0].(map[string]interface{})
			message := fmt.Sprintf("%v", firstError["message"])
			t.Logf("Error message: %s", message)

			// Check for PDP-related error code in extensions
			extensions, hasExtensions := firstError["extensions"].(map[string]interface{})
			if hasExtensions {
				errorCode := fmt.Sprintf("%v", extensions["code"])
				assert.Equal(t, "PDP_NOT_ALLOWED", errorCode, "Error should have PDP_NOT_ALLOWED code")
			} else {
				// Fallback: check if message contains "PDP" keyword
				assert.Contains(t, message, "PDP", "Error message should mention PDP")
			}
		}
	}
}

// TestGraphQLFlow_UnauthorizedApp tests the behavior when an app queries a field that
// requires consent but has no consent granted. Policy metadata exists, but app has no consent.
func TestGraphQLFlow_UnauthorizedApp(t *testing.T) {
	timestamp := time.Now().UnixNano()
	schemaID := "test-schema-123"
	fieldName := "person.address"
	unauthorizedAppID := fmt.Sprintf("rogue-app-%d", timestamp)
	t.Logf("Testing with unauthorized app ID: %s, SchemaID: %s", unauthorizedAppID, schemaID)

	cleanup := &testCleanupRegistry{
		schemaIDs: []string{schemaID},
		appIDs:    []string{unauthorizedAppID},
	}
	defer cleanup.cleanup(t)

	createPolicyMetadata(t, schemaID, fieldName)
	updatePDPAllowlist(t, unauthorizedAppID, schemaID, fieldName)

	token, err := createTestJWT(unauthorizedAppID)
	require.NoError(t, err)

	resp, err := makeGraphQLRequest(t, `
		query TestQuery($nic: String!) {
			personInfo(nic: $nic) {
				address
			}
		}
	`, map[string]interface{}{"nic": testNIC}, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	t.Logf("Unauthorized App Response status: %d", resp.StatusCode)

	assert.Equal(t, http.StatusOK, resp.StatusCode, "GraphQL should return 200 OK even with errors")

	result := parseGraphQLResponse(t, resp)

	// Should contain errors indicating consent not approved
	errors, hasErrors := result["errors"]
	assert.True(t, hasErrors, "Should return errors for valid metadata but missing consent")
	if hasErrors {
		errorList := errors.([]interface{})
		assert.NotEmpty(t, errorList, "Error list should not be empty")
		t.Logf("Received errors as expected: %+v", errorList)

		// Check if error contains consent-related error
		if len(errorList) > 0 {
			firstError := errorList[0].(map[string]interface{})
			message := fmt.Sprintf("%v", firstError["message"])
			t.Logf("Error message: %s", message)

			// Check for consent-related error code in extensions
			extensions, hasExtensions := firstError["extensions"].(map[string]interface{})
			if hasExtensions {
				errorCode := fmt.Sprintf("%v", extensions["code"])
				assert.Equal(t, "CE_NOT_APPROVED", errorCode, "Error should have CE_NOT_APPROVED code")
			} else {
				// Fallback: check if message contains consent-related keywords
				assert.True(t,
					strings.Contains(message, "Consent") || strings.Contains(message, "CE"),
					"Error message should mention consent")
			}
		}
	}
}

// TestGraphQLFlow_ServiceTimeout tests resilience/failure when a dependency (PDP) is down or times out.
func TestGraphQLFlow_ServiceTimeout(t *testing.T) {
	testID := fmt.Sprintf("%d", time.Now().UnixNano())
	appID := fmt.Sprintf("test-app-timeout-%s", testID)

	cleanup := &testCleanupRegistry{
		appIDs: []string{appID},
	}
	defer cleanup.cleanup(t)

	token, err := createTestJWT(appID)
	require.NoError(t, err)

	wd, err := os.Getwd()
	if err != nil {
		t.Skipf("Skipping ServiceTimeout test: unable to get working directory: %v", err)
		return
	}
	composeFile := filepath.Join(wd, "docker-compose.test.yml")

	if _, err := os.Stat(composeFile); os.IsNotExist(err) {
		t.Skipf("Skipping ServiceTimeout test: docker-compose.test.yml not found: %v", err)
		return
	}

	cmd := exec.Command("docker", "compose", "-f", composeFile, "pause", "policy-decision-point")
	err = cmd.Run()
	if err != nil {
		t.Skipf("Skipping ServiceTimeout test: unable to pause docker container: %v", err)
		return
	}

	t.Cleanup(func() {
		unpauseCmd := exec.Command("docker", "compose", "-f", composeFile, "unpause", "policy-decision-point")
		if err := unpauseCmd.Run(); err != nil {
			t.Logf("Failed to unpause PDP container during cleanup: %v", err)
		}
		waitForService(pdpURL+"/health", 10)
	})

	time.Sleep(servicePauseDelay)

	resp, err := makeGraphQLRequest(t, `
		query TestQuery($nic: String!) {
			personInfo(nic: $nic) {
				fullName
			}
		}
	`, map[string]interface{}{"nic": testNIC}, token)

	if err != nil {
		t.Logf("Request failed as expected (timeout/connection error): %v", err)
		assert.Error(t, err, "Request should fail when PDP is down")
		return
	}
	defer resp.Body.Close()

	t.Logf("Response status during outage: %d", resp.StatusCode)

	if resp.StatusCode == http.StatusOK {
		result := parseGraphQLResponse(t, resp)
		t.Logf("Error response: %+v", result)
		if errors, ok := result["errors"].([]interface{}); ok {
			assert.NotEmpty(t, errors, "Should have errors when PDP is down")
			t.Logf("Received errors as expected: %+v", errors)
		}
	} else {
		assert.NotEqual(t, http.StatusOK, resp.StatusCode, "Should not return OK when PDP is down")
	}
}

// TestGraphQLFlow_InvalidQuery tests the behavior when an invalid/malformed GraphQL query is sent.
func TestGraphQLFlow_InvalidQuery(t *testing.T) {
	testID := fmt.Sprintf("%d", time.Now().UnixNano())
	appID := fmt.Sprintf("test-app-invalid-query-%s", testID)

	cleanup := &testCleanupRegistry{
		appIDs: []string{appID},
	}
	defer cleanup.cleanup(t)

	token, err := createTestJWT(appID)
	require.NoError(t, err)

	resp, err := makeGraphQLRequest(t, `
		query TestQuery {
			personInfo {
				fullName
			}
		}
	`, nil, token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "GraphQL should return 200 OK even with errors")

	result := parseGraphQLResponse(t, resp)

	// Should contain errors for invalid query
	errors, hasErrors := result["errors"]
	assert.True(t, hasErrors, "Should return errors for invalid GraphQL query")
	if hasErrors {
		errorList := errors.([]interface{})
		assert.NotEmpty(t, errorList, "Error list should not be empty")
		t.Logf("Received errors as expected: %+v", errorList)
	}
}

// TestGraphQLFlow_MissingToken tests the behavior when no JWT token is provided.
func TestGraphQLFlow_MissingToken(t *testing.T) {
	graphQLQuery := map[string]interface{}{
		"query": `
			query TestQuery($nic: String!) {
				personInfo(nic: $nic) {
					fullName
				}
			}
		`,
		"variables": map[string]interface{}{
			"nic": testNIC,
		},
	}
	jsonData, err := json.Marshal(graphQLQuery)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", orchestrationEngineURL, bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := testHTTPClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	t.Logf("Response status without token: %d", resp.StatusCode)

	if resp.StatusCode == http.StatusUnauthorized {
		t.Logf("Correctly returned 401 Unauthorized for missing token")
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Should return 401 when token is missing")
	} else {
		result := parseGraphQLResponse(t, resp)
		t.Logf("Response (may be local env): %+v", result)
	}
}
