package namespace

import (
	"time"

	corev1 "k8s.io/api/core/v1"
)

// Namespace represents a Kubernetes namespace
type Namespace struct {
	Name              string
	Status            string
	CreationTimestamp time.Time
	Labels            map[string]string
	Annotations       map[string]string
	Phase             corev1.NamespacePhase
}

// NamespaceDetails contains detailed information about a namespace
type NamespaceDetails struct {
	Namespace
	ResourceQuotas []ResourceQuota
	LimitRanges    []LimitRange
}

// ResourceQuota represents resource quotas for a namespace
type ResourceQuota struct {
	Name   string
	Hard   ResourceList
	Used   ResourceList
	Scopes []string
}

// LimitRange represents limit ranges for a namespace
type LimitRange struct {
	Name    string
	Type    string
	Min     ResourceList
	Max     ResourceList
	Default ResourceList
}

// ResourceList represents a map of resource names to quantities
type ResourceList map[string]string

// NamespaceSortOption represents namespace sorting options
type NamespaceSortOption string

const (
	// SortByName sorts namespaces by name
	SortByName NamespaceSortOption = "name"
	// SortByAge sorts namespaces by creation timestamp
	SortByAge NamespaceSortOption = "age"
	// SortByStatus sorts namespaces by status
	SortByStatus NamespaceSortOption = "status"
)
