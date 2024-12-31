package describe

import (
	"context"
)

// DescribeService defines the interface for describing Kubernetes resources
type DescribeService interface {
	// DescribePod returns a detailed description of a pod
	DescribePod(ctx context.Context, namespace, name string) (*ResourceDescription, error)

	// DescribeDeployment returns a detailed description of a deployment
	DescribeDeployment(ctx context.Context, namespace, name string) (*ResourceDescription, error)

	// DescribeService returns a detailed description of a service
	DescribeService(ctx context.Context, namespace, name string) (*ResourceDescription, error)

	// DescribeNode returns a detailed description of a node
	DescribeNode(ctx context.Context, name string) (*ResourceDescription, error)

	// DescribeNamespace returns a detailed description of a namespace
	DescribeNamespace(ctx context.Context, name string) (*ResourceDescription, error)

	// Describe returns a detailed description of any supported resource
	Describe(ctx context.Context, resourceType ResourceType, namespace, name string) (*ResourceDescription, error)
}
