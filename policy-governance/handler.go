// handler.go
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"policy-governance/internal/models" // Import the models package
)

// PolicyDataFetcher defines the interface for fetching policy data.
// The GetPolicyFromDB method now accepts a consumerID.
type PolicyDataFetcher interface {
	GetPolicyFromDB(consumerID, subgraph, typ, field string) (models.PolicyRecord, error) // <-- UPDATED SIGNATURE
}

// PolicyGovernanceService handles policy verification.
type PolicyGovernanceService struct {
	Fetcher PolicyDataFetcher
}

// EvaluateAccessPolicy determines the access policy for the requested fields.
func (s *PolicyGovernanceService) EvaluateAccessPolicy(req models.PolicyRequest) models.PolicyResponse {
	resp := models.PolicyResponse{
		ConsumerID:             req.ConsumerID,
		AccessScopes:           []models.AccessScope{},
		OverallConsentRequired: false,
	}

	for _, field := range req.RequestedFields {
		accessScope := models.AccessScope{
			SubgraphName:           field.SubgraphName,
			TypeName:               field.TypeName,
			FieldName:              field.FieldName,
			ResolvedClassification: models.DENY,
		}

		policyFromDB, err := s.Fetcher.GetPolicyFromDB(req.ConsumerID, field.SubgraphName, field.TypeName, field.FieldName)
		if err != nil {
			log.Printf("Error fetching policy for ConsumerID %s, Field %s.%s.%s: %v. Defaulting to DENY.", req.ConsumerID, field.SubgraphName, field.TypeName, field.FieldName, err)
		} else {
			actualClassification := policyFromDB.Classification
			if actualClassification == "" {
				actualClassification = field.Classification
			}

			accessScope.ResolvedClassification = actualClassification

			switch actualClassification {
			case models.ALLOW_PROVIDER_CONSENT, models.ALLOW_CITIZEN_CONSENT, models.ALLOW_CONSENT:
				resp.OverallConsentRequired = true
			}
		}
		resp.AccessScopes = append(resp.AccessScopes, accessScope)
	}

	return resp
}

// DatabasePolicyFetcher is a concrete implementation of PolicyDataFetcher using a SQL database.
type DatabasePolicyFetcher struct {
	DB *sql.DB
}

// GetPolicyFromDB implements the PolicyDataFetcher interface for database lookup.
// The query now includes consumer_id in the WHERE clause and scans it.
func (f *DatabasePolicyFetcher) GetPolicyFromDB(consumerID, subgraph, typ, field string) (models.PolicyRecord, error) { // <-- UPDATED SIGNATURE
	var policy models.PolicyRecord
	// Updated query to include consumer_id
	query := `SELECT id, consumer_id, subgraph_name, type_name, field_name, classification
              FROM policies
              WHERE consumer_id = $1 AND subgraph_name = $2 AND type_name = $3 AND field_name = $4
              LIMIT 1`
	row := f.DB.QueryRow(query, consumerID, subgraph, typ, field)

	// Updated scan to include policy.ConsumerID
	err := row.Scan(&policy.ID, &policy.ConsumerID, &policy.SubgraphName, &policy.TypeName, &policy.FieldName, &policy.Classification) // <-- UPDATED SCAN
	if err == sql.ErrNoRows {
		return models.PolicyRecord{Classification: models.DENY}, nil
	}
	if err != nil {
		return models.PolicyRecord{}, fmt.Errorf("failed to scan policy: %w", err)
	}
	return policy, nil
}

// HandlePolicyRequest is the HTTP handler for policy requests.
// It accepts a pointer to PolicyGovernanceService to use the shared DB connection.
func HandlePolicyRequest(service *PolicyGovernanceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
			return
		}

		var req models.PolicyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request payload: %v", err), http.StatusBadRequest)
			return
		}

		resp := service.EvaluateAccessPolicy(req)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}
