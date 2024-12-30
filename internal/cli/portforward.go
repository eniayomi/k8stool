package cli

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	k8s "k8stool/internal/k8s/client"
	"k8stool/internal/k8s/portforward"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func getPortForwardCmd() *cobra.Command {
	var namespace string
	var address string
	var interactive bool
	var protocol string

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

  # Forward using UDP protocol
  k8stool port-forward pod nginx 8080:80 --protocol=udp

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
				currentCtx, err := client.ContextService.GetCurrent()
				if err != nil {
					return err
				}
				namespace = currentCtx.Namespace
			}

			// Handle interactive mode
			if interactive {
				return handleInteractivePortForward(client, namespace, address, protocol)
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
			portMappings := make([]portforward.PortMapping, 0, len(ports))
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

				// Convert string ports to uint16
				localPortNum, err := strconv.ParseUint(localPort, 10, 16)
				if err != nil {
					return fmt.Errorf("invalid local port: %s", localPort)
				}
				remotePortNum, err := strconv.ParseUint(remotePort, 10, 16)
				if err != nil {
					return fmt.Errorf("invalid remote port: %s", remotePort)
				}

				portMappings = append(portMappings, portforward.PortMapping{
					Local:    uint16(localPortNum),
					Remote:   uint16(remotePortNum),
					Address:  address,
					Protocol: protocol,
				})
			}

			// Validate port forwarding before starting
			if err := client.PortForwardService.ValidatePortForward(namespace, name, portMappings); err != nil {
				return fmt.Errorf("port forward validation failed: %v", err)
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

			opts := portforward.PortForwardOptions{
				Ports:        portMappings,
				StopChannel:  stopChan,
				ReadyChannel: readyChan,
				Streams: portforward.Streams{
					Out:    os.Stdout,
					ErrOut: os.Stderr,
				},
			}

			var result *portforward.PortForwardResult
			switch resourceType {
			case "pod", "po":
				result, err = client.PortForwardService.ForwardPodPort(namespace, name, opts)
			case "deployment", "deploy":
				result, err = client.PortForwardService.ForwardServicePort(namespace, name, opts)
			default:
				return fmt.Errorf("unsupported resource type: %s", resourceType)
			}

			if err != nil {
				return err
			}

			if result.Error != nil {
				return result.Error
			}

			// Wait for ready signal
			<-readyChan

			// Print forwarded ports
			fmt.Println("Port forwarding is ready:")
			for _, port := range result.Ports {
				fmt.Printf("  %s:%d -> %d\n", port.Address, port.Local, port.Remote)
			}

			// Wait for stop signal
			<-stopChan

			// Stop port forwarding
			if err := client.PortForwardService.StopForwarding(result); err != nil {
				fmt.Printf("Error stopping port forward: %v\n", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace")
	cmd.Flags().StringVar(&address, "address", "localhost", "Local address to bind to")
	cmd.Flags().StringVar(&protocol, "protocol", string(portforward.TCP), "Protocol to use (tcp or udp)")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive mode")

	return cmd
}

func handleInteractivePortForward(client *k8s.Client, namespace, address, protocol string) error {
	// First, let the user choose between pod and deployment
	resourceTypes := []string{"pod", "deployment"}
	resourcePrompt := promptui.Select{
		Label: "Select resource type",
		Items: resourceTypes,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "▸ {{ . | cyan }}",
			Inactive: "  {{ . }}",
			Selected: "✔ {{ . | green }}",
		},
	}

	resourceIdx, _, err := resourcePrompt.Run()
	if err != nil {
		return err
	}

	resourceType := resourceTypes[resourceIdx]

	var containerPorts []string
	var resourceName string

	if resourceType == "pod" {
		// Get list of pods
		podList, err := client.PodService.List(namespace, false, "", "")
		if err != nil {
			return err
		}

		// Create pod selection prompt
		podPrompt := promptui.Select{
			Label: "Select pod to port-forward",
			Items: podList,
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

		selectedPod := podList[idx]
		resourceName = selectedPod.Name

		// Get pod details to access container information
		podDetails, err := client.PodService.Get(namespace, selectedPod.Name)
		if err != nil {
			return err
		}

		// Get container ports
		for _, container := range podDetails.Containers {
			for _, port := range container.Ports {
				if port.ContainerPort > 0 {
					containerPorts = append(containerPorts,
						fmt.Sprintf("%d:%d (%s/%s)", port.ContainerPort, port.ContainerPort, container.Name, port.Protocol))
				}
			}
		}
	} else {
		// Get list of deployments
		deploymentList, err := client.DeploymentService.List(namespace, false, "")
		if err != nil {
			return err
		}

		// Create deployment selection prompt
		deploymentPrompt := promptui.Select{
			Label: "Select deployment to port-forward",
			Items: deploymentList,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}",
				Active:   "▸ {{ .Name | cyan }}",
				Inactive: "  {{ .Name }}",
				Selected: "✔ {{ .Name | green }}",
			},
		}

		idx, _, err := deploymentPrompt.Run()
		if err != nil {
			return err
		}

		selectedDeployment := deploymentList[idx]
		resourceName = selectedDeployment.Name

		// Get deployment details to access container information
		deploymentDetails, err := client.DeploymentService.Describe(namespace, selectedDeployment.Name)
		if err != nil {
			return err
		}

		// Get container ports
		for _, container := range deploymentDetails.Containers {
			for _, port := range container.Ports {
				containerPorts = append(containerPorts,
					fmt.Sprintf("%d:%d (%s)", port.ContainerPort, port.ContainerPort, container.Name))
			}
		}
	}

	if len(containerPorts) == 0 {
		return fmt.Errorf("no ports exposed by %s %s", resourceType, resourceName)
	}

	// Let user select port mapping
	portPrompt := promptui.Select{
		Label: "Select port mapping",
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

	// Parse selected port mapping
	parts := strings.Split(strings.Split(portMapping, " ")[0], ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid port mapping format")
	}

	remotePort, err := strconv.ParseUint(parts[1], 10, 16)
	if err != nil {
		return fmt.Errorf("invalid remote port: %v", err)
	}

	// Ask if user wants to specify a local port
	customPortPrompt := promptui.Prompt{
		Label:     "Do you want to specify a local port? (y/N)",
		IsConfirm: true,
	}

	useCustomPort, _ := customPortPrompt.Run()
	var localPort uint64

	if strings.ToLower(useCustomPort) == "y" {
		localPortPrompt := promptui.Prompt{
			Label:   "Enter local port",
			Default: parts[1], // Use remote port as default
			Validate: func(input string) error {
				port, err := strconv.ParseUint(input, 10, 16)
				if err != nil {
					return fmt.Errorf("invalid port number")
				}
				if port < 1 || port > 65535 {
					return fmt.Errorf("port must be between 1 and 65535")
				}
				return nil
			},
		}

		localPortStr, err := localPortPrompt.Run()
		if err != nil {
			return err
		}

		localPort, err = strconv.ParseUint(localPortStr, 10, 16)
		if err != nil {
			return fmt.Errorf("invalid local port: %v", err)
		}
	} else {
		localPort = remotePort
	}

	// Create port mapping
	portMappings := []portforward.PortMapping{
		{
			Local:    uint16(localPort),
			Remote:   uint16(remotePort),
			Address:  address,
			Protocol: protocol,
		},
	}

	// Validate port forwarding
	if err := client.PortForwardService.ValidatePortForward(namespace, resourceName, portMappings); err != nil {
		return fmt.Errorf("port forward validation failed: %v", err)
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

	fmt.Printf("Starting port forward for %s/%s...\n", resourceType, resourceName)

	opts := portforward.PortForwardOptions{
		Ports:        portMappings,
		StopChannel:  stopChan,
		ReadyChannel: readyChan,
		Streams: portforward.Streams{
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
	}

	var result *portforward.PortForwardResult
	if resourceType == "pod" {
		result, err = client.PortForwardService.ForwardPodPort(namespace, resourceName, opts)
	} else {
		result, err = client.PortForwardService.ForwardServicePort(namespace, resourceName, opts)
	}

	if err != nil {
		return err
	}

	if result.Error != nil {
		return result.Error
	}

	// Wait for ready signal
	<-readyChan

	// Print forwarded ports
	fmt.Println("Port forwarding is ready:")
	for _, port := range result.Ports {
		fmt.Printf("  %s:%d -> %d\n", port.Address, port.Local, port.Remote)
	}

	// Wait for stop signal
	<-stopChan

	// Stop port forwarding
	if err := client.PortForwardService.StopForwarding(result); err != nil {
		fmt.Printf("Error stopping port forward: %v\n", err)
	}

	return nil
}
