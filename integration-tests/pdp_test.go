package main

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
	pdpBaseURL = "http://127.0.0.1:8083/api/v1/policy"
)

func TestMain(m *testing.M) {
	// Simple wait for service availability
	waitForService(pdpBaseURL + "/metadata") // This endpoint might 405 on GET, but connection should work

	code := m.Run()
	os.Exit(code)
}

func waitForService(url string) {
	for i := 0; i < 30; i++ {
		resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte("{}")))
		// We expect 400 Bad Request or similar if service is up, but connection refused if down
		if err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(1 * time.Second)
	}
	fmt.Println("Service might not be ready, proceeding anyway...")
}

func TestPDP_Flow(t *testing.T) {
	schemaID := "schema-test-123"
	fieldName := "email"
	appID := "consumer-app-1"

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
	jsonData, _ := json.Marshal(reqBody)

	resp, err := http.Post(pdpBaseURL+"/metadata", "application/json", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check status code
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var createResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&createResult)
	
	records := createResult["records"].([]interface{})
	require.NotEmpty(t, records)
	record := records[0].(map[string]interface{})
	assert.NotEmpty(t, record["id"])
	assert.Equal(t, schemaID, record["schemaId"])
	assert.Equal(t, fieldName, record["fieldName"])

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
		updateJson, _ := json.Marshal(updateBody)
		
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
		decisionJson, _ := json.Marshal(decisionBody)

		resp, err := http.Post(pdpBaseURL+"/decide", "application/json", bytes.NewBuffer(decisionJson))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		var decisionResult map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&decisionResult)
		
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
		decisionJson, _ := json.Marshal(decisionBody)

		resp, err := http.Post(pdpBaseURL+"/decide", "application/json", bytes.NewBuffer(decisionJson))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		var decisionResult map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&decisionResult)
		
		// appAuthorized should be false
		assert.Equal(t, false, decisionResult["appAuthorized"])
	})
}
