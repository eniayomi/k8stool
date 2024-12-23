package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

func getVersionCmd(version, commit, date string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("k8stool version: %s\n", version)
			fmt.Printf("  Build Date: %s\n", date)
			fmt.Printf("  Commit: %s\n", commit)
			fmt.Printf("  Go version: %s\n", runtime.Version())
			fmt.Printf("  OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
	return cmd
}
