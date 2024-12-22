package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/eniayomi/k8stool/internal/k8s"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func getPodsCmd() *cobra.Command {
	var namespace string

	cmd := &cobra.Command{
		Use:   "pods [namespace]",
		Short: "List all pods in the specified namespace",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				namespace = args[0]
			}

			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			pods, err := client.ListPods(namespace)
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)

			fmt.Fprintln(w, "NAME\tREADY\tRESTARTS\tCONTROLLER\tAGE\tSTATUS")

			for _, pod := range pods {
				fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\n",
					pod.Name,
					pod.Ready,
					pod.Restarts,
					pod.Controller,
					pod.Age,
					colorizeStatus(pod.Status))
			}

			w.Flush()
			return nil
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
	return cmd
}

func colorizeStatus(status string) string {
	switch status {
	case "Running":
		return color.GreenString(status)
	case "Pending":
		return color.YellowString(status)
	case "Succeeded":
		return color.HiGreenString(status)
	case "Failed", "Evicted":
		return color.RedString(status)
	case "CrashLoopBackOff":
		return color.HiRedString(status)
	case "Completed":
		return color.HiGreenString(status)
	case "Terminating":
		return color.HiYellowString(status)
	default:
		return color.WhiteString(status)
	}
}
