package cli

import (
	"github.com/spf13/cobra"
)

// Add these variables at package level
var (
	version string
	commit  string
	date    string
)

var rootCmd = &cobra.Command{
	Use:   "k8stool",
	Short: "K8sTool is a CLI tool for managing Kubernetes clusters",
	Long: `A CLI tool that helps you interact with Kubernetes clusters,
allowing you to view pods, logs, deployments, and more.`,
}

// Update Execute to accept version info
func Execute(v, c, d string) error {
	version = v
	commit = c
	date = d
	return rootCmd.Execute()
}

func init() {
	// Add the 'get' command as a parent command
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get Kubernetes resources",
		Long:  `Get Kubernetes resources such as pods, deployments, services, etc.`,
	}

	// Add resource commands to the 'get' command
	getCmd.AddCommand(getPodsCmd())
	getCmd.AddCommand(getDeploymentsCmd())
	getCmd.AddCommand(getEventsCmd())

	// Add commands to root
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(getLogsCmd())
	rootCmd.AddCommand(getContextCmd())
	rootCmd.AddCommand(getNamespaceCmd())
	rootCmd.AddCommand(getDescribeCmd())
	rootCmd.AddCommand(getMetricsCmd())
	rootCmd.AddCommand(getExecCmd())
	rootCmd.AddCommand(getPortForwardCmd())
	rootCmd.AddCommand(getVersionCmd(version, commit, date))
}
