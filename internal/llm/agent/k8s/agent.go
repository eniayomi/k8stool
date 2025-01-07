package k8s

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"k8stool/internal/embeddings"
	"k8stool/internal/embeddings/generator/openai"
	"k8stool/internal/learning"
	"k8stool/internal/llm/config"
	openaitypes "k8stool/internal/llm/providers/openai"

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

// Agent handles natural language queries about Kubernetes
type Agent struct {
	client     openaitypes.Client
	embedStore embeddings.EmbeddingStore
	learnStore *learning.LearningStore
	k8sClient  kubernetes.Interface
	k8sConfig  *rest.Config
	k8sContext *K8sContext
	validator  ResourceValidator
	currentCtx map[string]string
	// Add conversation memory
	conversationHistory []ConversationTurn
}

// ConversationTurn represents a single turn in the conversation
type ConversationTurn struct {
	Query     string
	Response  string
	Params    TaskParams
	Timestamp time.Time
}

// NewAgent creates a new Kubernetes agent
func NewAgent(embedStore embeddings.EmbeddingStore, learnStore *learning.LearningStore) (*Agent, error) {
	// Load OpenAI configuration
	cfg, err := config.LoadOpenAIConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAI config: %w", err)
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

	agent := &Agent{
		client:     openai.NewClient(cfg.APIKey),
		embedStore: embedStore,
		learnStore: learnStore,
		k8sClient:  clientset,
		k8sConfig:  config,
		k8sContext: k8sContext,
		currentCtx: make(map[string]string),
	}

	// Initialize validator
	agent.validator = NewBasicValidator(clientset)

	return agent, nil
}

