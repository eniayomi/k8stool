package describe

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type service struct {
	clientset *kubernetes.Clientset
}

// NewDescribeService creates a new describe service instance
func NewDescribeService(clientset *kubernetes.Clientset) (DescribeService, error) {
	if clientset == nil {
		return nil, fmt.Errorf("kubernetes clientset is required")
	}
	return &service{clientset: clientset}, nil
}

// DescribePod returns a detailed description of a pod
func (s *service) DescribePod(ctx context.Context, namespace, name string) (*ResourceDescription, error) {
	pod, err := s.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	details := &PodDetails{
		Phase:      pod.Status.Phase,
		Conditions: pod.Status.Conditions,
		Node:       pod.Spec.NodeName,
		IP:         pod.Status.PodIP,
	}

	// Get container details
	for _, container := range pod.Spec.Containers {
		var state string
		var ready bool
		var restartCount int32

		for _, status := range pod.Status.ContainerStatuses {
			if status.Name == container.Name {
				if status.State.Running != nil {
					state = "Running"
				} else if status.State.Waiting != nil {
					state = "Waiting"
				} else if status.State.Terminated != nil {
					state = "Terminated"
				}
				ready = status.Ready
				restartCount = status.RestartCount
				break
			}
		}

		containerDetails := ContainerDetails{
			Name:         container.Name,
			Image:        container.Image,
			State:        state,
			Ready:        ready,
			RestartCount: restartCount,
		}

		// Add ports
		for _, port := range container.Ports {
			containerDetails.Ports = append(containerDetails.Ports, ContainerPort{
				Name:          port.Name,
				Protocol:      string(port.Protocol),
				ContainerPort: port.ContainerPort,
				HostPort:      port.HostPort,
			})
		}

		// Add resources
		containerDetails.Resources = ResourceRequirements{
			Requests: make(ResourceList),
			Limits:   make(ResourceList),
		}
		for resource, quantity := range container.Resources.Requests {
			containerDetails.Resources.Requests[string(resource)] = quantity.String()
		}
		for resource, quantity := range container.Resources.Limits {
			containerDetails.Resources.Limits[string(resource)] = quantity.String()
		}

		details.Containers = append(details.Containers, containerDetails)
	}

	// Get volume details
	for _, volume := range pod.Spec.Volumes {
		volumeDetails := VolumeDetails{
			Name: volume.Name,
		}

		// Determine volume type and source
		switch {
		case volume.ConfigMap != nil:
			volumeDetails.Type = "ConfigMap"
			volumeDetails.Source = volume.ConfigMap.Name
		case volume.Secret != nil:
			volumeDetails.Type = "Secret"
			volumeDetails.Source = volume.Secret.SecretName
		case volume.PersistentVolumeClaim != nil:
			volumeDetails.Type = "PersistentVolumeClaim"
			volumeDetails.Source = volume.PersistentVolumeClaim.ClaimName
		case volume.EmptyDir != nil:
			volumeDetails.Type = "EmptyDir"
		case volume.HostPath != nil:
			volumeDetails.Type = "HostPath"
			volumeDetails.Source = volume.HostPath.Path
		}

		details.Volumes = append(details.Volumes, volumeDetails)
	}

	return &ResourceDescription{
		Type:              Pod,
		Name:              pod.Name,
		Namespace:         pod.Namespace,
		CreationTimestamp: pod.CreationTimestamp.Time,
		Labels:            pod.Labels,
		Annotations:       pod.Annotations,
		Status:            string(pod.Status.Phase),
		Details:           details,
	}, nil
}

// DescribeDeployment returns a detailed description of a deployment
func (s *service) DescribeDeployment(ctx context.Context, namespace, name string) (*ResourceDescription, error) {
	deployment, err := s.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	details := struct {
		Replicas          int32             `json:"replicas"`
		AvailableReplicas int32             `json:"availableReplicas"`
		UpdatedReplicas   int32             `json:"updatedReplicas"`
		Strategy          string            `json:"strategy"`
		Selector          map[string]string `json:"selector"`
		Conditions        []string          `json:"conditions"`
	}{
		Replicas:          deployment.Status.Replicas,
		AvailableReplicas: deployment.Status.AvailableReplicas,
		UpdatedReplicas:   deployment.Status.UpdatedReplicas,
		Strategy:          string(deployment.Spec.Strategy.Type),
		Selector:          deployment.Spec.Selector.MatchLabels,
	}

	for _, condition := range deployment.Status.Conditions {
		details.Conditions = append(details.Conditions, fmt.Sprintf("%s: %s", condition.Type, condition.Status))
	}

	return &ResourceDescription{
		Type:              Deployment,
		Name:              deployment.Name,
		Namespace:         deployment.Namespace,
		CreationTimestamp: deployment.CreationTimestamp.Time,
		Labels:            deployment.Labels,
		Annotations:       deployment.Annotations,
		Status:            s.getDeploymentStatus(deployment),
		Details:           details,
	}, nil
}

