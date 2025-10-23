package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gov-dx-sandbox/audit-service/consumer"
	"github.com/gov-dx-sandbox/audit-service/services"
	redisclient "github.com/gov-dx-sandbox/shared/redis"
)

// main is the entry point for the audit service.
func main() {
	// 1. Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutdown signal received, canceling context...")
		cancel()
	}()

	// 2. Connect to Database (Your existing logic)
	dbConfig := NewDatabaseConfig()
	log.Printf("Connecting to database: %s:%s/%s", dbConfig.Host, dbConfig.Port, dbConfig.Database)

	db, err := ConnectDB(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := GracefulShutdown(db); err != nil {
			log.Printf("Error during database graceful shutdown: %v", err)
		}
	}()

	// Initialize database tables
	log.Println("Initializing database tables and views")
	if err := InitDatabase(db); err != nil {
		log.Printf("Failed to initialize database: %v", err)
		log.Println("Continuing with database initialization failure - some operations may not work")
	}

	// 3. Connect to Redis
	redisCfg := &redisclient.Config{
		Addr:     getEnvOrDefault("REDIS_ADDR", "localhost:6379"),
		Password: getEnvOrDefault("REDIS_PASSWORD", ""),
		DB:       0,
	}

	redisClient, err := redisclient.NewClient(redisCfg)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// 4. Create the processor and consumer
	// The processor (DatabaseEventProcessor) links our consumer to our database.
	auditService := services.NewAuditService(db)
	processor := consumer.NewDatabaseEventProcessor(auditService)
	streamConsumer, err := consumer.NewStreamConsumer(redisClient, processor)
	if err != nil {
		log.Fatalf("Failed to create stream consumer: %v", err)
	}

	// 5. Start the consumer in a new goroutine
	go streamConsumer.Start(ctx)

	log.Println("Audit service started. Waiting for events...")

	// 6. Wait for shutdown signal
	<-ctx.Done()
	log.Println("Audit service shutting down.")
}
