package configs

import (
	"encoding/json"
	"os"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/consent"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/federator"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/graphql"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/policy"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

// Cfg defines the configuration structure for the application.
type Cfg struct {
	Options    *federator.Options     `json:"Options"`
	MappingAST *graphql.MappingAST    `json:"MappingAST"`
	ArgMapping map[string]interface{} `json:"argMapping"`
	Schema     *ast.Document
	PdpConfig  *policy.PdpConfig `json:"PdpConfig"`
	CeConfig   *consent.CeConfig `json:"CeConfig"`
}

const ConfigFilePath = "./config.json"
const SDLFilePath = "./schema.graphql"

// AppConfig is a global variable to hold the application configuration.
var AppConfig *Cfg

func LoadSdlSchema() *ast.Document {
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

	return doc
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

	AppConfig.Schema = LoadSdlSchema()
}
