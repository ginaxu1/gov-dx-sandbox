package configs

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/auth"
	"github.com/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/graphql"
	"github.com/gov-dx-sandbox/exchange/orchestration-engine-go/provider"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

// Config is the top-level struct to hold all application configuration.
// The json tags (`json:"..."`) are essential for correctly mapping the
// keys from the config.json file to the fields in this struct.
type Config struct {
	CeURL       string                `json:"ceUrl"`
	PdpURL      string                `json:"pdpUrl"`
	Providers   []*ProviderConfig     `json:"providers"`
	ArgMappings []*graphql.ArgMapping `json:"argMappings"`
	Environment string                `json:"environment,omitempty"`
	Server      ServerConfig          `json:"server,omitempty"`
	Log         LogConfig             `json:"log,omitempty"`
	Services    ServicesConfig        `json:"services,omitempty"`
	PdpConfig   PdpConfig             `json:"pdpConfig,omitempty"`
	CeConfig    CeConfig              `json:"ceConfig,omitempty"`
	Schema      *string               `json:"schema,omitempty"`
	Sdl         *string               `json:"sdl,omitempty"`
	ArgMapping  []*graphql.ArgMapping `json:"argMapping,omitempty"`
}

// ProviderConfig represents a provider configuration
type ProviderConfig struct {
	ProviderKey string           `json:"providerKey"`
	ProviderURL string           `json:"providerUrl"`
	Auth        *auth.AuthConfig `json:"auth,omitempty"`
	SchemaID    string           `json:"schemaId"`
}

// ServerConfig holds the server-specific configuration.
type ServerConfig struct {
	Port string `json:"port"`
}

// LogConfig holds the logging configuration.
type LogConfig struct {
	Level string `json:"level"`
}

// ServicesConfig holds URLs for external services.
type ServicesConfig struct {
	PdpURL string `json:"pdp_url"`
}

// PdpConfig holds PDP service configuration
type PdpConfig struct {
	ClientURL string `json:"clientUrl"`
}

// CeConfig holds Consent Engine configuration
type CeConfig struct {
	ClientURL string `json:"clientUrl"`
}

// LoadConfigFromBytes unmarshals JSON into config (pure function, testable)
func LoadConfigFromBytes(data []byte) (*Config, error) {
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config JSON: %w", err)
	}

	// Derived config logic
	if config.PdpConfig.ClientURL == "" && config.PdpURL != "" {
		config.PdpConfig.ClientURL = config.PdpURL
	}
	if config.CeConfig.ClientURL == "" && config.CeURL != "" {
		config.CeConfig.ClientURL = config.CeURL
	}
	if config.ArgMapping == nil {
		config.ArgMapping = config.ArgMappings
	}

	return &config, nil
}

// LoadConfigFile reads a file and uses LoadConfigFromBytes (IO separated)
func LoadConfigFile(path string) (*Config, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file %s: %w", path, err)
	}
	return LoadConfigFromBytes(bytes)
}

// LoadConfig is usually called from main()
func LoadConfig() (*Config, error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "./config.json"
	}

	cfg, err := LoadConfigFile(path)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// GetProviders converts ProviderConfig slice to provider.Provider slice
func (c *Config) GetProviders() []*provider.Provider {
	providers := make([]*provider.Provider, len(c.Providers))
	for i, pConfig := range c.Providers {
		providers[i] = provider.NewProvider(
			pConfig.ProviderKey,
			pConfig.ProviderURL,
			pConfig.SchemaID,
			pConfig.Auth,
		)
	}
	return providers
}

// GetSchemaDocument parses the schema string and returns an AST document
func (c *Config) GetSchemaDocument() (*ast.Document, error) {
	if c.Schema == nil || *c.Schema == "" {
		return nil, fmt.Errorf("no schema defined in configuration")
	}

	src := source.NewSource(&source.Source{
		Body: []byte(*c.Schema),
		Name: "ConfigSchema",
	})

	schema, err := parser.Parse(parser.ParseParams{Source: src})
	if err != nil {
		return nil, fmt.Errorf("error parsing schema: %w", err)
	}

	return schema, nil
}
