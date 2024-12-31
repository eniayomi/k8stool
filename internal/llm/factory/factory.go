package factory

import (
	"fmt"

	"k8stool/internal/llm/providers/openai"
	"k8stool/internal/llm/types"
)

// Factory creates LLM providers
type Factory struct{}

// New creates a new Factory
func New() *Factory {
	return &Factory{}
}

// CreateProvider creates a new LLM provider based on the provider type
func (f *Factory) CreateProvider(providerType string, config interface{}) (types.LLMProvider, error) {
	switch providerType {
	case "openai":
		cfg, ok := config.(openai.Config)
		if !ok {
			return nil, fmt.Errorf("invalid config type for OpenAI provider")
		}
		return openai.New(cfg)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
}
