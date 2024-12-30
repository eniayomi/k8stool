package logs

import (
	"context"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type service struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

// NewLogService creates a new log service instance
func NewLogService(clientset *kubernetes.Clientset, config *rest.Config) (LogService, error) {
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

// GetLogs retrieves logs from a container in a pod
func (s *service) GetLogs(ctx context.Context, namespace, pod string, opts *LogOptions) (*LogResult, error) {
	if err := s.Validate(opts); err != nil {
		return nil, err
	}

	// First check if the pod exists
	podObj, err := s.clientset.CoreV1().Pods(namespace).Get(ctx, pod, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to find pod %s in namespace %s: %w", pod, namespace, err)
	}

	// If container is specified, verify it exists in the pod
	if opts.Container != "" {
		containerExists := false
		for _, container := range podObj.Spec.Containers {
			if container.Name == opts.Container {
				containerExists = true
				break
			}
		}
		if !containerExists {
			return nil, fmt.Errorf("container %s not found in pod %s", opts.Container, pod)
		}
	}

	req := s.buildLogRequest(namespace, pod, opts)
	logs, err := req.DoRaw(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	// Always write logs to the provided writer if one exists
	if opts.Writer != nil {
		if _, err := opts.Writer.Write(logs); err != nil {
			return nil, fmt.Errorf("failed to write logs: %w", err)
		}
	}

	// Return logs in the result
	return &LogResult{
		Logs: string(logs),
	}, nil
}

// StreamLogs streams logs from a container in a pod
func (s *service) StreamLogs(ctx context.Context, namespace, pod string, opts *LogOptions) (*LogConnection, error) {
	if err := s.Validate(opts); err != nil {
		return nil, err
	}

	// First check if the pod exists
	_, err := s.clientset.CoreV1().Pods(namespace).Get(ctx, pod, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to find pod %s in namespace %s: %w", pod, namespace, err)
	}

	req := s.buildLogRequest(namespace, pod, opts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create log stream: %w", err)
	}

	done := make(chan struct{})
	connection := &LogConnection{
		Reader: stream,
		Done:   done,
	}

	// Start streaming in a goroutine if a writer is provided
	if opts.Writer != nil {
		go func() {
			defer close(done)
			defer stream.Close()

			_, err := io.Copy(opts.Writer, stream)
			if err != nil && err != io.EOF {
				connection.Error = fmt.Errorf("error streaming logs: %w", err)
			}
		}()
	}

	return connection, nil
}

// Validate validates the log options
func (s *service) Validate(opts *LogOptions) error {
	if opts == nil {
		return fmt.Errorf("log options are required")
	}

	if opts.Follow && opts.Writer == nil {
		return fmt.Errorf("writer is required when following logs")
	}

	return nil
}

// buildLogRequest builds a request for retrieving container logs
func (s *service) buildLogRequest(namespace, pod string, opts *LogOptions) *rest.Request {
	var sinceTime *metav1.Time
	if opts.SinceTime != nil {
		sinceTime = &metav1.Time{Time: *opts.SinceTime}
	}

	return s.clientset.CoreV1().Pods(namespace).GetLogs(pod, &corev1.PodLogOptions{
		Container:    opts.Container,
		Follow:       opts.Follow,
		Previous:     opts.Previous,
		SinceTime:    sinceTime,
		SinceSeconds: opts.SinceSeconds,
		TailLines:    opts.TailLines,
		LimitBytes:   opts.LimitBytes,
		Timestamps:   opts.Timestamps,
	})
}
