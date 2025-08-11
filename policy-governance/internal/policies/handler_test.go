package policies

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"reflect" // New import for deep comparison

	"github.com/go-chi/chi/v5"

	"policy-governance/internal/models"
)

// MockPolicyRepository is a mock for the repository to isolate handler tests
type MockPolicyRepository struct {
	policy *models.PolicyMapping
	err    error
}

func (m *MockPolicyRepository) GetPolicy(ctx context.Context, consumerID, providerID string) (*models.PolicyMapping, error) {
	return m.policy, m.err
}

func TestGetAccessPolicy_Success(t *testing.T) {
	mockPolicy := &models.PolicyMapping{
		AccessTier:   "Tier 2",
		AccessBucket: "govt_access",
	}
	mockRepo := &MockPolicyRepository{policy: mockPolicy}
	handler := NewPolicyHandler(mockRepo)

	req := httptest.NewRequest("GET", "/policies/access-policy/dmt_service_id/drp_service", nil)
	rr := httptest.NewRecorder()

	// Use chi context to simulate URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("consumerID", "dmt_service_id")
	rctx.URLParams.Add("providerID", "drp_service")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.GetAccessPolicy(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expectedBody, _ := json.Marshal(mockPolicy)
	var expectedJSON, actualJSON map[string]interface{}
	json.Unmarshal(expectedBody, &expectedJSON)
	json.Unmarshal(rr.Body.Bytes(), &actualJSON)

	if !reflect.DeepEqual(expectedJSON, actualJSON) {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), string(expectedBody))
	}
}

func TestGetAccessPolicy_DefaultPolicy(t *testing.T) {
	// Repository returns nil for no policy found
	mockRepo := &MockPolicyRepository{policy: nil}
	handler := NewPolicyHandler(mockRepo)

	req := httptest.NewRequest("GET", "/policies/access-policy/non_existent/provider", nil)
	rr := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("consumerID", "non_existent")
	rctx.URLParams.Add("providerID", "provider")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.GetAccessPolicy(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expectedDefaultPolicy := &models.PolicyMapping{
		AccessTier:   "Tier 2",
		AccessBucket: "require_consent",
	}
	expectedBody, _ := json.Marshal(expectedDefaultPolicy)
	var expectedJSON, actualJSON map[string]interface{}
	json.Unmarshal(expectedBody, &expectedJSON)
	json.Unmarshal(rr.Body.Bytes(), &actualJSON)

	if !reflect.DeepEqual(expectedJSON, actualJSON) {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), string(expectedBody))
	}
}

func TestGetAccessPolicy_InvalidRequest(t *testing.T) {
	handler := NewPolicyHandler(&MockPolicyRepository{})

	// No consumerID and providerID in URL
	req := httptest.NewRequest("GET", "/policies/access-policy//", nil)
	rr := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.GetAccessPolicy(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestGetAccessPolicy_RepoError(t *testing.T) {
	// Mock repo returns an error
	mockRepo := &MockPolicyRepository{err: errors.New("database connection failed")}
	handler := NewPolicyHandler(mockRepo)

	req := httptest.NewRequest("GET", "/policies/access-policy/test_consumer/test_provider", nil)
	rr := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("consumerID", "test_consumer")
	rctx.URLParams.Add("providerID", "test_provider")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.GetAccessPolicy(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}