package main

import (
	"log"
	"os"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/configs"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/federator"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/provider"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/server"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/services"
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

	// Initialize API Server client if URL is configured
	var apiServerClient *services.APIServerClient
	apiServerURL := os.Getenv("API_SERVER_URL")
	if apiServerURL != "" {
		apiKey := os.Getenv("API_SERVER_API_KEY")
		apiServerClient = services.NewAPIServerClient(apiServerURL, apiKey)
		logger.Log.Info("API Server client initialized", "url", apiServerURL)
	} else {
		logger.Log.Info("API Server URL not configured, schema loading from API Server disabled")
	}

	var federationObject = federator.Initialize(providerHandler, nil, apiServerClient)

	server.RunServer(federationObject, apiServerClient)
}