// ProcessQuery handles a natural language query about Kubernetes
func (a *Agent) ProcessQuery(ctx context.Context, query string) (string, error) {
	// Add conversation context to the query
	var conversationContext strings.Builder
	if len(a.conversationHistory) > 0 {
		conversationContext.WriteString("Previous conversation:\n")
		// Use last 5 turns for context
		start := len(a.conversationHistory)
		if start > 5 {
			start = len(a.conversationHistory) - 5
		}
		for _, turn := range a.conversationHistory[start:] {
			conversationContext.WriteString(fmt.Sprintf("User: %s\nAssistant: %s\n", turn.Query, turn.Response))
		}
	}

	// First parse the query into task parameters with conversation context
	params, err := a.ParseQuery(ctx, query, conversationContext.String())
	if err != nil {
		return "", fmt.Errorf("failed to parse query: %w", err)
	}

	// Handle conversational queries
	switch params.ResourceType {
	case "conversation":
		switch params.Action {
		case "greet":
			response := fmt.Sprintf("Hello! I'm your Kubernetes AI assistant. You're currently in context %q and namespace %q. How can I help you?",
				a.k8sContext.CurrentContext,
				a.k8sContext.Namespace)

			// Record in conversation history
			a.conversationHistory = append(a.conversationHistory, ConversationTurn{
				Query:     query,
				Response:  response,
				Params:    *params,
				Timestamp: time.Now(),
			})

			return response, nil
		}
	case "help":
		if params.ResourceName == "" {
			response := "I can help you with Kubernetes operations. You can ask me about:\n" +
				"- Pods (list, describe, logs, exec)\n" +
				"- Deployments (list, describe, scale)\n" +
				"- Namespaces (list, switch)\n" +
				"- Contexts (list, switch)\n" +
				"- Events (get, watch)\n" +
				"- Metrics (pod and node usage)\n" +
				"- Port forwarding\n\n" +
				"Try asking in natural language, like:\n" +
				"- \"what pods are running?\"\n" +
				"- \"switch to production namespace\"\n"

			// Record in conversation history
			a.conversationHistory = append(a.conversationHistory, ConversationTurn{
				Query:     query,
				Response:  response,
				Params:    *params,
				Timestamp: time.Now(),
			})

			return response, nil
		}

		// Search for relevant documentation chunks
		searchQuery := query
		if params.ResourceName != "" {
			searchQuery = fmt.Sprintf("how to use %s command", params.ResourceName)
		}
		chunks, err := a.embedStore.Search(searchQuery, 3)
		if err != nil {
			return "", fmt.Errorf("failed to search documentation: %w", err)
		}

		// Apply learned relevance adjustments
		var chunkIDs []string
		var docContext strings.Builder
		for _, chunk := range chunks {
			chunkID := fmt.Sprintf("%s:%d-%d", chunk.Metadata.Source, chunk.Metadata.StartLine, chunk.Metadata.EndLine)
			chunkIDs = append(chunkIDs, chunkID)

			// Apply learned score adjustment
			score := a.learnStore.GetChunkScore(chunkID)
			if score > 1.2 { // Only include highly successful chunks
				docContext.WriteString(chunk.Content)
				docContext.WriteString("\n\n")
			}
		}

		// Get help response from OpenAI
		resp, err := a.client.CreateChatCompletion(ctx, openaitypes.ChatCompletionRequest{
			Model: "gpt-3.5-turbo",
			Messages: []openaitypes.ChatCompletionMessage{
				{Role: "system", Content: fmt.Sprintf(`You are an AI assistant for the k8stool command-line tool.
Based on the following documentation:

%s

Please help the user with their query. Be specific and provide command examples when relevant.`, docContext.String())},
				{Role: "user", Content: query},
			},
			Temperature: 0.2,
		})
		if err != nil {
			return "", fmt.Errorf("failed to get completion: %w", err)
		}

		response := resp.Choices[0].Message.Content

		// Record the interaction
		interaction := learning.Interaction{
			Query:      query,
			Response:   response,
			ChunksUsed: chunkIDs,
			Timestamp:  time.Now(),
			Context:    a.currentCtx,
		}
		interaction.Successful = true
		if err := a.learnStore.RecordInteraction(interaction); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to record interaction: %v\n", err)
		}

		// Record in conversation history
		a.conversationHistory = append(a.conversationHistory, ConversationTurn{
			Query:     query,
			Response:  response,
			Params:    *params,
			Timestamp: time.Now(),
		})

		return response, nil
	case "context":
		if params.Action == "get" {
			response := fmt.Sprintf("Current Context: %s\nNamespace: %s",
				a.k8sContext.CurrentContext,
				a.k8sContext.Namespace)

			// Record in conversation history
			a.conversationHistory = append(a.conversationHistory, ConversationTurn{
				Query:     query,
				Response:  response,
				Params:    *params,
				Timestamp: time.Now(),
			})

			return response, nil
		}
	}

	// If namespace is empty, use current namespace
	if params.Namespace == "" {
		params.Namespace = a.k8sContext.Namespace
	}

	// Handle the task based on resource type
	var result *TaskResult
	switch params.ResourceType {
	case "pod", "pods":
		result, err = a.PodHandler(ctx, *params)
	case "deployment", "deployments":
		result, err = a.DeploymentHandler(ctx, *params)
	case "namespace", "namespaces":
		result, err = a.NamespaceHandler(ctx, *params)
	case "event", "events":
		result, err = a.EventsHandler(ctx, *params)
	case "metrics":
		result, err = a.MetricsHandler(ctx, *params)
	case "portforward", "port-forward":
		result, err = a.PortForwardHandler(ctx, *params)
	case "exec":
		result, err = a.ExecHandler(ctx, *params)
	default:
		return "", fmt.Errorf("unsupported resource type: %s", params.ResourceType)
	}

	if err != nil {
		return "", err
	}

	// Record the interaction
	interaction := learning.Interaction{
		Query:      query,
		Response:   result.Output,
		Timestamp:  time.Now(),
		Context:    a.currentCtx,
		Successful: result.Success,
	}
	if err := a.learnStore.RecordInteraction(interaction); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to record interaction: %v\n", err)
	}

	// Record in conversation history
	a.conversationHistory = append(a.conversationHistory, ConversationTurn{
		Query:     query,
		Response:  result.Output,
		Params:    *params,
		Timestamp: time.Now(),
	})

	return result.Output, nil
}

// UpdateContext updates the current context (e.g., namespace, current command)
func (a *Agent) UpdateContext(key, value string) {
	a.currentCtx[key] = value
}

// GetContext returns the current Kubernetes context information
func (a *Agent) GetContext() *K8sContext {
	return a.k8sContext
}

// ValidateResource validates a Kubernetes resource
func (a *Agent) ValidateResource(ctx context.Context, resourceType, name, namespace string) error {
	return a.validator.ValidateResource(ctx, resourceType, name, namespace)
}

// ValidateOperation validates a Kubernetes operation
func (a *Agent) ValidateOperation(ctx context.Context, taskType TaskType, params map[string]interface{}) error {
	return a.validator.ValidateOperation(ctx, taskType, params)
}
