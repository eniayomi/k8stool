package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextCommands_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Save original stdout and restore it after tests
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	tests := []struct {
		name     string
		command  string
		args     []string
		wantErr  bool
		validate func(t *testing.T, output string)
	}{
		{
			name:    "list contexts",
			command: "list",
			args:    []string{},
			wantErr: false,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "CURRENT")
				assert.Contains(t, output, "NAME")
				assert.Contains(t, output, "CLUSTER")
			},
		},
		{
			name:    "show current context",
			command: "current",
			args:    []string{},
			wantErr: false,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "Current context:")
				assert.True(t, len(strings.TrimSpace(output)) > 0)
			},
		},
		{
			name:    "switch context with invalid name",
			command: "switch",
			args:    []string{"nonexistent-context"},
			wantErr: true,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "does not exist")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a pipe to capture output
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Create command
			cmd := getContextCmd()
			cmd.SetOut(w)
			cmd.SetErr(w)
			cmd.SetArgs(append([]string{tt.command}, tt.args...))

			// Execute command
			var err error
			err = cmd.Execute()
			if err != nil && !tt.wantErr {
				t.Fatalf("unexpected error executing command: %v", err)
			}

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
