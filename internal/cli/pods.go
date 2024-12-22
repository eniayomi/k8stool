package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"k8stool/internal/k8s"
	"k8stool/pkg/utils"

	"github.com/spf13/cobra"
)

func getPodsCmd() *cobra.Command {
	var namespace string
	var showMetrics bool
	var labelSelector string
	var statusFilter string
	var sortBy string
	var reverseSort bool
	var allNamespaces bool

	cmd := &cobra.Command{
		Use:   "pods [namespace]",
		Short: "List all pods in the specified namespace",
		Args:  cobra.MaximumNArgs(1),
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

			pods, err := client.ListPods(namespace, labelSelector)
			if err != nil {
				return err
			}

			if statusFilter != "" {
				var filteredPods []k8s.Pod
				for _, pod := range pods {
					if strings.EqualFold(pod.Status, statusFilter) {
						filteredPods = append(filteredPods, pod)
					}
				}
				pods = filteredPods
			}

			switch strings.ToLower(sortBy) {
			case "age":
				sort.Slice(pods, func(i, j int) bool {
					if reverseSort {
						return pods[i].Age < pods[j].Age
					}
					return pods[i].Age > pods[j].Age
				})
			case "status":
				sort.Slice(pods, func(i, j int) bool {
					if reverseSort {
						return pods[i].Status > pods[j].Status
					}
					return pods[i].Status < pods[j].Status
				})
			case "name":
				sort.Slice(pods, func(i, j int) bool {
					if reverseSort {
						return pods[i].Name > pods[j].Name
					}
					return pods[i].Name < pods[j].Name
				})
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)

			if allNamespaces {
				if showMetrics {
					fmt.Fprintln(w, "NAMESPACE\tNAME\tREADY\tRESTARTS\tCONTROLLER\tCPU\tMEMORY\tAGE\tSTATUS")
				} else {
					fmt.Fprintln(w, "NAMESPACE\tNAME\tREADY\tRESTARTS\tCONTROLLER\tAGE\tSTATUS")
				}
			} else {
				if showMetrics {
					fmt.Fprintln(w, "NAME\tREADY\tRESTARTS\tCONTROLLER\tCPU\tMEMORY\tAGE\tSTATUS")
				} else {
					fmt.Fprintln(w, "NAME\tREADY\tRESTARTS\tCONTROLLER\tAGE\tSTATUS")
				}
			}

			for _, pod := range pods {
				var metrics *k8s.PodMetrics
				if showMetrics {
					metrics, _ = client.GetPodMetrics(pod.Namespace, pod.Name)
				}

				if showMetrics {
					cpuUsage := "<none>"
					memUsage := "<none>"
					if metrics != nil {
						cpuUsage = metrics.CPU
						memUsage = metrics.Memory
					}
					if allNamespaces {
						fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\n",
							pod.Namespace,
							pod.Name,
							pod.Ready,
							pod.Restarts,
							pod.Controller,
							cpuUsage,
							memUsage,
							utils.FormatDuration(pod.Age),
							utils.ColorizeStatus(pod.Status))
					} else {
						fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\n",
							pod.Name,
							pod.Ready,
							pod.Restarts,
							pod.Controller,
							cpuUsage,
							memUsage,
							utils.FormatDuration(pod.Age),
							utils.ColorizeStatus(pod.Status))
					}
				} else {
					if allNamespaces {
						fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\t%s\n",
							pod.Namespace,
							pod.Name,
							pod.Ready,
							pod.Restarts,
							pod.Controller,
							utils.FormatDuration(pod.Age),
							utils.ColorizeStatus(pod.Status))
					} else {
						fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\n",
							pod.Name,
							pod.Ready,
							pod.Restarts,
							pod.Controller,
							utils.FormatDuration(pod.Age),
							utils.ColorizeStatus(pod.Status))
					}
				}
			}

			w.Flush()
			return nil
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
	cmd.Flags().BoolVar(&showMetrics, "metrics", false, "Show CPU and memory usage")
	cmd.Flags().StringVarP(&labelSelector, "selector", "l", "", "Label selector (e.g. 'app=nginx,env=prod')")
	cmd.Flags().StringVarP(&statusFilter, "status", "s", "", "Filter pods by status (e.g. Running, Pending, Failed)")
	cmd.Flags().StringVar(&sortBy, "sort", "", "Sort pods by field (age, status, name)")
	cmd.Flags().BoolVar(&reverseSort, "reverse", false, "Reverse the sort order")
	cmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "List pods across all namespaces")
	return cmd
}
