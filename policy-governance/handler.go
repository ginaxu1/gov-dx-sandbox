// handler.go
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings" // Import for string manipulation (e.g., StartsWith)

	"policy-governance/internal/models" // Import the models package

	"github.com/graphql-go/graphql/language/ast"    // GraphQL AST types
	"github.com/graphql-go/graphql/language/parser" // GraphQL parser
)

// GraphQLRequestBody mirrors the JSON structure sent by Apollo Router
// when `send_graphql_request_body: true` is set.
type GraphQLRequestBody struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
	OperationName string                 `json:"operationName"`
}

// typeToSubgraphMap maps GraphQL type names to their owning subgraph names.
// This is critical for Apollo Router's external authorization service
// to correctly identify the subgraph context for policy evaluation.
// This map MUST be kept in sync with your actual subgraph schemas and definitions.
var typeToSubgraphMap = map[string]string{
	// DRP Subgraph Types (from drp/types.bal)
	"PersonData":              "drp",
	"PersonInfo":              "drp",
	"CardInfo":                "drp",
	"LostCardReplacementInfo": "drp",
	"CitizenshipInfo":         "drp",
	"ParentInfo":              "drp",
	"Gender":                  "drp",
	"CardStatus":              "drp",
	"CivilStatus":             "drp",
	"CitizenshipType":         "drp",
	"person":                  "drp", // Assuming 'person' (lowercase) is a type name for PersonData

	// DMT Subgraph Types (from dmt/types.bal)
	"VehicleClass":  "dmt",
	"VehicleInfo":   "dmt",
	"DriverLicense": "dmt",
	"Vehicle":       "dmt", // Assuming 'Vehicle' (capitalized) might be a return type for Query.vehicle
	"vehicle":       "dmt", // Assuming 'vehicle' (lowercase) might be a return type
}

