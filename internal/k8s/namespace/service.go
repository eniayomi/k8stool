package namespace

import (
	"context"
	"fmt"
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type service struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

// newService creates a new namespace service instance
func newService(clientset *kubernetes.Clientset, config *rest.Config) Service {
	return &service{
		clientset: clientset,
		config:    config,
	}
}

// List returns all available namespaces
func (s *service) List() ([]Namespace, error) {
	namespaceList, err := s.clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	var namespaces []Namespace
	for _, ns := range namespaceList.Items {
		namespace := Namespace{
			Name:              ns.Name,
			Status:            string(ns.Status.Phase),
			CreationTimestamp: ns.CreationTimestamp.Time,
			Labels:            ns.Labels,
			Annotations:       ns.Annotations,
			Phase:             ns.Status.Phase,
		}
		namespaces = append(namespaces, namespace)
	}

	return namespaces, nil
}

// Get returns details for a specific namespace
func (s *service) Get(name string) (*NamespaceDetails, error) {
	ns, err := s.clientset.CoreV1().Namespaces().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace %q: %w", name, err)
	}

	quotas, err := s.GetResourceQuotas(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource quotas: %w", err)
	}

	limits, err := s.GetLimitRanges(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get limit ranges: %w", err)
	}

	details := &NamespaceDetails{
		Namespace: Namespace{
			Name:              ns.Name,
			Status:            string(ns.Status.Phase),
			CreationTimestamp: ns.CreationTimestamp.Time,
			Labels:            ns.Labels,
			Annotations:       ns.Annotations,
			Phase:             ns.Status.Phase,
		},
		ResourceQuotas: quotas,
		LimitRanges:    limits,
	}

	return details, nil
}

// Create creates a new namespace
func (s *service) Create(name string, labels, annotations map[string]string) error {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      labels,
			Annotations: annotations,
		},
	}

	_, err := s.clientset.CoreV1().Namespaces().Create(context.Background(), namespace, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create namespace %q: %w", name, err)
	}

	return nil
}

// Delete deletes a namespace
func (s *service) Delete(name string) error {
	err := s.clientset.CoreV1().Namespaces().Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete namespace %q: %w", name, err)
	}

	return nil
}

// GetResourceQuotas returns resource quotas for a namespace
func (s *service) GetResourceQuotas(namespace string) ([]ResourceQuota, error) {
	quotaList, err := s.clientset.CoreV1().ResourceQuotas(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list resource quotas: %w", err)
	}

	var quotas []ResourceQuota
	for _, quota := range quotaList.Items {
		resourceQuota := ResourceQuota{
			Name: quota.Name,
			Hard: make(ResourceList),
			Used: make(ResourceList),
		}

		for resource, quantity := range quota.Status.Hard {
			resourceQuota.Hard[string(resource)] = quantity.String()
		}

		for resource, quantity := range quota.Status.Used {
			resourceQuota.Used[string(resource)] = quantity.String()
		}

		for _, scope := range quota.Spec.Scopes {
			resourceQuota.Scopes = append(resourceQuota.Scopes, string(scope))
		}

		quotas = append(quotas, resourceQuota)
	}

	return quotas, nil
}

// GetLimitRanges returns limit ranges for a namespace
func (s *service) GetLimitRanges(namespace string) ([]LimitRange, error) {
	limitList, err := s.clientset.CoreV1().LimitRanges(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list limit ranges: %w", err)
	}

	var limits []LimitRange
	for _, limit := range limitList.Items {
		for _, item := range limit.Spec.Limits {
			limitRange := LimitRange{
				Name:    limit.Name,
				Type:    string(item.Type),
				Min:     make(ResourceList),
				Max:     make(ResourceList),
				Default: make(ResourceList),
			}

			for resource, quantity := range item.Min {
				limitRange.Min[string(resource)] = quantity.String()
			}

			for resource, quantity := range item.Max {
				limitRange.Max[string(resource)] = quantity.String()
			}

			for resource, quantity := range item.Default {
				limitRange.Default[string(resource)] = quantity.String()
			}

			limits = append(limits, limitRange)
		}
	}

	return limits, nil
}

// Sort sorts namespaces based on the given option
func (s *service) Sort(namespaces []Namespace, sortBy NamespaceSortOption) []Namespace {
	switch sortBy {
	case SortByName:
		sort.Slice(namespaces, func(i, j int) bool {
			return namespaces[i].Name < namespaces[j].Name
		})
	case SortByAge:
		sort.Slice(namespaces, func(i, j int) bool {
			return namespaces[i].CreationTimestamp.Before(namespaces[j].CreationTimestamp)
		})
	case SortByStatus:
		sort.Slice(namespaces, func(i, j int) bool {
			return namespaces[i].Status < namespaces[j].Status
		})
	}
	return namespaces
}
