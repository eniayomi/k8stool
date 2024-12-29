package events

import (
	"context"
	"fmt"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type service struct {
	clientset *kubernetes.Clientset
}

// NewEventService creates a new event service instance
func NewEventService(clientset *kubernetes.Clientset) (EventService, error) {
	if clientset == nil {
		return nil, fmt.Errorf("kubernetes clientset is required")
	}
	return &service{clientset: clientset}, nil
}

// List returns a list of events matching the given filter
func (s *service) List(ctx context.Context, namespace string, filter *EventFilter) (*EventList, error) {
	opts := metav1.ListOptions{}
	if filter != nil {
		var selectors []string

		// Apply filters
		if len(filter.Types) > 0 {
			types := make([]string, len(filter.Types))
			for i, t := range filter.Types {
				types[i] = string(t)
			}
			selectors = append(selectors, fmt.Sprintf("type=%s", strings.Join(types, ",")))
		}

		if len(filter.ResourceKinds) > 0 {
			selectors = append(selectors, fmt.Sprintf("involvedObject.kind=%s", strings.Join(filter.ResourceKinds, ",")))
		}

		if len(filter.ResourceNames) > 0 {
			selectors = append(selectors, fmt.Sprintf("involvedObject.name=%s", strings.Join(filter.ResourceNames, ",")))
		}

		if len(filter.Components) > 0 {
			selectors = append(selectors, fmt.Sprintf("source.component=%s", strings.Join(filter.Components, ",")))
		}

		if len(selectors) > 0 {
			opts.FieldSelector = strings.Join(selectors, ",")
		}
	}

	events, err := s.clientset.CoreV1().Events(namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	result := &EventList{
		Items: make([]Event, 0, len(events.Items)),
		Total: len(events.Items),
	}

	for _, event := range events.Items {
		if filter != nil && filter.Since != nil {
			if event.LastTimestamp.Time.Before(*filter.Since) {
				continue
			}
		}
		result.Items = append(result.Items, *FromCoreEvent(&event))
	}

	if filter != nil {
		// Apply sorting
		switch filter.SortBy {
		case SortByTime:
			sort.Slice(result.Items, func(i, j int) bool {
				return result.Items[i].LastTimestamp.After(result.Items[j].LastTimestamp)
			})
		case SortByCount:
			sort.Slice(result.Items, func(i, j int) bool {
				return result.Items[i].Count > result.Items[j].Count
			})
		case SortByType:
			sort.Slice(result.Items, func(i, j int) bool {
				return string(result.Items[i].Type) < string(result.Items[j].Type)
			})
		case SortByResource:
			sort.Slice(result.Items, func(i, j int) bool {
				if result.Items[i].ResourceKind == result.Items[j].ResourceKind {
					return result.Items[i].ResourceName < result.Items[j].ResourceName
				}
				return result.Items[i].ResourceKind < result.Items[j].ResourceKind
			})
		}

		// Apply limit
		if filter.Limit > 0 && len(result.Items) > filter.Limit {
			result.Items = result.Items[:filter.Limit]
		}
	}

	return result, nil
}

// ListForObject returns events related to a specific resource
func (s *service) ListForObject(ctx context.Context, namespace, kind, name string) (*EventList, error) {
	return s.List(ctx, namespace, &EventFilter{
		ResourceKinds: []string{kind},
		ResourceNames: []string{name},
		SortBy:        SortByTime,
	})
}

// Watch watches for events matching the given filter
func (s *service) Watch(ctx context.Context, namespace string, opts *EventOptions) (<-chan Event, error) {
	if opts == nil {
		opts = &EventOptions{
			BufferSize: 100,
		}
	}

	watchOpts := metav1.ListOptions{
		Watch: true,
	}

	if opts.Filter != nil {
		var selectors []string

		if len(opts.Filter.Types) > 0 {
			types := make([]string, len(opts.Filter.Types))
			for i, t := range opts.Filter.Types {
				types[i] = string(t)
			}
			selectors = append(selectors, fmt.Sprintf("type=%s", strings.Join(types, ",")))
		}

		if len(opts.Filter.ResourceKinds) > 0 {
			selectors = append(selectors, fmt.Sprintf("involvedObject.kind=%s", strings.Join(opts.Filter.ResourceKinds, ",")))
		}

		if len(opts.Filter.ResourceNames) > 0 {
			selectors = append(selectors, fmt.Sprintf("involvedObject.name=%s", strings.Join(opts.Filter.ResourceNames, ",")))
		}

		if len(selectors) > 0 {
			watchOpts.FieldSelector = strings.Join(selectors, ",")
		}
	}

	watcher, err := s.clientset.CoreV1().Events(namespace).Watch(ctx, watchOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to watch events: %w", err)
	}

	eventChan := make(chan Event, opts.BufferSize)

	go func() {
		defer watcher.Stop()
		defer close(eventChan)

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.ResultChan():
				if !ok {
					return
				}
				if e, ok := event.Object.(*corev1.Event); ok {
					eventChan <- *FromCoreEvent(e)
				}
			}
		}
	}()

	return eventChan, nil
}

// Get returns a specific event by name
func (s *service) Get(ctx context.Context, namespace, name string) (*Event, error) {
	event, err := s.clientset.CoreV1().Events(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return FromCoreEvent(event), nil
}
