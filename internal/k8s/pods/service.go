package pods

import (
	"context"
	"fmt"
	"io"
	"strings"
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

			// Add container ports
			for _, p := range c.Ports {
				port := ContainerPort{
					Name:          p.Name,
					ContainerPort: p.ContainerPort,
					HostPort:      p.HostPort,
					Protocol:      string(p.Protocol),
				}
				container.Ports = append(container.Ports, port)
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

	// Add container information
	for _, c := range p.Spec.Containers {
		container := ContainerInfo{
			Name:  c.Name,
			Image: c.Image,
		}

		// Add container ports
		for _, p := range c.Ports {
			port := ContainerPort{
				Name:          p.Name,
				ContainerPort: p.ContainerPort,
				HostPort:      p.HostPort,
				Protocol:      string(p.Protocol),
			}
			container.Ports = append(container.Ports, port)
		}

		pod.Containers = append(pod.Containers, container)
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
		Name:           pod.Name,
		Namespace:      pod.Namespace,
		HasPriority:    pod.Spec.Priority != nil,
		ServiceAccount: pod.Spec.ServiceAccountName,
		Node:           fmt.Sprintf("%s/%s", pod.Spec.NodeName, pod.Status.HostIP),
		NodeIP:         pod.Status.HostIP,
		StartTime:      pod.Status.StartTime.Time,
		Status:         string(pod.Status.Phase),
		Phase:          string(pod.Status.Phase),
		IP:             pod.Status.PodIP,
		IPs:            make([]string, 0),
		ControlledBy:   getControllerRef(pod),
		QoSClass:       string(pod.Status.QOSClass),
		CreationTime:   pod.CreationTimestamp.Time,
		Labels:         pod.Labels,
		Annotations:    pod.Annotations,
		NodeSelector:   pod.Spec.NodeSelector,
	}

	// Add IPs
	for _, ip := range pod.Status.PodIPs {
		details.IPs = append(details.IPs, ip.IP)
	}

	// Add container information
	for _, c := range pod.Spec.Containers {
		container := ContainerInfo{
			Name:         c.Name,
			Image:        c.Image,
			ContainerID:  getContainerID(pod, c.Name),
			ImageID:      getContainerImageID(pod, c.Name),
			Ports:        make([]ContainerPort, 0),
			State:        getContainerState(pod, c.Name),
			Ready:        isContainerReady(pod, c.Name),
			RestartCount: getContainerRestartCount(pod, c.Name),
		}

		// Add ports
		for _, p := range c.Ports {
			port := ContainerPort{
				Name:          p.Name,
				ContainerPort: p.ContainerPort,
				HostPort:      p.HostPort,
				Protocol:      string(p.Protocol),
			}
			container.Ports = append(container.Ports, port)
		}

		// Add resources
		container.Resources = Resources{
			Requests: Resource{
				CPU:    c.Resources.Requests.Cpu().String(),
				Memory: c.Resources.Requests.Memory().String(),
			},
			Limits: Resource{
				CPU:    c.Resources.Limits.Cpu().String(),
				Memory: c.Resources.Limits.Memory().String(),
			},
		}

		// Add readiness probe
		if c.ReadinessProbe != nil {
			probe := &Probe{
				Delay:            time.Duration(c.ReadinessProbe.InitialDelaySeconds) * time.Second,
				Timeout:          time.Duration(c.ReadinessProbe.TimeoutSeconds) * time.Second,
				Period:           time.Duration(c.ReadinessProbe.PeriodSeconds) * time.Second,
				SuccessThreshold: c.ReadinessProbe.SuccessThreshold,
				FailureThreshold: c.ReadinessProbe.FailureThreshold,
			}
			if c.ReadinessProbe.TCPSocket != nil {
				probe.Type = "tcp-socket"
				probe.Port = c.ReadinessProbe.TCPSocket.Port.IntVal
			} else if c.ReadinessProbe.HTTPGet != nil {
				probe.Type = "http-get"
				probe.Port = c.ReadinessProbe.HTTPGet.Port.IntVal
				probe.Path = c.ReadinessProbe.HTTPGet.Path
			} else if c.ReadinessProbe.Exec != nil {
				probe.Type = "exec"
			}
			container.ReadinessProbe = probe
		}

		// Add environment variables
		for _, env := range c.EnvFrom {
			var envFrom EnvFromSource
			if env.ConfigMapRef != nil {
				envFrom = EnvFromSource{
					Name:     env.ConfigMapRef.Name,
					Type:     "ConfigMap",
					Optional: *env.ConfigMapRef.Optional,
				}
			} else if env.SecretRef != nil {
				envFrom = EnvFromSource{
					Name:     env.SecretRef.Name,
					Type:     "Secret",
					Optional: *env.SecretRef.Optional,
				}
			}
			container.EnvFrom = append(container.EnvFrom, envFrom)
		}

		for _, env := range c.Env {
			envVar := EnvVar{
				Name: env.Name,
			}
			if env.Value != "" {
				envVar.Value = env.Value
			} else if env.ValueFrom != nil {
				if env.ValueFrom.ConfigMapKeyRef != nil {
					envVar.ValueFrom = fmt.Sprintf("configmap %s/%s", env.ValueFrom.ConfigMapKeyRef.Name, env.ValueFrom.ConfigMapKeyRef.Key)
				} else if env.ValueFrom.SecretKeyRef != nil {
					envVar.ValueFrom = fmt.Sprintf("secret %s/%s", env.ValueFrom.SecretKeyRef.Name, env.ValueFrom.SecretKeyRef.Key)
				} else if env.ValueFrom.FieldRef != nil {
					envVar.ValueFrom = fmt.Sprintf("field %s", env.ValueFrom.FieldRef.FieldPath)
				}
			}
			container.Env = append(container.Env, envVar)
		}

		// Add volume mounts
		for _, vm := range c.VolumeMounts {
			mount := VolumeMount{
				Name:      vm.Name,
				MountPath: vm.MountPath,
				ReadOnly:  vm.ReadOnly,
			}
			container.VolumeMounts = append(container.VolumeMounts, mount)
		}

		details.Containers = append(details.Containers, container)
	}

	// Add conditions
	for _, c := range pod.Status.Conditions {
		condition := PodCondition{
			Type:   string(c.Type),
			Status: string(c.Status),
		}
		details.Conditions = append(details.Conditions, condition)
	}

	// Add volumes
	for _, v := range pod.Spec.Volumes {
		volume := VolumeInfo{
			Name: v.Name,
		}
		if v.Projected != nil {
			volume.Type = "Projected"
			for _, s := range v.Projected.Sources {
				if s.ServiceAccountToken != nil && s.ServiceAccountToken.ExpirationSeconds != nil {
					volume.TokenExpirationSeconds = *s.ServiceAccountToken.ExpirationSeconds
				}
				if s.ConfigMap != nil {
					volume.ConfigMapName = s.ConfigMap.Name
					volume.ConfigMapOptional = s.ConfigMap.Optional
				}
				if s.DownwardAPI != nil {
					volume.DownwardAPI = true
				}
			}
		}
		details.Volumes = append(details.Volumes, volume)
	}

	// Add tolerations
	for _, t := range pod.Spec.Tolerations {
		toleration := Toleration{
			Key:               t.Key,
			Operator:          string(t.Operator),
			Value:             t.Value,
			Effect:            string(t.Effect),
			TolerationSeconds: t.TolerationSeconds,
		}
		details.Tolerations = append(details.Tolerations, toleration)
	}

	// Get events
	events, err := s.getEvents(namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get pod events: %w", err)
	}
	details.Events = events

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

	var totalCPU int64
	var totalMemory int64

	for _, container := range metrics.Containers {
		cpuQuantity := container.Usage.Cpu().MilliValue()
		memoryBytes := container.Usage.Memory().Value()

		totalCPU += cpuQuantity
		totalMemory += memoryBytes

		containerMetrics := ContainerMetrics{
			Name:   container.Name,
			CPU:    fmt.Sprintf("%dm", cpuQuantity),
			Memory: fmt.Sprintf("%dMi", memoryBytes/(1024*1024)),
		}
		podMetrics.Containers = append(podMetrics.Containers, containerMetrics)
	}

	// Set total pod metrics
	podMetrics.CPU = fmt.Sprintf("%dm", totalCPU)
	podMetrics.Memory = fmt.Sprintf("%dMi", totalMemory/(1024*1024))

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
			Type:    e.Type,
			Reason:  e.Reason,
			Age:     time.Since(e.FirstTimestamp.Time),
			From:    e.Source.Component,
			Message: e.Message,
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
	if s.metricsClient == nil {
		return fmt.Errorf("metrics-server not available: metrics client is nil")
	}

	for i := range pods {
		metrics, err := s.GetMetrics(pods[i].Namespace, pods[i].Name)
		if err != nil {
			// Set default metrics instead of showing warning
			pods[i].Metrics = &PodMetrics{
				Name:      pods[i].Name,
				Namespace: pods[i].Namespace,
				CPU:       "0m",
				Memory:    "0Mi",
			}
			continue
		}
		pods[i].Metrics = metrics
	}
	return nil
}

// ListMetrics returns resource usage metrics for all pods in a namespace
func (s *service) ListMetrics(namespace string) ([]PodMetrics, error) {
	metrics, err := s.metricsClient.MetricsV1beta1().PodMetricses(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pod metrics: %w", err)
	}

	var podMetricsList []PodMetrics

	for _, podMetrics := range metrics.Items {
		pm := PodMetrics{
			Name:      podMetrics.Name,
			Namespace: podMetrics.Namespace,
		}

		var totalCPU int64
		var totalMemory int64

		for _, container := range podMetrics.Containers {
			cpuQuantity := container.Usage.Cpu().MilliValue()
			memoryBytes := container.Usage.Memory().Value()

			totalCPU += cpuQuantity
			totalMemory += memoryBytes

			containerMetrics := ContainerMetrics{
				Name:   container.Name,
				CPU:    fmt.Sprintf("%dm", cpuQuantity),
				Memory: fmt.Sprintf("%dMi", memoryBytes/(1024*1024)),
			}
			pm.Containers = append(pm.Containers, containerMetrics)
		}

		// Set total pod metrics
		pm.CPU = fmt.Sprintf("%dm", totalCPU)
		pm.Memory = fmt.Sprintf("%dMi", totalMemory/(1024*1024))

		podMetricsList = append(podMetricsList, pm)
	}

	return podMetricsList, nil
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

func getContainerID(pod *corev1.Pod, containerName string) string {
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Name == containerName {
			// Strip the container runtime prefix (e.g., "docker://" or "containerd://")
			parts := strings.Split(cs.ContainerID, "://")
			if len(parts) == 2 {
				return parts[1]
			}
			return cs.ContainerID
		}
	}
	return ""
}

func getContainerImageID(pod *corev1.Pod, containerName string) string {
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Name == containerName {
			return cs.ImageID
		}
	}
	return ""
}

