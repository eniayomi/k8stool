package cli

import (
	"fmt"

	"github.com/eniayomi/k8stool/internal/k8s"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func getMetricsCmd() *cobra.Command {
	var namespace string

	cmd := &cobra.Command{
		Use:   "metrics [pod-name]",
		Short: "Show resource usage metrics for a pod",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			podName := args[0]

			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			metrics, err := client.GetPodMetrics(namespace, podName)
			if err != nil {
				return err
			}

			// Print pod-level metrics
			bold := color.New(color.Bold)
			bold.Println("Pod Metrics:")
			fmt.Printf("  Name:      %s\n", metrics.Name)
			fmt.Printf("  Namespace: %s\n", metrics.Namespace)
			fmt.Printf("  CPU:       %s\n", metrics.CPU)
			fmt.Printf("  Memory:    %s\n", metrics.Memory)

			// Print container-level metrics
			bold.Println("\nContainer Metrics:")
			for _, container := range metrics.Containers {
				fmt.Printf("\n  • %s:\n", container.Name)
				fmt.Printf("      CPU:    %s\n", container.CPU)
				fmt.Printf("      Memory: %s\n", container.Memory)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
	return cmd
}
