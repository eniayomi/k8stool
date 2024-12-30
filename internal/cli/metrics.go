package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	k8s "k8stool/internal/k8s/client"
	"k8stool/internal/k8s/metrics"

	"github.com/spf13/cobra"
)

func getMetricsCmd() *cobra.Command {
	var namespace string
	var allNamespaces bool
	var selector string
	var sortBy string
	var reverse bool

	cmd := &cobra.Command{
		Use:     "metrics (pods|nodes|<pod-name>)",
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
				currentCtx, err := client.ContextService.GetCurrent()
				if err != nil {
					return err
				}
				namespace = currentCtx.Namespace
			}

			resourceType := args[0]
			switch resourceType {
			case "pods", "pod", "po":
				// List all pod metrics in the namespace
				podMetrics, err := client.MetricsService.ListPodMetrics(namespace)
				if err != nil {
					return err
				}

				// Sort metrics if requested
				if sortBy != "" {
					podMetrics = client.MetricsService.Sort(podMetrics, metrics.MetricsSortOption(sortBy))
					if reverse {
						// Reverse the slice
						for i, j := 0, len(podMetrics)-1; i < j; i, j = i+1, j-1 {
							podMetrics[i], podMetrics[j] = podMetrics[j], podMetrics[i]
						}
					}
				}

				return printPodMetricsList(podMetrics)

			case "nodes", "node", "no":
				// List all node metrics
				nodeMetrics, err := client.MetricsService.ListNodeMetrics()
				if err != nil {
					return err
				}
				return printNodeMetricsList(nodeMetrics)

			default:
				// Try to get metrics for a specific pod
				podMetrics, err := client.MetricsService.GetPodMetrics(namespace, resourceType)
				if err != nil {
					return fmt.Errorf("pod '%s' not found or error getting metrics: %v", resourceType, err)
				}

				return printPodMetrics(podMetrics)
			}
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace")
	cmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "Show metrics for all namespaces")
	cmd.Flags().StringVarP(&selector, "selector", "l", "", "Label selector")
	cmd.Flags().StringVar(&sortBy, "sort", "", "Sort by (name, cpu, memory, age)")
	cmd.Flags().BoolVar(&reverse, "reverse", false, "Reverse sort order")

	return cmd
}

func printPodMetrics(metrics *metrics.PodMetrics) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "NAMESPACE\tPOD\tCPU(cores)\tCPU%\tMEMORY(bytes)\tMEMORY%")
	fmt.Fprintf(w, "%s\t%s\t%d\t%.1f%%\t%d\t%.1f%%\n",
		metrics.Namespace,
		metrics.Name,
		metrics.TotalResources.CPU.UsageNanoCores/1e9, // Convert to cores
		metrics.TotalResources.CPU.UsageCorePercent,
		metrics.TotalResources.Memory.UsageBytes,
		metrics.TotalResources.Memory.LimitUtilization*100)

	return nil
}

func printPodMetricsList(metrics []metrics.PodMetrics) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "NAMESPACE\tPOD\tCPU(cores)\tCPU%\tMEMORY(bytes)\tMEMORY%")
	for _, m := range metrics {
		fmt.Fprintf(w, "%s\t%s\t%d\t%.1f%%\t%d\t%.1f%%\n",
			m.Namespace,
			m.Name,
			m.TotalResources.CPU.UsageNanoCores/1e9, // Convert to cores
			m.TotalResources.CPU.UsageCorePercent,
			m.TotalResources.Memory.UsageBytes,
			m.TotalResources.Memory.LimitUtilization*100)
	}

	return nil
}

func printNodeMetricsList(metrics []metrics.NodeMetrics) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "NODE\tCPU(cores)\tCPU%\tMEMORY(bytes)\tMEMORY%\tPODS")
	for _, m := range metrics {
		fmt.Fprintf(w, "%s\t%d\t%.1f%%\t%d\t%.1f%%\t%d\n",
			m.Name,
			m.Resources.CPU.UsageNanoCores/1e9, // Convert to cores
			m.Resources.CPU.UsageCorePercent,
			m.Resources.Memory.UsageBytes,
			m.Resources.Memory.LimitUtilization*100,
			m.PodCount)
	}

	return nil
}
