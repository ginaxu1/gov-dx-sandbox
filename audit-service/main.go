package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gov-dx-sandbox/audit-service/handlers"
	"github.com/gov-dx-sandbox/audit-service/services"
	_ "github.com/lib/pq"
)

func main() {
	// Database configuration
	dbHost := getEnv("CHOREO_OPENDIF_DATABASE1_HOSTNAME", "localhost")
	dbPort := getEnv("CHOREO_OPENDIF_DATABASE1_PORT", "5432")
	dbUser := getEnv("CHOREO_OPENDIF_DATABASE1_USERNAME", "user")
	dbPassword := getEnv("CHOREO_OPENDIF_DATABASE1_PASSWORD", "password")
	dbName := getEnv("CHOREO_OPENDIF_DATABASE1_DATABASENAME", "gov_dx_sandbox")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	// Server configuration
	port := getEnv("PORT", "3001")

	// Connect to database
	db, err := connectDB(dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize services
	auditService := services.NewAuditService(db)

	// Initialize handlers
	auditHandler := handlers.NewAuditHandler(auditService)

	// Setup routes
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Audit Service is healthy"))
	})

	// Audit endpoints
	mux.HandleFunc("/audit/events", auditHandler.GetAuditEvents)            // Admin Portal
	mux.HandleFunc("/audit/providers", auditHandler.GetProviderAuditEvents) // Provider Portal
	mux.HandleFunc("/audit/consumers", auditHandler.GetConsumerAuditEvents) // Consumer Portal

	// Audit log ingestion endpoint (for api-server-go to send audit logs)
	mux.HandleFunc("/audit/logs", auditHandler.CreateAuditLog) // Create audit log

	// Manual audit log creation endpoint (for testing purposes)
	mux.HandleFunc("/audit/create", auditHandler.CreateAuditLogManual) // Manual create audit log

	// Start server
	log.Printf("Audit Service starting on port %s", port)
	log.Printf("Database: %s:%s/%s", dbHost, dbPort, dbName)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// connectDB establishes a connection to the PostgreSQL database
func connectDB(host, port, user, password, dbname, sslmode string) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

// getEnv gets an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
