package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPodsCommands_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Save original stdout and restore it after tests
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		validate func(t *testing.T, output string)
	}{
		{
			name:    "list pods",
			args:    []string{"list"},
			wantErr: false,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "NAMESPACE")
				assert.Contains(t, output, "NAME")
				assert.Contains(t, output, "STATUS")
			},
		},
		{
			name:    "list pods with namespace",
			args:    []string{"list", "-n", "kube-system"},
			wantErr: false,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "kube-system")
			},
		},
		{
			name:    "list pods with invalid namespace",
			args:    []string{"list", "-n", "nonexistent-namespace"},
			wantErr: false, // Not an error, just empty list
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "NAMESPACE")
				assert.Contains(t, output, "NAME")
				assert.Contains(t, output, "STATUS")
			},
		},
		{
			name:    "describe pod without name",
			args:    []string{"describe"},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, strings.ToLower(output), "error")
				assert.Contains(t, strings.ToLower(output), "pod name")
			},
		},
		{
			name:    "describe nonexistent pod",
			args:    []string{"describe", "nonexistent-pod"},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, strings.ToLower(output), "not found")
			},
		},
		{
			name:    "logs without pod name",
			args:    []string{"logs"},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, strings.ToLower(output), "error")
				assert.Contains(t, strings.ToLower(output), "pod name")
			},
		},
		{
			name:    "logs for nonexistent pod",
			args:    []string{"logs", "nonexistent-pod"},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, strings.ToLower(output), "not found")
			},
		},
		{
			name:    "exec without pod name",
			args:    []string{"exec"},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, strings.ToLower(output), "error")
				assert.Contains(t, strings.ToLower(output), "pod name")
			},
		},
		{
			name:    "exec for nonexistent pod",
			args:    []string{"exec", "nonexistent-pod", "--", "ls"},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, strings.ToLower(output), "not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a pipe to capture output
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Create command
			cmd := getPodsCmd()
			cmd.SetOut(w)
			cmd.SetErr(w)
			cmd.SetArgs(tt.args)

			// Execute command
			err := cmd.Execute()

			// Read output
			w.Close()
			var buf bytes.Buffer
			_, err = io.Copy(&buf, r)
			if err != nil {
				t.Fatalf("failed to copy response: %v", err)
			}
			output := buf.String()

			t.Logf("Command output:\n%s", output)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			tt.validate(t, output)
		})
	}
}
