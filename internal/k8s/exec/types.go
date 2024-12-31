package exec

import (
	"io"
)

// ExecOptions represents options for executing a command in a container
type ExecOptions struct {
	// Command is the command and arguments to execute
	Command []string `json:"command"`

	// Container is the name of the container to execute in
	// If not specified, the first container in the pod will be used
	Container string `json:"container,omitempty"`

	// Stdin enables stdin for the exec session
	Stdin bool `json:"stdin,omitempty"`

	// TTY enables TTY for the exec session
	TTY bool `json:"tty,omitempty"`

	// Streams configures the input/output streams for the exec session
	Streams *IOStreams `json:"-"`
}

// IOStreams holds the input/output streams for the exec session
type IOStreams struct {
	// In holds the input stream (stdin)
	In io.Reader

	// Out holds the output stream (stdout)
	Out io.Writer

	// ErrOut holds the error output stream (stderr)
	ErrOut io.Writer
}

// ExecResult represents the result of an exec operation
type ExecResult struct {
	// ExitCode is the exit code of the executed command
	ExitCode int `json:"exitCode"`

	// Error is any error that occurred during execution
	Error string `json:"error,omitempty"`
}

// ExecConnection represents an active exec connection
type ExecConnection struct {
	// Stdin is the stdin stream
	Stdin io.WriteCloser

	// Stdout is the stdout stream
	Stdout io.Reader

	// Stderr is the stderr stream
	Stderr io.Reader

	// TTY indicates if this is a TTY session
	TTY bool

	// TerminalSizeQueue handles terminal resize events
	TerminalSizeQueue TerminalSizeQueue
}

// TerminalSize represents the size of a terminal
type TerminalSize struct {
	// Width is the width of the terminal in characters
	Width uint16

	// Height is the height of the terminal in characters
	Height uint16
}

// TerminalSizeQueue is an interface for handling terminal resize events
type TerminalSizeQueue interface {
	// Next returns the new terminal size after the terminal has been resized
	Next() *TerminalSize
}
