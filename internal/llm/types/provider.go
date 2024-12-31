package types

import "context"

// Message represents a single message in a chat conversation
type Message struct {
	Role    string // system, user, assistant
	Content string
}

// CompletionOptions contains parameters for LLM completion requests
type CompletionOptions struct {
	Temperature      float32
	MaxTokens        int
	TopP             float32
	FrequencyPenalty float32
	PresencePenalty  float32
	Stop             []string
}

// CompletionChunk represents a chunk of streaming completion
type CompletionChunk struct {
	Content string
	Error   error
}

// LLMProvider defines the interface for language model interactions
type LLMProvider interface {
	// Complete generates a completion for a single prompt
	Complete(ctx context.Context, prompt string, opts CompletionOptions) (string, error)

	// CompleteChat generates a completion for a chat conversation
	CompleteChat(ctx context.Context, messages []Message, opts CompletionOptions) (string, error)

	// StreamComplete streams completion chunks for a single prompt
	StreamComplete(ctx context.Context, prompt string, opts CompletionOptions) (<-chan CompletionChunk, error)

	// StreamCompleteChat streams completion chunks for a chat conversation
	StreamCompleteChat(ctx context.Context, messages []Message, opts CompletionOptions) (<-chan CompletionChunk, error)
}
