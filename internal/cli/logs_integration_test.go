package cli

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogsCommands_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Save original stdout and restore it after tests
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	rootCmd := getLogsCmd()

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		validate func(t *testing.T, output string)
	}{
		{
			name:    "get pod logs",
			args:    []string{"pod", "nginx-default"},
			wantErr: false,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "nginx")
			},
		},
		{
			name:    "get deployment logs",
			args:    []string{"deployment", "nginx-default-deploy"},
			wantErr: false,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "nginx")
			},
		},
		{
			name:    "get logs with invalid pod",
			args:    []string{"pod", "nonexistent-pod"},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "failed to get logs: failed to find pod nonexistent-pod in namespace default: pods \"nonexistent-pod\" not found")
			},
		},
		{
			name:    "get logs with invalid deployment",
			args:    []string{"deployment", "nonexistent-deployment"},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "failed to get deployment: failed to get deployment: deployments.apps \"nonexistent-deployment\" not found")
			},
		},
		{
			name:    "get logs with container flag",
			args:    []string{"pod", "nginx-default", "-c", "nginx-default"},
			wantErr: false,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "nginx")
			},
		},
		{
			name:    "get logs with invalid container",
			args:    []string{"pod", "nginx-default", "-c", "nonexistent-container"},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "failed to get logs: container nonexistent-container not found in pod nginx-default")
			},
		},
		{
			name:    "get logs with tail flag",
			args:    []string{"pod", "nginx-default", "-c", "nginx-default", "--tail", "10"},
			wantErr: false,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "nginx")
			},
		},
		{
			name:    "get logs with since flag",
			args:    []string{"pod", "nginx-default", "-c", "nginx-default", "--since", "1h"},
			wantErr: false,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "nginx")
			},
		},
		{
			name:    "get logs with previous flag",
			args:    []string{"pod", "nginx-default", "-p"},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "failed to get logs: no previous terminated state found for container")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a pipe to capture output
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Create command
			cmd := rootCmd
			cmd.SetOut(w)
			cmd.SetErr(w)
			cmd.SetArgs(tt.args)

			// Execute command
			execErr := cmd.Execute()

			// Read output
			w.Close()
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, r); err != nil {
				t.Fatalf("failed to copy response: %v", err)
			}
			output := buf.String()

			t.Logf("Command output:\n%s", output)

			if tt.wantErr {
				assert.Error(t, execErr)
			} else {
				assert.NoError(t, execErr)
			}
			tt.validate(t, output)
		})
	}
}
