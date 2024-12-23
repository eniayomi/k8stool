package k8s

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"k8stool/pkg/utils"
	"os"
	"strings"
	"sync"
	"time"

	"net/http"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/transport/spdy"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Client struct {
	clientset     *kubernetes.Clientset
	metricsClient *metricsv1beta1.Clientset
	config        *rest.Config
	configFile    clientcmd.ClientConfig
	namespace     string
}

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

type Deployment struct {
	Name              string
	Namespace         string
	Replicas          int32
	ReadyReplicas     int32
	UpdatedReplicas   int32
	AvailableReplicas int32
	Age               time.Duration
	Status            string
}

type Context struct {
	Name    string
	Cluster string
}

type Namespace struct {
	Name   string
	Status string
}

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

type ContainerDetails struct {
	Name         string
	Image        string
	State        string
	Ready        bool
	RestartCount int32
	LastState    string
	Ports        []string
	Mounts       []MountDetails
	Resources    ResourceDetails
}

type EventDetails struct {
	Time    string
	Type    string
	Message string
}

type VolumeDetails struct {
	Name     string
	Type     string
	Source   string
	ReadOnly bool
}

type MountDetails struct {
	Name      string
	MountPath string
	ReadOnly  bool
}

type ResourceDetails struct {
	Requests ResourceQuantity
	Limits   ResourceQuantity
}

type ResourceQuantity struct {
	CPU    string
	Memory string
}

type PodMetrics struct {
	Name       string
	Namespace  string
	Containers []ContainerMetrics
	CPU        string
	Memory     string
}

type ContainerMetrics struct {
	Name   string
	CPU    string
	Memory string
}

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

type Container struct {
	Name string
}

type ExecOptions struct {
	Command []string
	TTY     bool
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
}

// Resource represents CPU or Memory
type Resource struct {
	CPU    string
	Memory string
}

// Resources represents container resources
type Resources struct {
	Limits   Resource
	Requests Resource
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

type PortMapping struct {
	Local  string
	Remote string
}

type ContainerPort struct {
	Name          string
	HostPort      int32
	ContainerPort int32
	Protocol      string
}

type Details struct {
	Events []Event
}

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
	Containers        []ContainerInfo
}

func NewClient() (*Client, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	metricsClient, err := metricsv1beta1.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client: %v", err)
	}

	namespace, _, err := kubeConfig.Namespace()
	if err != nil {
		namespace = "default"
	}

	return &Client{
		clientset:     clientset,
		metricsClient: metricsClient,
		config:        config,
		configFile:    kubeConfig,
		namespace:     namespace,
	}, nil
}

func (c *Client) ListPods(namespace string, allNamespaces bool, selector string, statusFilter string) ([]Pod, error) {
	var pods *corev1.PodList
	var err error

	if allNamespaces {
		pods, err = c.clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
			LabelSelector: selector,
		})
	} else {
		if namespace == "" {
			namespace = c.namespace
		}
		pods, err = c.clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: selector,
		})
	}

	if err != nil {
		return nil, err
	}

	result := make([]Pod, 0, len(pods.Items))
	for _, pod := range pods.Items {
		// Apply status filter if provided
		if statusFilter != "" && !strings.EqualFold(string(pod.Status.Phase), statusFilter) {
			continue
		}

		status := string(pod.Status.Phase)
		// Add color to pod status
		switch pod.Status.Phase {
		case corev1.PodRunning:
			status = utils.Green(status)
		case corev1.PodPending:
			status = utils.Yellow(status)
		case corev1.PodFailed:
			status = utils.Red(status)
		case corev1.PodSucceeded:
			status = utils.Blue(status)
		}

		var controllerName, controllerKind string
		for _, owner := range pod.OwnerReferences {
			controllerName = owner.Name
			controllerKind = owner.Kind
			break
		}

		// Add container information
		containers := make([]ContainerInfo, 0, len(pod.Spec.Containers))
		for _, container := range pod.Spec.Containers {
			ports := make([]ContainerPort, 0, len(container.Ports))
			for _, port := range container.Ports {
				ports = append(ports, ContainerPort{
					Name:          port.Name,
					ContainerPort: port.ContainerPort,
					Protocol:      string(port.Protocol),
				})
			}
			containers = append(containers, ContainerInfo{
				Name:  container.Name,
				Ports: ports,
			})
		}

		result = append(result, Pod{
			Name:           pod.Name,
			Namespace:      pod.Namespace,
			Ready:          getPodReady(pod.Status),
			Status:         status,
			Restarts:       getPodRestarts(pod.Status),
			Age:            time.Since(pod.CreationTimestamp.Time),
			IP:             pod.Status.PodIP,
			Node:           pod.Spec.NodeName,
			Labels:         pod.Labels,
			Controller:     controllerKind,
			ControllerName: controllerName,
			Containers:     containers,
		})
	}

	return result, nil
}

