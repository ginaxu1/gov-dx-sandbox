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
// Handles the retrieval of access policies based on consumer and provider IDs.
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

	// If no explicit policy is found, default to a restrictive but potentially accessible tier.
	// For example, defaulting to "Confidential" requiring consent.
	if policy == nil {
		policy = &models.PolicyMapping{
			AccessTier:   "Confidential",
			AccessBucket: "requires_consent",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policy)
}