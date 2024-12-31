package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	k8s "k8stool/internal/k8s/client"
	"k8stool/internal/k8s/context"
	"k8stool/pkg/utils"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func getNamespaceCmd() *cobra.Command {
	var interactive bool

	cmd := &cobra.Command{
		Use:     "namespace [namespace_name]",
		Aliases: []string{"ns"},
		Short:   "Manage Kubernetes namespaces",
		Long:    "Manage Kubernetes namespaces, including switching between namespaces and viewing namespace information.",
		Args:    cobra.MaximumNArgs(1),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip cluster connection for namespace commands
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize context service without cluster access
			contextService, err := context.NewContextOnlyService()
			if err != nil {
				return fmt.Errorf("failed to initialize context service: %w", err)
			}

			// If interactive mode is requested or no args provided with -i flag
			if interactive {
				client, err := k8s.NewClient()
				if err != nil {
					return fmt.Errorf("failed to initialize client: %w", err)
				}

				namespaces, err := client.NamespaceService.List()
				if err != nil {
					return fmt.Errorf("failed to list namespaces: %w", err)
				}

				current, err := contextService.GetCurrent()
				if err != nil {
					return fmt.Errorf("failed to get current context: %w", err)
				}

				var options []string
				for _, ns := range namespaces {
					name := ns.Name
					if name == current.Namespace {
						name += " (current)"
					}
					options = append(options, name)
				}

				prompt := &promptui.Select{
					Label: "Select namespace:",
					Items: options,
					Size:  10,
					Templates: &promptui.SelectTemplates{
						Active:   "→ {{ . | cyan }}",
						Inactive: "  {{ . | white }}",
						Selected: "✓ {{ . | green }}",
					},
				}

				idx, _, err := prompt.Run()
				if err != nil {
					return fmt.Errorf("failed to get user input: %w", err)
				}

				// Extract namespace name from selected option
				targetNamespace := strings.TrimSuffix(options[idx], " (current)")

				if err := contextService.SetNamespace(targetNamespace); err != nil {
					return fmt.Errorf("failed to switch namespace: %w", err)
				}

				fmt.Printf("Switched to namespace %q\n", targetNamespace)
				return nil
			}

			// If a namespace is provided, switch to it
			if len(args) > 0 {
				targetNamespace := args[0]

				// Initialize client to validate namespace
				client, err := k8s.NewClient()
				if err != nil {
					return fmt.Errorf("failed to initialize client: %w", err)
				}

				// Validate namespace exists
				_, err = client.NamespaceService.Get(targetNamespace)
				if err != nil {
					return fmt.Errorf("namespaces %q not found", targetNamespace)
				}

				if err := contextService.SetNamespace(targetNamespace); err != nil {
					return fmt.Errorf("failed to switch namespace: %w", err)
				}

				fmt.Printf("Switched to namespace %q\n", targetNamespace)
				return nil
			}

			// If no args provided and not in interactive mode, show current namespace
			current, err := contextService.GetCurrent()
			if err != nil {
				return fmt.Errorf("failed to get current context: %w", err)
			}

			fmt.Printf("Current namespace: %s\n", current.Namespace)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive mode")

	// Add subcommands
	cmd.AddCommand(getCurrentNamespaceCmd())
	cmd.AddCommand(listNamespacesCmd())
	cmd.AddCommand(switchNamespaceCmd())

	return cmd
}

func getCurrentNamespaceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show current namespace",
		Long:  "Display information about the current Kubernetes namespace.",
		RunE: func(cmd *cobra.Command, args []string) error {
			contextService, err := context.NewContextOnlyService()
			if err != nil {
				return fmt.Errorf("failed to initialize context service: %w", err)
			}

			current, err := contextService.GetCurrent()
			if err != nil {
				return fmt.Errorf("failed to get current context: %w", err)
			}

			fmt.Printf("Current namespace: %s\n", current.Namespace)
			return nil
		},
	}
}

func listNamespacesCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List available namespaces",
		Long:    "Display a list of all available Kubernetes namespaces.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return fmt.Errorf("failed to initialize client: %w", err)
			}

			namespaces, err := client.NamespaceService.List()
			if err != nil {
				return fmt.Errorf("failed to list namespaces: %w", err)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tSTATUS")
			for _, ns := range namespaces {
				fmt.Fprintf(w, "%s\t%s\n",
					ns.Name,
					utils.ColorizeStatus(ns.Status))
			}
			w.Flush()

			return nil
		},
	}
}
func switchNamespaceCmd() *cobra.Command {
	var interactive bool

	cmd := &cobra.Command{
		Use:   "switch [namespace]",
		Short: "Switch to a different namespace",
		Long:  "Switch to a different Kubernetes namespace, either by name or interactively.",
		RunE: func(cmd *cobra.Command, args []string) error {
			contextService, err := context.NewContextOnlyService()
			if err != nil {
				return fmt.Errorf("failed to initialize context service: %w", err)
			}

			client, err := k8s.NewClient()
			if err != nil {
				return fmt.Errorf("failed to initialize client: %w", err)
			}

			if interactive || len(args) == 0 {
				namespaces, err := client.NamespaceService.List()
				if err != nil {
					return fmt.Errorf("failed to list namespaces: %w", err)
				}

				current, err := contextService.GetCurrent()
				if err != nil {
					return fmt.Errorf("failed to get current context: %w", err)
				}

				var options []string
				for _, ns := range namespaces {
					name := ns.Name
					if name == current.Namespace {
						name += " (current)"
					}
					options = append(options, name)
				}

				prompt := &promptui.Select{
					Label: "Select namespace:",
					Items: options,
					Size:  10,
					Templates: &promptui.SelectTemplates{
						Active:   "→ {{ . | cyan }}",
						Inactive: "  {{ . | white }}",
						Selected: "✓ {{ . | green }}",
					},
				}

				idx, _, err := prompt.Run()
				if err != nil {
					return fmt.Errorf("failed to get user input: %w", err)
				}

				// Extract namespace name from selected option
				targetNamespace := strings.TrimSuffix(options[idx], " (current)")

				if err := contextService.SetNamespace(targetNamespace); err != nil {
					return fmt.Errorf("failed to switch namespace: %w", err)
				}

				fmt.Printf("Switched to namespace %q\n", targetNamespace)
			} else {
				targetNamespace := args[0]

				// Validate namespace exists
				_, err := client.NamespaceService.Get(targetNamespace)
				if err != nil {
					return fmt.Errorf("namespaces %q not found", targetNamespace)
				}

				if err := contextService.SetNamespace(targetNamespace); err != nil {
					return fmt.Errorf("failed to switch namespace: %w", err)
				}

				fmt.Printf("Switched to namespace %q\n", targetNamespace)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Select namespace interactively")

	return cmd
}
