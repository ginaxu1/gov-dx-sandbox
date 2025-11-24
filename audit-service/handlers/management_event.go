package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gov-dx-sandbox/audit-service/models"
	"github.com/gov-dx-sandbox/audit-service/services"
)

// ManagementEventHandler handles management event-related HTTP requests
type ManagementEventHandler struct {
	managementEventService *services.ManagementEventService
}

// NewManagementEventHandler creates a new management event handler
func NewManagementEventHandler(managementEventService *services.ManagementEventService) *ManagementEventHandler {
	return &ManagementEventHandler{
		managementEventService: managementEventService,
	}
}

// CreateManagementEvent handles POST /api/management-events (for creating new management events)
func (h *ManagementEventHandler) CreateManagementEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateManagementEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Create event
	event, err := h.managementEventService.CreateManagementEvent(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create event: %v", err), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Return the created event
	if err := json.NewEncoder(w).Encode(event); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetManagementEvents handles GET /api/management-events (for querying management events)
func (h *ManagementEventHandler) GetManagementEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse filter parameters
	filter := h.parseManagementEventFilterParams(r)

	// Get events
	response, err := h.managementEventService.GetManagementEvents(r.Context(), filter)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve events: %v", err), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// parseManagementEventFilterParams parses filter parameters from query string for management events
func (h *ManagementEventHandler) parseManagementEventFilterParams(r *http.Request) *models.ManagementEventFilter {
	filter := &models.ManagementEventFilter{}

	// Parse query parameters
	query := r.URL.Query()

	if eventType := query.Get("eventType"); eventType != "" {
		filter.EventType = &eventType
	}

	if status := query.Get("status"); status != "" {
		filter.Status = &status
	}

	if actorType := query.Get("actorType"); actorType != "" {
		filter.ActorType = &actorType
	}

	if actorID := query.Get("actorId"); actorID != "" {
		filter.ActorID = &actorID
	}

	if actorRole := query.Get("actorRole"); actorRole != "" {
		filter.ActorRole = &actorRole
	}

	if targetResource := query.Get("targetResource"); targetResource != "" {
		filter.TargetResource = &targetResource
	}

	if targetResourceID := query.Get("targetResourceId"); targetResourceID != "" {
		filter.TargetResourceID = &targetResourceID
	}

	if startDateStr := query.Get("startDate"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr := query.Get("endDate"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filter.EndDate = &endDate
		}
	}

	// Parse pagination parameters
	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 1000 {
			filter.Limit = limit
		}
	} else {
		filter.Limit = 50 // Default limit
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	return filter
}
