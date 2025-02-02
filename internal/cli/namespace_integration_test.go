package cli

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNamespaceCommands_Integration(t *testing.T) {
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
			name:    "list namespaces",
			args:    []string{"list"},
			wantErr: false,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "NAME")
				assert.Contains(t, output, "STATUS")
				assert.Contains(t, output, "Active")
				assert.Contains(t, output, "kube-system")
				assert.Contains(t, output, "default")
			},
		},
		{
			name:    "show current namespace",
			args:    []string{"current"},
			wantErr: false,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "Current namespace: ")
			},
		},
		{
			name:    "switch to non-existent namespace",
			args:    []string{"switch", "nonexistent-namespace"},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "Error: namespaces \"nonexistent-namespace\" not found")
			},
		},
		{
			name:    "interactive mode with switch command",
			args:    []string{"switch", "-i"},
			wantErr: true, // Will fail in non-interactive test environment
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "Error: failed to get user input")
			},
		},
		{
			name:    "interactive mode with -i flag",
			args:    []string{"-i"},
			wantErr: true, // Will fail in non-interactive test environment
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "Error: failed to get user input")
			},
		},
		{
			name:    "direct switch to non-existent namespace",
			args:    []string{"nonexistent-namespace"},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "Error: namespaces \"nonexistent-namespace\" not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a pipe to capture output
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Create fresh command for each test
			cmd := getNamespaceCmd()
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
