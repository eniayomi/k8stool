package utils

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
