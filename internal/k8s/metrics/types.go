package metrics

import (
	"time"
)

// ResourceMetrics represents resource usage metrics
type ResourceMetrics struct {
	CPU    CPUMetrics    `json:"cpu"`
	Memory MemoryMetrics `json:"memory"`
}

// CPUMetrics represents CPU usage metrics
type CPUMetrics struct {
	UsageNanoCores     int64   `json:"usageNanoCores"`
	UsageCorePercent   float64 `json:"usageCorePercent"`
	RequestMilliCores  int64   `json:"requestMilliCores"`
	LimitMilliCores    int64   `json:"limitMilliCores"`
	RequestUtilization float64 `json:"requestUtilization"`
	LimitUtilization   float64 `json:"limitUtilization"`
}

// MemoryMetrics represents memory usage metrics
type MemoryMetrics struct {
	UsageBytes         int64   `json:"usageBytes"`
	RequestBytes       int64   `json:"requestBytes"`
	LimitBytes         int64   `json:"limitBytes"`
	RequestUtilization float64 `json:"requestUtilization"`
	LimitUtilization   float64 `json:"limitUtilization"`
}

// PodMetrics represents metrics for a pod
type PodMetrics struct {
	Name              string                     `json:"name"`
	Namespace         string                     `json:"namespace"`
	CreationTimestamp time.Time                  `json:"creationTimestamp"`
	Containers        map[string]ResourceMetrics `json:"containers"`
	TotalResources    ResourceMetrics            `json:"totalResources"`
}

// NodeMetrics represents metrics for a node
type NodeMetrics struct {
	Name              string          `json:"name"`
	CreationTimestamp time.Time       `json:"creationTimestamp"`
	Resources         ResourceMetrics `json:"resources"`
	Allocatable       ResourceMetrics `json:"allocatable"`
	Capacity          ResourceMetrics `json:"capacity"`
	PodCount          int             `json:"podCount"`
}

// MetricsSortOption represents metrics sorting options
type MetricsSortOption string

const (
	// SortByName sorts metrics by resource name
	SortByName MetricsSortOption = "name"
	// SortByCPU sorts metrics by CPU usage
	SortByCPU MetricsSortOption = "cpu"
	// SortByMemory sorts metrics by memory usage
	SortByMemory MetricsSortOption = "memory"
	// SortByAge sorts metrics by creation timestamp
	SortByAge MetricsSortOption = "age"
)
