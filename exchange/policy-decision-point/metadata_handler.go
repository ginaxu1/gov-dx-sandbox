package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gov-dx-sandbox/exchange/policy-decision-point/models"
)

// MetadataHandler handles provider metadata updates
type MetadataHandler struct {
	dbService DatabaseServiceInterface
	evaluator *PolicyEvaluator
}

// NewMetadataHandler creates a new metadata handler
func NewMetadataHandler(dbService DatabaseServiceInterface, evaluator *PolicyEvaluator) *MetadataHandler {
	return &MetadataHandler{
		dbService: dbService,
		evaluator: evaluator,
	}
}

// UpdateProviderMetadata handles POST /metadata/update
func (h *MetadataHandler) UpdateProviderMetadata(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req models.ProviderMetadataUpdateRequest
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
	if req.ApplicationID == "" || len(req.Fields) == 0 {
		http.Error(w, "application_id and fields are required", http.StatusBadRequest)
		return
	}

	// Load current metadata from database
	metadata, err := h.dbService.GetAllProviderMetadata()
	if err != nil {
		slog.Error("Failed to load metadata from database", "error", err)
		http.Error(w, "Failed to load metadata", http.StatusInternalServerError)
		return
	}

	// Update metadata with new grants
	updated := 0
	var updateErrors []string
	for _, fieldGrant := range req.Fields {
		if err := h.updateFieldMetadata(metadata, req.ApplicationID, fieldGrant); err != nil {
			slog.Error("Failed to update field metadata", "field", fieldGrant.FieldName, "error", err)
			updateErrors = append(updateErrors, fmt.Sprintf("field %s: %v", fieldGrant.FieldName, err))
			continue
		}
		updated++
	}

	// If no fields were updated successfully, return an error
	if updated == 0 {
		slog.Error("No fields were updated successfully", "errors", updateErrors)
		http.Error(w, fmt.Sprintf("Failed to update any fields: %v", updateErrors), http.StatusBadRequest)
		return
	}

	// Save updated metadata to database
	if err := h.dbService.UpdateProviderMetadata(metadata); err != nil {
		slog.Error("Failed to save metadata to database", "error", err, "applicationId", req.ApplicationID)
		http.Error(w, fmt.Sprintf("Failed to save metadata: %v", err), http.StatusInternalServerError)
		return
	}

	// Refresh the policy evaluator with updated metadata
	if err := h.evaluator.RefreshMetadata(r.Context()); err != nil {
		slog.Warn("Failed to refresh policy evaluator", "error", err)
		// Don't fail the request, just log the warning
	}

	// Send response
	response := models.ProviderMetadataUpdateResponse{
		Success: true,
		Message: fmt.Sprintf("Updated %d fields for application %s", updated, req.ApplicationID),
		Updated: updated,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	slog.Info("Updated provider metadata", "applicationId", req.ApplicationID, "updated", updated)
}

// updateFieldMetadata updates a specific field's metadata with a new grant
func (h *MetadataHandler) updateFieldMetadata(metadata *models.ProviderMetadata, applicationID string, fieldGrant models.ProviderFieldGrant) error {
	fieldName := fieldGrant.FieldName

	// Validate input
	if fieldName == "" {
		return fmt.Errorf("field name cannot be empty")
	}
	if applicationID == "" {
		return fmt.Errorf("application ID cannot be empty")
	}

	// Ensure metadata.Fields is initialized
	if metadata.Fields == nil {
		metadata.Fields = make(map[string]models.ProviderMetadataField)
	}

	// Get or create field metadata
	field, exists := metadata.Fields[fieldName]
	if !exists {
		// Create new field metadata with default values
		field = models.ProviderMetadataField{
			Owner:             "external", // Default owner
			Provider:          "",         // Empty string will be converted to NULL in database
			ConsentRequired:   false,      // Default to false
			AccessControlType: "public",   // Default to public
			AllowList:         []models.PDPAllowListEntry{},
		}
	}

	// Parse grant duration to get expiration timestamp (ISO 8601 format)
	expiresAt, err := h.parseGrantDuration(fieldGrant.GrantDuration)
	if err != nil {
		slog.Error("Failed to parse grant duration", "duration", fieldGrant.GrantDuration, "error", err)
		return fmt.Errorf("invalid grant duration %s: %w", fieldGrant.GrantDuration, err)
	}

	// Create new allow list entry
	newEntry := models.PDPAllowListEntry{
		ConsumerID:    applicationID,
		ExpiresAt:     expiresAt,
		GrantDuration: fieldGrant.GrantDuration,
	}

	// Check if consumer already exists in allow list
	found := false
	for i, entry := range field.AllowList {
		if entry.ConsumerID == applicationID {
			// Update existing entry
			field.AllowList[i] = newEntry
			found = true
			break
		}
	}

	// Add new entry if not found
	if !found {
		field.AllowList = append(field.AllowList, newEntry)
	}

	// Update the field in metadata
	metadata.Fields[fieldName] = field

	return nil
}

// parseGrantDuration parses an ISO 8601 duration string and returns the expiration timestamp
// Supported format: ISO 8601 (P30D, P1M, P1Y, PT1H, PT30M, P1Y2M3DT4H5M6S)
func (h *MetadataHandler) parseGrantDuration(duration string) (int64, error) {
	if duration == "" {
		// Default to 30 days if no duration specified
		return time.Now().Add(30 * 24 * time.Hour).Unix(), nil
	}

	return parseISODuration(duration)
}

// parseInt parses an integer from string
func parseInt(s string) (int, error) {
	var result int
	for _, char := range s {
		if char < '0' || char > '9' {
			return 0, fmt.Errorf("invalid number: %s", s)
		}
		result = result*10 + int(char-'0')
	}
	return result, nil
}
