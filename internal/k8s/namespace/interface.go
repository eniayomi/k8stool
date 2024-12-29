package namespace

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Service defines the interface for namespace operations
type Service interface {
	// List returns all available namespaces
	List() ([]Namespace, error)

	// Get returns details for a specific namespace
	Get(name string) (*NamespaceDetails, error)

	// Create creates a new namespace
	Create(name string, labels, annotations map[string]string) error

	// Delete deletes a namespace
	Delete(name string) error

	// GetResourceQuotas returns resource quotas for a namespace
	GetResourceQuotas(namespace string) ([]ResourceQuota, error)

	// GetLimitRanges returns limit ranges for a namespace
	GetLimitRanges(namespace string) ([]LimitRange, error)

	// Sort sorts namespaces based on the given option
	Sort(namespaces []Namespace, sortBy NamespaceSortOption) []Namespace
}

// NewNamespaceService creates a new namespace service instance
func NewNamespaceService(clientset *kubernetes.Clientset, config *rest.Config) (Service, error) {
	if clientset == nil {
		return nil, fmt.Errorf("kubernetes client is required")
	}
	if config == nil {
		return nil, fmt.Errorf("rest config is required")
	}
	return newService(clientset, config), nil
}
