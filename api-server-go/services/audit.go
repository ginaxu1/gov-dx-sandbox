package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/models"
)

// AuditService handles communication with the audit-service
type AuditService struct {
	auditServiceURL string
	httpClient      *http.Client
}

// NewAuditService creates a new audit service
func NewAuditService(auditServiceURL string) *AuditService {
	return &AuditService{
		auditServiceURL: auditServiceURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second, // Short timeout to avoid blocking main requests
		},
	}
}

// SendAuditLog sends an audit log to the audit-service
func (s *AuditService) SendAuditLog(ctx context.Context, auditReq *models.AuditLogRequest) error {
	// Convert to new simplified log structure
	logReq := models.LogRequest{
		Status:        s.mapTransactionStatus(auditReq.TransactionStatus),
		RequestedData: s.extractGraphQLQuery(auditReq.RequestedData),
		ConsumerID:    auditReq.ConsumerID,
		ProviderID:    auditReq.ProviderID,
	}

	// Serialize the request
	reqBody, err := json.Marshal(logReq)
	if err != nil {
		return fmt.Errorf("failed to marshal audit request: %w", err)
	}

	// Create HTTP request to new /api/logs endpoint
	req, err := http.NewRequestWithContext(ctx, "POST", s.auditServiceURL+"/api/logs", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create audit request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send audit request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("audit service returned status %d", resp.StatusCode)
	}

	// Parse response
	var logResp models.Log
	if err := json.NewDecoder(resp.Body).Decode(&logResp); err != nil {
		slog.Warn("Failed to parse audit response", "error", err)
		// Don't return error here as the audit was likely successful
	}

	slog.Debug("Audit log sent successfully", "log_id", logResp.ID, "status", logResp.Status)
	return nil
}

// SendAuditLogAsync sends an audit log asynchronously to avoid blocking the main request
func (s *AuditService) SendAuditLogAsync(auditReq *models.AuditLogRequest) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.SendAuditLog(ctx, auditReq); err != nil {
			slog.Error("Failed to send audit log asynchronously",
				"error", err,
				"event_id", auditReq.EventID,
				"consumer_id", auditReq.ConsumerID,
				"provider_id", auditReq.ProviderID)
		}
	}()
}

// ExtractConsumerIDFromRequest extracts consumer ID from request path, headers, query params, or body
func (s *AuditService) ExtractConsumerIDFromRequest(r *http.Request) string {
	// 1. Try to extract from path first
	if consumerID := s.extractConsumerIDFromPath(r.URL.Path); consumerID != "" {
		return consumerID
	}

	// 2. Try to extract from headers
	if consumerID := r.Header.Get("X-Consumer-ID"); consumerID != "" {
		return consumerID
	}
	if consumerID := r.Header.Get("X-User-ID"); consumerID != "" {
		return consumerID
	}

	// 3. Try to extract from query parameters
	if consumerID := r.URL.Query().Get("consumerId"); consumerID != "" {
		return consumerID
	}
	if consumerID := r.URL.Query().Get("consumer_id"); consumerID != "" {
		return consumerID
	}

	// 4. Try to extract from request body
	if consumerID := s.extractConsumerIDFromBody(r); consumerID != "" {
		return consumerID
	}

	return ""
}

// extractConsumerIDFromPath extracts consumer ID from request path
func (s *AuditService) extractConsumerIDFromPath(path string) string {
	// Extract consumer ID from paths like /consumers/{consumerId} or /consumer-applications/{consumerId}
	if strings.HasPrefix(path, "/consumers/") {
		// Extract ID after /consumers/
		parts := strings.Split(path, "/")
		if len(parts) >= 3 {
			return parts[2] // /consumers/{id}
		}
	}
	if strings.HasPrefix(path, "/consumer-applications/") {
		// Extract ID after /consumer-applications/
		parts := strings.Split(path, "/")
		if len(parts) >= 3 {
			return parts[2] // /consumer-applications/{id}
		}
	}
	return ""
}

// ExtractProviderIDFromRequest extracts provider ID from request path, headers, query params, or body
func (s *AuditService) ExtractProviderIDFromRequest(r *http.Request) string {
	// 1. Try to extract from path first
	if providerID := s.extractProviderIDFromPath(r.URL.Path); providerID != "" {
		return providerID
	}

	// 2. Try to extract from headers
	if providerID := r.Header.Get("X-Provider-ID"); providerID != "" {
		return providerID
	}

	// 3. Try to extract from query parameters
	if providerID := r.URL.Query().Get("providerId"); providerID != "" {
		return providerID
	}
	if providerID := r.URL.Query().Get("provider_id"); providerID != "" {
		return providerID
	}

	// 4. Try to extract from request body
	if providerID := s.extractProviderIDFromBody(r); providerID != "" {
		return providerID
	}

	return ""
}

