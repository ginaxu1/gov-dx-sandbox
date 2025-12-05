package federator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/gov-dx-sandbox/exchange/orchestration-engine-go/auth"
	"github.com/gov-dx-sandbox/exchange/orchestration-engine-go/configs"
	"github.com/gov-dx-sandbox/exchange/orchestration-engine-go/consent"
	"github.com/gov-dx-sandbox/exchange/orchestration-engine-go/internals/errors"
	"github.com/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
	"github.com/gov-dx-sandbox/exchange/orchestration-engine-go/middleware"
	auth2 "github.com/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/auth"
	"github.com/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/graphql"
	"github.com/gov-dx-sandbox/exchange/orchestration-engine-go/policy"
	"github.com/gov-dx-sandbox/exchange/orchestration-engine-go/provider"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
	"golang.org/x/oauth2/clientcredentials"
)

// Context key for audit metadata
type contextKey string

const auditMetadataKey contextKey = "auditMetadata"

// AuditMetadata holds metadata needed for audit logging
type AuditMetadata struct {
	ConsumerAppID    string
	ProviderFieldMap *[]ProviderLevelFieldRecord
}

// NewContextWithAuditMetadata creates a new context with audit metadata
func NewContextWithAuditMetadata(ctx context.Context, metadata *AuditMetadata) context.Context {
	return context.WithValue(ctx, auditMetadataKey, metadata)
}

// AuditMetadataFromContext retrieves audit metadata from context
func AuditMetadataFromContext(ctx context.Context) *AuditMetadata {
	metadata, ok := ctx.Value(auditMetadataKey).(*AuditMetadata)
	if !ok {
		return nil
	}
	return metadata
}

// Federator struct that includes all the context needed for federation.
type Federator struct {
	Configs         *configs.Config
	ProviderHandler *provider.Handler
	Client          *http.Client
	Schema          *ast.Document
	SchemaService   interface{} // Will be *services.SchemaService, using interface{} to avoid circular import
}

type FederationServiceAST struct {
	ServiceKey string
	SchemaID   string
	QueryAst   *ast.Document
}

type federationServiceRequest struct {
	ServiceKey     string
	SchemaID       string
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
	ServiceKey string              `json:"ProviderKey"`
	Responses  []*ProviderResponse `json:"responses"`
}

// GetProviderResponse Returns the specific provider response by service key
func (f *FederationResponse) GetProviderResponse(providerKey string) *ProviderResponse {
	for _, resp := range f.Responses {
		if resp.ServiceKey == providerKey {
			return resp
		}
	}
	return nil
}

