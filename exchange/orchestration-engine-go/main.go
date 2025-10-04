package main

import (
	"os"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/configs"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/database"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/server"
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

	// Start the schema management server
	schemaServer := server.NewSchemaServer(db)
	if err := schemaServer.Initialize(); err != nil {
		logger.Log.Error("Failed to initialize schema server", "error", err)
		os.Exit(1)
	}

	// Start the server on port 8081
	if err := schemaServer.Start(":8081"); err != nil {
		logger.Log.Error("Failed to start schema server", "error", err)
		os.Exit(1)
	}
}
