package describe

import (
	"time"

	corev1 "k8s.io/api/core/v1"
)

// ResourceType represents the type of Kubernetes resource
type ResourceType string

const (
	// Pod resource type
	Pod ResourceType = "pod"
	// Deployment resource type
	Deployment ResourceType = "deployment"
	// Service resource type
	Service ResourceType = "service"
	// Node resource type
	Node ResourceType = "node"
	// Namespace resource type
	Namespace ResourceType = "namespace"
)

// ResourceDescription contains detailed information about a Kubernetes resource
type ResourceDescription struct {
	// Type is the resource type
	Type ResourceType `json:"type"`

	// Name is the resource name
	Name string `json:"name"`

	// Namespace is the resource namespace
	Namespace string `json:"namespace,omitempty"`

	// CreationTimestamp is when the resource was created
	CreationTimestamp time.Time `json:"creationTimestamp"`

	// Labels are the resource labels
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations are the resource annotations
	Annotations map[string]string `json:"annotations,omitempty"`

	// Status is the current status
	Status string `json:"status"`

	// Events are recent events related to the resource
	Events []Event `json:"events,omitempty"`

	// Details contains resource-specific details
	Details interface{} `json:"details"`
}

// Event represents a Kubernetes event
type Event struct {
	// Type is the event type (Normal, Warning)
	Type string `json:"type"`

	// Reason is the reason for the event
	Reason string `json:"reason"`

	// Message is the event message
	Message string `json:"message"`

	// Count is how many times this event has occurred
	Count int32 `json:"count"`

	// FirstTimestamp is when the event was first recorded
	FirstTimestamp time.Time `json:"firstTimestamp"`

	// LastTimestamp is when the event was last recorded
	LastTimestamp time.Time `json:"lastTimestamp"`

	// Source is what reported this event
	Source string `json:"source"`
}

// PodDetails contains pod-specific details
type PodDetails struct {
	// Phase is the current phase of the pod
	Phase corev1.PodPhase `json:"phase"`

	// Conditions are the current pod conditions
	Conditions []corev1.PodCondition `json:"conditions"`

	// Node is the name of the node running the pod
	Node string `json:"node"`

	// IP is the pod's IP address
	IP string `json:"ip"`

	// Containers are the pod's containers
	Containers []ContainerDetails `json:"containers"`

	// Volumes are the pod's volumes
	Volumes []VolumeDetails `json:"volumes"`
}

// ContainerDetails contains container-specific details
type ContainerDetails struct {
	// Name is the container name
	Name string `json:"name"`

	// Image is the container image
	Image string `json:"image"`

	// State is the current container state
	State string `json:"state"`

	// Ready indicates if the container is ready
	Ready bool `json:"ready"`

	// RestartCount is how many times the container has restarted
	RestartCount int32 `json:"restartCount"`

	// Ports are the container ports
	Ports []ContainerPort `json:"ports,omitempty"`

	// Resources are the container resource requests/limits
	Resources ResourceRequirements `json:"resources"`
}

// ContainerPort represents a network port in a container
type ContainerPort struct {
	// Name is the port name
	Name string `json:"name,omitempty"`

	// Protocol is the port protocol
	Protocol string `json:"protocol"`

	// ContainerPort is the port number
	ContainerPort int32 `json:"containerPort"`

	// HostPort is the host port number
	HostPort int32 `json:"hostPort,omitempty"`
}

// ResourceRequirements represents compute resource requirements
type ResourceRequirements struct {
	// Requests are the minimum required resources
	Requests ResourceList `json:"requests,omitempty"`

	// Limits are the maximum allowed resources
	Limits ResourceList `json:"limits,omitempty"`
}

// ResourceList represents resource quantities
type ResourceList map[string]string

// VolumeDetails contains volume-specific details
type VolumeDetails struct {
	// Name is the volume name
	Name string `json:"name"`

	// Type is the volume type
	Type string `json:"type"`

	// Source is the volume source
	Source string `json:"source"`

	// MountPath is where the volume is mounted
	MountPath string `json:"mountPath,omitempty"`
}