// PolicyDataFetcher defines the interface for fetching policy data.
type PolicyDataFetcher interface {
	GetPolicyFromDB(consumerID, subgraph, typ, field string) (models.PolicyRecord, error)
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
			ResolvedClassification: models.DENY, // Default to DENY for safety
		}

		policyFromDB, err := s.Fetcher.GetPolicyFromDB(req.ConsumerID, field.SubgraphName, field.TypeName, field.FieldName)
		if err != nil {
			// Log error but proceed, using the DENY default
			log.Printf("Error fetching policy for ConsumerID %s, Field %s.%s.%s: %v. Defaulting to DENY.", req.ConsumerID, field.SubgraphName, field.TypeName, field.FieldName, err)
		} else {
			// If policy found, use its classification. If not, the default DENY remains.
			accessScope.ResolvedClassification = policyFromDB.Classification
			// Check if consent is required based on the resolved classification from DB
			switch policyFromDB.Classification {
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
func (f *DatabasePolicyFetcher) GetPolicyFromDB(consumerID, subgraph, typ, field string) (models.PolicyRecord, error) {
	var policy models.PolicyRecord
	query := `SELECT id, consumer_id, subgraph_name, type_name, field_name, classification
			  FROM policies
			  WHERE consumer_id = $1 AND subgraph_name = $2 AND type_name = $3 AND field_name = $4
			  LIMIT 1`
	row := f.DB.QueryRow(query, consumerID, subgraph, typ, field)

	err := row.Scan(&policy.ID, &policy.ConsumerID, &policy.SubgraphName, &policy.TypeName, &policy.FieldName, &policy.Classification)
	if err == sql.ErrNoRows {
		// If no specific policy found, explicitly return DENY (or your desired default)
		// This ensures policies that aren't defined are denied by default.
		return models.PolicyRecord{Classification: models.DENY}, nil
	}
	if err != nil {
		return models.PolicyRecord{}, fmt.Errorf("failed to scan policy: %w", err)
	}
	return policy, nil
}

// parseGraphQLQuery parses the GraphQL query string and extracts requested fields
// with their subgraph and type context.
// It is now designed to only support 'query' operations.
func parseGraphQLQuery(queryString string, schema *ast.Document) ([]models.RequestedField, error) {
	if queryString == "" {
		return []models.RequestedField{}, nil
	}

	// Parse the GraphQL query string into an AST
	doc, err := parser.Parse(parser.ParseParams{Source: queryString})
	if err != nil {
		return nil, fmt.Errorf("failed to parse GraphQL query: %w", err)
	}

	requestedFields := []models.RequestedField{}
	visitedFields := make(map[string]struct{}) // To prevent duplicate entries

	// Helper function for recursive AST traversal
	var collectFields func(selections []ast.Selection, currentTypeName string)
	collectFields = func(selections []ast.Selection, currentTypeName string) {
		for _, sel := range selections {
			switch node := sel.(type) {
			case *ast.Field:
				fieldName := node.Name.Value
				if strings.HasPrefix(fieldName, "__") { // Skip introspection fields
					continue
				}

				// Determine subgraph name for the current field
				subgraphName := typeToSubgraphMap[currentTypeName]
				if subgraphName == "" {
					// Special handling for root Query fields that return specific subgraph types
					switch fieldName {
					case "getPersonDataByNic", "overallConsentStatus", "person", "getPersonByNic":
						subgraphName = "drp"
					case "vehicle", "getVehicleById", "vehicleInfoById", "driverLicenseById", "driverLicensesByOwnerId", "vehicleClasses", "vehicleClassById", "getVehicleInfos":
						subgraphName = "dmt"
					default:
						log.Printf("Warning: Unknown subgraph for type '%s' and field '%s'. Defaulting to 'unknown'. Please update typeToSubgraphMap.", currentTypeName, fieldName)
						subgraphName = "unknown" // Fallback
					}
				}

				fieldIdentifier := fmt.Sprintf("%s.%s.%s", subgraphName, currentTypeName, fieldName)
				if _, ok := visitedFields[fieldIdentifier]; !ok {
					visitedFields[fieldIdentifier] = struct{}{}
					requestedFields = append(requestedFields, models.RequestedField{
						SubgraphName: subgraphName,
						TypeName:     currentTypeName,
						FieldName:    fieldName,
						Context:      map[string]interface{}{}, // Empty context for now
					})
				}

				// Recursively collect nested selections
				if node.SelectionSet != nil {
					// Determine the type name for the next level of recursion (the type of the current field's value)
					nextTypeName := ""
					// This mapping is based on the expected return types of fields in your schema.
					// This is still a heuristic without a full GraphQL schema object for type resolution.
					switch fieldName {
					case "person": // If Query.person returns a type named 'person'
						nextTypeName = "person"
					case "getPersonByNic": // If 'person.getPersonByNic' returns 'PersonData'
						nextTypeName = "PersonData"
					case "vehicle": // If Query.vehicle returns a type named 'Vehicle' (capitalized)
						nextTypeName = "Vehicle"
					case "vehicleInfoById", "getVehicleInfos": // If 'Vehicle.vehicleInfoById' or 'Query.getVehicleInfos' returns 'VehicleInfo'
						nextTypeName = "VehicleInfo"
					case "license", "driverLicenseById", "driverLicensesByOwnerId": // If these return 'DriverLicense'
						nextTypeName = "DriverLicense"
					case "cardInfo":
						nextTypeName = "CardInfo"
					case "citizenshipInfo":
						nextTypeName = "CitizenshipInfo"
					case "parentInfo":
						nextTypeName = "ParentInfo"
					case "vehicleClass", "vehicleClassById", "vehicleClasses":
						nextTypeName = "VehicleClass"
					default:
						// Fallback: If no direct mapping for the field's return type,
						// try to infer from the field's name (camelCase to PascalCase)
						// or default to currentTypeName if it's a known complex type.
						if len(fieldName) > 0 {
							tempTypeName := strings.ToUpper(fieldName[:1]) + fieldName[1:]
							if _, ok := typeToSubgraphMap[tempTypeName]; ok {
								nextTypeName = tempTypeName
							} else {
								// If the field name itself is a known type (e.g., 'vehicle' for type 'vehicle')
								if _, ok := typeToSubgraphMap[fieldName]; ok {
									nextTypeName = fieldName
								} else {
									log.Printf("Warning: Could not infer nextTypeName for field '%s' on type '%s'. Nested fields might be misclassified.", fieldName, currentTypeName)
									// If no specific inference, and it has selections, it must be a complex type.
									// For now, we'll let it proceed with the warning, as the subsequent
									// fields will have their own currentTypeName context.
								}
							}
						}
					}

					if nextTypeName != "" {
						collectFields(node.SelectionSet.Selections, nextTypeName)
					} else {
						// If nextTypeName could not be determined, but there are nested selections,
						// we still need to recurse. The currentTypeName for the nested fields
						// will be 'unknown' or a fallback, which might lead to incorrect subgraph mapping.
						// This is a limitation without a full GraphQL schema object for type resolution.
						collectFields(node.SelectionSet.Selections, currentTypeName) // Fallback: use current type name, might be inaccurate
					}
				}

			case *ast.InlineFragment:
				if node.TypeCondition != nil {
					collectFields(node.SelectionSet.Selections, node.TypeCondition.Name.Value)
				}
			case *ast.FragmentSpread:
				// Fragment spreads would require access to the entire document's fragment definitions,
				// which are available in the top-level AST document.
				// For this example, we assume `parseGraphQLQuery` could receive or access them.
				// For simplicity in this direct traversal, we'll log a warning.
				log.Printf("Warning: FragmentSpread '%s' encountered. Full fragment resolution for policy is complex and not fully implemented in this AST traversal.", node.Name.Value)
				// To properly resolve, you'd need to find the fragment definition in the main AST document
				// and recursively call collectFields on its SelectionSet with its TypeCondition.
			default:
				log.Printf("Warning: Unknown selection type: %T", sel)
			}
		}
	}

	// Start traversal from the operation definition (Query, Mutation, Subscription)
	for _, def := range doc.Definitions {
		if opDef, ok := def.(*ast.OperationDefinition); ok {
			// Only process 'query' operations as per the requirement
			if opDef.Operation == "query" {
				rootTypeName := "Query" // For 'query' operations, the root type is always 'Query'
				collectFields(opDef.SelectionSet.Selections, rootTypeName)
			} else {
				log.Printf("Info: Skipping non-query operation type '%s'. Only 'query' operations are supported for policy evaluation.", opDef.Operation)
			}
		}
	}

	return requestedFields, nil
}

// HandlePolicyRequest is the HTTP handler for policy requests.
// It accepts a pointer to PolicyGovernanceService to use the shared DB connection.
func HandlePolicyRequest(service *PolicyGovernanceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
			return
		}

		// Decode the incoming GraphQL request body from Apollo Router
		var graphqlReqBody GraphQLRequestBody
		if err := json.NewDecoder(r.Body).Decode(&graphqlReqBody); err != nil {
			http.Error(w, fmt.Sprintf("Invalid GraphQL request payload: %v", err), http.StatusBadRequest)
			return
		}

		// Extract consumerId from headers, as it's no longer in the body
		consumerID := r.Header.Get("x-consumer-id")
		if consumerID == "" {
			consumerID = "anonymous-consumer" // Default if not provided
		}

		// Parse the GraphQL query string to get requested fields
		requestedFields, err := parseGraphQLQuery(graphqlReqBody.Query, nil) // schema is not needed here
		if err != nil {
			log.Printf("Error parsing GraphQL query: %v", err)
			http.Error(w, fmt.Sprintf("Error parsing GraphQL query: %v", err), http.StatusInternalServerError)
			return
		}

		// Construct the PolicyRequest for evaluation
		policyRequest := models.PolicyRequest{
			ConsumerID:      consumerID,
			RequestedFields: requestedFields,
		}

		resp := service.EvaluateAccessPolicy(policyRequest)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}
