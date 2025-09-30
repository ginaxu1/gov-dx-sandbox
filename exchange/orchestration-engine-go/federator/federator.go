package federator

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/auth"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/configs"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/consent"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/internals/errors"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/graphql"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/policy"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/provider"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

// Federator struct that includes all the context needed for federation.
type Federator struct {
	ProviderHandler *provider.Handler
	Client          *http.Client
	Schema          *ast.Document
}

type FederationServiceAST struct {
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

type ProviderResponse struct {
	ServiceKey string
	Response   graphql.Response `json:"response"`
}

type FederationResponse struct {
	ServiceKey string             `json:"serviceKey"`
	Responses  []ProviderResponse `json:"responses"`
}

// GetProviderResponse Returns the specific provider response by service key
func (f *FederationResponse) GetProviderResponse(providerKey string) *ProviderResponse {
	for _, resp := range f.Responses {
		if resp.ServiceKey == providerKey {
			return &resp
		}
	}
	return nil
}

// Initialize sets up the Federator with providers and an HTTP client.
func Initialize(providerHandler *provider.Handler) *Federator {
	options := configs.AppConfig.Options

	federator := &Federator{
		ProviderHandler: providerHandler,
	}

	// Initialize with options if provided

	if options != nil {
		for _, p := range options.Providers {
			// print service url
			logger.Log.Info("Adding Provider from the Config File", "Provider Key", p.ServiceKey, "Provider Url", p.ServiceUrl)
			providerHandler.AddProvider(p.ServiceKey, p)
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
func (f *Federator) FederateQuery(request graphql.Request, consumerInfo *auth.ConsumerAssertion) graphql.Response {
	// Convert the query string into its ast
	src := source.NewSource(&source.Source{
		Body: []byte(request.Query),
		Name: "Query",
	})

	doc, err := parser.Parse(parser.ParseParams{Source: src})
	if err != nil {
		logger.Log.Error("Failed to parse query", "Error", err)
	}

	schema := configs.AppConfig.Schema

	// Collect the directives from the query
	schemaCollection, err := ProviderSchemaCollector(schema, doc)
	if err != nil {
		logger.Log.Error("Failed to collect provider schema", "Error", err)
		return graphql.Response{
			Data: nil,
			Errors: []interface{}{
				err.(*graphql.JSONError),
			},
		}
	}

	requiredArguments := FindRequiredArguments(schemaCollection.ProviderFieldMap, configs.AppConfig.ArgMapping)

	extractedArgs := ExtractRequiredArguments(requiredArguments, schemaCollection.Arguments)

	// check whether there are variables in the request
	if request.Variables != nil {
		// if there are variables, replace the argument values with the variable values
		PushVariablesFromVariableDefinition(request, extractedArgs, schemaCollection.VariableDefinitions)
	}

	pdpClient := policy.NewPdpClient(configs.AppConfig.PdpConfig.ClientUrl)
	ceClient := consent.NewCEClient(configs.AppConfig.CeConfig.ClientUrl)

	pdpResponse, err := pdpClient.MakePdpRequest(&policy.PdpRequest{
		ConsumerId:     consumerInfo.Subscriber,
		AppId:          consumerInfo.ApplicationId,
		RequestId:      "request_123",
		RequiredFields: schemaCollection.ProviderFieldMap,
	})
	if err != nil {
		logger.Log.Info("PDP request failed", "error", err)
		return graphql.Response{
			Data: nil,
			Errors: []interface{}{
				map[string]interface{}{
					"message": "PDP request failed",
					"extensions": map[string]interface{}{
						"code": errors.CodePDPError,
					},
				},
			},
		}
	}

	if pdpResponse == nil {
		logger.Log.Error("Failed to get response from PDP")
		return graphql.Response{
			Data: nil,
			Errors: []interface{}{
				map[string]interface{}{
					"message": "Failed to get response from PDP",
					"extensions": map[string]interface{}{
						"code": errors.CodePDPNoResponse,
					},
				},
			},
		}
	}

	if !pdpResponse.Allowed {
		logger.Log.Info("Request not allowed by PDP")
		return graphql.Response{
			Data: nil,
			Errors: []interface{}{
				map[string]interface{}{
					"message": "Request not allowed by PDP",
					"extensions": map[string]interface{}{
						"code": errors.CodePDPNotAllowed,
					},
				},
			},
		}
	}

	// check whether the arguments contain the citizen id
	if len(extractedArgs) == 0 || extractedArgs[0].Value.GetValue() == nil {
		logger.Log.Info("Citizen ID argument is missing or invalid")
		return graphql.Response{
			Data: nil,
			Errors: []interface{}{
				map[string]interface{}{
					"message": "Citizen ID argument is missing or invalid",
					"extensions": map[string]interface{}{
						"code": errors.CodeMissingEntityIdentifier,
					},
				},
			},
		}
	}

	if pdpResponse.ConsentRequired {
		logger.Log.Info("Consent required for fields", "fields", pdpResponse.ConsentRequiredFields)

		ceRequest := &consent.CERequest{
			AppId:     consumerInfo.ApplicationId,
			Purpose:   "testing",
			SessionId: "session_123",
			DataFields: []consent.DataOwnerRecord{
				{
					OwnerType: "citizen",
					OwnerId:   extractedArgs[0].Value.GetValue().(string),
					Fields:    pdpResponse.ConsentRequiredFields,
				},
			},
		}

		ceResp, err := ceClient.MakeConsentRequest(ceRequest)
		if err != nil {
			logger.Log.Info("CE request failed", "error", err)
			return graphql.Response{
				Data: nil,
				Errors: []interface{}{
					map[string]interface{}{
						"message": "CE request failed",
						"extensions": map[string]interface{}{
							"code": errors.CodeCEError,
						},
					},
				},
			}
		}

		// log the consent response
		logger.Log.Info("Consent Response", "response", ceResp)

		if ceResp.Status != "approved" {
			logger.Log.Info("Consent not approved")
			return graphql.Response{
				Data: nil,
				Errors: []interface{}{
					map[string]interface{}{
						"message": "Consent not approved",
						"extensions": map[string]interface{}{
							"code":             errors.CodeCENotApproved,
							"consentPortalUrl": ceResp.ConsentPortalUrl,
							"consentStatus":    ceResp.Status,
						},
					},
				},
			}
		}
	}

	logger.Log.Info("Consent approved, proceeding with query execution")

	splitRequests, err := QueryBuilder(schemaCollection.ProviderFieldMap, extractedArgs)
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

	// Build schema info map for array-aware processing
	schemaInfoMap, err := BuildSchemaInfoMap(schema, doc)
	if err != nil {
		logger.Log.Error("Failed to build schema info map", "Error", err)
		return graphql.Response{
			Data: nil,
			Errors: []interface{}{
				err.(*graphql.JSONError),
			},
		}
	}

	// Transform the federated responses back to the original query structure using array-aware processing
	response := AccumulateResponseWithSchemaInfo(doc, responses, schemaInfoMap)

	return response
}

func (f *Federator) performFederation(r *federationRequest) *FederationResponse {
	FederationResponse := &FederationResponse{
		Responses: make([]ProviderResponse, 0, len(r.FederationServiceRequest)),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex // to safely append to FederationResponse.Responses

	for _, request := range r.FederationServiceRequest {
		p, exists := f.ProviderHandler.GetProvider(request.ServiceKey)
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

			response, err := prov.PerformRequest(reqBody)
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
			FederationResponse.Responses = append(FederationResponse.Responses, ProviderResponse{
				ServiceKey: req.ServiceKey,
				Response:   bodyJson,
			})
			mu.Unlock()
		}(request, p)
	}

	wg.Wait()
	return FederationResponse
}

func (f *Federator) mergeResponses(responses []ProviderResponse) graphql.Response {
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
