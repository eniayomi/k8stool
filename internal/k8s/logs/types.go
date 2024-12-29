package logs

import (
	"io"
	"time"
)

// LogOptions represents options for retrieving container logs
type LogOptions struct {
	// Container is the name of the container to get logs from
	// If not specified, the first container in the pod will be used
	Container string `json:"container,omitempty"`

	// Follow indicates if the logs should be streamed
	Follow bool `json:"follow,omitempty"`

	// Previous indicates if the previously terminated container logs should be returned
	Previous bool `json:"previous,omitempty"`

	// SinceTime returns logs newer than a specific time
	SinceTime *time.Time `json:"sinceTime,omitempty"`

	// SinceSeconds returns logs newer than a relative duration in seconds
	SinceSeconds *int64 `json:"sinceSeconds,omitempty"`

	// TailLines limits the number of lines to return from the end of the logs
	TailLines *int64 `json:"tailLines,omitempty"`

	// LimitBytes limits the number of bytes to return from the server
	LimitBytes *int64 `json:"limitBytes,omitempty"`

	// Timestamps includes timestamps on each line in the log output
	Timestamps bool `json:"timestamps,omitempty"`

	// Writer specifies where to write the logs
	Writer io.Writer `json:"-"`
}

// LogResult represents the result of a log retrieval operation
type LogResult struct {
	// Error is any error that occurred during log retrieval
	Error string `json:"error,omitempty"`
}

// LogConnection represents an active log streaming connection
type LogConnection struct {
	// Reader is the log stream reader
	Reader io.ReadCloser

	// Done is closed when the log streaming is complete
	Done chan struct{}

	// Error holds any error that occurred during streaming
	Error error
}
