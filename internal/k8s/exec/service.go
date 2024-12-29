package exec

import (
	"context"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type service struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

// NewExecService creates a new exec service instance
func NewExecService(clientset *kubernetes.Clientset, config *rest.Config) (ExecService, error) {
	if clientset == nil {
		return nil, fmt.Errorf("kubernetes clientset is required")
	}
	if config == nil {
		return nil, fmt.Errorf("rest config is required")
	}
	return &service{
		clientset: clientset,
		config:    config,
	}, nil
}

// Exec executes a command in a container and returns the result
func (s *service) Exec(ctx context.Context, namespace, pod string, opts *ExecOptions) (*ExecResult, error) {
	if err := s.Validate(opts); err != nil {
		return nil, err
	}

	// Create a buffer to capture output
	var stdout, stderr io.Writer
	if opts.Streams != nil {
		stdout = opts.Streams.Out
		stderr = opts.Streams.ErrOut
	}

	// Create the exec request
	req := s.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: opts.Container,
		Command:   opts.Command,
		Stdin:     opts.Stdin,
		Stdout:    stdout != nil,
		Stderr:    stderr != nil,
		TTY:       opts.TTY,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(s.config, "POST", req.URL())
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	// Execute the command
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  opts.Streams.In,
		Stdout: stdout,
		Stderr: stderr,
		Tty:    opts.TTY,
	})

	if err != nil {
		return &ExecResult{
			ExitCode: -1,
			Error:    err.Error(),
		}, nil
	}

	return &ExecResult{
		ExitCode: 0,
	}, nil
}

// Stream executes a command in a container and streams the input/output
func (s *service) Stream(ctx context.Context, namespace, pod string, opts *ExecOptions) (*ExecConnection, error) {
	if err := s.Validate(opts); err != nil {
		return nil, err
	}

	// Create the exec request
	req := s.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: opts.Container,
		Command:   opts.Command,
		Stdin:     opts.Stdin,
		Stdout:    true,
		Stderr:    true,
		TTY:       opts.TTY,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(s.config, "POST", req.URL())
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	// Create pipes for stdin, stdout, and stderr
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()

	// Start streaming in a goroutine
	go func() {
		err := exec.StreamWithContext(ctx, remotecommand.StreamOptions{
			Stdin:  stdinReader,
			Stdout: stdoutWriter,
			Stderr: stderrWriter,
			Tty:    opts.TTY,
		})
		if err != nil {
			// Close all pipes on error
			stdinReader.CloseWithError(err)
			stdoutWriter.CloseWithError(err)
			stderrWriter.CloseWithError(err)
		}
	}()

	return &ExecConnection{
		Stdin:  stdinWriter,
		Stdout: stdoutReader,
		Stderr: stderrReader,
		TTY:    opts.TTY,
	}, nil
}

// Validate validates the exec options
func (s *service) Validate(opts *ExecOptions) error {
	if opts == nil {
		return fmt.Errorf("exec options are required")
	}

	if len(opts.Command) == 0 {
		return fmt.Errorf("command is required")
	}

	if opts.TTY && opts.Streams == nil {
		return fmt.Errorf("streams are required when TTY is enabled")
	}

	if opts.Stdin && opts.Streams == nil {
		return fmt.Errorf("streams are required when stdin is enabled")
	}

	return nil
}
