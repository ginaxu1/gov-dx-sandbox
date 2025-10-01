package handlers

import (
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
	return &SchemaHandlers{
		schemaService: schemaService,
	}
}

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
		return
	}

	schema, err := h.schemaService.CreateSchema(&req)
	if err != nil {
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
}
