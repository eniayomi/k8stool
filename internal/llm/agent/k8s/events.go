package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EventsHandler handles Kubernetes events operations
func (a *Agent) EventsHandler(ctx context.Context, params TaskParams) (*TaskResult, error) {
	switch params.Action {
	case "get", "list":
		return a.getEvents(ctx, params)
	case "watch":
		return a.watchEvents(ctx, params)
	default:
		return nil, fmt.Errorf("unsupported events action: %s", params.Action)
	}
}

// getEvents retrieves events for a namespace or specific resource
func (a *Agent) getEvents(ctx context.Context, params TaskParams) (*TaskResult, error) {
	namespace := params.Namespace
	if namespace == "" {
		namespace = a.k8sContext.Namespace
	}

	listOptions := metav1.ListOptions{
		Limit: 50, // Limit to last 50 events by default
	}

	// If resource name is specified, filter events for that resource
	if params.ResourceName != "" {
		listOptions.FieldSelector = fmt.Sprintf("involvedObject.name=%s", params.ResourceName)
		if params.ResourceType != "" {
			listOptions.FieldSelector += fmt.Sprintf(",involvedObject.kind=%s", params.ResourceType)
		}
	}

	events, err := a.k8sClient.CoreV1().Events(namespace).List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	var output strings.Builder
	output.WriteString("LAST SEEN\tTYPE\tREASON\tOBJECT\tMESSAGE\n")
	output.WriteString("---------\t----\t------\t------\t-------\n")

	for _, event := range events.Items {
		lastSeen := time.Since(event.LastTimestamp.Time).Round(time.Second)
		object := fmt.Sprintf("%s/%s", strings.ToLower(event.InvolvedObject.Kind), event.InvolvedObject.Name)
		output.WriteString(fmt.Sprintf("%v\t%s\t%s\t%s\t%s\n",
			lastSeen,
			event.Type,
			event.Reason,
			object,
			event.Message,
		))
	}

	return &TaskResult{
		Success: true,
		Output:  output.String(),
	}, nil
}

// watchEvents watches events in real-time
func (a *Agent) watchEvents(ctx context.Context, params TaskParams) (*TaskResult, error) {
	namespace := params.Namespace
	if namespace == "" {
		namespace = a.k8sContext.Namespace
	}

	listOptions := metav1.ListOptions{
		Watch: true,
	}

	// If resource name is specified, filter events for that resource
	if params.ResourceName != "" {
		listOptions.FieldSelector = fmt.Sprintf("involvedObject.name=%s", params.ResourceName)
		if params.ResourceType != "" {
			listOptions.FieldSelector += fmt.Sprintf(",involvedObject.kind=%s", params.ResourceType)
		}
	}

	watcher, err := a.k8sClient.CoreV1().Events(namespace).Watch(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to watch events: %w", err)
	}
	defer watcher.Stop()

	var output strings.Builder
	output.WriteString("Watching events... (Press Ctrl+C to stop)\n\n")
	output.WriteString("TIME\tTYPE\tREASON\tOBJECT\tMESSAGE\n")
	output.WriteString("----\t----\t------\t------\t-------\n")

	for event := range watcher.ResultChan() {
		e, ok := event.Object.(*corev1.Event)
		if !ok {
			continue
		}

		object := fmt.Sprintf("%s/%s", strings.ToLower(e.InvolvedObject.Kind), e.InvolvedObject.Name)
		output.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\n",
			e.LastTimestamp.Format("15:04:05"),
			e.Type,
			e.Reason,
			object,
			e.Message,
		))
	}

	return &TaskResult{
		Success: true,
		Output:  output.String(),
	}, nil
}
