package pods

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned"
)

type service struct {
	clientset     *kubernetes.Clientset
	metricsClient *metricsv1beta1.Clientset
	config        *rest.Config
}

// NewPodService creates a new pod service instance
func NewPodService(clientset *kubernetes.Clientset, metricsClient *metricsv1beta1.Clientset, config *rest.Config) Service {
	return &service{
		clientset:     clientset,
		metricsClient: metricsClient,
		config:        config,
	}
}

// List returns a list of pods based on the given filters
func (s *service) List(namespace string, allNamespaces bool, selector string, statusFilter string) ([]Pod, error) {
	var pods []Pod
	var listOptions metav1.ListOptions

	if selector != "" {
		listOptions.LabelSelector = selector
	}

	if allNamespaces {
		namespace = ""
	}

	podList, err := s.clientset.CoreV1().Pods(namespace).List(context.Background(), listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	for _, p := range podList.Items {
		if statusFilter != "" && string(p.Status.Phase) != statusFilter {
			continue
		}

		pod := Pod{
			Name:      p.Name,
			Namespace: p.Namespace,
			Ready:     getPodReady(p.Status),
			Status:    string(p.Status.Phase),
			Restarts:  getPodRestarts(p.Status),
			Age:       time.Since(p.CreationTimestamp.Time),
			IP:        p.Status.PodIP,
			Node:      p.Spec.NodeName,
			Labels:    p.Labels,
		}

		// Add controller reference if available
		if len(p.OwnerReferences) > 0 {
			owner := p.OwnerReferences[0]
			pod.Controller = owner.Kind
			pod.ControllerName = owner.Name
		}

		// Add container information
		for _, c := range p.Spec.Containers {
			container := ContainerInfo{
				Name:  c.Name,
				Image: c.Image,
			}
			pod.Containers = append(pod.Containers, container)
		}

		pods = append(pods, pod)
	}

	return pods, nil
}

// Get returns a specific pod by name
func (s *service) Get(namespace, name string) (*Pod, error) {
	p, err := s.clientset.CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	pod := &Pod{
		Name:      p.Name,
		Namespace: p.Namespace,
		Ready:     getPodReady(p.Status),
		Status:    string(p.Status.Phase),
		Restarts:  getPodRestarts(p.Status),
		Age:       time.Since(p.CreationTimestamp.Time),
		IP:        p.Status.PodIP,
		Node:      p.Spec.NodeName,
		Labels:    p.Labels,
	}

	return pod, nil
}

// GetLogs retrieves logs from a pod's container
func (s *service) GetLogs(namespace, name string, container string, opts LogOptions) error {
	pod, err := s.clientset.CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}

	if opts.AllContainers {
		var wg sync.WaitGroup
		var mutex sync.Mutex

		for _, c := range pod.Spec.Containers {
			wg.Add(1)
			go func(containerName string) {
				defer wg.Done()
				err := s.getContainerLogs(pod, containerName, opts, &mutex)
				if err != nil {
					fmt.Fprintf(opts.Writer, "Error getting logs for container %s: %v\n", containerName, err)
				}
			}(c.Name)
		}

		wg.Wait()
		return nil
	}

	return s.getContainerLogs(pod, container, opts, nil)
}

// Describe returns detailed information about a pod
func (s *service) Describe(namespace, name string) (*PodDetails, error) {
	pod, err := s.clientset.CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	details := &PodDetails{
		Name:         pod.Name,
		Namespace:    pod.Namespace,
		Node:         pod.Spec.NodeName,
		Status:       string(pod.Status.Phase),
		IP:           pod.Status.PodIP,
		CreationTime: pod.CreationTimestamp.Time,
		Age:          time.Since(pod.CreationTimestamp.Time),
		Labels:       pod.Labels,
		NodeSelector: pod.Spec.NodeSelector,
	}

	// Add container information
	for _, c := range pod.Spec.Containers {
		container := ContainerInfo{
			Name:         c.Name,
			Image:        c.Image,
			Resources:    getResourcesFromContainer(c),
			VolumeMounts: getVolumeMountsFromContainer(c),
			Ports:        getPortsFromContainer(c),
		}
		details.Containers = append(details.Containers, container)
	}

	// Add volume information
	for _, v := range pod.Spec.Volumes {
		volume := Volume{
			Name: v.Name,
		}
		details.Volumes = append(details.Volumes, volume)
	}

	return details, nil
}

// GetMetrics returns resource usage metrics for a pod
func (s *service) GetMetrics(namespace, name string) (*PodMetrics, error) {
	metrics, err := s.metricsClient.MetricsV1beta1().PodMetricses(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod metrics: %w", err)
	}

	podMetrics := &PodMetrics{
		Name:      metrics.Name,
		Namespace: metrics.Namespace,
	}

	for _, container := range metrics.Containers {
		containerMetrics := ContainerMetrics{
			Name:   container.Name,
			CPU:    container.Usage.Cpu().String(),
			Memory: container.Usage.Memory().String(),
		}
		podMetrics.Containers = append(podMetrics.Containers, containerMetrics)
	}

	return podMetrics, nil
}

