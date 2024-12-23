package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"k8stool/internal/k8s"

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

func listNamespacesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all namespaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			namespaces, current, err := client.GetNamespaces()
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.TabIndent)
			fmt.Fprintln(w, "CURRENT\tNAME\tSTATUS")

			for _, ns := range namespaces {
				currentMarker := " "
				if ns.Name == current {
					currentMarker = "*"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", currentMarker, ns.Name, ns.Status)
			}

			return w.Flush()
		},
	}
}

func switchNamespaceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "switch [namespace-name]",
		Short: "Switch to a different namespace",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			// If namespace is provided as argument, switch directly
			if len(args) == 1 {
				return client.SwitchNamespace(args[0])
			}

			// Otherwise, show interactive prompt
			namespaces, current, err := client.GetNamespaces()
			if err != nil {
				return err
			}

			var namespaceNames []string
			for _, ns := range namespaces {
				if ns.Name == current {
					namespaceNames = append(namespaceNames, fmt.Sprintf("%s (current)", ns.Name))
				} else {
					namespaceNames = append(namespaceNames, ns.Name)
				}
			}

			prompt := promptui.Select{
				Label: "Select Namespace",
				Items: namespaceNames,
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

			selectedNS := namespaces[idx].Name
			if selectedNS == current {
				fmt.Printf("Already using namespace %q\n", selectedNS)
				return nil
			}

			if err := client.SwitchNamespace(selectedNS); err != nil {
				return err
			}

			fmt.Printf("Switched to namespace %q\n", selectedNS)
			return nil
		},
	}
}

func currentNamespaceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show current namespace",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			_, current, err := client.GetNamespaces()
			if err != nil {
				return err
			}

			fmt.Printf("Current namespace: %s\n", current)
			return nil
		},
	}
}
