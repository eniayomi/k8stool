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

func getPodsCmd() *cobra.Command {
	var allNamespaces bool
	var showMetrics bool
	var labelSelector string
	var statusFilter string
	var sortBy string
	var reverse bool
	var namespace string

	cmd := &cobra.Command{
		Use:     "pods",
		Aliases: []string{"pod", "po"},
		Short:   "List pods",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			// If namespace flag not provided, use the client's current namespace
			if !allNamespaces && namespace == "" {
				namespace = client.GetNamespace()
			}

			pods, err := client.ListPods(namespace, allNamespaces, labelSelector, statusFilter)
			if err != nil {
				return err
			}

			if showMetrics {
				if err := client.AddPodMetrics(pods); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Could not get metrics: %v\n", err)
				}
			}

			// Sort pods
			if err := sortPods(pods, sortBy, reverse); err != nil {
				return err
			}

			return printPods(pods, showMetrics)
		},
	}

	cmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "List pods in all namespaces")
	cmd.Flags().BoolVar(&showMetrics, "metrics", false, "Show CPU and Memory metrics")
	cmd.Flags().StringVarP(&labelSelector, "selector", "l", "", "Label selector")
	cmd.Flags().StringVarP(&statusFilter, "status", "s", "", "Filter by status")
	cmd.Flags().StringVar(&sortBy, "sort", "", "Sort by (name, status, age)")
	cmd.Flags().BoolVar(&reverse, "reverse", false, "Reverse sort order")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace")

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

	if showMetrics {
		if showNamespace {
			fmt.Fprintln(w, "NAMESPACE\tNAME\tREADY\tRESTARTS\tIP\tNODE\tCPU\tMEMORY\tAGE\tSTATUS")
		} else {
			fmt.Fprintln(w, "NAME\tREADY\tRESTARTS\tIP\tNODE\tCPU\tMEMORY\tAGE\tSTATUS")
		}
	} else {
		if showNamespace {
			fmt.Fprintln(w, "NAMESPACE\tNAME\tREADY\tRESTARTS\tIP\tNODE\tAGE\tSTATUS")
		} else {
			fmt.Fprintln(w, "NAME\tREADY\tRESTARTS\tIP\tNODE\tAGE\tSTATUS")
		}
	}

	for _, pod := range pods {
		age := utils.FormatDuration(pod.Age)

		if showMetrics {
			cpu := "<none>"
			mem := "<none>"
			if pod.Metrics != nil {
				cpu = pod.Metrics.CPU
				mem = pod.Metrics.Memory
			}
			if showNamespace {
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
					pod.Namespace, pod.Name, pod.Ready, pod.Restarts,
					pod.IP, pod.Node, cpu, mem, age, pod.Status)
			} else {
				fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
					pod.Name, pod.Ready, pod.Restarts,
					pod.IP, pod.Node, cpu, mem, age, pod.Status)
			}
		} else {
			if showNamespace {
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\t%s\t%s\n",
					pod.Namespace, pod.Name, pod.Ready, pod.Restarts,
					pod.IP, pod.Node, age, pod.Status)
			} else {
				fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\t%s\n",
					pod.Name, pod.Ready, pod.Restarts,
					pod.IP, pod.Node, age, pod.Status)
			}
		}
	}

	return nil
}
