package events

import (
	"context"
)

// EventService defines the interface for managing Kubernetes resources
type EventService interface {
	// List returns a list of events matching the given filter
	List(ctx context.Context, namespace string, filter *EventFilter) (*EventList, error)

	// ListForObject returns events related to a specific resource
	ListForObject(ctx context.Context, namespace, kind, name string) (*EventList, error)

	// Watch watches for events matching the given filter
	Watch(ctx context.Context, namespace string, opts *EventOptions) (<-chan Event, error)

	// Get returns a specific event by name
	Get(ctx context.Context, namespace, name string) (*Event, error)
}
