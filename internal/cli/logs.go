package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8stool/internal/k8s"
)

func getLogsCmd() *cobra.Command {
	var namespace string

	cmd := &cobra.Command{
		Use:   "logs [pod-name]",
		Short: "Get logs from a pod",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			podName := args[0]
			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			logs, err := client.GetPodLogs(podName, namespace)
			if err != nil {
				return err
			}

			fmt.Print(logs)
			return nil
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
	return cmd
}