func (c *Client) GetPodLogs(namespace, name string, container string, opts LogOptions) error {
	if namespace == "" {
		namespace = c.namespace
	}

	podLogOpts := &corev1.PodLogOptions{
		Container:    container,
		Follow:       opts.Follow,
		Previous:     opts.Previous,
		SinceSeconds: opts.SinceSeconds,
	}

	if opts.SinceTime != nil {
		podLogOpts.SinceTime = &metav1.Time{Time: *opts.SinceTime}
	}

	if opts.TailLines >= 0 {
		podLogOpts.TailLines = &opts.TailLines
	}

	req := c.clientset.CoreV1().Pods(namespace).GetLogs(name, podLogOpts)
	stream, err := req.Stream(context.TODO())
	if err != nil {
		return fmt.Errorf("error opening stream: %v", err)
	}
	defer stream.Close()

	_, err = io.Copy(opts.Writer, stream)
	return err
}

func (c *Client) ListDeployments(namespace string, allNamespaces bool, labelSelector string) ([]Deployment, error) {
	var listOptions metav1.ListOptions
	if labelSelector != "" {
		listOptions.LabelSelector = labelSelector
	}

	// Use all namespaces or the provided/stored namespace
	ns := ""
	if !allNamespaces {
		ns = namespace
	}

	deployments, err := c.clientset.AppsV1().Deployments(ns).List(context.TODO(), listOptions)
	if err != nil {
		return nil, err
	}

	result := make([]Deployment, 0, len(deployments.Items))
	for _, d := range deployments.Items {
		deployment := Deployment{
			Name:              d.Name,
			Namespace:         d.Namespace,
			Replicas:          *d.Spec.Replicas,
			ReadyReplicas:     d.Status.ReadyReplicas,
			UpdatedReplicas:   d.Status.UpdatedReplicas,
			AvailableReplicas: d.Status.AvailableReplicas,
			Age:               time.Since(d.CreationTimestamp.Time),
			Status:            getDeploymentStatus(d),
		}
		result = append(result, deployment)
	}

	return result, nil
}

func getDeploymentStatus(d appsv1.Deployment) string {
	if d.Spec.Replicas != nil && *d.Spec.Replicas == 0 {
		return "Scaled to 0"
	}
	if d.Status.ReadyReplicas == 0 {
		return "Not Ready"
	}
	if d.Status.ReadyReplicas == d.Status.Replicas {
		return "Ready"
	}
	return "Partially Ready"
}

func (c *Client) GetContexts() ([]Context, string, error) {
	config, err := clientcmd.NewDefaultPathOptions().GetStartingConfig()
	if err != nil {
		return nil, "", err
	}

	var contexts []Context
	for name, ctx := range config.Contexts {
		contexts = append(contexts, Context{
			Name:    name,
			Cluster: ctx.Cluster,
		})
	}

	return contexts, config.CurrentContext, nil
}

func (c *Client) GetCurrentContext() (string, error) {
	config, err := clientcmd.NewDefaultPathOptions().GetStartingConfig()
	if err != nil {
		return "", err
	}

	return config.CurrentContext, nil
}

func (c *Client) SwitchContext(contextName string) error {
	configAccess := clientcmd.NewDefaultPathOptions()
	config, err := configAccess.GetStartingConfig()
	if err != nil {
		return err
	}

	// Verify the context exists
	if _, exists := config.Contexts[contextName]; !exists {
		return fmt.Errorf("context %q does not exist", contextName)
	}

	// Set the current context
	config.CurrentContext = contextName

	// Save the config
	return clientcmd.ModifyConfig(configAccess, *config, true)
}

func (c *Client) GetNamespaces() ([]Namespace, string, error) {
	namespaces, err := c.clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, "", err
	}

	config, err := clientcmd.NewDefaultPathOptions().GetStartingConfig()
	if err != nil {
		return nil, "", err
	}

	currentContext := config.Contexts[config.CurrentContext]
	currentNamespace := "default"
	if currentContext != nil {
		currentNamespace = currentContext.Namespace
	}

	var nsList []Namespace
	for _, ns := range namespaces.Items {
		nsList = append(nsList, Namespace{
			Name:   ns.Name,
			Status: string(ns.Status.Phase),
		})
	}

	return nsList, currentNamespace, nil
}

