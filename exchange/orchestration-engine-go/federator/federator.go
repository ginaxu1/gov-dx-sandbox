package federator

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
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
	SchemaService   interface{} // Will be *services.SchemaService, using interface{} to avoid circular import
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
func Initialize(providerHandler *provider.Handler, schemaService interface{}) *Federator {
	federator := &Federator{
		ProviderHandler: providerHandler,
		SchemaService:   schemaService,
	}

	// Initialize with providers from config if available
	if configs.AppConfig != nil && configs.AppConfig.Providers != nil {
		for _, p := range configs.AppConfig.Providers {
			// Convert ProviderConfig to Provider
			providerInstance := &provider.Provider{
				ServiceUrl: p.ProviderURL,
				ServiceKey: p.ProviderKey,
				Auth:       p.Auth,
			}
			// print service url
			logger.Log.Info("Adding Provider from the Config File", "Provider Key", p.ProviderKey, "Provider Url", p.ProviderURL)
			providerHandler.AddProvider(p.ProviderKey, providerInstance)
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

	// Get schema document from database or config
	var schema *ast.Document

	// First try to get from database if schema service is available
	if f.SchemaService != nil {
		// Use reflection to call GetActiveSchema method
		schemaServiceValue := reflect.ValueOf(f.SchemaService)
		if schemaServiceValue.IsValid() && !schemaServiceValue.IsNil() {
			getActiveSchemaMethod := schemaServiceValue.MethodByName("GetActiveSchema")
			if getActiveSchemaMethod.IsValid() {
				results := getActiveSchemaMethod.Call([]reflect.Value{})
				if len(results) >= 2 && !results[1].IsNil() {
					// Error occurred
					logger.Log.Warn("Failed to get active schema from database", "Error", results[1].Interface())
				} else if len(results) >= 1 && !results[0].IsNil() {
					// Got schema from database
					schemaRecord := results[0].Interface()
					// Extract SDL from schema record using reflection
					schemaRecordValue := reflect.ValueOf(schemaRecord)
					// If it's a pointer, dereference it
					if schemaRecordValue.Kind() == reflect.Ptr {
						schemaRecordValue = schemaRecordValue.Elem()
					}
					sdlField := schemaRecordValue.FieldByName("SDL")
					if sdlField.IsValid() && sdlField.Kind() == reflect.String {
						sdlString := sdlField.String()
						src := source.NewSource(&source.Source{
							Body: []byte(sdlString),
							Name: "ActiveSchema",
						})
						schema, err = parser.Parse(parser.ParseParams{Source: src})
						if err != nil {
							logger.Log.Error("Failed to parse active schema from database", "Error", err)
							schema = nil
						}
					}
				}
			}
		}
	} else {
		logger.Log.Info("SchemaService is nil, skipping database schema lookup")
	}

	// Fallback to config if no schema from database
	if schema == nil && configs.AppConfig != nil && configs.AppConfig.Schema != nil {
		schema, err = configs.AppConfig.GetSchemaDocument()
		if err != nil {
			logger.Log.Warn("Failed to get schema from config", "Error", err)
			schema = nil
		}
	}

	// Final fallback to schema.graphql file if no schema from database or config
	if schema == nil {
		logger.Log.Info("No schema found in database or config, attempting to load schema.graphql file")
		schema, err = f.loadSchemaFromFile()
		if err != nil {
			logger.Log.Error("Failed to load schema from file", "Error", err)
			return graphql.Response{
				Data: nil,
				Errors: []interface{}{
					&graphql.JSONError{
						Message: "No active schema found. Please create and activate a schema using the schema management API first, or ensure schema.graphql file exists.",
					},
				},
			}
		}
	}

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

	// Safely get argument mapping with nil check
	var argMapping []*graphql.ArgMapping
	if configs.AppConfig != nil && configs.AppConfig.ArgMapping != nil {
		argMapping = configs.AppConfig.ArgMapping
	}

	var requiredArguments = FindRequiredArguments(schemaCollection.ProviderFieldMap, argMapping)

	var extractedArgs = ExtractRequiredArguments(requiredArguments, schemaCollection.Arguments)

	// check whether there are variables in the request
	if request.Variables != nil {
		// if there are variables, replace the argument values with the variable values
		PushVariablesFromVariableDefinition(request, extractedArgs, schemaCollection.VariableDefinitions)
	}

	// Safely initialize PDP and CE clients with nil checks
	var pdpClient *policy.PdpClient
	var ceClient *consent.CEClient

	if configs.AppConfig != nil {
		if configs.AppConfig.PdpConfig.ClientURL != "" {
			pdpClient = policy.NewPdpClient(configs.AppConfig.PdpConfig.ClientURL)
		}
		if configs.AppConfig.CeConfig.ClientURL != "" {
			ceClient = consent.NewCEClient(configs.AppConfig.CeConfig.ClientURL)
		}
	}

	// Check if PDP client is available before making request
	var pdpResponse *policy.PdpResponse
	if pdpClient == nil {
		logger.Log.Warn("PDP client not available, skipping policy check")
		// Continue without PDP check - this allows the system to work without PDP
	} else {
		var err error
		pdpResponse, err = pdpClient.MakePdpRequest(&policy.PdpRequest{
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

	// Handle consent check only if PDP client was available and consent is required
	if pdpClient != nil && pdpResponse != nil && pdpResponse.ConsentRequired {
		logger.Log.Info("Consent required for fields", "fields", pdpResponse.ConsentRequiredFields)

		// Check if CE client is available
		if ceClient == nil {
			logger.Log.Warn("CE client not available, skipping consent check")
			return graphql.Response{
				Data: nil,
				Errors: []interface{}{
					map[string]interface{}{
						"message": "Consent required but consent engine not available",
						"extensions": map[string]interface{}{
							"code": errors.CodeCEError,
						},
					},
				},
			}
		}

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
	var schemaInfoMap map[string]*SourceSchemaInfo
	if schema != nil {
		schemaInfoMap, err = BuildSchemaInfoMap(schema, doc)
		if err != nil {
			logger.Log.Error("Failed to build schema info map", "Error", err)
		}
	}
	// Error handling is done above in the if block

	// Transform the federated responses back to the original query structure using array-aware processing
	var response = AccumulateResponseWithSchemaInfo(doc, responses, schemaInfoMap)

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

// loadSchemaFromFile loads the schema from schema.graphql file as a fallback
func (f *Federator) loadSchemaFromFile() (*ast.Document, error) {
	// Try to read schema.graphql file from current directory
	schemaData, err := os.ReadFile("schema.graphql")
	if err != nil {
		// Try alternative paths
		alternativePaths := []string{
			"./schema.graphql",
			"../schema.graphql",
			"../../schema.graphql",
		}

		for _, path := range alternativePaths {
			schemaData, err = os.ReadFile(path)
			if err == nil {
				logger.Log.Info("Successfully found schema.graphql at", "path", path)
				break
			}
		}

		if err != nil {
			return nil, fmt.Errorf("could not find schema.graphql file in any expected location: %w", err)
		}
	}

	// Parse the schema file
	src := source.NewSource(&source.Source{
		Body: schemaData,
		Name: "SchemaFile",
	})

	schema, err := parser.Parse(parser.ParseParams{Source: src})
	if err != nil {
		return nil, err
	}

	logger.Log.Info("Successfully loaded schema from schema.graphql file")
	return schema, nil
}
