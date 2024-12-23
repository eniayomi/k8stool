package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"k8stool/internal/k8s"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func getContextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "context",
		Aliases: []string{"ctx"},
		Short:   "Show, list or switch Kubernetes contexts",
		Long:    `Show current context, list all available contexts in your kubeconfig, or switch to a different context`,
	}

	cmd.AddCommand(listContextCmd())
	cmd.AddCommand(currentContextCmd())
	cmd.AddCommand(switchContextCmd())

	return cmd
}

func listContextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available contexts",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			contexts, current, err := client.GetContexts()
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.TabIndent)
			fmt.Fprintln(w, "CURRENT\tNAME\tCLUSTER")

			for _, ctx := range contexts {
				currentMarker := " "
				if ctx.Name == current {
					currentMarker = "*"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", currentMarker, ctx.Name, ctx.Cluster)
			}

			return w.Flush()
		},
	}
}

func currentContextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show current context",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			current, err := client.GetCurrentContext()
			if err != nil {
				return err
			}

			fmt.Printf("Current context: %s\n", current)
			return nil
		},
	}
}

func switchContextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "switch [context-name]",
		Short: "Switch Kubernetes context",
		Long:  `Switch Kubernetes context either by providing the name or selecting interactively`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			// If context name is provided as argument, switch directly
			if len(args) == 1 {
				if err := client.SwitchContext(args[0]); err != nil {
					return err
				}
				fmt.Printf("Switched to context %q\n", args[0])
				return nil
			}

			// Otherwise, show interactive prompt
			contexts, current, err := client.GetContexts()
			if err != nil {
				return err
			}

			var contextNames []string
			for _, ctx := range contexts {
				if ctx.Name == current {
					contextNames = append(contextNames, fmt.Sprintf("%s (current)", ctx.Name))
				} else {
					contextNames = append(contextNames, ctx.Name)
				}
			}

			prompt := promptui.Select{
				Label: "Select Kubernetes Context",
				Items: contextNames,
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
				return fmt.Errorf("prompt failed: %v", err)
			}

			selectedContext := contexts[idx].Name
			if selectedContext == current {
				fmt.Printf("Already using context %q\n", selectedContext)
				return nil
			}

			if err := client.SwitchContext(selectedContext); err != nil {
				return err
			}

			fmt.Printf("Switched to context %q\n", selectedContext)
			return nil
		},
	}
}
