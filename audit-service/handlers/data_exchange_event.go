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

// DataExchangeEventHandler handles DataExchangeEvent-related HTTP requests
type DataExchangeEventHandler struct {
	DataExchangeEventService *services.DataExchangeEventService
}

// NewDataExchangeEventHandler creates a new DataExchangeEvent handler
func NewDataExchangeEventHandler(DataExchangeEventService *services.DataExchangeEventService) *DataExchangeEventHandler {
	return &DataExchangeEventHandler{
		DataExchangeEventService: DataExchangeEventService,
	}
}

// CreateDataExchangeEvent handles POST /api/data-exchange-events (for data exchange logging from OE)
func (h *DataExchangeEventHandler) CreateDataExchangeEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateDataExchangeEventRequest

	// Decode request body into DataExchangeEvent struct
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request payload: %v", err), http.StatusBadRequest)
		return
	}

	// Create event
	event, err := h.DataExchangeEventService.CreateDataExchangeEvent(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create event log: %v", err), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Return the created event log
	if err := json.NewEncoder(w).Encode(event); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetDataExchangeEvents handles GET /api/data-exchange-events (for Admin Portal and Entity Portals)
func (h *DataExchangeEventHandler) GetDataExchangeEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse filter parameters
	filter := h.parseDataExchangeEventFilterParams(r)

	// Get events
	response, err := h.DataExchangeEventService.GetDataExchangeEvents(r.Context(), filter)
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

// parseDataExchangeEventFilterParams parses filter parameters from query string for event logs
func (h *DataExchangeEventHandler) parseDataExchangeEventFilterParams(r *http.Request) *models.DataExchangeEventFilter {
	filter := &models.DataExchangeEventFilter{}

	query := r.URL.Query()

	if status := query.Get("status"); status != "" {
		filter.Status = &status
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

	if applicationId := query.Get("applicationId"); applicationId != "" {
		filter.ApplicationID = &applicationId
	}

	if schemaId := query.Get("schemaId"); schemaId != "" {
		filter.SchemaID = &schemaId
	}

	if consumerId := query.Get("consumerId"); consumerId != "" {
		filter.ConsumerID = &consumerId
	}

	if providerId := query.Get("providerId"); providerId != "" {
		filter.ProviderID = &providerId
	}

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
