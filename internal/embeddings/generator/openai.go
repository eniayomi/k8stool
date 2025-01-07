package generator

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// OpenAIGenerator implements embeddings.Generator using OpenAI's API
type OpenAIGenerator struct {
	client *openai.Client
	model  openai.EmbeddingModel
}

// NewOpenAIGenerator creates a new OpenAI-based embedding generator
func NewOpenAIGenerator(apiKey string) *OpenAIGenerator {
	return &OpenAIGenerator{
		client: openai.NewClient(apiKey),
		model:  openai.AdaEmbeddingV2,
	}
}

// Generate creates an embedding for the given text
func (g *OpenAIGenerator) Generate(text string) ([]float32, error) {
	resp, err := g.client.CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
		Input: []string{text},
		Model: g.model,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data received")
	}

	return resp.Data[0].Embedding, nil
}

// GenerateBatch creates embeddings for multiple texts
func (g *OpenAIGenerator) GenerateBatch(texts []string) ([][]float32, error) {
	resp, err := g.client.CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
		Input: texts,
		Model: g.model,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings: %w", err)
	}

	embeddings := make([][]float32, len(resp.Data))
	for i, data := range resp.Data {
		embeddings[i] = data.Embedding
	}

	return embeddings, nil
}
