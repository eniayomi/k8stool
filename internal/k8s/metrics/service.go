package metrics

import (
	"context"
	"fmt"
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

type service struct {
	clientset     *kubernetes.Clientset
	metricsClient *metrics.Clientset
	config        *rest.Config
}

// newService creates a new metrics service instance
func newService(clientset *kubernetes.Clientset, metricsClient *metrics.Clientset, config *rest.Config) Service {
	return &service{
		clientset:     clientset,
		metricsClient: metricsClient,
		config:        config,
	}
}

// GetPodMetrics returns metrics for a specific pod
func (s *service) GetPodMetrics(namespace, name string) (*PodMetrics, error) {
	podMetrics, err := s.metricsClient.MetricsV1beta1().PodMetricses(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod metrics: %w", err)
	}

	pod, err := s.clientset.CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	metrics := &PodMetrics{
		Name:              podMetrics.Name,
		Namespace:         podMetrics.Namespace,
		CreationTimestamp: pod.CreationTimestamp.Time,
		Containers:        make(map[string]ResourceMetrics),
		TotalResources:    ResourceMetrics{},
	}

	for _, container := range podMetrics.Containers {
		containerMetrics := s.calculateContainerMetrics(container, pod)
		metrics.Containers[container.Name] = containerMetrics
		metrics.TotalResources.CPU.UsageNanoCores += containerMetrics.CPU.UsageNanoCores
		metrics.TotalResources.Memory.UsageBytes += containerMetrics.Memory.UsageBytes
	}

	return metrics, nil
}

// ListPodMetrics returns metrics for all pods in a namespace
func (s *service) ListPodMetrics(namespace string) ([]PodMetrics, error) {
	podMetricsList, err := s.metricsClient.MetricsV1beta1().PodMetricses(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pod metrics: %w", err)
	}

	var metrics []PodMetrics
	for _, podMetrics := range podMetricsList.Items {
		pod, err := s.clientset.CoreV1().Pods(namespace).Get(context.Background(), podMetrics.Name, metav1.GetOptions{})
		if err != nil {
			continue // Skip pods that can't be found
		}

		metric := PodMetrics{
			Name:              podMetrics.Name,
			Namespace:         podMetrics.Namespace,
			CreationTimestamp: pod.CreationTimestamp.Time,
			Containers:        make(map[string]ResourceMetrics),
			TotalResources:    ResourceMetrics{},
		}

		for _, container := range podMetrics.Containers {
			containerMetrics := s.calculateContainerMetrics(container, pod)
			metric.Containers[container.Name] = containerMetrics
			metric.TotalResources.CPU.UsageNanoCores += containerMetrics.CPU.UsageNanoCores
			metric.TotalResources.Memory.UsageBytes += containerMetrics.Memory.UsageBytes
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// GetNodeMetrics returns metrics for a specific node
func (s *service) GetNodeMetrics(name string) (*NodeMetrics, error) {
	nodeMetrics, err := s.metricsClient.MetricsV1beta1().NodeMetricses().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get node metrics: %w", err)
	}

	node, err := s.clientset.CoreV1().Nodes().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	pods, err := s.clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", name),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	metrics := &NodeMetrics{
		Name:              nodeMetrics.Name,
		CreationTimestamp: node.CreationTimestamp.Time,
		Resources:         s.calculateNodeMetrics(nodeMetrics),
		Allocatable:       s.calculateNodeResourceMetrics(node.Status.Allocatable),
		Capacity:          s.calculateNodeResourceMetrics(node.Status.Capacity),
		PodCount:          len(pods.Items),
	}

	return metrics, nil
}

// ListNodeMetrics returns metrics for all nodes
func (s *service) ListNodeMetrics() ([]NodeMetrics, error) {
	nodeMetricsList, err := s.metricsClient.MetricsV1beta1().NodeMetricses().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list node metrics: %w", err)
	}

	var metrics []NodeMetrics
	for _, nodeMetrics := range nodeMetricsList.Items {
		node, err := s.clientset.CoreV1().Nodes().Get(context.Background(), nodeMetrics.Name, metav1.GetOptions{})
		if err != nil {
			continue // Skip nodes that can't be found
		}

		pods, err := s.clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{
			FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeMetrics.Name),
		})
		if err != nil {
			continue // Skip nodes where we can't list pods
		}

		metric := NodeMetrics{
			Name:              nodeMetrics.Name,
			CreationTimestamp: node.CreationTimestamp.Time,
			Resources:         s.calculateNodeMetrics(&nodeMetrics),
			Allocatable:       s.calculateNodeResourceMetrics(node.Status.Allocatable),
			Capacity:          s.calculateNodeResourceMetrics(node.Status.Capacity),
			PodCount:          len(pods.Items),
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// Sort sorts metrics based on the given option
func (s *service) Sort(podMetrics []PodMetrics, sortBy MetricsSortOption) []PodMetrics {
	switch sortBy {
	case SortByName:
		sort.Slice(podMetrics, func(i, j int) bool {
			return podMetrics[i].Name < podMetrics[j].Name
		})
	case SortByCPU:
		sort.Slice(podMetrics, func(i, j int) bool {
			return podMetrics[i].TotalResources.CPU.UsageNanoCores > podMetrics[j].TotalResources.CPU.UsageNanoCores
		})
	case SortByMemory:
		sort.Slice(podMetrics, func(i, j int) bool {
			return podMetrics[i].TotalResources.Memory.UsageBytes > podMetrics[j].TotalResources.Memory.UsageBytes
		})
	case SortByAge:
		sort.Slice(podMetrics, func(i, j int) bool {
			return podMetrics[i].CreationTimestamp.Before(podMetrics[j].CreationTimestamp)
		})
	}
	return podMetrics
}

// Helper functions for calculating metrics

func (s *service) calculateContainerMetrics(containerMetrics interface{}, pod *corev1.Pod) ResourceMetrics {
	// Implementation depends on the metrics API version and structure
	// This is a placeholder that needs to be implemented based on the actual metrics API
	return ResourceMetrics{}
}

func (s *service) calculateNodeMetrics(nodeMetrics interface{}) ResourceMetrics {
	// Implementation depends on the metrics API version and structure
	// This is a placeholder that needs to be implemented based on the actual metrics API
	return ResourceMetrics{}
}

func (s *service) calculateNodeResourceMetrics(resources corev1.ResourceList) ResourceMetrics {
	// Implementation depends on the resource conversion logic
	// This is a placeholder that needs to be implemented based on the actual resource types
	return ResourceMetrics{}
}
