package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	// Serialize the request
	reqBody, err := json.Marshal(auditReq)
	if err != nil {
		return fmt.Errorf("failed to marshal audit request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", s.auditServiceURL+"/audit/logs", bytes.NewBuffer(reqBody))
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
	var auditResp models.AuditLogResponse
	if err := json.NewDecoder(resp.Body).Decode(&auditResp); err != nil {
		slog.Warn("Failed to parse audit response", "error", err)
		// Don't return error here as the audit was likely successful
	}

	slog.Debug("Audit log sent successfully", "event_id", auditReq.EventID, "status", auditResp.Status)
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

// ExtractConsumerIDFromPath extracts consumer ID from request path
func (s *AuditService) ExtractConsumerIDFromPath(path string) string {
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

// ExtractProviderIDFromPath extracts provider ID from request path
func (s *AuditService) ExtractProviderIDFromPath(path string) string {
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
