package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"k8stool/internal/k8s"

	"github.com/spf13/cobra"
)

func getMetricsCmd() *cobra.Command {
	var namespace string
	var allNamespaces bool
	var selector string

	cmd := &cobra.Command{
		Use:     "metrics (pods|nodes)",
		Aliases: []string{"top"},
		Short:   "Show metrics for pods or nodes",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			// If namespace flag not provided and not all namespaces, use the client's current namespace
			if !allNamespaces && namespace == "" {
				namespace = client.GetNamespace()
			}

			resourceType := args[0]
			switch resourceType {
			case "pods", "pod", "po":
				metrics, err := client.GetPodMetrics(namespace, selector)
				if err != nil {
					return err
				}
				return printPodMetrics(metrics)
			case "nodes", "node", "no":
				return client.GetNodeMetrics(selector)
			default:
				return fmt.Errorf("unsupported resource type: %s", resourceType)
			}
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace")
	cmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "Show metrics for all namespaces")
	cmd.Flags().StringVarP(&selector, "selector", "l", "", "Label selector")

	return cmd
}

func printPodMetrics(metrics *k8s.PodMetrics) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "NAMESPACE\tPOD\tCPU\tMEMORY")
	// Add implementation for printing metrics
	return nil
}