func (c *Client) SwitchNamespace(namespace string) error {
	// First, verify the namespace exists
	_, err := c.clientset.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("namespace %s not found: %v", namespace, err)
	}

	// Get current context
	raw, err := c.configFile.RawConfig()
	if err != nil {
		return err
	}

	currentContext := raw.CurrentContext
	if currentContext == "" {
		return fmt.Errorf("no current context found")
	}

	// Update namespace in current context
	if raw.Contexts[currentContext] == nil {
		return fmt.Errorf("current context %s not found in kubeconfig", currentContext)
	}

	raw.Contexts[currentContext].Namespace = namespace

	// Save the updated config
	if err := clientcmd.ModifyConfig(clientcmd.NewDefaultPathOptions(), raw, true); err != nil {
		return fmt.Errorf("failed to modify kubeconfig: %v", err)
	}

	// Update the client's namespace
	c.namespace = namespace

	return nil
}

func (c *Client) DescribePod(namespace, name string) (*PodDetails, error) {
	pod, err := c.clientset.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	containers := make([]ContainerInfo, 0, len(pod.Spec.Containers))
	for _, c := range pod.Spec.Containers {
		var status corev1.ContainerStatus
		for _, s := range pod.Status.ContainerStatuses {
			if s.Name == c.Name {
				status = s
				break
			}
		}

		// Convert container ports
		ports := make([]ContainerPort, 0, len(c.Ports))
		for _, port := range c.Ports {
			ports = append(ports, ContainerPort{
				Name:          port.Name,
				HostPort:      port.HostPort,
				ContainerPort: port.ContainerPort,
				Protocol:      string(port.Protocol),
			})
		}

		// Convert volume mounts
		mounts := make([]VolumeMount, 0, len(c.VolumeMounts))
		for _, m := range c.VolumeMounts {
			mounts = append(mounts, VolumeMount{
				Name:      m.Name,
				MountPath: m.MountPath,
				ReadOnly:  m.ReadOnly,
			})
		}

		container := ContainerInfo{
			Name:         c.Name,
			Image:        c.Image,
			Ready:        status.Ready,
			RestartCount: status.RestartCount,
			Ports:        ports,
			VolumeMounts: mounts,
			Resources: Resources{
				Limits: Resource{
					CPU:    c.Resources.Limits.Cpu().String(),
					Memory: c.Resources.Limits.Memory().String(),
				},
				Requests: Resource{
					CPU:    c.Resources.Requests.Cpu().String(),
					Memory: c.Resources.Requests.Memory().String(),
				},
			},
		}
		containers = append(containers, container)
	}

	// Convert volumes
	volumes := make([]Volume, 0, len(pod.Spec.Volumes))
	for _, v := range pod.Spec.Volumes {
		volume := Volume{
			Name: v.Name,
		}
		// Determine volume type and source
		if v.Secret != nil {
			volume.Type = "Secret"
			volume.Source = v.Secret.SecretName
		} else if v.ConfigMap != nil {
			volume.Type = "ConfigMap"
			volume.Source = v.ConfigMap.Name
		} else if v.PersistentVolumeClaim != nil {
			volume.Type = "PersistentVolumeClaim"
			volume.Source = v.PersistentVolumeClaim.ClaimName
		} else if v.EmptyDir != nil {
			volume.Type = "EmptyDir"
			volume.Source = ""
		} else if v.HostPath != nil {
			volume.Type = "HostPath"
			volume.Source = v.HostPath.Path
		}
		// Add more volume types as needed
		volumes = append(volumes, volume)
	}

	// Set NodeSelector (use empty map if none specified)
	nodeSelector := pod.Spec.NodeSelector
	if nodeSelector == nil {
		nodeSelector = make(map[string]string)
	}
	// details.Events = events

	return &PodDetails{
		Name:         pod.Name,
		Namespace:    pod.Namespace,
		Node:         pod.Spec.NodeName,
		Status:       string(pod.Status.Phase),
		IP:           pod.Status.PodIP,
		Age:          time.Since(pod.CreationTimestamp.Time),
		Labels:       pod.Labels,
		NodeSelector: nodeSelector,
		Volumes:      volumes,
		Containers:   containers,
	}, nil
}

func getContainerState(state corev1.ContainerState) string {
	if state.Running != nil {
		return "Running"
	}
	if state.Waiting != nil {
		return fmt.Sprintf("Waiting (%s)", state.Waiting.Reason)
	}
	if state.Terminated != nil {
		return fmt.Sprintf("Terminated (%s)", state.Terminated.Reason)
	}
	return ""
}

