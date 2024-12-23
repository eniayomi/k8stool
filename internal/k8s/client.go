package k8s

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Client struct {
	clientset     *kubernetes.Clientset
	metricsClient *metricsv1beta1.Clientset
	config        *rest.Config
}

type Pod struct {
	Name       string
	Namespace  string
	Ready      string
	Status     string
	Restarts   int32
	Age        time.Duration
	Controller string
	Containers []Container
}

type Deployment struct {
	Name      string
	Namespace string
	Ready     string
	UpToDate  int32
	Available int32
	Age       time.Duration
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
	Labels       map[string]string
	NodeSelector map[string]string
	Containers   []ContainerInfo
	Volumes      []Volume
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
	LastSeen time.Duration
	Type     string
	Reason   string
	Object   string
	Message  string
}

type LogOptions struct {
	Follow       bool
	Previous     bool
	TailLines    int64
	Writer       io.Writer
	SinceTime    *time.Time
	SinceSeconds *int64
}

type Container struct {
	Name string
}

type ExecOptions struct {
	Command []string
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	TTY     bool
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
}

// Volume represents a pod volume
type Volume struct {
	Name   string
	Type   string
	Source string
}

// VolumeMount represents a container's volume mount
type VolumeMount struct {
	Name      string
	MountPath string
	ReadOnly  bool
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
		return nil, err
	}

	return &Client{
		clientset:     clientset,
		metricsClient: metricsClient,
		config:        config,
	}, nil
}

func (c *Client) ListPods(namespace, labelSelector string) ([]Pod, error) {
	pods, err := c.clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}

	var podList []Pod
	for _, pod := range pods.Items {
		// Calculate ready containers
		readyContainers := 0
		totalContainers := len(pod.Spec.Containers)
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.Ready {
				readyContainers++
			}
		}

		// Calculate restarts
		var restarts int32
		for _, containerStatus := range pod.Status.ContainerStatuses {
			restarts += containerStatus.RestartCount
		}

		// Calculate age
		age := time.Since(pod.CreationTimestamp.Time)

		// Determine controller type
		controller := "<none>"
		if len(pod.OwnerReferences) > 0 {
			owner := pod.OwnerReferences[0]
			switch owner.Kind {
			case "ReplicaSet":
				// Check if it's part of a Deployment
				if rs, err := c.clientset.AppsV1().ReplicaSets(namespace).Get(context.TODO(), owner.Name, metav1.GetOptions{}); err == nil {
					if len(rs.OwnerReferences) > 0 && rs.OwnerReferences[0].Kind == "Deployment" {
						controller = "Deployment"
					} else {
						controller = "ReplicaSet"
					}
				}
			case "Job":
				// Check if it's part of a CronJob
				if job, err := c.clientset.BatchV1().Jobs(namespace).Get(context.TODO(), owner.Name, metav1.GetOptions{}); err == nil {
					if len(job.OwnerReferences) > 0 && job.OwnerReferences[0].Kind == "CronJob" {
						controller = "CronJob"
					} else {
						controller = "Job"
					}
				}
			default:
				controller = owner.Kind
			}
		}

		podList = append(podList, Pod{
			Name:       pod.Name,
			Namespace:  pod.Namespace,
			Ready:      fmt.Sprintf("%d/%d", readyContainers, totalContainers),
			Status:     string(pod.Status.Phase),
			Restarts:   restarts,
			Age:        age,
			Controller: controller,
		})
	}

	return podList, nil
}

func (c *Client) GetPodLogs(namespace, podName, containerName string, opts LogOptions) error {
	podLogOpts := &corev1.PodLogOptions{
		Container:    containerName,
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

	req := c.clientset.CoreV1().Pods(namespace).GetLogs(podName, podLogOpts)
	stream, err := req.Stream(context.TODO())
	if err != nil {
		return fmt.Errorf("error opening stream: %v", err)
	}
	defer stream.Close()

	_, err = io.Copy(opts.Writer, stream)
	return err
}

func (c *Client) ListDeployments(namespace string) ([]Deployment, error) {
	deployments, err := c.clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var deploymentList []Deployment
	for _, d := range deployments.Items {
		age := time.Since(d.CreationTimestamp.Time)
		ready := fmt.Sprintf("%d/%d", d.Status.ReadyReplicas, d.Status.Replicas)

		deploymentList = append(deploymentList, Deployment{
			Name:      d.Name,
			Namespace: d.Namespace,
			Ready:     ready,
			UpToDate:  d.Status.UpdatedReplicas,
			Available: d.Status.AvailableReplicas,
			Age:       age,
		})
	}

	return deploymentList, nil
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
	// Verify namespace exists
	_, err := c.clientset.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("namespace %q does not exist", namespace)
	}

	configAccess := clientcmd.NewDefaultPathOptions()
	config, err := configAccess.GetStartingConfig()
	if err != nil {
		return err
	}

	context := config.Contexts[config.CurrentContext]
	if context == nil {
		return fmt.Errorf("current context is not set")
	}

	context.Namespace = namespace

	return clientcmd.ModifyConfig(configAccess, *config, true)
}

