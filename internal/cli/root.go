package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "k8stool",
	Short: "K8sTool is a CLI tool for managing Kubernetes clusters",
	Long: `A CLI tool that helps you interact with Kubernetes clusters,
allowing you to view pods, logs, deployments, and more.`,
}

func Execute() error {
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
}
