package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/models"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/services"
)

// SchemaMappingHandler handles HTTP requests for schema mapping
type SchemaMappingHandler struct {
	schemaMappingService *services.SchemaMappingService
	compatibilityChecker *services.CompatibilityChecker
}

// NewSchemaMappingHandler creates a new schema mapping handler
func NewSchemaMappingHandler(schemaMappingService *services.SchemaMappingService, compatibilityChecker *services.CompatibilityChecker) *SchemaMappingHandler {
	return &SchemaMappingHandler{
		schemaMappingService: schemaMappingService,
		compatibilityChecker: compatibilityChecker,
	}
}

// CheckCompatibility handles POST /admin/schemas/compatibility/check
func (h *SchemaMappingHandler) CheckCompatibility(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CompatibilityCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Get old schema by version
	oldSchema, err := h.schemaMappingService.GetUnifiedSchemaByVersion(req.OldVersion)
	if err != nil {
		logger.Log.Error("Failed to get old schema", "version", req.OldVersion, "error", err)
		http.Error(w, "Old schema version not found", http.StatusNotFound)
		return
	}

	// Check compatibility
	result := h.compatibilityChecker.CheckCompatibilitySimple(oldSchema.SDL, req.NewSDL)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Note: Most admin handler methods have been removed as admin routes are now consolidated into /sdl/* routes
// The following methods are still needed for /sdl/* routes

// GetProviderSchemas handles GET /sdl/providers
func (h *SchemaMappingHandler) GetProviderSchemas(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	schemas, err := h.schemaMappingService.GetProviderSchemas()
	if err != nil {
		logger.Log.Error("Failed to get provider schemas", "error", err)
		http.Error(w, "Failed to get provider schemas", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schemas)
}

// UpdateFieldMapping handles PUT /sdl/mappings/{mapping_id}
func (h *SchemaMappingHandler) UpdateFieldMapping(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract mapping ID from URL path
	path := r.URL.Path
	mappingID := path[len("/sdl/mappings/"):]

	if mappingID == "" {
		http.Error(w, "Mapping ID is required", http.StatusBadRequest)
		return
	}

	var req models.FieldMappingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	response, err := h.schemaMappingService.UpdateFieldMapping(mappingID, &req)
	if err != nil {
		logger.Log.Error("Failed to update field mapping", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteFieldMapping handles DELETE /sdl/mappings/{mapping_id}
func (h *SchemaMappingHandler) DeleteFieldMapping(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract mapping ID from URL path
	path := r.URL.Path
	mappingID := path[len("/sdl/mappings/"):]

	if mappingID == "" {
		http.Error(w, "Mapping ID is required", http.StatusBadRequest)
		return
	}

	err := h.schemaMappingService.DeleteFieldMapping(mappingID)
	if err != nil {
		logger.Log.Error("Failed to delete field mapping", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Field Mapping Handlers for Active Schema (used by /sdl/* routes)

// CreateFieldMappingForActiveSchema handles POST /sdl/mappings
func (h *SchemaMappingHandler) CreateFieldMappingForActiveSchema(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get active unified schema
	activeSchema, err := h.schemaMappingService.GetActiveUnifiedSchema()
	if err != nil {
		logger.Log.Error("Failed to get active unified schema", "error", err)
		http.Error(w, "No active unified schema found", http.StatusNotFound)
		return
	}

	var req models.FieldMappingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	response, err := h.schemaMappingService.CreateFieldMapping(activeSchema.ID, &req)
	if err != nil {
		logger.Log.Error("Failed to create field mapping", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetFieldMappingsForActiveSchema handles GET /sdl/mappings
func (h *SchemaMappingHandler) GetFieldMappingsForActiveSchema(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get active unified schema
	activeSchema, err := h.schemaMappingService.GetActiveUnifiedSchema()
	if err != nil {
		logger.Log.Error("Failed to get active unified schema", "error", err)
		http.Error(w, "No active unified schema found", http.StatusNotFound)
		return
	}

	mappings, err := h.schemaMappingService.GetFieldMappings(activeSchema.ID)
	if err != nil {
		logger.Log.Error("Failed to get field mappings", "error", err)
		http.Error(w, "Failed to get field mappings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mappings)
}
