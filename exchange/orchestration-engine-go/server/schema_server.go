package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/database"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/handlers"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/services"
	_ "github.com/lib/pq"
)

// SchemaServer represents the schema management server
type SchemaServer struct {
	db            *sql.DB
	schemaService *services.SchemaServiceImpl
	handlers      *handlers.SchemaHandlers
}

// NewSchemaServer creates a new schema server instance
func NewSchemaServer(db *sql.DB) *SchemaServer {
	schemaDB := database.NewSchemaDB(db)
	schemaService := services.NewSchemaService(schemaDB)
	schemaHandlers := handlers.NewSchemaHandlers(schemaService)

	return &SchemaServer{
		db:            db,
		schemaService: schemaService,
		handlers:      schemaHandlers,
	}
}

// Initialize sets up the database and creates necessary tables
func (s *SchemaServer) Initialize() error {
	schemaDB := database.NewSchemaDB(s.db)

	// Create unified_schemas table
	if err := schemaDB.CreateSchemaTable(); err != nil {
		return fmt.Errorf("failed to create schema table: %w", err)
	}

	// Create schema_versions table
	if err := schemaDB.CreateSchemaVersionsTable(); err != nil {
		return fmt.Errorf("failed to create schema versions table: %w", err)
	}

	return nil
}

// SetupRoutes configures the HTTP routes for schema management
func (s *SchemaServer) SetupRoutes(r *gin.Engine) {
	// Schema management API routes
	api := r.Group("/api")
	{
		// Schema CRUD operations
		api.POST("/schemas", s.handlers.CreateSchema)
		api.GET("/schemas", s.handlers.GetAllSchemas)
		api.GET("/schemas/active", s.handlers.GetActiveSchema)
		api.GET("/schemas/versions", s.handlers.GetSchemaVersions)
		api.GET("/schemas/:version", s.handlers.GetSchema)
		api.PUT("/schemas/:version", s.handlers.UpdateSchema)
		api.DELETE("/schemas/:version", s.handlers.DeleteSchema)

		// Schema status management
		api.PUT("/schemas/:version/status", s.handlers.UpdateSchemaStatus)
		api.POST("/schemas/:version/activate", s.handlers.ActivateVersion)
		api.POST("/schemas/:version/deactivate", s.handlers.DeactivateVersion)

		// Schema validation and compatibility
		api.POST("/schemas/validate", s.handlers.ValidateSDL)
		api.POST("/schemas/check-compatibility", s.handlers.CheckCompatibility)

		// Schema version history
		api.GET("/schemas/versions/history", s.handlers.GetSchemaVersionsHistory)
		api.GET("/schemas/:version/history", s.handlers.GetSchemaVersionsByVersion)

		// GraphQL query execution
		api.POST("/graphql", s.handlers.ExecuteQuery)
	}
}

// Start starts the schema management server
func (s *SchemaServer) Start(port string) error {
	// Initialize database
	if err := s.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Setup Gin router
	r := gin.Default()

	// Add CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Setup routes
	s.SetupRoutes(r)

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Start server
	log.Printf("Schema management server starting on port %s", port)
	return r.Run(":" + port)
}

// GetSchemaService returns the schema service instance
func (s *SchemaServer) GetSchemaService() *services.SchemaServiceImpl {
	return s.schemaService
}

// ConnectDatabase establishes a connection to the PostgreSQL database
func ConnectDatabase(host, port, user, password, dbname, sslmode string) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
