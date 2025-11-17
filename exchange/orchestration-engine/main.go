package main

import (
	"context"
	"log"
	"os"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine/configs"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine/federator"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine/logger"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine/middleware"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine/provider"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine/server"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/pkg/monitoring"
)

func main() {
	logger.Init()

	ctx := context.Background()
	shutdown, err := monitoring.Setup(ctx, monitoring.Config{
		ServiceName: "orchestration-engine",
	})
	if err != nil {
		log.Fatalf("Failed to initialize telemetry: %v", err)
	}
	defer func() {
		_ = shutdown(context.Background())
	}()

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
