package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	k8s "k8stool/internal/k8s/client"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func getNamespaceCmd() *cobra.Command {
	var interactive bool
	var list bool

	cmd := &cobra.Command{
		Use:     "namespace [namespace]",
		Aliases: []string{"ns"},
		Short:   "Switch or view current namespace",
		Long: `Switch to a different namespace or view current namespace.
Examples:
  # Switch to a specific namespace
  k8stool namespace kube-system
  k8stool ns kube-system

  # List all namespaces
  k8stool ns --list
  k8stool ns -l

  # Interactive namespace selection
  k8stool ns -i
  k8stool ns --interactive`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			// List namespaces
			if list {
				namespaces, err := client.ListNamespaces()
				if err != nil {
					return err
				}
				currentNs, err := client.GetCurrentNamespace()
				if err != nil {
					return err
				}

				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				fmt.Fprintln(w, "NAMESPACE\tSTATUS\tCURRENT")
				for _, ns := range namespaces {
					current := ""
					if ns.Name == currentNs {
						current = "*"
					}
					fmt.Fprintf(w, "%s\t%s\t%s\n", ns.Name, ns.Status.Phase, current)
				}
				return w.Flush()
			}

			// Interactive mode
			if interactive {
				namespaces, err := client.ListNamespaces()
				if err != nil {
					return err
				}

				var options []string
				for _, ns := range namespaces {
					options = append(options, ns.Name)
				}

				prompt := promptui.Select{
					Label: "Select namespace",
					Items: options,
				}

				_, result, err := prompt.Run()
				if err != nil {
					return fmt.Errorf("prompt failed: %v", err)
				}

				if err := client.SwitchNamespace(result); err != nil {
					return err
				}

				fmt.Printf("Switched to namespace: %s\n", result)
				return nil
			}

			// If no args, show current namespace
			if len(args) == 0 {
				ns, err := client.GetCurrentNamespace()
				if err != nil {
					return err
				}
				fmt.Printf("Current namespace: %s\n", ns)
				return nil
			}

			// Switch namespace
			newNs := args[0]
			if err := client.SwitchNamespace(newNs); err != nil {
				return err
			}

			fmt.Printf("Switched to namespace: %s\n", newNs)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&list, "list", "l", false, "List all namespaces")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive namespace selection")

	return cmd
}
