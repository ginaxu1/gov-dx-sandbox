package policy

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
)

type PdpConfig struct {
	ClientUrl string `json:"pdpUrl,omitempty"`
}

type RequiredField struct {
	ProviderKey string `json:"providerKey"`
	SchemaId    string `json:"schemaId"`
	FieldName   string `json:"fieldName"`
}

type PdpRequest struct {
	ConsumerId     string          `json:"consumerId"`
	AppId          string          `json:"applicationId"`
	RequestId      string          `json:"requestId"`
	RequiredFields []RequiredField `json:"requiredFields"`
}

type ConsentRequiredField struct {
	FieldName   string  `json:"fieldName"`
	SchemaID    string  `json:"schemaId"`
	DisplayName *string `json:"displayName,omitempty"`
	Description *string `json:"description,omitempty"`
	Owner       *string `json:"owner,omitempty"`
}

type PdpResponse struct {
	AppAuthorized         bool                   `json:"appAuthorized"`
	ConsentRequired       bool                   `json:"appRequiresOwnerConsent"`
	ConsentRequiredFields []ConsentRequiredField `json:"consentRequiredFields"`
}

type PdpClient struct {
	httpClient *http.Client
	baseUrl    string
}

func NewPdpClient(baseUrl string) *PdpClient {
	return &PdpClient{
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
		baseUrl: baseUrl,
	}
}

func (p *PdpClient) MakePdpRequest(request *PdpRequest) (*PdpResponse, error) {
	// Implement the logic to make a PDP request using p.httpClient
	requestBody, err := json.Marshal(request)
	if err != nil {
		// handle error
		logger.Log.Error("Failed to marshal PDP request", "error", err)
		return nil, err
	}

	// log the json request body
	logger.Log.Info("PDP Request Body", "body", string(requestBody))

	response, err := p.httpClient.Post(p.baseUrl+"/api/v1/policy/decide", "application/json", bytes.NewReader(requestBody))
	if err != nil {
		// handle error
		logger.Log.Error("Failed to make PDP request", "error", err)
		return nil, err
	}
	defer response.Body.Close()

	var pdpResponse PdpResponse
	err = json.NewDecoder(response.Body).Decode(&pdpResponse)
	if err != nil {
		// handle error
		logger.Log.Error("Failed to decode PDP response", "error", err)
		return nil, err
	}

	return &pdpResponse, nil
}
