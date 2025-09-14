package federator

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/configs"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/federator"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/graphql"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/provider"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

// Federator struct that includes all the context needed for federation.
type Federator struct {
	Providers map[string]*provider.Provider
	Client    *http.Client
	Schema    *ast.Document
}

type federationServiceAST struct {
	ServiceKey string
	QueryAst   *ast.Document
}

type federationServiceRequest struct {
	ServiceKey     string
	GraphQLRequest graphql.Request
}

type federationRequest struct {
	// Define fields as needed
	FederationServiceRequest []*federationServiceRequest
}

type providerResponse struct {
	ServiceKey string
	Response   graphql.Response `json:"response"`
}

type federationResponse struct {
	ServiceKey string             `json:"serviceKey"`
	Responses  []providerResponse `json:"responses"`
}

// GetProviderResponse Returns the specific provider response by service key
func (f *federationResponse) GetProviderResponse(providerKey string) *providerResponse {
	for _, resp := range f.Responses {
		if resp.ServiceKey == providerKey {
			return &resp
		}
	}
	return nil
}

// Initialize sets up the Federator with providers and an HTTP client.
func Initialize() *Federator {
	var options *federator.Options
	if configs.AppConfig != nil {
		options = configs.AppConfig.Options
	}

	federator := &Federator{}
	federator.Providers = make(map[string]*provider.Provider)
	// Initialize with options if provided

	if options != nil {
		for _, p := range options.Providers {
			// print service url
			logger.Log.Info("Adding Provider from the Config File", "Provider Key", p.ServiceKey, "Provider Url", p.ServiceUrl)
			federator.Providers[p.ServiceKey] = p
		}
	} else {
		logger.Log.Info("No Providers found in the Config File")
	}

	// Initialize HTTP client with timeout and connection pooling
	federator.Client = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
		},
	}

	return federator
}

// FederateQuery takes a raw GraphQL query, splits it into sub-queries for each service,
// sends them to the respective providers, and merges the responses.
func (f *Federator) FederateQuery(request graphql.Request) graphql.Response {

	// Convert the query string into its ast
	src := source.NewSource(&source.Source{
		Body: []byte(request.Query),
		Name: "Query",
	})

	doc, err := parser.Parse(parser.ParseParams{Source: src})

	if err != nil {
		logger.Log.Error("Failed to parse query", "Error", err)
	}

	splitRequests, err := QueryBuilder(doc)

	if err != nil {
		logger.Log.Error("Failed to build queries", "Error", err)
		return graphql.Response{
			Data: nil,
			Errors: []interface{}{
				err.(*graphql.JSONError),
			},
		}
	}

	if len(splitRequests) == 0 {
		logger.Log.Info("No valid service queries found in the request")
		return graphql.Response{
			Data: nil,
			Errors: []interface{}{
				map[string]interface{}{
					"message": "No valid service queries found in the request",
				},
			},
		}
	}

	federationRequest := &federationRequest{
		FederationServiceRequest: splitRequests,
	}
	responses := f.performFederation(federationRequest)

	// Transform the federated responses back to the original query structure

	var response = AccumulateResponse(doc, responses)

	return response

	//return f.mergeResponses(responses.Responses)
}

func (f *Federator) performFederation(r *federationRequest) *federationResponse {
	federationResponse := &federationResponse{
		Responses: make([]providerResponse, 0, len(r.FederationServiceRequest)),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex // to safely append to federationResponse.Responses

	for _, request := range r.FederationServiceRequest {
		p, exists := f.Providers[request.ServiceKey]
		if !exists {
			logger.Log.Info("Provider not found", "Provider Key", request.ServiceKey)
			continue
		}

		wg.Add(1)
		go func(req *federationServiceRequest, prov *provider.Provider) {
			defer wg.Done()

			reqBody, err := json.Marshal(req.GraphQLRequest)
			if err != nil {
				logger.Log.Info("Failed to marshal request", "Provider Key", req.ServiceKey, "Error", err)
				return
			}

			response, err := f.Client.Post(prov.ServiceUrl, "application/json", bytes.NewBuffer(reqBody))
			if err != nil {
				logger.Log.Info("Request failed to the Provider", "Provider Key", req.ServiceKey, "Error", err)
				return
			}
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				logger.Log.Error("Failed to read response body", "Provider Key", req.ServiceKey, "Error", err)
				return
			}

			var bodyJson graphql.Response
			err = json.Unmarshal(body, &bodyJson)
			if err != nil {
				logger.Log.Error("Failed to unmarshal response", "Provider Key", req.ServiceKey, "Error", err)
				return
			}

			// Thread-safe append
			mu.Lock()
			federationResponse.Responses = append(federationResponse.Responses, providerResponse{
				ServiceKey: req.ServiceKey,
				Response:   bodyJson,
			})
			mu.Unlock()
		}(request, p)
	}

	wg.Wait()
	return federationResponse
}

func (f *Federator) mergeResponses(responses []providerResponse) graphql.Response {
	merged := graphql.Response{
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
