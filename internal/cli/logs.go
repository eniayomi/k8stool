package cli

import (
	"fmt"
	"os"
	"time"

	"k8stool/internal/k8s"

	"github.com/spf13/cobra"
)

func getLogsCmd() *cobra.Command {
	var namespace string
	var follow bool
	var previous bool
	var tail int64
	var container string
	var since string
	var sinceTime string

	cmd := &cobra.Command{
		Use:   "logs [pod-name]",
		Short: "View logs from a pod's containers",
		Long: `View logs from a pod's containers.
Example: k8stool logs nginx-pod
         k8stool logs nginx-pod -f`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			podName := args[0]

			// Parse time filters
			var sinceSeconds *int64
			var startTime *time.Time

			if since != "" {
				duration, err := time.ParseDuration(since)
				if err != nil {
					return fmt.Errorf("invalid duration: %v", err)
				}
				seconds := int64(duration.Seconds())
				sinceSeconds = &seconds
			}

			if sinceTime != "" {
				t, err := time.Parse(time.RFC3339, sinceTime)
				if err != nil {
					return fmt.Errorf("invalid time format: %v", err)
				}
				startTime = &t
			}

			// Get container name if not specified
			if container == "" {
				pod, err := client.GetPod(namespace, podName)
				if err != nil {
					return err
				}
				if len(pod.Containers) > 0 {
					container = pod.Containers[0].Name
				}
			}

			return client.GetPodLogs(namespace, podName, container, k8s.LogOptions{
				Follow:       follow,
				Previous:     previous,
				TailLines:    tail,
				Writer:       os.Stdout,
				SinceTime:    startTime,
				SinceSeconds: sinceSeconds,
			})
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output")
	cmd.Flags().BoolVarP(&previous, "previous", "p", false, "Print the logs for the previous instance of the container")
	cmd.Flags().Int64VarP(&tail, "tail", "t", -1, "Lines of recent log file to display")
	cmd.Flags().StringVarP(&container, "container", "c", "", "Print the logs of this container")
	cmd.Flags().StringVar(&since, "since", "", "Show logs since duration (e.g. 1h, 5m, 30s)")
	cmd.Flags().StringVar(&sinceTime, "since-time", "", "Show logs since specific time (RFC3339 format)")

	return cmd
}
