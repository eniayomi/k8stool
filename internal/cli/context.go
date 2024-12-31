package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"k8stool/internal/k8s/context"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func getContextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "context",
		Aliases: []string{"ctx"},
		Short:   "Manage Kubernetes contexts",
		Long:    "Manage Kubernetes contexts, including switching between contexts and viewing context information.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip cluster connection for context commands
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize context service without cluster access
			contextService, err := context.NewContextOnlyService()
			if err != nil {
				return fmt.Errorf("failed to initialize context service: %w", err)
			}

			current, err := contextService.GetCurrent()
			if err != nil {
				return fmt.Errorf("failed to get current context: %w", err)
			}

			fmt.Printf("Current context: %s\n", current.Name)
			return nil
		},
	}

	// Add subcommands
	cmd.AddCommand(getCurrentContextCmd())
	cmd.AddCommand(listContextsCmd())
	cmd.AddCommand(switchContextCmd())

	return cmd
}

func getCurrentContextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show current context",
		Long:  "Display information about the current Kubernetes context.",
		RunE: func(cmd *cobra.Command, args []string) error {
			contextService, err := context.NewContextOnlyService()
			if err != nil {
				return fmt.Errorf("failed to initialize context service: %w", err)
			}

			current, err := contextService.GetCurrent()
			if err != nil {
				return fmt.Errorf("failed to get current context: %w", err)
			}

			fmt.Printf("Current context: %s\n", current.Name)
			fmt.Printf("Cluster: %s\n", current.Cluster)
			fmt.Printf("User: %s\n", current.User)
			if current.Namespace != "" {
				fmt.Printf("Namespace: %s\n", current.Namespace)
			}

			return nil
		},
	}
}

func listContextsCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List available contexts",
		Long:    "Display a list of all available Kubernetes contexts.",
		RunE: func(cmd *cobra.Command, args []string) error {
			contextService, err := context.NewContextOnlyService()
			if err != nil {
				return fmt.Errorf("failed to initialize context service: %w", err)
			}

			contexts, err := contextService.List()
			if err != nil {
				return fmt.Errorf("failed to list contexts: %w", err)
			}

			// Sort contexts by name
			contexts = contextService.Sort(contexts, context.SortByName)

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tCLUSTER\tUSER\tNAMESPACE\tACTIVE")
			for _, ctx := range contexts {
				active := ""
				if ctx.IsActive {
					active = "*"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					ctx.Name,
					ctx.Cluster,
					ctx.User,
					ctx.Namespace,
					active)
			}
			w.Flush()

			return nil
		},
	}
}

func switchContextCmd() *cobra.Command {
	var interactive bool

	cmd := &cobra.Command{
		Use:   "switch [context]",
		Short: "Switch to a different context",
		Long:  "Switch to a different Kubernetes context, either by name or interactively.",
		RunE: func(cmd *cobra.Command, args []string) error {
			contextService, err := context.NewContextOnlyService()
			if err != nil {
				return fmt.Errorf("failed to initialize context service: %w", err)
			}

			contexts, err := contextService.List()
			if err != nil {
				return fmt.Errorf("failed to list contexts: %w", err)
			}

			var targetContext string

			if interactive || len(args) == 0 {
				// Sort contexts by name for consistent ordering
				contexts = contextService.Sort(contexts, context.SortByName)

				var options []string
				for _, ctx := range contexts {
					name := ctx.Name
					if ctx.IsActive {
						name += " (current)"
					}
					options = append(options, name)
				}

				prompt := &promptui.Select{
					Label: "Select context:",
					Items: options,
					Size:  10,
					Templates: &promptui.SelectTemplates{
						Label:    "{{ . }}",
						Active:   "\U0001F449 {{ . | cyan }}",
						Inactive: "  {{ . | white }}",
						Selected: "\U0001F44D {{ . | green }}",
					},
				}

				idx, _, err := prompt.Run()
				if err != nil {
					return fmt.Errorf("failed to get user input: %v", err)
				}

				// Extract context name from selected option
				targetContext = strings.TrimSuffix(options[idx], " (current)")
			} else {
				targetContext = args[0]
			}

			if err := contextService.SwitchContext(targetContext); err != nil {
				return fmt.Errorf("failed to switch context: %w", err)
			}

			fmt.Printf("Switched to context %q\n", targetContext)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Select context interactively")

	return cmd
}
