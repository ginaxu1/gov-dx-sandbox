package main

import (
	"os"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/configs"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/database"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/federator"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/provider"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/server"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/services"
)

func main() {
	logger.Init()
	configs.LoadConfig()

	// Initialize database connection
	dbConfig := database.NewDatabaseConfig()
	db, err := database.ConnectDB(dbConfig)
	if err != nil {
		logger.Log.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := database.GracefulShutdown(db); err != nil {
			logger.Log.Error("Error during database graceful shutdown", "error", err)
		}
	}()

	// Initialize database tables
	if err := database.InitDatabase(db); err != nil {
		logger.Log.Error("Failed to initialize database tables", "error", err)
		os.Exit(1)
	}

	// Initialize services
	schemaService := services.NewSchemaService(db)

	var providerHandler = provider.NewProviderHandler(configs.AppConfig.Providers)

	providerHandler.StartTokenRefreshProcess()

	var federationObject = federator.Initialize(providerHandler)

	server.RunServer(federationObject, schemaService)
}