// Initialize sets up the Federator with providers and an HTTP client.
func Initialize(configs *configs.Config, providerHandler *provider.Handler, schemaService interface{}) *Federator {
	federator := &Federator{
		ProviderHandler: providerHandler,
		SchemaService:   schemaService,
		Configs:         configs,
	}

	// Initialize with providers from config if available
	if configs.Providers != nil {
		for _, p := range configs.Providers {
			// Convert ProviderConfig to Provider
			providerInstance := &provider.Provider{
				ServiceUrl: p.ProviderURL,
				ServiceKey: p.ProviderKey,
				SchemaID:   p.SchemaID,
				Auth:       p.Auth,
			}

			if p.Auth != nil && p.Auth.Type == auth2.AuthTypeOAuth2 {
				providerInstance.OAuth2Config = &clientcredentials.Config{
					ClientID:     p.Auth.ClientID,
					ClientSecret: p.Auth.ClientSecret,
					TokenURL:     p.Auth.TokenURL,
				}
			}

			// print service url
			logger.Log.Info("Adding Provider from the Config File", "Provider Key", p.ProviderKey, "Provider Url", p.ProviderURL)
			providerHandler.AddProvider(providerInstance)
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
func (f *Federator) FederateQuery(ctx context.Context, request graphql.Request, consumerInfo *auth.ConsumerAssertion) graphql.Response {
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
	if schema == nil && f.Configs.Schema != nil {
		schema, err = f.Configs.GetSchemaDocument()
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
	if f.Configs.ArgMapping != nil {
		argMapping = f.Configs.ArgMapping
	}

	requiredArguments := FindRequiredArguments(schemaCollection.ProviderFieldMap, argMapping)

	extractedArgs := ExtractRequiredArguments(requiredArguments, schemaCollection.Arguments)

	// check whether there are variables in the request
	if request.Variables != nil {
		// if there are variables, replace the argument values with the variable values
		PushVariablesFromVariableDefinition(request, extractedArgs, schemaCollection.VariableDefinitions)
	}

	// Safely initialize PDP and CE clients with nil checks
	var pdpClient *policy.PdpClient
	var ceClient *consent.CEClient

	if f.Configs.PdpConfig.ClientURL != "" {
		pdpClient = policy.NewPdpClient(f.Configs.PdpConfig.ClientURL)
	}
	if f.Configs.CeConfig.ClientURL != "" {
		ceClient = consent.NewCEClient(f.Configs.CeConfig.ClientURL)
	}

	// Check if PDP client is available before making request
	var pdpResponse *policy.PdpResponse
	if pdpClient == nil {
		logger.Log.Warn("PDP client not available, skipping policy check")
		// Continue without PDP check - this allows the system to work without PDP
	} else {
		var err error

		pdpRequest := &policy.PdpRequest{
			ConsumerId: consumerInfo.Subscriber,
			AppId:      consumerInfo.ApplicationId,
			RequestId:  "request_123",
		}

		requiredFields := make([]policy.RequiredField, 0)

		for _, field := range *schemaCollection.ProviderFieldMap {
			requiredFields = append(requiredFields, policy.RequiredField{
				ProviderKey: field.ServiceKey,
				SchemaId:    field.SchemaId,
				FieldName:   field.FieldPath,
			})
		}

		pdpRequest.RequiredFields = requiredFields

		pdpResponse, err = pdpClient.MakePdpRequest(pdpRequest)
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

		if !pdpResponse.AppAuthorized {
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

		fields := make([]consent.ConsentField, len(pdpResponse.ConsentRequiredFields))
		for i, f := range pdpResponse.ConsentRequiredFields {
			fields[i].FieldName = f.FieldName
			fields[i].SchemaID = f.SchemaID
		}

		ceRequest := &consent.CERequest{
			AppId:     consumerInfo.ApplicationId,
			Purpose:   "testing",
			SessionId: "session_123",
			ConsentRequirements: []consent.ConsentRequirement{
				{
					Owner:   "citizen",
					OwnerID: extractedArgs[0].Value.GetValue().(string),
					Fields:  fields,
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

	// Inject audit metadata into context
	auditMetadata := &AuditMetadata{
		ConsumerAppID:    consumerInfo.ApplicationId,
		ProviderFieldMap: schemaCollection.ProviderFieldMap,
	}
	ctxWithAudit := NewContextWithAuditMetadata(ctx, auditMetadata)

	responses := f.performFederation(ctxWithAudit, federationRequest)

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
	response := AccumulateResponseWithSchemaInfo(doc, responses, schemaInfoMap)

	return response
}

func (f *Federator) performFederation(ctx context.Context, r *federationRequest) *FederationResponse {
	FederationResponse := &FederationResponse{
		Responses: make([]*ProviderResponse, 0, len(r.FederationServiceRequest)),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex // to safely append to FederationResponse.Responses

	for _, request := range r.FederationServiceRequest {
		p, exists := f.ProviderHandler.GetProvider(request.ServiceKey, request.SchemaID)
		if !exists {
			logger.Log.Info("Provider not found", "Provider Key", request.ServiceKey)
			continue
		}

		wg.Add(1)
		go func(req *federationServiceRequest, prov *provider.Provider) {
			defer wg.Done()

			logAudit := func(status string, err error) {
				f.logAuditEvent(ctx, req.SchemaID, req, status, err)
			}

			reqBody, err := json.Marshal(req.GraphQLRequest)
			if err != nil {
				logger.Log.Info("Failed to marshal request", "Provider Key", req.ServiceKey, "Error", err)
				logAudit("failure", err)
				return
			}

			response, err := prov.PerformRequest(ctx, reqBody)
			if err != nil {
				logger.Log.Info("Request failed to the Provider", "Provider Key", req.ServiceKey, "Error", err)
				logAudit("failure", err)
				return
			}
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				logger.Log.Error("Failed to read response body", "Provider Key", req.ServiceKey, "Error", err)
				logAudit("failure", err)
				return
			}

			var bodyJson graphql.Response
			err = json.Unmarshal(body, &bodyJson)
			if err != nil {
				logger.Log.Error("Failed to unmarshal response", "Provider Key", req.ServiceKey, "Error", err)
				logAudit("failure", err)
				return
			}

			// Determine status based on response
			status := "success"
			if len(bodyJson.Errors) > 0 || response.StatusCode >= 400 {
				status = "failure"
			}

			// Log audit event
			logAudit(status, nil)

			// Thread-safe append
			mu.Lock()
			FederationResponse.Responses = append(FederationResponse.Responses, &ProviderResponse{
				ServiceKey: req.ServiceKey,
				Response:   bodyJson,
			})
			mu.Unlock()
		}(request, p)
	}

	wg.Wait()
	return FederationResponse
}

// logAuditEvent logs a data exchange event to the audit service asynchronously
func (f *Federator) logAuditEvent(ctx context.Context, providerSchemaID string, req *federationServiceRequest, status string, err error) {
	// Retrieve metadata from context
	metadata := AuditMetadataFromContext(ctx)
	if metadata == nil {
		logger.Log.Warn("Audit metadata missing from context, skipping audit log")
		return
	}

	// Extract requested fields for this provider
	requestedFields := make([]string, 0)
	if metadata.ProviderFieldMap != nil {
		for _, field := range *metadata.ProviderFieldMap {
			if field.SchemaId == req.SchemaID && field.ServiceKey == req.ServiceKey {
				requestedFields = append(requestedFields, field.FieldPath)
			}
		}
	}

	// Prepare requested data as JSON
	requestedDataMap := map[string]interface{}{
		"fields": requestedFields,
		"query":  req.GraphQLRequest.Query,
	}
	requestedDataJSON, jsonErr := json.Marshal(requestedDataMap)
	if jsonErr != nil {
		logger.Log.Error("Failed to marshal requested data for audit", "error", jsonErr)
		return
	}

	// Prepare additional info for audit
	additionalInfo := map[string]interface{}{
		"serviceKey": req.ServiceKey,
	}
	if err != nil {
		additionalInfo["error"] = err.Error()
	}
	additionalInfoJSON, jsonErr := json.Marshal(additionalInfo)
	if jsonErr != nil {
		logger.Log.Error("Failed to marshal additional info for audit", "error", jsonErr)
		additionalInfoJSON = []byte("{}")
	}

	// Create audit request for data exchange event
	auditRequest := &middleware.DataExchangeEventAuditRequest{
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Status:        status,
		ApplicationID: metadata.ConsumerAppID,
		SchemaID:      providerSchemaID,
		RequestedData: json.RawMessage(requestedDataJSON),
		// Note: OnBehalfOfOwnerID, ConsumerID, and ProviderID are not populated here
		// to avoid expensive lookup calls. The audit service can handle missing member IDs.
		AdditionalInfo: json.RawMessage(additionalInfoJSON),
	}

	// Log the audit event asynchronously using the global middleware function
	middleware.LogAuditEvent(auditRequest)
}

func (f *Federator) mergeResponses(responses []*ProviderResponse) graphql.Response {
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
