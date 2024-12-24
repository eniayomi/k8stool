package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Use the package variables directly
func getVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("k8stool version: %s\n", Version)
			fmt.Printf("  Build Date: %s\n", Date)
			fmt.Printf("  Commit: %s\n", Commit)
			fmt.Printf("  Go version: %s\n", runtime.Version())
			fmt.Printf("  OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
	return cmd
}
