package k8s

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

// MetricsHandler handles resource metrics operations
func (a *Agent) MetricsHandler(ctx context.Context, params TaskParams) (*TaskResult, error) {
	switch params.Action {
	case "get", "top":
		if params.ResourceType == "pod" || params.ResourceType == "pods" {
			return a.getPodMetrics(ctx, params)
		} else if params.ResourceType == "node" || params.ResourceType == "nodes" {
			return a.getNodeMetrics(ctx, params)
		}
		return nil, fmt.Errorf("unsupported resource type for metrics: %s", params.ResourceType)
	default:
		return nil, fmt.Errorf("unsupported metrics action: %s", params.Action)
	}
}

// getPodMetrics gets resource usage metrics for pods
func (a *Agent) getPodMetrics(ctx context.Context, params TaskParams) (*TaskResult, error) {
	namespace := params.Namespace
	if namespace == "" {
		namespace = a.k8sContext.Namespace
	}

	// Create metrics client
	metricsClient, err := metrics.NewForConfig(a.k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client: %w", err)
	}

	// Get pod metrics
	var podMetrics *v1beta1.PodMetricsList
	if params.ResourceName != "" {
		metric, err := metricsClient.MetricsV1beta1().PodMetricses(namespace).Get(ctx, params.ResourceName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get pod metrics: %w", err)
		}
		podMetrics = &v1beta1.PodMetricsList{
			Items: []v1beta1.PodMetrics{*metric},
		}
	} else {
		podMetrics, err = metricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to list pod metrics: %w", err)
		}
	}

	var output strings.Builder
	output.WriteString("POD\tCPU(cores)\tMEMORY(bytes)\n")
	output.WriteString("---\t----------\t-------------\n")

	for _, pod := range podMetrics.Items {
		var cpuTotal int64
		var memoryTotal int64

		for _, container := range pod.Containers {
			cpu := container.Usage.Cpu().MilliValue()
			memory := container.Usage.Memory().Value()
			cpuTotal += cpu
			memoryTotal += memory
		}

		output.WriteString(fmt.Sprintf("%s\t%dm\t%dMi\n",
			pod.Name,
			cpuTotal,
			memoryTotal/(1024*1024),
		))
	}

	return &TaskResult{
		Success: true,
		Output:  output.String(),
	}, nil
}

// getNodeMetrics gets resource usage metrics for nodes
func (a *Agent) getNodeMetrics(ctx context.Context, params TaskParams) (*TaskResult, error) {
	// Create metrics client
	metricsClient, err := metrics.NewForConfig(a.k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client: %w", err)
	}

	// Get node metrics
	var nodeMetrics *v1beta1.NodeMetricsList
	if params.ResourceName != "" {
		metric, err := metricsClient.MetricsV1beta1().NodeMetricses().Get(ctx, params.ResourceName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get node metrics: %w", err)
		}
		nodeMetrics = &v1beta1.NodeMetricsList{
			Items: []v1beta1.NodeMetrics{*metric},
		}
	} else {
		nodeMetrics, err = metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to list node metrics: %w", err)
		}
	}

	var output strings.Builder
	output.WriteString("NODE\tCPU(cores)\tCPU%\tMEMORY(bytes)\tMEMORY%\n")
	output.WriteString("----\t----------\t----\t-------------\t-------\n")

	for _, node := range nodeMetrics.Items {
		cpu := node.Usage.Cpu().MilliValue()
		memory := node.Usage.Memory().Value()

		// Get node capacity
		nodeInfo, err := a.k8sClient.CoreV1().Nodes().Get(ctx, node.Name, metav1.GetOptions{})
		if err != nil {
			continue
		}

		cpuCapacity := nodeInfo.Status.Capacity.Cpu().MilliValue()
		memoryCapacity := nodeInfo.Status.Capacity.Memory().Value()

		cpuPercent := float64(cpu) / float64(cpuCapacity) * 100
		memoryPercent := float64(memory) / float64(memoryCapacity) * 100

		output.WriteString(fmt.Sprintf("%s\t%dm\t%.1f%%\t%dMi\t%.1f%%\n",
			node.Name,
			cpu,
			cpuPercent,
			memory/(1024*1024),
			memoryPercent,
		))
	}

	return &TaskResult{
		Success: true,
		Output:  output.String(),
	}, nil
}
