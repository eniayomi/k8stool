package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	k8s "k8stool/internal/k8s/client"
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
	rootCmd.AddCommand(getVersionCmd())
}

// initializeClient initializes the Kubernetes client configuration
func initializeClient() error {
	var config *rest.Config
	var err error

	// Try to build config from kubeconfig file
	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		// Try in-cluster config
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Initialize the client
	if err := k8s.Initialize(clientset, config); err != nil {
		return fmt.Errorf("failed to initialize client: %w", err)
	}

	return nil
}
