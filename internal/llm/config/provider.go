package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"k8stool/internal/llm/providers/openai"
)

// ProviderType represents the type of LLM provider
type ProviderType string

const (
	OpenAIProvider ProviderType = "openai"
)

// OpenAIOptions holds configuration options for OpenAI
type OpenAIOptions struct {
	APIKey string
	Model  string
	OrgID  string
}

// ConfigureOpenAI configures OpenAI provider with the given options
func ConfigureOpenAI(opts OpenAIOptions) error {
	if opts.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	// Validate model
	switch opts.Model {
	case "", "gpt-4":
		opts.Model = "gpt-4"
	case "gpt-3.5-turbo":
		// Valid model
	default:
		return fmt.Errorf("unsupported model: %s", opts.Model)
	}

	config := openai.Config{
		APIKey: opts.APIKey,
		Model:  opts.Model,
		OrgID:  opts.OrgID,
	}

	if err := saveConfig(config); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println("\nConfiguration saved successfully!")
	fmt.Printf("Provider: OpenAI\n")
	fmt.Printf("Model: %s\n", opts.Model)
	if opts.OrgID != "" {
		fmt.Printf("Organization ID: %s\n", opts.OrgID)
	}

	return nil
}

// ProviderInfo holds information about a provider
type ProviderInfo struct {
	Type    ProviderType
	Name    string
	Active  bool
	Model   string
	HasAuth bool
}

// ListProviders returns information about all supported providers
func ListProviders() ([]ProviderInfo, error) {
	providers := []ProviderInfo{
		{
			Type: OpenAIProvider,
			Name: "OpenAI",
		},
	}

	// Check if OpenAI is configured
	config, err := loadOpenAIConfig()
	if err == nil && config.APIKey != "" {
		// Update OpenAI provider info
		providers[0].Active = true
		providers[0].Model = config.Model
		providers[0].HasAuth = true
	}

	return providers, nil
}

// loadOpenAIConfig reads the OpenAI configuration from file
func loadOpenAIConfig() (openai.Config, error) {
	var config openai.Config

	configFile := fmt.Sprintf("%s/config.env", getConfigDir())
	data, err := os.ReadFile(configFile)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "OPENAI_API_KEY":
			config.APIKey = value
		case "OPENAI_MODEL":
			config.Model = value
		case "OPENAI_ORG_ID":
			config.OrgID = value
		}
	}

	return config, nil
}

// ConfigureProvider handles the interactive configuration of LLM providers
func ConfigureProvider() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Available LLM Providers:")
	fmt.Println("1. OpenAI")
	fmt.Print("\nSelect a provider (1): ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	if choice == "" || choice == "1" {
		return configureOpenAI(reader)
	}

	return fmt.Errorf("invalid choice: only OpenAI is supported at the moment")
}

func configureOpenAI(reader *bufio.Reader) error {
	fmt.Print("\nEnter your OpenAI API Key: ")
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)

	if apiKey == "" {
		return fmt.Errorf("API Key cannot be empty")
	}

	fmt.Print("Enter Organization ID (optional, press Enter to skip): ")
	orgID, _ := reader.ReadString('\n')
	orgID = strings.TrimSpace(orgID)

	fmt.Println("\nSelect Model:")
	fmt.Println("1. GPT-4 (Recommended)")
	fmt.Println("2. GPT-3.5-Turbo")
	fmt.Print("\nChoose a model (1): ")

	modelChoice, _ := reader.ReadString('\n')
	modelChoice = strings.TrimSpace(modelChoice)

	var model string
	switch modelChoice {
	case "", "1":
		model = "gpt-4"
	case "2":
		model = "gpt-3.5-turbo"
	default:
		fmt.Println("Invalid choice. Using GPT-4 as default.")
		model = "gpt-4"
	}

	return ConfigureOpenAI(OpenAIOptions{
		APIKey: apiKey,
		Model:  model,
		OrgID:  orgID,
	})
}

func saveConfig(config openai.Config) error {
	// Create config directory if it doesn't exist
	configDir := getConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Save the configuration
	configFile := fmt.Sprintf("%s/config.env", configDir)
	f, err := os.OpenFile(configFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	// Write configuration
	fmt.Fprintf(f, "OPENAI_API_KEY=%s\n", config.APIKey)
	fmt.Fprintf(f, "OPENAI_MODEL=%s\n", config.Model)
	if config.OrgID != "" {
		fmt.Fprintf(f, "OPENAI_ORG_ID=%s\n", config.OrgID)
	}

	return nil
}

func getConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".k8stool"
	}
	return fmt.Sprintf("%s/.k8stool", homeDir)
}

// LoadOpenAIConfig loads the OpenAI configuration from the config file
func LoadOpenAIConfig() (openai.Config, error) {
	return loadOpenAIConfig()
}
