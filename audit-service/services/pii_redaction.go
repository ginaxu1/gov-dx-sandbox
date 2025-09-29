package services

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
)

// PIIRedactionService handles redaction of PII data
type PIIRedactionService struct {
	// Pattern to match ID fields that exist in our database schema (_id suffix)
	idPattern *regexp.Regexp
}

// NewPIIRedactionService creates a new PII redaction service
func NewPIIRedactionService() *PIIRedactionService {
	// Compile regex pattern to match ID fields that exist in our database schema
	// Matches: consumer_id, provider_id, consent_id, owner_id, entity_id, submission_id, schema_id
	pattern := regexp.MustCompile(`(?i)(_id)$`)

	return &PIIRedactionService{
		idPattern: pattern,
	}
}

// RedactData removes all ID fields from the data (consumer_id, provider_id, etc.)
func (s *PIIRedactionService) RedactData(data interface{}) (interface{}, error) {
	switch v := data.(type) {
	case map[string]interface{}:
		return s.redactMap(v), nil
	case []interface{}:
		return s.redactArray(v), nil
	default:
		return v, nil
	}
}

// redactMap recursively redacts a map
func (s *PIIRedactionService) redactMap(m map[string]interface{}) map[string]interface{} {
	redacted := make(map[string]interface{})

	for key, value := range m {
		// Check if the key ends with _id (matches our database schema)
		if s.idPattern.MatchString(key) {
			// Skip this field entirely
			continue
		}

		// Recursively redact the value
		switch v := value.(type) {
		case map[string]interface{}:
			redacted[key] = s.redactMap(v)
		case []interface{}:
			redacted[key] = s.redactArray(v)
		default:
			redacted[key] = v
		}
	}

	return redacted
}

// redactArray recursively redacts an array
func (s *PIIRedactionService) redactArray(arr []interface{}) []interface{} {
	redacted := make([]interface{}, 0, len(arr))

	for _, item := range arr {
		switch v := item.(type) {
		case map[string]interface{}:
			redacted = append(redacted, s.redactMap(v))
		case []interface{}:
			redacted = append(redacted, s.redactArray(v))
		default:
			redacted = append(redacted, v)
		}
	}

	return redacted
}

// ExtractAndHashCitizenID extracts citizen ID from data and returns both the hash and redacted data
func (s *PIIRedactionService) ExtractAndHashCitizenID(data interface{}) (string, interface{}, error) {
	var citizenID string
	var redactedData interface{}

	// First, extract the citizen ID before redaction
	citizenID = s.ExtractCitizenID(data)

	// Then redact the data
	redactedData, err := s.RedactData(data)
	if err != nil {
		return "", nil, fmt.Errorf("failed to redact data: %w", err)
	}

	// Hash the citizen ID
	hashedID := s.HashCitizenID(citizenID)

	return hashedID, redactedData, nil
}

// ExtractCitizenID extracts citizen ID from data (looks for common citizen ID field names)
func (s *PIIRedactionService) ExtractCitizenID(data interface{}) string {
	return s.extractCitizenIDPrivate(data)
}

// extractCitizenIDPrivate extracts citizen ID from data using regex pattern matching (private method to avoid recursion)
func (s *PIIRedactionService) extractCitizenIDPrivate(data interface{}) string {
	switch v := data.(type) {
	case map[string]interface{}:
		// Look for ID field names that exist in our database schema
		citizenIDFields := []string{
			"consumer_id", "provider_id", "consent_id", "owner_id", "entity_id",
			"submission_id", "schema_id",
		}

		for _, field := range citizenIDFields {
			if value, exists := v[field]; exists {
				if str, ok := value.(string); ok {
					return str
				}
			}
		}

		// If not found in common fields, recursively search
		for _, value := range v {
			if id := s.extractCitizenIDPrivate(value); id != "" {
				return id
			}
		}

	case []interface{}:
		// Search in array elements
		for _, item := range v {
			if id := s.extractCitizenIDPrivate(item); id != "" {
				return id
			}
		}
	}

	return ""
}

// HashCitizenID creates a secure hash of the citizen ID
func (s *PIIRedactionService) HashCitizenID(citizenID string) string {
	if citizenID == "" {
		return ""
	}

	// Use SHA-256 for hashing
	hash := sha256.Sum256([]byte(citizenID))
	return hex.EncodeToString(hash[:])
}

// RedactJSONString redacts PII from a JSON string
func (s *PIIRedactionService) RedactJSONString(jsonStr string) (string, error) {
	var data interface{}

	// Parse JSON
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Redact data
	redactedData, err := s.RedactData(data)
	if err != nil {
		return "", fmt.Errorf("failed to redact data: %w", err)
	}

	// Convert back to JSON
	redactedJSON, err := json.Marshal(redactedData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal redacted data: %w", err)
	}

	return string(redactedJSON), nil
}

// ExtractCitizenIDAndRedactJSONString extracts citizen ID and redacts PII from a JSON string
func (s *PIIRedactionService) ExtractCitizenIDAndRedactJSONString(jsonStr string) (string, string, error) {
	var data interface{}

	// Parse JSON
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Extract citizen ID and redact data
	hashedID, redactedData, err := s.ExtractAndHashCitizenID(data)
	if err != nil {
		return "", "", fmt.Errorf("failed to extract and redact: %w", err)
	}

	// Convert redacted data back to JSON
	redactedJSON, err := json.Marshal(redactedData)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal redacted data: %w", err)
	}

	return hashedID, string(redactedJSON), nil
}

// ValidateCitizenID validates if a citizen ID looks reasonable
func (s *PIIRedactionService) ValidateCitizenID(citizenID string) bool {
	if citizenID == "" {
		return false
	}

	// Basic validation - should be alphanumeric and reasonable length
	if len(citizenID) < 3 || len(citizenID) > 50 {
		return false
	}

	// Should contain only alphanumeric characters and common separators
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)
	return validPattern.MatchString(citizenID)
}
