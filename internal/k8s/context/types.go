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

// ClusterInfo contains information about the Kubernetes cluster
type ClusterInfo struct {
	Server                string
	CertificateAuthority  string
	InsecureSkipTLSVerify bool
	APIVersion            string
	MasterVersion         string
	Platform              string
}

// ContextSortOption represents context sorting options
type ContextSortOption string

const (
	// SortByName sorts contexts by name
	SortByName ContextSortOption = "name"
	// SortByCluster sorts contexts by cluster name
	SortByCluster ContextSortOption = "cluster"
	// SortByActive sorts contexts by active status (active first)
	SortByActive ContextSortOption = "active"
)
