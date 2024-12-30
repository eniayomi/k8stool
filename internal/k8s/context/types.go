package context

// Context represents a Kubernetes context configuration
type Context struct {
	Name        string
	Cluster     string
	User        string
	Namespace   string
	IsActive    bool
	ClusterInfo ClusterInfo
}

// ClusterInfo contains information about a Kubernetes cluster
type ClusterInfo struct {
	Version   string
	NodeCount int
}

// ContextSortOption defines how contexts should be sorted
type ContextSortOption int

const (
	// SortByName sorts contexts by name
	SortByName ContextSortOption = iota
	// SortByCluster sorts contexts by cluster name
	SortByCluster
	// SortByNamespace sorts contexts by namespace
	SortByNamespace
)
