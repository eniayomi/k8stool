package k8s

import (
	"context"
	"fmt"
	ctx "k8stool/internal/k8s/context"
	"k8stool/internal/k8s/deployments"
	desc "k8stool/internal/k8s/describe"
	"k8stool/internal/k8s/events"
	ex "k8stool/internal/k8s/exec"
	"k8stool/internal/k8s/logs"
	"k8stool/internal/k8s/metrics"
	ns "k8stool/internal/k8s/namespace"
	"k8stool/internal/k8s/pods"
	pf "k8stool/internal/k8s/portforward"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned"
)

// Type aliases for pods package
type Pod = pods.Pod
type PodDetails = pods.PodDetails
type ContainerInfo = pods.ContainerInfo
type ContainerPort = pods.ContainerPort
type Volume = pods.Volume
type VolumeMount = pods.VolumeMount
type Event = pods.Event
type ListOptions = pods.ListOptions

// Type aliases for deployments package
type Deployment = deployments.Deployment
type DeploymentDetails = deployments.DeploymentDetails
type DeploymentMetrics = deployments.DeploymentMetrics
type DeploymentOptions = deployments.DeploymentOptions

// Type aliases for events package
type EventType = events.EventType
type EventList = events.EventList
type EventFilter = events.EventFilter
type EventSortOption = events.EventSortOption
type EventOptions = events.EventOptions

// Type aliases for namespace package
type Namespace = ns.Namespace
type NamespaceDetails = ns.NamespaceDetails
type ResourceQuota = ns.ResourceQuota
type LimitRange = ns.LimitRange
type ResourceList = ns.ResourceList
type NamespaceSortOption = ns.NamespaceSortOption

// Type aliases for metrics package
type ResourceMetrics = metrics.ResourceMetrics
type CPUMetrics = metrics.CPUMetrics
type MemoryMetrics = metrics.MemoryMetrics
type PodMetrics = metrics.PodMetrics
type NodeMetrics = metrics.NodeMetrics
type MetricsSortOption = metrics.MetricsSortOption

// Type aliases for context package
type Context = ctx.Context
type ClusterInfo = ctx.ClusterInfo
type ContextSortOption = ctx.ContextSortOption

// Type aliases for logs package
type LogOptions = logs.LogOptions
type LogResult = logs.LogResult
type LogConnection = logs.LogConnection

// Type aliases for exec package
type ExecOptions = ex.ExecOptions
type ExecResult = ex.ExecResult
type ExecConnection = ex.ExecConnection
type IOStreams = ex.IOStreams
type TerminalSize = ex.TerminalSize
type TerminalSizeQueue = ex.TerminalSizeQueue

// Type aliases for portforward package
type PortForwardOptions = pf.PortForwardOptions
type PortMapping = pf.PortMapping
type Streams = pf.Streams
type ForwardedPort = pf.ForwardedPort
type PortForwardResult = pf.PortForwardResult
type PortForwardDirection = pf.PortForwardDirection
type PortForwardProtocol = pf.PortForwardProtocol

// Type aliases for describe package
type ResourceType = desc.ResourceType
type ResourceDescription = desc.ResourceDescription
type ContainerDetails = desc.ContainerDetails
type VolumeDetails = desc.VolumeDetails
type ResourceRequirements = desc.ResourceRequirements

type Client struct {
	clientset          *kubernetes.Clientset
	metricsClient      *metricsv1beta1.Clientset
	config             *rest.Config
	configFile         clientcmd.ClientConfig
	namespace          string
	PodService         pods.Service
	DeploymentService  deployments.Service
	EventService       events.EventService
	NamespaceService   ns.Service
	MetricsService     metrics.Service
	ContextService     ctx.Service
	LogService         logs.LogService
	ExecService        ex.ExecService
	PortForwardService pf.Service
	DescribeSvc        desc.DescribeService
}

