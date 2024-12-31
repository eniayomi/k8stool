package exec

import (
	"context"
)

// ExecService defines the interface for executing commands in containers
type ExecService interface {
	// Exec executes a command in a container and returns the result
	Exec(ctx context.Context, namespace, pod string, opts *ExecOptions) (*ExecResult, error)

	// Stream executes a command in a container and streams the input/output
	Stream(ctx context.Context, namespace, pod string, opts *ExecOptions) (*ExecConnection, error)

	// Validate validates the exec options
	Validate(opts *ExecOptions) error
}
