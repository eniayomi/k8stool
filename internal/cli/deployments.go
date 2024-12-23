package cli

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"k8stool/internal/k8s"
	"k8stool/pkg/utils"

	"github.com/spf13/cobra"
)

func getDeploymentsCmd() *cobra.Command {
	var namespace string
	var allNamespaces bool
	var labelSelector string
	var sortBy string
	var reverse bool
	var showMetrics bool

	cmd := &cobra.Command{
		Use:     "deployments",
		Aliases: []string{"deployment", "deploy"},
		Short:   "List deployments",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			// If namespace flag not provided and not all namespaces, use the client's current namespace
			if !allNamespaces && namespace == "" {
				namespace = client.GetNamespace()
			}

			deployments, err := client.ListDeployments(namespace, allNamespaces, labelSelector)
			if err != nil {
				return err
			}

			// Sort deployments if requested
			if err := sortDeployments(deployments, sortBy, reverse); err != nil {
				return err
			}

			// If metrics flag is set, add metrics information
			if showMetrics {
				if err := client.AddDeploymentMetrics(deployments); err != nil {
					return fmt.Errorf("failed to get metrics: %v", err)
				}
			}

			return printDeployments(deployments, showMetrics)
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace")
	cmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "List deployments in all namespaces")
	cmd.Flags().StringVarP(&labelSelector, "selector", "l", "", "Label selector")
	cmd.Flags().StringVar(&sortBy, "sort", "", "Sort by (name, status, age)")
	cmd.Flags().BoolVar(&reverse, "reverse", false, "Reverse sort order")
	cmd.Flags().BoolVar(&showMetrics, "metrics", false, "Show resource metrics")

	return cmd
}

func sortDeployments(deployments []k8s.Deployment, sortBy string, reverse bool) error {
	switch sortBy {
	case "":
		return nil
	case "name":
		sort.Slice(deployments, func(i, j int) bool {
			if reverse {
				return deployments[i].Name > deployments[j].Name
			}
			return deployments[i].Name < deployments[j].Name
		})
	case "status":
		sort.Slice(deployments, func(i, j int) bool {
			if reverse {
				return deployments[i].Status > deployments[j].Status
			}
			return deployments[i].Status < deployments[j].Status
		})
	case "age":
		sort.Slice(deployments, func(i, j int) bool {
			if reverse {
				return deployments[i].Age > deployments[j].Age
			}
			return deployments[i].Age < deployments[j].Age
		})
	default:
		return fmt.Errorf("invalid sort key: %s", sortBy)
	}
	return nil
}

func printDeployments(deployments []k8s.Deployment, showMetrics bool) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Check if we need to show namespace column by checking if deployments are from different namespaces
	showNamespace := false
	if len(deployments) > 0 {
		ns := deployments[0].Namespace
		for _, deploy := range deployments[1:] {
			if deploy.Namespace != ns {
				showNamespace = true
				break
			}
		}
	}

	// Print header based on what columns we're showing
	if showNamespace {
		if showMetrics {
			fmt.Fprintln(w, "NAMESPACE\tNAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE\tSTATUS\tCPU\tMEMORY")
		} else {
			fmt.Fprintln(w, "NAMESPACE\tNAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE\tSTATUS")
		}
	} else {
		if showMetrics {
			fmt.Fprintln(w, "NAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE\tSTATUS\tCPU\tMEMORY")
		} else {
			fmt.Fprintln(w, "NAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE\tSTATUS")
		}
	}

	for _, d := range deployments {
		ready := fmt.Sprintf("%d/%d", d.ReadyReplicas, d.Replicas)
		age := utils.FormatDuration(d.Age)

		if showNamespace {
			if showMetrics {
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%s\t%s\t%s\t%s\n",
					d.Namespace, d.Name, ready, d.UpdatedReplicas,
					d.AvailableReplicas, age, d.Status,
					d.Metrics.CPU, d.Metrics.Memory)
			} else {
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%s\t%s\n",
					d.Namespace, d.Name, ready, d.UpdatedReplicas,
					d.AvailableReplicas, age, d.Status)
			}
		} else {
			if showMetrics {
				fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\t%s\t%s\t%s\n",
					d.Name, ready, d.UpdatedReplicas,
					d.AvailableReplicas, age, d.Status,
					d.Metrics.CPU, d.Metrics.Memory)
			} else {
				fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\t%s\n",
					d.Name, ready, d.UpdatedReplicas,
					d.AvailableReplicas, age, d.Status)
			}
		}
	}

	return nil
}
