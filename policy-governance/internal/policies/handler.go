// internal/policies/handler.go
package policies

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"policy-governance/internal/models"
	"policy-governance/internal/repository"
)

type PolicyHandler struct {
	repo repository.PolicyRepositoryInterface
}

func NewPolicyHandler(repo repository.PolicyRepositoryInterface) *PolicyHandler {
	return &PolicyHandler{repo: repo}
}

// GetAccessPolicy handles requests to retrieve an access policy.
func (h *PolicyHandler) GetAccessPolicy(w http.ResponseWriter, r *http.Request) {
	consumerID := chi.URLParam(r, "consumerID")
	providerID := chi.URLParam(r, "providerID")

	if consumerID == "" || providerID == "" {
		http.Error(w, "consumerID and providerID must be provided", http.StatusBadRequest)
		return
	}

	policy, err := h.repo.GetPolicy(r.Context(), consumerID, providerID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Default to a restrictive policy if no explicit mapping is found
	if policy == nil {
		policy = &models.PolicyMapping{
			AccessTier:   "Tier 2",
			AccessBucket: "require_consent",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policy)
}