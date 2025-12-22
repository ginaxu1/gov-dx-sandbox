package main

import (
	"log"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine/configs"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine/federator"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine/logger"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine/middleware"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine/provider"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine/server"
	auditclient "github.com/gov-dx-sandbox/shared/audit"
)

func main() {
	logger.Init()

	// Load configuration with proper error handling
	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize audit middleware
	// All configuration comes from config.json for consistency
	auditClient := auditclient.NewClient(config.AuditConfig.ServiceURL)
	auditclient.InitializeGlobalAudit(auditClient)

	// Initialize audit configuration (actorType, actorID)
	// Note: targetType is determined per API call, not from global config
	middleware.InitializeAuditConfig(
		config.AuditConfig.ActorType,
		config.AuditConfig.ActorID,
	)

	providerHandler := provider.NewProviderHandler(config.GetProviders())

	federationObject, err := federator.Initialize(config, providerHandler, nil)
	if err != nil {
		log.Fatalf("Failed to initialize federator: %v", err)
	}

	server.RunServer(federationObject)
}
