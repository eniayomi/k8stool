package logs

import (
	"context"
)

// LogService defines the interface for retrieving container logs
type LogService interface {
	// GetLogs retrieves logs from a container in a pod
	GetLogs(ctx context.Context, namespace, pod string, opts *LogOptions) (*LogResult, error)

	// StreamLogs streams logs from a container in a pod
	StreamLogs(ctx context.Context, namespace, pod string, opts *LogOptions) (*LogConnection, error)

	// Validate validates the log options
	Validate(opts *LogOptions) error
}
