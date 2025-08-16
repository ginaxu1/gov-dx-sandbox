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
// This abstraction allows for easy mocking in tests.
type PolicyDataFetcher interface {
	GetPolicyFromDB(subgraph, typ, field string) (models.PolicyRecord, error)
}

// PolicyGovernanceService handles policy verification.
// It now embeds the PolicyDataFetcher interface.
type PolicyGovernanceService struct {
	Fetcher PolicyDataFetcher // Use the interface for fetching data
}

// GetPolicyFromDB is now part of the PolicyDataFetcher interface.
// The actual implementation (using *sql.DB) will be in a concrete type that implements this.
// For example, a `DatabasePolicyFetcher` struct.

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
			ResolvedClassification: models.DENIED, // Default to DENIED
			ConsentRequired:        false,
			ConsentType:            "",
		}

		// Fetch policy from the database using the Fetcher interface
		policyFromDB, err := s.Fetcher.GetPolicyFromDB(field.SubgraphName, field.TypeName, field.FieldName)
		if err != nil {
			log.Printf("Error fetching policy for %s.%s.%s: %v. Defaulting to DENIED.", field.SubgraphName, field.TypeName, field.FieldName, err)
			// Keep default DENIED
		} else {
			// Use the classification from the database, or fall back to requested if no specific policy
			actualClassification := policyFromDB.Classification
			if actualClassification == "" { // If DB didn't have a specific policy, use the requested one
				actualClassification = field.Classification
			}

			switch actualClassification {
			case models.ALLOW:
				accessScope.ResolvedClassification = models.ALLOW
			case models.ALLOW_PROVIDER_CONSENT:
				accessScope.ResolvedClassification = models.ALLOW_PROVIDER_CONSENT
				accessScope.ConsentRequired = true
				accessScope.ConsentType = "provider"
				resp.OverallConsentRequired = true
			case models.ALLOW_CITIZEN_CONSENT:
				accessScope.ResolvedClassification = models.ALLOW_CITIZEN_CONSENT
				accessScope.ConsentRequired = true
				accessScope.ConsentType = "citizen"
				resp.OverallConsentRequired = true
			case models.ALLOW_CONSENT:
				accessScope.ResolvedClassification = models.ALLOW_CONSENT
				accessScope.ConsentRequired = true
				if _, ok := field.Context["citizenId"]; ok {
					accessScope.ConsentType = "citizen"
				} else {
					accessScope.ConsentType = "provider"
				}
				resp.OverallConsentRequired = true
			case models.DENIED:
				accessScope.ResolvedClassification = models.DENIED
			default:
				accessScope.ResolvedClassification = models.DENIED
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
func (f *DatabasePolicyFetcher) GetPolicyFromDB(subgraph, typ, field string) (models.PolicyRecord, error) {
	var policy models.PolicyRecord
	query := `SELECT id, subgraph_name, type_name, field_name, classification FROM policies WHERE subgraph_name = $1 AND type_name = $2 AND field_name = $3 LIMIT 1`
	row := f.DB.QueryRow(query, subgraph, typ, field)

	err := row.Scan(&policy.ID, &policy.SubgraphName, &policy.TypeName, &policy.FieldName, &policy.Classification)
	if err == sql.ErrNoRows {
		return models.PolicyRecord{Classification: models.DENIED}, nil // Default if not found
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
