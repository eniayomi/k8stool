package pods

// Service defines the interface for pod operations
type Service interface {
	// List returns a list of pods based on the given filters
	List(namespace string, allNamespaces bool, selector string, statusFilter string) ([]Pod, error)

	// Get returns a specific pod by name
	Get(namespace, name string) (*Pod, error)

	// GetLogs retrieves logs from a pod's container
	GetLogs(namespace, name string, container string, opts LogOptions) error

	// Describe returns detailed information about a pod
	Describe(namespace, name string) (*PodDetails, error)

	// GetMetrics returns resource usage metrics for a pod
	GetMetrics(namespace, name string) (*PodMetrics, error)

	// GetEvents returns events related to a pod
	GetEvents(namespace, name string) ([]Event, error)

	// Exec executes a command in a pod's container
	Exec(namespace, name, container string, opts ExecOptions) error

	// AddMetrics adds metrics information to a list of pods
	AddMetrics(pods []Pod) error
}

// NewService creates a new pod service instance
func NewService(client interface{}) (Service, error) {
	// Implementation will be added in service.go
	return nil, nil
}
