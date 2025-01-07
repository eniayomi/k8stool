package cli

import (
	"github.com/spf13/cobra"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Agent related commands",
	Long:  `Commands for managing the AI agent configuration and operation.`,
}

var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "Provider related commands",
	Long:  `Commands for managing LLM providers like OpenAI.`,
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure LLM provider",
	Long:  `Interactive configuration for LLM providers like OpenAI.`,
	Run: func(cmd *cobra.Command, args []string) {
		ConfigureProvider()
	},
}

func init() {
	// Add provider command to agent
	agentCmd.AddCommand(providerCmd)

	// Add config command to provider
	providerCmd.AddCommand(configCmd)
}
