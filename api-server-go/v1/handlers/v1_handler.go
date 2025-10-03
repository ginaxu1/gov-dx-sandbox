package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gov-dx-sandbox/api-server-go/shared/utils"
	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"github.com/gov-dx-sandbox/api-server-go/v1/services"
	"gorm.io/gorm"
)

// V1Handler handles all V1 API routes
type V1Handler struct {
	providerService *services.ProviderService
	consumerService *services.ConsumerService
	entityService   *services.EntityService
}

// NewV1Handler creates a new V1 handler
func NewV1Handler(db *gorm.DB) *V1Handler {
	entityService := services.NewEntityService(db)
	return &V1Handler{
		entityService:   entityService,
		providerService: services.NewProviderService(db, entityService),
		consumerService: services.NewConsumerService(db, entityService),
	}
}

// SetupV1Routes configures all V1 API routes
func (h *V1Handler) SetupV1Routes(mux *http.ServeMux) {
	// Provider routes
	mux.Handle("/api/v1/providers", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleProviders)))

	// Consumer routes
	mux.Handle("/api/v1/consumers", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleConsumers)))

	// Entity routes
	mux.Handle("/api/v1/entities", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleEntities)))
}

// handleProviders handles provider-related routes
func (h *V1Handler) handleProviders(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/providers")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	// Handle collection endpoint: GET /api/v1/providers and POST /api/v1/providers
	if len(parts) == 1 && parts[0] == "" {
		switch r.Method {
		case http.MethodGet:
			h.getAllProviders(w, r)
		case http.MethodPost:
			h.createProvider(w, r)
		default:
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	if len(parts) < 1 || parts[0] == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Provider ID is required")
		return
	}

	providerID := parts[0]

	// Handle base provider endpoint: GET /api/v1/providers/:providerId and PUT /api/v1/providers/:providerId
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			h.getProvider(w, r, providerID)
		case http.MethodPut:
			h.updateProvider(w, r, providerID)
		default:
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	// Handle provider schemas: GET /api/v1/providers/:providerId/schemas
	if len(parts) == 2 && parts[1] == "schemas" {
		if r.Method == http.MethodGet {
			h.getProviderSchemas(w, r, providerID)
		} else {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	// Handle provider schema submissions: /api/v1/providers/:providerId/schema-submissions
	if len(parts) == 2 && parts[1] == "schema-submissions" {
		switch r.Method {
		case http.MethodGet:
			h.getProviderSchemaSubmissions(w, r, providerID)
		case http.MethodPost:
			h.createProviderSchemaSubmission(w, r, providerID)
		default:
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	utils.RespondWithError(w, http.StatusNotFound, "Endpoint not found")
}

// handleConsumers handles consumer-related routes
func (h *V1Handler) handleConsumers(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/consumers")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	// Handle collection endpoint: GET /api/v1/consumers and POST /api/v1/consumers
	if len(parts) == 1 && parts[0] == "" {
		switch r.Method {
		case http.MethodGet:
			h.getAllConsumers(w, r)
		case http.MethodPost:
			h.createConsumer(w, r)
		default:
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	if len(parts) < 1 || parts[0] == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Consumer ID is required")
		return
	}

	consumerID := parts[0]

	// Handle base consumer endpoint: GET /api/v1/consumers/:consumerId and PUT /api/v1/consumers/:consumerId
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			h.getConsumer(w, r, consumerID)
		case http.MethodPut:
			h.updateConsumer(w, r, consumerID)
		default:
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	// Handle consumer applications: GET /api/v1/consumers/:consumerId/applications
	if len(parts) == 2 && parts[1] == "applications" {
		if r.Method == http.MethodGet {
			h.getConsumerApplications(w, r, consumerID)
		} else {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	// Handle consumer application submissions: /api/v1/consumers/:consumerId/application-submissions
	if len(parts) == 2 && parts[1] == "application-submissions" {
		switch r.Method {
		case http.MethodGet:
			h.getConsumerApplicationSubmissions(w, r, consumerID)
		case http.MethodPost:
			h.createConsumerApplicationSubmission(w, r, consumerID)
		default:
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	utils.RespondWithError(w, http.StatusNotFound, "Endpoint not found")
}

// handleEntities handles entity-related routes
func (h *V1Handler) handleEntities(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/entities")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	// Handle collection endpoint: GET /api/v1/entities and POST /api/v1/entities
	if len(parts) == 1 && parts[0] == "" {
		switch r.Method {
		case http.MethodGet:
			h.getAllEntities(w, r)
		case http.MethodPost:
			h.createEntity(w, r)
		default:
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	if len(parts) < 1 || parts[0] == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Entity ID is required")
		return
	}

	entityID := parts[0]

	// Handle base entity endpoint: GET /api/v1/entities/:entityId and PUT /api/v1/entities/:entityId
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			h.getEntity(w, r, entityID)
		case http.MethodPut:
			h.updateEntity(w, r, entityID)
		default:
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	utils.RespondWithError(w, http.StatusNotFound, "Endpoint not found")
}

// Provider handlers
func (h *V1Handler) createProvider(w http.ResponseWriter, r *http.Request) {
	var req models.CreateProviderRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	provider, err := h.providerService.CreateProvider(&req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, provider)
}

func (h *V1Handler) updateProvider(w http.ResponseWriter, r *http.Request, providerID string) {
	var req models.UpdateProviderRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	provider, err := h.providerService.UpdateProvider(providerID, &req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, provider)
}

func (h *V1Handler) getProvider(w http.ResponseWriter, r *http.Request, providerID string) {
	provider, err := h.providerService.GetProvider(providerID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, provider)
}

func (h *V1Handler) getProviderSchemas(w http.ResponseWriter, r *http.Request, providerID string) {
	schemas, err := h.providerService.GetProviderSchemas(providerID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := models.CollectionResponse{
		Items: schemas,
		Count: len(schemas),
	}
	utils.RespondWithSuccess(w, http.StatusOK, response)
}

func (h *V1Handler) getProviderSchemaSubmissions(w http.ResponseWriter, r *http.Request, providerID string) {
	status := r.URL.Query().Get("status")

	submissions, err := h.providerService.GetProviderSchemaSubmissions(providerID, status)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := models.CollectionResponse{
		Items: submissions,
		Count: len(submissions),
	}
	utils.RespondWithSuccess(w, http.StatusOK, response)
}

func (h *V1Handler) createProviderSchemaSubmission(w http.ResponseWriter, r *http.Request, providerID string) {
	var req models.CreateProviderSchemaSubmissionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	submission, err := h.providerService.CreateProviderSchemaSubmission(providerID, req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, submission)
}

// Consumer handlers
func (h *V1Handler) createConsumer(w http.ResponseWriter, r *http.Request) {
	var req models.CreateConsumerRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	consumer, err := h.consumerService.CreateConsumer(&req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, consumer)
}

func (h *V1Handler) updateConsumer(w http.ResponseWriter, r *http.Request, consumerID string) {
	var req models.UpdateConsumerRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	consumer, err := h.consumerService.UpdateConsumer(consumerID, &req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, consumer)
}

func (h *V1Handler) getConsumer(w http.ResponseWriter, r *http.Request, consumerID string) {
	consumer, err := h.consumerService.GetConsumer(consumerID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, consumer)
}

func (h *V1Handler) getConsumerApplications(w http.ResponseWriter, r *http.Request, consumerID string) {
	applications, err := h.consumerService.GetConsumerApplications(consumerID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := models.CollectionResponse{
		Items: applications,
		Count: len(applications),
	}
	utils.RespondWithSuccess(w, http.StatusOK, response)
}

func (h *V1Handler) getConsumerApplicationSubmissions(w http.ResponseWriter, r *http.Request, consumerID string) {
	status := r.URL.Query().Get("status")

	submissions, err := h.consumerService.GetConsumerApplicationSubmissions(consumerID, status)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := models.CollectionResponse{
		Items: submissions,
		Count: len(submissions),
	}
	utils.RespondWithSuccess(w, http.StatusOK, response)
}

func (h *V1Handler) createConsumerApplicationSubmission(w http.ResponseWriter, r *http.Request, consumerID string) {
	var req models.CreateConsumerApplicationSubmissionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	submission, err := h.consumerService.CreateConsumerApplicationSubmission(consumerID, req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, submission)
}

// Entity handlers
func (h *V1Handler) createEntity(w http.ResponseWriter, r *http.Request) {
	var req models.CreateEntityRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	entity, err := h.entityService.CreateEntity(&req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, entity)
}

func (h *V1Handler) updateEntity(w http.ResponseWriter, r *http.Request, entityID string) {
	var req models.UpdateEntityRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	entity, err := h.entityService.UpdateEntity(entityID, &req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, entity)
}

func (h *V1Handler) getEntity(w http.ResponseWriter, r *http.Request, entityID string) {
	entity, err := h.entityService.GetEntity(entityID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, entity)
}

// Collection handlers
func (h *V1Handler) getAllEntities(w http.ResponseWriter, r *http.Request) {
	entities, err := h.entityService.GetAllEntities()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := models.CollectionResponse{
		Items: entities,
		Count: len(entities),
	}
	utils.RespondWithSuccess(w, http.StatusOK, response)
}

func (h *V1Handler) getAllProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := h.providerService.GetAllProviders()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := models.CollectionResponse{
		Items: providers,
		Count: len(providers),
	}
	utils.RespondWithSuccess(w, http.StatusOK, response)
}

func (h *V1Handler) getAllConsumers(w http.ResponseWriter, r *http.Request) {
	consumers, err := h.consumerService.GetAllConsumers()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := models.CollectionResponse{
		Items: consumers,
		Count: len(consumers),
	}
	utils.RespondWithSuccess(w, http.StatusOK, response)
}