// DescribeService returns a detailed description of a service
func (s *service) DescribeService(ctx context.Context, namespace, name string) (*ResourceDescription, error) {
	svc, err := s.clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	details := struct {
		Type            string             `json:"type"`
		ClusterIP       string             `json:"clusterIP"`
		ExternalIPs     []string           `json:"externalIPs,omitempty"`
		Ports           []ServicePort      `json:"ports"`
		Selector        map[string]string  `json:"selector,omitempty"`
		LoadBalancer    LoadBalancerStatus `json:"loadBalancer,omitempty"`
		SessionAffinity string             `json:"sessionAffinity"`
	}{
		Type:            string(svc.Spec.Type),
		ClusterIP:       svc.Spec.ClusterIP,
		ExternalIPs:     svc.Spec.ExternalIPs,
		Selector:        svc.Spec.Selector,
		SessionAffinity: string(svc.Spec.SessionAffinity),
	}

	for _, port := range svc.Spec.Ports {
		details.Ports = append(details.Ports, ServicePort{
			Name:       port.Name,
			Protocol:   string(port.Protocol),
			Port:       port.Port,
			TargetPort: port.TargetPort.String(),
			NodePort:   port.NodePort,
		})
	}

	if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
		for _, ingress := range svc.Status.LoadBalancer.Ingress {
			details.LoadBalancer.Ingress = append(details.LoadBalancer.Ingress, LoadBalancerIngress{
				IP:       ingress.IP,
				Hostname: ingress.Hostname,
			})
		}
	}

	return &ResourceDescription{
		Type:              ResourceType("service"),
		Name:              svc.Name,
		Namespace:         svc.Namespace,
		CreationTimestamp: svc.CreationTimestamp.Time,
		Labels:            svc.Labels,
		Annotations:       svc.Annotations,
		Status:            s.getServiceStatus(svc),
		Details:           details,
	}, nil
}

// DescribeNode returns a detailed description of a node
func (s *service) DescribeNode(ctx context.Context, name string) (*ResourceDescription, error) {
	node, err := s.clientset.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	details := struct {
		Addresses   []NodeAddress    `json:"addresses"`
		Capacity    ResourceList     `json:"capacity"`
		Allocatable ResourceList     `json:"allocatable"`
		Conditions  []NodeCondition  `json:"conditions"`
		Info        NodeSystemInfo   `json:"info"`
		Images      []ContainerImage `json:"images"`
	}{
		Capacity:    make(ResourceList),
		Allocatable: make(ResourceList),
	}

	for _, addr := range node.Status.Addresses {
		details.Addresses = append(details.Addresses, NodeAddress{
			Type:    string(addr.Type),
			Address: addr.Address,
		})
	}

	for resource, quantity := range node.Status.Capacity {
		details.Capacity[string(resource)] = quantity.String()
	}

	for resource, quantity := range node.Status.Allocatable {
		details.Allocatable[string(resource)] = quantity.String()
	}

	for _, condition := range node.Status.Conditions {
		details.Conditions = append(details.Conditions, NodeCondition{
			Type:               string(condition.Type),
			Status:             string(condition.Status),
			LastHeartbeatTime:  condition.LastHeartbeatTime.Time,
			LastTransitionTime: condition.LastTransitionTime.Time,
			Reason:             condition.Reason,
			Message:            condition.Message,
		})
	}

	details.Info = NodeSystemInfo{
		MachineID:               node.Status.NodeInfo.MachineID,
		SystemUUID:              node.Status.NodeInfo.SystemUUID,
		BootID:                  node.Status.NodeInfo.BootID,
		KernelVersion:           node.Status.NodeInfo.KernelVersion,
		OSImage:                 node.Status.NodeInfo.OSImage,
		ContainerRuntimeVersion: node.Status.NodeInfo.ContainerRuntimeVersion,
		KubeletVersion:          node.Status.NodeInfo.KubeletVersion,
		KubeProxyVersion:        "",
		OperatingSystem:         node.Status.NodeInfo.OperatingSystem,
		Architecture:            node.Status.NodeInfo.Architecture,
	}

	for _, image := range node.Status.Images {
		var names []string
		names = append(names, image.Names...)
		details.Images = append(details.Images, ContainerImage{
			Names:     names,
			SizeBytes: image.SizeBytes,
		})
	}

	return &ResourceDescription{
		Type:              Node,
		Name:              node.Name,
		CreationTimestamp: node.CreationTimestamp.Time,
		Labels:            node.Labels,
		Annotations:       node.Annotations,
		Status:            s.getNodeStatus(node),
		Details:           details,
	}, nil
}

