package agent

import (
	"context"
	"k8stool/internal/llm/types"
)

// Agent represents an AI agent that can understand and execute tasks
type Agent struct {
	llmProvider types.LLMProvider
	memory      Memory
	config      Config
}

// Config holds the configuration for the AI agent
type Config struct {
	MaxTokens    int
	Temperature  float32
	SystemPrompt string
	MemorySize   int
	MaxRetries   int
}

// Memory represents the agent's conversation history and context
type Memory struct {
	messages []types.Message
	maxSize  int
}

// Response represents the agent's response to a query
type Response struct {
	Content string
	Error   error
}

// Task represents a specific task for the agent to execute
type Task struct {
	Instruction string
	Context     map[string]interface{}
	Priority    int
}

// AgentInterface defines the methods that an AI agent must implement
type AgentInterface interface {
	// Process handles a user query and returns a response
	Process(ctx context.Context, input string) Response

	// ExecuteTask performs a specific task and returns the result
	ExecuteTask(ctx context.Context, task Task) Response

	// Learn updates the agent's knowledge or behavior based on feedback
	Learn(ctx context.Context, feedback string) error

	// Reset clears the agent's memory and state
	Reset() error
}
