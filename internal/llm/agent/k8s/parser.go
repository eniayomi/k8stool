package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"k8stool/internal/learning"
	openaitypes "k8stool/internal/llm/providers/openai"
)

// ParseQuery parses a natural language query into task parameters
func (a *Agent) ParseQuery(ctx context.Context, query string, conversationContext string) (*TaskParams, error) {
	// Handle simple conversational queries directly
	if query == "hello" || query == "hi" || query == "hey" {
		return &TaskParams{
			Action:       "greet",
			ResourceType: "conversation",
		}, nil
	}

	// First, search for relevant documentation chunks to understand the query
	chunks, err := a.embedStore.Search(query, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to search documentation: %w", err)
	}

	// Build context from relevant chunks
	var docContext strings.Builder
	docContext.WriteString(conversationContext)
	docContext.WriteString("\n\n")
	for _, chunk := range chunks {
		docContext.WriteString(chunk.Content)
		docContext.WriteString("\n\n")
	}

	// Get completion from OpenAI to parse the query with context
	resp, err := a.client.CreateChatCompletion(ctx, openaitypes.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []openaitypes.ChatCompletionMessage{
			{Role: "system", Content: fmt.Sprintf(`You are an AI assistant for the k8stool command-line tool.
Based on the following conversation history and documentation:

%s

Parse the user's query into a structured task. The response must be a valid JSON object with these fields:
- action: The action to perform (e.g., get, list, describe, logs, exec, port-forward, help)
- resourceType: The type of resource (e.g., pod, deployment, namespace, events, metrics, help)
- resourceName: The name of the resource (if specified)
- namespace: The namespace (if specified)
- containerName: The container name for exec/logs operations (if specified)
- command: Array of command and arguments for exec operations (if specified)
- flags: Map of additional flags/parameters

Example responses:

For "show me the logs of the curl pod":
{
  "action": "logs",
  "resourceType": "pod",
  "resourceName": "curl"
}

For "what's the current namespace":
{
  "action": "get",
  "resourceType": "namespace"
}

For "show me pods in monitoring":
{
  "action": "list",
  "resourceType": "pod",
  "namespace": "monitoring"
}

For "show me failed pods":
{
  "action": "list",
  "resourceType": "pod",
  "flags": {
    "status": "Failed"
  }
}

For "is the curl pod running":
{
  "action": "list",
  "resourceType": "pod",
  "resourceName": "curl"
}

Remember to ONLY respond with a valid JSON object, nothing else.`, docContext.String())},
			{Role: "user", Content: query},
		},
		Temperature: 0.1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no completion choices returned")
	}

	// Parse the response
	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	var params TaskParams
	if err := json.Unmarshal([]byte(content), &params); err != nil {
		return nil, fmt.Errorf("failed to parse completion response: %w", err)
	}

	// Record this parsing interaction for learning
	interaction := learning.Interaction{
		Query:    query,
		Response: content,
		Context: map[string]string{
			"action":       params.Action,
			"resourceType": params.ResourceType,
		},
		Timestamp:  time.Now(),
		Successful: true,
	}
	if err := a.learnStore.RecordInteraction(interaction); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to record parsing interaction: %v\n", err)
	}

	return &params, nil
}
