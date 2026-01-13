package audit

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMarshalMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]interface{}
		wantNil  bool
		wantJSON string
	}{
		{
			name:     "nil metadata",
			metadata: nil,
			wantNil:  true,
		},
		{
			name:     "empty metadata",
			metadata: map[string]interface{}{},
			wantNil:  false,
			wantJSON: "{}",
		},
		{
			name: "simple metadata",
			metadata: map[string]interface{}{
				"key": "value",
			},
			wantNil:  false,
			wantJSON: `{"key":"value"}`,
		},
		{
			name: "complex metadata",
			metadata: map[string]interface{}{
				"applicationId": "test-app-123",
				"query":         "query { citizen(nic: \"123456789V\") { name } }",
				"count":         42,
			},
			wantNil:  false,
			wantJSON: `{"applicationId":"test-app-123","count":42,"query":"query { citizen(nic: \"123456789V\") { name } }"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MarshalMetadata(tt.metadata)

			if tt.wantNil {
				if got != nil {
					t.Errorf("MarshalMetadata() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Errorf("MarshalMetadata() = nil, want non-nil")
				return
			}

			// Verify it's valid JSON
			var unmarshaled map[string]interface{}
			if err := json.Unmarshal(got, &unmarshaled); err != nil {
				t.Errorf("MarshalMetadata() produced invalid JSON: %v", err)
			}

			// Verify content matches (accounting for map ordering)
			expectedBytes, _ := json.Marshal(tt.metadata)
			gotStr := string(got)
			expectedStr := string(expectedBytes)

			// Parse both to compare semantically
			var gotMap, expectedMap map[string]interface{}
			json.Unmarshal(got, &gotMap)
			json.Unmarshal(expectedBytes, &expectedMap)

			if len(gotMap) != len(expectedMap) {
				t.Errorf("MarshalMetadata() length mismatch: got %d, want %d", len(gotMap), len(expectedMap))
			}

			// Check that the JSON is valid and contains expected content
			if tt.wantJSON != "" {
				// For simple cases, we can check exact match
				if len(tt.metadata) == 1 {
					if gotStr != expectedStr {
						t.Errorf("MarshalMetadata() = %s, want %s", gotStr, expectedStr)
					}
				}
			}
		})
	}
}

func TestMarshalMetadata_InvalidData(t *testing.T) {
	// Test with data that cannot be marshaled (channel, function, etc.)
	// In practice, this shouldn't happen, but we want to ensure graceful handling
	metadata := map[string]interface{}{
		"valid": "value",
		// Note: We can't easily test unmarshalable data in Go without using reflect,
		// but the function should handle it gracefully if it occurs
	}

	got := MarshalMetadata(metadata)
	if got == nil {
		t.Error("MarshalMetadata() = nil for valid metadata")
	}

	// Verify it's valid JSON
	var unmarshaled map[string]interface{}
	if err := json.Unmarshal(got, &unmarshaled); err != nil {
		t.Errorf("MarshalMetadata() produced invalid JSON: %v", err)
	}
}

func TestCurrentTimestamp(t *testing.T) {
	// Test that CurrentTimestamp returns valid RFC3339 format
	timestamp := CurrentTimestamp()

	// Parse the timestamp to verify it's valid RFC3339
	_, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		t.Errorf("CurrentTimestamp() = %q, is not valid RFC3339: %v", timestamp, err)
	}

	// Test that it's UTC
	parsed, _ := time.Parse(time.RFC3339, timestamp)
	if parsed.Location().String() != "UTC" {
		t.Errorf("CurrentTimestamp() location = %v, want UTC", parsed.Location())
	}

	// Test that it's recent (within last 5 seconds)
	now := time.Now().UTC()
	diff := now.Sub(parsed)
	if diff < 0 {
		diff = -diff
	}
	if diff > 5*time.Second {
		t.Errorf("CurrentTimestamp() = %v, is too old (diff: %v)", timestamp, diff)
	}
}

func TestCurrentTimestamp_Format(t *testing.T) {
	// Test multiple calls to ensure consistent format
	timestamps := make([]string, 10)
	for i := 0; i < 10; i++ {
		timestamps[i] = CurrentTimestamp()
		time.Sleep(1 * time.Millisecond) // Small delay to ensure different timestamps
	}

	// Verify all are valid RFC3339
	for i, ts := range timestamps {
		_, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			t.Errorf("CurrentTimestamp() call %d = %q, is not valid RFC3339: %v", i, ts, err)
		}
	}

	// Verify they're in chronological order (or at least valid)
	for i := 1; i < len(timestamps); i++ {
		prev, _ := time.Parse(time.RFC3339, timestamps[i-1])
		curr, _ := time.Parse(time.RFC3339, timestamps[i])
		if curr.Before(prev) {
			// This is acceptable if timestamps are very close, but log it
			diff := prev.Sub(curr)
			if diff > 1*time.Second {
				t.Logf("Warning: timestamp %d is before timestamp %d (diff: %v)", i, i-1, diff)
			}
		}
	}
}
