package k8s

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodHandler handles pod-related operations
func (a *Agent) PodHandler(ctx context.Context, params TaskParams) (*TaskResult, error) {
	// Validate the resource
	if err := a.ValidateResource(ctx, "pod", params.ResourceName, params.Namespace); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	switch params.Action {
	case "inspect":
		return a.inspectPod(ctx, params)
	case "logs":
		// Validate log operation
		if err := a.ValidateOperation(ctx, TaskPodLogs, params.ExtraParams); err != nil {
			return nil, fmt.Errorf("log operation validation failed: %w", err)
		}
		return a.getPodLogs(ctx, params)
	case "list":
		return a.listPods(ctx, params)
	default:
		return nil, fmt.Errorf("unsupported pod action: %s", params.Action)
	}
}

// inspectPod retrieves detailed information about a pod
func (a *Agent) inspectPod(ctx context.Context, params TaskParams) (*TaskResult, error) {
	// Get the pod
	pod, err := a.k8sClient.CoreV1().Pods(params.Namespace).Get(ctx, params.ResourceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod %s: %w", params.ResourceName, err)
	}

	// Build the output
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Pod: %s\n", pod.Name))
	output.WriteString(fmt.Sprintf("Namespace: %s\n", pod.Namespace))
	output.WriteString(fmt.Sprintf("Status: %s\n", pod.Status.Phase))
	output.WriteString(fmt.Sprintf("Node: %s\n", pod.Spec.NodeName))
	output.WriteString(fmt.Sprintf("IP: %s\n", pod.Status.PodIP))

	// Container details
	output.WriteString("\nContainers:\n")
	for _, container := range pod.Spec.Containers {
		output.WriteString(fmt.Sprintf("- %s:\n", container.Name))
		output.WriteString(fmt.Sprintf("  Image: %s\n", container.Image))
		if len(container.Ports) > 0 {
			output.WriteString("  Ports:\n")
			for _, port := range container.Ports {
				output.WriteString(fmt.Sprintf("  - %d/%s\n", port.ContainerPort, port.Protocol))
			}
		}
	}

	return &TaskResult{
		Success: true,
		Output:  output.String(),
	}, nil
}

// getPodLogs retrieves logs from a pod
func (a *Agent) getPodLogs(ctx context.Context, params TaskParams) (*TaskResult, error) {
	// Get log options from params
	tailLines := int64(100) // Default to last 100 lines
	if val, ok := params.ExtraParams["tail"]; ok {
		if lines, ok := val.(int64); ok {
			tailLines = lines
		}
	}

	container := ""
	if val, ok := params.ExtraParams["container"]; ok {
		if name, ok := val.(string); ok {
			container = name
		}
	}

	// Set up log options
	logOptions := &corev1.PodLogOptions{
		Container: container,
		TailLines: &tailLines,
	}

	// Get the logs
	req := a.k8sClient.CoreV1().Pods(params.Namespace).GetLogs(params.ResourceName, logOptions)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pod logs: %w", err)
	}
	defer podLogs.Close()

	// Read the logs
	buf := make([]byte, 2048)
	var output strings.Builder
	for {
		n, err := podLogs.Read(buf)
		if n == 0 {
			break
		}
		if err != nil {
			break
		}
		output.Write(buf[:n])
	}

	return &TaskResult{
		Success: true,
		Output:  output.String(),
	}, nil
}

// listPods lists all pods in a namespace
func (a *Agent) listPods(ctx context.Context, params TaskParams) (*TaskResult, error) {
	// Get the pods
	pods, err := a.k8sClient.CoreV1().Pods(params.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	// Build the output
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Pods in namespace %s:\n", params.Namespace))
	output.WriteString("NAME\t\tSTATUS\t\tNODE\t\tREADY\n")
	output.WriteString("----\t\t------\t\t----\t\t-----\n")

	for _, pod := range pods.Items {
		ready := fmt.Sprintf("%d/%d", getPodReadyContainers(pod.Status.ContainerStatuses), len(pod.Spec.Containers))
		output.WriteString(fmt.Sprintf("%s\t\t%s\t\t%s\t\t%s\n",
			pod.Name,
			pod.Status.Phase,
			pod.Spec.NodeName,
			ready,
		))
	}

	return &TaskResult{
		Success: true,
		Output:  output.String(),
	}, nil
}

// Helper functions

// getPodContainerStatus returns the ready status of a container
func getPodContainerStatus(statuses []corev1.ContainerStatus, containerName string) bool {
	for _, status := range statuses {
		if status.Name == containerName {
			return status.Ready
		}
	}
	return false
}

// getPodReadyContainers returns the number of ready containers in a pod
func getPodReadyContainers(statuses []corev1.ContainerStatus) int {
	ready := 0
	for _, status := range statuses {
		if status.Ready {
			ready++
		}
	}
	return ready
}