// extractProviderIDFromPath extracts provider ID from request path
func (s *AuditService) extractProviderIDFromPath(path string) string {
	// Extract provider ID from paths like /providers/{providerId}
	if strings.HasPrefix(path, "/providers/") {
		// Extract ID after /providers/
		parts := strings.Split(path, "/")
		if len(parts) >= 3 {
			return parts[2] // /providers/{id}
		}
	}
	return ""
}

// DetermineTransactionStatus determines if the transaction was successful based on status code
func (s *AuditService) DetermineTransactionStatus(statusCode int) string {
	if statusCode >= 200 && statusCode < 300 {
		return "SUCCESS"
	}
	return "FAILURE"
}

// mapTransactionStatus maps the old transaction status to new simplified status
func (s *AuditService) mapTransactionStatus(transactionStatus string) string {
	if transactionStatus == "SUCCESS" {
		return "success"
	}
	return "failure"
}

// extractConsumerIDFromBody extracts consumer ID from request body
func (s *AuditService) extractConsumerIDFromBody(r *http.Request) string {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		return ""
	}

	// Restore the body for the next handler
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	// Try to parse as JSON
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return ""
	}

	// Look for common consumer ID field names
	consumerIDFields := []string{
		"consumerId", "consumer_id", "userId", "user_id",
		"clientId", "client_id", "appId", "app_id",
	}

	for _, field := range consumerIDFields {
		if value, exists := data[field]; exists {
			if str, ok := value.(string); ok && str != "" {
				return str
			}
		}
	}

	return ""
}

// extractProviderIDFromBody extracts provider ID from request body
func (s *AuditService) extractProviderIDFromBody(r *http.Request) string {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		return ""
	}

	// Restore the body for the next handler
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	// Try to parse as JSON
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return ""
	}

	// Look for common provider ID field names
	providerIDFields := []string{
		"providerId", "provider_id", "serviceId", "service_id",
		"sourceId", "source_id", "targetId", "target_id",
	}

	for _, field := range providerIDFields {
		if value, exists := data[field]; exists {
			if str, ok := value.(string); ok && str != "" {
				return str
			}
		}
	}

	return ""
}

// ExtractGraphQLQueryFromRequest extracts GraphQL query from request headers, body, or query params
func (s *AuditService) ExtractGraphQLQueryFromRequest(r *http.Request) string {
	// 1. Try to extract from headers
	if query := r.Header.Get("X-GraphQL-Query"); query != "" {
		return query
	}

	// 2. Try to extract from query parameters
	if query := r.URL.Query().Get("query"); query != "" {
		return query
	}

	// 3. Try to extract from request body
	if query := s.extractGraphQLQueryFromBody(r); query != "" {
		return query
	}

	return ""
}

// extractGraphQLQueryFromBody extracts GraphQL query from request body
func (s *AuditService) extractGraphQLQueryFromBody(r *http.Request) string {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		return ""
	}

	// Restore the body for the next handler
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	// Try to parse as JSON
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		// If not JSON, check if it's a raw GraphQL query
		bodyStr := string(body)
		if strings.Contains(strings.ToLower(bodyStr), "query") ||
			strings.Contains(strings.ToLower(bodyStr), "mutation") ||
			strings.Contains(strings.ToLower(bodyStr), "subscription") {
			return bodyStr
		}
		return ""
	}

	// Look for common GraphQL query fields
	queryFields := []string{"query", "Query", "operationName", "operation"}
	for _, field := range queryFields {
		if query, ok := data[field].(string); ok && query != "" {
			return query
		}
	}

	// If no query field found, return the raw data as string
	return string(body)
}

// extractGraphQLQuery extracts GraphQL query from request data (legacy method for backward compatibility)
func (s *AuditService) extractGraphQLQuery(requestData json.RawMessage) string {
	if len(requestData) == 0 {
		return "No query data"
	}

	// Try to parse as JSON and extract query field
	var data map[string]interface{}
	if err := json.Unmarshal(requestData, &data); err != nil {
		// If not JSON, return as string
		return string(requestData)
	}

	// Look for common GraphQL query fields
	if query, ok := data["query"].(string); ok {
		return query
	}
	if query, ok := data["Query"].(string); ok {
		return query
	}
	if query, ok := data["operationName"].(string); ok {
		return query
	}

	// If no query field found, return the raw data as string
	return string(requestData)
}
