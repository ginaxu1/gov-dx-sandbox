package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gov-dx-sandbox/api-server-go/models"
	"github.com/gov-dx-sandbox/api-server-go/services"
	"github.com/gov-dx-sandbox/exchange/shared/utils"
)

// APIServer manages all API routes and handlers
type APIServer struct {
	consumerService *services.ConsumerService
	providerService *services.ProviderService
	adminService    *services.AdminService
	grantsService   *services.GrantsService
	authService     *services.AuthService
}

// NewAPIServer creates a new API server instance
func NewAPIServer() *APIServer {
	providerService := services.NewProviderService()
	grantsService := services.NewGrantsService()

	// Initialize Asgardeo service
	var asgardeoService *services.AsgardeoService
	asgardeoBaseURL := os.Getenv("ASGARDEO_BASE_URL")
	asgardeoClientID := os.Getenv("ASGARDEO_CLIENT_ID")
	asgardeoClientSecret := os.Getenv("ASGARDEO_CLIENT_SECRET")

	if asgardeoBaseURL != "" && asgardeoClientID != "" && asgardeoClientSecret != "" {
		asgardeoService = services.NewAsgardeoService(asgardeoBaseURL)
		slog.Info("Asgardeo service initialized",
			"baseURL", asgardeoBaseURL,
			"clientID", asgardeoClientID)
	} else {
		slog.Warn("Asgardeo service not configured - missing required environment variables",
			"ASGARDEO_BASE_URL", asgardeoBaseURL != "",
			"ASGARDEO_CLIENT_ID", asgardeoClientID != "",
			"ASGARDEO_CLIENT_SECRET", asgardeoClientSecret != "")
	}

	// Initialize consumer service with Asgardeo integration
	var consumerService *services.ConsumerService
	if asgardeoService != nil {
		consumerService = services.NewConsumerServiceWithAsgardeo(asgardeoService)
	} else {
		consumerService = services.NewConsumerService()
	}

	authService := services.NewAuthService(consumerService)

	return &APIServer{
		consumerService: consumerService,
		providerService: providerService,
		adminService:    services.NewAdminServiceWithServices(consumerService, providerService),
		grantsService:   grantsService,
		authService:     authService,
	}
}

// ProviderService returns the provider service instance
func (s *APIServer) ProviderService() *services.ProviderService {
	return s.providerService
}

// GetProviderService returns the provider service instance (alias for consistency)
func (s *APIServer) GetProviderService() *services.ProviderService {
	return s.providerService
}

// SetupRoutes configures all API routes
func (s *APIServer) SetupRoutes(mux *http.ServeMux) {
	// Consumer routes
	mux.Handle("/consumers", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleConsumers)))
	mux.Handle("/consumers/", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleConsumerByID)))

	// Consumer application routes (RESTful)
	mux.Handle("/consumer-applications", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleConsumerApplications)))
	mux.Handle("/consumer-applications/", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleConsumerApplicationByID)))

	// Provider submission routes
	mux.Handle("/provider-submissions", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleProviderSubmissions)))
	mux.Handle("/provider-submissions/", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleProviderSubmissionByID)))

	// provider routes
	mux.Handle("/providers", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleProvidersCollection)))
	mux.Handle("/providers/", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleProviders)))

	// Admin routes
	mux.Handle("/admin/", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleAdmin)))

	// Allow List Management routes
	mux.Handle("/admin/fields/", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleAllowListRoutes)))

	// Authentication routes
	mux.Handle("/auth/token", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleAuthToken)))
	mux.Handle("/auth/validate", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleAsgardeoTokenValidate)))
	mux.Handle("/auth/exchange", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleTokenExchange)))
}