func formatAge(d time.Duration) string {
	if d.Hours() > 24*365 {
		years := int(d.Hours() / (24 * 365))
		days := int((d.Hours() - float64(years)*24*365) / 24)
		if days > 0 {
			return fmt.Sprintf("%dy%dd", years, days)
		}
		return fmt.Sprintf("%dy", years)
	}
	if d.Hours() > 24*30 {
		months := int(d.Hours() / (24 * 30))
		days := int((d.Hours() - float64(months)*24*30) / 24)
		if days > 0 {
			return fmt.Sprintf("%dM%dd", months, days)
		}
		return fmt.Sprintf("%dM", months)
	}
	if d.Hours() > 24 {
		days := int(d.Hours() / 24)
		hours := int(d.Hours()) % 24
		if hours > 0 {
			return fmt.Sprintf("%dd%dh", days, hours)
		}
		return fmt.Sprintf("%dd", days)
	}
	if d.Hours() >= 1 {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		if minutes > 0 {
			return fmt.Sprintf("%dh%dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}
	if d.Minutes() >= 1 {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
}

func (c *Client) DescribePod(namespace, name string) (*PodDetails, error) {
	pod, err := c.clientset.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	details := &PodDetails{
		Name:         pod.Name,
		Namespace:    pod.Namespace,
		Node:         pod.Spec.NodeName,
		Status:       string(pod.Status.Phase),
		IP:           pod.Status.PodIP,
		CreationTime: pod.CreationTimestamp.Time,
		Labels:       pod.Labels,
		NodeSelector: pod.Spec.NodeSelector,
		Containers:   make([]ContainerInfo, 0),
		Volumes:      make([]Volume, 0),
	}

	// Add container details
	for i, c := range pod.Spec.Containers {
		containerStatus := pod.Status.ContainerStatuses[i]
		container := ContainerInfo{
			Name:         c.Name,
			Image:        c.Image,
			Ready:        containerStatus.Ready,
			RestartCount: containerStatus.RestartCount,
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
			VolumeMounts: make([]VolumeMount, 0),
		}

		// Add volume mounts
		for _, vm := range c.VolumeMounts {
			container.VolumeMounts = append(container.VolumeMounts, VolumeMount{
				Name:      vm.Name,
				MountPath: vm.MountPath,
				ReadOnly:  vm.ReadOnly,
			})
		}

		// Determine container state
		if containerStatus.State.Running != nil {
			container.State = "Running"
		} else if containerStatus.State.Waiting != nil {
			container.State = "Waiting"
		} else if containerStatus.State.Terminated != nil {
			container.State = "Terminated"
		}

		details.Containers = append(details.Containers, container)
	}

	// Add volume details
	for _, v := range pod.Spec.Volumes {
		volume := Volume{
			Name: v.Name,
		}

		if v.ConfigMap != nil {
			volume.Type = "ConfigMap"
			volume.Source = v.ConfigMap.Name
		} else if v.Secret != nil {
			volume.Type = "Secret"
			volume.Source = v.Secret.SecretName
		} else if v.PersistentVolumeClaim != nil {
			volume.Type = "PVC"
			volume.Source = v.PersistentVolumeClaim.ClaimName
		} else if v.EmptyDir != nil {
			volume.Type = "EmptyDir"
			volume.Source = "N/A"
		}

		details.Volumes = append(details.Volumes, volume)
	}

	// Get events
	events, err := c.GetPodEvents(namespace, name)
	if err != nil {
		return nil, err
	}
	details.Events = events

	return details, nil
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
	selector := fmt.Sprintf("involvedObject.kind=Pod,involvedObject.name=%s", podName)
	events, err := c.clientset.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: selector,
	})
	if err != nil {
		return nil, err
	}

	var eventList []Event
	for _, event := range events.Items {
		lastSeen := time.Since(event.LastTimestamp.Time)
		eventList = append(eventList, Event{
			LastSeen: lastSeen,
			Type:     event.Type,
			Reason:   event.Reason,
			Object:   fmt.Sprintf("Pod/%s", event.InvolvedObject.Name),
			Message:  event.Message,
		})
	}

	return eventList, nil
}

func (c *Client) GetPod(namespace, name string) (*Pod, error) {
	pod, err := c.clientset.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	containers := make([]Container, 0)
	for _, c := range pod.Spec.Containers {
		containers = append(containers, Container{
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
		return err
	}

	return exec.Stream(remotecommand.StreamOptions{
		Stdin:  opts.Stdin,
		Stdout: opts.Stdout,
		Stderr: opts.Stderr,
		Tty:    opts.TTY,
	})
}

func (c *Client) DescribeDeployment(namespace, name string) error {
	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Basic Info
	fmt.Fprintf(w, "Name:\t%s\n", deployment.Name)
	fmt.Fprintf(w, "Namespace:\t%s\n", deployment.Namespace)
	fmt.Fprintf(w, "CreationTimestamp:\t%s\n", deployment.CreationTimestamp.Format(time.RFC3339))

	// Labels
	fmt.Fprintf(w, "Labels:\t")
	if len(deployment.Labels) == 0 {
		fmt.Fprintf(w, "<none>")
	} else {
		first := true
		for k, v := range deployment.Labels {
			if !first {
				fmt.Fprintf(w, ", ")
			}
			fmt.Fprintf(w, "%s=%s", k, v)
			first = false
		}
	}
	fmt.Fprintf(w, "\n")

	// Annotations
	fmt.Fprintf(w, "Annotations:\t")
	if len(deployment.Annotations) == 0 {
		fmt.Fprintf(w, "<none>")
	} else {
		first := true
		for k, v := range deployment.Annotations {
			if !first {
				fmt.Fprintf(w, ", ")
			}
			fmt.Fprintf(w, "%s=%s", k, v)
			first = false
		}
	}
	fmt.Fprintf(w, "\n")

	// Selector
	fmt.Fprintf(w, "Selector:\t")
	if deployment.Spec.Selector == nil {
		fmt.Fprintf(w, "<none>")
	} else {
		first := true
		for k, v := range deployment.Spec.Selector.MatchLabels {
			if !first {
				fmt.Fprintf(w, ", ")
			}
			fmt.Fprintf(w, "%s=%s", k, v)
			first = false
		}
	}
	fmt.Fprintf(w, "\n")

	// Replicas
	fmt.Fprintf(w, "Replicas:\t%d desired | %d updated | %d total | %d available | %d unavailable\n",
		*deployment.Spec.Replicas,
		deployment.Status.UpdatedReplicas,
		deployment.Status.Replicas,
		deployment.Status.AvailableReplicas,
		deployment.Status.UnavailableReplicas)

	// Strategy
	fmt.Fprintf(w, "Strategy:\t%s\n", string(deployment.Spec.Strategy.Type))
	if deployment.Spec.Strategy.RollingUpdate != nil {
		fmt.Fprintf(w, "RollingUpdate Strategy:\tmax unavailable: %s, max surge: %s\n",
			deployment.Spec.Strategy.RollingUpdate.MaxUnavailable.String(),
			deployment.Spec.Strategy.RollingUpdate.MaxSurge.String())
	}

	// Pod Template
	fmt.Fprintf(w, "\nPod Template:\n")
	fmt.Fprintf(w, "  Labels:\t")
	if len(deployment.Spec.Template.Labels) == 0 {
		fmt.Fprintf(w, "<none>")
	} else {
		first := true
		for k, v := range deployment.Spec.Template.Labels {
			if !first {
				fmt.Fprintf(w, ", ")
			}
			fmt.Fprintf(w, "%s=%s", k, v)
			first = false
		}
	}
	fmt.Fprintf(w, "\n")

	// Node Selector
	fmt.Fprintf(w, "  Node Selector:\t")
	if len(deployment.Spec.Template.Spec.NodeSelector) == 0 {
		fmt.Fprintf(w, "<none>")
	} else {
		first := true
		for k, v := range deployment.Spec.Template.Spec.NodeSelector {
			if !first {
				fmt.Fprintf(w, ", ")
			}
			fmt.Fprintf(w, "%s=%s", k, v)
			first = false
		}
	}
	fmt.Fprintf(w, "\n")

	// Containers
	fmt.Fprintf(w, "\nContainers:\n")
	for _, container := range deployment.Spec.Template.Spec.Containers {
		fmt.Fprintf(w, "  %s:\n", container.Name)
		fmt.Fprintf(w, "    Image:\t%s\n", container.Image)

		// Ports
		fmt.Fprintf(w, "    Ports:\t")
		if len(container.Ports) == 0 {
			fmt.Fprintf(w, "<none>")
		} else {
			first := true
			for _, port := range container.Ports {
				if !first {
					fmt.Fprintf(w, ", ")
				}
				fmt.Fprintf(w, "%d/%s", port.ContainerPort, port.Protocol)
				first = false
			}
		}
		fmt.Fprintf(w, "\n")

		// Resources
		fmt.Fprintf(w, "    Limits:\t")
		if len(container.Resources.Limits) == 0 {
			fmt.Fprintf(w, "<none>")
		} else {
			first := true
			for k, v := range container.Resources.Limits {
				if !first {
					fmt.Fprintf(w, ", ")
				}
				fmt.Fprintf(w, "%s=%s", k, v.String())
				first = false
			}
		}
		fmt.Fprintf(w, "\n")

		fmt.Fprintf(w, "    Requests:\t")
		if len(container.Resources.Requests) == 0 {
			fmt.Fprintf(w, "<none>")
		} else {
			first := true
			for k, v := range container.Resources.Requests {
				if !first {
					fmt.Fprintf(w, ", ")
				}
				fmt.Fprintf(w, "%s=%s", k, v.String())
				first = false
			}
		}
		fmt.Fprintf(w, "\n")

		// Volume Mounts
		fmt.Fprintf(w, "    Volume Mounts:\t")
		if len(container.VolumeMounts) == 0 {
			fmt.Fprintf(w, "<none>")
		} else {
			fmt.Fprintf(w, "\n")
			for _, vm := range container.VolumeMounts {
				ro := ""
				if vm.ReadOnly {
					ro = " (ro)"
				}
				fmt.Fprintf(w, "      %s:\t%s%s\n", vm.Name, vm.MountPath, ro)
			}
		}
		fmt.Fprintf(w, "\n")
	}

	// Volumes
	fmt.Fprintf(w, "  Volumes:\t")
	if len(deployment.Spec.Template.Spec.Volumes) == 0 {
		fmt.Fprintf(w, "<none>")
	} else {
		fmt.Fprintf(w, "\n")
		for _, v := range deployment.Spec.Template.Spec.Volumes {
			fmt.Fprintf(w, "    %s:\n", v.Name)
			if v.ConfigMap != nil {
				fmt.Fprintf(w, "      ConfigMap:\t%s\n", v.ConfigMap.Name)
			} else if v.Secret != nil {
				fmt.Fprintf(w, "      Secret:\t%s\n", v.Secret.SecretName)
			} else if v.PersistentVolumeClaim != nil {
				fmt.Fprintf(w, "      PVC:\t%s\n", v.PersistentVolumeClaim.ClaimName)
			} else if v.EmptyDir != nil {
				fmt.Fprintf(w, "      EmptyDir:\t{}\n")
			}
		}
	}
	fmt.Fprintf(w, "\n")

	// Conditions
	fmt.Fprintf(w, "\nConditions:\n")
	fmt.Fprintf(w, "  Type\tStatus\tLastUpdateTime\tLastTransitionTime\tReason\tMessage\n")
	for _, condition := range deployment.Status.Conditions {
		fmt.Fprintf(w, "  %s\t%s\t%s\t%s\t%s\t%s\n",
			condition.Type,
			condition.Status,
			condition.LastUpdateTime.Format(time.RFC3339),
			condition.LastTransitionTime.Format(time.RFC3339),
			condition.Reason,
			condition.Message)
	}

	return nil
}

func (p *PodDetails) Print(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	defer tw.Flush()

	fmt.Fprintf(tw, "Name:\t%s\n", p.Name)
	fmt.Fprintf(tw, "Namespace:\t%s\n", p.Namespace)
	fmt.Fprintf(tw, "Node:\t%s\n", p.Node)
	fmt.Fprintf(tw, "Status:\t%s\n", p.Status)
	fmt.Fprintf(tw, "IP:\t%s\n", p.IP)
	fmt.Fprintf(tw, "Created:\t%s\n", p.CreationTime.Format(time.RFC3339))

	// Labels
	fmt.Fprintf(tw, "\nLabels:\t")
	if len(p.Labels) == 0 {
		fmt.Fprintf(tw, "<none>")
	} else {
		first := true
		for k, v := range p.Labels {
			if !first {
				fmt.Fprintf(tw, ", ")
			}
			fmt.Fprintf(tw, "%s=%s", k, v)
			first = false
		}
	}
	fmt.Fprintf(tw, "\n")

	// Node Selector
	fmt.Fprintf(tw, "Node Selector:\t")
	if len(p.NodeSelector) == 0 {
		fmt.Fprintf(tw, "<none>")
	} else {
		first := true
		for k, v := range p.NodeSelector {
			if !first {
				fmt.Fprintf(tw, ", ")
			}
			fmt.Fprintf(tw, "%s=%s", k, v)
			first = false
		}
	}
	fmt.Fprintf(tw, "\n")

	// Containers
	fmt.Fprintf(tw, "\nContainers:\n")
	for _, c := range p.Containers {
		fmt.Fprintf(tw, "  %s:\n", c.Name)
		fmt.Fprintf(tw, "    Image:\t%s\n", c.Image)
		fmt.Fprintf(tw, "    State:\t%s\n", c.State)
		fmt.Fprintf(tw, "    Ready:\t%v\n", c.Ready)
		fmt.Fprintf(tw, "    Restarts:\t%d\n", c.RestartCount)

		// Resources
		fmt.Fprintf(tw, "    Limits:\t")
		if c.Resources.Limits.CPU == "" && c.Resources.Limits.Memory == "" {
			fmt.Fprintf(tw, "<none>\n")
		} else {
			fmt.Fprintf(tw, "\n")
			if c.Resources.Limits.CPU != "" {
				fmt.Fprintf(tw, "      CPU:\t%s\n", c.Resources.Limits.CPU)
			}
			if c.Resources.Limits.Memory != "" {
				fmt.Fprintf(tw, "      Memory:\t%s\n", c.Resources.Limits.Memory)
			}
		}

		fmt.Fprintf(tw, "    Requests:\t")
		if c.Resources.Requests.CPU == "" && c.Resources.Requests.Memory == "" {
			fmt.Fprintf(tw, "<none>\n")
		} else {
			fmt.Fprintf(tw, "\n")
			if c.Resources.Requests.CPU != "" {
				fmt.Fprintf(tw, "      CPU:\t%s\n", c.Resources.Requests.CPU)
			}
			if c.Resources.Requests.Memory != "" {
				fmt.Fprintf(tw, "      Memory:\t%s\n", c.Resources.Requests.Memory)
			}
		}

		// Volume Mounts
		fmt.Fprintf(tw, "    Volume Mounts:\t")
		if len(c.VolumeMounts) == 0 {
			fmt.Fprintf(tw, "<none>\n")
		} else {
			fmt.Fprintf(tw, "\n")
			for _, vm := range c.VolumeMounts {
				ro := ""
				if vm.ReadOnly {
					ro = " (ro)"
				}
				fmt.Fprintf(tw, "      %s:\t%s%s\n", vm.Name, vm.MountPath, ro)
			}
		}
	}

	// Volumes
	fmt.Fprintf(tw, "\nVolumes:\t")
	if len(p.Volumes) == 0 {
		fmt.Fprintf(tw, "<none>\n")
	} else {
		fmt.Fprintf(tw, "\n")
		for _, v := range p.Volumes {
			fmt.Fprintf(tw, "  %s:\n", v.Name)
			fmt.Fprintf(tw, "    Type:\t%s\n", v.Type)
			fmt.Fprintf(tw, "    Source:\t%s\n", v.Source)
		}
	}

	return nil
}
