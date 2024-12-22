package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/eniayomi/k8stool/internal/k8s"
)

func getNamespaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "namespace",
		Short: "List or switch namespaces",
		Long:  `List all namespaces or switch to a different namespace`,
	}

	cmd.AddCommand(listNamespacesCmd())
	cmd.AddCommand(switchNamespaceCmd())
	cmd.AddCommand(currentNamespaceCmd())

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
