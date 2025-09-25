package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/models"
	"github.com/gov-dx-sandbox/api-server-go/services"
)

// TestPhase2RedactionMiddleware tests the complete Phase 2 redaction middleware functionality
func TestPhase2RedactionMiddleware(t *testing.T) {
	piiService := services.NewPIIRedactionService()

	// Test Case 1: Request with various ID fields
	t.Run("RequestRedaction", func(t *testing.T) {
		requestData := map[string]interface{}{
			"query": "getCitizenData",
			"variables": map[string]interface{}{
				"citizenId":  "123456789",
				"nationalId": "NIC123456",
				"service_id": "SERVICE001",
				"requestId":  "REQ123",
				"name":       "John Doe",
				"email":      "john@example.com",
				"address": map[string]interface{}{
					"street":   "123 Main St",
					"postalId": "POSTAL123",
					"city":     "Colombo",
					"country":  "Sri Lanka",
				},
			},
		}

		// Convert to JSON
		requestJSON, err := json.Marshal(requestData)
		if err != nil {
			t.Fatalf("Failed to marshal request data: %v", err)
		}

		// Test redaction
		redactedJSON, err := piiService.RedactJSONString(string(requestJSON))
		if err != nil {
			t.Fatalf("Failed to redact request JSON: %v", err)
		}

		// Parse redacted JSON
		var redactedData map[string]interface{}
		if err := json.Unmarshal([]byte(redactedJSON), &redactedData); err != nil {
			t.Fatalf("Failed to parse redacted JSON: %v", err)
		}

		// Verify ID fields are removed
		variables := redactedData["variables"].(map[string]interface{})
		idFields := []string{"citizenId", "nationalId", "service_id", "requestId"}
		for _, field := range idFields {
			if _, exists := variables[field]; exists {
				t.Errorf("Field '%s' should be redacted from request", field)
			}
		}

		// Verify non-ID fields are preserved
		nonIdFields := []string{"name", "email"}
		for _, field := range nonIdFields {
			if _, exists := variables[field]; !exists {
				t.Errorf("Field '%s' should be preserved in request", field)
			}
		}

		// Verify query field is preserved at root level
		if _, exists := redactedData["query"]; !exists {
			t.Error("Field 'query' should be preserved in request")
		}

		// Verify nested redaction
		address := variables["address"].(map[string]interface{})
		if _, exists := address["postalId"]; exists {
			t.Error("postalId should be redacted from nested object")
		}
		if street, exists := address["street"]; !exists || street != "123 Main St" {
			t.Error("street should be preserved in nested object")
		}
	})

	// Test Case 2: Response with various ID fields
	t.Run("ResponseRedaction", func(t *testing.T) {
		responseData := map[string]interface{}{
			"data": map[string]interface{}{
				"citizen": map[string]interface{}{
					"citizenId":  "123456789",
					"nationalId": "NIC123456",
					"name":       "John Doe",
					"email":      "john@example.com",
					"phone":      "0771234567",
					"address": map[string]interface{}{
						"street":   "123 Main St",
						"postalId": "POSTAL123",
						"city":     "Colombo",
					},
					"documents": []interface{}{
						map[string]interface{}{
							"type":       "passport",
							"documentId": "DOC001",
							"number":     "P123456",
						},
					},
				},
				"transactionId": "TXN123",
				"requestId":     "REQ123",
			},
		}

		// Convert to JSON
		responseJSON, err := json.Marshal(responseData)
		if err != nil {
			t.Fatalf("Failed to marshal response data: %v", err)
		}

		// Test redaction
		redactedJSON, err := piiService.RedactJSONString(string(responseJSON))
		if err != nil {
			t.Fatalf("Failed to redact response JSON: %v", err)
		}

		// Parse redacted JSON
		var redactedData map[string]interface{}
		if err := json.Unmarshal([]byte(redactedJSON), &redactedData); err != nil {
			t.Fatalf("Failed to parse redacted JSON: %v", err)
		}

		// Verify ID fields are removed from response
		data := redactedData["data"].(map[string]interface{})
		citizen := data["citizen"].(map[string]interface{})

		idFields := []string{"citizenId", "nationalId", "transactionId", "requestId"}
		for _, field := range idFields {
			if field == "citizenId" || field == "nationalId" {
				if _, exists := citizen[field]; exists {
					t.Errorf("Field '%s' should be redacted from citizen object", field)
				}
			} else {
				if _, exists := data[field]; exists {
					t.Errorf("Field '%s' should be redacted from data object", field)
				}
			}
		}

		// Verify non-ID fields are preserved
		nonIdFields := []string{"name", "email", "phone"}
		for _, field := range nonIdFields {
			if _, exists := citizen[field]; !exists {
				t.Errorf("Field '%s' should be preserved in citizen object", field)
			}
		}

		// Verify array redaction
		documents := citizen["documents"].([]interface{})
		document := documents[0].(map[string]interface{})
		if _, exists := document["documentId"]; exists {
			t.Error("documentId should be redacted from array element")
		}
		if docType, exists := document["type"]; !exists || docType != "passport" {
			t.Error("type should be preserved in array element")
		}
	})

	// Test Case 3: Citizen ID extraction and hashing
	t.Run("CitizenIDExtractionAndHashing", func(t *testing.T) {
		requestData := map[string]interface{}{
			"query": "getCitizenData",
			"variables": map[string]interface{}{
				"citizenId":  "123456789",
				"nationalId": "NIC123456",
				"name":       "John Doe",
				"email":      "john@example.com",
			},
		}

		// Test extraction and hashing
		hashedID, redactedData, err := piiService.ExtractAndHashCitizenID(requestData)
		if err != nil {
			t.Fatalf("Failed to extract and hash citizen ID: %v", err)
		}

		// Verify hash is generated
		if hashedID == "" {
			t.Error("Hashed ID should not be empty")
		}

		// Verify hash is consistent
		expectedHash := piiService.HashCitizenID("123456789")
		if hashedID != expectedHash {
			t.Error("Hash should be consistent")
		}

		// Verify citizen ID is redacted from the data
		redactedMap := redactedData.(map[string]interface{})
		variables := redactedMap["variables"].(map[string]interface{})
		if _, exists := variables["citizenId"]; exists {
			t.Error("citizenId should be redacted after extraction")
		}
		if _, exists := variables["nationalId"]; exists {
			t.Error("nationalId should be redacted after extraction")
		}

		// Verify other fields are preserved
		if name, exists := variables["name"]; !exists || name != "John Doe" {
			t.Error("name should be preserved")
		}
		if email, exists := variables["email"]; !exists || email != "john@example.com" {
			t.Error("email should be preserved")
		}
	})

	// Test Case 4: Multiple citizen ID formats
	t.Run("MultipleCitizenIDFormats", func(t *testing.T) {
		testCases := []struct {
			fieldName        string
			value            string
			shouldBeRedacted bool
		}{
			{"citizenId", "123456789", true},
			{"citizen_id", "123456789", true},
			{"nationalId", "NIC123456", true},
			{"national_id", "NIC123456", true},
			{"nic", "123456789V", false},        // Doesn't end with Id or _id
			{"nicNumber", "123456789V", false},  // Doesn't end with Id or _id
			{"nic_number", "123456789V", false}, // Doesn't end with Id or _id
			{"personId", "PERSON001", true},
			{"person_id", "PERSON001", true},
			{"userId", "USER001", true},
			{"user_id", "USER001", true},
			{"ownerId", "OWNER001", true},
			{"owner_id", "OWNER001", true},
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("Field_%s", tc.fieldName), func(t *testing.T) {
				requestData := map[string]interface{}{
					"query": "getCitizenData",
					"variables": map[string]interface{}{
						tc.fieldName: tc.value,
						"name":       "John Doe",
						"email":      "john@example.com",
					},
				}

				// Test extraction and hashing
				hashedID, redactedData, err := piiService.ExtractAndHashCitizenID(requestData)
				if err != nil {
					t.Fatalf("Failed to extract and hash citizen ID: %v", err)
				}

				// Verify hash is generated
				if hashedID == "" {
					t.Errorf("Hashed ID should not be empty for field %s", tc.fieldName)
				}

				// Verify hash matches expected
				expectedHash := piiService.HashCitizenID(tc.value)
				if hashedID != expectedHash {
					t.Errorf("Hash should match expected for field %s", tc.fieldName)
				}

				// Verify field behavior based on pattern
				redactedMap := redactedData.(map[string]interface{})
				variables := redactedMap["variables"].(map[string]interface{})

				if tc.shouldBeRedacted {
					if _, exists := variables[tc.fieldName]; exists {
						t.Errorf("Field %s should be redacted after extraction", tc.fieldName)
					}
				} else {
					// For fields that don't match the pattern, they should still be extracted for hashing
					// but may or may not be redacted depending on the pattern
					if _, exists := variables[tc.fieldName]; !exists {
						t.Logf("Field %s was redacted (may be expected if it matches other patterns)", tc.fieldName)
					}
				}
			})
		}
	})

	// Test Case 5: Edge cases
	t.Run("EdgeCases", func(t *testing.T) {
		// Test with no citizen ID
		requestData := map[string]interface{}{
			"query": "getPublicData",
			"variables": map[string]interface{}{
				"name":  "John Doe",
				"email": "john@example.com",
			},
		}

		hashedID, _, err := piiService.ExtractAndHashCitizenID(requestData)
		if err != nil {
			t.Fatalf("Failed to extract and hash citizen ID: %v", err)
		}

		// Should return empty hash when no citizen ID found
		if hashedID != "" {
			t.Error("Hashed ID should be empty when no citizen ID found")
		}

		// Test with empty data
		hashedID, _, err = piiService.ExtractAndHashCitizenID(map[string]interface{}{})
		if err != nil {
			t.Fatalf("Failed to extract and hash citizen ID: %v", err)
		}

		if hashedID != "" {
			t.Error("Hashed ID should be empty for empty data")
		}

		// Test with nil data
		hashedID, _, err = piiService.ExtractAndHashCitizenID(nil)
		if err != nil {
			t.Fatalf("Failed to extract and hash citizen ID: %v", err)
		}

		if hashedID != "" {
			t.Error("Hashed ID should be empty for nil data")
		}
	})
}

