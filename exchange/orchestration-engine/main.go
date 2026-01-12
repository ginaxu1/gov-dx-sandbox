package main

import (
	"log"
	"os"

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
	// Environment variable takes precedence over config.json for flexibility
	auditServiceURL := os.Getenv("AUDIT_SERVICE_URL")
	if auditServiceURL == "" {
		auditServiceURL = config.AuditConfig.ServiceURL
	}
	auditClient := auditclient.NewClient(auditServiceURL)
	auditclient.InitializeGlobalAudit(auditClient)

	// Initialize audit configuration (actorType, actorID, targetType)
	middleware.InitializeAuditConfig(
		config.AuditConfig.ActorType,
		config.AuditConfig.ActorID,
		config.AuditConfig.TargetType,
	)

	providerHandler := provider.NewProviderHandler(config.GetProviders())

	federationObject := federator.Initialize(config, providerHandler, nil)

	server.RunServer(federationObject)
}
