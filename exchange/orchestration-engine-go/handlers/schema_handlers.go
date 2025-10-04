package handlers

import (
<<<<<<< HEAD
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
=======
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/models"
)

// SchemaHandlers handles HTTP requests for schema management
type SchemaHandlers struct {
	schemaService models.SchemaService
}

// NewSchemaHandlers creates a new schema handlers instance
func NewSchemaHandlers(schemaService models.SchemaService) *SchemaHandlers {
>>>>>>> 8d51df8 (OE add database.go and schema endpoints, update schema functionality)
	return &SchemaHandlers{
		schemaService: schemaService,
	}
}

<<<<<<< HEAD
// CreateSchema handles POST /api/schemas
func (h *SchemaHandlers) CreateSchema(c *gin.Context) {
	var req models.CreateSchemaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
=======
// CreateSchema handles POST /sdl
func (h *SchemaHandlers) CreateSchema(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateSchemaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.Version == "" {
		http.Error(w, "Version is required", http.StatusBadRequest)
		return
	}
	if req.SDL == "" {
		http.Error(w, "SDL is required", http.StatusBadRequest)
		return
	}
	if req.CreatedBy == "" {
		http.Error(w, "CreatedBy is required", http.StatusBadRequest)
		return
	}
	if req.ChangeType == "" {
		http.Error(w, "ChangeType is required", http.StatusBadRequest)
		return
	}

	// Validate change type
	validChangeTypes := map[models.VersionChangeType]bool{
		models.VersionChangeTypeMajor: true,
		models.VersionChangeTypeMinor: true,
		models.VersionChangeTypePatch: true,
	}
	if !validChangeTypes[req.ChangeType] {
		http.Error(w, "Invalid change type. Must be 'major', 'minor', or 'patch'", http.StatusBadRequest)
>>>>>>> 8d51df8 (OE add database.go and schema endpoints, update schema functionality)
		return
	}

	schema, err := h.schemaService.CreateSchema(&req)
	if err != nil {
<<<<<<< HEAD
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
=======
		http.Error(w, "Failed to create schema: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := &models.SchemaVersionResponse{
		ID:                schema.ID,
		Version:           schema.Version,
		SDL:               schema.SDL,
		CreatedAt:         schema.CreatedAt,
		CreatedBy:         schema.CreatedBy,
		Status:            schema.Status,
		ChangeType:        schema.ChangeType,
		Notes:             schema.Notes,
		PreviousVersionID: schema.PreviousVersionID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetSchemaVersions handles GET /sdl/versions
func (h *SchemaHandlers) GetSchemaVersions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	statusParam := r.URL.Query().Get("status")
	limitParam := r.URL.Query().Get("limit")
	offsetParam := r.URL.Query().Get("offset")

	// Set defaults
	limit := 50
	offset := 0
	var status *models.SchemaStatus

	// Parse limit
	if limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	// Parse offset
	if offsetParam != "" {
		if parsed, err := strconv.Atoi(offsetParam); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Parse status filter
	if statusParam != "" {
		s := models.SchemaStatus(statusParam)
		validStatuses := map[models.SchemaStatus]bool{
			models.SchemaStatusActive:     true,
			models.SchemaStatusInactive:   true,
			models.SchemaStatusDeprecated: true,
		}
		if validStatuses[s] {
			status = &s
		}
	}

	schemas, total, err := h.schemaService.GetAllSchemaVersions(status, limit, offset)
	if err != nil {
		http.Error(w, "Failed to get schema versions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to response format
	versions := make([]models.SchemaVersionResponse, len(schemas))
	for i, schema := range schemas {
		versions[i] = models.SchemaVersionResponse{
			ID:                schema.ID,
			Version:           schema.Version,
			SDL:               schema.SDL,
			CreatedAt:         schema.CreatedAt,
			CreatedBy:         schema.CreatedBy,
			Status:            schema.Status,
			ChangeType:        schema.ChangeType,
			Notes:             schema.Notes,
			PreviousVersionID: schema.PreviousVersionID,
		}
	}

	response := &models.SchemaVersionsListResponse{
		Versions: versions,
		Total:    total,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSchemaVersion handles GET /sdl/versions/{version}
func (h *SchemaHandlers) GetSchemaVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract version from URL path
	// This is a simplified extraction - in a real implementation, you'd use a router
	version := r.URL.Path[len("/sdl/versions/"):]
	if version == "" {
		http.Error(w, "Version parameter is required", http.StatusBadRequest)
		return
	}

	schema, err := h.schemaService.GetSchemaVersion(version)
	if err != nil {
		http.Error(w, "Failed to get schema version: "+err.Error(), http.StatusNotFound)
		return
	}

	response := &models.SchemaVersionResponse{
		ID:                schema.ID,
		Version:           schema.Version,
		SDL:               schema.SDL,
		CreatedAt:         schema.CreatedAt,
		CreatedBy:         schema.CreatedBy,
		Status:            schema.Status,
		ChangeType:        schema.ChangeType,
		Notes:             schema.Notes,
		PreviousVersionID: schema.PreviousVersionID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateSchemaStatus handles PUT /sdl/versions/{version}/status
func (h *SchemaHandlers) UpdateSchemaStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract version from URL path
	version := r.URL.Path[len("/sdl/versions/"):]
	if version == "" {
		http.Error(w, "Version parameter is required", http.StatusBadRequest)
		return
	}

	// Remove "/status" suffix
	if len(version) > 7 && version[len(version)-7:] == "/status" {
		version = version[:len(version)-7]
	}

	var req models.UpdateSchemaStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	err := h.schemaService.UpdateSchemaStatus(version, req.IsActive, req.Reason)
	if err != nil {
		http.Error(w, "Failed to update schema status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the updated schema
	schema, err := h.schemaService.GetSchemaVersion(version)
	if err != nil {
		http.Error(w, "Failed to get updated schema: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := &models.SchemaVersionResponse{
		ID:                schema.ID,
		Version:           schema.Version,
		SDL:               schema.SDL,
		CreatedAt:         schema.CreatedAt,
		CreatedBy:         schema.CreatedBy,
		Status:            schema.Status,
		ChangeType:        schema.ChangeType,
		Notes:             schema.Notes,
		PreviousVersionID: schema.PreviousVersionID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetActiveSchema handles GET /sdl/active
func (h *SchemaHandlers) GetActiveSchema(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	schema, err := h.schemaService.GetActiveSchema()
	if err != nil {
		http.Error(w, "Failed to get active schema: "+err.Error(), http.StatusNotFound)
		return
	}

	response := &models.SchemaVersionResponse{
		ID:                schema.ID,
		Version:           schema.Version,
		SDL:               schema.SDL,
		CreatedAt:         schema.CreatedAt,
		CreatedBy:         schema.CreatedBy,
		Status:            schema.Status,
		ChangeType:        schema.ChangeType,
		Notes:             schema.Notes,
		PreviousVersionID: schema.PreviousVersionID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
>>>>>>> 8d51df8 (OE add database.go and schema endpoints, update schema functionality)
}
