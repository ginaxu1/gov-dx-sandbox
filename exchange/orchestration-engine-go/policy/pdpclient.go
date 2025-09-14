package policy

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
)

type PdpConfig struct {
	ClientUrl string `json:"pdpUrl,omitempty"`
}

type PdpRequest struct {
	ConsumerId     string   `json:"consumer_id"`
	AppId          string   `json:"app_id"`
	RequestId      string   `json:"request_id"`
	RequiredFields []string `json:"required_fields"`
}

type PdpResponse struct {
	Allowed               bool     `json:"allow"`
	ConsentRequired       bool     `json:"consent_required"`
	ConsentRequiredFields []string `json:"consent_required_fields"`
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

	response, err := p.httpClient.Post(p.baseUrl+"/decide", "application/json", bytes.NewReader(requestBody))

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
