package main

import (
	"log"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/configs"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/server"
)

func main() {
	// Load configuration
	config := configs.LoadSchemaConfig()

	// Connect to database
	db, err := server.ConnectDatabase(
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
		config.Database.DBName,
		config.Database.SSLMode,
	)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create schema server
	schemaServer := server.NewSchemaServer(db)

	// Start the server
	if err := schemaServer.Start(config.Server.Port); err != nil {
		log.Fatalf("Failed to start schema server: %v", err)
	}
}
