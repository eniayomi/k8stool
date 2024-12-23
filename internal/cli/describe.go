package cli

import (
	"fmt"
	"os"

	"k8stool/internal/k8s"

	"github.com/spf13/cobra"
)

func getDescribeCmd() *cobra.Command {
	var namespace string

	cmd := &cobra.Command{
		Use:   "describe [resource] [name]",
		Short: "Show details of a specific resource",
		Long: `Show details of a specific resource.
Supported resources: pod/po, deployment/deploy`,
		Example: `  # Describe a pod
  k8stool describe pod nginx-pod
  k8stool describe po nginx-pod

  # Describe a deployment
  k8stool describe deployment nginx
  k8stool describe deploy nginx`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			resource := args[0]
			name := args[1]

			switch resource {
			case "pod", "po":
				details, err := client.DescribePod(namespace, name)
				if err != nil {
					return err
				}
				return details.Print(os.Stdout)
			case "deployment", "deploy":
				return client.DescribeDeployment(namespace, name)
			default:
				return fmt.Errorf("unsupported resource type: %s", resource)
			}
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
	return cmd
}

func getResourceValue(value string) string {
	if value == "" || value == "0" {
		return "<none>"
	}
	return value
}
