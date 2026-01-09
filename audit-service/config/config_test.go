package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnums_DefaultValues(t *testing.T) {
	// Test loading with non-existent file (should return defaults)
	enums, err := LoadEnums("/nonexistent/path/enums.yaml")
	if err != nil {
		t.Fatalf("Expected no error for non-existent file, got: %v", err)
	}

	if enums == nil {
		t.Fatal("Expected default enums, got nil")
	}

	// Verify default values are present
	if len(enums.EventTypes) == 0 {
		t.Error("Expected default event types")
	}
	if len(enums.EventActions) == 0 {
		t.Error("Expected default event actions")
	}
	if len(enums.ActorTypes) == 0 {
		t.Error("Expected default actor types")
	}
	if len(enums.TargetTypes) == 0 {
		t.Error("Expected default target types")
	}
}

func TestLoadEnums_ValidYAML(t *testing.T) {
	// Create a temporary YAML file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "enums.yaml")
	configContent := `enums:
  eventTypes:
    - POLICY_CHECK
    - MANAGEMENT_EVENT
  eventActions:
    - CREATE
    - READ
  actorTypes:
    - SERVICE
    - ADMIN
  targetTypes:
    - SERVICE
    - RESOURCE
`

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	enums, err := LoadEnums(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded values
	if len(enums.EventTypes) != 2 {
		t.Errorf("Expected 2 event types, got %d", len(enums.EventTypes))
	}
	if enums.EventTypes[0] != "POLICY_CHECK" {
		t.Errorf("Expected first event type to be POLICY_CHECK, got %s", enums.EventTypes[0])
	}
}

func TestAuditEnums_Validation(t *testing.T) {
	enums := &AuditEnums{
		EventTypes:   []string{"POLICY_CHECK", "MANAGEMENT_EVENT"},
		EventActions: []string{"CREATE", "READ"},
		ActorTypes:   []string{"SERVICE", "ADMIN"},
		TargetTypes:  []string{"SERVICE", "RESOURCE"},
	}
	// Initialize maps (normally done by LoadEnums)
	enums.InitializeMaps()

	// Test valid values
	if !enums.IsValidEventType("POLICY_CHECK") {
		t.Error("POLICY_CHECK should be valid")
	}
	if !enums.IsValidEventAction("CREATE") {
		t.Error("CREATE should be valid")
	}
	if !enums.IsValidActorType("SERVICE") {
		t.Error("SERVICE should be valid")
	}
	if !enums.IsValidTargetType("RESOURCE") {
		t.Error("RESOURCE should be valid")
	}

	// Test invalid values
	if enums.IsValidEventType("INVALID") {
		t.Error("INVALID should not be valid")
	}
	if enums.IsValidEventAction("INVALID") {
		t.Error("INVALID should not be valid")
	}
	if enums.IsValidActorType("INVALID") {
		t.Error("INVALID should not be valid")
	}
	if enums.IsValidTargetType("INVALID") {
		t.Error("INVALID should not be valid")
	}

	// Test empty values (should be allowed for nullable fields)
	if !enums.IsValidEventType("") {
		t.Error("Empty event type should be valid (nullable)")
	}
	if !enums.IsValidEventAction("") {
		t.Error("Empty event action should be valid (nullable)")
	}
}