func NewClient() (*Client, error) {
	// Load kubeconfig
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	// Get config
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubernetes config: %w", err)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Create metrics client
	metricsClient, err := metricsv1beta1.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client: %w", err)
	}

	// Get namespace from context
	namespace, _, err := kubeConfig.Namespace()
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace from context: %w", err)
	}

	client := &Client{
		clientset:     clientset,
		metricsClient: metricsClient,
		config:        config,
		configFile:    kubeConfig,
		namespace:     namespace,
	}

	// Initialize pod service
	podService := pods.NewPodService(clientset, metricsClient, config)
	client.PodService = podService

	// Initialize deployment service
	deploymentService, err := deployments.NewDeploymentService(clientset, metricsClient, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment service: %w", err)
	}
	client.DeploymentService = deploymentService

	// Initialize event service
	eventService, err := events.NewEventService(clientset)
	if err != nil {
		return nil, fmt.Errorf("failed to create event service: %w", err)
	}
	client.EventService = eventService

	// Initialize namespace service
	namespaceService, err := ns.NewNamespaceService(clientset, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create namespace service: %w", err)
	}
	client.NamespaceService = namespaceService

	// Initialize metrics service
	metricsService, err := metrics.NewMetricsService(clientset, metricsClient, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics service: %w", err)
	}
	client.MetricsService = metricsService

	// Initialize context service
	contextService, err := ctx.NewContextService(clientset, config, kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create context service: %w", err)
	}
	client.ContextService = contextService

	// Initialize log service
	logService, err := logs.NewLogService(clientset, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create log service: %w", err)
	}
	client.LogService = logService

	// Initialize exec service
	execService, err := ex.NewExecService(clientset, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec service: %w", err)
	}
	client.ExecService = execService

	// Initialize portforward service
	portForwardService, err := pf.NewPortForwardService(clientset, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create portforward service: %w", err)
	}
	client.PortForwardService = portForwardService

	// Initialize describe service
	describeService, err := desc.NewDescribeService(clientset)
	if err != nil {
		return nil, fmt.Errorf("failed to create describe service: %w", err)
	}
	client.DescribeSvc = describeService

	return client, nil
}

func (c *Client) DescribePod(namespace, name string) (*PodDetails, error) {
	return c.PodService.Describe(namespace, name)
}

func (c *Client) GetPodMetrics(namespace, podName string) (*PodMetrics, error) {
	return c.MetricsService.GetPodMetrics(namespace, podName)
}

func (c *Client) AddPodMetrics(pods []Pod) error {
	return c.PodService.AddMetrics(pods)
}

func (c *Client) GetPodLogs(namespace, name string, container string, opts logs.LogOptions) error {
	// Set container if provided
	if container != "" {
		opts.Container = container
	}

	// Get logs
	result, err := c.LogService.GetLogs(context.Background(), namespace, name, &opts)
	if err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}

	// Check for error in result
	if result.Error != "" {
		return fmt.Errorf(result.Error)
	}

	// Write logs to the provided writer or stdout
	if opts.Writer != nil {
		// Logs were already written to the writer in GetLogs
		return nil
	} else if result.Logs != "" {
		fmt.Print(result.Logs)
	}

	return nil
}

func (c *Client) ExecInPod(namespace, podName, containerName string, opts ExecOptions) error {
	result, err := c.ExecService.Exec(context.Background(), namespace, podName, &opts)
	if err != nil {
		return err
	}
	if result.Error != "" {
		return fmt.Errorf(result.Error)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("command exited with code %d", result.ExitCode)
	}
	return nil
}

// Deployment methods
func (c *Client) ListDeployments(namespace string, allNamespaces bool, selector string) ([]Deployment, error) {
	return c.DeploymentService.List(namespace, allNamespaces, selector)
}

func (c *Client) GetDeployment(namespace, name string) (*Deployment, error) {
	return c.DeploymentService.Get(namespace, name)
}

func (c *Client) DescribeDeployment(namespace, name string) (*DeploymentDetails, error) {
	return c.DeploymentService.Describe(namespace, name)
}

func (c *Client) GetDeploymentMetrics(namespace, name string) (*DeploymentMetrics, error) {
	return c.DeploymentService.GetMetrics(namespace, name)
}

func (c *Client) ScaleDeployment(namespace, name string, replicas int32) error {
	return c.DeploymentService.Scale(namespace, name, replicas)
}

func (c *Client) UpdateDeployment(namespace, name string, opts DeploymentOptions) error {
	return c.DeploymentService.Update(namespace, name, opts)
}

