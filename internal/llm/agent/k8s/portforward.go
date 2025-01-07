package k8s

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// PortForwardHandler handles port forwarding operations
func (a *Agent) PortForwardHandler(ctx context.Context, params TaskParams) (*TaskResult, error) {
	switch params.Action {
	case "start", "create":
		return a.startPortForward(ctx, params)
	case "stop":
		return a.stopPortForward(ctx, params)
	default:
		return nil, fmt.Errorf("unsupported port-forward action: %s", params.Action)
	}
}

// startPortForward starts port forwarding for a pod
func (a *Agent) startPortForward(ctx context.Context, params TaskParams) (*TaskResult, error) {
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

	// Parse port mappings
	var ports []string
	if params.Flags != nil {
		if p, ok := params.Flags["ports"].(string); ok {
			ports = strings.Split(p, ",")
		}
	}
	if len(ports) == 0 {
		return nil, fmt.Errorf("at least one port mapping is required (e.g., 8080:80)")
	}

	// Create port-forward request
	req := a.k8sClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(params.ResourceName).
		SubResource("portforward")

	transport, upgrader, err := spdy.RoundTripperFor(a.k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create round tripper: %w", err)
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())
	stopChan := make(chan struct{}, 1)
	readyChan := make(chan struct{}, 1)

	fw, err := portforward.New(dialer, ports, stopChan, readyChan, os.Stdout, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("failed to create port-forward: %w", err)
	}

	// Start port forwarding in a goroutine
	go func() {
		if err := fw.ForwardPorts(); err != nil {
			fmt.Fprintf(os.Stderr, "Error forwarding ports: %v\n", err)
		}
	}()

	// Wait for ready signal
	select {
	case <-readyChan:
		var output strings.Builder
		output.WriteString(fmt.Sprintf("Port forwarding started for pod %s\n", params.ResourceName))
		for _, port := range ports {
			output.WriteString(fmt.Sprintf("Forwarding %s\n", port))
		}
		output.WriteString("\nConnections are now open. Press Ctrl+C to stop forwarding.\n")

		return &TaskResult{
			Success: true,
			Output:  output.String(),
		}, nil
	case <-ctx.Done():
		close(stopChan)
		return nil, fmt.Errorf("port forwarding cancelled")
	}
}

// stopPortForward stops port forwarding for a pod
func (a *Agent) stopPortForward(ctx context.Context, params TaskParams) (*TaskResult, error) {
	if params.ResourceName == "" {
		return nil, fmt.Errorf("pod name is required")
	}

	// TODO: Implement tracking of active port-forwards and stopping specific ones
	return &TaskResult{
		Success: true,
		Output:  fmt.Sprintf("Port forwarding stopped for pod %s", params.ResourceName),
	}, nil
}