func getContainerState(pod *corev1.Pod, containerName string) ContainerState {
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Name == containerName {
			state := ContainerState{}
			if cs.State.Running != nil {
				state.Status = "Running"
				state.Started = cs.State.Running.StartedAt.Time
			} else if cs.State.Waiting != nil {
				state.Status = "Waiting"
				state.Reason = cs.State.Waiting.Reason
				state.Message = cs.State.Waiting.Message
			} else if cs.State.Terminated != nil {
				state.Status = "Terminated"
				state.ExitCode = cs.State.Terminated.ExitCode
				state.Reason = cs.State.Terminated.Reason
				state.Message = cs.State.Terminated.Message
			}
			return state
		}
	}
	return ContainerState{}
}

func isContainerReady(pod *corev1.Pod, containerName string) bool {
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Name == containerName {
			return cs.Ready
		}
	}
	return false
}

func getContainerRestartCount(pod *corev1.Pod, containerName string) int32 {
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Name == containerName {
			return cs.RestartCount
		}
	}
	return 0
}

func getControllerRef(pod *corev1.Pod) string {
	if ref := metav1.GetControllerOf(pod); ref != nil {
		return fmt.Sprintf("%s/%s", ref.Kind, ref.Name)
	}
	return ""
}

// getEvents returns events for a pod
func (s *service) getEvents(namespace, name string) ([]Event, error) {
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
			Type:    e.Type,
			Reason:  e.Reason,
			Age:     time.Since(e.FirstTimestamp.Time),
			From:    e.Source.Component,
			Message: e.Message,
		}
		podEvents = append(podEvents, event)
	}

	return podEvents, nil
}
