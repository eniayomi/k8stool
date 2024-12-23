package cli

import (
	"os"

	"k8stool/internal/k8s"

	"github.com/spf13/cobra"
)

func getExecCmd() *cobra.Command {
	var namespace string
	var container string
	var command []string

	cmd := &cobra.Command{
		Use:   "exec [pod-name] [shell|command]",
		Short: "Execute a command in a container",
		Long: `Execute a command in a container.
Example: k8stool exec nginx-pod bash     # Start bash shell
         k8stool exec nginx-pod sh       # Start sh shell
         k8stool exec nginx-pod -- ls    # Run specific command

You can use either 'pod' or 'po' when referring to the pod name.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			podName := args[0]

			// Find command after "--"
			cmdIndex := -1
			for i, arg := range args {
				if arg == "--" {
					cmdIndex = i
					break
				}
			}

			if cmdIndex == -1 {
				if len(args) > 1 {
					// Check if second argument is a shell request
					switch args[1] {
					case "bash":
						command = []string{"/bin/bash"}
					case "sh":
						command = []string{"/bin/sh"}
					default:
						// Try bash first, fallback to sh
						if err := client.ExecInPod(namespace, podName, container, k8s.ExecOptions{
							Command: []string{"/bin/bash"},
							Stdin:   os.Stdin,
							Stdout:  os.Stdout,
							Stderr:  os.Stderr,
							TTY:     true,
						}); err != nil {
							command = []string{"/bin/sh"}
						} else {
							return nil
						}
					}
				} else {
					// No shell specified, use smart detection
					if err := client.ExecInPod(namespace, podName, container, k8s.ExecOptions{
						Command: []string{"/bin/bash"},
						Stdin:   os.Stdin,
						Stdout:  os.Stdout,
						Stderr:  os.Stderr,
						TTY:     true,
					}); err != nil {
						command = []string{"/bin/sh"}
					} else {
						return nil
					}
				}
			} else {
				command = args[cmdIndex+1:]
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

			return client.ExecInPod(namespace, podName, container, k8s.ExecOptions{
				Command: command,
				Stdin:   os.Stdin,
				Stdout:  os.Stdout,
				Stderr:  os.Stderr,
				TTY:     true,
			})
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
	cmd.Flags().StringVarP(&container, "container", "c", "", "Container name")

	return cmd
}
