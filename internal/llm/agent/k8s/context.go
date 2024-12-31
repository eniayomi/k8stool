package k8s

import (
	"context"
	"fmt"
	"strings"

	k8scontext "k8stool/internal/k8s/context"
)

// ContextHandler handles context-related operations
func (a *Agent) ContextHandler(ctx context.Context, params TaskParams) (*TaskResult, error) {
	// Create context service
	contextService, err := k8scontext.NewContextOnlyService()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize context service: %w", err)
	}

	switch params.Action {
	case "get":
		return a.getCurrentContext(contextService)
	case "list":
		return a.listContexts(contextService)
	case "switch", "set", "change":
		return a.switchContext(contextService, params)
	default:
		return nil, fmt.Errorf("unsupported context action: %s", params.Action)
	}
}

// getCurrentContext gets information about the current context
func (a *Agent) getCurrentContext(service k8scontext.Service) (*TaskResult, error) {
	current, err := service.GetCurrent()
	if err != nil {
		return nil, fmt.Errorf("failed to get current context: %w", err)
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Current Context: %s\n", current.Name))
	output.WriteString(fmt.Sprintf("Cluster: %s\n", current.Cluster))
	output.WriteString(fmt.Sprintf("User: %s\n", current.User))
	if current.Namespace != "" {
		output.WriteString(fmt.Sprintf("Namespace: %s\n", current.Namespace))
	}

	return &TaskResult{
		Success: true,
		Output:  output.String(),
	}, nil
}

// listContexts lists all available contexts
func (a *Agent) listContexts(service k8scontext.Service) (*TaskResult, error) {
	contexts, err := service.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list contexts: %w", err)
	}

	// Sort contexts by name
	contexts = service.Sort(contexts, k8scontext.SortByName)

	var output strings.Builder
	output.WriteString("Available Contexts:\n")
	output.WriteString("NAME\t\tCLUSTER\t\tUSER\t\tNAMESPACE\n")
	output.WriteString("----\t\t-------\t\t----\t\t---------\n")

	for _, ctx := range contexts {
		current := " "
		if ctx.IsActive {
			current = "*"
		}
		output.WriteString(fmt.Sprintf("%s %s\t\t%s\t\t%s\t\t%s\n",
			current,
			ctx.Name,
			ctx.Cluster,
			ctx.User,
			ctx.Namespace,
		))
	}

	return &TaskResult{
		Success: true,
		Output:  output.String(),
	}, nil
}

// switchContext switches to a different context
func (a *Agent) switchContext(service k8scontext.Service, params TaskParams) (*TaskResult, error) {
	if params.ResourceName == "" {
		return nil, fmt.Errorf("context name is required")
	}

	if err := service.SwitchContext(params.ResourceName); err != nil {
		return nil, fmt.Errorf("failed to switch context: %w", err)
	}

	// Get the new context to update our agent's state
	current, err := service.GetCurrent()
	if err != nil {
		return nil, fmt.Errorf("failed to get current context: %w", err)
	}

	// Update the agent's context
	a.k8sContext = &K8sContext{
		CurrentContext: current.Name,
		Namespace:      current.Namespace,
		ClusterInfo:    current.Cluster,
	}

	return &TaskResult{
		Success: true,
		Output:  fmt.Sprintf("Switched to context %q\n", params.ResourceName),
	}, nil
}
