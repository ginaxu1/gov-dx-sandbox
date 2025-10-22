package main

import (
	"log"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/configs"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/federator"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/provider"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/server"
)

func main() {
	logger.Init()

	// Load configuration with proper error handling
	_, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Check if AppConfig is properly initialized
	if configs.AppConfig == nil {
		log.Fatal("Configuration not properly initialized")
	}

	var providerHandler = provider.NewProviderHandler(configs.AppConfig.GetProviders())

	var federationObject = federator.Initialize(providerHandler, nil)

	server.RunServer(federationObject)
}
