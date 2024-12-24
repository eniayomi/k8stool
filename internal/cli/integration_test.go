package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextCommands_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	tests := []struct {
		name       string
		command    string
		args       []string
		wantOutput string
		wantErr    bool
	}{
		{
			name:    "list contexts",
			command: "context",
			args:    []string{"list"},
			wantErr: false,
		},
		{
			name:    "show current context",
			command: "context",
			args:    []string{"current"},
			wantErr: false,
		},
		{
			name:    "switch context with invalid name",
			command: "context",
			args:    []string{"switch", "nonexistent-context"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRootCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(append([]string{tt.command}, tt.args...))

			err := cmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, buf.String())
			}
		})
	}
}
