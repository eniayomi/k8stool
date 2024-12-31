package cli

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecCommands_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Save original stdout and restore it after tests
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	rootCmd := getExecCmd()

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		validate func(t *testing.T, output string)
	}{
		{
			name:    "exec ls command",
			args:    []string{"nginx-default", "ls", "/"},
			wantErr: false,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "etc")
				assert.Contains(t, output, "usr")
			},
		},
		{
			name:    "exec with invalid pod",
			args:    []string{"nonexistent-pod", "ls"},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "failed to get pod: failed to get pod: pods \"nonexistent-pod\" not found")
			},
		},
		{
			name:    "exec with invalid command",
			args:    []string{"nginx-default", "invalid-command"},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "unable to start container process: exec: \"invalid-command\": executable file not found in $PATH")
			},
		},
		{
			name:    "exec with container flag",
			args:    []string{"nginx-default", "-c", "nginx-default", "ls", "/"},
			wantErr: false,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "etc")
				assert.Contains(t, output, "usr")
			},
		},
		{
			name:    "exec with invalid container",
			args:    []string{"nginx-default", "-c", "nonexistent-container", "ls"},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "container \"nonexistent-container\" not found in pod \"nginx-default\"")
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
