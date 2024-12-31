package context

import (
	"context"
	"fmt"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type service struct {
	clientset  *kubernetes.Clientset
	config     *rest.Config
	kubeconfig clientcmd.ClientConfig
}

// newService creates a new context service instance
func newService(clientset *kubernetes.Clientset, config *rest.Config, kubeconfig clientcmd.ClientConfig) Service {
	return &service{
		clientset:  clientset,
		config:     config,
		kubeconfig: kubeconfig,
	}
}

// NewContextOnlyService creates a new context service instance without requiring cluster access
func NewContextOnlyService() (Service, error) {
	// Load kubeconfig
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	return &service{
		kubeconfig: kubeConfig,
	}, nil
}

// List returns all available contexts
func (s *service) List() ([]Context, error) {
	rawConfig, err := s.kubeconfig.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	currentContext := rawConfig.CurrentContext
	var contexts []Context

	for name, ctx := range rawConfig.Contexts {
		context := Context{
			Name:      name,
			Cluster:   ctx.Cluster,
			User:      ctx.AuthInfo,
			Namespace: ctx.Namespace,
			IsActive:  name == currentContext,
		}

		contexts = append(contexts, context)
	}

	return contexts, nil
}

// GetCurrent returns the current context
func (s *service) GetCurrent() (*Context, error) {
	rawConfig, err := s.kubeconfig.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	currentContext := rawConfig.CurrentContext
	ctx, exists := rawConfig.Contexts[currentContext]
	if !exists {
		return nil, fmt.Errorf("current context %q not found", currentContext)
	}

	context := &Context{
		Name:      currentContext,
		Cluster:   ctx.Cluster,
		User:      ctx.AuthInfo,
		Namespace: ctx.Namespace,
		IsActive:  true,
	}

	return context, nil
}

// SwitchContext switches to a different context
func (s *service) SwitchContext(name string) error {
	configAccess := clientcmd.NewDefaultPathOptions()
	config, err := configAccess.GetStartingConfig()
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	if _, exists := config.Contexts[name]; !exists {
		return fmt.Errorf("context %q not found", name)
	}

	config.CurrentContext = name

	if err := clientcmd.ModifyConfig(configAccess, *config, true); err != nil {
		return fmt.Errorf("failed to modify kubeconfig: %w", err)
	}

	return nil
}

// SetNamespace sets the default namespace for the current context
func (s *service) SetNamespace(namespace string) error {
	configAccess := clientcmd.NewDefaultPathOptions()
	config, err := configAccess.GetStartingConfig()
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	currentContext := config.CurrentContext
	if currentContext == "" {
		return fmt.Errorf("no current context")
	}

	ctx, exists := config.Contexts[currentContext]
	if !exists {
		return fmt.Errorf("current context %q not found", currentContext)
	}

	ctx.Namespace = namespace

	if err := clientcmd.ModifyConfig(configAccess, *config, true); err != nil {
		return fmt.Errorf("failed to modify kubeconfig: %w", err)
	}

	return nil
}

// GetClusterInfo returns information about the current cluster
func (s *service) GetClusterInfo() (*ClusterInfo, error) {
	if s.clientset == nil {
		return nil, nil
	}

	version, err := s.clientset.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get server version: %w", err)
	}

	nodes, err := s.clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	return &ClusterInfo{
		Version:   version.String(),
		NodeCount: len(nodes.Items),
	}, nil
}

// Sort sorts contexts based on the given option
func (s *service) Sort(contexts []Context, sortBy ContextSortOption) []Context {
	switch sortBy {
	case SortByName:
		sort.Slice(contexts, func(i, j int) bool {
			return contexts[i].Name < contexts[j].Name
		})
	case SortByCluster:
		sort.Slice(contexts, func(i, j int) bool {
			return contexts[i].Cluster < contexts[j].Cluster
		})
	case SortByNamespace:
		sort.Slice(contexts, func(i, j int) bool {
			return contexts[i].Namespace < contexts[j].Namespace
		})
	}
	return contexts
}
