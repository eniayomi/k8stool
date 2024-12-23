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
		Use:   "describe (pod|deployment) [name]",
		Short: "Show details of a specific resource",
		Long: `Show details of a specific resource.
Example: k8stool describe pod nginx-pod
         k8stool describe deployment nginx`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			// If namespace flag not provided, use the client's current namespace
			if namespace == "" {
				namespace = client.GetNamespace()
			}

			resourceType := args[0]
			name := args[1]

			switch resourceType {
			case "pod", "po":
				pod, err := client.DescribePod(namespace, name)
				if err != nil {
					return err
				}

				details, err := client.GetDetails(namespace, resourceType, name)
				if err != nil {
					return err
				}

				return pod.Print(os.Stdout, details)
			case "deployment", "deploy":
				deployment, err := client.DescribeDeployment(namespace, name)
				if err != nil {
					return err
				}

				details, err := client.GetDetails(namespace, resourceType, name)
				if err != nil {
					return err
				}

				return deployment.Print(os.Stdout, details)
			default:
				return fmt.Errorf("unsupported resource type: %s", resourceType)
			}
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace")

	return cmd
}

func getResourceValue(value string) string {
	if value == "" || value == "0" {
		return "<none>"
	}
	return value
}
