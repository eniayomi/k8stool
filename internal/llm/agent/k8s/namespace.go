package k8s

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespaceHandler handles namespace-related operations
func (a *Agent) NamespaceHandler(ctx context.Context, params TaskParams) (*TaskResult, error) {
	switch params.Action {
	case "get":
		return &TaskResult{
			Output:  fmt.Sprintf("Current Namespace: %s", a.k8sContext.Namespace),
			Success: true,
		}, nil
	case "list":
		namespaces, err := a.k8sClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to list namespaces: %w", err)
		}

		var output strings.Builder
		output.WriteString("Available Namespaces:\n")
		for _, ns := range namespaces.Items {
			if ns.Name == a.k8sContext.Namespace {
				output.WriteString(fmt.Sprintf("* %s (current)\n", ns.Name))
			} else {
				output.WriteString(fmt.Sprintf("  %s\n", ns.Name))
			}
		}

		return &TaskResult{
			Output:  output.String(),
			Success: true,
		}, nil
	case "switch", "use":
		if params.ResourceName == "" {
			return nil, fmt.Errorf("namespace name is required")
		}

		// Verify namespace exists
		_, err := a.k8sClient.CoreV1().Namespaces().Get(ctx, params.ResourceName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("namespace %q not found: %w", params.ResourceName, err)
		}

		// Update kubeconfig
		cmd := exec.Command("kubectl", "config", "set-context", "--current", "--namespace", params.ResourceName)
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("failed to switch namespace: %w", err)
		}

		// Update agent's context
		a.k8sContext.Namespace = params.ResourceName

		return &TaskResult{
			Output:  fmt.Sprintf("Switched to namespace %q", params.ResourceName),
			Success: true,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported namespace action: %s", params.Action)
	}
}
