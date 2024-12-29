package pods

import (
	"io"
	"time"
)

// Pod represents a Kubernetes pod with essential information
type Pod struct {
	Name           string
	Namespace      string
	Ready          string
	Status         string
	Restarts       int32
	Age            time.Duration
	IP             string
	Node           string
	Labels         map[string]string
	Controller     string
	ControllerName string
	Metrics        *PodMetrics
	Containers     []ContainerInfo
}

// PodDetails contains detailed information about a pod
type PodDetails struct {
	Name         string
	Namespace    string
	Node         string
	Status       string
	IP           string
	CreationTime time.Time
	Age          time.Duration
	Labels       map[string]string
	NodeSelector map[string]string
	Volumes      []Volume
	Containers   []ContainerInfo
	Events       []Event
}

// ContainerInfo represents a container in a pod
type ContainerInfo struct {
	Name         string
	Image        string
	State        string
	Ready        bool
	RestartCount int32
	Resources    Resources
	VolumeMounts []VolumeMount
	Ports        []ContainerPort
}

// Volume represents a pod volume
type Volume struct {
	Name     string
	Type     string
	Source   string
	ReadOnly bool
}

// VolumeMount represents a container's volume mount
type VolumeMount struct {
	Name      string
	MountPath string
	ReadOnly  bool
}

// ContainerPort represents a container's port configuration
type ContainerPort struct {
	Name          string
	HostPort      int32
	ContainerPort int32
	Protocol      string
}

// Resources represents container resources
type Resources struct {
	Limits   Resource
	Requests Resource
}

// Resource represents CPU or Memory
type Resource struct {
	CPU    string
	Memory string
}

// PodMetrics contains resource usage metrics for a pod
type PodMetrics struct {
	Name       string
	Namespace  string
	Containers []ContainerMetrics
	CPU        string
	Memory     string
}

// ContainerMetrics contains resource usage metrics for a container
type ContainerMetrics struct {
	Name   string
	CPU    string
	Memory string
}

// Event represents a Kubernetes event
type Event struct {
	Type      string
	Reason    string
	Age       time.Duration
	From      string
	Message   string
	Count     int32
	FirstSeen time.Time
	LastSeen  time.Time
	Object    string
}

// LogOptions configures how to retrieve container logs
type LogOptions struct {
	Follow        bool
	Previous      bool
	TailLines     int64
	Writer        io.Writer
	SinceTime     *time.Time
	SinceSeconds  *int64
	Container     string
	AllContainers bool
}

// ExecOptions configures how to execute commands in a container
type ExecOptions struct {
	Command []string
	TTY     bool
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
}
