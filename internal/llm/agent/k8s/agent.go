package k8s

import (
	"context"
	"fmt"

	"k8stool/internal/llm/config"
	"k8stool/internal/llm/providers/openai"
	"k8stool/internal/llm/types"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// AgentConfig holds the agent configuration
type AgentConfig struct {
	MaxTokens   int
	Temperature float32
}

// DefaultConfig returns the default agent configuration
func DefaultConfig() AgentConfig {
	return AgentConfig{
		MaxTokens:   2000,
		Temperature: 0.7,
	}
}

// Agent represents an AI agent that can understand and execute Kubernetes operations
type Agent struct {
	llm        types.LLMProvider
	config     AgentConfig
	k8sClient  kubernetes.Interface
	k8sConfig  *rest.Config
	k8sContext *K8sContext
	validator  *BasicValidator
}

// NewAgent creates a new Kubernetes AI agent
func NewAgent() (*Agent, error) {
	// Load the OpenAI configuration
	cfg, err := config.LoadOpenAIConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAI config: %w", err)
	}

	// Create the OpenAI provider
	provider, err := openai.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI provider: %w", err)
	}

	// Load kubeconfig
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	// Get config
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get Kubernetes config: %w", err)
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Get current context info
	rawConfig, err := kubeConfig.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get raw config: %w", err)
	}

	currentContext := rawConfig.CurrentContext
	context := rawConfig.Contexts[currentContext]
	if context == nil {
		return nil, fmt.Errorf("current context %q not found", currentContext)
	}

	k8sContext := &K8sContext{
		CurrentContext: currentContext,
		Namespace:      context.Namespace,
		ClusterInfo:    context.Cluster,
	}

	// Create validator
	validator := NewBasicValidator(clientset)

	return &Agent{
		llm:        provider,
		config:     DefaultConfig(),
		k8sClient:  clientset,
		k8sConfig:  config,
		k8sContext: k8sContext,
		validator:  validator,
	}, nil
}

// ProcessQuery handles a natural language query about Kubernetes
func (a *Agent) ProcessQuery(ctx context.Context, query string) (string, error) {
	// Parse the query into task parameters
	params, err := a.ParseQuery(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to parse query: %w", err)
	}

	// Execute the task
	result, err := a.ExecuteTask(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to execute task: %w", err)
	}

	// Format the response
	systemPrompt := types.Message{
		Role: "system",
		Content: fmt.Sprintf(`You are a Kubernetes expert AI assistant. You have access to the following Kubernetes context:
- Current Context: %s
- Current Namespace: %s
- Current Cluster: %s

Your role is to:
1. Take the raw output from a Kubernetes operation
2. Format it in a clear, human-friendly way
3. Add any relevant explanations or suggestions
4. Highlight any potential issues or warnings

Always use the current context and namespace information provided above.`,
			a.k8sContext.CurrentContext,
			a.k8sContext.Namespace,
			a.k8sContext.ClusterInfo,
		),
	}

	// Create the task result message
	taskMsg := types.Message{
		Role:    "user",
		Content: fmt.Sprintf("Operation: %s %s\nResult:\n%s", params.ResourceType, params.Action, result.Output),
	}

	// Get formatted response
	opts := types.CompletionOptions{
		Temperature: a.config.Temperature,
		MaxTokens:   a.config.MaxTokens,
	}

	response, err := a.llm.CompleteChat(ctx, []types.Message{systemPrompt, taskMsg}, opts)
	if err != nil {
		return result.Output, nil // Fall back to raw output if formatting fails
	}

	return response, nil
}

// ValidateResource validates a Kubernetes resource
func (a *Agent) ValidateResource(ctx context.Context, resourceType, name, namespace string) error {
	return a.validator.ValidateResource(ctx, resourceType, name, namespace)
}

// ValidateOperation validates a Kubernetes operation
func (a *Agent) ValidateOperation(ctx context.Context, taskType TaskType, params map[string]interface{}) error {
	return a.validator.ValidateOperation(ctx, taskType, params)
}