// DescribeNamespace returns a detailed description of a namespace
func (s *service) DescribeNamespace(ctx context.Context, name string) (*ResourceDescription, error) {
	namespace, err := s.clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace: %w", err)
	}

	details := struct {
		Phase      string   `json:"phase"`
		Finalizers []string `json:"finalizers,omitempty"`
	}{
		Phase:      string(namespace.Status.Phase),
		Finalizers: []string{},
	}

	for _, finalizer := range namespace.Spec.Finalizers {
		details.Finalizers = append(details.Finalizers, string(finalizer))
	}

	return &ResourceDescription{
		Type:              Namespace,
		Name:              namespace.Name,
		CreationTimestamp: namespace.CreationTimestamp.Time,
		Labels:            namespace.Labels,
		Annotations:       namespace.Annotations,
		Status:            string(namespace.Status.Phase),
		Details:           details,
	}, nil
}

// Describe returns a detailed description of any supported resource
func (s *service) Describe(ctx context.Context, resourceType ResourceType, namespace, name string) (*ResourceDescription, error) {
	switch resourceType {
	case Pod:
		return s.DescribePod(ctx, namespace, name)
	case Deployment:
		return s.DescribeDeployment(ctx, namespace, name)
	case ResourceType("service"):
		return s.DescribeService(ctx, namespace, name)
	case Node:
		return s.DescribeNode(ctx, name)
	case Namespace:
		return s.DescribeNamespace(ctx, name)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// Helper types and functions

type ServicePort struct {
	Name       string `json:"name,omitempty"`
	Protocol   string `json:"protocol"`
	Port       int32  `json:"port"`
	TargetPort string `json:"targetPort"`
	NodePort   int32  `json:"nodePort,omitempty"`
}

type LoadBalancerStatus struct {
	Ingress []LoadBalancerIngress `json:"ingress,omitempty"`
}

type LoadBalancerIngress struct {
	IP       string `json:"ip,omitempty"`
	Hostname string `json:"hostname,omitempty"`
}

type NodeAddress struct {
	Type    string `json:"type"`
	Address string `json:"address"`
}

type NodeCondition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	LastHeartbeatTime  time.Time `json:"lastHeartbeatTime"`
	LastTransitionTime time.Time `json:"lastTransitionTime"`
	Reason             string    `json:"reason"`
	Message            string    `json:"message"`
}

type NodeSystemInfo struct {
	MachineID               string `json:"machineID"`
	SystemUUID              string `json:"systemUUID"`
	BootID                  string `json:"bootID"`
	KernelVersion           string `json:"kernelVersion"`
	OSImage                 string `json:"osImage"`
	ContainerRuntimeVersion string `json:"containerRuntimeVersion"`
	KubeletVersion          string `json:"kubeletVersion"`
	KubeProxyVersion        string `json:"kubeProxyVersion"`
	OperatingSystem         string `json:"operatingSystem"`
	Architecture            string `json:"architecture"`
}

type ContainerImage struct {
	Names     []string `json:"names"`
	SizeBytes int64    `json:"sizeBytes"`
}

func (s *service) getDeploymentStatus(deployment *appsv1.Deployment) string {
	if deployment.Generation <= deployment.Status.ObservedGeneration {
		if deployment.Status.UpdatedReplicas == *deployment.Spec.Replicas {
			if deployment.Status.AvailableReplicas == *deployment.Spec.Replicas {
				return "Available"
			}
			return "Progressing"
		}
		return "Updating"
	}
	return "Unknown"
}

func (s *service) getServiceStatus(svc *corev1.Service) string {
	if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			return "Active"
		}
		return "Pending"
	}
	return "Active"
}

func (s *service) getNodeStatus(node *corev1.Node) string {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			if condition.Status == corev1.ConditionTrue {
				return "Ready"
			}
			return "NotReady"
		}
	}
	return "Unknown"
}
