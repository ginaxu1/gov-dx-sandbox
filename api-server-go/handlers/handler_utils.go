package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gov-dx-sandbox/api-server-go/models"
	"github.com/gov-dx-sandbox/exchange/shared/utils"
)

// HandlerFunc represents a function that handles HTTP requests and returns a result or error
type HandlerFunc func(r *http.Request) (interface{}, error)

// RequestParser represents a function that parses a request body into a specific type
type RequestParser func([]byte) (interface{}, error)

// ServiceMethod represents a function that calls a service method
type ServiceMethod func(interface{}) (interface{}, error)

// GenericHandler handles common HTTP patterns with error handling and response formatting
func GenericHandler(w http.ResponseWriter, r *http.Request, handler HandlerFunc) {
	result, err := handler(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, result)
}

// CreateHandler handles POST requests with request parsing and service calls
func CreateHandler(w http.ResponseWriter, r *http.Request, parser RequestParser, serviceMethod ServiceMethod) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Failed to marshal request")
		return
	}

	parsedReq, err := parser(bodyBytes)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := serviceMethod(parsedReq)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, result)
}

// UpdateHandler handles PUT requests with request parsing and service calls
func UpdateHandler(w http.ResponseWriter, r *http.Request, id string, parser RequestParser, serviceMethod ServiceMethod) {
	if r.Method != http.MethodPut {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if id == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "ID is required")
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Failed to marshal request")
		return
	}

	parsedReq, err := parser(bodyBytes)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := serviceMethod(parsedReq)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, result)
}

// GetHandler handles GET requests for retrieving items
func GetHandler(w http.ResponseWriter, r *http.Request, id string, serviceMethod ServiceMethod) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if id == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "ID is required")
		return
	}

	result, err := serviceMethod(id)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, result)
}

// ListHandler handles GET requests for retrieving collections
func ListHandler(w http.ResponseWriter, r *http.Request, serviceMethod ServiceMethod) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	result, err := serviceMethod(nil)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve items")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, result)
}

// DeleteHandler handles DELETE requests
func DeleteHandler(w http.ResponseWriter, r *http.Request, id string, serviceMethod ServiceMethod) {
	if r.Method != http.MethodDelete {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if id == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "ID is required")
		return
	}

	_, err := serviceMethod(id)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Item deleted successfully"})
}

// Request parsers for different types
func ParseCreateConsumerRequest(body []byte) (interface{}, error) {
	var req models.CreateConsumerRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("failed to parse consumer request: %w", err)
	}
	return req, nil
}

func ParseCreateConsumerAppRequest(body []byte) (interface{}, error) {
	var req models.CreateConsumerAppRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("failed to parse consumer app request: %w", err)
	}
	return req, nil
}

func ParseUpdateConsumerAppRequest(body []byte) (interface{}, error) {
	var req models.UpdateConsumerAppRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("failed to parse update consumer app request: %w", err)
	}
	return req, nil
}

func ParseCreateProviderSubmissionRequest(body []byte) (interface{}, error) {
	var req models.CreateProviderSubmissionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("failed to parse provider submission request: %w", err)
	}
	return req, nil
}

func ParseUpdateProviderSubmissionRequest(body []byte) (interface{}, error) {
	var req models.UpdateProviderSubmissionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("failed to parse update provider submission request: %w", err)
	}
	return req, nil
}

func ParseCreateProviderSchemaRequest(body []byte) (interface{}, error) {
	var req models.CreateProviderSchemaRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("failed to parse provider schema request: %w", err)
	}
	return req, nil
}

func ParseCreateProviderSchemaSDLRequest(body []byte) (interface{}, error) {
	var req models.CreateProviderSchemaSDLRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("failed to parse provider schema SDL request: %w", err)
	}
	return req, nil
}

// ExtractFieldNameFromPath extracts field name from URL path like /admin/fields/{fieldName}/allow-list
func ExtractFieldNameFromPath(path string) string {
	// Pattern: /admin/fields/{fieldName}/allow-list or /admin/fields/{fieldName}/allow-list/{consumerId}
	parts := strings.Split(path, "/")
	if len(parts) >= 4 && parts[1] == "admin" && parts[2] == "fields" {
		return parts[3]
	}
	return ""
}

// ExtractConsumerIDFromPath extracts consumer ID from URL path like /admin/fields/{fieldName}/allow-list/{consumerId}
func ExtractConsumerIDFromPath(path string) string {
	// Pattern: /admin/fields/{fieldName}/allow-list/{consumerId}
	parts := strings.Split(path, "/")
	if len(parts) >= 6 && parts[1] == "admin" && parts[2] == "fields" && parts[4] == "allow-list" {
		return parts[5]
	}
	return ""
}
