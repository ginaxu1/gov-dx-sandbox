package main

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// ISO 8601 Duration parsing utilities
// Supports formats like: P30D, P1M, P1Y, PT1H, PT30M, P1Y2M3DT4H5M6S

// parseISODuration parses an ISO 8601 duration string and returns the expiration timestamp
func parseISODuration(duration string) (int64, error) {
	if duration == "" {
		// Default to 30 days if no duration specified
		return time.Now().Add(30 * 24 * time.Hour).Unix(), nil
	}

	// Validate ISO 8601 duration format
	if !isValidISODuration(duration) {
		return 0, fmt.Errorf("invalid ISO 8601 duration format: %s", duration)
	}

	// Parse the duration
	parsedDuration, err := parseDurationString(duration)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration %s: %w", duration, err)
	}

	// Calculate expiration time
	expiresAt := time.Now().Add(parsedDuration)
	return expiresAt.Unix(), nil
}

// isValidISODuration validates if a string is a valid ISO 8601 duration
func isValidISODuration(duration string) bool {
	// ISO 8601 duration pattern: P(\\d+Y)?(\\d+M)?(\\d+D)?(T(\\d+H)?(\\d+M)?(\\d+(\\.\\d+)?S)?)?
	pattern := `^P(?:\d+Y)?(?:\d+M)?(?:\d+D)?(?:T(?:\d+H)?(?:\d+M)?(?:\d+(?:\.\d+)?S)?)?$`
	matched, _ := regexp.MatchString(pattern, duration)
	return matched
}

// parseDurationString parses an ISO 8601 duration string into a time.Duration
func parseDurationString(duration string) (time.Duration, error) {
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
