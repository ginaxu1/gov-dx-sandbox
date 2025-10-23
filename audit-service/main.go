package main

import (
	"context"
	"fmt"
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

	// 2. Connect to Database with explicit configuration
	dbConfig := &DatabaseConfig{
		Host:            getEnvOrDefault("CHOREO_DB_AUDIT_HOSTNAME", "localhost"),
		Port:            getEnvOrDefault("CHOREO_DB_AUDIT_PORT", "5432"),
		Username:        getEnvOrDefault("CHOREO_DB_AUDIT_USERNAME", "postgres"),
		Password:        getEnvOrDefault("CHOREO_DB_AUDIT_PASSWORD", "password"),
		Database:        getEnvOrDefault("CHOREO_DB_AUDIT_DATABASENAME", "gov_dx_sandbox"),
		SSLMode:         getEnvOrDefault("DB_SSLMODE", "require"),
		MaxOpenConns:    parseIntOrDefault("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    parseIntOrDefault("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: parseDurationOrDefault("DB_CONN_MAX_LIFETIME", "1h"),
		ConnMaxIdleTime: parseDurationOrDefault("DB_CONN_MAX_IDLE_TIME", "30m"),
		QueryTimeout:    parseDurationOrDefault("DB_QUERY_TIMEOUT", "30s"),
		ConnectTimeout:  parseDurationOrDefault("DB_CONNECT_TIMEOUT", "10s"),
		RetryAttempts:   parseIntOrDefault("DB_RETRY_ATTEMPTS", 10),
		RetryDelay:      parseDurationOrDefault("DB_RETRY_DELAY", "2s"),
	}

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

	// Generate a unique consumer name, e.g., from the hostname.
	hostname, err := os.Hostname()
	if err != nil {
		hostname = fmt.Sprintf("pid-%d", os.Getpid())
	}
	consumerName := "audit-service-" + hostname
	log.Printf("Initializing consumer with name: %s", consumerName)

	// Pass the new name to the constructor
	streamConsumer, err := consumer.NewStreamConsumer(redisClient, processor, consumerName)
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
