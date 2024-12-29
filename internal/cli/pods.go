package cli

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	k8s "k8stool/internal/k8s/client"
	"k8stool/pkg/utils"

	"github.com/spf13/cobra"
)

func getPodsCmd() *cobra.Command {
	var namespace string
	var allNamespaces bool
	var labelSelector string
	var fieldSelector string
	var sortBy string
	var reverse bool
	var showMetrics bool

	cmd := &cobra.Command{
		Use:     "pods",
		Aliases: []string{"pod", "po"},
		Short:   "List pods",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			// If namespace flag not provided and not all namespaces, use the client's current namespace
			if !allNamespaces && namespace == "" {
				namespace = client.GetNamespace()
			}

			pods, err := client.ListPods(namespace, allNamespaces, labelSelector, fieldSelector)
			if err != nil {
				return err
			}

			// Sort pods if requested
			if err := sortPods(pods, sortBy, reverse); err != nil {
				return err
			}

			// If metrics flag is set, add metrics information
			if showMetrics {
				if err := client.AddPodMetrics(pods); err != nil {
					return fmt.Errorf("failed to get metrics: %v", err)
				}
			}

			return printPods(pods, showMetrics)
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace")
	cmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "List pods in all namespaces")
	cmd.Flags().StringVarP(&labelSelector, "selector", "l", "", "Label selector")
	cmd.Flags().StringVarP(&fieldSelector, "field-selector", "f", "", "Field selector")
	cmd.Flags().StringVar(&sortBy, "sort", "", "Sort by (name, status, age)")
	cmd.Flags().BoolVar(&reverse, "reverse", false, "Reverse sort order")
	cmd.Flags().BoolVar(&showMetrics, "metrics", false, "Show resource metrics")

	return cmd
}

func sortPods(pods []k8s.Pod, sortBy string, reverse bool) error {
	switch sortBy {
	case "":
		return nil
	case "name":
		sort.Slice(pods, func(i, j int) bool {
			if reverse {
				return pods[i].Name > pods[j].Name
			}
			return pods[i].Name < pods[j].Name
		})
	case "status":
		sort.Slice(pods, func(i, j int) bool {
			if reverse {
				return pods[i].Status > pods[j].Status
			}
			return pods[i].Status < pods[j].Status
		})
	case "age":
		sort.Slice(pods, func(i, j int) bool {
			if reverse {
				return pods[i].Age > pods[j].Age
			}
			return pods[i].Age < pods[j].Age
		})
	default:
		return fmt.Errorf("invalid sort key: %s", sortBy)
	}
	return nil
}

func printPods(pods []k8s.Pod, showMetrics bool) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Check if we need to show namespace column by checking if pods are from different namespaces
	showNamespace := false
	if len(pods) > 0 {
		ns := pods[0].Namespace
		for _, pod := range pods[1:] {
			if pod.Namespace != ns {
				showNamespace = true
				break
			}
		}
	}

	// Print header based on what columns we're showing
	if showNamespace {
		if showMetrics {
			fmt.Fprintln(w, "NAMESPACE\tNAME\tREADY\tRESTARTS\tAGE\tCPU\tMEMORY\tSTATUS")
		} else {
			fmt.Fprintln(w, "NAMESPACE\tNAME\tREADY\tRESTARTS\tAGE\tSTATUS")
		}
	} else {
		if showMetrics {
			fmt.Fprintln(w, "NAME\tREADY\tRESTARTS\tAGE\tCPU\tMEMORY\tSTATUS")
		} else {
			fmt.Fprintln(w, "NAME\tREADY\tRESTARTS\tAGE\tSTATUS")
		}
	}

	for _, p := range pods {
		age := utils.FormatDuration(p.Age)

		if showNamespace {
			if showMetrics {
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\t%s\t%s\n",
					p.Namespace, p.Name, p.Ready, p.Restarts,
					age, p.Metrics.CPU, p.Metrics.Memory, p.Status)
			} else {
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n",
					p.Namespace, p.Name, p.Ready, p.Restarts, age, p.Status)
			}
		} else {
			if showMetrics {
				fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\t%s\n",
					p.Name, p.Ready, p.Restarts, age,
					p.Metrics.CPU, p.Metrics.Memory, p.Status)
			} else {
				fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n",
					p.Name, p.Ready, p.Restarts, age, p.Status)
			}
		}
	}

	return nil
}
