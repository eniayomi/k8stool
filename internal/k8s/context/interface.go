package context

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Service defines the interface for context operations
type Service interface {
	// List returns all available contexts
	List() ([]Context, error)

	// GetCurrent returns the current context
	GetCurrent() (*Context, error)

	// SwitchContext switches to a different context
	SwitchContext(name string) error

	// SetNamespace sets the default namespace for the current context
	SetNamespace(namespace string) error

	// GetClusterInfo returns information about the current cluster
	GetClusterInfo() (*ClusterInfo, error)

	// Sort sorts contexts based on the given option
	Sort(contexts []Context, sortBy ContextSortOption) []Context
}

// NewContextService creates a new context service instance
func NewContextService(clientset *kubernetes.Clientset, config *rest.Config, kubeconfig clientcmd.ClientConfig) (Service, error) {
	if clientset == nil {
		return nil, fmt.Errorf("kubernetes client is required")
	}
	if config == nil {
		return nil, fmt.Errorf("rest config is required")
	}
	if kubeconfig == nil {
		return nil, fmt.Errorf("kubeconfig is required")
	}
	return newService(clientset, config, kubeconfig), nil
}
