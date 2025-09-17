package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/handlers"
	"github.com/gov-dx-sandbox/api-server-go/models"
)

func TestGenericHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		handler := func(r *http.Request) (interface{}, error) {
			return map[string]string{"message": "success"}, nil
		}

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handlers.GenericHandler(w, req, handler)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["message"] != "success" {
			t.Errorf("Expected message 'success', got %s", response["message"])
		}
	})

	t.Run("Error", func(t *testing.T) {
		handler := func(r *http.Request) (interface{}, error) {
			return nil, http.ErrBodyReadAfterClose
		}

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handlers.GenericHandler(w, req, handler)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestCreateHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		parser := func(body []byte) (interface{}, error) {
			var req models.CreateConsumerRequest
			if err := json.Unmarshal(body, &req); err != nil {
				return nil, err
			}
			return req, nil
		}

		serviceMethod := func(req interface{}) (interface{}, error) {
			consumerReq := req.(models.CreateConsumerRequest)
			return map[string]string{
				"consumerId":   "consumer_123",
				"consumerName": consumerReq.ConsumerName,
			}, nil
		}

		reqBody := map[string]string{
			"consumerName": "Test Consumer",
			"contactEmail": "test@example.com",
			"phoneNumber":  "1234567890",
		}

		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/consumers", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.CreateHandler(w, req, parser, serviceMethod)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["consumerId"] != "consumer_123" {
			t.Errorf("Expected consumerId 'consumer_123', got %s", response["consumerId"])
		}
	})

	t.Run("WrongMethod", func(t *testing.T) {
		parser := func(body []byte) (interface{}, error) {
			return nil, nil
		}

		serviceMethod := func(req interface{}) (interface{}, error) {
			return nil, nil
		}

		req := httptest.NewRequest("GET", "/consumers", nil)
		w := httptest.NewRecorder()

		handlers.CreateHandler(w, req, parser, serviceMethod)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		parser := func(body []byte) (interface{}, error) {
			return nil, nil
		}

		serviceMethod := func(req interface{}) (interface{}, error) {
			return nil, nil
		}

		req := httptest.NewRequest("POST", "/consumers", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.CreateHandler(w, req, parser, serviceMethod)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("ParserError", func(t *testing.T) {
		parser := func(body []byte) (interface{}, error) {
			return nil, http.ErrBodyReadAfterClose
		}

		serviceMethod := func(req interface{}) (interface{}, error) {
			return nil, nil
		}

		reqBody := map[string]string{"test": "data"}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/consumers", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.CreateHandler(w, req, parser, serviceMethod)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("ServiceError", func(t *testing.T) {
		parser := func(body []byte) (interface{}, error) {
			var req models.CreateConsumerRequest
			json.Unmarshal(body, &req)
			return req, nil
		}

		serviceMethod := func(req interface{}) (interface{}, error) {
			return nil, http.ErrBodyReadAfterClose
		}

		reqBody := map[string]string{
			"consumerName": "Test Consumer",
			"contactEmail": "test@example.com",
			"phoneNumber":  "1234567890",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/consumers", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.CreateHandler(w, req, parser, serviceMethod)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestUpdateHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		parser := func(body []byte) (interface{}, error) {
			var req models.UpdateConsumerRequest
			if err := json.Unmarshal(body, &req); err != nil {
				return nil, err
			}
			return req, nil
		}

		serviceMethod := func(req interface{}) (interface{}, error) {
			updateReq := req.(models.UpdateConsumerRequest)
			return map[string]string{
				"consumerId":   "consumer_123",
				"consumerName": *updateReq.ConsumerName,
			}, nil
		}

		reqBody := map[string]string{
			"consumerName": "Updated Consumer",
		}

		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("PUT", "/consumers/consumer_123", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.UpdateHandler(w, req, "consumer_123", parser, serviceMethod)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["consumerName"] != "Updated Consumer" {
			t.Errorf("Expected consumerName 'Updated Consumer', got %s", response["consumerName"])
		}
	})

	t.Run("WrongMethod", func(t *testing.T) {
		parser := func(body []byte) (interface{}, error) {
			return nil, nil
		}

		serviceMethod := func(req interface{}) (interface{}, error) {
			return nil, nil
		}

		req := httptest.NewRequest("GET", "/consumers/123", nil)
		w := httptest.NewRecorder()

		handlers.UpdateHandler(w, req, "123", parser, serviceMethod)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("MissingID", func(t *testing.T) {
		parser := func(body []byte) (interface{}, error) {
			return nil, nil
		}

		serviceMethod := func(req interface{}) (interface{}, error) {
			return nil, nil
		}

		req := httptest.NewRequest("PUT", "/consumers/", nil)
		w := httptest.NewRecorder()

		handlers.UpdateHandler(w, req, "", parser, serviceMethod)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestGetHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		serviceMethod := func(id interface{}) (interface{}, error) {
			return map[string]string{
				"consumerId":   id.(string),
				"consumerName": "Test Consumer",
			}, nil
		}

		req := httptest.NewRequest("GET", "/consumers/consumer_123", nil)
		w := httptest.NewRecorder()

		handlers.GetHandler(w, req, "consumer_123", serviceMethod)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["consumerId"] != "consumer_123" {
			t.Errorf("Expected consumerId 'consumer_123', got %s", response["consumerId"])
		}
	})

	t.Run("WrongMethod", func(t *testing.T) {
		serviceMethod := func(id interface{}) (interface{}, error) {
			return nil, nil
		}

		req := httptest.NewRequest("POST", "/consumers/123", nil)
		w := httptest.NewRecorder()

		handlers.GetHandler(w, req, "123", serviceMethod)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("MissingID", func(t *testing.T) {
		serviceMethod := func(id interface{}) (interface{}, error) {
			return nil, nil
		}

		req := httptest.NewRequest("GET", "/consumers/", nil)
		w := httptest.NewRecorder()

		handlers.GetHandler(w, req, "", serviceMethod)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("ServiceError", func(t *testing.T) {
		serviceMethod := func(id interface{}) (interface{}, error) {
			return nil, http.ErrBodyReadAfterClose
		}

		req := httptest.NewRequest("GET", "/consumers/nonexistent", nil)
		w := httptest.NewRecorder()

		handlers.GetHandler(w, req, "nonexistent", serviceMethod)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestListHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		serviceMethod := func(req interface{}) (interface{}, error) {
			return []map[string]string{
				{"consumerId": "consumer_1", "consumerName": "Consumer 1"},
				{"consumerId": "consumer_2", "consumerName": "Consumer 2"},
			}, nil
		}

		req := httptest.NewRequest("GET", "/consumers", nil)
		w := httptest.NewRecorder()

		handlers.ListHandler(w, req, serviceMethod)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response []map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if len(response) != 2 {
			t.Errorf("Expected 2 items, got %d", len(response))
		}
	})

	t.Run("WrongMethod", func(t *testing.T) {
		serviceMethod := func(req interface{}) (interface{}, error) {
			return nil, nil
		}

		req := httptest.NewRequest("POST", "/consumers", nil)
		w := httptest.NewRecorder()

		handlers.ListHandler(w, req, serviceMethod)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("ServiceError", func(t *testing.T) {
		serviceMethod := func(req interface{}) (interface{}, error) {
			return nil, http.ErrBodyReadAfterClose
		}

		req := httptest.NewRequest("GET", "/consumers", nil)
		w := httptest.NewRecorder()

		handlers.ListHandler(w, req, serviceMethod)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})
}

func TestDeleteHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		serviceMethod := func(id interface{}) (interface{}, error) {
			return map[string]string{"message": "deleted"}, nil
		}

		req := httptest.NewRequest("DELETE", "/consumers/consumer_123", nil)
		w := httptest.NewRecorder()

		handlers.DeleteHandler(w, req, "consumer_123", serviceMethod)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["message"] != "Item deleted successfully" {
			t.Errorf("Expected message 'Item deleted successfully', got %s", response["message"])
		}
	})

	t.Run("WrongMethod", func(t *testing.T) {
		serviceMethod := func(id interface{}) (interface{}, error) {
			return nil, nil
		}

		req := httptest.NewRequest("GET", "/consumers/123", nil)
		w := httptest.NewRecorder()

		handlers.DeleteHandler(w, req, "123", serviceMethod)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("MissingID", func(t *testing.T) {
		serviceMethod := func(id interface{}) (interface{}, error) {
			return nil, nil
		}

		req := httptest.NewRequest("DELETE", "/consumers/", nil)
		w := httptest.NewRecorder()

		handlers.DeleteHandler(w, req, "", serviceMethod)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("ServiceError", func(t *testing.T) {
		serviceMethod := func(id interface{}) (interface{}, error) {
			return nil, http.ErrBodyReadAfterClose
		}

		req := httptest.NewRequest("DELETE", "/consumers/nonexistent", nil)
		w := httptest.NewRecorder()

		handlers.DeleteHandler(w, req, "nonexistent", serviceMethod)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestRequestParsers(t *testing.T) {
	t.Run("ParseCreateConsumerRequest", func(t *testing.T) {
		reqData := models.CreateConsumerRequest{
			ConsumerName: "Test Consumer",
			ContactEmail: "test@example.com",
			PhoneNumber:  "1234567890",
		}

		bodyBytes, err := json.Marshal(reqData)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		parsed, err := handlers.ParseCreateConsumerRequest(bodyBytes)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		req := parsed.(models.CreateConsumerRequest)
		if req.ConsumerName != reqData.ConsumerName {
			t.Errorf("Expected consumer name %s, got %s", reqData.ConsumerName, req.ConsumerName)
		}
	})

	t.Run("ParseCreateConsumerAppRequest", func(t *testing.T) {
		reqData := models.CreateConsumerAppRequest{
			ConsumerID:     "consumer_123",
			RequiredFields: map[string]bool{"person.fullName": true},
		}

		bodyBytes, err := json.Marshal(reqData)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		parsed, err := handlers.ParseCreateConsumerAppRequest(bodyBytes)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		req := parsed.(models.CreateConsumerAppRequest)
		if req.ConsumerID != reqData.ConsumerID {
			t.Errorf("Expected consumer ID %s, got %s", reqData.ConsumerID, req.ConsumerID)
		}
	})

	t.Run("ParseUpdateConsumerAppRequest", func(t *testing.T) {
		status := models.StatusApproved
		reqData := models.UpdateConsumerAppRequest{
			Status: &status,
		}

		bodyBytes, err := json.Marshal(reqData)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		parsed, err := handlers.ParseUpdateConsumerAppRequest(bodyBytes)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		req := parsed.(models.UpdateConsumerAppRequest)
		if *req.Status != status {
			t.Errorf("Expected status %s, got %s", status, *req.Status)
		}
	})

	t.Run("ParseCreateProviderSubmissionRequest", func(t *testing.T) {
		reqData := models.CreateProviderSubmissionRequest{
			ProviderName: "Test Provider",
			ContactEmail: "provider@example.com",
			PhoneNumber:  "1234567890",
			ProviderType: models.ProviderTypeGovernment,
		}

		bodyBytes, err := json.Marshal(reqData)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		parsed, err := handlers.ParseCreateProviderSubmissionRequest(bodyBytes)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		req := parsed.(models.CreateProviderSubmissionRequest)
		if req.ProviderName != reqData.ProviderName {
			t.Errorf("Expected provider name %s, got %s", reqData.ProviderName, req.ProviderName)
		}
	})

	t.Run("ParseUpdateProviderSubmissionRequest", func(t *testing.T) {
		status := models.SubmissionStatusApproved
		reqData := models.UpdateProviderSubmissionRequest{
			Status: &status,
		}

		bodyBytes, err := json.Marshal(reqData)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		parsed, err := handlers.ParseUpdateProviderSubmissionRequest(bodyBytes)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		req := parsed.(models.UpdateProviderSubmissionRequest)
		if *req.Status != status {
			t.Errorf("Expected status %s, got %s", status, *req.Status)
		}
	})

	t.Run("ParseCreateProviderSchemaRequest", func(t *testing.T) {
		reqData := models.CreateProviderSchemaRequest{
			ProviderID: "provider_123",
			FieldConfigurations: models.FieldConfigurations{
				"PersonData": {
					"fullName": {
						Source:      "authoritative",
						IsOwner:     true,
						Description: "Full name",
					},
				},
			},
		}

		bodyBytes, err := json.Marshal(reqData)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		parsed, err := handlers.ParseCreateProviderSchemaRequest(bodyBytes)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		req := parsed.(models.CreateProviderSchemaRequest)
		if req.ProviderID != reqData.ProviderID {
			t.Errorf("Expected provider ID %s, got %s", reqData.ProviderID, req.ProviderID)
		}
	})

	t.Run("ParseCreateProviderSchemaSDLRequest", func(t *testing.T) {
		reqData := models.CreateProviderSchemaSDLRequest{
			SDL: "type Query { test: String }",
		}

		bodyBytes, err := json.Marshal(reqData)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		parsed, err := handlers.ParseCreateProviderSchemaSDLRequest(bodyBytes)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		req := parsed.(models.CreateProviderSchemaSDLRequest)
		if req.SDL != reqData.SDL {
			t.Errorf("Expected SDL %s, got %s", reqData.SDL, req.SDL)
		}
	})

	t.Run("ParserError", func(t *testing.T) {
		invalidJSON := []byte("invalid json")

		_, err := handlers.ParseCreateConsumerRequest(invalidJSON)
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})
}

func TestPathExtractors(t *testing.T) {
	t.Run("ExtractFieldNameFromPath", func(t *testing.T) {
		tests := []struct {
			path     string
			expected string
		}{
			{"/admin/fields/person.fullName/allow-list", "person.fullName"},
			{"/admin/fields/person.email/allow-list/consumer_123", "person.email"},
			{"/admin/fields/test.field/allow-list", "test.field"},
			{"/admin/fields/", ""},
			{"/admin/fields", ""},
			{"/admin/other/person.fullName", ""},
			{"/fields/person.fullName", ""},
		}

		for _, test := range tests {
			result := handlers.ExtractFieldNameFromPath(test.path)
			if result != test.expected {
				t.Errorf("For path '%s', expected '%s', got '%s'", test.path, test.expected, result)
			}
		}
	})

	t.Run("ExtractConsumerIDFromPath", func(t *testing.T) {
		tests := []struct {
			path     string
			expected string
		}{
			{"/admin/fields/person.fullName/allow-list/consumer_123", "consumer_123"},
			{"/admin/fields/person.email/allow-list/consumer_456", "consumer_456"},
			{"/admin/fields/test.field/allow-list/test_consumer", "test_consumer"},
			{"/admin/fields/person.fullName/allow-list", ""},
			{"/admin/fields/person.fullName", ""},
			{"/admin/fields", ""},
			{"/admin/other/person.fullName/allow-list/consumer_123", ""},
			{"/fields/person.fullName/allow-list/consumer_123", ""},
		}

		for _, test := range tests {
			result := handlers.ExtractConsumerIDFromPath(test.path)
			if result != test.expected {
				t.Errorf("For path '%s', expected '%s', got '%s'", test.path, test.expected, result)
			}
		}
	})
}
