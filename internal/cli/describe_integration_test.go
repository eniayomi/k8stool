package cli

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDescribeCommands_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Save original stdout and restore it after tests
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	rootCmd := getDescribeCmd()

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		validate func(t *testing.T, output string)
	}{
		{
			name:    "describe pod",
			args:    []string{"pod", "curl-default"},
			wantErr: false,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "Name:")
				assert.Contains(t, output, "Namespace:")
				assert.Contains(t, output, "Node:")
				assert.Contains(t, output, "Status:")
				assert.Contains(t, output, "IP:")
				assert.Contains(t, output, "Containers:")
			},
		},
		{
			name:    "describe deployment",
			args:    []string{"deployment", "curl-default-deploy"},
			wantErr: false,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "Name:")
				assert.Contains(t, output, "Namespace:")
				assert.Contains(t, output, "Replicas:")
				assert.Contains(t, output, "Strategy")
				assert.Contains(t, output, "Selector:")
				assert.Contains(t, output, "Containers:")
			},
		},
		{
			name:    "describe invalid pod",
			args:    []string{"pod", "nonexistent-pod"},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "failed to get pod: pods \"nonexistent-pod\" not found")
			},
		},
		{
			name:    "describe invalid deployment",
			args:    []string{"deployment", "nonexistent-deployment"},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "failed to get deployment: deployments.apps \"nonexistent-deployment\" not found")
			},
		},
		{
			name:    "describe pod with invalid resource type",
			args:    []string{"invalid-type", "curl-default"},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "unsupported resource type: invalid-type")
			},
		},
		{
			name:    "describe pod in different namespace",
			args:    []string{"pod", "curl-deploy", "--namespace", "integration-test"},
			wantErr: false,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "Name:")
				assert.Contains(t, output, "Namespace: integration-test")
				assert.Contains(t, output, "Node:")
				assert.Contains(t, output, "Status:")
				assert.Contains(t, output, "IP:")
				assert.Contains(t, output, "Containers:")
			},
		},
		{
			name:    "describe deployment in different namespace",
			args:    []string{"deployment", "curl-deploy", "--namespace", "integration-test"},
			wantErr: false,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "Name:")
				assert.Contains(t, output, "Namespace: integration-test")
				assert.Contains(t, output, "Replicas:")
				assert.Contains(t, output, "Strategy")
				assert.Contains(t, output, "Selector:")
				assert.Contains(t, output, "Containers:")
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
