package main

import (
	"log"
	"os"

	"github.com/gov-dx-sandbox/exchange/orchestration-engine-go/configs"
	"github.com/gov-dx-sandbox/exchange/orchestration-engine-go/federator"
	"github.com/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
	"github.com/gov-dx-sandbox/exchange/orchestration-engine-go/middleware"
	"github.com/gov-dx-sandbox/exchange/orchestration-engine-go/provider"
	"github.com/gov-dx-sandbox/exchange/orchestration-engine-go/server"
)

func main() {
	logger.Init()

	// Load configuration with proper error handling
	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize audit middleware
	auditServiceURL := os.Getenv("CHOREO_AUDIT_CONNECTION_SERVICEURL")
	middleware.NewAuditMiddleware(auditServiceURL)

	providerHandler := provider.NewProviderHandler(config.GetProviders())

	federationObject := federator.Initialize(config, providerHandler, nil)

	server.RunServer(federationObject)
}
