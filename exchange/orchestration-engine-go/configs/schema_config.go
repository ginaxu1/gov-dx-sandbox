package configs

import (
	"os"
	"strconv"
	"strings"
)

// SchemaConfig defines the configuration structure for schema management
type SchemaConfig struct {
	DefaultVersion      string `json:"default_version"`
	MaxVersions         int    `json:"max_versions"`
	CompatibilityCheck  bool   `json:"compatibility_check"`
	AutoActivate        bool   `json:"auto_activate"`
	BackupRetentionDays int    `json:"backup_retention_days"`
	EnableVersioning    bool   `json:"enable_versioning"`
	EnableContractTests bool   `json:"enable_contract_tests"`
	SchemaCacheSize     int    `json:"schema_cache_size"`
	VersionHistoryLimit int    `json:"version_history_limit"`
}

// DefaultSchemaConfig returns the default schema configuration
func DefaultSchemaConfig() *SchemaConfig {
	return &SchemaConfig{
		DefaultVersion:      "1.0.0",
		MaxVersions:         100,
		CompatibilityCheck:  true,
		AutoActivate:        false,
		BackupRetentionDays: 30,
		EnableVersioning:    true,
		EnableContractTests: true,
		SchemaCacheSize:     50,
		VersionHistoryLimit: 10,
	}
}

// LoadSchemaConfig loads schema configuration from environment variables and config file
func LoadSchemaConfig() *SchemaConfig {
	config := DefaultSchemaConfig()

	// Load from environment variables
	if val := os.Getenv("SCHEMA_VERSION_DEFAULT"); val != "" {
		config.DefaultVersion = val
	}

	if val := os.Getenv("SCHEMA_MAX_VERSIONS"); val != "" {
		if maxVersions, err := strconv.Atoi(val); err == nil && maxVersions > 0 {
			config.MaxVersions = maxVersions
		}
	}

	if val := os.Getenv("SCHEMA_COMPATIBILITY_CHECK"); val != "" {
		config.CompatibilityCheck = strings.ToLower(val) == "true"
	}

	if val := os.Getenv("SCHEMA_AUTO_ACTIVATE"); val != "" {
		config.AutoActivate = strings.ToLower(val) == "true"
	}

	if val := os.Getenv("SCHEMA_BACKUP_RETENTION_DAYS"); val != "" {
		if retentionDays, err := strconv.Atoi(val); err == nil && retentionDays > 0 {
			config.BackupRetentionDays = retentionDays
		}
	}

	if val := os.Getenv("SCHEMA_ENABLE_VERSIONING"); val != "" {
		config.EnableVersioning = strings.ToLower(val) == "true"
	}

	if val := os.Getenv("SCHEMA_ENABLE_CONTRACT_TESTS"); val != "" {
		config.EnableContractTests = strings.ToLower(val) == "true"
	}

	if val := os.Getenv("SCHEMA_CACHE_SIZE"); val != "" {
		if cacheSize, err := strconv.Atoi(val); err == nil && cacheSize > 0 {
			config.SchemaCacheSize = cacheSize
		}
	}

	if val := os.Getenv("SCHEMA_VERSION_HISTORY_LIMIT"); val != "" {
		if historyLimit, err := strconv.Atoi(val); err == nil && historyLimit > 0 {
			config.VersionHistoryLimit = historyLimit
		}
	}

	return config
}

// Validate validates the schema configuration
func (sc *SchemaConfig) Validate() error {
	if sc.DefaultVersion == "" {
		return &ConfigError{Field: "DefaultVersion", Message: "default version cannot be empty"}
	}

	if sc.MaxVersions <= 0 {
		return &ConfigError{Field: "MaxVersions", Message: "max versions must be greater than 0"}
	}

	if sc.BackupRetentionDays < 0 {
		return &ConfigError{Field: "BackupRetentionDays", Message: "backup retention days cannot be negative"}
	}

	if sc.SchemaCacheSize <= 0 {
		return &ConfigError{Field: "SchemaCacheSize", Message: "schema cache size must be greater than 0"}
	}

	if sc.VersionHistoryLimit <= 0 {
		return &ConfigError{Field: "VersionHistoryLimit", Message: "version history limit must be greater than 0"}
	}

	return nil
}

