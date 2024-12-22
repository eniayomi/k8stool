package utils

import (
	"fmt"
	"time"
)

// FormatDuration formats a duration into a human-readable string (e.g., "2d", "5h", "3m")
func FormatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24

	if days >= 365 {
		years := days / 365
		remainingDays := days % 365
		if remainingDays > 0 {
			return fmt.Sprintf("%dy%dd", years, remainingDays)
		}
		return fmt.Sprintf("%dy", years)
	}
	if days > 0 {
		if hours > 0 {
			return fmt.Sprintf("%dd%dh", days, hours)
		}
		return fmt.Sprintf("%dd", days)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dm", minutes)
}

// FormatResourceValue formats resource values (CPU/Memory) with appropriate units
func FormatResourceValue(value string) string {
	if value == "" {
		return "<none>"
	}
	return value
}

// TruncateString truncates a string if it's longer than maxLen
func TruncateString(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen-3] + "..."
}
