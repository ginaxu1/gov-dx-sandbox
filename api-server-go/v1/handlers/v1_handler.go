package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gov-dx-sandbox/api-server-go/shared/utils"
	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"github.com/gov-dx-sandbox/api-server-go/v1/services"
	"gorm.io/gorm"
)

// V1Handler handles all V1 API routes
type V1Handler struct {
	providerService    *services.ProviderService
	consumerService    *services.ConsumerService
	entityService      *services.EntityService
	applicationService *services.ApplicationService
	schemaService      *services.SchemaService
}

// NewV1Handler creates a new V1 handler
func NewV1Handler(db *gorm.DB) *V1Handler {
	entityService := services.NewEntityService(db)
	pdpServiceURL := os.Getenv("PDP_SERVICE_URL")
	if pdpServiceURL == "" {
		slog.Error("PDP_SERVICE_URL environment variable is not set or empty")
		panic("PDP_SERVICE_URL environment variable is not set or empty")
	}
	pdpService := services.NewPDPService(pdpServiceURL)
	slog.Info("PDP Service URL", "url", pdpServiceURL)
	return &V1Handler{
		entityService:      entityService,
		providerService:    services.NewProviderService(db, entityService),
		consumerService:    services.NewConsumerService(db, entityService),
		schemaService:      services.NewSchemaService(db, pdpService),
		applicationService: services.NewApplicationService(db, pdpService),
	}
}

