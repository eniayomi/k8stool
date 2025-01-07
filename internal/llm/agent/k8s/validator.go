package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// BasicValidator provides basic validation for Kubernetes resources and operations
type BasicValidator struct {
	client kubernetes.Interface
}

// NewBasicValidator creates a new basic validator
func NewBasicValidator(client kubernetes.Interface) *BasicValidator {
	return &BasicValidator{
		client: client,
	}
}

// ValidateResource validates if a Kubernetes resource exists
func (v *BasicValidator) ValidateResource(ctx context.Context, resourceType, name, namespace string) error {
	// For list operations, no need to validate specific resource
	if name == "" {
		return nil
	}

	switch resourceType {
	case "pod", "pods":
		_, err := v.client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to validate pod %s/%s: %w", namespace, name, err)
		}
	case "deployment", "deployments":
		_, err := v.client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to validate deployment %s/%s: %w", namespace, name, err)
		}
	case "namespace", "namespaces":
		if name == "" {
			return nil // Current namespace is always valid
		}
		_, err := v.client.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to validate namespace %s: %w", name, err)
		}
	case "service", "services":
		_, err := v.client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to validate service %s/%s: %w", namespace, name, err)
		}
	}

	return nil
}

// ValidateOperation validates if a Kubernetes operation is allowed
func (v *BasicValidator) ValidateOperation(ctx context.Context, taskType TaskType, params map[string]interface{}) error {
	switch taskType {
	case TaskPodLogs, TaskExec:
		// These operations require a specific pod
		if name, ok := params["name"].(string); !ok || name == "" {
			return fmt.Errorf("pod name is required for %s operation", taskType)
		}
	case TaskDeployScale:
		// Scale operation requires a deployment and replicas
		if name, ok := params["name"].(string); !ok || name == "" {
			return fmt.Errorf("deployment name is required for scale operation")
		}
		if _, ok := params["replicas"].(int32); !ok {
			return fmt.Errorf("replicas count is required for scale operation")
		}
	case TaskList, TaskNamespaceList, TaskContextList:
		// List operations don't require specific resource names
		return nil
	}

	return nil
}
