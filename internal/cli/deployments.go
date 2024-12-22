package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/eniayomi/k8stool/internal/k8s"
	"github.com/spf13/cobra"
)

func getDeploymentsCmd() *cobra.Command {
	var namespace string

	cmd := &cobra.Command{
		Use:   "deployments [namespace]",
		Short: "List all deployments in the specified namespace",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// If namespace is provided as an argument, use it
			if len(args) > 0 {
				namespace = args[0]
			}

			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			deployments, err := client.ListDeployments(namespace)
			if err != nil {
				return err
			}

			// Initialize tabwriter
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.TabIndent)

			// Print headers
			fmt.Fprintln(w, "NAME\tREPLICAS\tAVAILABLE\tREADY")

			// Print deployments
			for _, d := range deployments {
				fmt.Fprintf(w, "%s\t%d\t%d\t%s\n",
					d.Name,
					d.DesiredReplicas,
					d.AvailableReplicas,
					fmt.Sprintf("%d/%d", d.AvailableReplicas, d.DesiredReplicas))
			}

			w.Flush()
			return nil
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
	return cmd
}