// SetupV1Routes configures all V1 API routes
func (h *V1Handler) SetupV1Routes(mux *http.ServeMux) {
	// Provider routes
	mux.Handle("/api/v1/providers", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleProviders)))
	mux.Handle("/api/v1/providers/", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleProviders)))

	// Schema routes
	mux.Handle("/api/v1/schemas", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleSchemas)))
	mux.Handle("/api/v1/schemas/", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleSchemas)))

	// SchemaSubmission routes
	mux.Handle("/api/v1/schema-submissions", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleSchemaSubmissions)))
	mux.Handle("/api/v1/schema-submissions/", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleSchemaSubmissions)))

	// Consumer routes
	mux.Handle("/api/v1/consumers", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleConsumers)))
	mux.Handle("/api/v1/consumers/", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleConsumers)))

	// Application routes
	mux.Handle("/api/v1/applications", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleApplications)))
	mux.Handle("/api/v1/applications/", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleApplications)))

	// ApplicationSubmission routes
	mux.Handle("/api/v1/application-submissions", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleApplicationSubmissions)))
	mux.Handle("/api/v1/application-submissions/", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleApplicationSubmissions)))

	// Entity routes
	mux.Handle("/api/v1/entities", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleEntities)))
	mux.Handle("/api/v1/entities/", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleEntities)))
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
		switch r.Method {
		case http.MethodGet:
			h.getAllSchemas(w, r, &providerID)
		default:
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	// Handle provider schema submissions: /api/v1/providers/:providerId/schema-submissions?status=pending&status=rejected
	if len(parts) == 2 && parts[1] == "schema-submissions" {
		switch r.Method {
		case http.MethodGet:
			status := r.URL.Query()["status"]
			h.getAllSchemaSubmissions(w, r, &providerID, &status)
		case http.MethodPost:
			h.createSchemaSubmission(w, r, &providerID)
		default:
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	// Handle specific provider schema: PUT /api/v1/providers/:providerId/schemas/:schemaId
	if len(parts) == 3 && parts[1] == "schemas" {
		switch r.Method {
		case http.MethodPut:
			h.updateSchema(w, r, parts[2])
		default:
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	// Handle specific provider schema submission: PUT /api/v1/providers/:providerId/schema-submissions/:submissionId
	if len(parts) == 3 && parts[1] == "schema-submissions" {
		switch r.Method {
		case http.MethodPut:
			h.updateSchemaSubmission(w, r, parts[2])
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
			h.getAllApplications(w, r, &consumerID)
		} else {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	// Handle consumer application submissions: /api/v1/consumers/:consumerId/application-submissions?status=pending&status=rejected
	if len(parts) == 2 && parts[1] == "application-submissions" {
		switch r.Method {
		case http.MethodGet:
			status := r.URL.Query()["status"]
			h.getAllApplicationSubmissions(w, r, &consumerID, &status)
		case http.MethodPost:
			h.createApplicationSubmission(w, r, &consumerID)
		default:
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	// Handle specific consumer application: PUT /api/v1/consumers/:consumerId/applications/:applicationId
	if len(parts) == 3 && parts[1] == "applications" {
		switch r.Method {
		case http.MethodPut:
			h.updateApplication(w, r, parts[2])
		default:
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	// Handle specific consumer application submission: PUT /api/v1/consumers/:consumerId/application-submissions/:submissionId
	if len(parts) == 3 && parts[1] == "application-submissions" {
		switch r.Method {
		case http.MethodPut:
			h.updateApplicationSubmission(w, r, parts[2])
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

// handleSchemas handles schema-related routes
func (h *V1Handler) handleSchemas(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/schemas")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	// Handle collection endpoint: GET /api/v1/schemas and POST /api/v1/schemas
	if len(parts) == 1 && parts[0] == "" {
		switch r.Method {
		case http.MethodGet:
			providerID := r.URL.Query().Get("providerId")
			h.getAllSchemas(w, r, &providerID)
		case http.MethodPost:
			h.createSchema(w, r)
		default:
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}
	if len(parts) < 1 || parts[0] == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Schema ID is required")
		return
	}
	schemaID := parts[0]

	// Handle specific schema endpoint: GET /api/v1/schemas/:schemaId and PUT /api/v1/schemas/:schemaId
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			h.getSchema(w, r, schemaID)
		case http.MethodPut:
			h.updateSchema(w, r, schemaID)
		default:
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	utils.RespondWithError(w, http.StatusNotFound, "Endpoint not found")
}

// handleSchemaSubmissions handles schema submission-related routes
func (h *V1Handler) handleSchemaSubmissions(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/schema-submissions")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	// Handle collection endpoint: GET /api/v1/schema-submissions and POST /api/v1/schema-submissions
	if len(parts) == 1 && parts[0] == "" {
		switch r.Method {
		case http.MethodGet:
			status := r.URL.Query()["status"]
			providerID := r.URL.Query().Get("providerId")
			h.getAllSchemaSubmissions(w, r, &providerID, &status)
		case http.MethodPost:
			h.createSchemaSubmission(w, r, nil)
		default:
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}
	if len(parts) < 1 || parts[0] == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Submission ID is required")
		return
	}
	submissionID := parts[0]
	// Handle specific schema submission endpoint: GET /api/v1/schema-submissions/:submissionId and PUT /api/v1/schema-submissions/:submissionId
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			h.getSchemaSubmission(w, r, submissionID)
		case http.MethodPut:
			h.updateSchemaSubmission(w, r, submissionID)
		default:
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	utils.RespondWithError(w, http.StatusNotFound, "Endpoint not found")
}

// handleApplications handles application-related routes
func (h *V1Handler) handleApplications(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/applications")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	// Handle collection endpoint: GET /api/v1/applications and POST /api/v1/applications
	if len(parts) == 1 && parts[0] == "" {
		switch r.Method {
		case http.MethodGet:
			consumerID := r.URL.Query().Get("consumerId")
			h.getAllApplications(w, r, &consumerID)
		case http.MethodPost:
			h.createApplication(w, r)
		default:
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}
	if len(parts) < 1 || parts[0] == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Application ID is required")
		return
	}

	applicationID := parts[0]
	// Handle specific application endpoint: GET /api/v1/applications/:applicationId and PUT /api/v1/applications/:applicationId
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			h.getApplication(w, r, applicationID)
		case http.MethodPut:
			h.updateApplication(w, r, applicationID)
		default:
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	utils.RespondWithError(w, http.StatusNotFound, "Endpoint not found")
}

// handleApplicationSubmissions handles application submission-related routes
func (h *V1Handler) handleApplicationSubmissions(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/application-submissions")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	// Handle collection endpoint: GET /api/v1/application-submissions and POST /api/v1/application-submissions
	if len(parts) == 1 && parts[0] == "" {
		switch r.Method {
		case http.MethodGet:
			status := r.URL.Query()["status"]
			consumerID := r.URL.Query().Get("consumerId")
			h.getAllApplicationSubmissions(w, r, &consumerID, &status)
		case http.MethodPost:
			h.createApplicationSubmission(w, r, nil)
		default:
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	if len(parts) < 1 || parts[0] == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Submission ID is required")
		return
	}

	submissionID := parts[0]
	// Handle specific application submission endpoint: GET /api/v1/application-submissions/:submissionId and PUT /api/v1/application-submissions/:submissionId
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			h.getApplicationSubmission(w, r, submissionID)
		case http.MethodPut:
			h.updateApplicationSubmission(w, r, submissionID)
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

// Schema handlers
func (h *V1Handler) getAllSchemaSubmissions(w http.ResponseWriter, r *http.Request, providerID *string, statusFilter *[]string) {
	submissions, err := h.schemaService.GetSchemaSubmissions(providerID, statusFilter)
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

func (h *V1Handler) getSchemaSubmission(w http.ResponseWriter, r *http.Request, submissionID string) {
	submission, err := h.schemaService.GetSchemaSubmission(submissionID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, submission)
}

func (h *V1Handler) createSchemaSubmission(w http.ResponseWriter, r *http.Request, providerID *string) {
	var req models.CreateSchemaSubmissionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if providerID != nil {
		req.ProviderID = *providerID
	}

	submission, err := h.schemaService.CreateSchemaSubmission(&req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, submission)
}

func (h *V1Handler) updateSchemaSubmission(w http.ResponseWriter, r *http.Request, submissionID string) {
	var req models.UpdateSchemaSubmissionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	submission, err := h.schemaService.UpdateSchemaSubmission(submissionID, &req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, submission)
}

func (h *V1Handler) getAllSchemas(w http.ResponseWriter, r *http.Request, providerID *string) {
	schemas, err := h.schemaService.GetSchemas(providerID)
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

func (h *V1Handler) getSchema(w http.ResponseWriter, r *http.Request, submissionID string) {
	schema, err := h.schemaService.GetSchema(submissionID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, schema)
}

func (h *V1Handler) createSchema(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSchemaRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	schema, err := h.schemaService.CreateSchema(&req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, schema)
}

func (h *V1Handler) updateSchema(w http.ResponseWriter, r *http.Request, schemaID string) {
	var req models.UpdateSchemaRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	schema, err := h.schemaService.UpdateSchema(schemaID, &req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, schema)
}

// Application handlers
func (h *V1Handler) getAllApplicationSubmissions(w http.ResponseWriter, r *http.Request, consumerID *string, statusFilter *[]string) {
	submissions, err := h.applicationService.GetApplicationSubmissions(consumerID, statusFilter)
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

func (h *V1Handler) getApplicationSubmission(w http.ResponseWriter, r *http.Request, submissionID string) {
	submission, err := h.applicationService.GetApplicationSubmission(submissionID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, submission)
}

func (h *V1Handler) createApplicationSubmission(w http.ResponseWriter, r *http.Request, consumerID *string) {
	var req models.CreateApplicationSubmissionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if consumerID != nil {
		req.ConsumerID = *consumerID
	}

	submission, err := h.applicationService.CreateApplicationSubmission(&req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, submission)
}

func (h *V1Handler) updateApplicationSubmission(w http.ResponseWriter, r *http.Request, submissionID string) {
	var req models.UpdateApplicationSubmissionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	submission, err := h.applicationService.UpdateApplicationSubmission(submissionID, &req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, submission)
}

func (h *V1Handler) getAllApplications(w http.ResponseWriter, r *http.Request, consumerID *string) {
	applications, err := h.applicationService.GetApplications(consumerID)
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

func (h *V1Handler) getApplication(w http.ResponseWriter, r *http.Request, submissionID string) {
	application, err := h.applicationService.GetApplication(submissionID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, application)
}

func (h *V1Handler) createApplication(w http.ResponseWriter, r *http.Request) {
	var req models.CreateApplicationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	application, err := h.applicationService.CreateApplication(&req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, application)
}

func (h *V1Handler) updateApplication(w http.ResponseWriter, r *http.Request, applicationID string) {
	var req models.UpdateApplicationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	application, err := h.applicationService.UpdateApplication(applicationID, &req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, application)
}
