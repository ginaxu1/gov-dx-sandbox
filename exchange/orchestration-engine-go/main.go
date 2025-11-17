package main

import (
	"context"
	"log"
	"os"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/configs"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/federator"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/middleware"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/provider"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/server"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/telemetry"
)

func main() {
	logger.Init()

	ctx := context.Background()
	shutdown, err := telemetry.Init(ctx, "orchestration-engine")
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

	var providerHandler = provider.NewProviderHandler(config.GetProviders())

	var federationObject = federator.Initialize(config, providerHandler, nil)

	server.RunServer(federationObject)
}
