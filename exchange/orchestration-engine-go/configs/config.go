package configs

import (
	"encoding/json"
	"os"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/consent"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/graphql"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/policy"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/provider"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

// Options defines the configuration options for the federator.
type Options struct {
	Providers []*provider.Provider `json:"providers,omitempty"`
}

// Cfg defines the configuration structure for the application.
type Cfg struct {
	*Options
	Environment string `json:"environment,omitempty"`
	*graphql.MappingAST
	Schema *ast.Document
	Sdl    []byte
	*policy.PdpConfig
	*consent.CeConfig
	*SchemaConfig
}

const ConfigFilePath = "./config.json"
const SDLFilePath = "./schema.graphql"

// AppConfig is a global variable to hold the application configuration.
var AppConfig *Cfg

func LoadSdlSchema(AppConfig *Cfg) {
	schema, err := os.ReadFile(SDLFilePath)

	if err != nil {
		panic(err)
	}

	// Wrap as a GraphQL source
	src := source.NewSource(&source.Source{
		Body: schema,
		Name: "SchemaSDL",
	})

	doc, err := parser.Parse(parser.ParseParams{Source: src})

	if err != nil {
		panic(err)
	}

	AppConfig.Sdl = schema
	AppConfig.Schema = doc
}

// LoadConfig reads the configuration from the config.json file and unmarshal it into the AppConfig variable.
func LoadConfig() {
	if AppConfig != nil {
		return
	}

	AppConfig = &Cfg{}

	file, err := os.ReadFile(ConfigFilePath)

	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(file, AppConfig)

	if err != nil {
		panic(err)
	}

	// Load schema configuration
	AppConfig.SchemaConfig = LoadSchemaConfig()

	LoadSdlSchema(AppConfig)
}
