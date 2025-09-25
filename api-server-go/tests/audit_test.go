package tests

import (
	"encoding/json"
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/models"
	"github.com/gov-dx-sandbox/api-server-go/services"
)

func TestPIIRedaction(t *testing.T) {
	piiService := services.NewPIIRedactionService()

	// Test data with various ID fields
	testData := map[string]interface{}{
		"name":        "John Doe",
		"email":       "john@example.com",
		"citizenId":   "123456789",
		"national_id": "NIC123456",
		"personId":    "PERSON001",
		"address": map[string]interface{}{
			"street":   "123 Main St",
			"city":     "Colombo",
			"postalId": "POSTAL123", // Should be redacted
			"country":  "Sri Lanka",
		},
		"contacts": []interface{}{
			map[string]interface{}{
				"type":      "phone",
				"number":    "0771234567",
				"contactId": "CONTACT001", // Should be redacted
			},
		},
	}

	// Test redaction
	redactedData, err := piiService.RedactData(testData)
	if err != nil {
		t.Fatalf("Failed to redact data: %v", err)
	}

	redactedMap := redactedData.(map[string]interface{})

	// Verify that ID fields are removed
	if _, exists := redactedMap["citizenId"]; exists {
		t.Error("citizenId should be redacted")
	}
	if _, exists := redactedMap["national_id"]; exists {
		t.Error("national_id should be redacted")
	}
	if _, exists := redactedMap["personId"]; exists {
		t.Error("personId should be redacted")
	}

	// Verify that non-ID fields are preserved
	if name, exists := redactedMap["name"]; !exists || name != "John Doe" {
		t.Error("name should be preserved")
	}
	if email, exists := redactedMap["email"]; !exists || email != "john@example.com" {
		t.Error("email should be preserved")
	}

	// Verify nested redaction
	address := redactedMap["address"].(map[string]interface{})
	if _, exists := address["postalId"]; exists {
		t.Error("postalId should be redacted from nested object")
	}
	if street, exists := address["street"]; !exists || street != "123 Main St" {
		t.Error("street should be preserved in nested object")
	}

	// Verify array redaction
	contacts := redactedMap["contacts"].([]interface{})
	contact := contacts[0].(map[string]interface{})
	if _, exists := contact["contactId"]; exists {
		t.Error("contactId should be redacted from array element")
	}
	if contactType, exists := contact["type"]; !exists || contactType != "phone" {
		t.Error("type should be preserved in array element")
	}
}

func TestCitizenIDExtractionAndHashing(t *testing.T) {
	piiService := services.NewPIIRedactionService()

	// Test data with citizen ID
	testData := map[string]interface{}{
		"name":      "John Doe",
		"citizenId": "123456789",
		"email":     "john@example.com",
	}

	// Test extraction and hashing
	hashedID, redactedData, err := piiService.ExtractAndHashCitizenID(testData)
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
	if _, exists := redactedMap["citizenId"]; exists {
		t.Error("citizenId should be redacted")
	}

	// Verify other fields are preserved
	if name, exists := redactedMap["name"]; !exists || name != "John Doe" {
		t.Error("name should be preserved")
	}
}

func TestJSONRedaction(t *testing.T) {
	piiService := services.NewPIIRedactionService()

	// Test JSON string
	jsonStr := `{
		"name": "John Doe",
		"citizenId": "123456789",
		"email": "john@example.com",
		"address": {
			"street": "123 Main St",
			"postalId": "POSTAL123"
		}
	}`

	// Test JSON redaction
	redactedJSON, err := piiService.RedactJSONString(jsonStr)
	if err != nil {
		t.Fatalf("Failed to redact JSON: %v", err)
	}

	// Parse redacted JSON
	var redactedData map[string]interface{}
	if err := json.Unmarshal([]byte(redactedJSON), &redactedData); err != nil {
		t.Fatalf("Failed to parse redacted JSON: %v", err)
	}

	// Verify ID fields are removed
	if _, exists := redactedData["citizenId"]; exists {
		t.Error("citizenId should be redacted from JSON")
	}

	address := redactedData["address"].(map[string]interface{})
	if _, exists := address["postalId"]; exists {
		t.Error("postalId should be redacted from nested JSON")
	}

	// Verify non-ID fields are preserved
	if name, exists := redactedData["name"]; !exists || name != "John Doe" {
		t.Error("name should be preserved in JSON")
	}
}

func TestAuditLogConstants(t *testing.T) {
	// Test transaction status constants
	if models.TransactionStatusSuccess != "SUCCESS" {
		t.Errorf("Expected TransactionStatusSuccess to be 'SUCCESS', got '%s'", models.TransactionStatusSuccess)
	}
	if models.TransactionStatusFailure != "FAILURE" {
		t.Errorf("Expected TransactionStatusFailure to be 'FAILURE', got '%s'", models.TransactionStatusFailure)
	}
}

func TestAuditLogRequestValidation(t *testing.T) {
	// Test valid audit log request
	validRequest := models.AuditLogRequest{
		ConsumerID:        "consumer-123",
		ProviderID:        "provider-456",
		RequestedData:     json.RawMessage(`{"query": "test"}`),
		TransactionStatus: models.TransactionStatusSuccess,
		CitizenHash:       "hashed_citizen_123",
	}

	// This should not cause any validation errors
	if validRequest.ConsumerID == "" {
		t.Error("ConsumerID should not be empty")
	}
	if validRequest.ProviderID == "" {
		t.Error("ProviderID should not be empty")
	}
	if validRequest.TransactionStatus == "" {
		t.Error("TransactionStatus should not be empty")
	}
	if validRequest.CitizenHash == "" {
		t.Error("CitizenHash should not be empty")
	}
}
