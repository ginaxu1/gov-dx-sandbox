package consent

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
)

type CeConfig struct {
	ClientUrl string `json:"ceUrl,omitempty"`
}

type CERequest struct {
	AppId               string               `json:"app_id"`
	ConsentRequirements []ConsentRequirement `json:"consent_requirements"`
	Purpose             string               `json:"purpose"`
	SessionId           string               `json:"session_id"`
}

type ConsentRequirement struct {
	Owner   string         `json:"owner"`
	OwnerID string         `json:"owner_id"`
	Fields  []ConsentField `json:"fields"`
}

type ConsentField struct {
	FieldName string `json:"fieldName"`
	SchemaID  string `json:"schemaId"`
}

type CEResponse struct {
	Status           string `json:"status"`
	ConsentPortalUrl string `json:"consent_portal_url"`
}

type CEClient struct {
	httpClient *http.Client
	baseUrl    string
}

func NewCEClient(baseUrl string) *CEClient {
	return &CEClient{
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
		baseUrl: baseUrl,
	}
}

func (p *CEClient) MakeConsentRequest(request *CERequest) (*CEResponse, error) {
	// Implement the logic to make a Consent request using p.httpClient
	requestBody, err := json.Marshal(request)
	if err != nil {
		// handle error
		logger.Log.Error("Failed to marshal Consent request", "error", err)
		return nil, err
	}

	logger.Log.Info("Making Consent Request to Consent Engine", "url", p.baseUrl+"/consents")
	response, err := p.httpClient.Post(p.baseUrl+"/consents", "application/json", bytes.NewReader(requestBody))

	if err != nil {
		// handle error
		logger.Log.Error("Consent Request Failed", "error", err)
		return nil, err
	}
	defer response.Body.Close()

	// convert to string and log
	logger.Log.Info("Consent Response Status", "status", response.Status)

	var pdpResponse CEResponse
	err = json.NewDecoder(response.Body).Decode(&pdpResponse)

	if err != nil {
		// handle error
		logger.Log.Error("Failed to decode CE response", "error", err)
		return nil, err
	}

	return &pdpResponse, nil
}
