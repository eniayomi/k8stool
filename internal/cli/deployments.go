package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"k8stool/internal/k8s"
	"k8stool/pkg/utils"

	"github.com/spf13/cobra"
)

func getDeploymentsCmd() *cobra.Command {
	var namespace string
	var allNamespaces bool

	cmd := &cobra.Command{
		Use:     "deployments [namespace]",
		Aliases: []string{"deploy"},
		Short:   "List all deployments in the specified namespace",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			if len(args) > 0 {
				namespace = args[0]
			}

			if allNamespaces {
				namespace = ""
			}

			deployments, err := client.ListDeployments(namespace)
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)

			if allNamespaces {
				fmt.Fprintln(w, "NAMESPACE\tNAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE")
			} else {
				fmt.Fprintln(w, "NAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE")
			}

			for _, d := range deployments {
				if allNamespaces {
					fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%s\n",
						d.Namespace,
						d.Name,
						d.Ready,
						d.UpToDate,
						d.Available,
						utils.FormatDuration(d.Age))
				} else {
					fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\n",
						d.Name,
						d.Ready,
						d.UpToDate,
						d.Available,
						utils.FormatDuration(d.Age))
				}
			}

			w.Flush()
			return nil
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
	cmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "List deployments across all namespaces")
	return cmd
}
