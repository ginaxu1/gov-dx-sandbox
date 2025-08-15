package policies

import (
	"encoding/json"
	"net/http"

	"policy-governance/internal/database"
	"policy-governance/internal/models"
)

func PolicyHandler(w http.ResponseWriter, r *http.Request) {
	consumerID := r.Header.Get("X-Consumer-Id")
	if consumerID == "" {
		sendResponse(w, false, "X-Consumer-Id header is missing")
		return
	}

	var reqBody models.RequestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		sendResponse(w, false, "Invalid request body: "+err.Error())
		return
	}

	if reqBody.ConsumerID != "" && reqBody.ConsumerID != consumerID {
		sendResponse(w, false, "Consumer ID mismatch between header and body")
		return
	}

	consumerPolicy, err := database.GetConsumerPolicy(consumerID)
	if err != nil {
		sendResponse(w, false, "Failed to retrieve policy for consumer: "+consumerID)
		return
	}

	for _, requestedField := range reqBody.RequestedFields {
		subgraph := requestedField.SubgraphName
		typeName := requestedField.TypeName
		fieldName := requestedField.FieldName

		subgraphPolicies, ok := consumerPolicy.Subgraphs[subgraph]
		if !ok {
			sendResponse(w, false, "Unauthorized access: Subgraph '"+subgraph+"' is not allowed.")
			return
		}

		allowedFields, ok := subgraphPolicies[typeName]
		if !ok {
			sendResponse(w, false, "Unauthorized access: Type '"+typeName+"' in subgraph '"+subgraph+"' is not allowed.")
			return
		}

		if !contains(allowedFields, fieldName) {
			sendResponse(w, false, "Unauthorized access: Field '"+fieldName+"' in type '"+typeName+"' of subgraph '"+subgraph+"' is not allowed.")
			return
		}

		switch requestedField.Classification {
		case "ALLOW_PROVIDER_CONSENT":
		case "ALLOW_CITIZEN_CONSENT":
		case "ALLOW_CONSENT":
		case "DENY":
			sendResponse(w, false, "Access to field '"+requestedField.FieldName+"' is explicitly denied.")
			return
		case "ALLOW":
		}
	}

	sendResponse(w, true, "Request is authorized")
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func sendResponse(w http.ResponseWriter, authorized bool, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := models.ResponseBody{
		Authorized: authorized,
		Message:    message,
	}
	json.NewEncoder(w).Encode(response)
}
