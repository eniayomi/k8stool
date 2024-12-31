package events

import (
	"time"

	corev1 "k8s.io/api/core/v1"
)

// EventType represents the type of event
type EventType string

const (
	// Normal event type
	Normal EventType = "Normal"
	// Warning event type
	Warning EventType = "Warning"
	// Error event type
	Error EventType = "Error"
)

// Event represents a Kubernetes event with additional metadata
type Event struct {
	// Type is the event type (Normal, Warning, Error)
	Type EventType `json:"type"`

	// Name is the event name
	Name string `json:"name"`

	// Namespace is the event namespace
	Namespace string `json:"namespace"`

	// ResourceKind is the kind of resource this event is about
	ResourceKind string `json:"resourceKind"`

	// ResourceName is the name of the resource this event is about
	ResourceName string `json:"resourceName"`

	// Reason is a short, machine understandable string that gives the reason
	// for the transition into the object's current status
	Reason string `json:"reason"`

	// Message is a human-readable description of the status of this operation
	Message string `json:"message"`

	// Component is the component reporting this event
	Component string `json:"component"`

	// Host is the name of the host where this event was generated
	Host string `json:"host"`

	// FirstTimestamp is the time when this event was first observed
	FirstTimestamp time.Time `json:"firstTimestamp"`

	// LastTimestamp is the time when this event was last observed
	LastTimestamp time.Time `json:"lastTimestamp"`

	// Count is the number of times this event has occurred
	Count int32 `json:"count"`

	// IsWarning indicates if this is a warning event
	IsWarning bool `json:"isWarning"`
}

// EventList represents a list of events
type EventList struct {
	// Items is the list of events
	Items []Event `json:"items"`

	// Total is the total number of events
	Total int `json:"total"`
}

// EventFilter represents filters for event queries
type EventFilter struct {
	// Types are the event types to include
	Types []EventType `json:"types,omitempty"`

	// ResourceKinds are the resource kinds to include
	ResourceKinds []string `json:"resourceKinds,omitempty"`

	// ResourceNames are the resource names to include
	ResourceNames []string `json:"resourceNames,omitempty"`

	// Components are the components to include
	Components []string `json:"components,omitempty"`

	// Since is the time since when to include events
	Since *time.Time `json:"since,omitempty"`

	// SortBy defines the sorting criteria
	SortBy EventSortOption `json:"sortBy,omitempty"`

	// Limit is the maximum number of events to return
	Limit int `json:"limit,omitempty"`
}

// EventSortOption represents event sorting options
type EventSortOption string

const (
	// SortByTime sorts events by time
	SortByTime EventSortOption = "time"
	// SortByCount sorts events by count
	SortByCount EventSortOption = "count"
	// SortByType sorts events by type
	SortByType EventSortOption = "type"
	// SortByResource sorts events by resource
	SortByResource EventSortOption = "resource"
)

// EventOptions represents options for watching events
type EventOptions struct {
	// Filter specifies the event filter
	Filter *EventFilter `json:"filter,omitempty"`

	// IncludeManaged indicates whether to include managed fields
	IncludeManaged bool `json:"includeManaged,omitempty"`

	// BufferSize is the size of the event buffer
	BufferSize int `json:"bufferSize,omitempty"`
}

// FromCoreEvent converts a core event to an Event
func FromCoreEvent(e *corev1.Event) *Event {
	return &Event{
		Type:           EventType(e.Type),
		Name:           e.Name,
		Namespace:      e.Namespace,
		ResourceKind:   e.InvolvedObject.Kind,
		ResourceName:   e.InvolvedObject.Name,
		Reason:         e.Reason,
		Message:        e.Message,
		Component:      e.Source.Component,
		Host:           e.Source.Host,
		FirstTimestamp: e.FirstTimestamp.Time,
		LastTimestamp:  e.LastTimestamp.Time,
		Count:          e.Count,
		IsWarning:      e.Type == string(Warning),
	}
}
