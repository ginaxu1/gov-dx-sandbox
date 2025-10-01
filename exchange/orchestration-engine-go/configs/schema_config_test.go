package configs

import (
	"os"
	"testing"
)

func TestLoadSchemaConfig(t *testing.T) {
	// Test with default values
	config := LoadSchemaConfig()

	if config.Schema.DefaultVersion != "latest" {
		t.Errorf("Expected default version latest, got %s", config.Schema.DefaultVersion)
	}

	if config.Schema.MaxVersions != 10 {
		t.Errorf("Expected max versions 10, got %d", config.Schema.MaxVersions)
	}

	if !config.Schema.CompatibilityCheck {
		t.Errorf("Expected compatibility check to be true, got %v", config.Schema.CompatibilityCheck)
	}

	if config.Schema.AutoActivate {
		t.Errorf("Expected auto activate to be false, got %v", config.Schema.AutoActivate)
	}
}

func TestLoadSchemaConfigWithEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("SCHEMA_DEFAULT_VERSION", "2.0.0")
	os.Setenv("SCHEMA_MAX_VERSIONS", "50")
	os.Setenv("SCHEMA_COMPATIBILITY_CHECK", "false")
	os.Setenv("SCHEMA_AUTO_ACTIVATE", "true")

	// Load config
	config := LoadSchemaConfig()

	if config.Schema.DefaultVersion != "2.0.0" {
		t.Errorf("Expected version 2.0.0, got %s", config.Schema.DefaultVersion)
	}

	if config.Schema.MaxVersions != 50 {
		t.Errorf("Expected max versions 50, got %d", config.Schema.MaxVersions)
	}

	if config.Schema.CompatibilityCheck {
		t.Errorf("Expected compatibility check to be false, got %v", config.Schema.CompatibilityCheck)
	}

	if !config.Schema.AutoActivate {
		t.Errorf("Expected auto activate to be true, got %v", config.Schema.AutoActivate)
	}

	// Clean up
	os.Unsetenv("SCHEMA_DEFAULT_VERSION")
	os.Unsetenv("SCHEMA_MAX_VERSIONS")
	os.Unsetenv("SCHEMA_COMPATIBILITY_CHECK")
	os.Unsetenv("SCHEMA_AUTO_ACTIVATE")
}
