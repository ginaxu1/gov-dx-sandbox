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

<<<<<<< HEAD
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

=======
	// Create schema
>>>>>>> e62b19e (Clean up and unit tests)
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

	response := map[string]interface{}{
		"success": true,
		"message": "Schema created successfully",
		"schema":  schema,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSchemaVersions handles GET /sdl/versions
func (h *SchemaHandlers) GetSchemaVersions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters for filtering
	statusParam := r.URL.Query().Get("status")
	limitParam := r.URL.Query().Get("limit")
	offsetParam := r.URL.Query().Get("offset")

	// Parse limit and offset
	limit := 100 // default limit
	offset := 0  // default offset

	if limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetParam != "" {
		if o, err := strconv.Atoi(offsetParam); err == nil && o >= 0 {
			offset = o
		}
	}

	// Parse status filter
	var status *models.SchemaStatus
	if statusParam != "" {
		s := models.SchemaStatus(statusParam)
		status = &s
	}

	// Get schema versions
	versions, total, err := h.schemaService.GetAllSchemaVersions(status, limit, offset)
	if err != nil {
		http.Error(w, "Failed to retrieve schema versions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"versions": versions,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
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
	// This would typically be done with a router like mux
	// For now, we'll get it from query parameter
	version := r.URL.Query().Get("version")
	if version == "" {
		http.Error(w, "Version parameter is required", http.StatusBadRequest)
		return
	}

	schema, err := h.schemaService.GetSchemaVersion(version)
	if err != nil {
		http.Error(w, "Schema version not found: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schema)
}

// UpdateSchemaStatus handles PUT /sdl/versions/{version}/status
func (h *SchemaHandlers) UpdateSchemaStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract version from URL path
	version := r.URL.Query().Get("version")
	if version == "" {
		http.Error(w, "Version parameter is required", http.StatusBadRequest)
		return
	}

	var req models.UpdateSchemaStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Update schema status
	err := h.schemaService.UpdateSchemaStatus(version, req.IsActive, req.Reason)
	if err != nil {
		http.Error(w, "Failed to update schema status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Schema status updated successfully",
		"version": version,
		"active":  req.IsActive,
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
		http.Error(w, "No active schema found: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schema)
}

// ListVersions handles GET /sdl/versions - List all schema versions
func (h *SchemaHandlers) ListVersions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters for filtering
	statusParam := r.URL.Query().Get("status")
	limitParam := r.URL.Query().Get("limit")
	offsetParam := r.URL.Query().Get("offset")

	// Parse limit and offset
	limit := 100 // default limit
	offset := 0  // default offset

	if limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetParam != "" {
		if o, err := strconv.Atoi(offsetParam); err == nil && o >= 0 {
			offset = o
		}
	}

	// Parse status filter
	var status *models.SchemaStatus
	if statusParam != "" {
		s := models.SchemaStatus(statusParam)
		status = &s
	}

	// Get schema versions
	versions, total, err := h.schemaService.GetAllSchemaVersions(status, limit, offset)
	if err != nil {
		http.Error(w, "Failed to retrieve schema versions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"versions": versions,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetVersion handles GET /sdl/versions/{version} - Get specific schema version
func (h *SchemaHandlers) GetVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract version from URL path
	// This would typically be done with a router like mux
	// For now, we'll get it from query parameter
	version := r.URL.Query().Get("version")
	if version == "" {
		http.Error(w, "Version parameter is required", http.StatusBadRequest)
		return
	}

	schema, err := h.schemaService.GetSchemaVersion(version)
	if err != nil {
		http.Error(w, "Schema version not found: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schema)
}

// ActivateVersion handles POST /sdl/versions/{version}/activate - Activate specific schema version
func (h *SchemaHandlers) ActivateVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract version from URL path
	// This would typically be done with a router like mux
	// For now, we'll get it from query parameter
	version := r.URL.Query().Get("version")
	if version == "" {
		http.Error(w, "Version parameter is required", http.StatusBadRequest)
		return
	}

	// Parse request body for activation reason
	var req struct {
		Reason string `json:"reason,omitempty"`
	}
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&req)
	}

	// Activate the schema version
	reason := req.Reason
	if reason == "" {
		reason = "Manual activation"
	}

	err := h.schemaService.UpdateSchemaStatus(version, true, &reason)
	if err != nil {
		http.Error(w, "Failed to activate schema version: "+err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Schema version activated successfully",
		"version": version,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeactivateVersion handles POST /sdl/versions/{version}/deactivate - Deactivate specific schema version
func (h *SchemaHandlers) DeactivateVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract version from URL path
	version := r.URL.Query().Get("version")
	if version == "" {
		http.Error(w, "Version parameter is required", http.StatusBadRequest)
		return
	}

	// Parse request body for deactivation reason
	var req struct {
		Reason string `json:"reason,omitempty"`
	}
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&req)
	}

	// Deactivate the schema version
	reason := req.Reason
	if reason == "" {
		reason = "Manual deactivation"
	}

	err := h.schemaService.UpdateSchemaStatus(version, false, &reason)
	if err != nil {
		http.Error(w, "Failed to deactivate schema version: "+err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Schema version deactivated successfully",
		"version": version,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSchemaVersionsInfo handles GET /sdl/versions/info - Get information about loaded schema versions
func (h *SchemaHandlers) GetSchemaVersionsInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get active schema version
	activeVersion, err := h.schemaService.GetActiveSchemaVersion()
	if err != nil {
		activeVersion = "none"
	}

	// Get all loaded versions
	loadedVersions := h.schemaService.GetSchemaVersions()

	// Get version counts
	versionCounts := make(map[string]int)
	for version := range loadedVersions {
		// Check if version is loaded in memory
		if h.schemaService.IsSchemaVersionLoaded(version) {
			versionCounts["loaded"]++
		}
	}

	response := map[string]interface{}{
		"active_version":   activeVersion,
		"loaded_versions":  loadedVersions,
		"version_counts":   versionCounts,
		"total_versions":   len(loadedVersions),
		"loaded_in_memory": versionCounts["loaded"],
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
>>>>>>> 8d51df8 (OE add database.go and schema endpoints, update schema functionality)
}
