package cli

import (
	"fmt"
	"io"
	k8s "k8stool/internal/k8s/client"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func getExecCmd() *cobra.Command {
	var namespace string
	var container string
	var tty bool
	var stdin bool

	// Helper function to check if command is interactive
	isInteractiveCommand := func(cmd []string) bool {
		if len(cmd) == 0 {
			return false
		}
		interactive := []string{"bash", "sh", "zsh"}
		base := filepath.Base(cmd[0])
		for _, shell := range interactive {
			if base == shell {
				return true
			}
		}
		return false
	}

	cmd := &cobra.Command{
		Use:   "exec POD [COMMAND] [args...]",
		Short: "Execute a command in a container",
		Long: `Execute a command in a container.

Flags:
  -i, --stdin    Pass stdin to the container (required for interactive sessions)
  -t, --tty      Allocate a pseudo-TTY (requires -i flag)
  -c, --container Specify container within pod
  -n, --namespace Target namespace (defaults to current namespace)

Examples:
  # Run non-interactive command in pod
  k8stool exec nginx-pod -- ls /

  # Run command in specific container
  k8stool exec nginx-pod -c nginx -- ls /

  # Run interactive shell (automatically enables -it)
  k8stool exec nginx-pod -- bash

  # Run interactive shell explicitly
  k8stool exec -it nginx-pod -- /bin/bash

  # Run interactive command in specific namespace
  k8stool exec -n mynamespace -it nginx-pod -- /bin/sh`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			if namespace == "" {
				namespace = client.GetNamespace()
			}

			podName := args[0]
			var command []string
			if len(args) > 1 {
				command = args[1:]
			} else if tty {
				command = []string{"/bin/sh"}
			} else {
				return fmt.Errorf("command is required when not using -t flag")
			}

			// Auto-enable interactive mode for shell commands
			if isInteractiveCommand(command) && !tty && !stdin {
				tty = true
				stdin = true
			}

			// Set up IO
			var stdinReader io.Reader
			if stdin {
				stdinReader = os.Stdin
			}

			// Only enable TTY if both -t and -i are set
			if tty && !stdin {
				return fmt.Errorf("you must specify -i when using -t")
			}

			return client.ExecInPod(namespace, podName, container, k8s.ExecOptions{
				Command: command,
				TTY:     tty,
				Stdin:   stdinReader,
				Stdout:  os.Stdout,
				Stderr:  os.Stderr,
			})
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace")
	cmd.Flags().StringVarP(&container, "container", "c", "", "Container name within pod")
	cmd.Flags().BoolVarP(&tty, "tty", "t", false, "Allocate a pseudo-TTY (requires -i)")
	cmd.Flags().BoolVarP(&stdin, "stdin", "i", false, "Pass stdin to the container")

	return cmd
}
