package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	k8scontext "k8stool/internal/k8s/context"
	k8snamespace "k8stool/internal/k8s/namespace"

	"k8s.io/client-go/kubernetes"
)

// NamespaceHandler handles namespace-related operations
func (a *Agent) NamespaceHandler(ctx context.Context, params TaskParams) (*TaskResult, error) {
	// Create namespace service
	clientset, ok := a.k8sClient.(*kubernetes.Clientset)
	if !ok {
		return nil, fmt.Errorf("invalid kubernetes client type")
	}

	namespaceService, err := k8snamespace.NewNamespaceService(clientset, a.k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize namespace service: %w", err)
	}

	// Create context service for namespace switching
	contextService, err := k8scontext.NewContextOnlyService()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize context service: %w", err)
	}

	switch params.Action {
	case "get":
		return a.getCurrentNamespace(namespaceService, params)
	case "list":
		return a.listNamespaces(namespaceService)
	case "switch", "set", "change":
		return a.switchNamespace(namespaceService, contextService, params)
	default:
		return nil, fmt.Errorf("unsupported namespace action: %s", params.Action)
	}
}

// getCurrentNamespace gets information about the current or specified namespace
func (a *Agent) getCurrentNamespace(service k8snamespace.Service, params TaskParams) (*TaskResult, error) {
	name := params.ResourceName
	if name == "" {
		name = a.k8sContext.Namespace
	}

	details, err := service.Get(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace details: %w", err)
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Namespace: %s\n", details.Name))
	output.WriteString(fmt.Sprintf("Status: %s\n", details.Status))
	output.WriteString(fmt.Sprintf("Created: %s\n", details.CreationTimestamp.Format("2006-01-02 15:04:05")))

	if len(details.Labels) > 0 {
		output.WriteString("\nLabels:\n")
		for k, v := range details.Labels {
			output.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
	}

	if len(details.ResourceQuotas) > 0 {
		output.WriteString("\nResource Quotas:\n")
		for _, quota := range details.ResourceQuotas {
			output.WriteString(fmt.Sprintf("  %s:\n", quota.Name))
			for resource, hard := range quota.Hard {
				used := quota.Used[resource]
				output.WriteString(fmt.Sprintf("    %s: %s/%s\n", resource, used, hard))
			}
		}
	}

	return &TaskResult{
		Success: true,
		Output:  output.String(),
	}, nil
}

// listNamespaces lists all available namespaces
func (a *Agent) listNamespaces(service k8snamespace.Service) (*TaskResult, error) {
	namespaces, err := service.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	// Sort namespaces by name
	namespaces = service.Sort(namespaces, k8snamespace.SortByName)

	var output strings.Builder
	output.WriteString("Available Namespaces:\n")
	output.WriteString("NAME\t\tSTATUS\t\tAGE\n")
	output.WriteString("----\t\t------\t\t---\n")

	for _, ns := range namespaces {
		current := " "
		if ns.Name == a.k8sContext.Namespace {
			current = "*"
		}
		age := time.Since(ns.CreationTimestamp).Round(time.Second)
		output.WriteString(fmt.Sprintf("%s %s\t\t%s\t\t%s\n",
			current,
			ns.Name,
			ns.Status,
			age,
		))
	}

	return &TaskResult{
		Success: true,
		Output:  output.String(),
	}, nil
}

// switchNamespace switches to a different namespace
func (a *Agent) switchNamespace(service k8snamespace.Service, contextService k8scontext.Service, params TaskParams) (*TaskResult, error) {
	if params.ResourceName == "" {
		return nil, fmt.Errorf("namespace name is required")
	}

	// Verify namespace exists
	_, err := service.Get(params.ResourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace: %w", err)
	}

	// Update the context's namespace
	if err := contextService.SetNamespace(params.ResourceName); err != nil {
		return nil, fmt.Errorf("failed to switch namespace: %w", err)
	}

	// Update the agent's context
	a.k8sContext.Namespace = params.ResourceName

	return &TaskResult{
		Success: true,
		Output:  fmt.Sprintf("Switched to namespace %q\n", params.ResourceName),
	}, nil
}