func (c *Client) GetPodMetrics(namespace, podName string) (*PodMetrics, error) {
	metrics, err := c.metricsClient.MetricsV1beta1().PodMetricses(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod metrics: %v", err)
	}

	podMetrics := &PodMetrics{
		Name:      metrics.Name,
		Namespace: metrics.Namespace,
	}

	var totalCPU, totalMemory int64
	for _, container := range metrics.Containers {
		cpu := container.Usage.Cpu().MilliValue()
		memory := container.Usage.Memory().Value() / (1024 * 1024) // Convert to Mi

		totalCPU += cpu
		totalMemory += memory

		podMetrics.Containers = append(podMetrics.Containers, ContainerMetrics{
			Name:   container.Name,
			CPU:    fmt.Sprintf("%dm", cpu),
			Memory: fmt.Sprintf("%dMi", memory),
		})
	}

	podMetrics.CPU = fmt.Sprintf("%dm", totalCPU)
	podMetrics.Memory = fmt.Sprintf("%dMi", totalMemory)

	return podMetrics, nil
}

func (c *Client) GetPodEvents(namespace, podName string) ([]Event, error) {
	events, err := c.clientset.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s", podName),
	})
	if err != nil {
		return nil, err
	}

	result := make([]Event, 0, len(events.Items))
	for _, e := range events.Items {
		event := Event{
			Type:      e.Type,
			Reason:    e.Reason,
			Age:       time.Since(e.CreationTimestamp.Time),
			From:      e.Source.Component,
			Message:   e.Message,
			Count:     e.Count,
			FirstSeen: e.FirstTimestamp.Time,
			LastSeen:  e.LastTimestamp.Time,
			Object:    fmt.Sprintf("%s/%s", strings.ToLower(e.InvolvedObject.Kind), e.InvolvedObject.Name),
		}
		result = append(result, event)
	}

	return result, nil
}

func (c *Client) GetDeploymentEvents(namespace, deploymentName string) ([]Event, error) {
	events, err := c.clientset.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.kind=Deployment,involvedObject.name=%s", deploymentName),
	})
	if err != nil {
		return nil, err
	}

	result := make([]Event, 0, len(events.Items))
	for _, e := range events.Items {
		event := Event{
			Type:      e.Type,
			Reason:    e.Reason,
			Age:       time.Since(e.CreationTimestamp.Time),
			From:      e.Source.Component,
			Message:   e.Message,
			Count:     e.Count,
			FirstSeen: e.FirstTimestamp.Time,
			LastSeen:  e.LastTimestamp.Time,
			Object:    fmt.Sprintf("%s/%s", strings.ToLower(e.InvolvedObject.Kind), e.InvolvedObject.Name),
		}
		result = append(result, event)
	}

	return result, nil
}

func (c *Client) GetDetails(namespace, resourceType, name string) (*Details, error) {
	var fieldSelector string
	switch resourceType {
	case "pod", "po":
		fieldSelector = fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Pod", name)
	case "deployment", "deploy":
		fieldSelector = fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Deployment", name)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	events, err := c.clientset.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, err
	}

	result := make([]Event, 0, len(events.Items))
	for _, e := range events.Items {
		event := Event{
			Type:      e.Type,
			Reason:    e.Reason,
			Age:       time.Since(e.CreationTimestamp.Time),
			From:      e.Source.Component,
			Message:   e.Message,
			Count:     e.Count,
			FirstSeen: e.FirstTimestamp.Time,
			LastSeen:  e.LastTimestamp.Time,
			Object:    fmt.Sprintf("%s/%s", strings.ToLower(e.InvolvedObject.Kind), e.InvolvedObject.Name),
		}
		result = append(result, event)
	}

	return &Details{
		Events: result,
	}, nil
}

func (c *Client) GetPod(namespace, name string) (*Pod, error) {
	pod, err := c.clientset.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	containers := make([]ContainerInfo, 0)
	for _, c := range pod.Spec.Containers {
		containers = append(containers, ContainerInfo{
			Name: c.Name,
		})
	}

	return &Pod{
		Name:       pod.Name,
		Namespace:  pod.Namespace,
		Containers: containers,
	}, nil
}

