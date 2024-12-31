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
	Name           string
	Namespace      string
	HasPriority    bool
	ServiceAccount string
	Node           string
	NodeIP         string
	StartTime      time.Time
	Status         string
	Phase          string
	IP             string
	IPs            []string
	ControlledBy   string
	QoSClass       string
	CreationTime   time.Time
	Labels         map[string]string
	Annotations    map[string]string
	NodeSelector   map[string]string

	// Container information
	Containers []ContainerInfo

	// Pod conditions
	Conditions []PodCondition

	// Volume information
	Volumes []VolumeInfo

	// Tolerations
	Tolerations []Toleration

	// Events
	Events []Event
}

// ContainerInfo represents a container in a pod
type ContainerInfo struct {
	Name           string
	ContainerID    string
	Image          string
	ImageID        string
	Ports          []ContainerPort
	State          ContainerState
	Ready          bool
	RestartCount   int32
	Resources      Resources
	VolumeMounts   []VolumeMount
	ReadinessProbe *Probe
	EnvFrom        []EnvFromSource
	Env            []EnvVar
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
	ContainerPort int32
	HostPort      int32
	Protocol      string
}

// Resources represents container resources
type Resources struct {
	Requests Resource
	Limits   Resource
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
	Type    string
	Reason  string
	Age     time.Duration
	From    string
	Message string
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

// ListOptions configures how to list pods
type ListOptions struct {
	Namespace     string
	AllNamespaces bool
	LabelSelector string
	FieldSelector string
}

type ContainerState struct {
	Status   string // Running, Waiting, Terminated
	Started  time.Time
	Reason   string
	ExitCode int32
	Message  string
}

type Probe struct {
	Type             string // tcp-socket, http-get, exec
	Port             int32
	Path             string
	Delay            time.Duration
	Timeout          time.Duration
	Period           time.Duration
	SuccessThreshold int32
	FailureThreshold int32
}

type EnvFromSource struct {
	Name     string
	Type     string // ConfigMap or Secret
	Optional bool
}

type EnvVar struct {
	Name      string
	Value     string
	ValueFrom string
}

type PodCondition struct {
	Type   string
	Status string
}

type VolumeInfo struct {
	Name                   string
	Type                   string
	TokenExpirationSeconds int64
	ConfigMapName          string
	ConfigMapOptional      *bool
	DownwardAPI            bool
}

type Toleration struct {
	Key               string
	Operator          string
	Value             string
	Effect            string
	TolerationSeconds *int64
}
