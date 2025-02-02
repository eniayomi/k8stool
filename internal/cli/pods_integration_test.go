package cli

import (
	"bytes"
	"io"
	"os"
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

	rootCmd := getPodsCmd()

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		validate func(t *testing.T, output string)
	}{
		{
			name:    "list pods",
			args:    []string{},
			wantErr: false,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "NAME")
				assert.Contains(t, output, "STATUS")
			},
		},
		{
			name:    "list pods with namespace",
			args:    []string{"-n", "kube-system"},
			wantErr: false,
			validate: func(t *testing.T, output string) {
				// Look for pods that we know exist in our k3d cluster
				assert.Contains(t, output, "coredns")
				// or
				assert.Contains(t, output, "local-path-provisioner")
			},
		},
		{
			name:    "list pods with invalid namespace",
			args:    []string{"-n", "nonexistent-namespace"},
			wantErr: false,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "NAME")
				assert.Contains(t, output, "STATUS")
			},
		},
		{
			name:    "list pods with all namespaces",
			args:    []string{"--all-namespaces"},
			wantErr: false,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "NAMESPACE")
				assert.Contains(t, output, "NAME")
				assert.Contains(t, output, "STATUS")
				assert.Contains(t, output, "kube-system")
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
