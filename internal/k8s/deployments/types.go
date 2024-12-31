package deployments

import (
	"time"
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
	Selector          map[string]string
}

// DeploymentDetails contains detailed information about a deployment
type DeploymentDetails struct {
	Name              string
	Namespace         string
	CreationTime      time.Time
	Labels            map[string]string
	Annotations       map[string]string
	Selector          map[string]string
	Replicas          int32
	UpdatedReplicas   int32
	ReadyReplicas     int32
	AvailableReplicas int32
	Strategy          string
	MinReadySeconds   int32

	// Pod template details
	TemplateLabels      map[string]string
	TemplateAnnotations map[string]string

	// Rolling update strategy
	RollingUpdateStrategy *RollingUpdateStrategy

	// Containers
	Containers []ContainerInfo

	// Environment variables from configmaps/secrets
	EnvironmentFrom []EnvironmentFrom
	Environment     []EnvVar

	// Deployment conditions
	Conditions []DeploymentCondition

	// ReplicaSet info
	OldReplicaSets []ReplicaSetInfo
	NewReplicaSet  ReplicaSetInfo

	// Events
	Events []Event
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

type RollingUpdateStrategy struct {
	MaxUnavailable int32
	MaxSurge       int32
}

type EnvironmentFrom struct {
	Name     string
	Type     string // ConfigMap or Secret
	Optional bool
}

type DeploymentCondition struct {
	Type   string
	Status string
	Reason string
}

type Event struct {
	Type    string
	Reason  string
	Age     time.Duration
	From    string
	Message string
}

type ContainerInfo struct {
	Name         string
	Image        string
	Ports        []ContainerPort
	Resources    Resources
	VolumeMounts []VolumeMount
}

type ContainerPort struct {
	Name          string
	ContainerPort int32
	HostPort      int32
	Protocol      string
}

type Resources struct {
	Requests Resource
	Limits   Resource
}

type Resource struct {
	CPU    string
	Memory string
}

type VolumeMount struct {
	Name      string
	MountPath string
	ReadOnly  bool
}

// ReplicaSet info
type ReplicaSetInfo struct {
	Name            string
	ReplicasCreated string // e.g. "0/0" or "1/1"
}

type EnvVar struct {
	Name      string
	Value     string
	ValueFrom string // e.g. "configmap key" or "secret key"
}
