package main

import (
	"database/sql"
	"encoding/json"
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
	dbHost := getEnv("CHOREO_DB_AUDIT_HOSTNAME", getEnv("DB_HOST", "localhost"))
	dbPort := getEnv("CHOREO_DB_AUDIT_PORT", getEnv("DB_PORT", "5432"))
	dbUser := getEnv("CHOREO_DB_AUDIT_USERNAME", getEnv("DB_USER", "user"))
	dbPassword := getEnv("CHOREO_DB_AUDIT_PASSWORD", getEnv("DB_PASSWORD", "password"))
	dbName := getEnv("CHOREO_DB_AUDIT_DATABASENAME", getEnv("DB_NAME", "gov_dx_sandbox"))
	dbSSLMode := getEnv("DB_SSLMODE", "require")

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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := map[string]string{
			"status": "healthy",
		}

		json.NewEncoder(w).Encode(response)
	})

	// API endpoints for log access
	mux.HandleFunc("/api/logs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			auditHandler.GetLogs(w, r)
		} else if r.Method == http.MethodPost {
			auditHandler.CreateLog(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Start server
	log.Printf("Audit Service starting on port %s", port)
	log.Printf("Database: %s:%s/%s", dbHost, dbPort, dbName)
	log.Printf("Database configuration - Choreo Host: %s, Fallback Host: %s",
		os.Getenv("CHOREO_DB_AUDIT_HOSTNAME"), os.Getenv("DB_HOST"))

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
