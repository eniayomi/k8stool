package k8s

import (
	"context"
	"fmt"
	"regexp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// BasicValidator implements basic validation for Kubernetes resources and operations
type BasicValidator struct {
	client kubernetes.Interface
}

// NewBasicValidator creates a new basic validator
func NewBasicValidator(client kubernetes.Interface) *BasicValidator {
	return &BasicValidator{
		client: client,
	}
}

// ValidateResource validates if a resource exists and is accessible
func (v *BasicValidator) ValidateResource(ctx context.Context, resourceType, name, namespace string) error {
	// Validate resource name format
	if err := v.validateResourceName(name); err != nil {
		return err
	}

	// Validate namespace exists
	if namespace != "" {
		if err := v.validateNamespace(ctx, namespace); err != nil {
			return err
		}
	}

	// Validate resource exists based on type
	switch resourceType {
	case "pod":
		_, err := v.client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("pod validation failed: %w", err)
		}
	case "deployment":
		_, err := v.client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("deployment validation failed: %w", err)
		}
	default:
		return fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	return nil
}

// ValidateOperation validates if an operation can be performed
func (v *BasicValidator) ValidateOperation(ctx context.Context, taskType TaskType, params map[string]interface{}) error {
	switch taskType {
	case TaskDeployScale:
		return v.validateScaleOperation(params)
	case TaskPodLogs:
		return v.validateLogsOperation(params)
	case TaskResourceApply:
		return v.validateApplyOperation(params)
	case TaskResourceDelete:
		return v.validateDeleteOperation(params)
	default:
		return nil // No specific validation for other operations
	}
}

// validateResourceName checks if the resource name follows Kubernetes naming conventions
func (v *BasicValidator) validateResourceName(name string) error {
	if name == "" {
		return fmt.Errorf("resource name cannot be empty")
	}

	// Kubernetes resource names must consist of lowercase alphanumeric characters, '-' or '.'
	matched, err := regexp.MatchString("^[a-z0-9][a-z0-9.-]*[a-z0-9]$", name)
	if err != nil {
		return fmt.Errorf("name validation error: %w", err)
	}
	if !matched {
		return fmt.Errorf("invalid resource name format: %s", name)
	}

	return nil
}

// validateNamespace checks if the namespace exists
func (v *BasicValidator) validateNamespace(ctx context.Context, namespace string) error {
	_, err := v.client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("namespace validation failed: %w", err)
	}
	return nil
}

// validateScaleOperation validates parameters for scaling operations
func (v *BasicValidator) validateScaleOperation(params map[string]interface{}) error {
	replicas, ok := params["replicas"].(int32)
	if !ok {
		return fmt.Errorf("replicas parameter must be an integer")
	}

	if replicas < 0 {
		return fmt.Errorf("replicas cannot be negative")
	}

	// Add a reasonable upper limit to prevent accidental large scale operations
	if replicas > 100 {
		return fmt.Errorf("scaling to more than 100 replicas requires manual confirmation")
	}

	return nil
}

// validateLogsOperation validates parameters for log operations
func (v *BasicValidator) validateLogsOperation(params map[string]interface{}) error {
	if tail, ok := params["tail"].(int64); ok {
		if tail < 0 {
			return fmt.Errorf("tail lines cannot be negative")
		}
		if tail > 10000 {
			return fmt.Errorf("requesting more than 10000 lines requires manual confirmation")
		}
	}

	return nil
}

// validateApplyOperation validates parameters for apply operations
func (v *BasicValidator) validateApplyOperation(params map[string]interface{}) error {
	// Add validation for apply operations
	// For example, validate YAML structure, resource limits, etc.
	return nil
}

// validateDeleteOperation validates parameters for delete operations
func (v *BasicValidator) validateDeleteOperation(params map[string]interface{}) error {
	// Add validation for delete operations
	// For example, prevent deletion of system resources
	return nil
}
