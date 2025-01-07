package k8s

import (
	"fmt"
	"time"
)

// formatAge formats a duration since a timestamp in a human-readable format
func formatAge(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "Less than a minute"
	}

	if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", minutes)
	}

	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}

	if duration < 30*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	}

	months := int(duration.Hours() / 24 / 30)
	if months == 1 {
		return "1 month"
	}
	return fmt.Sprintf("%d months", months)
}
