package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gov-dx-sandbox/api-server-go/idp"
	"github.com/gov-dx-sandbox/api-server-go/idp/idpfactory"
	"github.com/gov-dx-sandbox/api-server-go/shared/utils"
	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"github.com/gov-dx-sandbox/api-server-go/v1/services"

	"gorm.io/gorm"
)

// V1Handler handles all V1 API routes
type V1Handler struct {
	memberService      *services.MemberService
	applicationService *services.ApplicationService
	schemaService      *services.SchemaService
}

// NewV1Handler creates a new V1 handler
func NewV1Handler(db *gorm.DB) (*V1Handler, error) {
	// Get scopes from environment variable, fallback to default if not set
	asgScopesEnv := os.Getenv("ASGARDEO_SCOPES")
	var scopes []string
	if asgScopesEnv != "" {
		// Split by space to handle multiple scopes
		scopes = strings.Fields(asgScopesEnv)
	}
	// Create the NewIdpProvider
	idpProvider, err := idpfactory.NewIdpAPIProvider(idpfactory.FactoryConfig{
		ProviderType: idp.ProviderAsgardeo,
		BaseURL:      os.Getenv("ASGARDEO_BASE_URL"),
		ClientID:     os.Getenv("ASGARDEO_CLIENT_ID"),
		ClientSecret: os.Getenv("ASGARDEO_CLIENT_SECRET"),
		Scopes:       scopes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create IDP provider: %w", err)
	}
	memberService := services.NewMemberService(db, &idpProvider)

	pdpServiceURL := os.Getenv("CHOREO_PDP_CONNECTION_SERVICEURL")
	if pdpServiceURL == "" {
		return nil, fmt.Errorf("CHOREO_PDP_CONNECTION_SERVICEURL environment variable not set")
	}

	pdpServiceAPIKey := os.Getenv("CHOREO_PDP_CONNECTION_CHOREOAPIKEY")
	if pdpServiceAPIKey == "" {
		return nil, fmt.Errorf("CHOREO_PDP_CONNECTION_CHOREOAPIKEY environment variable not set")
	}

	pdpService := services.NewPDPService(pdpServiceURL, pdpServiceAPIKey)
	slog.Info("PDP Service URL", "url", pdpServiceURL)
	return &V1Handler{
		memberService:      memberService,
		schemaService:      services.NewSchemaService(db, pdpService),
		applicationService: services.NewApplicationService(db, pdpService),
	}, nil
}

// SetupV1Routes configures all V1 API routes
func (h *V1Handler) SetupV1Routes(mux *http.ServeMux) {
	// Schema routes
	mux.Handle("/api/v1/schemas", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleSchemas)))
	mux.Handle("/api/v1/schemas/", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleSchemas)))

	// SchemaSubmission routes
	mux.Handle("/api/v1/schema-submissions", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleSchemaSubmissions)))
	mux.Handle("/api/v1/schema-submissions/", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleSchemaSubmissions)))

	// Application routes
	mux.Handle("/api/v1/applications", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleApplications)))
	mux.Handle("/api/v1/applications/", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleApplications)))

	// ApplicationSubmission routes
	mux.Handle("/api/v1/application-submissions", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleApplicationSubmissions)))
	mux.Handle("/api/v1/application-submissions/", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleApplicationSubmissions)))

	// Member routes
	mux.Handle("/api/v1/members", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleMembers)))
	mux.Handle("/api/v1/members/", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleMembers)))
}

// handleMembers handles member-related routes
func (h *V1Handler) handleMembers(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/members")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	// Handle collection endpoint: GET /api/v1/members and POST /api/v1/members
	if len(parts) == 1 && parts[0] == "" {
		switch r.Method {
		case http.MethodGet:
			idpUserId := r.URL.Query().Get("idpUserId")
			email := r.URL.Query().Get("email")
			h.getAllMembers(w, r, &idpUserId, &email)
		case http.MethodPost:
			h.createMember(w, r)
		default:
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	if len(parts) < 1 || parts[0] == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Member ID is required")
		return
	}

	memberId := parts[0]

	// Handle base member endpoint: GET /api/v1/members/:memberId and PUT /api/v1/members/:memberId
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			h.getMember(w, r, memberId)
		case http.MethodPut:
			h.updateMember(w, r, memberId)
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
			memberId := r.URL.Query().Get("memberId")
			h.getAllSchemas(w, r, &memberId)
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
	schemaId := parts[0]

	// Handle specific schema endpoint: GET /api/v1/schemas/:schemaId and PUT /api/v1/schemas/:schemaId
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			h.getSchema(w, r, schemaId)
		case http.MethodPut:
			h.updateSchema(w, r, schemaId)
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
			memberId := r.URL.Query().Get("memberId")
			h.getAllSchemaSubmissions(w, r, &memberId, &status)
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
	submissionId := parts[0]
	// Handle specific schema submission endpoint: GET /api/v1/schema-submissions/:submissionId and PUT /api/v1/schema-submissions/:submissionId
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			h.getSchemaSubmission(w, r, submissionId)
		case http.MethodPut:
			h.updateSchemaSubmission(w, r, submissionId)
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
			memberId := r.URL.Query().Get("memberId")
			h.getAllApplications(w, r, &memberId)
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

	applicationId := parts[0]
	// Handle specific application endpoint: GET /api/v1/applications/:applicationId and PUT /api/v1/applications/:applicationId
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			h.getApplication(w, r, applicationId)
		case http.MethodPut:
			h.updateApplication(w, r, applicationId)
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
			memberId := r.URL.Query().Get("memberId")
			h.getAllApplicationSubmissions(w, r, &memberId, &status)
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

	submissionId := parts[0]
	// Handle specific application submission endpoint: GET /api/v1/application-submissions/:submissionId and PUT /api/v1/application-submissions/:submissionId
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			h.getApplicationSubmission(w, r, submissionId)
		case http.MethodPut:
			h.updateApplicationSubmission(w, r, submissionId)
		default:
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}
	utils.RespondWithError(w, http.StatusNotFound, "Endpoint not found")
}

