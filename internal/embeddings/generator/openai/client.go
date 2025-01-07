package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	openaitypes "k8stool/internal/llm/providers/openai"
)

// Client is a simple OpenAI API client
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new OpenAI client
func NewClient(apiKey string) openaitypes.Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// CreateChatCompletion sends a chat completion request to OpenAI
func (c *Client) CreateChatCompletion(ctx context.Context, req openaitypes.ChatCompletionRequest) (*openaitypes.ChatCompletionResponse, error) {
	// Convert request to JSON
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response
	var result openaitypes.ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
