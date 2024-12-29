package cli

import (
	"fmt"
	"os"
	"time"

	k8s "k8stool/internal/k8s/client"

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
	var allContainers bool

	cmd := &cobra.Command{
		Use:   "logs (pod|deployment) [name]",
		Short: "View logs from containers",
		Long: `View logs from containers in pods or deployments.
Example: k8stool logs pod nginx-pod
         k8stool logs deployment nginx
         k8stool logs deploy nginx`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			// If namespace flag not provided, use the client's current namespace
			if namespace == "" {
				namespace = client.GetNamespace()
			}

			resourceType := args[0]
			name := args[1]

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

			switch resourceType {
			case "pod", "po":
				return client.GetPodLogs(namespace, name, container, k8s.LogOptions{
					Follow:       follow,
					Previous:     previous,
					TailLines:    tail,
					Writer:       os.Stdout,
					SinceTime:    startTime,
					SinceSeconds: sinceSeconds,
				})
			case "deployment", "deploy":
				return client.GetDeploymentLogs(namespace, name, k8s.LogOptions{
					Follow:        follow,
					Previous:      previous,
					TailLines:     tail,
					Writer:        os.Stdout,
					SinceTime:     startTime,
					SinceSeconds:  sinceSeconds,
					Container:     container,
					AllContainers: allContainers,
				})
			default:
				return fmt.Errorf("unsupported resource type: %s", resourceType)
			}
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace")
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output")
	cmd.Flags().BoolVarP(&previous, "previous", "p", false, "Print the logs for the previous instance")
	cmd.Flags().Int64VarP(&tail, "tail", "t", -1, "Lines of recent log file to display")
	cmd.Flags().StringVarP(&container, "container", "c", "", "Print the logs of this container")
	cmd.Flags().StringVar(&since, "since", "", "Show logs since duration (e.g. 1h, 5m, 30s)")
	cmd.Flags().StringVar(&sinceTime, "since-time", "", "Show logs since specific time (RFC3339 format)")
	cmd.Flags().BoolVarP(&allContainers, "all-containers", "a", false, "Get logs from all containers")

	return cmd
}