func (c *Client) ExecInPod(namespace, podName, containerName string, opts ExecOptions) error {
	pod, err := c.clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("pod %s not found: %v", podName, err)
	}

	if pod.Status.Phase != corev1.PodRunning {
		return fmt.Errorf("pod %s is not running (status: %s)", podName, pod.Status.Phase)
	}

	// If container name is not specified, use the first container
	if containerName == "" {
		if len(pod.Spec.Containers) == 0 {
			return fmt.Errorf("no containers found in pod %s", podName)
		}
		containerName = pod.Spec.Containers[0].Name
	}

	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: containerName,
		Command:   opts.Command,
		Stdin:     opts.Stdin != nil,
		Stdout:    true,
		Stderr:    true,
		TTY:       opts.TTY,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("failed to create executor: %v", err)
	}

	streamOptions := remotecommand.StreamOptions{
		Stdin:             opts.Stdin,
		Stdout:            opts.Stdout,
		Stderr:            opts.Stderr,
		Tty:               opts.TTY,
		TerminalSizeQueue: nil,
	}

	return exec.Stream(streamOptions)
}

func (c *Client) DescribeDeployment(namespace, name string) (*DeploymentDetails, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	containers := make([]ContainerInfo, 0, len(deployment.Spec.Template.Spec.Containers))
	for _, c := range deployment.Spec.Template.Spec.Containers {
		ports := make([]ContainerPort, 0, len(c.Ports))
		for _, port := range c.Ports {
			ports = append(ports, ContainerPort{
				Name:          port.Name,
				HostPort:      port.HostPort,
				ContainerPort: port.ContainerPort,
				Protocol:      string(port.Protocol),
			})
		}

		container := ContainerInfo{
			Name:  c.Name,
			Image: c.Image,
			Ports: ports,
			Resources: Resources{
				Limits: Resource{
					CPU:    c.Resources.Limits.Cpu().String(),
					Memory: c.Resources.Limits.Memory().String(),
				},
				Requests: Resource{
					CPU:    c.Resources.Requests.Cpu().String(),
					Memory: c.Resources.Requests.Memory().String(),
				},
			},
		}
		containers = append(containers, container)
	}

	return &DeploymentDetails{
		Name:              deployment.Name,
		Namespace:         deployment.Namespace,
		Replicas:          *deployment.Spec.Replicas,
		UpdatedReplicas:   deployment.Status.UpdatedReplicas,
		ReadyReplicas:     deployment.Status.ReadyReplicas,
		AvailableReplicas: deployment.Status.AvailableReplicas,
		Strategy:          string(deployment.Spec.Strategy.Type),
		MinReadySeconds:   deployment.Spec.MinReadySeconds,
		Age:               time.Since(deployment.CreationTimestamp.Time),
		Labels:            deployment.Labels,
		Selector:          deployment.Spec.Selector.MatchLabels,
		Containers:        containers,
	}, nil
}

