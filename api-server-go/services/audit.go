package services

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/models"
	"github.com/gov-dx-sandbox/shared/redis"
)

// AuditService handles communication with the audit-service via Redis
type AuditService struct {
	redisClient *redis.RedisClient
}

// NewAuditService creates a new audit service with Redis support
func NewAuditService() *AuditService {
	// Try to connect to Redis
	redisClient, err := redis.NewClient(&redis.Config{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	if err != nil {
		slog.Warn("Failed to connect to Redis, audit service will be in degraded state", "error", err)
		redisClient = nil
	}

	return &AuditService{
		redisClient: redisClient,
	}
}

// SendAuditLogAsync sends an audit log asynchronously to Redis stream
func (s *AuditService) SendAuditLogAsync(auditReq *models.AuditLogRequest) {
	if s.redisClient == nil {
		slog.Warn("Redis client not available, skipping audit log",
			"event_id", auditReq.EventID,
			"consumer_id", auditReq.ConsumerID,
			"provider_id", auditReq.ProviderID)
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := s.sendAuditLogToRedis(ctx, auditReq); err != nil {
			slog.Error("Failed to send audit log to Redis stream",
				"error", err,
				"event_id", auditReq.EventID,
				"consumer_id", auditReq.ConsumerID,
				"provider_id", auditReq.ProviderID)
		} else {
			slog.Debug("Audit log sent to Redis stream successfully",
				"event_id", auditReq.EventID)
		}
	}()
}

// sendAuditLogToRedis sends an audit log to Redis stream
func (s *AuditService) sendAuditLogToRedis(ctx context.Context, auditReq *models.AuditLogRequest) error {
	// Convert to simple map for Redis stream
	event := map[string]interface{}{
		"event_id":           auditReq.EventID.String(),
		"consumer_id":        auditReq.ConsumerID,
		"provider_id":        auditReq.ProviderID,
		"requested_data":     string(auditReq.RequestedData),
		"response_data":      string(auditReq.ResponseData),
		"transaction_status": auditReq.TransactionStatus,
		"user_agent":         auditReq.UserAgent,
		"ip_address":         auditReq.IPAddress,
		"timestamp":          time.Now().Unix(),
	}

	// Use the new PublishAuditEvent method
	_, err := s.redisClient.PublishAuditEvent(ctx, "audit-events", event)
	return err
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

	// 4. Try to extract from request body (using shared body reader)
	body, err := s.readRequestBodyOnce(r)
	if err == nil && len(body) > 0 {
		if consumerID := s.extractConsumerIDFromBodyData(body); consumerID != "" {
			return consumerID
		}
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

// ExtractConsumerIDFromRequestWithBody extracts consumer ID from request using pre-read body data
func (s *AuditService) ExtractConsumerIDFromRequestWithBody(r *http.Request, body []byte) string {
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

	// 4. Try to extract from pre-read body data
	if len(body) > 0 {
		if consumerID := s.extractConsumerIDFromBodyData(body); consumerID != "" {
			return consumerID
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

	// 4. Try to extract from request body (using shared body reader)
	body, err := s.readRequestBodyOnce(r)
	if err == nil && len(body) > 0 {
		if providerID := s.extractProviderIDFromBodyData(body); providerID != "" {
			return providerID
		}
	}

	return ""
}

// ExtractProviderIDFromRequestWithBody extracts provider ID from request using pre-read body data
func (s *AuditService) ExtractProviderIDFromRequestWithBody(r *http.Request, body []byte) string {
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

	// 4. Try to extract from pre-read body data
	if len(body) > 0 {
		if providerID := s.extractProviderIDFromBodyData(body); providerID != "" {
			return providerID
		}
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

// readRequestBodyOnce reads the request body once and caches it for reuse
func (s *AuditService) readRequestBodyOnce(r *http.Request) ([]byte, error) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	// Restore the body for the next handler
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	return body, nil
}

// extractConsumerIDFromBodyData extracts consumer ID from cached body data
func (s *AuditService) extractConsumerIDFromBodyData(body []byte) string {
	if len(body) == 0 {
		return ""
	}

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

// extractProviderIDFromBodyData extracts provider ID from cached body data
func (s *AuditService) extractProviderIDFromBodyData(body []byte) string {
	if len(body) == 0 {
		return ""
	}

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

	// 3. Try to extract from request body (using shared body reader)
	body, err := s.readRequestBodyOnce(r)
	if err == nil && len(body) > 0 {
		if query := s.extractGraphQLQueryFromBodyData(body); query != "" {
			return query
		}
	}

	return ""
}

// ExtractGraphQLQueryFromRequestWithBody extracts GraphQL query from request using pre-read body data
func (s *AuditService) ExtractGraphQLQueryFromRequestWithBody(r *http.Request, body []byte) string {
	// 1. Try to extract from headers
	if query := r.Header.Get("X-GraphQL-Query"); query != "" {
		return query
	}

	// 2. Try to extract from query parameters
	if query := r.URL.Query().Get("query"); query != "" {
		return query
	}

	// 3. Try to extract from pre-read body data
	if len(body) > 0 {
		if query := s.extractGraphQLQueryFromBodyData(body); query != "" {
			return query
		}
	}

	return ""
}

// ExtractAllFromRequestWithBody extracts consumer ID, provider ID, and GraphQL query from request using a single body read
func (s *AuditService) ExtractAllFromRequestWithBody(r *http.Request) (consumerID, providerID, graphqlQuery string, body []byte, err error) {
	// Read the request body once
	body, err = s.readRequestBodyOnce(r)
	if err != nil {
		return "", "", "", nil, err
	}

	// Extract all information using the pre-read body data
	consumerID = s.ExtractConsumerIDFromRequestWithBody(r, body)
	providerID = s.ExtractProviderIDFromRequestWithBody(r, body)
	graphqlQuery = s.ExtractGraphQLQueryFromRequestWithBody(r, body)

	return consumerID, providerID, graphqlQuery, body, nil
}

// extractGraphQLQueryFromBodyData extracts GraphQL query from cached body data
func (s *AuditService) extractGraphQLQueryFromBodyData(body []byte) string {
	if len(body) == 0 {
		return ""
	}

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
