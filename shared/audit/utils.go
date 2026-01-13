package audit

import (
	"encoding/json"
	"log/slog"
	"time"
)

// MarshalMetadata safely marshals metadata to json.RawMessage.
// Returns empty JSON object "{}" on error to ensure valid JSON.
// Returns nil if metadata is nil.
func MarshalMetadata(metadata map[string]interface{}) json.RawMessage {
	if metadata == nil {
		return nil
	}
	bytes, err := json.Marshal(metadata)
	if err != nil {
		slog.Error("Failed to marshal metadata for audit", "error", err)
		return json.RawMessage("{}")
	}
	return json.RawMessage(bytes)
}

// CurrentTimestamp returns current UTC time in RFC3339 format.
// This provides a consistent timestamp format across all audit logs.
func CurrentTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