// Generic handler for collection endpoints (GET all, POST create)
func (s *APIServer) handleCollection(w http.ResponseWriter, r *http.Request, getter func() (interface{}, error), creator func(interface{}) (interface{}, error)) {
	switch r.Method {
	case http.MethodGet:
		items, err := getter()
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve items")
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, items)
	case http.MethodPost:
		if creator == nil {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		var req interface{}
		if err := utils.ParseJSONRequest(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		item, err := creator(req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.RespondWithSuccess(w, http.StatusCreated, item)
	default:
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Generic handler for item endpoints (GET, PUT, DELETE by ID)
func (s *APIServer) handleItem(w http.ResponseWriter, r *http.Request, getter func(string) (interface{}, error), updater func(string, interface{}) (interface{}, error), deleter func(string) error) {
	id := utils.ExtractIDFromPathString(r.URL.Path)
	if id == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "ID is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		item, err := getter(id)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "Item not found")
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, item)
	case http.MethodPut:
		var req interface{}
		if err := utils.ParseJSONRequest(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		item, err := updater(id, req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, item)
	case http.MethodDelete:
		if err := deleter(id); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete item")
			return
		}
		utils.RespondWithSuccess(w, http.StatusNoContent, nil)
	default:
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Helper functions for parsing requests

// Service method wrappers for use with generic handlers
// GetConsumerService returns the consumer service instance
func (s *APIServer) GetConsumerService() *services.ConsumerService {
	return s.consumerService
}

func (s *APIServer) createConsumerServiceMethod(req interface{}) (interface{}, error) {
	createReq := req.(models.CreateConsumerRequest)
	return s.consumerService.CreateConsumer(createReq)
}

func (s *APIServer) parseAndCreateConsumer(req interface{}) (interface{}, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	var createReq models.CreateConsumerRequest
	if err := json.Unmarshal(reqBytes, &createReq); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	return s.consumerService.CreateConsumer(createReq)
}

// Consumer handlers
func (s *APIServer) handleConsumers(w http.ResponseWriter, r *http.Request) {
	s.handleCollection(w, r,
		func() (interface{}, error) {
			consumers, err := s.consumerService.GetAllConsumers()
			if err != nil {
				return nil, err
			}
			return utils.CreateCollectionResponse(consumers, len(consumers)), nil
		},
		func(req interface{}) (interface{}, error) {
			return s.parseAndCreateConsumer(req)
		},
	)
}

func (s *APIServer) handleConsumerByID(w http.ResponseWriter, r *http.Request) {
	id := utils.ExtractIDFromPathString(r.URL.Path)
	if id == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Consumer ID is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		// GET /consumers/{consumerId} - Get specific consumer
		consumer, err := s.consumerService.GetConsumer(id)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "Consumer not found")
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, consumer)
	case http.MethodPut:
		// PUT /consumers/{consumerId} - Update consumer
		var req models.UpdateConsumerRequest
		if err := utils.ParseJSONRequest(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		consumer, err := s.consumerService.UpdateConsumer(id, req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, consumer)
	case http.MethodDelete:
		// DELETE /consumers/{consumerId} - Delete consumer
		err := s.consumerService.DeleteConsumer(id)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete consumer")
			return
		}
		utils.RespondWithSuccess(w, http.StatusNoContent, nil)
	default:
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Consumer application handlers (RESTful)
func (s *APIServer) handleConsumerApplications(w http.ResponseWriter, r *http.Request) {
	// GET /consumer-applications - Get all applications (admin view)
	s.handleCollection(w, r,
		func() (interface{}, error) {
			apps, err := s.consumerService.GetAllConsumerApps()
			if err != nil {
				return nil, err
			}
			return utils.CreateCollectionResponse(apps, len(apps)), nil
		},
		nil, // No POST at this level - use /consumer-applications/:consumerId
	)
}

func (s *APIServer) handleConsumerApplicationByID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	id := utils.ExtractIDFromPathString(path)

	if id == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "ID is required")
		return
	}

	// Determine if this is a consumer ID or submission ID based on the ID format and HTTP method
	if strings.HasPrefix(id, "consumer_") {
		// This is a consumer ID - handle consumer-specific operations
		s.handleConsumerApplicationsForConsumer(w, r, id)
	} else if strings.HasPrefix(id, "sub_") {
		// This is a submission ID - handle individual application operations
		s.handleIndividualConsumerApplication(w, r, id)
	} else {
		// Unknown ID format - try consumer first, then submission
		if r.Method == "POST" || r.Method == "GET" {
			// For POST/GET, assume it's a consumer ID
			s.handleConsumerApplicationsForConsumer(w, r, id)
		} else {
			// For other methods, assume it's a submission ID
			s.handleIndividualConsumerApplication(w, r, id)
		}
	}
}

