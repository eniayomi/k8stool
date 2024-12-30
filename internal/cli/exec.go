package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"

	k8s "k8stool/internal/k8s/client"
	"k8stool/internal/k8s/exec"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// terminalSizeQueue implements exec.TerminalSizeQueue
type terminalSizeQueue struct {
	sync.Mutex
	sizes chan *exec.TerminalSize
}

func newTerminalSizeQueue() *terminalSizeQueue {
	return &terminalSizeQueue{
		sizes: make(chan *exec.TerminalSize, 1),
	}
}

func (t *terminalSizeQueue) Next() *exec.TerminalSize {
	size := <-t.sizes
	return size
}

// handleTerminalResize handles terminal resize events for TTY sessions
func handleTerminalResize(ctx context.Context, conn *exec.ExecConnection) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return
	}

	sizeQueue := newTerminalSizeQueue()
	conn.TerminalSizeQueue = sizeQueue

	// Get the initial terminal size
	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return
	}

	// Send the initial size
	sizeQueue.sizes <- &exec.TerminalSize{
		Width:  uint16(width),
		Height: uint16(height),
	}

	// Create a channel to receive window resize signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)

	go func() {
		defer close(sizeQueue.sizes)
		for {
			select {
			case <-ctx.Done():
				return
			case <-sigCh:
				width, height, err := term.GetSize(int(os.Stdin.Fd()))
				if err != nil {
					continue
				}
				sizeQueue.sizes <- &exec.TerminalSize{
					Width:  uint16(width),
					Height: uint16(height),
				}
			}
		}
	}()
}

func getExecCmd() *cobra.Command {
	var container string
	var tty bool
	var stdin bool

	cmd := &cobra.Command{
		Use:   "exec POD [COMMAND] [args...]",
		Short: "Execute a command in a container",
		Long: `Execute a command in a container.
Examples:
  # Execute 'ls' in pod 'nginx'
  k8stool exec nginx ls

  # Execute 'ls' in pod 'nginx' container 'web'
  k8stool exec nginx -c web ls

  # Execute 'bash' in pod 'nginx' with TTY
  k8stool exec -it nginx bash`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			podName := args[0]
			command := args[1:]

			currentCtx, err := client.ContextService.GetCurrent()
			if err != nil {
				return err
			}
			namespace := currentCtx.Namespace

			// Get pod to validate container exists
			pod, err := client.PodService.Get(namespace, podName)
			if err != nil {
				return fmt.Errorf("failed to get pod %q: %w", podName, err)
			}

			// If container is specified, validate it exists
			if container != "" {
				containerExists := false
				for _, c := range pod.Containers {
					if c.Name == container {
						containerExists = true
						break
					}
				}
				if !containerExists {
					containerNames := make([]string, len(pod.Containers))
					for i, c := range pod.Containers {
						containerNames[i] = c.Name
					}
					return fmt.Errorf("container %q not found in pod %q. Available containers: %v",
						container, podName, containerNames)
				}
			} else if len(pod.Containers) > 1 {
				// If no container is specified and pod has multiple containers, show available containers
				containerNames := make([]string, len(pod.Containers))
				for i, c := range pod.Containers {
					containerNames[i] = c.Name
				}
				return fmt.Errorf("pod %q has multiple containers. Please specify one using -c flag. Available containers: %v",
					podName, containerNames)
			} else if len(pod.Containers) == 1 {
				// If no container is specified and pod has only one container, use it
				container = pod.Containers[0].Name
			} else {
				return fmt.Errorf("no containers found in pod %q", podName)
			}

			// Create streams for input/output
			streams := &exec.IOStreams{
				In:     os.Stdin,
				Out:    os.Stdout,
				ErrOut: os.Stderr,
			}

			// Create exec options
			opts := &exec.ExecOptions{
				Command:   command,
				Container: container,
				TTY:       tty,
				Stdin:     stdin,
				Streams:   streams,
			}

			// Validate options
			if err := client.ExecService.Validate(opts); err != nil {
				return err
			}

			ctx := context.Background()

			// If using TTY or stdin, use Stream instead of Exec
			if tty || stdin {
				// Create a context that can be cancelled
				ctx, cancel := context.WithCancel(ctx)
				defer cancel()

				// Handle interrupt signals
				sigCh := make(chan os.Signal, 1)
				signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
				go func() {
					<-sigCh
					cancel()
				}()

				// Stream the command
				conn, err := client.ExecService.Stream(ctx, namespace, podName, opts)
				if err != nil {
					return err
				}

				// Handle TTY resize if needed
				if tty && stdin {
					go handleTerminalResize(ctx, conn)
				}

				// Copy stdin to the container if enabled
				if stdin {
					go func() {
						defer conn.Stdin.Close()
						_, _ = io.Copy(conn.Stdin, os.Stdin)
					}()
				}

				// Copy output from the container
				if conn.TTY {
					_, _ = io.Copy(os.Stdout, conn.Stdout)
				} else {
					go func() {
						_, _ = io.Copy(os.Stdout, conn.Stdout)
					}()
					_, _ = io.Copy(os.Stderr, conn.Stderr)
				}

				return nil
			}

			// For non-interactive commands, use Exec
			result, err := client.ExecService.Exec(ctx, namespace, podName, opts)
			if err != nil {
				return err
			}

			if result.Error != "" {
				return fmt.Errorf("exec error: %s", result.Error)
			}

			if result.ExitCode != 0 {
				return fmt.Errorf("command exited with code %d", result.ExitCode)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&container, "container", "c", "", "Container name. If omitted, the first container in the pod will be chosen")
	cmd.Flags().BoolVarP(&tty, "tty", "t", false, "Allocate a pseudo-TTY")
	cmd.Flags().BoolVarP(&stdin, "stdin", "i", false, "Pass stdin to the container")

	return cmd
}