func (p PodDetails) Print(w io.Writer, details *Details) error {
	fmt.Fprintf(w, "Name:         %s\n", p.Name)
	fmt.Fprintf(w, "Namespace:    %s\n", p.Namespace)
	fmt.Fprintf(w, "Node:         %s\n", p.Node)
	fmt.Fprintf(w, "Status:       %s\n", p.Status)
	fmt.Fprintf(w, "IP:           %s\n", p.IP)
	fmt.Fprintf(w, "Age:          %s\n", utils.FormatDuration(p.Age))

	if len(p.Labels) > 0 {
		fmt.Fprintf(w, "Labels:\n")
		for k, v := range p.Labels {
			fmt.Fprintf(w, "  %s: %s\n", k, v)
		}
	}

	fmt.Fprintf(w, "Node Selector: ")
	if len(p.NodeSelector) > 0 {
		fmt.Fprintf(w, "\n")
		for k, v := range p.NodeSelector {
			fmt.Fprintf(w, "  %s: %s\n", k, v)
		}
	} else {
		fmt.Fprintf(w, "<none>\n")
	}

	if len(p.Volumes) > 0 {
		fmt.Fprintf(w, "\nVolumes:\n")
		for _, v := range p.Volumes {
			fmt.Fprintf(w, "  %s:\n", v.Name)
			fmt.Fprintf(w, "    Type:     %s\n", v.Type)
			if v.Source != "" {
				fmt.Fprintf(w, "    Source:   %s\n", v.Source)
			}
			if v.ReadOnly {
				fmt.Fprintf(w, "    ReadOnly: true\n")
			}
		}
	}

	fmt.Fprintf(w, "\nContainers:\n")
	for _, c := range p.Containers {
		fmt.Fprintf(w, "  %s:\n", c.Name)
		fmt.Fprintf(w, "    Image:          %s\n", c.Image)
		fmt.Fprintf(w, "    Ready:          %v\n", c.Ready)
		fmt.Fprintf(w, "    Restart Count:  %d\n", c.RestartCount)

		if len(c.Ports) > 0 {
			fmt.Fprintf(w, "    Ports:\n")
			for _, port := range c.Ports {
				if port.Name != "" {
					fmt.Fprintf(w, "      %s: %d/%s", port.Name, port.ContainerPort, port.Protocol)
				} else {
					fmt.Fprintf(w, "      %d/%s", port.ContainerPort, port.Protocol)
				}
				if port.HostPort != 0 {
					fmt.Fprintf(w, " -> %d", port.HostPort)
				}
				fmt.Fprintf(w, "\n")
			}
		}

		if len(c.VolumeMounts) > 0 {
			fmt.Fprintf(w, "    Mounts:\n")
			for _, m := range c.VolumeMounts {
				fmt.Fprintf(w, "      %s -> %s", m.Name, m.MountPath)
				if m.ReadOnly {
					fmt.Fprintf(w, " (ro)")
				}
				fmt.Fprintf(w, "\n")
			}
		}

		fmt.Fprintf(w, "    Resources:\n")
		fmt.Fprintf(w, "      Limits:\n")
		fmt.Fprintf(w, "        CPU:     %s\n", c.Resources.Limits.CPU)
		fmt.Fprintf(w, "        Memory:  %s\n", c.Resources.Limits.Memory)
		fmt.Fprintf(w, "      Requests:\n")
		fmt.Fprintf(w, "        CPU:     %s\n", c.Resources.Requests.CPU)
		fmt.Fprintf(w, "        Memory:  %s\n", c.Resources.Requests.Memory)
	}

	if details != nil && len(details.Events) > 0 {
		fmt.Fprintf(w, "\nEvents:\n")
		fmt.Fprintf(w, "  TYPE\tREASON\tAGE\tFROM\tMESSAGE\n")
		for _, e := range details.Events {
			fmt.Fprintf(w, "  %s\t%s\t%s\t%s\t%s\n",
				e.Type,
				e.Reason,
				utils.FormatDuration(e.Age),
				e.From,
				e.Message,
			)
		}
	}

	return nil
}

func (c *Client) GetDeploymentLogs(namespace, name string, opts LogOptions) error {
	if namespace == "" {
		namespace = c.namespace
	}

	// Get deployment
	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// Get pods for deployment
	selector := deployment.Spec.Selector.MatchLabels
	labelSelector := []string{}
	for key, value := range selector {
		labelSelector = append(labelSelector, fmt.Sprintf("%s=%s", key, value))
	}

	pods, err := c.clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: strings.Join(labelSelector, ","),
	})
	if err != nil {
		return err
	}

	if len(pods.Items) == 0 {
		return fmt.Errorf("no pods found for deployment %s", name)
	}

	// Create a channel for errors
	errChan := make(chan error, len(pods.Items))
	var wg sync.WaitGroup

	// Create a mutex for synchronized writing
	var writeMutex sync.Mutex

	// Get logs from each pod
	for _, pod := range pods.Items {
		wg.Add(1)
		go func(pod corev1.Pod) {
			defer wg.Done()

			// If container is specified, only get logs for that container
			if opts.Container != "" {
				err := c.getPodContainerLogs(&pod, opts.Container, opts, &writeMutex)
				if err != nil {
					errChan <- fmt.Errorf("pod %s container %s: %v", pod.Name, opts.Container, err)
				}
				return
			}

			// If all-containers flag is set, get logs from all containers
			if opts.AllContainers {
				for _, container := range pod.Spec.Containers {
					err := c.getPodContainerLogs(&pod, container.Name, opts, &writeMutex)
					if err != nil {
						errChan <- fmt.Errorf("pod %s container %s: %v", pod.Name, container.Name, err)
					}
				}
				return
			}

			// Default: get logs from first container
			if len(pod.Spec.Containers) > 0 {
				err := c.getPodContainerLogs(&pod, pod.Spec.Containers[0].Name, opts, &writeMutex)
				if err != nil {
					errChan <- fmt.Errorf("pod %s container %s: %v", pod.Name, pod.Spec.Containers[0].Name, err)
				}
			}
		}(pod)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errChan)

	// Check for errors
	var errors []string
	for err := range errChan {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors getting logs:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

func (c *Client) getPodContainerLogs(pod *corev1.Pod, containerName string, opts LogOptions, mutex *sync.Mutex) error {
	podLogOpts := &corev1.PodLogOptions{
		Container:    containerName,
		Follow:       opts.Follow,
		Previous:     opts.Previous,
		SinceSeconds: opts.SinceSeconds,
	}

	// Convert time.Time to metav1.Time
	if opts.SinceTime != nil {
		podLogOpts.SinceTime = &metav1.Time{Time: *opts.SinceTime}
	}

	if opts.TailLines >= 0 {
		podLogOpts.TailLines = &opts.TailLines
	}

	req := c.clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, podLogOpts)
	stream, err := req.Stream(context.TODO())
	if err != nil {
		return fmt.Errorf("error opening stream: %v", err)
	}
	defer stream.Close()

	// Add pod and container name as prefix to each line
	prefix := fmt.Sprintf("[pod/%s/%s] ", pod.Name, containerName)
	reader := bufio.NewReader(stream)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// Synchronized write with prefix
		mutex.Lock()
		fmt.Fprintf(opts.Writer, "%s%s", prefix, line)
		mutex.Unlock()
	}

	return nil
}

