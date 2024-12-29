package deployments

import (
	"context"
	"fmt"
	"time"

	"k8stool/internal/k8s/pods"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned"
)

type service struct {
	clientset     *kubernetes.Clientset
	metricsClient *metricsv1beta1.Clientset
	config        *rest.Config
}

// newService creates a new deployment service instance
func newService(clientset *kubernetes.Clientset, metricsClient *metricsv1beta1.Clientset, config *rest.Config) Service {
	return &service{
		clientset:     clientset,
		metricsClient: metricsClient,
		config:        config,
	}
}

// List returns a list of deployments based on the given filters
func (s *service) List(namespace string, allNamespaces bool, selector string) ([]Deployment, error) {
	var deployments []Deployment
	var listOptions metav1.ListOptions

	if selector != "" {
		listOptions.LabelSelector = selector
	}

	if allNamespaces {
		namespace = ""
	}

	deployList, err := s.clientset.AppsV1().Deployments(namespace).List(context.Background(), listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	for _, d := range deployList.Items {
		deployment := Deployment{
			Name:              d.Name,
			Namespace:         d.Namespace,
			Replicas:          *d.Spec.Replicas,
			ReadyReplicas:     d.Status.ReadyReplicas,
			UpdatedReplicas:   d.Status.UpdatedReplicas,
			AvailableReplicas: d.Status.AvailableReplicas,
			Age:               time.Since(d.CreationTimestamp.Time),
			Status:            getDeploymentStatus(d),
			Selector:          d.Spec.Selector,
		}
		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

// Get returns a specific deployment by name
func (s *service) Get(namespace, name string) (*Deployment, error) {
	d, err := s.clientset.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	deployment := &Deployment{
		Name:              d.Name,
		Namespace:         d.Namespace,
		Replicas:          *d.Spec.Replicas,
		ReadyReplicas:     d.Status.ReadyReplicas,
		UpdatedReplicas:   d.Status.UpdatedReplicas,
		AvailableReplicas: d.Status.AvailableReplicas,
		Age:               time.Since(d.CreationTimestamp.Time),
		Status:            getDeploymentStatus(*d),
		Selector:          d.Spec.Selector,
	}

	return deployment, nil
}

// Describe returns detailed information about a deployment
func (s *service) Describe(namespace, name string) (*DeploymentDetails, error) {
	d, err := s.clientset.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	details := &DeploymentDetails{
		Name:              d.Name,
		Namespace:         d.Namespace,
		Replicas:          *d.Spec.Replicas,
		UpdatedReplicas:   d.Status.UpdatedReplicas,
		ReadyReplicas:     d.Status.ReadyReplicas,
		AvailableReplicas: d.Status.AvailableReplicas,
		Strategy:          string(d.Spec.Strategy.Type),
		MinReadySeconds:   d.Spec.MinReadySeconds,
		Age:               time.Since(d.CreationTimestamp.Time),
		Labels:            d.Labels,
		Selector:          d.Spec.Selector.MatchLabels,
	}

	// Add container information
	for _, c := range d.Spec.Template.Spec.Containers {
		container := pods.ContainerInfo{
			Name:  c.Name,
			Image: c.Image,
		}
		details.Containers = append(details.Containers, container)
	}

	// Get events
	events, err := s.getDeploymentEvents(namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment events: %w", err)
	}
	details.Events = events

	return details, nil
}

// GetMetrics returns resource usage metrics for a deployment
func (s *service) GetMetrics(namespace, name string) (*DeploymentMetrics, error) {
	// Get deployment to get selector
	d, err := s.clientset.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	// Convert selector to string
	selector := metav1.FormatLabelSelector(d.Spec.Selector)

	// Get pod metrics for all pods in deployment
	podMetrics, err := s.metricsClient.MetricsV1beta1().PodMetricses(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod metrics: %w", err)
	}

	// Aggregate metrics
	var totalCPU, totalMemory int64
	for _, pod := range podMetrics.Items {
		for _, container := range pod.Containers {
			totalCPU += container.Usage.Cpu().MilliValue()
			totalMemory += container.Usage.Memory().Value()
		}
	}

	metrics := &DeploymentMetrics{
		Name:      name,
		Namespace: namespace,
		CPU:       fmt.Sprintf("%dm", totalCPU),
		Memory:    fmt.Sprintf("%dMi", totalMemory/(1024*1024)),
	}

	return metrics, nil
}

// Scale updates the number of replicas for a deployment
func (s *service) Scale(namespace, name string, replicas int32) error {
	scale, err := s.clientset.AppsV1().Deployments(namespace).GetScale(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment scale: %w", err)
	}

	scale.Spec.Replicas = replicas
	_, err = s.clientset.AppsV1().Deployments(namespace).UpdateScale(context.Background(), name, scale, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update deployment scale: %w", err)
	}

	return nil
}

// Update updates a deployment's configuration
func (s *service) Update(namespace, name string, opts DeploymentOptions) error {
	deployment, err := s.clientset.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	if opts.Replicas != nil {
		deployment.Spec.Replicas = opts.Replicas
	}

	if opts.Image != "" {
		// Update image for all containers
		for i := range deployment.Spec.Template.Spec.Containers {
			deployment.Spec.Template.Spec.Containers[i].Image = opts.Image
		}
	}

	_, err = s.clientset.AppsV1().Deployments(namespace).Update(context.Background(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update deployment: %w", err)
	}

	return nil
}

// AddMetrics adds metrics information to a list of deployments
func (s *service) AddMetrics(deployments []Deployment) error {
	for i := range deployments {
		metrics, err := s.GetMetrics(deployments[i].Namespace, deployments[i].Name)
		if err != nil {
			continue // Skip if metrics are not available
		}
		deployments[i].Metrics = metrics
	}
	return nil
}

// Helper functions

func getDeploymentStatus(d appsv1.Deployment) string {
	if d.Generation <= d.Status.ObservedGeneration {
		if d.Spec.Replicas != nil && d.Status.UpdatedReplicas < *d.Spec.Replicas {
			return "Progressing"
		}
		if d.Status.Replicas > d.Status.UpdatedReplicas {
			return "Progressing"
		}
		if d.Status.AvailableReplicas < d.Status.UpdatedReplicas {
			return "Progressing"
		}
		return "Available"
	}
	return "Progressing"
}

func (s *service) getDeploymentEvents(namespace, name string) ([]pods.Event, error) {
	fieldSelector := fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s,involvedObject.kind=Deployment", name, namespace)
	events, err := s.clientset.CoreV1().Events(namespace).List(context.Background(), metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment events: %w", err)
	}

	var deploymentEvents []pods.Event
	for _, e := range events.Items {
		event := pods.Event{
			Type:      e.Type,
			Reason:    e.Reason,
			Age:       time.Since(e.FirstTimestamp.Time),
			From:      e.Source.Component,
			Message:   e.Message,
			Count:     e.Count,
			FirstSeen: e.FirstTimestamp.Time,
			LastSeen:  e.LastTimestamp.Time,
			Object:    fmt.Sprintf("%s/%s", e.InvolvedObject.Kind, e.InvolvedObject.Name),
		}
		deploymentEvents = append(deploymentEvents, event)
	}

	return deploymentEvents, nil
}
