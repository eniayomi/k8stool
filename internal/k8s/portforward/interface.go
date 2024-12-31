package portforward

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Service defines the interface for port forwarding operations
type Service interface {
	// ForwardPodPort forwards one or more local ports to a pod
	ForwardPodPort(namespace, pod string, options PortForwardOptions) (*PortForwardResult, error)

	// ForwardServicePort forwards one or more local ports to a service
	ForwardServicePort(namespace, service string, options PortForwardOptions) (*PortForwardResult, error)

	// StopForwarding stops an active port forward
	StopForwarding(result *PortForwardResult) error

	// ValidatePortForward validates if port forwarding is possible
	ValidatePortForward(namespace, resource string, ports []PortMapping) error

	// GetForwardedPorts returns a list of currently forwarded ports
	GetForwardedPorts() []ForwardedPort
}

// NewPortForwardService creates a new port forward service instance
func NewPortForwardService(clientset *kubernetes.Clientset, config *rest.Config) (Service, error) {
	if clientset == nil {
		return nil, fmt.Errorf("kubernetes client is required")
	}
	if config == nil {
		return nil, fmt.Errorf("rest config is required")
	}
	return newService(clientset, config), nil
}
