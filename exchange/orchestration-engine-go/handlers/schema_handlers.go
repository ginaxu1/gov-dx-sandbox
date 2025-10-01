package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/models"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/services"
)

type SchemaHandlers struct {
	schemaService *services.SchemaServiceImpl
}

// NewSchemaHandlers creates a new schema handlers instance
func NewSchemaHandlers(schemaService *services.SchemaServiceImpl) *SchemaHandlers {
	return &SchemaHandlers{
		schemaService: schemaService,
	}
}

// CreateSchema handles POST /api/schemas
func (h *SchemaHandlers) CreateSchema(c *gin.Context) {
	var req models.CreateSchemaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schema, err := h.schemaService.CreateSchema(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, schema)
}

// GetSchema handles GET /api/schemas/:version
func (h *SchemaHandlers) GetSchema(c *gin.Context) {
	version := c.Param("version")

	schema, err := h.schemaService.GetSchemaByVersion(version)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Schema not found"})
		return
	}

	c.JSON(http.StatusOK, schema)
}

// GetActiveSchema handles GET /api/schemas/active
func (h *SchemaHandlers) GetActiveSchema(c *gin.Context) {
	schema, err := h.schemaService.GetActiveSchema()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active schema found"})
		return
	}

	c.JSON(http.StatusOK, schema)
}

// GetAllSchemas handles GET /api/schemas
func (h *SchemaHandlers) GetAllSchemas(c *gin.Context) {
	schemas, err := h.schemaService.GetAllSchemas()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"schemas": schemas})
}

// UpdateSchemaStatus handles PUT /api/schemas/:version/status
func (h *SchemaHandlers) UpdateSchemaStatus(c *gin.Context) {
	version := c.Param("version")

	var req models.UpdateSchemaStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.schemaService.UpdateSchemaStatus(version, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Schema status updated successfully"})
}

// DeleteSchema handles DELETE /api/schemas/:version
func (h *SchemaHandlers) DeleteSchema(c *gin.Context) {
	version := c.Param("version")

	err := h.schemaService.DeleteSchema(version)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Schema deleted successfully"})
}

// ActivateVersion handles POST /api/schemas/:version/activate
func (h *SchemaHandlers) ActivateVersion(c *gin.Context) {
	version := c.Param("version")

	err := h.schemaService.ActivateVersion(version)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Schema version activated successfully"})
}

// DeactivateVersion handles POST /api/schemas/:version/deactivate
func (h *SchemaHandlers) DeactivateVersion(c *gin.Context) {
	version := c.Param("version")

	err := h.schemaService.DeactivateVersion(version)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Schema version deactivated successfully"})
}

// GetSchemaVersions handles GET /api/schemas/versions
func (h *SchemaHandlers) GetSchemaVersions(c *gin.Context) {
	versions, err := h.schemaService.GetSchemaVersions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"versions": versions})
}

// CheckCompatibility handles POST /api/schemas/check-compatibility
func (h *SchemaHandlers) CheckCompatibility(c *gin.Context) {
	var req struct {
		SDL string `json:"sdl" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	check, err := h.schemaService.CheckCompatibility(req.SDL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, check)
}

// ValidateSDL handles POST /api/schemas/validate
func (h *SchemaHandlers) ValidateSDL(c *gin.Context) {
	var req struct {
		SDL string `json:"sdl" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.schemaService.ValidateSDL(req.SDL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "SDL is valid"})
}

// ExecuteQuery handles POST /api/graphql
func (h *SchemaHandlers) ExecuteQuery(c *gin.Context) {
	var req models.GraphQLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.schemaService.ExecuteQuery(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateSchema handles PUT /api/schemas/:version
func (h *SchemaHandlers) UpdateSchema(c *gin.Context) {
	version := c.Param("version")

	var req models.CreateSchemaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if schema exists
	existing, err := h.schemaService.GetSchemaByVersion(version)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Schema not found"})
		return
	}

	// Validate the new SDL
	if err := h.schemaService.ValidateSDL(req.SDL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check compatibility if it's the active schema
	if existing.Status == "active" {
		check, err := h.schemaService.CheckCompatibility(req.SDL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if !check.Compatible {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":              "Schema update would introduce breaking changes",
				"compatibilityCheck": check,
			})
			return
		}
	}

	// Update the schema
	// Note: In a real implementation, you might want to create a new version instead of updating
	// For now, we'll update the existing schema
	existing.SDL = req.SDL
	existing.Description = req.Description
	existing.UpdatedAt = time.Now()

	// Save the updated schema
	// TODO: Implement update method in database layer
	c.JSON(http.StatusOK, gin.H{"message": "Schema updated successfully"})
}

// GetSchemaVersionsHistory handles GET /api/schemas/versions/history
func (h *SchemaHandlers) GetSchemaVersionsHistory(c *gin.Context) {
	versions, err := h.schemaService.GetAllSchemaVersions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"versions": versions})
}

// GetSchemaVersionsByVersion handles GET /api/schemas/:version/history
func (h *SchemaHandlers) GetSchemaVersionsByVersion(c *gin.Context) {
	version := c.Param("version")

	versions, err := h.schemaService.GetSchemaVersionsByVersion(version)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"versions": versions})
}
