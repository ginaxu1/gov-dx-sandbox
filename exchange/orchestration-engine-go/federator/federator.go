package federator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Provider struct that represents a provider information.
type Provider struct {
	ServiceUrl string `json:"serviceUrl,omitempty"`
	ServiceKey string `json:"serviceKey,omitempty"`
	ApiKey     string `json:"apiKey,omitempty"`
}

// Federator struct that includes all the context needed for federation.
type Federator struct {
	Providers map[string]*Provider
	Client    *http.Client
}

type FederationServiceRequest struct {
	ServiceKey   string
	GraphqlQuery GraphQLRequest
}

type FederationRequest struct {
	// Define fields as needed
	FederationServiceRequest []*FederationServiceRequest
}

type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

type FederatorOptions struct {
	Providers []*Provider `json:"providers,omitempty"`
}

type ProviderResponse struct {
	ServiceKey string
	Response   GraphQLResponse `json:"response"`
}

type GraphQLResponse struct {
	Data   map[string]interface{} `json:"data,omitempty"`
	Errors []interface{}          `json:"errors,omitempty"`
}

type FederationResponse struct {
	ServiceKey string             `json:"serviceKey"`
	Responses  []ProviderResponse `json:"responses"`
}

func Initialize(options *FederatorOptions) *Federator {
	federator := &Federator{}
	federator.Providers = make(map[string]*Provider)
	// Initialize with options if provided

	if options.Providers != nil {
		for _, provider := range options.Providers {
			// print service url
			fmt.Printf("Adding provider: %s\n", provider.ServiceUrl)
			federator.Providers[provider.ServiceKey] = provider
		}
	}
	federator.Client = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
		},
	}

	return federator
}

func (f *Federator) AddProvider(url string) {
	provider := &Provider{ServiceUrl: url}
	f.Providers[url] = provider
}

func (f *Federator) HandleQuery(request GraphQLRequest) GraphQLResponse {
	splitRequests := splitQuery(request.Query)
	federationRequest := &FederationRequest{
		FederationServiceRequest: splitRequests,
	}
	responses := f.Federate(federationRequest)

	return f.MergeResponses(responses.Responses)
}

func (f *Federator) Federate(r *FederationRequest) *FederationResponse {
	federationResponse := &FederationResponse{
		Responses: make([]ProviderResponse, 0, len(r.FederationServiceRequest)),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex // to safely append to federationResponse.Responses

	for _, request := range r.FederationServiceRequest {
		provider, exists := f.Providers[request.ServiceKey]
		if !exists {
			fmt.Printf("Provider not found: %s\n", request.ServiceKey)
			continue
		}

		wg.Add(1)
		go func(req *FederationServiceRequest, prov *Provider) {
			defer wg.Done()

			reqBody, err := json.Marshal(req.GraphqlQuery)
			if err != nil {
				fmt.Printf("Failed to marshal request for %s: %v\n", req.ServiceKey, err)
				return
			}

			response, err := f.Client.Post(prov.ServiceUrl, "application/json", bytes.NewBuffer(reqBody))
			if err != nil {
				fmt.Printf("Request failed for %s: %v\n", req.ServiceKey, err)
				return
			}
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				fmt.Printf("Failed to read response for %s: %v\n", req.ServiceKey, err)
				return
			}

			var bodyJson GraphQLResponse
			err = json.Unmarshal(body, &bodyJson)
			if err != nil {
				fmt.Printf("Failed to unmarshal response for %s: %v\n", req.ServiceKey, err)
				return
			}

			// Thread-safe append
			mu.Lock()
			federationResponse.Responses = append(federationResponse.Responses, ProviderResponse{
				ServiceKey: req.ServiceKey,
				Response:   bodyJson,
			})
			mu.Unlock()
		}(request, provider)
	}

	wg.Wait()
	return federationResponse
}

func (f *Federator) MergeResponses(responses []ProviderResponse) GraphQLResponse {
	merged := GraphQLResponse{
		Data:   make(map[string]interface{}),
		Errors: make([]interface{}, 0),
	}

	for _, resp := range responses {
		if resp.Response.Data != nil {
			for k, v := range resp.Response.Data {
				// wrap v with service key
				merged.Data[resp.ServiceKey] = map[string]interface{}{
					k: v,
				}
			}
		}
		if resp.Response.Errors != nil {
			merged.Errors = append(merged.Errors, resp.Response.Errors...)
		}
	}

	return merged
}
