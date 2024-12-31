package metrics

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

// Service defines the interface for metrics operations
type Service interface {
	// GetPodMetrics returns metrics for a specific pod
	GetPodMetrics(namespace, name string) (*PodMetrics, error)

	// ListPodMetrics returns metrics for all pods in a namespace
	ListPodMetrics(namespace string) ([]PodMetrics, error)

	// GetNodeMetrics returns metrics for a specific node
	GetNodeMetrics(name string) (*NodeMetrics, error)

	// ListNodeMetrics returns metrics for all nodes
	ListNodeMetrics() ([]NodeMetrics, error)

	// Sort sorts metrics based on the given option
	Sort(podMetrics []PodMetrics, sortBy MetricsSortOption) []PodMetrics
}

// NewMetricsService creates a new metrics service instance
func NewMetricsService(clientset *kubernetes.Clientset, metricsClient *metrics.Clientset, config *rest.Config) (Service, error) {
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