// Member handlers
func (h *V1Handler) createMember(w http.ResponseWriter, r *http.Request) {
	var req models.CreateMemberRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	member, err := h.memberService.CreateMember(&req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, member)
}

func (h *V1Handler) updateMember(w http.ResponseWriter, r *http.Request, memberId string) {
	var req models.UpdateMemberRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	member, err := h.memberService.UpdateMember(memberId, &req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, member)
}

func (h *V1Handler) getMember(w http.ResponseWriter, r *http.Request, memberId string) {
	member, err := h.memberService.GetMember(memberId)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, member)
}

func (h *V1Handler) getAllMembers(w http.ResponseWriter, r *http.Request, idpUserId *string, email *string) {
	members, err := h.memberService.GetAllMembers(idpUserId, email)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := models.CollectionResponse{
		Items: members,
		Count: len(members),
	}
	utils.RespondWithSuccess(w, http.StatusOK, response)
}

// Schema handlers
func (h *V1Handler) getAllSchemaSubmissions(w http.ResponseWriter, r *http.Request, memberId *string, statusFilter *[]string) {
	submissions, err := h.schemaService.GetSchemaSubmissions(memberId, statusFilter)
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

func (h *V1Handler) getSchemaSubmission(w http.ResponseWriter, r *http.Request, submissionId string) {
	submission, err := h.schemaService.GetSchemaSubmission(submissionId)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, submission)
}

func (h *V1Handler) createSchemaSubmission(w http.ResponseWriter, r *http.Request, memberId *string) {
	var req models.CreateSchemaSubmissionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if memberId != nil {
		req.MemberID = *memberId
	}

	submission, err := h.schemaService.CreateSchemaSubmission(&req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, submission)
}

func (h *V1Handler) updateSchemaSubmission(w http.ResponseWriter, r *http.Request, submissionId string) {
	var req models.UpdateSchemaSubmissionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	submission, err := h.schemaService.UpdateSchemaSubmission(submissionId, &req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, submission)
}

func (h *V1Handler) getAllSchemas(w http.ResponseWriter, r *http.Request, memberId *string) {
	schemas, err := h.schemaService.GetSchemas(memberId)
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

func (h *V1Handler) getSchema(w http.ResponseWriter, r *http.Request, submissionId string) {
	schema, err := h.schemaService.GetSchema(submissionId)
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

func (h *V1Handler) updateSchema(w http.ResponseWriter, r *http.Request, schemaId string) {
	var req models.UpdateSchemaRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	schema, err := h.schemaService.UpdateSchema(schemaId, &req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, schema)
}

// Application handlers
func (h *V1Handler) getAllApplicationSubmissions(w http.ResponseWriter, r *http.Request, memberId *string, statusFilter *[]string) {
	submissions, err := h.applicationService.GetApplicationSubmissions(memberId, statusFilter)
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

func (h *V1Handler) getApplicationSubmission(w http.ResponseWriter, r *http.Request, submissionId string) {
	submission, err := h.applicationService.GetApplicationSubmission(submissionId)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, submission)
}

func (h *V1Handler) createApplicationSubmission(w http.ResponseWriter, r *http.Request, memberId *string) {
	var req models.CreateApplicationSubmissionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if memberId != nil {
		req.MemberID = *memberId
	}

	submission, err := h.applicationService.CreateApplicationSubmission(&req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, submission)
}

func (h *V1Handler) updateApplicationSubmission(w http.ResponseWriter, r *http.Request, submissionId string) {
	var req models.UpdateApplicationSubmissionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	submission, err := h.applicationService.UpdateApplicationSubmission(submissionId, &req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, submission)
}

func (h *V1Handler) getAllApplications(w http.ResponseWriter, r *http.Request, memberId *string) {
	applications, err := h.applicationService.GetApplications(memberId)
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

func (h *V1Handler) getApplication(w http.ResponseWriter, r *http.Request, submissionId string) {
	application, err := h.applicationService.GetApplication(submissionId)
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

func (h *V1Handler) updateApplication(w http.ResponseWriter, r *http.Request, applicationId string) {
	var req models.UpdateApplicationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	application, err := h.applicationService.UpdateApplication(applicationId, &req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, application)
}
