package openai

import (
	"context"
	"fmt"

	"k8stool/internal/llm/types"

	"github.com/sashabaranov/go-openai"
)

// Config holds OpenAI-specific configuration
type Config struct {
	APIKey string
	Model  string
	OrgID  string
}

// Provider implements the LLMProvider interface for OpenAI
type Provider struct {
	client *openai.Client
	config Config
}

// New creates a new OpenAI provider
func New(config Config) (*Provider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	if config.Model == "" {
		config.Model = openai.GPT4 // Default to GPT-4
	}

	// Create client config without org ID first
	clientConfig := openai.DefaultConfig(config.APIKey)

	// Only set org ID if provided
	if config.OrgID != "" {
		clientConfig.OrgID = config.OrgID
	} else {
		fmt.Printf("Debug: No Organization ID provided\n")
	}

	// Test the configuration with a simple API call
	client := openai.NewClientWithConfig(clientConfig)
	_, err := client.ListModels(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to validate OpenAI configuration: %w", err)
	}

	return &Provider{
		client: client,
		config: config,
	}, nil
}

// Complete implements single prompt completion
func (p *Provider) Complete(ctx context.Context, prompt string, opts types.CompletionOptions) (string, error) {
	req := openai.CompletionRequest{
		Model:            p.config.Model,
		Prompt:           prompt,
		Temperature:      float32(opts.Temperature),
		MaxTokens:        opts.MaxTokens,
		TopP:             float32(opts.TopP),
		FrequencyPenalty: float32(opts.FrequencyPenalty),
		PresencePenalty:  float32(opts.PresencePenalty),
		Stop:             opts.Stop,
	}

	resp, err := p.client.CreateCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("OpenAI completion error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no completion choices returned")
	}

	return resp.Choices[0].Text, nil
}

// CompleteChat implements chat completion
func (p *Provider) CompleteChat(ctx context.Context, messages []types.Message, opts types.CompletionOptions) (string, error) {
	var chatMessages []openai.ChatCompletionMessage
	for _, msg := range messages {
		chatMessages = append(chatMessages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	req := openai.ChatCompletionRequest{
		Model:            p.config.Model,
		Messages:         chatMessages,
		Temperature:      float32(opts.Temperature),
		MaxTokens:        opts.MaxTokens,
		TopP:             float32(opts.TopP),
		FrequencyPenalty: float32(opts.FrequencyPenalty),
		PresencePenalty:  float32(opts.PresencePenalty),
		Stop:             opts.Stop,
	}

	resp, err := p.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("OpenAI chat completion error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no chat completion choices returned")
	}

	return resp.Choices[0].Message.Content, nil
}

// StreamComplete implements streaming completion
func (p *Provider) StreamComplete(ctx context.Context, prompt string, opts types.CompletionOptions) (<-chan types.CompletionChunk, error) {
	stream := make(chan types.CompletionChunk)

	req := openai.CompletionRequest{
		Model:            p.config.Model,
		Prompt:           prompt,
		Temperature:      float32(opts.Temperature),
		MaxTokens:        opts.MaxTokens,
		TopP:             float32(opts.TopP),
		FrequencyPenalty: float32(opts.FrequencyPenalty),
		PresencePenalty:  float32(opts.PresencePenalty),
		Stop:             opts.Stop,
		Stream:           true,
	}

	go func() {
		defer close(stream)

		streamResp, err := p.client.CreateCompletionStream(ctx, req)
		if err != nil {
			stream <- types.CompletionChunk{Error: fmt.Errorf("create stream error: %w", err)}
			return
		}
		defer streamResp.Close()

		for {
			resp, err := streamResp.Recv()
			if err != nil {
				stream <- types.CompletionChunk{Error: fmt.Errorf("stream receive error: %w", err)}
				return
			}

			if len(resp.Choices) > 0 {
				stream <- types.CompletionChunk{Content: resp.Choices[0].Text}
			}
		}
	}()

	return stream, nil
}

// StreamCompleteChat implements streaming chat completion
func (p *Provider) StreamCompleteChat(ctx context.Context, messages []types.Message, opts types.CompletionOptions) (<-chan types.CompletionChunk, error) {
	stream := make(chan types.CompletionChunk)

	var chatMessages []openai.ChatCompletionMessage
	for _, msg := range messages {
		chatMessages = append(chatMessages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	req := openai.ChatCompletionRequest{
		Model:            p.config.Model,
		Messages:         chatMessages,
		Temperature:      float32(opts.Temperature),
		MaxTokens:        opts.MaxTokens,
		TopP:             float32(opts.TopP),
		FrequencyPenalty: float32(opts.FrequencyPenalty),
		PresencePenalty:  float32(opts.PresencePenalty),
		Stop:             opts.Stop,
		Stream:           true,
	}

	go func() {
		defer close(stream)

		streamResp, err := p.client.CreateChatCompletionStream(ctx, req)
		if err != nil {
			stream <- types.CompletionChunk{Error: fmt.Errorf("create chat stream error: %w", err)}
			return
		}
		defer streamResp.Close()

		for {
			resp, err := streamResp.Recv()
			if err != nil {
				stream <- types.CompletionChunk{Error: fmt.Errorf("chat stream receive error: %w", err)}
				return
			}

			if len(resp.Choices) > 0 {
				stream <- types.CompletionChunk{Content: resp.Choices[0].Delta.Content}
			}
		}
	}()

	return stream, nil
}
