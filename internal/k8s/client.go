package k8s

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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
	NodeName     string
	Status       string
	PodIP        string
	CreatedAt    string
	Labels       map[string]string
	Containers   []ContainerDetails
	Events       []EventDetails
	Volumes      []VolumeDetails
	NodeSelector map[string]string
	Tolerations  []corev1.Toleration
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

func (c *Client) GetPodLogs(podName, namespace string) (string, error) {
	req := c.clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{})
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		return "", err
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
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

	// Get pod events
	events, err := c.clientset.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s", name),
	})
	if err != nil {
		return nil, err
	}

	details := &PodDetails{
		Name:         pod.Name,
		Namespace:    pod.Namespace,
		NodeName:     pod.Spec.NodeName,
		Status:       string(pod.Status.Phase),
		PodIP:        pod.Status.PodIP,
		CreatedAt:    pod.CreationTimestamp.Format("2006-01-02 15:04:05"),
		Labels:       pod.Labels,
		NodeSelector: pod.Spec.NodeSelector,
		Tolerations:  pod.Spec.Tolerations,
	}

	// Get volumes
	for _, volume := range pod.Spec.Volumes {
		volumeDetail := VolumeDetails{
			Name: volume.Name,
		}

		// Determine volume type and source
		switch {
		case volume.ConfigMap != nil:
			volumeDetail.Type = "ConfigMap"
			volumeDetail.Source = volume.ConfigMap.Name
		case volume.Secret != nil:
			volumeDetail.Type = "Secret"
			volumeDetail.Source = volume.Secret.SecretName
		case volume.PersistentVolumeClaim != nil:
			volumeDetail.Type = "PVC"
			volumeDetail.Source = volume.PersistentVolumeClaim.ClaimName
		case volume.EmptyDir != nil:
			volumeDetail.Type = "EmptyDir"
			volumeDetail.Source = "N/A"
		case volume.HostPath != nil:
			volumeDetail.Type = "HostPath"
			volumeDetail.Source = volume.HostPath.Path
		default:
			volumeDetail.Type = "Other"
			volumeDetail.Source = "N/A"
		}

		details.Volumes = append(details.Volumes, volumeDetail)
	}

	// Get container details
	for _, container := range pod.Spec.Containers {
		var containerStatus *corev1.ContainerStatus
		for i := range pod.Status.ContainerStatuses {
			if pod.Status.ContainerStatuses[i].Name == container.Name {
				containerStatus = &pod.Status.ContainerStatuses[i]
				break
			}
		}

		cd := ContainerDetails{
			Name:  container.Name,
			Image: container.Image,
		}

		// Add volume mounts
		for _, mount := range container.VolumeMounts {
			cd.Mounts = append(cd.Mounts, MountDetails{
				Name:      mount.Name,
				MountPath: mount.MountPath,
				ReadOnly:  mount.ReadOnly,
			})
		}

		// Add resource requests and limits
		cd.Resources = ResourceDetails{
			Requests: ResourceQuantity{
				CPU:    container.Resources.Requests.Cpu().String(),
				Memory: container.Resources.Requests.Memory().String(),
			},
			Limits: ResourceQuantity{
				CPU:    container.Resources.Limits.Cpu().String(),
				Memory: container.Resources.Limits.Memory().String(),
			},
		}

		if containerStatus != nil {
			cd.Ready = containerStatus.Ready
			cd.RestartCount = containerStatus.RestartCount
			cd.State = getContainerState(containerStatus.State)
			cd.LastState = getContainerState(containerStatus.LastTerminationState)
		}

		// Get container ports
		for _, port := range container.Ports {
			cd.Ports = append(cd.Ports, fmt.Sprintf("%d/%s", port.ContainerPort, port.Protocol))
		}

		details.Containers = append(details.Containers, cd)
	}

	// Get recent events
	for _, event := range events.Items {
		details.Events = append(details.Events, EventDetails{
			Time:    event.LastTimestamp.Format("2006-01-02 15:04:05"),
			Type:    event.Type,
			Message: event.Message,
		})
	}

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
