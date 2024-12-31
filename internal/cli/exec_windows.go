//go:build windows
// +build windows

package cli

import (
	"fmt"
	"os"

	k8s "k8stool/internal/k8s/client"
	"k8stool/internal/k8s/pods"

	"github.com/spf13/cobra"
)

func getExecCmd() *cobra.Command {
	var container string
	var tty bool

	cmd := &cobra.Command{
		Use:   "exec [-c CONTAINER] POD COMMAND [args...]",
		Short: "Execute a command in a container",
		Long:  "Execute a command in a container. If the pod has multiple containers, use -c to specify which container to execute in.",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return fmt.Errorf("failed to initialize client: %w", err)
			}

			podName := args[0]
			command := args[1:]

			// Get current namespace
			currentCtx, err := client.ContextService.GetCurrent()
			if err != nil {
				return fmt.Errorf("failed to get current context: %w", err)
			}

			// Get pod to validate it exists and get container info
			pod, err := client.PodService.Get(currentCtx.Namespace, podName)
			if err != nil {
				return fmt.Errorf("failed to get pod: %w", err)
			}

			// If container not specified and pod has multiple containers, error out
			if container == "" && len(pod.Containers) > 1 {
				return fmt.Errorf("pod has multiple containers, use -c to specify which container to execute in")
			}

			// If container not specified, use the first container
			if container == "" {
				container = pod.Containers[0].Name
			}

			// Validate container exists in pod
			containerExists := false
			for _, c := range pod.Containers {
				if c.Name == container {
					containerExists = true
					break
				}
			}
			if !containerExists {
				return fmt.Errorf("container %q not found in pod %q", container, podName)
			}

			// Create exec options
			execOpts := pods.ExecOptions{
				Command: command,
				TTY:     tty,
				Stdin:   os.Stdin,
				Stdout:  os.Stdout,
				Stderr:  os.Stderr,
			}

			// Execute command in container
			return client.PodService.Exec(currentCtx.Namespace, podName, container, execOpts)
		},
	}

	cmd.Flags().StringVarP(&container, "container", "c", "", "Container name. If omitted, the first container in the pod will be chosen")
	cmd.Flags().BoolVarP(&tty, "tty", "t", false, "Allocate a pseudo-TTY")

	return cmd
}
