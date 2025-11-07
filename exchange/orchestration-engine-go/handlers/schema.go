package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/services"
	"github.com/go-chi/chi/v5"
)

// SchemaHandler handles HTTP requests for schema management
type SchemaHandler struct {
	schemaService   *services.SchemaService
	apiServerClient *services.APIServerClient
}

// NewSchemaHandler creates a new schema handler
func NewSchemaHandler(schemaService *services.SchemaService, apiServerClient *services.APIServerClient) *SchemaHandler {
	return &SchemaHandler{
		schemaService:   schemaService,
		apiServerClient: apiServerClient,
	}
}

// CreateSchemaRequest represents a request to create a new schema
type CreateSchemaRequest struct {
	Version   string `json:"version"`
	SDL       string `json:"sdl"`
	CreatedBy string `json:"created_by"`
}

// ValidateSDLRequest represents a request to validate SDL
type ValidateSDLRequest struct {
	SDL string `json:"sdl"`
}

// CreateSchema handles POST /sdl - create a new schema version
func (h *SchemaHandler) CreateSchema(w http.ResponseWriter, r *http.Request) {

	if h.schemaService == nil {
		http.Error(w, "Schema management not available - database not connected", http.StatusServiceUnavailable)
		return
	}

	var req CreateSchemaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.SDL == "" || req.CreatedBy == "" {
		http.Error(w, "SDL and created_by are required", http.StatusBadRequest)
		return
	}

	if req.Version == "" {
		req.Version = "1.0.0" // Default version
	}

	schema, err := h.schemaService.CreateSchema(req.Version, req.SDL, req.CreatedBy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schema)
}

// GetSchemas handles GET /sdl/versions - get all schema versions
func (h *SchemaHandler) GetSchemas(w http.ResponseWriter, r *http.Request) {

	if h.schemaService == nil {
		http.Error(w, "Schema management not available - database not connected", http.StatusServiceUnavailable)
		return
	}

	schemas, err := h.schemaService.GetAllSchemas()
	if err != nil {
		http.Error(w, "Internal server error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schemas)
}

// GetActiveSchema handles GET /sdl - get the active schema
func (h *SchemaHandler) GetActiveSchema(w http.ResponseWriter, r *http.Request) {
	if h.schemaService == nil {
		http.Error(w, "Schema management not available - database not connected", http.StatusServiceUnavailable)
		return
	}

	schema, err := h.schemaService.GetActiveSchema()
	if err != nil {
		http.Error(w, "Internal server error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if schema == nil {
		http.Error(w, "No active schema found", http.StatusNotFound)
		return
	}

	response := map[string]string{"sdl": schema.SDL}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ActivateSchema handles POST /sdl/versions/{version}/activate - activate a schema version
func (h *SchemaHandler) ActivateSchema(w http.ResponseWriter, r *http.Request) {

	if h.schemaService == nil {
		http.Error(w, "Schema management not available - database not connected", http.StatusServiceUnavailable)
		return
	}

	// Extract version from URL path (simplified)
	version := chi.URLParam(r, "version")

	err := h.schemaService.ActivateSchema(version)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Schema activated successfully"})
}

// ValidateSDL handles POST /sdl/validate - validate SDL syntax
func (h *SchemaHandler) ValidateSDL(w http.ResponseWriter, r *http.Request) {

	if h.schemaService == nil {
		http.Error(w, "Schema management not available - database not connected", http.StatusServiceUnavailable)
		return
	}

	var req ValidateSDLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	valid := h.schemaService.ValidateSDL(req.SDL)
	response := map[string]bool{"valid": valid}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CheckCompatibility handles POST /sdl/check-compatibility - check backward compatibility
func (h *SchemaHandler) CheckCompatibility(w http.ResponseWriter, r *http.Request) {

	if h.schemaService == nil {
		http.Error(w, "Schema management not available - database not connected", http.StatusServiceUnavailable)
		return
	}

	var req ValidateSDLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	compatible, reason := h.schemaService.CheckCompatibility(req.SDL)
	response := map[string]interface{}{
		"compatible": compatible,
		"reason":     reason,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RegisterSchemaRequest represents a request to register a schema to API Server
type RegisterSchemaRequest struct {
	SchemaID   string `json:"schemaId"`
	SchemaName string `json:"schemaName"`
	SDL        string `json:"sdl"`
	Version    string `json:"version"`
}

// RegisterSchema handles POST /schemas/register - register unified schema to API Server
func (h *SchemaHandler) RegisterSchema(w http.ResponseWriter, r *http.Request) {
	if h.apiServerClient == nil {
		http.Error(w, "API Server client not configured", http.StatusServiceUnavailable)
		return
	}

	var req RegisterSchemaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.SDL == "" {
		http.Error(w, "SDL is required", http.StatusBadRequest)
		return
	}

	if req.SchemaID == "" {
		req.SchemaID = "unified-schema-v1"
	}

	if req.SchemaName == "" {
		req.SchemaName = "Unified Schema"
	}

	if req.Version == "" {
		req.Version = "1.0.0"
	}

	// Validate SDL syntax
	if h.schemaService != nil && !h.schemaService.ValidateSDL(req.SDL) {
		http.Error(w, "Invalid SDL syntax", http.StatusBadRequest)
		return
	}

	// Register to API Server
	schema, err := h.apiServerClient.RegisterSchema(req.SchemaID, req.SchemaName, req.SDL, req.Version)
	if err != nil {
		http.Error(w, "Failed to register schema: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(schema)
}
