package k8s

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// ExecHandler handles container exec operations
func (a *Agent) ExecHandler(ctx context.Context, params TaskParams) (*TaskResult, error) {
	switch params.Action {
	case "exec", "run":
		return a.execInContainer(ctx, params)
	default:
		return nil, fmt.Errorf("unsupported exec action: %s", params.Action)
	}
}

// execInContainer executes a command in a container
func (a *Agent) execInContainer(ctx context.Context, params TaskParams) (*TaskResult, error) {
	if params.ResourceName == "" {
		return nil, fmt.Errorf("pod name is required")
	}

	namespace := params.Namespace
	if namespace == "" {
		namespace = a.k8sContext.Namespace
	}

	// Validate pod exists
	if err := a.ValidateResource(ctx, "pod", params.ResourceName, namespace); err != nil {
		return nil, err
	}

	// Get pod to find container
	pod, err := a.k8sClient.CoreV1().Pods(namespace).Get(ctx, params.ResourceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	// Determine container name
	containerName := ""
	if params.ContainerName != "" {
		// Verify container exists
		found := false
		for _, container := range pod.Spec.Containers {
			if container.Name == params.ContainerName {
				found = true
				containerName = container.Name
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("container %q not found in pod %q", params.ContainerName, params.ResourceName)
		}
	} else {
		// Use first container if not specified
		if len(pod.Spec.Containers) == 0 {
			return nil, fmt.Errorf("no containers found in pod %q", params.ResourceName)
		}
		containerName = pod.Spec.Containers[0].Name
	}

	// Create exec request
	req := a.k8sClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(params.ResourceName).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: containerName,
		Command:   params.Command,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(a.k8sConfig, "POST", req.URL())
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	// Create buffers for output
	var stdout, stderr strings.Builder

	// Execute command
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w\nStderr: %s", err, stderr.String())
	}

	// Combine output
	var output strings.Builder
	if stdout.Len() > 0 {
		output.WriteString("Output:\n")
		output.WriteString(stdout.String())
	}
	if stderr.Len() > 0 {
		if output.Len() > 0 {
			output.WriteString("\n")
		}
		output.WriteString("Error Output:\n")
		output.WriteString(stderr.String())
	}

	return &TaskResult{
		Success: true,
		Output:  output.String(),
	}, nil
}
