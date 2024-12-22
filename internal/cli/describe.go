package cli

import (
	"fmt"

	"github.com/eniayomi/k8stool/internal/k8s"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func getDescribeCmd() *cobra.Command {
	var namespace string

	cmd := &cobra.Command{
		Use:   "describe [resource] [name]",
		Short: "Show details of a specific resource",
		Long:  `Show detailed information about a specific Kubernetes resource`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			resourceType := args[0]
			name := args[1]

			if resourceType != "pod" {
				return fmt.Errorf("only pod description is supported currently")
			}

			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			pod, err := client.DescribePod(namespace, name)
			if err != nil {
				return err
			}

			// Print pod details
			bold := color.New(color.Bold)

			bold.Println("Pod Details:")
			fmt.Printf("  Name:            %s\n", pod.Name)
			fmt.Printf("  Namespace:       %s\n", pod.Namespace)
			fmt.Printf("  Node:            %s\n", pod.NodeName)
			fmt.Printf("  Status:          %s\n", colorizeStatus(pod.Status))
			fmt.Printf("  IP:              %s\n", pod.PodIP)
			fmt.Printf("  Created:         %s\n", pod.CreatedAt)

			if len(pod.NodeSelector) > 0 {
				fmt.Printf("  Node-Selectors:   %s: %s", pod.NodeSelector)
			} else {
				fmt.Printf("  Node-Selectors:   <none>\n")
			}

			if len(pod.Tolerations) > 0 {
				bold.Println("\nTolerations:")
				for _, toleration := range pod.Tolerations {
					fmt.Printf("  • Key: %s, Operator: %s, Value: %s, Effect: %s\n",
						toleration.Key,
						toleration.Operator,
						toleration.Value,
						toleration.Effect)
				}
			}

			if len(pod.Volumes) > 0 {
				bold.Println("\nVolumes:")
				for _, vol := range pod.Volumes {
					fmt.Printf("  • %s (%s):\n", vol.Name, vol.Type)
					fmt.Printf("    Source: %s\n", vol.Source)
					if vol.ReadOnly {
						fmt.Printf("    ReadOnly: true\n")
					}
				}
			}

			bold.Println("\nContainers:")
			for _, container := range pod.Containers {
				fmt.Printf("\n  • %s:\n", container.Name)
				fmt.Printf("      Image:         %s\n", container.Image)
				fmt.Printf("      State:         %s\n", container.State)
				fmt.Printf("      Ready:         %v\n", container.Ready)
				fmt.Printf("      Restart Count: %d\n", container.RestartCount)

				fmt.Printf("\n      Resources:\n")
				fmt.Printf("        Requests:\n")
				fmt.Printf("          CPU:    %s\n", getResourceValue(container.Resources.Requests.CPU))
				fmt.Printf("          Memory: %s\n", getResourceValue(container.Resources.Requests.Memory))
				fmt.Printf("        Limits:\n")
				fmt.Printf("          CPU:    %s\n", getResourceValue(container.Resources.Limits.CPU))
				fmt.Printf("          Memory: %s\n", getResourceValue(container.Resources.Limits.Memory))

				if len(container.Mounts) > 0 {
					fmt.Printf("\n      Mounts:\n")
					for _, mount := range container.Mounts {
						fmt.Printf("        • %s -> %s", mount.Name, mount.MountPath)
						if mount.ReadOnly {
							fmt.Printf(" (ro)")
						}
						fmt.Println()
					}
				}
			}

			if len(pod.Events) > 0 {
				bold.Println("\nEvents:")
				for _, event := range pod.Events {
					fmt.Printf("  %s  %s  %s\n",
						event.Time,
						event.Type,
						event.Message)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
	return cmd
}

func getResourceValue(value string) string {
	if value == "" {
		return "<none>"
	}
	return value
}