func (c *Client) PortForwardPod(namespace, podName, address string, ports []PortMapping, stopChan, readyChan chan struct{}) error {
	if namespace == "" {
		namespace = c.namespace
	}

	pod, err := c.clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if pod.Status.Phase != corev1.PodRunning {
		return fmt.Errorf("pod %s is not running (status: %s)", podName, pod.Status.Phase)
	}

	return c.forwardPorts(namespace, podName, address, ports, stopChan, readyChan)
}

func (c *Client) PortForwardDeployment(namespace, deployName, address string, ports []PortMapping, stopChan, readyChan chan struct{}) error {
	if namespace == "" {
		namespace = c.namespace
	}

	// Get deployment
	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deployName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// Get pods for deployment
	selector := deployment.Spec.Selector.MatchLabels
	labelSelector := []string{}
	for key, value := range selector {
		labelSelector = append(labelSelector, fmt.Sprintf("%s=%s", key, value))
	}

	pods, err := c.clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: strings.Join(labelSelector, ","),
	})
	if err != nil {
		return err
	}

	if len(pods.Items) == 0 {
		return fmt.Errorf("no pods found for deployment %s", deployName)
	}

	// Find first running pod
	var targetPod *corev1.Pod
	for i := range pods.Items {
		if pods.Items[i].Status.Phase == corev1.PodRunning {
			targetPod = &pods.Items[i]
			break
		}
	}

	if targetPod == nil {
		return fmt.Errorf("no running pods found for deployment %s", deployName)
	}

	fmt.Printf("Forwarding ports to pod %s\n", targetPod.Name)
	return c.forwardPorts(namespace, targetPod.Name, address, ports, stopChan, readyChan)
}

func (c *Client) forwardPorts(namespace, podName, address string, ports []PortMapping, stopChan, readyChan chan struct{}) error {
	// Convert port mappings to string slices
	portStrings := make([]string, len(ports))
	for i, port := range ports {
		portStrings[i] = fmt.Sprintf("%s:%s", port.Local, port.Remote)
	}

	// Create URL for port forwarding
	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(podName).
		SubResource("portforward")

	transport, upgrader, err := spdy.RoundTripperFor(c.config)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())

	// Create port forwarder
	fw, err := portforward.NewOnAddresses(dialer, []string{address}, portStrings, stopChan, readyChan, os.Stdout, os.Stderr)
	if err != nil {
		return err
	}

	// Start port forwarding
	return fw.ForwardPorts()
}

func (c *Client) GetCurrentNamespace() (string, error) {
	raw, err := c.configFile.RawConfig()
	if err != nil {
		return "", err
	}

	ctx := raw.CurrentContext
	if ctx == "" {
		return "default", nil
	}

	return raw.Contexts[ctx].Namespace, nil
}

func (c *Client) GetNamespace() string {
	return c.namespace
}

func (c *Client) ListNamespaces() ([]corev1.Namespace, error) {
	namespaces, err := c.clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return namespaces.Items, nil
}

// Helper functions for pod status
func getPodReady(status corev1.PodStatus) string {
	ready := 0
	total := len(status.ContainerStatuses)
	for _, container := range status.ContainerStatuses {
		if container.Ready {
			ready++
		}
	}
	return fmt.Sprintf("%d/%d", ready, total)
}

func getPodRestarts(status corev1.PodStatus) int32 {
	var restarts int32
	for _, container := range status.ContainerStatuses {
		restarts += container.RestartCount
	}
	return restarts
}

