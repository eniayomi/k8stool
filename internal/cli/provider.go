package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"k8stool/internal/llm/providers/openai"
)

// ConfigureProvider handles the interactive configuration of LLM providers
func ConfigureProvider() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Available LLM Providers:")
	fmt.Println("1. OpenAI")
	fmt.Print("\nSelect a provider (1): ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	if choice == "" || choice == "1" {
		configureOpenAI(reader)
	} else {
		fmt.Println("Invalid choice. Only OpenAI is supported at the moment.")
		os.Exit(1)
	}
}

func configureOpenAI(reader *bufio.Reader) {
	fmt.Print("\nEnter your OpenAI API Key: ")
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)

	if apiKey == "" {
		fmt.Println("API Key cannot be empty")
		os.Exit(1)
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

	config := openai.Config{
		APIKey: apiKey,
		Model:  model,
		OrgID:  orgID,
	}

	// Save the configuration
	err := saveConfig(config)
	if err != nil {
		fmt.Printf("Error saving configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nConfiguration saved successfully!")
	fmt.Printf("Provider: OpenAI\n")
	fmt.Printf("Model: %s\n", model)
	if orgID != "" {
		fmt.Printf("Organization ID: %s\n", orgID)
	}
}

func saveConfig(config openai.Config) error {
	// Create config directory if it doesn't exist
	configDir := getConfigDir()
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
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
