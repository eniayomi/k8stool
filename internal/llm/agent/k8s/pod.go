package k8s

import (
	"context"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// formatDuration formats a duration into a human-readable string
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "0m"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	if d < 30*24*time.Hour {
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

// PodHandler handles pod-related operations
func (a *Agent) PodHandler(ctx context.Context, params TaskParams) (*TaskResult, error) {
	switch params.Action {
	case "logs":
		// Get logs for the specified pod
		req := a.k8sClient.CoreV1().Pods(params.Namespace).GetLogs(params.ResourceName, &corev1.PodLogOptions{
			Container: params.ContainerName,
			Follow:    false,
			Previous:  false,
		})

		logs, err := req.Stream(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get logs stream: %w", err)
		}
		defer logs.Close()

		// Read logs
		logBytes, err := io.ReadAll(logs)
		if err != nil {
			return nil, fmt.Errorf("failed to read logs: %w", err)
		}

		return &TaskResult{
			Output:  string(logBytes),
			Success: true,
		}, nil

	case "list":
		// Build list options
		opts := metav1.ListOptions{}

		// Handle status filtering
		if params.Flags != nil {
			if status, ok := params.Flags["status"].(string); ok {
				opts.FieldSelector = fmt.Sprintf("status.phase=%s", status)
			}
		}

		// Get pods
		pods, err := a.k8sClient.CoreV1().Pods(params.Namespace).List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list pods: %w", err)
		}

		// Filter failed pods if requested
		var filteredPods []corev1.Pod
		if params.Flags != nil && params.Flags["status"] == "Failed" {
			for _, pod := range pods.Items {
				if pod.Status.Phase == corev1.PodFailed {
					filteredPods = append(filteredPods, pod)
				}
			}
			pods.Items = filteredPods
		}

		// Format output
		var output strings.Builder
		w := tabwriter.NewWriter(&output, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tREADY\tRESTARTS\tIP\tNODE\tAGE\tSTATUS")

		for _, pod := range pods.Items {
			// Get container ready count
			readyCount := 0
			totalCount := len(pod.Spec.Containers)
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.Ready {
					readyCount++
				}
			}

			// Calculate restarts
			restarts := 0
			for _, containerStatus := range pod.Status.ContainerStatuses {
				restarts += int(containerStatus.RestartCount)
			}

			// Calculate age
			age := time.Since(pod.CreationTimestamp.Time)
			ageStr := formatDuration(age)

			fmt.Fprintf(w, "%s\t%d/%d\t%d\t%s\t%s\t%s\t%s\n",
				pod.Name,
				readyCount,
				totalCount,
				restarts,
				pod.Status.PodIP,
				pod.Spec.NodeName,
				ageStr,
				string(pod.Status.Phase))
		}
		w.Flush()

		return &TaskResult{
			Output:  output.String(),
			Success: true,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported action: %s", params.Action)
	}
}
