package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseGrantDuration parses a grant duration string and returns the expiration timestamp
// Supported formats: "30d" (days), "1M" (months), "1y" (years), "24h" (hours), "60m" (minutes)
func ParseGrantDuration(duration string) (int64, error) {
	if duration == "" {
		// Default to 30 days if no duration specified
		return time.Now().Add(30 * 24 * time.Hour).Unix(), nil
	}

	duration = strings.ToLower(strings.TrimSpace(duration))

	// Extract number and unit
	var numStr string
	var unit string

	for i, char := range duration {
		if char >= '0' && char <= '9' {
			numStr += string(char)
		} else {
			unit = duration[i:]
			break
		}
	}

	if numStr == "" {
		return 0, fmt.Errorf("invalid duration format: %s", duration)
	}

	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, fmt.Errorf("invalid duration number: %s", numStr)
	}

	now := time.Now()
	var expiration time.Time

	switch unit {
	case "m", "min", "minute", "minutes":
		expiration = now.Add(time.Duration(num) * time.Minute)
	case "h", "hour", "hours":
		expiration = now.Add(time.Duration(num) * time.Hour)
	case "d", "day", "days":
		expiration = now.AddDate(0, 0, num)
	case "M", "month", "months":
		expiration = now.AddDate(0, num, 0)
	case "y", "year", "years":
		expiration = now.AddDate(num, 0, 0)
	default:
		return 0, fmt.Errorf("unsupported duration unit: %s", unit)
	}

	return expiration.Unix(), nil
}

// FormatGrantDuration formats a duration in a human-readable way
func FormatGrantDuration(duration string) string {
	if duration == "" {
		return "30d"
	}
	return duration
}
