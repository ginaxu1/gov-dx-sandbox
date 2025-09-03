package configs

import (
	"encoding/json"
	"os"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/federator"
)

// Cfg defines the configuration structure for the application.
type Cfg struct {
	*federator.Options
}

const ConfigFilePath = "./config.json"

// AppConfig is a global variable to hold the application configuration.
var AppConfig *Cfg

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
}

func IsProviderExists(providerKey string) bool {
	if AppConfig == nil || AppConfig.Options == nil {
		return false
	}

	for _, p := range AppConfig.Options.Providers {
		if p.ServiceKey == providerKey {
			return true
		}
	}
	return false
}
