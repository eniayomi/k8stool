package utils

import (
	"github.com/fatih/color"
)

var (
	Green    = color.New(color.FgGreen).SprintFunc()
	Yellow   = color.New(color.FgYellow).SprintFunc()
	Red      = color.New(color.FgRed).SprintFunc()
	Blue     = color.New(color.FgBlue).SprintFunc()
	Bold     = color.New(color.Bold).SprintFunc()
	HiGreen  = color.New(color.FgHiGreen).SprintFunc()
	HiYellow = color.New(color.FgHiYellow).SprintFunc()
	HiRed    = color.New(color.FgHiRed).SprintFunc()
)

// ColorizeStatus returns a colored string based on the status
func ColorizeStatus(status string) string {
	switch status {
	case "Running":
		return Green(status)
	case "Pending":
		return Yellow(status)
	case "Succeeded":
		return HiGreen(status)
	case "Failed", "Evicted":
		return Red(status)
	case "CrashLoopBackOff":
		return HiRed(status)
	case "Completed":
		return HiGreen(status)
	case "Terminating":
		return HiYellow(status)
	default:
		return status
	}
}
