package cli

import (
	"fmt"
	"path/filepath"

	k8s "k8stool/internal/k8s/client"

	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

// Version information
var (
	Version string
	Commit  string
	Date    string
)

// Command flags
var (
	kubeconfig string
	namespace  string
	verbose    bool
)

var rootCmd = &cobra.Command{
	Use:   "k8stool",
	Short: "K8sTool is a CLI tool for managing Kubernetes clusters",
	Long: `A CLI tool that helps you interact with Kubernetes clusters,
allowing you to view pods, logs, deployments, and more.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip client initialization for commands that don't need it
		if cmd.Name() == "embeddings" || cmd.Parent().Name() == "embeddings" {
			return nil
		}
		return initializeClient()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Initialize flags
	if home := homedir.HomeDir(); home != "" {
		rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "path to the kubeconfig file")
	} else {
		rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "path to the kubeconfig file")
	}

	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "the namespace to use")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")

	// Add commands to root
	rootCmd.AddCommand(getCmd())
	rootCmd.AddCommand(describeCmd())
	rootCmd.AddCommand(getLogsCmd())
	rootCmd.AddCommand(execCmd())
	rootCmd.AddCommand(portForwardCmd())
	rootCmd.AddCommand(contextCmd())
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(getNamespaceCmd())
	rootCmd.AddCommand(getMetricsCmd())
	rootCmd.AddCommand(NewAgentCmd())
	rootCmd.AddCommand(NewEmbeddingsCmd())
}

// getCmd returns the get command
func getCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get (pods|deployments|events)",
		Short: "Display one or many resources",
		Long:  `Display one or many resources.`,
	}

	cmd.AddCommand(getPodsCmd())
	cmd.AddCommand(getDeploymentsCmd())
	cmd.AddCommand(getEventsCmd())

	return cmd
}

// describeCmd returns the describe command
func describeCmd() *cobra.Command {
	return getDescribeCmd()
}

// execCmd returns the exec command
func execCmd() *cobra.Command {
	return getExecCmd()
}

// portForwardCmd returns the port-forward command
func portForwardCmd() *cobra.Command {
	return getPortForwardCmd()
}

// contextCmd returns the context command
func contextCmd() *cobra.Command {
	return getContextCmd()
}

// versionCmd returns the version command
func versionCmd() *cobra.Command {
	return getVersionCmd()
}

// initializeClient initializes the Kubernetes client configuration
func initializeClient() error {
	// Initialize the client
	_, err := k8s.NewClient()
	if err != nil {
		return fmt.Errorf("failed to initialize client: %w", err)
	}

	return nil
}
