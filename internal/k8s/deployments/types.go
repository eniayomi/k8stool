package deployments

import (
	"time"

	"k8stool/internal/k8s/pods"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deployment represents a Kubernetes deployment with essential information
type Deployment struct {
	Name              string
	Namespace         string
	Replicas          int32
	ReadyReplicas     int32
	UpdatedReplicas   int32
	AvailableReplicas int32
	Age               time.Duration
	Status            string
	Metrics           *DeploymentMetrics
	Selector          *metav1.LabelSelector
}

// DeploymentDetails contains detailed information about a deployment
type DeploymentDetails struct {
	Name              string
	Namespace         string
	Replicas          int32
	UpdatedReplicas   int32
	ReadyReplicas     int32
	AvailableReplicas int32
	Strategy          string
	MinReadySeconds   int32
	Age               time.Duration
	Labels            map[string]string
	Selector          map[string]string
	Containers        []pods.ContainerInfo
	Events            []pods.Event
}

// DeploymentMetrics contains resource usage metrics for a deployment
type DeploymentMetrics struct {
	Name      string
	Namespace string
	CPU       string
	Memory    string
}

// DeploymentOptions configures deployment operations
type DeploymentOptions struct {
	Replicas *int32
	Image    string
}