func (c *Client) AddDeploymentMetrics(deployments []Deployment) error {
	return c.DeploymentService.AddMetrics(deployments)
}

// Event methods
func (c *Client) ListEvents(ctx context.Context, namespace string, filter *EventFilter) (*EventList, error) {
	return c.EventService.List(ctx, namespace, filter)
}

func (c *Client) ListEventsForObject(ctx context.Context, namespace, kind, name string) (*EventList, error) {
	return c.EventService.ListForObject(ctx, namespace, kind, name)
}

func (c *Client) WatchEvents(ctx context.Context, namespace string, opts *EventOptions) (<-chan events.Event, error) {
	return c.EventService.Watch(ctx, namespace, opts)
}

func (c *Client) GetEvent(ctx context.Context, namespace, name string) (*events.Event, error) {
	return c.EventService.Get(ctx, namespace, name)
}

// Namespace methods
func (c *Client) ListNamespaces() ([]Namespace, error) {
	return c.NamespaceService.List()
}

func (c *Client) GetNamespace(name string) (*NamespaceDetails, error) {
	return c.NamespaceService.Get(name)
}

func (c *Client) CreateNamespace(name string, labels, annotations map[string]string) error {
	return c.NamespaceService.Create(name, labels, annotations)
}

func (c *Client) DeleteNamespace(name string) error {
	return c.NamespaceService.Delete(name)
}

func (c *Client) GetNamespaceResourceQuotas(namespace string) ([]ResourceQuota, error) {
	return c.NamespaceService.GetResourceQuotas(namespace)
}

func (c *Client) GetNamespaceLimitRanges(namespace string) ([]LimitRange, error) {
	return c.NamespaceService.GetLimitRanges(namespace)
}

func (c *Client) SortNamespaces(namespaces []Namespace, sortBy NamespaceSortOption) []Namespace {
	return c.NamespaceService.Sort(namespaces, sortBy)
}

// Metrics methods
func (c *Client) ListPodMetrics(namespace string) ([]PodMetrics, error) {
	return c.MetricsService.ListPodMetrics(namespace)
}

func (c *Client) GetNodeMetrics(name string) (*NodeMetrics, error) {
	return c.MetricsService.GetNodeMetrics(name)
}

func (c *Client) ListNodeMetrics() ([]NodeMetrics, error) {
	return c.MetricsService.ListNodeMetrics()
}

func (c *Client) SortMetrics(podMetrics []PodMetrics, sortBy MetricsSortOption) []PodMetrics {
	return c.MetricsService.Sort(podMetrics, sortBy)
}

// Context methods
func (c *Client) ListContexts() ([]Context, error) {
	return c.ContextService.List()
}

func (c *Client) GetCurrentContext() (*Context, error) {
	return c.ContextService.GetCurrent()
}

func (c *Client) SwitchContext(name string) error {
	return c.ContextService.SwitchContext(name)
}

func (c *Client) SetNamespace(namespace string) error {
	return c.ContextService.SetNamespace(namespace)
}

func (c *Client) GetClusterInfo() (*ClusterInfo, error) {
	return c.ContextService.GetClusterInfo()
}

func (c *Client) SortContexts(contexts []Context, sortBy ContextSortOption) []Context {
	return c.ContextService.Sort(contexts, sortBy)
}

func (c *Client) GetContexts() ([]Context, error) {
	return c.ContextService.List()
}

// Log methods
func (c *Client) GetLogs(ctx context.Context, namespace, pod string, opts *logs.LogOptions) (*logs.LogResult, error) {
	return c.LogService.GetLogs(ctx, namespace, pod, opts)
}

func (c *Client) StreamLogs(ctx context.Context, namespace, pod string, opts *logs.LogOptions) (*logs.LogConnection, error) {
	return c.LogService.StreamLogs(ctx, namespace, pod, opts)
}

func (c *Client) ValidateLogOptions(opts *logs.LogOptions) error {
	return c.LogService.Validate(opts)
}

// Exec methods
func (c *Client) Exec(ctx context.Context, namespace, pod string, opts *ex.ExecOptions) (*ex.ExecResult, error) {
	return c.ExecService.Exec(ctx, namespace, pod, opts)
}

