package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/gov-dx-sandbox/exchange/policy-decision-point/models"
)

// MetadataHandler handles policy metadata operations
type MetadataHandler struct {
	dbService DatabaseServiceInterface
}

// NewMetadataHandler creates a new metadata handler
func NewMetadataHandler(dbService DatabaseServiceInterface) *MetadataHandler {
	return &MetadataHandler{
		dbService: dbService,
	}
}

// CreatePolicyMetadata handles POST /policy-metadata
func (h *MetadataHandler) CreatePolicyMetadata(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req models.PolicyMetadataCreateRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Failed to read request body", "error", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		slog.Error("Failed to parse request body", "error", err)
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.FieldName == "" || req.DisplayName == "" || req.Source == "" || req.AccessControlType == "" {
		http.Error(w, "field_name, display_name, source, and access_control_type are required", http.StatusBadRequest)
		return
	}

	// Create policy metadata record
	id, err := h.dbService.CreatePolicyMetadata(&req)
	if err != nil {
		slog.Error("Failed to create policy metadata", "error", err, "field_name", req.FieldName)
		http.Error(w, fmt.Sprintf("Failed to create policy metadata: %v", err), http.StatusInternalServerError)
		return
	}

	// Send response
	response := models.PolicyMetadataCreateResponse{
		Success: true,
		Message: fmt.Sprintf("Created policy metadata for field %s", req.FieldName),
		ID:      id,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

	slog.Info("Created policy metadata", "id", id, "field_name", req.FieldName)
}

// UpdateAllowList handles POST /allow-list
func (h *MetadataHandler) UpdateAllowList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req models.AllowListUpdateRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Failed to read request body", "error", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		slog.Error("Failed to parse request body", "error", err)
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.FieldName == "" || req.ApplicationID == "" || req.ExpiresAt == "" {
		http.Error(w, "field_name, application_id, and expires_at are required", http.StatusBadRequest)
		return
	}

	// Update allow list
	err = h.dbService.UpdateAllowList(&req)
	if err != nil {
		slog.Error("Failed to update allow list", "error", err, "field_name", req.FieldName, "application_id", req.ApplicationID)
		http.Error(w, fmt.Sprintf("Failed to update allow list: %v", err), http.StatusInternalServerError)
		return
	}

	// Send response
	response := models.AllowListUpdateResponse{
		Success: true,
		Message: fmt.Sprintf("Updated allow list for field %s with application %s", req.FieldName, req.ApplicationID),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	slog.Info("Updated allow list", "field_name", req.FieldName, "application_id", req.ApplicationID)
}
