package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gov-dx-sandbox/audit-service/redis"

	"github.com/gov-dx-sandbox/audit-service/consumer"
	"github.com/gov-dx-sandbox/audit-service/handlers"
	"github.com/gov-dx-sandbox/audit-service/middleware"
	"github.com/gov-dx-sandbox/audit-service/services"
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

	// 2. Connect to Database using GORM
	dbConfig := &GormDatabaseConfig{
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

	db, err := NewGormDatabase(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				log.Printf("Error during database graceful shutdown: %v", err)
			}
		}
	}()

	// Initialize database tables using GORM AutoMigrate
	log.Println("Initializing database tables and views")
	if err := AutoMigrate(db); err != nil {
		log.Printf("Failed to initialize database: %v", err)
		log.Println("Continuing with database initialization failure - some operations may not work")
	}

	// 3. Connect to Redis
	redisCfg := &redis.Config{
		Addr:     getEnvOrDefault("REDIS_ADDR", "localhost:6379"),
		Password: getEnvOrDefault("REDIS_PASSWORD", ""),
		DB:       0,
	}

	redisClient, err := redis.NewClient(redisCfg)
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
	log.Println("Redis Stream consumer started.")

	// 6. Setup and Start the HTTP API Server
	apiHandler := handlers.NewAuditHandler(auditService)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/logs", apiHandler.HandleAuditLogs)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"audit-service"}`))
	})
	mux.HandleFunc("GET /version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"version":"1.0.0","service":"audit-service"}`))
	})

	// Add CORS middleware
	handler := middleware.NewCORSMiddleware()(mux)

	httpServer := &http.Server{
		Addr:    ":" + getEnvOrDefault("PORT", "3001"),
		Handler: handler,
	}

	// Start the server in a goroutine so it doesn't block
	go func() {
		log.Printf("HTTP API server starting on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server ListenAndServe error: %v", err)
		}
		log.Println("HTTP server stopped.")
	}()

	log.Println("Audit service started. Available endpoints:")
	log.Println("  GET  /health - Health check")
	log.Println("  GET  /version - Version info")
	log.Println("  GET  /api/logs - Retrieve audit logs")

	// 7. Wait for shutdown signal
	<-ctx.Done()
	log.Println("Audit service shutting down...")

	// 8. Gracefully shut down the HTTP server
	//    Give it configurable time to finish any active requests.
	shutdownTimeout := parseShutdownTimeout("SHUTDOWN_TIMEOUT", "5s")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server graceful shutdown failed: %v", err)
	}
}

// parseShutdownTimeout gets environment variable as duration or returns default value
func parseShutdownTimeout(key, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	if parsed, err := time.ParseDuration(defaultValue); err == nil {
		return parsed
	}
	return 5 * time.Second // fallback
}
