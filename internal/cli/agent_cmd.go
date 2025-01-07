package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8stool/internal/embeddings/store"
	"k8stool/internal/learning"
	"k8stool/internal/llm/agent/k8s"
	"k8stool/internal/llm/config"

	"github.com/spf13/cobra"
)

// NewAgentCmd creates a new agent command
func NewAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent [query]",
		Short: "AI agent for Kubernetes operations",
		Long: `Interact with Kubernetes using natural language through an AI agent.
If no query is provided, starts an interactive chat session.

Examples:
  # Start interactive chat session
  k8stool agent

  # One-shot query
  k8stool agent "how many pods are running in default namespace?"

  # Configure OpenAI provider
  k8stool agent provider config

  # List available providers
  k8stool agent provider list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load OpenAI configuration
			cfg, err := config.LoadOpenAIConfig()
			if err != nil {
				return fmt.Errorf("failed to load OpenAI config: %w", err)
			}

			// Initialize stores
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}

			k8sToolDir := filepath.Join(homeDir, ".k8stool")
			embedStore := store.NewFileStore(cfg.APIKey)
			if err := embedStore.Load(filepath.Join(k8sToolDir, "embeddings.json")); err != nil {
				return fmt.Errorf("failed to load embeddings: %w", err)
			}

			learnStore, err := learning.New(filepath.Join(k8sToolDir, "learning.json"))
			if err != nil {
				return fmt.Errorf("failed to initialize learning store: %w", err)
			}

			// Create agent
			agent, err := k8s.NewAgent(embedStore, learnStore)
			if err != nil {
				return fmt.Errorf("failed to create agent: %w", err)
			}

			// If no args, start interactive mode
			if len(args) == 0 {
				fmt.Println("Starting chat with Kubernetes AI Agent (type 'exit' or 'quit' to end)")
				fmt.Println("------------------------------------------------------------")
				fmt.Printf("\nCurrent Context: %s\nCurrent Namespace: %s\n\n", agent.GetContext().CurrentContext, agent.GetContext().Namespace)

				reader := bufio.NewReader(os.Stdin)
				for {
					fmt.Print("> ")
					input, err := reader.ReadString('\n')
					if err != nil {
						return fmt.Errorf("failed to read input: %w", err)
					}

					input = strings.TrimSpace(input)
					if input == "" {
						continue
					}

					// Check for exit commands
					if input == "exit" || input == "quit" {
						fmt.Println("Goodbye!")
						return nil
					}

					// Process query
					response, err := agent.ProcessQuery(context.Background(), input)
					if err != nil {
						fmt.Printf("Error: %v\n", err)
						continue
					}

					fmt.Printf("\n%s\n\n", response)
				}
			}

			// One-shot mode: process single question
			query := strings.Join(args, " ")
			response, err := agent.ProcessQuery(context.Background(), query)
			if err != nil {
				return fmt.Errorf("failed to process query: %w", err)
			}

			fmt.Println(response)
			return nil
		},
	}

	// Add provider subcommand
	cmd.AddCommand(newProviderCmd())

	return cmd
}

// newProviderCmd creates a new provider command
func newProviderCmd() *cobra.Command {
	var (
		apiKey   string
		orgID    string
		model    string
		provider string
	)

	providerCmd := &cobra.Command{
		Use:   "provider",
		Short: "LLM provider management",
		Long:  `Commands for managing LLM providers like OpenAI.`,
	}

	// Add config subcommand
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configure LLM provider",
		Long: `Configure LLM provider settings. Can be used in interactive or non-interactive mode.

Interactive mode (no flags):
  k8stool agent provider config

Non-interactive mode (with flags):
  k8stool agent provider config --provider openai --api-key <key> --model gpt-4`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no flags are provided, use interactive mode
			if !cmd.Flags().Changed("provider") && !cmd.Flags().Changed("api-key") {
				return config.ConfigureProvider()
			}

			// Non-interactive mode
			if provider == "" {
				return fmt.Errorf("provider is required in non-interactive mode")
			}
			if apiKey == "" {
				return fmt.Errorf("api-key is required in non-interactive mode")
			}

			switch provider {
			case "openai":
				if model == "" {
					model = "gpt-4" // Default model
				}
				return config.ConfigureOpenAI(config.OpenAIOptions{
					APIKey: apiKey,
					Model:  model,
					OrgID:  orgID,
				})
			default:
				return fmt.Errorf("unsupported provider: %s", provider)
			}
		},
	}

	// Add flags for non-interactive mode
	configCmd.Flags().StringVar(&provider, "provider", "", "Provider to configure (e.g., openai)")
	configCmd.Flags().StringVar(&apiKey, "api-key", "", "API key for the provider")
	configCmd.Flags().StringVar(&orgID, "org-id", "", "Organization ID (optional)")
	configCmd.Flags().StringVar(&model, "model", "", "Model to use (e.g., gpt-4, gpt-3.5-turbo)")

	// Add list subcommand
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List available providers",
		Long:  `List all supported LLM providers and show which one is active.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			providers, err := config.ListProviders()
			if err != nil {
				return fmt.Errorf("failed to list providers: %w", err)
			}

			fmt.Println("Available LLM Providers:")
			fmt.Println("------------------------")

			for _, p := range providers {
				status := " "
				if p.Active {
					status = "*"
				}

				fmt.Printf("[%s] %s", status, p.Name)

				if p.Active {
					fmt.Printf(" (active, using %s)", p.Model)
				}

				if !p.HasAuth {
					fmt.Printf(" (not configured)")
				}

				fmt.Println()
			}

			return nil
		},
	}

	providerCmd.AddCommand(configCmd)
	providerCmd.AddCommand(listCmd)
	return providerCmd
}