// Handle operations for a specific consumer's applications
func (s *APIServer) handleConsumerApplicationsForConsumer(w http.ResponseWriter, r *http.Request, consumerID string) {
	switch r.Method {
	case http.MethodGet:
		// GET /consumer-applications/:consumerId - Get applications for specific consumer
		apps, err := s.consumerService.GetConsumerAppsByConsumerID(consumerID)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, utils.CreateCollectionResponse(apps, len(apps)))
	case http.MethodPost:
		// POST /consumer-applications/:consumerId - Create application for specific consumer
		var req models.CreateConsumerAppRequest
		if err := utils.ParseJSONRequest(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Set the consumer ID from the URL
		req.ConsumerID = consumerID

		app, err := s.consumerService.CreateConsumerApp(req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.RespondWithSuccess(w, http.StatusCreated, app)
	default:
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Handle operations for individual consumer applications
func (s *APIServer) handleIndividualConsumerApplication(w http.ResponseWriter, r *http.Request, submissionID string) {
	switch r.Method {
	case http.MethodGet:
		// GET /consumer-applications/:submissionId - Get specific application
		app, err := s.consumerService.GetConsumerApp(submissionID)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, app)
	case http.MethodPut:
		// PUT /consumer-applications/:submissionId - Update application (admin approval)
		var req models.UpdateConsumerAppRequest
		if err := utils.ParseJSONRequest(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		response, err := s.consumerService.UpdateConsumerApp(submissionID, req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.RespondWithSuccess(w, http.StatusOK, response)
	default:
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleProvidersCollection handles /providers (GET all providers)
func (s *APIServer) handleProvidersCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// GET /providers - List all providers
		profiles, err := s.providerService.GetAllProviderProfiles()
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve providers")
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, profiles)
	default:
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleProviderByID handles /providers/{providerId} (GET specific provider)
func (s *APIServer) handleProviderByID(w http.ResponseWriter, r *http.Request, providerID string) {
	switch r.Method {
	case http.MethodGet:
		// GET /providers/{providerId} - Get specific provider
		profile, err := s.providerService.GetProviderProfile(providerID)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "Provider not found")
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, profile)
	default:
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleProviders handles provider routes
func (s *APIServer) handleProviders(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Extract provider ID from path like /providers/{provider-id}/...
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	if len(pathParts) < 2 || pathParts[0] != "providers" {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid provider path")
		return
	}

	providerID := pathParts[1]
	if providerID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Provider ID is required")
		return
	}

	// Handle base provider endpoints
	if len(pathParts) == 2 {
		// /providers/{providerId} - Get specific provider
		if r.Method == http.MethodGet {
			s.handleProviderByID(w, r, providerID)
			return
		}
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Check if this is a schemas sub-resource (approved schemas only)
	if len(pathParts) == 3 && pathParts[2] == "schemas" {
		s.handleProviderSchemas(w, r, providerID)
		return
	}

	// Check if this is a schema-submissions sub-resource
	if len(pathParts) == 3 && pathParts[2] == "schema-submissions" {
		s.handleProviderSchemaSubmissions(w, r, providerID)
		return
	}

	// Check if this is a specific schema submission (e.g., /providers/:id/schema-submissions/:schemaId)
	if len(pathParts) == 4 && pathParts[2] == "schema-submissions" {
		schemaID := pathParts[3]
		s.handleProviderSchemaSubmissionByID(w, r, providerID, schemaID)
		return
	}

	// Handle other provider sub-resources here in the future
	utils.RespondWithError(w, http.StatusNotFound, "Resource not found")
}

// handleProviderSchemaSubmissions handles /providers/:provider-id/schema-submissions
func (s *APIServer) handleProviderSchemaSubmissions(w http.ResponseWriter, r *http.Request, providerID string) {
	switch r.Method {
	case http.MethodPost:
		// POST /providers/:provider-id/schema-submissions - Create new schema submission or modify existing
		var req models.CreateProviderSchemaSubmissionRequest
		if err := utils.ParseJSONRequest(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		schema, err := s.providerService.CreateProviderSchemaSubmission(providerID, req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.RespondWithSuccess(w, http.StatusCreated, schema)
	case http.MethodGet:
		// GET /providers/:provider-id/schema-submissions - List all schema submissions for provider
		schemas, err := s.providerService.GetProviderSchemasByProviderID(providerID)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.RespondWithSuccess(w, http.StatusOK, utils.CreateCollectionResponse(schemas, len(schemas)))
	default:
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleProviderSchemaSubmissionByID handles /providers/:provider-id/schema-submissions/:schemaId
func (s *APIServer) handleProviderSchemaSubmissionByID(w http.ResponseWriter, r *http.Request, providerID, schemaID string) {
	switch r.Method {
	case http.MethodGet:
		// GET /providers/:provider-id/schema-submissions/:schemaId - Get specific schema
		schema, err := s.providerService.GetProviderSchema(schemaID)

		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		// Verify the schema belongs to the provider
		if schema.ProviderID != providerID {
			utils.RespondWithError(w, http.StatusNotFound, "Schema not found for this provider")
			return
		}

		utils.RespondWithSuccess(w, http.StatusOK, schema)
	case http.MethodPut:
		// PUT /providers/:provider-id/schema-submissions/:schemaId - Update schema (status changes)
		var req models.UpdateProviderSchemaRequest
		if err := utils.ParseJSONRequest(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Check if this is a status update to "pending" (submit for review)
		if req.Status != nil && *req.Status == "pending" {
			// Submit schema for review (draft -> pending)
			schema, err := s.providerService.SubmitSchemaForReview(schemaID)
			if err != nil {
				utils.RespondWithError(w, http.StatusBadRequest, err.Error())
				return
			}

			// Verify the schema belongs to the provider
			if schema.ProviderID != providerID {
				utils.RespondWithError(w, http.StatusNotFound, "Schema not found for this provider")
				return
			}

			utils.RespondWithSuccess(w, http.StatusOK, schema)
			return
		}

		// Regular schema update (admin approval/rejection)
		schema, err := s.providerService.UpdateProviderSchema(schemaID, req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Verify the schema belongs to the provider
		if schema.ProviderID != providerID {
			utils.RespondWithError(w, http.StatusNotFound, "Schema not found for this provider")
			return
		}

		utils.RespondWithSuccess(w, http.StatusOK, schema)
	default:
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Admin handler
func (s *APIServer) handleAdmin(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/admin")

	switch {
	case path == "/metrics" && r.Method == http.MethodGet:
		// GET /admin/metrics - Get system metrics
		metrics, err := s.adminService.GetMetrics()
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get metrics")
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, metrics)
	case path == "/recent-activity" && r.Method == http.MethodGet:
		// GET /admin/recent-activity - Get recent system activity
		activity, err := s.adminService.GetRecentActivity()
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get recent activity")
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, activity)
	case path == "/statistics" && r.Method == http.MethodGet:
		// GET /admin/statistics - Get detailed statistics by resource type
		stats, err := s.adminService.GetStatistics()
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get statistics")
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, stats)
	default:
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleProviderSchemas handles /providers/:provider-id/schemas
func (s *APIServer) handleProviderSchemas(w http.ResponseWriter, r *http.Request, providerID string) {
	switch r.Method {
	case http.MethodGet:
		// GET /providers/:provider-id/schemas - List approved schemas for provider
		schemas, err := s.providerService.GetApprovedSchemasByProviderID(providerID)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.RespondWithSuccess(w, http.StatusOK, utils.CreateCollectionResponse(schemas, len(schemas)))
	default:
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Provider submission handlers

// handleProviderSubmissions handles /provider-submissions
func (s *APIServer) handleProviderSubmissions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// GET /provider-submissions - List all provider submissions
		submissions, err := s.providerService.GetAllProviderSubmissions()
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve provider submissions")
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, utils.CreateCollectionResponse(submissions, len(submissions)))
	case http.MethodPost:
		// POST /provider-submissions - Create new provider submission
		var req models.CreateProviderSubmissionRequest
		if err := utils.ParseJSONRequest(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		submission, err := s.providerService.CreateProviderSubmission(req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.RespondWithSuccess(w, http.StatusCreated, submission)
	default:
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleProviderSubmissionByID handles /provider-submissions/{submissionId}
func (s *APIServer) handleProviderSubmissionByID(w http.ResponseWriter, r *http.Request) {
	submissionID := utils.ExtractIDFromPathString(r.URL.Path)
	if submissionID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Submission ID is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		// GET /provider-submissions/{submissionId} - Get specific provider submission
		submission, err := s.providerService.GetProviderSubmission(submissionID)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "Provider submission not found")
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, submission)
	case http.MethodPut:
		// PUT /provider-submissions/{submissionId} - Update provider submission (admin approval/rejection)
		var req models.UpdateProviderSubmissionRequest
		if err := utils.ParseJSONRequest(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		submission, err := s.providerService.UpdateProviderSubmission(submissionID, req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.RespondWithSuccess(w, http.StatusOK, submission)
	default:
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Allow List Management Handlers

// handleAllowListManagement handles /admin/fields/{fieldName}/allow-list
func (s *APIServer) handleAllowListManagement(w http.ResponseWriter, r *http.Request) {
	fieldName := ExtractFieldNameFromPath(r.URL.Path)
	if fieldName == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Field name is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		// GET /admin/fields/{fieldName}/allow-list - List consumers in allow_list
		response, err := s.grantsService.GetAllowListForField(fieldName)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, response)

	case http.MethodPost:
		// POST /admin/fields/{fieldName}/allow-list - Add consumer to allow_list
		var req models.AllowListManagementRequest
		if err := utils.ParseJSONRequest(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		response, err := s.grantsService.AddConsumerToAllowList(fieldName, req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.RespondWithSuccess(w, http.StatusCreated, response)

	default:
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleAllowListConsumerManagement handles /admin/fields/{fieldName}/allow-list/{consumerId}
func (s *APIServer) handleAllowListConsumerManagement(w http.ResponseWriter, r *http.Request) {
	fieldName := ExtractFieldNameFromPath(r.URL.Path)
	consumerID := ExtractConsumerIDFromPath(r.URL.Path)

	if fieldName == "" || consumerID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Field name and consumer ID are required")
		return
	}

	switch r.Method {
	case http.MethodPut:
		// PUT /admin/fields/{fieldName}/allow-list/{consumerId} - Update consumer in allow_list
		var req models.AllowListManagementRequest
		if err := utils.ParseJSONRequest(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		response, err := s.grantsService.UpdateConsumerInAllowList(fieldName, consumerID, req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, response)

	case http.MethodDelete:
		// DELETE /admin/fields/{fieldName}/allow-list/{consumerId} - Remove consumer from allow_list
		response, err := s.grantsService.RemoveConsumerFromAllowList(fieldName, consumerID)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, response)

	default:
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleAllowListRoutes routes allow_list management requests
func (s *APIServer) handleAllowListRoutes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Check if this is a consumer-specific allow_list request
	// Pattern: /admin/fields/{fieldName}/allow-list/{consumerId}
	if strings.Contains(path, "/allow-list/") && len(strings.Split(path, "/")) >= 6 {
		s.handleAllowListConsumerManagement(w, r)
		return
	}

	// Pattern: /admin/fields/{fieldName}/allow-list
	if strings.HasSuffix(path, "/allow-list") {
		s.handleAllowListManagement(w, r)
		return
	}

	// If no pattern matches, return 404
	utils.RespondWithError(w, http.StatusNotFound, "Allow list endpoint not found")
}

// Authentication handlers

// handleAuthToken handles POST /auth/token - Authenticate consumer and get access token
func (s *APIServer) handleAuthToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req models.AuthRequest
	if err := utils.ParseJSONRequest(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.ConsumerID == "" || req.Secret == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "consumerId and secret are required")
		return
	}

	// Validate input length and format
	if len(req.ConsumerID) > 100 || len(req.Secret) > 200 {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid input length")
		return
	}

	// Authenticate consumer
	response, err := s.authService.AuthenticateConsumer(req)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, response)
}

// handleTokenExchange handles POST /auth/exchange - Exchange API credentials for Asgardeo token
func (s *APIServer) handleTokenExchange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req models.TokenExchangeRequest
	if err := utils.ParseJSONRequest(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.APIKey == "" || req.APISecret == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "apiKey and apiSecret are required")
		return
	}

	// Exchange credentials for Asgardeo token
	response, err := s.consumerService.ExchangeCredentialsForToken(req)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, response)
}

// handleAsgardeoTokenValidate handles POST /auth/validate - Validate Asgardeo access token
func (s *APIServer) handleAsgardeoTokenValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req models.ValidateTokenRequest
	if err := utils.ParseJSONRequest(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Token == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "token is required")
		return
	}

	// Validate Asgardeo token
	response, err := s.consumerService.ValidateAsgardeoToken(req.Token)
	if err != nil {
		// Check if the error is due to service not being configured
		if strings.Contains(err.Error(), "not configured") {
			utils.RespondWithError(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to validate Asgardeo token")
		return
	}

	// If validation failed due to service not being configured, return 503
	if !response.Valid && response.Error != "" && strings.Contains(response.Error, "not configured") {
		utils.RespondWithError(w, http.StatusServiceUnavailable, response.Error)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, response)
}
