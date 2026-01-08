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
	"github.com/gov-dx-sandbox/exchange/shared/monitoring"
)

func main() {
	logger.Init()

	// Initialize monitoring/observability (optional - can be disabled via ENABLE_OBSERVABILITY=false)
	// Services will continue to function normally even if observability is disabled
	if monitoring.IsObservabilityEnabled() {
		monitoringConfig := monitoring.DefaultConfig("orchestration-engine")
		if err := monitoring.Initialize(monitoringConfig); err != nil {
			logger.Log.Warn("Failed to initialize monitoring (service will continue)", "error", err)
		}
	} else {
		logger.Log.Info("Observability disabled via environment variable")
	}

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
