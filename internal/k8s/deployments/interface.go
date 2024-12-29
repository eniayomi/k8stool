package deployments

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned"
)

// Service defines the interface for deployment operations
type Service interface {
	// List returns a list of deployments based on the given filters
	List(namespace string, allNamespaces bool, selector string) ([]Deployment, error)

	// Get returns a specific deployment by name
	Get(namespace, name string) (*Deployment, error)

	// Describe returns detailed information about a deployment
	Describe(namespace, name string) (*DeploymentDetails, error)

	// GetMetrics returns resource usage metrics for a deployment
	GetMetrics(namespace, name string) (*DeploymentMetrics, error)

	// Scale updates the number of replicas for a deployment
	Scale(namespace, name string, replicas int32) error

	// Update updates a deployment's configuration
	Update(namespace, name string, opts DeploymentOptions) error

	// AddMetrics adds metrics information to a list of deployments
	AddMetrics(deployments []Deployment) error
}

// NewDeploymentService creates a new deployment service instance
func NewDeploymentService(clientset *kubernetes.Clientset, metricsClient *metricsv1beta1.Clientset, config *rest.Config) (Service, error) {
	if clientset == nil {
		return nil, fmt.Errorf("kubernetes client is required")
	}
	if metricsClient == nil {
		return nil, fmt.Errorf("metrics client is required")
	}
	if config == nil {
		return nil, fmt.Errorf("rest config is required")
	}
	return newService(clientset, metricsClient, config), nil
}
