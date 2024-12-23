package cli

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"k8stool/internal/k8s"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func getPortForwardCmd() *cobra.Command {
	var namespace string
	var address string
	var interactive bool

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
  k8stool port-forward pod nginx 80

  # Interactive mode
  k8stool port-forward -i`,
		Aliases: []string{"pf"},
		Args:    cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			// If namespace flag not provided, use the client's current namespace
			if namespace == "" {
				namespace = client.GetNamespace()
			}

			// Handle interactive mode
			if interactive {
				return handleInteractivePortForward(client, namespace, address)
			}

			// Original non-interactive logic continues here
			if len(args) < 2 {
				return fmt.Errorf("resource type and name are required")
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
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive mode")

	return cmd
}

func handleInteractivePortForward(client *k8s.Client, namespace, address string) error {
	// Get list of pods
	pods, err := client.ListPods(namespace, false, "", "")
	if err != nil {
		return err
	}

	// Create pod selection prompt
	podPrompt := promptui.Select{
		Label: "Select pod to port-forward",
		Items: pods,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "▸ {{ .Name | cyan }}",
			Inactive: "  {{ .Name }}",
			Selected: "✔ {{ .Name | green }}",
		},
	}

	idx, _, err := podPrompt.Run()
	if err != nil {
		return err
	}

	selectedPod := pods[idx]

	// Get container ports
	containerPorts := []string{}
	for _, container := range selectedPod.Containers {
		for _, port := range container.Ports {
			containerPorts = append(containerPorts,
				fmt.Sprintf("%d:%d (%s)", port.ContainerPort, port.ContainerPort, container.Name))
		}
	}

	if len(containerPorts) == 0 {
		return fmt.Errorf("no ports exposed by pod %s", selectedPod.Name)
	}

	// Create port selection prompt
	portPrompt := promptui.Select{
		Label: "Select port mapping (container:local)",
		Items: containerPorts,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "▸ {{ . | cyan }}",
			Inactive: "  {{ . }}",
			Selected: "✔ {{ . | green }}",
		},
	}

	_, portMapping, err := portPrompt.Run()
	if err != nil {
		return err
	}

	// Parse port mapping and create PortMapping slice
	parts := strings.Split(strings.Fields(portMapping)[0], ":")
	portMappings := []k8s.PortMapping{
		{
			Local:  parts[0],
			Remote: parts[1],
		},
	}

	// Start port forwarding
	stopChan := make(chan struct{}, 1)
	readyChan := make(chan struct{})

	return client.PortForwardPod(namespace, selectedPod.Name, address, portMappings, stopChan, readyChan)
}
