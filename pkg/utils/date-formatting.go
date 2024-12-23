package utils

import (
	"fmt"
	"time"
)

func FormatDuration(d time.Duration) string {
	if d.Hours() > 24*365 {
		years := int(d.Hours() / (24 * 365))
		days := int((d.Hours() - float64(years)*24*365) / 24)
		if days > 0 {
			return fmt.Sprintf("%dy%dd", years, days)
		}
		return fmt.Sprintf("%dy", years)
	}
	if d.Hours() > 24*30 {
		months := int(d.Hours() / (24 * 30))
		days := int((d.Hours() - float64(months)*24*30) / 24)
		if days > 0 {
			return fmt.Sprintf("%dM%dd", months, days)
		}
		return fmt.Sprintf("%dM", months)
	}
	if d.Hours() > 24 {
		days := int(d.Hours() / 24)
		hours := int(d.Hours()) % 24
		if hours > 0 {
			return fmt.Sprintf("%dd%dh", days, hours)
		}
		return fmt.Sprintf("%dd", days)
	}
	if d.Hours() >= 1 {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		if minutes > 0 {
			return fmt.Sprintf("%dh%dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}
	if d.Minutes() >= 1 {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
}
