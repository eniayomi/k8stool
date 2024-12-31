package k8s

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeploymentHandler handles deployment-related operations
func (a *Agent) DeploymentHandler(ctx context.Context, params TaskParams) (*TaskResult, error) {
	// Validate the resource
	if err := a.ValidateResource(ctx, "deployment", params.ResourceName, params.Namespace); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	switch params.Action {
	case "inspect":
		return a.inspectDeployment(ctx, params)
	case "scale":
		// Validate scale operation
		if err := a.ValidateOperation(ctx, TaskDeployScale, params.ExtraParams); err != nil {
			return nil, fmt.Errorf("scale operation validation failed: %w", err)
		}
		return a.scaleDeployment(ctx, params)
	case "list":
		return a.listDeployments(ctx, params)
	default:
		return nil, fmt.Errorf("unsupported deployment action: %s", params.Action)
	}
}

// inspectDeployment retrieves detailed information about a deployment
func (a *Agent) inspectDeployment(ctx context.Context, params TaskParams) (*TaskResult, error) {
	// Get the deployment
	deployment, err := a.k8sClient.AppsV1().Deployments(params.Namespace).Get(ctx, params.ResourceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment %s: %w", params.ResourceName, err)
	}

	// Build the output
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Deployment: %s\n", deployment.Name))
	output.WriteString(fmt.Sprintf("Namespace: %s\n", deployment.Namespace))
	output.WriteString(fmt.Sprintf("Replicas: %d/%d\n", deployment.Status.ReadyReplicas, deployment.Status.Replicas))
	output.WriteString(fmt.Sprintf("Strategy: %s\n", deployment.Spec.Strategy.Type))

	// Pod template details
	output.WriteString("\nPod Template:\n")
	output.WriteString("Containers:\n")
	for _, container := range deployment.Spec.Template.Spec.Containers {
		output.WriteString(fmt.Sprintf("- %s:\n", container.Name))
		output.WriteString(fmt.Sprintf("  Image: %s\n", container.Image))
		if len(container.Ports) > 0 {
			output.WriteString("  Ports:\n")
			for _, port := range container.Ports {
				output.WriteString(fmt.Sprintf("  - %d/%s\n", port.ContainerPort, port.Protocol))
			}
		}
	}

	// Selector details
	output.WriteString("\nSelector:\n")
	for key, value := range deployment.Spec.Selector.MatchLabels {
		output.WriteString(fmt.Sprintf("- %s: %s\n", key, value))
	}

	return &TaskResult{
		Success: true,
		Output:  output.String(),
	}, nil
}

// scaleDeployment scales a deployment to a specified number of replicas
func (a *Agent) scaleDeployment(ctx context.Context, params TaskParams) (*TaskResult, error) {
	// Get replicas from params
	replicas, ok := params.ExtraParams["replicas"].(int32)
	if !ok {
		return nil, fmt.Errorf("replicas parameter is required and must be an integer")
	}

	// Get the deployment
	deployment, err := a.k8sClient.AppsV1().Deployments(params.Namespace).Get(ctx, params.ResourceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment %s: %w", params.ResourceName, err)
	}

	// Update replicas
	deployment.Spec.Replicas = &replicas

	// Apply the update
	_, err = a.k8sClient.AppsV1().Deployments(params.Namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to scale deployment: %w", err)
	}

	return &TaskResult{
		Success: true,
		Output:  fmt.Sprintf("Successfully scaled deployment %s to %d replicas", params.ResourceName, replicas),
	}, nil
}

// listDeployments lists all deployments in a namespace
func (a *Agent) listDeployments(ctx context.Context, params TaskParams) (*TaskResult, error) {
	// Get the deployments
	deployments, err := a.k8sClient.AppsV1().Deployments(params.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	// Build the output
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Deployments in namespace %s:\n", params.Namespace))
	output.WriteString("NAME\t\tREPLICAS\t\tAVAILABLE\t\tUP-TO-DATE\n")
	output.WriteString("----\t\t--------\t\t---------\t\t----------\n")

	for _, deployment := range deployments.Items {
		output.WriteString(fmt.Sprintf("%s\t\t%d/%d\t\t%d\t\t%d\n",
			deployment.Name,
			deployment.Status.ReadyReplicas,
			deployment.Status.Replicas,
			deployment.Status.AvailableReplicas,
			deployment.Status.UpdatedReplicas,
		))
	}

	return &TaskResult{
		Success: true,
		Output:  output.String(),
	}, nil
}