// IsVersioningEnabled returns true if schema versioning is enabled
func (sc *SchemaConfig) IsVersioningEnabled() bool {
	return sc.EnableVersioning
}

// IsCompatibilityCheckEnabled returns true if compatibility checking is enabled
func (sc *SchemaConfig) IsCompatibilityCheckEnabled() bool {
	return sc.CompatibilityCheck
}

// IsAutoActivateEnabled returns true if auto-activation is enabled
func (sc *SchemaConfig) IsAutoActivateEnabled() bool {
	return sc.AutoActivate
}

// IsContractTestsEnabled returns true if contract tests are enabled
func (sc *SchemaConfig) IsContractTestsEnabled() bool {
	return sc.EnableContractTests
}

// GetMaxVersions returns the maximum number of schema versions allowed
func (sc *SchemaConfig) GetMaxVersions() int {
	return sc.MaxVersions
}

// GetDefaultVersion returns the default schema version
func (sc *SchemaConfig) GetDefaultVersion() string {
	return sc.DefaultVersion
}

// GetBackupRetentionDays returns the number of days to retain backups
func (sc *SchemaConfig) GetBackupRetentionDays() int {
	return sc.BackupRetentionDays
}

// GetSchemaCacheSize returns the maximum number of schemas to cache in memory
func (sc *SchemaConfig) GetSchemaCacheSize() int {
	return sc.SchemaCacheSize
}

// GetVersionHistoryLimit returns the maximum number of versions to keep in history
func (sc *SchemaConfig) GetVersionHistoryLimit() int {
	return sc.VersionHistoryLimit
}

// ConfigError represents a configuration validation error
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return "configuration error in " + e.Field + ": " + e.Message
}

// SchemaConfigManager manages schema configuration
type SchemaConfigManager struct {
	config *SchemaConfig
}

// NewSchemaConfigManager creates a new schema configuration manager
func NewSchemaConfigManager() *SchemaConfigManager {
	return &SchemaConfigManager{
		config: LoadSchemaConfig(),
	}
}

// GetConfig returns the current schema configuration
func (scm *SchemaConfigManager) GetConfig() *SchemaConfig {
	return scm.config
}

// ReloadConfig reloads the schema configuration from environment variables
func (scm *SchemaConfigManager) ReloadConfig() error {
	scm.config = LoadSchemaConfig()
	return scm.config.Validate()
}

// UpdateConfig updates the schema configuration
func (scm *SchemaConfigManager) UpdateConfig(newConfig *SchemaConfig) error {
	if err := newConfig.Validate(); err != nil {
		return err
	}
	scm.config = newConfig
	return nil
}

// GetEnvironmentVariables returns a map of all schema-related environment variables
func (scm *SchemaConfigManager) GetEnvironmentVariables() map[string]string {
	return map[string]string{
		"SCHEMA_VERSION_DEFAULT":       scm.config.DefaultVersion,
		"SCHEMA_MAX_VERSIONS":          strconv.Itoa(scm.config.MaxVersions),
		"SCHEMA_COMPATIBILITY_CHECK":   strconv.FormatBool(scm.config.CompatibilityCheck),
		"SCHEMA_AUTO_ACTIVATE":         strconv.FormatBool(scm.config.AutoActivate),
		"SCHEMA_BACKUP_RETENTION_DAYS": strconv.Itoa(scm.config.BackupRetentionDays),
		"SCHEMA_ENABLE_VERSIONING":     strconv.FormatBool(scm.config.EnableVersioning),
		"SCHEMA_ENABLE_CONTRACT_TESTS": strconv.FormatBool(scm.config.EnableContractTests),
		"SCHEMA_CACHE_SIZE":            strconv.Itoa(scm.config.SchemaCacheSize),
		"SCHEMA_VERSION_HISTORY_LIMIT": strconv.Itoa(scm.config.VersionHistoryLimit),
	}
}
