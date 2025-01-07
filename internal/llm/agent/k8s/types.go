package k8s

import (
	"context"

	"k8s.io/client-go/kubernetes"
)

// TaskType represents different types of Kubernetes operations
type TaskType string

const (
	// Task types for different Kubernetes operations
	TaskPodInspect      TaskType = "pod_inspect"
	TaskPodLogs         TaskType = "pod_logs"
	TaskDeployInspect   TaskType = "deployment_inspect"
	TaskDeployScale     TaskType = "deployment_scale"
	TaskTroubleshoot    TaskType = "troubleshoot"
	TaskResourceApply   TaskType = "resource_apply"
	TaskResourceDelete  TaskType = "resource_delete"
	TaskContextSwitch   TaskType = "context_switch"
	TaskNamespaceSwitch TaskType = "namespace_switch"
	TaskContextList     TaskType = "context_list"
	TaskContextGet      TaskType = "context_get"
	TaskNamespaceList   TaskType = "namespace_list"
	TaskNamespaceGet    TaskType = "namespace_get"
	TaskGet             TaskType = "get"
	TaskList            TaskType = "list"
	TaskDescribe        TaskType = "describe"
	TaskLogs            TaskType = "logs"
	TaskExec            TaskType = "exec"
	TaskPortForward     TaskType = "port-forward"
)

// K8sContext holds information about the current Kubernetes context
type K8sContext struct {
	CurrentContext string
	Namespace      string
	ClusterInfo    string
}

// TaskHandler handles Kubernetes-specific operations
type TaskHandler struct {
	client    kubernetes.Interface
	context   K8sContext
	validator ResourceValidator
}

// ResourceValidator validates Kubernetes resources and operations
type ResourceValidator interface {
	ValidateResource(ctx context.Context, resourceType, name, namespace string) error
	ValidateOperation(ctx context.Context, taskType TaskType, params map[string]interface{}) error
}

// TaskResult represents the result of a Kubernetes operation
type TaskResult struct {
	Success       bool
	Output        string
	Error         error
	Suggestions   []string
	Resources     []string // Affected resources
	NoExplanation bool     // Skip explanation formatting
}

// TaskParams holds parameters for Kubernetes operations
type TaskParams struct {
	ResourceType  string
	ResourceName  string
	Namespace     string
	Action        string
	ExtraParams   map[string]interface{}
	ContainerName string
	Command       []string
	Flags         map[string]interface{}
}

// New creates a new Kubernetes task handler
func New(client kubernetes.Interface, context K8sContext) *TaskHandler {
	return &TaskHandler{
		client:  client,
		context: context,
	}
}

// SetValidator sets the resource validator for the task handler
func (h *TaskHandler) SetValidator(validator ResourceValidator) {
	h.validator = validator
}

// GetContext returns the current Kubernetes context
func (h *TaskHandler) GetContext() K8sContext {
	return h.context
}

// UpdateContext updates the Kubernetes context information
func (h *TaskHandler) UpdateContext(context K8sContext) {
	h.context = context
}