// GetEvents returns events related to a pod
func (s *service) GetEvents(namespace, name string) ([]Event, error) {
	fieldSelector := fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s,involvedObject.kind=Pod", name, namespace)
	events, err := s.clientset.CoreV1().Events(namespace).List(context.Background(), metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod events: %w", err)
	}

	var podEvents []Event
	for _, e := range events.Items {
		event := Event{
			Type:      e.Type,
			Reason:    e.Reason,
			Age:       time.Since(e.FirstTimestamp.Time),
			From:      e.Source.Component,
			Message:   e.Message,
			Count:     e.Count,
			FirstSeen: e.FirstTimestamp.Time,
			LastSeen:  e.LastTimestamp.Time,
			Object:    fmt.Sprintf("%s/%s", e.InvolvedObject.Kind, e.InvolvedObject.Name),
		}
		podEvents = append(podEvents, event)
	}

	return podEvents, nil
}

// Exec executes a command in a pod's container
func (s *service) Exec(namespace, name, container string, opts ExecOptions) error {
	req := s.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(name).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   opts.Command,
			Stdin:     opts.Stdin != nil,
			Stdout:    opts.Stdout != nil,
			Stderr:    opts.Stderr != nil,
			TTY:       opts.TTY,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(s.config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("failed to create executor: %w", err)
	}

	return exec.Stream(remotecommand.StreamOptions{
		Stdin:  opts.Stdin,
		Stdout: opts.Stdout,
		Stderr: opts.Stderr,
		Tty:    opts.TTY,
	})
}

// AddMetrics adds metrics information to a list of pods
func (s *service) AddMetrics(pods []Pod) error {
	for i := range pods {
		metrics, err := s.GetMetrics(pods[i].Namespace, pods[i].Name)
		if err != nil {
			continue // Skip if metrics are not available
		}
		pods[i].Metrics = metrics
	}
	return nil
}

// Helper functions

func getPodReady(status corev1.PodStatus) string {
	ready := 0
	total := len(status.ContainerStatuses)
	for _, cs := range status.ContainerStatuses {
		if cs.Ready {
			ready++
		}
	}
	return fmt.Sprintf("%d/%d", ready, total)
}

func getPodRestarts(status corev1.PodStatus) int32 {
	var restarts int32
	for _, cs := range status.ContainerStatuses {
		restarts += cs.RestartCount
	}
	return restarts
}

func (s *service) getContainerLogs(pod *corev1.Pod, containerName string, opts LogOptions, mutex *sync.Mutex) error {
	logOptions := &corev1.PodLogOptions{
		Follow:     opts.Follow,
		Previous:   opts.Previous,
		TailLines:  &opts.TailLines,
		Container:  containerName,
		Timestamps: true,
	}

	if opts.SinceTime != nil {
		logOptions.SinceTime = &metav1.Time{Time: *opts.SinceTime}
	}

	if opts.SinceSeconds != nil {
		logOptions.SinceSeconds = opts.SinceSeconds
	}

	req := s.clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, logOptions)
	stream, err := req.Stream(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get log stream: %w", err)
	}
	defer stream.Close()

	if mutex != nil {
		mutex.Lock()
		defer mutex.Unlock()
	}

	_, err = io.Copy(opts.Writer, stream)
	return err
}

func getResourcesFromContainer(c corev1.Container) Resources {
	return Resources{
		Limits: Resource{
			CPU:    c.Resources.Limits.Cpu().String(),
			Memory: c.Resources.Limits.Memory().String(),
		},
		Requests: Resource{
			CPU:    c.Resources.Requests.Cpu().String(),
			Memory: c.Resources.Requests.Memory().String(),
		},
	}
}

func getVolumeMountsFromContainer(c corev1.Container) []VolumeMount {
	var mounts []VolumeMount
	for _, m := range c.VolumeMounts {
		mount := VolumeMount{
			Name:      m.Name,
			MountPath: m.MountPath,
			ReadOnly:  m.ReadOnly,
		}
		mounts = append(mounts, mount)
	}
	return mounts
}

func getPortsFromContainer(c corev1.Container) []ContainerPort {
	var ports []ContainerPort
	for _, p := range c.Ports {
		port := ContainerPort{
			Name:          p.Name,
			HostPort:      p.HostPort,
			ContainerPort: p.ContainerPort,
			Protocol:      string(p.Protocol),
		}
		ports = append(ports, port)
	}
	return ports
}
