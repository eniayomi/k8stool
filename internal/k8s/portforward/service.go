package portforward

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type service struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
	forwards  map[string]*portforward.PortForwarder
	mu        sync.Mutex
}

// newService creates a new port forward service instance
func newService(clientset *kubernetes.Clientset, config *rest.Config) Service {
	return &service{
		clientset: clientset,
		config:    config,
		forwards:  make(map[string]*portforward.PortForwarder),
	}
}

// ForwardPodPort forwards one or more local ports to a pod
func (s *service) ForwardPodPort(namespace, pod string, options PortForwardOptions) (*PortForwardResult, error) {
	if err := s.ValidatePortForward(namespace, pod, options.Ports); err != nil {
		return nil, err
	}

	req := s.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(pod).
		SubResource("portforward")

	return s.forwardPorts(req.URL(), options)
}

// ForwardServicePort forwards one or more local ports to a service
func (s *service) ForwardServicePort(namespace, service string, options PortForwardOptions) (*PortForwardResult, error) {
	svc, err := s.clientset.CoreV1().Services(namespace).Get(context.Background(), service, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	// Get pods for the service
	var selectors []string
	for k, v := range svc.Spec.Selector {
		selectors = append(selectors, fmt.Sprintf("%s=%s", k, v))
	}
	labelSelector := strings.Join(selectors, ",")

	pods, err := s.clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pods found for service %s", service)
	}

	// Forward to the first available pod
	pod := pods.Items[0]
	return s.ForwardPodPort(namespace, pod.Name, options)
}

// StopForwarding stops an active port forward
func (s *service) StopForwarding(result *PortForwardResult) error {
	if result == nil {
		return fmt.Errorf("port forward result is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, port := range result.Ports {
		key := fmt.Sprintf("%s:%d", port.Address, port.Local)
		if forwarder, exists := s.forwards[key]; exists {
			forwarder.Close()
			delete(s.forwards, key)
		}
		if port.Listener != nil {
			port.Listener.Close()
		}
	}

	return nil
}

// ValidatePortForward validates if port forwarding is possible
func (s *service) ValidatePortForward(namespace, resource string, ports []PortMapping) error {
	if len(ports) == 0 {
		return fmt.Errorf("at least one port mapping is required")
	}

	for _, port := range ports {
		if port.Local == 0 {
			return fmt.Errorf("local port is required")
		}
		if port.Remote == 0 {
			return fmt.Errorf("remote port is required")
		}

		// Check if local port is available
		if port.Address == "" {
			port.Address = "localhost"
		}
		listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", port.Address, port.Local))
		if err != nil {
			return fmt.Errorf("local port %d is not available: %w", port.Local, err)
		}
		listener.Close()
	}

	return nil
}

// GetForwardedPorts returns a list of currently forwarded ports
func (s *service) GetForwardedPorts() []ForwardedPort {
	s.mu.Lock()
	defer s.mu.Unlock()

	var ports []ForwardedPort
	for _, forwarder := range s.forwards {
		fwdPorts, err := forwarder.GetPorts()
		if err != nil {
			continue
		}
		for _, port := range fwdPorts {
			ports = append(ports, ForwardedPort{
				Local:  uint16(port.Local),
				Remote: uint16(port.Remote),
			})
		}
	}

	return ports
}

// Helper functions

func (s *service) forwardPorts(reqURL *url.URL, options PortForwardOptions) (*PortForwardResult, error) {
	transport, upgrader, err := spdy.RoundTripperFor(s.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create round tripper: %w", err)
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", reqURL)

	var ports []string
	for _, mapping := range options.Ports {
		ports = append(ports, fmt.Sprintf("%d:%d", mapping.Local, mapping.Remote))
	}

	if options.StopChannel == nil {
		options.StopChannel = make(chan struct{})
	}
	if options.ReadyChannel == nil {
		options.ReadyChannel = make(chan struct{})
	}

	fw, err := portforward.New(dialer, ports, options.StopChannel, options.ReadyChannel, options.Streams.Out, options.Streams.ErrOut)
	if err != nil {
		return nil, fmt.Errorf("failed to create port forwarder: %w", err)
	}

	var forwardedPorts []ForwardedPort
	for _, mapping := range options.Ports {
		key := fmt.Sprintf("%s:%d", mapping.Address, mapping.Local)
		s.mu.Lock()
		s.forwards[key] = fw
		s.mu.Unlock()

		forwardedPorts = append(forwardedPorts, ForwardedPort{
			Local:    mapping.Local,
			Remote:   mapping.Remote,
			Address:  mapping.Address,
			Protocol: mapping.Protocol,
		})
	}

	go func() {
		err := fw.ForwardPorts()
		if err != nil {
			fmt.Printf("port forwarding failed: %v\n", err)
		}
	}()

	<-options.ReadyChannel

	return &PortForwardResult{
		Ports: forwardedPorts,
	}, nil
}
