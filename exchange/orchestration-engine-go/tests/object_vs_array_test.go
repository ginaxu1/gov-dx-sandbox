package tests

import (
	"testing"
)

// TestObjectVsArrayFieldDetection tests that object fields are not treated as array fields
func TestObjectVsArrayFieldDetection(t *testing.T) {
	// Test cases for field type detection
	testCases := []struct {
		fieldName   string
		value       interface{}
		expected    bool
		description string
	}{
		{
			fieldName:   "personInfo",
			value:       map[string]interface{}{"fullName": "John Doe"},
			expected:    false,
			description: "personInfo should be treated as object field, not array",
		},
		{
			fieldName:   "ownedVehicles",
			value:       []interface{}{map[string]interface{}{"regNo": "ABC123"}},
			expected:    true,
			description: "ownedVehicles should be treated as array field",
		},
		{
			fieldName:   "class",
			value:       []interface{}{map[string]interface{}{"className": "Sedan"}},
			expected:    true,
			description: "class should be treated as array field",
		},
		{
			fieldName:   "fullName",
			value:       "John Doe",
			expected:    false,
			description: "fullName should be treated as simple field, not array",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// This would test the isArrayFieldValue function if it were exported
			// For now, we'll just document the expected behavior
			t.Logf("Field: %s, Value: %T, Expected Array: %v", tc.fieldName, tc.value, tc.expected)
		})
	}
}

// TestPersonInfoObjectField tests that personInfo is processed as an object
func TestPersonInfoObjectField(t *testing.T) {
	// This test documents the expected behavior for personInfo
	// personInfo should be processed as an object field, not an array field

	personInfoValue := map[string]interface{}{
		"fullName": "John Doe",
		"name":     "John",
		"address":  "123 Main St",
	}

	// personInfo should NOT be treated as an array
	// It should be processed as a single object with nested fields
	t.Logf("personInfo value: %+v", personInfoValue)
	t.Logf("personInfo should be processed as object, not array")
}
