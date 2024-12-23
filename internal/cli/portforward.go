package cli

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"k8stool/internal/k8s"

	"github.com/spf13/cobra"
)

func getPortForwardCmd() *cobra.Command {
	var namespace string
	var address string

	cmd := &cobra.Command{
		Use:   "port-forward (pod|deployment) NAME [LOCAL_PORT:]REMOTE_PORT [...[LOCAL_PORT_N:]REMOTE_PORT_N]",
		Short: "Forward local ports to a pod or deployment",
		Long: `Forward one or more local ports to a pod or deployment.
Examples:
  # Forward local port 8080 to pod port 80
  k8stool port-forward pod nginx 8080:80

  # Forward local port 8080 to deployment port 80
  k8stool port-forward deployment nginx 8080:80

  # Forward multiple ports
  k8stool port-forward pod nginx 8080:80 9090:90

  # Forward same remote port if local port not specified
  k8stool port-forward pod nginx 80`,
		Aliases: []string{"pf"},
		Args:    cobra.MinimumNArgs(2),
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
			ports := args[2:]

			// If no ports specified, return error
			if len(ports) == 0 {
				return fmt.Errorf("at least one port mapping is required")
			}

			// Parse port mappings
			portMappings := make([]k8s.PortMapping, 0, len(ports))
			for _, port := range ports {
				parts := strings.Split(port, ":")
				if len(parts) > 2 {
					return fmt.Errorf("invalid port mapping: %s", port)
				}

				var localPort, remotePort string
				if len(parts) == 1 {
					// If only one port specified, use same port for local and remote
					localPort = parts[0]
					remotePort = parts[0]
				} else {
					localPort = parts[0]
					remotePort = parts[1]
				}

				portMappings = append(portMappings, k8s.PortMapping{
					Local:  localPort,
					Remote: remotePort,
				})
			}

			// Handle interrupt signal
			signals := make(chan os.Signal, 1)
			signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
			defer signal.Stop(signals)

			// Start port forwarding
			stopChan := make(chan struct{}, 1)
			readyChan := make(chan struct{})

			go func() {
				sig := <-signals
				fmt.Printf("\nReceived signal: %v\n", sig)
				close(stopChan)
			}()

			fmt.Printf("Starting port forward for %s/%s...\n", resourceType, name)

			switch resourceType {
			case "pod", "po":
				return client.PortForwardPod(namespace, name, address, portMappings, stopChan, readyChan)
			case "deployment", "deploy":
				return client.PortForwardDeployment(namespace, name, address, portMappings, stopChan, readyChan)
			default:
				return fmt.Errorf("unsupported resource type: %s", resourceType)
			}
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace")
	cmd.Flags().StringVar(&address, "address", "localhost", "Local address to bind to")

	return cmd
}
