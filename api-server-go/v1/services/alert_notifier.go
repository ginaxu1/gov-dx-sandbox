package services

import (
	"log/slog"
)

// LoggingAlertNotifier implements AlertNotifier using structured logging
// In production, this could be extended to send to PagerDuty, Slack, etc.
type LoggingAlertNotifier struct{}

// NewLoggingAlertNotifier creates a new logging-based alert notifier
func NewLoggingAlertNotifier() *LoggingAlertNotifier {
	return &LoggingAlertNotifier{}
}

// SendAlert sends a high-priority alert
func (n *LoggingAlertNotifier) SendAlert(severity string, message string, details map[string]interface{}) error {
	// Log at error level for critical alerts
	if severity == "critical" {
		slog.Error("CRITICAL ALERT",
			"message", message,
			"severity", severity,
			"details", details)
	} else {
		slog.Warn("ALERT",
			"message", message,
			"severity", severity,
			"details", details)
	}

	// In production, you would add integrations here:
	// - Send to PagerDuty
	// - Send to Slack channel
	// - Send to monitoring system (Datadog, New Relic, etc.)
	// - Send email to on-call engineer

	return nil
}

// PagerDutyAlertNotifier implements AlertNotifier using PagerDuty (example implementation)
type PagerDutyAlertNotifier struct {
	integrationKey string
	httpClient     interface{} // Would be *http.Client in real implementation
}

// NewPagerDutyAlertNotifier creates a new PagerDuty alert notifier
// This is a placeholder - actual implementation would require PagerDuty API client
func NewPagerDutyAlertNotifier(integrationKey string) *PagerDutyAlertNotifier {
	return &PagerDutyAlertNotifier{
		integrationKey: integrationKey,
	}
}

// SendAlert sends a high-priority alert to PagerDuty
func (n *PagerDutyAlertNotifier) SendAlert(severity string, message string, details map[string]interface{}) error {
	// Placeholder - actual implementation would:
	// 1. Create PagerDuty event
	// 2. Send HTTP POST to PagerDuty Events API
	// 3. Handle response

	slog.Info("PagerDuty alert sent (placeholder)",
		"severity", severity,
		"message", message,
		"details", details)
	panic("PagerDutyAlertNotifier.SendAlert called: PagerDuty integration not implemented. This is a placeholder and must not be used in production.")
}