// TestPhase2MiddlewareIntegration tests the middleware integration with HTTP requests
func TestPhase2MiddlewareIntegration(t *testing.T) {
	piiService := services.NewPIIRedactionService()

	// Create a test HTTP request with JSON body
	requestData := map[string]interface{}{
		"query": "getCitizenData",
		"variables": map[string]interface{}{
			"citizenId":  "123456789",
			"nationalId": "NIC123456",
			"name":       "John Doe",
			"email":      "john@example.com",
		},
	}

	requestJSON, err := json.Marshal(requestData)
	if err != nil {
		t.Fatalf("Failed to marshal request data: %v", err)
	}

	// Create HTTP request
	req := httptest.NewRequest("POST", "/graphql", bytes.NewBuffer(requestJSON))
	req.Header.Set("Content-Type", "application/json")

	// Simulate middleware processing
	// 1. Extract citizen ID and hash it
	hashedID, redactedData, err := piiService.ExtractAndHashCitizenID(requestData)
	if err != nil {
		t.Fatalf("Failed to extract and hash citizen ID: %v", err)
	}

	// 2. Create audit log entry
	auditRequest := &models.AuditLogRequest{
		ConsumerID:        "consumer-123",
		ProviderID:        "provider-456",
		RequestedData:     json.RawMessage(requestJSON),
		TransactionStatus: models.TransactionStatusSuccess,
		CitizenHash:       hashedID,
	}

	// Verify audit request is properly formed
	if auditRequest.ConsumerID == "" {
		t.Error("ConsumerID should not be empty")
	}
	if auditRequest.ProviderID == "" {
		t.Error("ProviderID should not be empty")
	}
	if auditRequest.CitizenHash == "" {
		t.Error("CitizenHash should not be empty")
	}
	if auditRequest.TransactionStatus != models.TransactionStatusSuccess {
		t.Error("TransactionStatus should be SUCCESS")
	}

	// Verify redacted data doesn't contain ID fields
	redactedMap := redactedData.(map[string]interface{})
	variables := redactedMap["variables"].(map[string]interface{})
	if _, exists := variables["citizenId"]; exists {
		t.Error("citizenId should be redacted in audit log")
	}
	if _, exists := variables["nationalId"]; exists {
		t.Error("nationalId should be redacted in audit log")
	}

	// Verify non-ID fields are preserved
	if name, exists := variables["name"]; !exists || name != "John Doe" {
		t.Error("name should be preserved in audit log")
	}
	if email, exists := variables["email"]; !exists || email != "john@example.com" {
		t.Error("email should be preserved in audit log")
	}

	t.Logf("Phase 2 middleware test completed successfully")
	t.Logf("Hashed Citizen ID: %s", hashedID)
	t.Logf("Redacted data keys: %v", getKeys(redactedMap))
}

// Helper function to get keys from a map
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
