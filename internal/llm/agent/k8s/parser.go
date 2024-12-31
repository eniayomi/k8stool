package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"k8stool/internal/llm/types"
)

// ParseQuery parses a natural language query into task parameters
func (a *Agent) ParseQuery(ctx context.Context, query string) (*TaskParams, error) {
	// Create the system prompt
	systemPrompt := types.Message{
		Role: "system",
		Content: `You are the AI assistant for k8stool, a command-line tool for managing Kubernetes clusters. You understand both natural language queries and k8stool's command structure.

Command Structure:
k8stool [resource] [action] [name] [flags]

Available Resources:
- context (alias: ctx): Manage Kubernetes contexts
- namespace (alias: ns): Manage Kubernetes namespaces
- pod (alias: po): Manage Kubernetes pods
- deployment (alias: deploy): Manage Kubernetes deployments

Common Actions:
- get/show/display: Show details of a resource
- list/ls: List all resources
- switch/set/change: Change to a different context or namespace
- logs: Get logs from a pod
- scale: Scale a deployment

Your role is to convert natural language queries into structured task parameters.
You should identify:
1. The resource type (pod, deployment, namespace, context)
2. The action to perform (inspect, logs, list, scale, get, switch, set, change)
3. The resource name (if applicable)
4. Any additional parameters

For namespace and context queries:
- "What's the current namespace?" -> {"resource_type": "namespace", "action": "get"}
- "List all namespaces" -> {"resource_type": "namespace", "action": "list"}
- "What's the current context?" -> {"resource_type": "context", "action": "get"}
- "Show all contexts" -> {"resource_type": "context", "action": "list"}
- "Switch to context my-context" -> {"resource_type": "context", "action": "switch", "resource_name": "my-context"}
- "Change context to my-context" -> {"resource_type": "context", "action": "change", "resource_name": "my-context"}

For greetings, help requests, or invalid queries, respond with:
{
  "resource_type": "help",
  "action": "show",
  "resource_name": "",
  "namespace": "",
  "extra_params": {}
}

For all other queries, respond with a JSON object containing:
{
  "resource_type": "pod|deployment|namespace|context",
  "action": "inspect|logs|list|scale|get|switch|set|change",
  "resource_name": "name",
  "namespace": "namespace",
  "extra_params": {
    "key": "value"
  }
}

For example:
Query: "Show me the logs of the nginx pod"
Response: {
  "resource_type": "pod",
  "action": "logs",
  "resource_name": "nginx",
  "namespace": "default",
  "extra_params": {
    "tail": 100
  }
}

Important: Your response must be a valid JSON object. Do not include any additional text or explanation.`,
	}

	// Create the user query
	userQuery := types.Message{
		Role:    "user",
		Content: query,
	}

	// Get completion
	opts := types.CompletionOptions{
		Temperature: 0.1, // Low temperature for more deterministic results
		MaxTokens:   500,
	}

	response, err := a.llm.CompleteChat(ctx, []types.Message{systemPrompt, userQuery}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	// Debug: Print the raw response
	if strings.TrimSpace(response) == "" {
		return nil, fmt.Errorf("empty response from LLM")
	}

	// Try to clean the response by finding the first '{' and last '}'
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("invalid JSON response format: %s", response)
	}

	jsonStr := response[start : end+1]

	// Parse the JSON response
	var result struct {
		ResourceType string                 `json:"resource_type"`
		Action       string                 `json:"action"`
		ResourceName string                 `json:"resource_name"`
		Namespace    string                 `json:"namespace"`
		ExtraParams  map[string]interface{} `json:"extra_params"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w\nResponse: %s", err, response)
	}

	// Use current namespace if not specified
	if result.Namespace == "" {
		result.Namespace = a.k8sContext.Namespace
	}

	// Convert to TaskParams
	params := &TaskParams{
		ResourceType: result.ResourceType,
		ResourceName: result.ResourceName,
		Namespace:    result.Namespace,
		Action:       result.Action,
		ExtraParams:  result.ExtraParams,
	}

	return params, nil
}

// ExecuteTask executes a Kubernetes task based on the parsed parameters
func (a *Agent) ExecuteTask(ctx context.Context, params *TaskParams) (*TaskResult, error) {
	switch strings.ToLower(params.ResourceType) {
	case "pod":
		return a.PodHandler(ctx, *params)
	case "deployment":
		return a.DeploymentHandler(ctx, *params)
	case "namespace":
		return a.NamespaceHandler(ctx, *params)
	case "context":
		return a.ContextHandler(ctx, *params)
	case "help":
		return &TaskResult{
			Success: true,
			Output: `Welcome to k8stool! I can help you with Kubernetes operations. Here are some example commands:

Contexts:
- "What's my current context?"
- "List all contexts"
- "Switch to context my-context"

Namespaces:
- "What's my current namespace?"
- "List all namespaces"
- "Switch to namespace my-namespace"

Pods:
- "List all pods"
- "Show me the logs of pod my-pod"
- "Get details of pod my-pod"

Deployments:
- "List all deployments"
- "Scale deployment my-deployment to 3 replicas"
- "Get details of deployment my-deployment"

You can also use k8stool's command syntax directly:
k8stool [resource] [action] [name] [flags]

Resources: context (ctx), namespace (ns), pod (po), deployment (deploy)
Actions: get, list (ls), switch, logs, scale

Feel free to ask any questions about your Kubernetes cluster!`,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", params.ResourceType)
	}
}