func (c *Client) StreamExec(ctx context.Context, namespace, pod string, opts *ex.ExecOptions) (*ex.ExecConnection, error) {
	return c.ExecService.Stream(ctx, namespace, pod, opts)
}

func (c *Client) ValidateExecOptions(opts *ex.ExecOptions) error {
	return c.ExecService.Validate(opts)
}

// PortForward methods
func (c *Client) ForwardPodPort(namespace, pod string, options PortForwardOptions) (*PortForwardResult, error) {
	return c.PortForwardService.ForwardPodPort(namespace, pod, options)
}

func (c *Client) ForwardServicePort(namespace, service string, options PortForwardOptions) (*PortForwardResult, error) {
	return c.PortForwardService.ForwardServicePort(namespace, service, options)
}

func (c *Client) StopForwarding(result *PortForwardResult) error {
	return c.PortForwardService.StopForwarding(result)
}

func (c *Client) ValidatePortForward(namespace, resource string, ports []PortMapping) error {
	return c.PortForwardService.ValidatePortForward(namespace, resource, ports)
}

func (c *Client) GetForwardedPorts() []ForwardedPort {
	return c.PortForwardService.GetForwardedPorts()
}

// Describe methods
func (c *Client) DescribeResource(ctx context.Context, resourceType ResourceType, namespace, name string) (*ResourceDescription, error) {
	return c.DescribeSvc.Describe(ctx, resourceType, namespace, name)
}

func (c *Client) DescribeService(ctx context.Context, namespace, name string) (*ResourceDescription, error) {
	return c.DescribeSvc.DescribeService(ctx, namespace, name)
}

func (c *Client) DescribeNode(ctx context.Context, name string) (*ResourceDescription, error) {
	return c.DescribeSvc.DescribeNode(ctx, name)
}

func (c *Client) DescribeNamespace(ctx context.Context, name string) (*ResourceDescription, error) {
	return c.DescribeSvc.DescribeNamespace(ctx, name)
}

// ListPods returns a list of pods based on the given options
func (c *Client) ListPods(opts *ListOptions) ([]Pod, error) {
	return c.PodService.List(opts.Namespace, opts.AllNamespaces, opts.LabelSelector, "")
}

// GetDeploymentLogs retrieves logs from all pods in a deployment
func (c *Client) GetDeploymentLogs(namespace, name string, opts LogOptions) error {
	// Get deployment
	deployment, err := c.DeploymentService.Get(namespace, name)
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	// Convert selector map to string
	var selectorStr string
	for k, v := range deployment.Selector {
		if selectorStr != "" {
			selectorStr += ","
		}
		selectorStr += fmt.Sprintf("%s=%s", k, v)
	}

	// Get pods for deployment
	pods, err := c.PodService.List(namespace, false, selectorStr, "")
	if err != nil {
		return fmt.Errorf("failed to get pods for deployment: %w", err)
	}

	if len(pods) == 0 {
		return fmt.Errorf("no pods found for deployment %s", name)
	}

	// Get logs from each pod
	for _, pod := range pods {
		// If container is specified, only get logs for that container
		if opts.Container != "" {
			err = c.GetPodLogs(namespace, pod.Name, opts.Container, opts)
			if err != nil {
				return fmt.Errorf("failed to get logs for pod %s: %w", pod.Name, err)
			}
			continue
		}

		// If all containers requested, get logs for each container
		if opts.AllContainers {
			for _, container := range pod.Containers {
				err = c.GetPodLogs(namespace, pod.Name, container.Name, opts)
				if err != nil {
					return fmt.Errorf("failed to get logs for container %s in pod %s: %w", container.Name, pod.Name, err)
				}
			}
			continue
		}

		// Otherwise, get logs from the first container
		if len(pod.Containers) > 0 {
			err = c.GetPodLogs(namespace, pod.Name, pod.Containers[0].Name, opts)
			if err != nil {
				return fmt.Errorf("failed to get logs for pod %s: %w", pod.Name, err)
			}
		}
	}

	return nil
}

// GetCurrentNamespace returns the current namespace
func (c *Client) GetCurrentNamespace() string {
	return c.namespace
}

// ... existing code ...
