package pdp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/gov-dx-sandbox/tests/integration/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	pdpBaseURL = "http://127.0.0.1:8083/api/v1/policy"
)

func TestMain(m *testing.M) {
	// Simple wait for service availability
	if err := testutils.WaitForService(pdpBaseURL + "/metadata"); err != nil {
		fmt.Printf("Service not available: %v\n", err)
		os.Exit(1)
	}

	// Optionally connect to database for verification
	// This allows tests to verify database state if needed
	// Set TEST_VERIFY_DB=true to enable database verification
	if os.Getenv("TEST_VERIFY_DB") == "true" {
		// We'll set up the DB connection in individual tests if needed
		// This is optional - integration tests can work without direct DB access
	}

	code := m.Run()
	os.Exit(code)
}

func TestPDP_Flow(t *testing.T) {
	schemaID := "schema-test-123"
	fieldName := "email"
	appID := "consumer-app-1"

	// Optionally set up database connection for verification
	// This is optional - the test works without it, but enables DB state verification
	var db *gorm.DB
	if os.Getenv("TEST_VERIFY_DB") == "true" {
		db = testutils.SetupPostgresTestDB(t)
		if db != nil {
			defer testutils.CleanupTestData(t, db)
		}
	}

	// 1. Create Policy Metadata
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

	resp, err := http.Post(pdpBaseURL+"/metadata", "application/json", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check status code
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var createResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createResult)
	require.NoError(t, err)
	
	recordsRaw, ok := createResult["records"]
	require.True(t, ok, "response missing 'records' field")
	records, ok := recordsRaw.([]interface{})
	require.True(t, ok, "'records' field is not an array")
	require.NotEmpty(t, records, "'records' array is empty")
	recordRaw := records[0]
	record, ok := recordRaw.(map[string]interface{})
	require.True(t, ok, "record is not an object")
	assert.NotEmpty(t, record["id"])
	assert.Equal(t, schemaID, record["schemaId"])
	assert.Equal(t, fieldName, record["fieldName"])

	// Verify database state if DB connection is available
	if db != nil {
		assert.True(t, testutils.VerifyPolicyMetadataExists(t, db, schemaID, fieldName),
			"Policy metadata should exist in database after creation")
	}

	// 2. Update Allowlist
	t.Run("UpdateAllowlist", func(t *testing.T) {
		updateBody := map[string]interface{}{
			"applicationId": appID,
			"grantDuration": "30d",
			"records": []map[string]interface{}{
				{
					"schemaId":  schemaID,
					"fieldName": fieldName,
				},
			},
		}
		updateJson, err := json.Marshal(updateBody)
		require.NoError(t, err)
		
		resp, err := http.Post(pdpBaseURL+"/update-allowlist", "application/json", bytes.NewBuffer(updateJson))
		require.NoError(t, err)
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// 3. Check Decision (Allowed)
	t.Run("CheckDecision_Allowed", func(t *testing.T) {
		decisionBody := map[string]interface{}{
			"applicationId": appID,
			"requiredFields": []map[string]interface{}{
				{
					"schemaId":  schemaID,
					"fieldName": fieldName,
				},
			},
		}
		decisionJson, err := json.Marshal(decisionBody)
		require.NoError(t, err)

		resp, err := http.Post(pdpBaseURL+"/decide", "application/json", bytes.NewBuffer(decisionJson))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		var decisionResult map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&decisionResult)
		require.NoError(t, err)
		
		// appAuthorized should be true because we added it to allow list
		assert.Equal(t, true, decisionResult["appAuthorized"])
	})

	// 4. Check Decision (Denied - different app)
	t.Run("CheckDecision_Denied", func(t *testing.T) {
		decisionBody := map[string]interface{}{
			"applicationId": "unknown-app",
			"requiredFields": []map[string]interface{}{
				{
					"schemaId":  schemaID,
					"fieldName": fieldName,
				},
			},
		}
		decisionJson, err := json.Marshal(decisionBody)
		require.NoError(t, err)

		resp, err := http.Post(pdpBaseURL+"/decide", "application/json", bytes.NewBuffer(decisionJson))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		var decisionResult map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&decisionResult)
		require.NoError(t, err)
		
		// appAuthorized should be false
		assert.Equal(t, false, decisionResult["appAuthorized"])
	})
}
