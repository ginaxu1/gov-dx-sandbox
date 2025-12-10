package utils

import (
	"testing"
)

func TestGenerateTraceID(t *testing.T) {
	// Generate multiple IDs
	id1 := GenerateTraceID()
	id2 := GenerateTraceID()

	// Check length (16 bytes = 32 hex chars)
	if len(id1) != 32 {
		t.Errorf("Expected trace ID length 32, got %d", len(id1))
	}

	// Check uniqueness
	if id1 == id2 {
		t.Error("Generated identical trace IDs")
	}
}
