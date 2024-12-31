package cli

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	k8s "k8stool/internal/k8s/client"
	"k8stool/internal/k8s/pods"
	"k8stool/pkg/utils"

	"github.com/spf13/cobra"
)

func getPodsCmd() *cobra.Command {
	var allNamespaces bool
	var selector string
	var sortBy string
	var reverse bool
	var showMetrics bool

	cmd := &cobra.Command{
		Use:     "pods",
		Aliases: []string{"pod", "po"},
		Short:   "Get pods",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			// Get namespace from flag or current context
			namespace := cmd.Flag("namespace").Value.String()
			if !allNamespaces && namespace == "" {
				currentCtx, err := client.ContextService.GetCurrent()
				if err != nil {
					return fmt.Errorf("failed to get current context: %v", err)
				}
				namespace = currentCtx.Namespace
			}

			// List pods using the service
			podList, err := client.PodService.List(namespace, allNamespaces, selector, "")
			if err != nil {
				return err
			}

			// Sort pods if requested
			if sortBy != "" {
				sort.Slice(podList, func(i, j int) bool {
					var result bool
					switch sortBy {
					case "name":
						result = podList[i].Name < podList[j].Name
					case "status":
						result = podList[i].Status < podList[j].Status
					case "age":
						result = podList[i].Age < podList[j].Age
					default:
						return false
					}
					if reverse {
						return !result
					}
					return result
				})
			}

			// Add metrics if requested
			if showMetrics {
				if err := client.PodService.AddMetrics(podList); err != nil {
					return fmt.Errorf("failed to get metrics: %v", err)
				}
			}

			return printPods(podList, showMetrics)
		},
	}

	cmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "List pods across all namespaces")
	cmd.Flags().StringVarP(&selector, "selector", "l", "", "Selector (label query) to filter on")
	cmd.Flags().StringVar(&sortBy, "sort", "", "Sort by (name, status, age)")
	cmd.Flags().BoolVar(&reverse, "reverse", false, "Reverse sort order")
	cmd.Flags().BoolVar(&showMetrics, "metrics", false, "Show resource metrics")

	return cmd
}

func printPods(pods []pods.Pod, showMetrics bool) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
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
			fmt.Fprintln(w, "NAME\tREADY\tRESTARTS\tIP\tNODE\tCPU\tMEMORY\tAGE\tSTATUS")
		} else {
			fmt.Fprintln(w, "NAME\tREADY\tRESTARTS\tIP\tNODE\tAGE\tSTATUS")
		}
	} else {
		if showMetrics {
			fmt.Fprintln(w, "NAME\tREADY\tRESTARTS\tIP\tNODE\tCPU\tMEMORY\tAGE\tSTATUS")
		} else {
			fmt.Fprintln(w, "NAME\tREADY\tRESTARTS\tIP\tNODE\tAGE\tSTATUS")
		}
	}

	for _, pod := range pods {
		ready := pod.Ready
		age := utils.FormatDuration(pod.Age)
		restartCount := fmt.Sprintf("%d", pod.Restarts)

		if showNamespace {
			if showMetrics && pod.Metrics != nil {
				cpu := "<none>"
				mem := "<none>"
				if pod.Metrics != nil {
					cpu = pod.Metrics.CPU
					mem = pod.Metrics.Memory
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
					pod.Name, ready,
					restartCount, pod.IP, pod.Node,
					cpu, mem, age,
					utils.ColorizeStatus(pod.Status))
			} else {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
					pod.Name, ready,
					restartCount, pod.IP, pod.Node,
					age, utils.ColorizeStatus(pod.Status))
			}
		} else {
			if showMetrics && pod.Metrics != nil {
				cpu := "<none>"
				mem := "<none>"
				if pod.Metrics != nil {
					cpu = pod.Metrics.CPU
					mem = pod.Metrics.Memory
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
					pod.Name, ready,
					restartCount, pod.IP, pod.Node,
					cpu, mem, age,
					utils.ColorizeStatus(pod.Status))
			} else {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
					pod.Name, ready,
					restartCount, pod.IP, pod.Node,
					age, utils.ColorizeStatus(pod.Status))
			}
		}
	}

	return nil
}
