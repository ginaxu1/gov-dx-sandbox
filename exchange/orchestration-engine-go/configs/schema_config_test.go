package configs

import (
	"os"
	"testing"
)

func TestLoadSchemaConfig(t *testing.T) {
	// Test with default values
	config := LoadSchemaConfig()

	if config.DefaultVersion != "1.0.0" {
		t.Errorf("Expected default version 1.0.0, got %s", config.DefaultVersion)
	}

	if config.MaxVersions != 100 {
		t.Errorf("Expected max versions 100, got %d", config.MaxVersions)
	}

	if !config.CompatibilityCheck {
		t.Errorf("Expected compatibility check to be true, got %v", config.CompatibilityCheck)
	}

	if config.AutoActivate {
		t.Errorf("Expected auto activate to be false, got %v", config.AutoActivate)
	}
}

func TestLoadSchemaConfigWithEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("SCHEMA_VERSION_DEFAULT", "2.0.0")
	os.Setenv("SCHEMA_MAX_VERSIONS", "50")
	os.Setenv("SCHEMA_COMPATIBILITY_CHECK", "false")
	os.Setenv("SCHEMA_AUTO_ACTIVATE", "true")

	// Load config
	config := LoadSchemaConfig()

	if config.DefaultVersion != "2.0.0" {
		t.Errorf("Expected version 2.0.0, got %s", config.DefaultVersion)
	}

	if config.MaxVersions != 50 {
		t.Errorf("Expected max versions 50, got %d", config.MaxVersions)
	}

	if config.CompatibilityCheck {
		t.Errorf("Expected compatibility check to be false, got %v", config.CompatibilityCheck)
	}

	if !config.AutoActivate {
		t.Errorf("Expected auto activate to be true, got %v", config.AutoActivate)
	}

	// Clean up
	os.Unsetenv("SCHEMA_VERSION_DEFAULT")
	os.Unsetenv("SCHEMA_MAX_VERSIONS")
	os.Unsetenv("SCHEMA_COMPATIBILITY_CHECK")
	os.Unsetenv("SCHEMA_AUTO_ACTIVATE")
}

func TestSchemaConfigValidation(t *testing.T) {
	config := LoadSchemaConfig()

	err := config.Validate()
	if err != nil {
		t.Errorf("Expected validation to pass, got error: %v", err)
	}

	// Test invalid config
	config.MaxVersions = -1
	err = config.Validate()
	if err == nil {
		t.Errorf("Expected validation to fail for negative MaxVersions")
	}
}

func TestSchemaConfigMethods(t *testing.T) {
	config := LoadSchemaConfig()

	if !config.IsVersioningEnabled() {
		t.Errorf("Expected versioning to be enabled")
	}

	if !config.IsCompatibilityCheckEnabled() {
		t.Errorf("Expected compatibility check to be enabled")
	}

	if config.IsAutoActivateEnabled() {
		t.Errorf("Expected auto activate to be disabled")
	}

	if config.GetMaxVersions() != 100 {
		t.Errorf("Expected max versions 100, got %d", config.GetMaxVersions())
	}

	if config.GetVersionHistoryLimit() != 10 {
		t.Errorf("Expected version history limit 10, got %d", config.GetVersionHistoryLimit())
	}
}
