package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

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

// GetUnifiedSchemas handles GET /admin/unified-schemas
func (h *SchemaMappingHandler) GetUnifiedSchemas(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	schemas, err := h.schemaMappingService.GetUnifiedSchemas()
	if err != nil {
		logger.Log.Error("Failed to get unified schemas", "error", err)
		http.Error(w, "Failed to get unified schemas", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schemas)
}

// GetLatestUnifiedSchema handles GET /admin/unified-schemas/latest
func (h *SchemaMappingHandler) GetLatestUnifiedSchema(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	schema, err := h.schemaMappingService.GetActiveUnifiedSchema()
	if err != nil {
		logger.Log.Error("Failed to get active unified schema", "error", err)
		http.Error(w, "Failed to get active unified schema", http.StatusInternalServerError)
		return
	}

	if schema == nil {
		http.Error(w, "No active unified schema found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schema)
}

// CreateUnifiedSchema handles POST /admin/unified-schemas
func (h *SchemaMappingHandler) CreateUnifiedSchema(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateUnifiedSchemaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Check backward compatibility if there's an active schema
	activeSchema, err := h.schemaMappingService.GetActiveUnifiedSchema()
	if err == nil && activeSchema != nil {
		// Use simple compatibility checker for now
		compatibilityResult := h.compatibilityChecker.CheckCompatibilitySimple(activeSchema.SDL, req.SDL)
		if !compatibilityResult.Compatible {
			logger.Log.Warn("Backward compatibility check failed", "breaking_changes", compatibilityResult.BreakingChanges)
			http.Error(w, "Backward compatibility check failed", http.StatusBadRequest)
			return
		}
	}

	response, err := h.schemaMappingService.CreateUnifiedSchema(&req)
	if err != nil {
		logger.Log.Error("Failed to create unified schema", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// ActivateUnifiedSchema handles PUT /admin/unified-schemas/{version}/activate
func (h *SchemaMappingHandler) ActivateUnifiedSchema(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract version from URL path
	version := r.URL.Path[len("/admin/unified-schemas/"):]
	version = version[:len(version)-len("/activate")]

	response, err := h.schemaMappingService.ActivateUnifiedSchema(version)
	if err != nil {
		logger.Log.Error("Failed to activate unified schema", "version", version, "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetProviderSchemas handles GET /admin/provider-schemas
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

// Field Mapping Handlers

// CreateFieldMapping handles POST /admin/unified-schemas/{version}/mappings
func (h *SchemaMappingHandler) CreateFieldMapping(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract version from URL path
	version := r.URL.Path[len("/admin/unified-schemas/"):]
	version = version[:len(version)-len("/mappings")]

	// Get unified schema ID by version
	unifiedSchema, err := h.schemaMappingService.GetUnifiedSchemaByVersion(version)
	if err != nil {
		logger.Log.Error("Failed to get unified schema", "version", version, "error", err)
		http.Error(w, "Unified schema not found", http.StatusNotFound)
		return
	}

	var req models.FieldMappingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	response, err := h.schemaMappingService.CreateFieldMapping(unifiedSchema.ID, &req)
	if err != nil {
		logger.Log.Error("Failed to create field mapping", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetFieldMappings handles GET /admin/unified-schemas/{version}/mappings
func (h *SchemaMappingHandler) GetFieldMappings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract version from URL path
	version := r.URL.Path[len("/admin/unified-schemas/"):]
	version = version[:len(version)-len("/mappings")]

	// Get unified schema ID by version
	unifiedSchema, err := h.schemaMappingService.GetUnifiedSchemaByVersion(version)
	if err != nil {
		logger.Log.Error("Failed to get unified schema", "version", version, "error", err)
		http.Error(w, "Unified schema not found", http.StatusNotFound)
		return
	}

	mappings, err := h.schemaMappingService.GetFieldMappings(unifiedSchema.ID)
	if err != nil {
		logger.Log.Error("Failed to get field mappings", "error", err)
		http.Error(w, "Failed to get field mappings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mappings)
}

// UpdateFieldMapping handles PUT /admin/unified-schemas/{version}/mappings/{mapping_id}
func (h *SchemaMappingHandler) UpdateFieldMapping(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract mapping ID from URL path
	mappingID := r.URL.Path[len("/admin/unified-schemas/"):]
	// Find the last slash to get mapping ID
	lastSlash := strings.LastIndex(mappingID, "/")
	if lastSlash == -1 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	mappingID = mappingID[lastSlash+1:]

	var req models.FieldMappingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	response, err := h.schemaMappingService.UpdateFieldMapping(mappingID, &req)
	if err != nil {
		logger.Log.Error("Failed to update field mapping", "mapping_id", mappingID, "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteFieldMapping handles DELETE /admin/unified-schemas/{version}/mappings/{mapping_id}
func (h *SchemaMappingHandler) DeleteFieldMapping(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract mapping ID from URL path
	mappingID := r.URL.Path[len("/admin/unified-schemas/"):]
	// Find the last slash to get mapping ID
	lastSlash := strings.LastIndex(mappingID, "/")
	if lastSlash == -1 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	mappingID = mappingID[lastSlash+1:]

	if err := h.schemaMappingService.DeleteFieldMapping(mappingID); err != nil {
		logger.Log.Error("Failed to delete field mapping", "mapping_id", mappingID, "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Compatibility Check Handler

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
