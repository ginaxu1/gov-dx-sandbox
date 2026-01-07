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
	auditclient "github.com/gov-dx-sandbox/shared/audit"
)

// auditClientAdapter adapts shared/audit.Client to middleware.AuditClient interface
type auditClientAdapter struct {
	client *auditclient.Client
}

func (a *auditClientAdapter) LogEvent(ctx context.Context, event *middleware.AuditLogRequest) {
	// Convert middleware DTO to shared audit DTO
	auditRequest := &auditclient.AuditLogRequest{
		TraceID:            event.TraceID,
		Timestamp:          event.Timestamp,
		EventType:          event.EventType,
		EventAction:        event.EventAction,
		Status:             event.Status,
		ActorType:          event.ActorType,
		ActorID:            event.ActorID,
		TargetType:         event.TargetType,
		TargetID:           event.TargetID,
		RequestMetadata:    event.RequestMetadata,
		ResponseMetadata:   event.ResponseMetadata,
		AdditionalMetadata: event.AdditionalMetadata,
	}
	a.client.LogEvent(ctx, auditRequest)
}

func (a *auditClientAdapter) IsEnabled() bool {
	return a.client.IsEnabled()
}

func main() {
	logger.Init()

	// Load configuration with proper error handling
	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize audit middleware
	auditServiceURL := os.Getenv("CHOREO_AUDIT_CONNECTION_SERVICEURL")
	sharedAuditClient := auditclient.NewClient(auditServiceURL)
	auditClient := &auditClientAdapter{client: sharedAuditClient}
	middleware.NewAuditMiddleware(auditClient)

	providerHandler := provider.NewProviderHandler(config.GetProviders())

	federationObject := federator.Initialize(config, providerHandler, nil)

	server.RunServer(federationObject)
}
