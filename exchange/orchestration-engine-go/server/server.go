package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/auth"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/database"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/federator"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/handlers"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/graphql"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/services"
)

type Response struct {
	Message string `json:"message"`
}

const DefaultPort = "4000"

// RunServer starts a simple HTTP server with a health check endpoint.
func RunServer(f *federator.Federator) {
	mux := http.NewServeMux()

	// Initialize database connection
	dbConnectionString := getDatabaseConnectionString()
	schemaDB, err := database.NewSchemaDB(dbConnectionString)
	if err != nil {
		logger.Log.Error("Failed to connect to database", "error", err)
		// Continue without database for now
		schemaDB = nil
	}

	// Initialize schema mapping database connection
	var schemaMappingDB *database.SchemaMappingDB
	if schemaDB != nil {
		schemaMappingDB, err = database.NewSchemaMappingDB(dbConnectionString)
		if err != nil {
			logger.Log.Error("Failed to connect to schema mapping database", "error", err)
			schemaMappingDB = nil
		}
	}

	// Initialize schema service and handler
	var schemaService *services.SchemaService
	if schemaDB != nil {
		schemaService = services.NewSchemaService(schemaDB)
	} else {
		// Fallback to in-memory service if database is not available
		schemaService = nil
		logger.Log.Warn("Running without database - schema management disabled")
	}

	schemaHandler := handlers.NewSchemaHandler(schemaService)

	// Initialize schema mapping service and handler
	var schemaMappingService *services.SchemaMappingService
	var schemaMappingHandler *handlers.SchemaMappingHandler
	if schemaMappingDB != nil {
		schemaMappingService = services.NewSchemaMappingService(schemaMappingDB)
		compatibilityChecker := services.NewCompatibilityChecker()
		schemaMappingHandler = handlers.NewSchemaMappingHandler(schemaMappingService, compatibilityChecker)
	} else {
		logger.Log.Warn("Running without schema mapping database - schema mapping disabled")
	}

	// Set the schema service in the federator
	f.SchemaService = schemaService
	// /health route
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		resp := Response{Message: "OpenDIF Server is Healthy!"}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			return
		}
	})

	// Schema management routes
	mux.HandleFunc("/sdl", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			schemaHandler.GetActiveSchema(w, r)
		} else if r.Method == http.MethodPost {
			schemaHandler.CreateSchema(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/sdl/versions", schemaHandler.GetSchemas)
	mux.HandleFunc("/sdl/validate", schemaHandler.ValidateSDL)
	mux.HandleFunc("/sdl/check-compatibility", schemaHandler.CheckCompatibility)

	// Handle activation endpoint with proper path matching
	mux.HandleFunc("/sdl/versions/", func(w http.ResponseWriter, r *http.Request) {
		// Check if this is an activation request
		if strings.HasSuffix(r.URL.Path, "/activate") && r.Method == http.MethodPost {
			schemaHandler.ActivateSchema(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	// Schema Mapping Routes (Admin Portal)
	if schemaMappingHandler != nil {
		// Unified Schema Management
		mux.HandleFunc("/admin/unified-schemas", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				schemaMappingHandler.GetUnifiedSchemas(w, r)
			} else if r.Method == http.MethodPost {
				schemaMappingHandler.CreateUnifiedSchema(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})
		mux.HandleFunc("/admin/unified-schemas/latest", schemaMappingHandler.GetLatestUnifiedSchema)

		// Provider Schema Management
		mux.HandleFunc("/admin/provider-schemas", schemaMappingHandler.GetProviderSchemas)

		// Unified Schema Management with Field Mappings
		mux.HandleFunc("/admin/unified-schemas/", func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// Handle activation endpoint
			if strings.HasSuffix(path, "/activate") && r.Method == http.MethodPut {
				schemaMappingHandler.ActivateUnifiedSchema(w, r)
				return
			}

			// Handle field mappings
			if strings.Contains(path, "/mappings") {
				if strings.Contains(path, "/mappings/") && (r.Method == http.MethodPut || r.Method == http.MethodDelete) {
					// Update or delete specific mapping
					if r.Method == http.MethodPut {
						schemaMappingHandler.UpdateFieldMapping(w, r)
					} else {
						schemaMappingHandler.DeleteFieldMapping(w, r)
					}
				} else if r.Method == http.MethodPost {
					// Create new mapping
					schemaMappingHandler.CreateFieldMapping(w, r)
				} else if r.Method == http.MethodGet {
					// Get all mappings
					schemaMappingHandler.GetFieldMappings(w, r)
				} else {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
				return
			}

			// No matching route
			http.NotFound(w, r)
		})

		// Compatibility Check
		mux.HandleFunc("/admin/schemas/compatibility/check", schemaMappingHandler.CheckCompatibility)
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request body
		var req graphql.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
			return
		}

		// decode the token
		consumerAssertion, err := auth.GetConsumerJwtFromToken(r)

		if err != nil {
			logger.Log.Error("Failed to get consumer JWT from token", "error", err)
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Add panic recovery for federator calls
		var response graphql.Response
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Log.Error("Panic in FederateQuery", "panic", r, "stack", string(debug.Stack()))
					response = graphql.Response{
						Data: nil,
						Errors: []interface{}{
							map[string]interface{}{
								"message": fmt.Sprintf("Internal server error: %v", r),
							},
						},
					}
				}
			}()
			response = f.FederateQuery(req, consumerAssertion)
		}()

		w.WriteHeader(http.StatusOK)
		// Set content type to application/json

		w.Header().Set("Content-Type", "application/json")

		err = json.NewEncoder(w).Encode(response)

		if err != nil {
			logger.Log.Error("Failed to write response", "error", err)
			return
		}
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	// Convert port to string with colon prefix
	// e.g., "8000" -> ":8000"
	// This is needed for http.ListenAndServe
	// which expects the port in the format ":port"
	// If the port already has a colon, we don't add another one
	if port[0] != ':' {
		port = ":" + port
	}

	logger.Log.Info("Server is Listening", "port", port)

	if err := http.ListenAndServe(port, corsMiddleware(mux)); err != nil {
		logger.Log.Error("Failed to start server", "error", err)
	} else {
		logger.Log.Info("Server stopped")
	}
}

// corsMiddleware sets CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow all origins
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// Allow specific methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		// Allow specific headers
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight (OPTIONS) requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getDatabaseConnectionString returns the database connection string from environment variables
func getDatabaseConnectionString() string {
	// Check for Choreo environment variables first
	choreoHost := os.Getenv("CHOREO_DB_OE_HOSTNAME")
	choreoUser := os.Getenv("CHOREO_DB_OE_USERNAME")
	choreoPassword := os.Getenv("CHOREO_DB_OE_PASSWORD")
	choreoDB := os.Getenv("CHOREO_DB_OE_DATABASENAME")

	// Use Choreo variables if available, otherwise fall back to standard environment variables
	var host, port, user, password, dbname, sslmode string

	if choreoHost != "" {
		host = choreoHost
		port = getEnv("CHOREO_DB_OE_PORT", "5432")
		user = choreoUser
		password = choreoPassword
		dbname = choreoDB
		sslmode = "require" // Choreo typically requires SSL
	} else {
		host = getEnv("DB_HOST", "localhost")
		port = getEnv("DB_PORT", "5432")
		user = getEnv("DB_USER", "postgres")
		password = getEnv("DB_PASSWORD", "password")
		dbname = getEnv("DB_NAME", "orchestration_engine")
		sslmode = getEnv("DB_SSLMODE", "disable")
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
