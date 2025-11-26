package service

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/exchange/consent-engine/v1/models"
	"github.com/gov-dx-sandbox/exchange/shared/utils"
)

// Utility functions for consent management

// generateConsentID generates a unique consent ID
func generateConsentID() uuid.UUID {
	return uuid.New()
}

// getDefaultGrantDuration returns the default grant duration
func getDefaultGrantDuration(duration string) string {
	if duration == "" {
		return "1h" // Default to 1 hour
	}
	return duration
}

// calculateExpiresAt calculates the expiry time based on grant duration
func calculateExpiresAt(grantDuration string, createdAt time.Time) (time.Time, error) {
	duration, err := parseISODuration(grantDuration)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid grant duration format: %w", err)
	}
	return createdAt.Add(duration), nil
}

// parseISODuration parses an ISO 8601 duration string and returns the duration
// Supports formats like: P30D, P1M, P1Y, PT1H, PT30M, P1Y2M3DT4H5M6S
func parseISODuration(duration string) (time.Duration, error) {
	if duration == "" {
		// Default to 1 hour if no duration specified
		return time.Hour, nil
	}

	// Check if it's ISO 8601 format (starts with 'P')
	if len(duration) > 0 && duration[0] == 'P' {
		return parseISO8601Duration(duration)
	}

	// Fallback to legacy format parsing
	return utils.ParseExpiryTime(duration)
}

// parseISO8601Duration parses an ISO 8601 duration string into a time.Duration
func parseISO8601Duration(duration string) (time.Duration, error) {
	// Validate ISO 8601 duration format
	if !isValidISODuration(duration) {
		return 0, fmt.Errorf("invalid ISO 8601 duration format: %s", duration)
	}

	// Remove the 'P' prefix
	if len(duration) == 0 || duration[0] != 'P' {
		return 0, fmt.Errorf("duration must start with 'P'")
	}
	duration = duration[1:]

	var total time.Duration
	var err error

	// Check if there's a time component (starts with 'T')
	timeIndex := -1
	for i, char := range duration {
		if char == 'T' {
			timeIndex = i
			break
		}
	}

	// Parse date components (before 'T' or entire string if no 'T')
	datePart := duration
	if timeIndex != -1 {
		datePart = duration[:timeIndex]
	}

	// Parse years
	years, datePart, err := parseComponent(datePart, "Y")
	if err != nil {
		return 0, err
	}
	total += time.Duration(years) * 365 * 24 * time.Hour

	// Parse months
	months, datePart, err := parseComponent(datePart, "M")
	if err != nil {
		return 0, err
	}
	total += time.Duration(months) * 30 * 24 * time.Hour // Approximate month as 30 days

	// Parse days
	days, _, err := parseComponent(datePart, "D")
	if err != nil {
		return 0, err
	}
	total += time.Duration(days) * 24 * time.Hour

	// Parse time components (after 'T')
	if timeIndex != -1 {
		timePart := duration[timeIndex+1:]

		// Parse hours
		hours, timePart, err := parseComponent(timePart, "H")
		if err != nil {
			return 0, err
		}
		total += time.Duration(hours) * time.Hour

		// Parse minutes
		minutes, timePart, err := parseComponent(timePart, "M")
		if err != nil {
			return 0, err
		}
		total += time.Duration(minutes) * time.Minute

		// Parse seconds
		seconds, _, err := parseComponent(timePart, "S")
		if err != nil {
			return 0, err
		}
		total += time.Duration(seconds) * time.Second
	}

	return total, nil
}

// isValidISODuration validates if a string is a valid ISO 8601 duration
func isValidISODuration(duration string) bool {
	// ISO 8601 duration pattern: P(\d+Y)?(\d+M)?(\d+D)?(T(\d+H)?(\d+M)?(\d+(\.\d+)?S)?)?
	pattern := `^P(?:\d+Y)?(?:\d+M)?(?:\d+D)?(?:T(?:\d+H)?(?:\d+M)?(?:\d+(?:\.\d+)?S)?)?$`
	matched, _ := regexp.MatchString(pattern, duration)
	return matched
}

// parseComponent extracts a numeric component from a duration string
func parseComponent(duration, suffix string) (int, string, error) {
	pattern := fmt.Sprintf(`(\d+)%s`, suffix)
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(duration)

	if len(matches) == 0 {
		return 0, duration, nil
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, duration, err
	}

	// Remove the matched part from the duration string
	remaining := re.ReplaceAllString(duration, "")
	return value, remaining, nil
}

// getAllFields extracts all fields from data fields
func getAllFields(dataFields []models.DataField) []string {
	var allFields []string
	for _, field := range dataFields {
		allFields = append(allFields, field.Fields...)
	}
	return allFields
}

// isValidStatusTransition checks if a status transition is valid
func isValidStatusTransition(current, new models.ConsentStatus) bool {
	validTransitions := map[models.ConsentStatus][]models.ConsentStatus{
		models.StatusPending:  {models.StatusApproved, models.StatusRejected, models.StatusExpired},                       // Initial decision
		models.StatusApproved: {models.StatusApproved, models.StatusRejected, models.StatusRevoked, models.StatusExpired}, // Direct approval flow: approved->approved (success), approved->rejected (direct rejection), approved->revoked (user revocation), approved->expired (expiry)
		models.StatusRejected: {models.StatusExpired},                                                                     // Terminal state - can only transition to expired
		models.StatusExpired:  {models.StatusExpired},                                                                     // Terminal state - can only stay expired
		models.StatusRevoked:  {models.StatusExpired},                                                                     // Terminal state - can only transition to expired
	}

	allowed, exists := validTransitions[current]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == new {
			return true
		}
	}
	return false
}