// AddPodMetrics adds metrics information to pods
func (c *Client) AddPodMetrics(pods []Pod) error {
	metrics, err := c.metricsClient.MetricsV1beta1().PodMetricses("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	metricsMap := make(map[string]*PodMetrics)
	for _, m := range metrics.Items {
		key := fmt.Sprintf("%s/%s", m.Namespace, m.Name)
		var cpu, mem int64
		for _, c := range m.Containers {
			cpu += c.Usage.Cpu().MilliValue()
			mem += c.Usage.Memory().Value()
		}
		metricsMap[key] = &PodMetrics{
			CPU:    fmt.Sprintf("%dm", cpu),
			Memory: formatBytes(mem),
		}
	}

	for i := range pods {
		key := fmt.Sprintf("%s/%s", pods[i].Namespace, pods[i].Name)
		if metrics, ok := metricsMap[key]; ok {
			pods[i].Metrics = metrics
		}
	}

	return nil
}

// Helper function to format bytes
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1fGi", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1fMi", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1fKi", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

func (c *Client) GetNodeMetrics(selector string) error {
	// Implementation here
	return nil
}

func (c *Client) GetEvents(namespace string, allNamespaces bool, selector string, watch bool) error {
	// Implementation here
	return nil
}

func (d DeploymentDetails) Print(w io.Writer, details *Details) error {
	fmt.Fprintf(w, "Name:               %s\n", d.Name)
	fmt.Fprintf(w, "Namespace:          %s\n", d.Namespace)
	fmt.Fprintf(w, "Strategy:           %s\n", d.Strategy)
	fmt.Fprintf(w, "MinReadySeconds:    %d\n", d.MinReadySeconds)
	fmt.Fprintf(w, "Replicas:           %d\n", d.Replicas)
	fmt.Fprintf(w, "Updated Replicas:   %d\n", d.UpdatedReplicas)
	fmt.Fprintf(w, "Ready Replicas:     %d\n", d.ReadyReplicas)
	fmt.Fprintf(w, "Available Replicas: %d\n", d.AvailableReplicas)
	fmt.Fprintf(w, "Age:                %s\n", utils.FormatDuration(d.Age))

	if len(d.Labels) > 0 {
		fmt.Fprintf(w, "\nLabels:\n")
		for k, v := range d.Labels {
			fmt.Fprintf(w, "  %s: %s\n", k, v)
		}
	}

	if len(d.Selector) > 0 {
		fmt.Fprintf(w, "\nSelector:\n")
		for k, v := range d.Selector {
			fmt.Fprintf(w, "  %s: %s\n", k, v)
		}
	}

	fmt.Fprintf(w, "\nContainers:\n")
	for _, c := range d.Containers {
		fmt.Fprintf(w, "  %s:\n", c.Name)
		fmt.Fprintf(w, "    Image:     %s\n", c.Image)

		if len(c.Ports) > 0 {
			fmt.Fprintf(w, "    Ports:\n")
			for _, port := range c.Ports {
				if port.Name != "" {
					fmt.Fprintf(w, "      %s: %d/%s", port.Name, port.ContainerPort, port.Protocol)
				} else {
					fmt.Fprintf(w, "      %d/%s", port.ContainerPort, port.Protocol)
				}
				if port.HostPort != 0 {
					fmt.Fprintf(w, " -> %d", port.HostPort)
				}
				fmt.Fprintf(w, "\n")
			}
		}

		fmt.Fprintf(w, "    Resources:\n")
		fmt.Fprintf(w, "      Limits:\n")
		fmt.Fprintf(w, "        CPU:     %s\n", c.Resources.Limits.CPU)
		fmt.Fprintf(w, "        Memory:  %s\n", c.Resources.Limits.Memory)
		fmt.Fprintf(w, "      Requests:\n")
		fmt.Fprintf(w, "        CPU:     %s\n", c.Resources.Requests.CPU)
		fmt.Fprintf(w, "        Memory:  %s\n", c.Resources.Requests.Memory)
	}

	if details != nil && len(details.Events) > 0 {
		fmt.Fprintf(w, "\nEvents:\n")
		fmt.Fprintf(w, "  TYPE\tREASON\tAGE\tFROM\tMESSAGE\n")
		for _, e := range details.Events {
			eventType := e.Type
			if e.Type == "Normal" {
				eventType = utils.Green(e.Type)
			} else if e.Type == "Warning" {
				eventType = utils.Yellow(e.Type)
			}

			fmt.Fprintf(w, "  %s\t%s\t%s\t%s\t%s\n",
				eventType,
				e.Reason,
				utils.FormatDuration(e.Age),
				e.From,
				e.Message,
			)
		}
	}

	return nil
}
