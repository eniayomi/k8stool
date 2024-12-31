package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	k8s "k8stool/internal/k8s/client"
	"k8stool/internal/k8s/deployments"
	"k8stool/internal/k8s/pods"

	"github.com/spf13/cobra"
)

// resourceTypeAliases maps shorthand names to their full resource types
var resourceTypeAliases = map[string]string{
	"po":          "pod",
	"pods":        "pod",
	"deploy":      "deployment",
	"deployments": "deployment",
}

func getDescribeCmd() *cobra.Command {
	var namespace string

	cmd := &cobra.Command{
		Use:     "describe TYPE NAME",
		Aliases: []string{"desc"},
		Short:   "Show details of a specific resource",
		Long: `Show detailed information about a specific Kubernetes resource.

Supported resource types:
  - pod (po, pods)
  - deployment (deploy, deployments)

Examples:
  # Describe a pod
  k8stool describe pod my-pod

  # Describe a deployment
  k8stool describe deploy my-deployment

  # Describe a pod in a specific namespace
  k8stool describe pod my-pod --namespace my-namespace`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			resourceType := strings.ToLower(args[0])
			name := args[1]

			// Map resource type alias to actual type
			if actualType, ok := resourceTypeAliases[resourceType]; ok {
				resourceType = actualType
			}

			// Use provided namespace or fallback to current context's namespace
			ns := namespace
			if ns == "" {
				currentCtx, err := client.GetCurrentContext()
				if err != nil {
					return err
				}
				ns = currentCtx.Namespace
			}

			switch resourceType {
			case "pod":
				details, err := client.PodService.Describe(ns, name)
				if err != nil {
					return err
				}
				return printPodDetails(details)
			case "deployment":
				details, err := client.DeploymentService.Describe(ns, name)
				if err != nil {
					return err
				}
				return printDeploymentDetails(details)
			default:
				return fmt.Errorf("unsupported resource type: %s", resourceType)
			}
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Namespace of the resource")
	return cmd
}

func printPodDetails(details *pods.PodDetails) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Basic Info
	fmt.Fprintf(w, "Name:\t%s\n", details.Name)
	fmt.Fprintf(w, "Namespace:\t%s\n", details.Namespace)
	if details.HasPriority {
		fmt.Fprintf(w, "Priority:\t%v\n", details.HasPriority)
	}
	fmt.Fprintf(w, "Service Account:\t%s\n", details.ServiceAccount)
	fmt.Fprintf(w, "Node:\t%s\n", details.Node)
	if details.NodeIP != "" {
		fmt.Fprintf(w, "Node IP:\t%s\n", details.NodeIP)
	}
	fmt.Fprintf(w, "Start Time:\t%s\n", details.StartTime.Format("Mon, 02 Jan 2006 15:04:05 -0700"))

	// Labels and Annotations
	if len(details.Labels) > 0 {
		fmt.Fprintf(w, "Labels:\t\n")
		for k, v := range details.Labels {
			fmt.Fprintf(w, "  %s=%s\n", k, v)
		}
	}
	if len(details.Annotations) > 0 {
		fmt.Fprintf(w, "Annotations:\t\n")
		for k, v := range details.Annotations {
			fmt.Fprintf(w, "  %s=%s\n", k, v)
		}
	}

	// Status and IP
	fmt.Fprintf(w, "Status:\t%s\n", details.Status)
	fmt.Fprintf(w, "IP:\t%s\n", details.IP)
	if len(details.IPs) > 0 {
		fmt.Fprintf(w, "IPs:\n")
		for _, ip := range details.IPs {
			fmt.Fprintf(w, "  IP:\t%s\n", ip)
		}
	}

	if details.ControlledBy != "" {
		fmt.Fprintf(w, "Controlled By:\t%s\n", details.ControlledBy)
	}

	// Containers
	fmt.Fprintf(w, "Containers:\n")
	for _, c := range details.Containers {
		fmt.Fprintf(w, "  %s:\n", c.Name)
		if c.ContainerID != "" {
			fmt.Fprintf(w, "    Container ID:\t%s\n", c.ContainerID)
		}
		fmt.Fprintf(w, "    Image:\t%s\n", c.Image)
		if c.ImageID != "" {
			fmt.Fprintf(w, "    Image ID:\t%s\n", c.ImageID)
		}

		// Ports
		if len(c.Ports) > 0 {
			for _, p := range c.Ports {
				fmt.Fprintf(w, "    Port:\t%d/%s\n", p.ContainerPort, p.Protocol)
				if p.HostPort > 0 {
					fmt.Fprintf(w, "    Host Port:\t%d/%s\n", p.HostPort, p.Protocol)
				}
			}
		}

		// Container State
		fmt.Fprintf(w, "    State:\t%s\n", c.State.Status)
		if !c.State.Started.IsZero() {
			fmt.Fprintf(w, "      Started:\t%s\n", c.State.Started.Format("Mon, 02 Jan 2006 15:04:05 -0700"))
		}
		fmt.Fprintf(w, "    Ready:\t%v\n", c.Ready)
		fmt.Fprintf(w, "    Restart Count:\t%d\n", c.RestartCount)

		// Resources
		if c.Resources.Requests.CPU != "" || c.Resources.Requests.Memory != "" {
			fmt.Fprintf(w, "    Requests:\n")
			if c.Resources.Requests.CPU != "" {
				fmt.Fprintf(w, "      cpu:\t%s\n", c.Resources.Requests.CPU)
			}
			if c.Resources.Requests.Memory != "" {
				fmt.Fprintf(w, "      memory:\t%s\n", c.Resources.Requests.Memory)
			}
		}

		// Readiness Probe
		if c.ReadinessProbe != nil {
			fmt.Fprintf(w, "    Readiness:\t%s :%d delay=%s timeout=%s period=%s #success=%d #failure=%d\n",
				c.ReadinessProbe.Type,
				c.ReadinessProbe.Port,
				c.ReadinessProbe.Delay,
				c.ReadinessProbe.Timeout,
				c.ReadinessProbe.Period,
				c.ReadinessProbe.SuccessThreshold,
				c.ReadinessProbe.FailureThreshold,
			)
		}

		// Environment Variables
		if len(c.EnvFrom) > 0 {
			fmt.Fprintf(w, "    Environment Variables from:\n")
			for _, env := range c.EnvFrom {
				fmt.Fprintf(w, "      %s\t%s\tOptional: %v\n", env.Name, env.Type, env.Optional)
			}
		}
		if len(c.Env) > 0 {
			fmt.Fprintf(w, "    Environment:\n")
			for _, env := range c.Env {
				if env.Value != "" {
					fmt.Fprintf(w, "      %s:\t%s\n", env.Name, env.Value)
				} else if env.ValueFrom != "" {
					fmt.Fprintf(w, "      %s:\t%s\n", env.Name, env.ValueFrom)
				}
			}
		}

		// Volume Mounts
		if len(c.VolumeMounts) > 0 {
			fmt.Fprintf(w, "    Mounts:\n")
			for _, vm := range c.VolumeMounts {
				fmt.Fprintf(w, "      %s from %s (ro: %v)\n", vm.MountPath, vm.Name, vm.ReadOnly)
			}
		}
	}

	// Conditions
	if len(details.Conditions) > 0 {
		fmt.Fprintf(w, "Conditions:\n")
		fmt.Fprintf(w, "  Type\tStatus\n")
		fmt.Fprintf(w, "  ----\t------\n")
		for _, c := range details.Conditions {
			fmt.Fprintf(w, "  %s\t%s\n", c.Type, c.Status)
		}
	}

	// Volumes
	if len(details.Volumes) > 0 {
		fmt.Fprintf(w, "Volumes:\n")
		for _, v := range details.Volumes {
			fmt.Fprintf(w, "  %s:\n", v.Name)
			fmt.Fprintf(w, "    Type:\t%s\n", v.Type)
			if v.TokenExpirationSeconds > 0 {
				fmt.Fprintf(w, "    TokenExpirationSeconds:\t%d\n", v.TokenExpirationSeconds)
			}
			if v.ConfigMapName != "" {
				fmt.Fprintf(w, "    ConfigMapName:\t%s\n", v.ConfigMapName)
			}
			if v.ConfigMapOptional != nil {
				fmt.Fprintf(w, "    ConfigMapOptional:\t%v\n", *v.ConfigMapOptional)
			}
			if v.DownwardAPI {
				fmt.Fprintf(w, "    DownwardAPI:\t%v\n", v.DownwardAPI)
			}
		}
	}

	// QoS Class
	if details.QoSClass != "" {
		fmt.Fprintf(w, "QoS Class:\t%s\n", details.QoSClass)
	}

	// Node Selectors
	if len(details.NodeSelector) > 0 {
		fmt.Fprintf(w, "Node-Selectors:\t")
		selectors := make([]string, 0)
		for k, v := range details.NodeSelector {
			selectors = append(selectors, fmt.Sprintf("%s=%s", k, v))
		}
		fmt.Fprintf(w, "%s\n", strings.Join(selectors, ","))
	} else {
		fmt.Fprintf(w, "Node-Selectors:\t<none>\n")
	}

	// Tolerations
	if len(details.Tolerations) > 0 {
		fmt.Fprintf(w, "Tolerations:\t")
		for i, t := range details.Tolerations {
			if i > 0 {
				fmt.Fprintf(w, "\t\t")
			}
			fmt.Fprintf(w, "%s", t.Key)
			if t.Operator != "" {
				fmt.Fprintf(w, ":%s", t.Operator)
			}
			if t.Value != "" {
				fmt.Fprintf(w, " %s", t.Value)
			}
			if t.Effect != "" {
				fmt.Fprintf(w, " %s", t.Effect)
			}
			if t.TolerationSeconds != nil {
				fmt.Fprintf(w, " for %ds", *t.TolerationSeconds)
			}
			fmt.Fprintf(w, "\n")
		}
	}

	// Events
	if len(details.Events) > 0 {
		fmt.Fprintf(w, "Events:\n")
		fmt.Fprintf(w, "Type\tReason\tAge\tFrom\tMessage\n")
		fmt.Fprintf(w, "----\t------\t---\t----\t-------\n")
		for _, e := range details.Events {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				e.Type,
				e.Reason,
				e.Age.Round(time.Second),
				e.From,
				e.Message,
			)
		}
	} else {
		fmt.Fprintf(w, "Events:\t<none>\n")
	}

	return nil
}

func printDeploymentDetails(details *deployments.DeploymentDetails) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Basic Info
	fmt.Fprintf(w, "Name:\t%s\n", details.Name)
	fmt.Fprintf(w, "Namespace:\t%s\n", details.Namespace)
	fmt.Fprintf(w, "CreationTimestamp:\t%s\n", details.CreationTime.Format("Mon, 02 Jan 2006 15:04:05 -0700"))

	// Labels and Annotations
	if len(details.Labels) > 0 {
		fmt.Fprintf(w, "Labels:\t\n")
		for k, v := range details.Labels {
			fmt.Fprintf(w, "  %s=%s\n", k, v)
		}
	}
	if len(details.Annotations) > 0 {
		fmt.Fprintf(w, "Annotations:\t\n")
		for k, v := range details.Annotations {
			fmt.Fprintf(w, "  %s=%s\n", k, v)
		}
	}

	// Selector
	if len(details.Selector) > 0 {
		fmt.Fprintf(w, "Selector:\t")
		selectorPairs := make([]string, 0)
		for k, v := range details.Selector {
			selectorPairs = append(selectorPairs, fmt.Sprintf("%s=%s", k, v))
		}
		fmt.Fprintf(w, "%s\n", strings.Join(selectorPairs, ","))
	}

	// Replicas Status
	fmt.Fprintf(w, "Replicas:\t%d desired | %d updated | %d total | %d available | %d unavailable\n",
		details.Replicas,
		details.UpdatedReplicas,
		details.Replicas,
		details.AvailableReplicas,
		details.Replicas-details.AvailableReplicas)

	// Strategy
	fmt.Fprintf(w, "StrategyType:\t%s\n", details.Strategy)
	fmt.Fprintf(w, "MinReadySeconds:\t%d\n", details.MinReadySeconds)
	if details.RollingUpdateStrategy != nil {
		fmt.Fprintf(w, "RollingUpdateStrategy:\t%d max unavailable, %d max surge\n",
			details.RollingUpdateStrategy.MaxUnavailable,
			details.RollingUpdateStrategy.MaxSurge)
	}

	// Pod Template
	fmt.Fprintf(w, "Pod Template:\n")
	fmt.Fprintf(w, "  Labels:\t\n")
	for k, v := range details.TemplateLabels {
		fmt.Fprintf(w, "    %s=%s\n", k, v)
	}
	if len(details.TemplateAnnotations) > 0 {
		fmt.Fprintf(w, "  Annotations:\t\n")
		for k, v := range details.TemplateAnnotations {
			fmt.Fprintf(w, "    %s=%s\n", k, v)
		}
	}

	// Containers
	fmt.Fprintf(w, "  Containers:\n")
	for _, c := range details.Containers {
		fmt.Fprintf(w, "   %s:\n", c.Name)
		fmt.Fprintf(w, "    Image:\t%s\n", c.Image)
		if len(c.Ports) > 0 {
			fmt.Fprintf(w, "    Port:\t%d/%s\n", c.Ports[0].ContainerPort, c.Ports[0].Protocol)
			if c.Ports[0].HostPort > 0 {
				fmt.Fprintf(w, "    Host Port:\t%d/%s\n", c.Ports[0].HostPort, c.Ports[0].Protocol)
			}
		}

		// Check for resource requests and limits
		hasRequests := c.Resources.Requests.CPU != "" || c.Resources.Requests.Memory != ""
		hasLimits := c.Resources.Limits.CPU != "" || c.Resources.Limits.Memory != ""

		if hasRequests {
			fmt.Fprintf(w, "    Requests:\n")
			if c.Resources.Requests.CPU != "" {
				fmt.Fprintf(w, "      cpu:\t%s\n", c.Resources.Requests.CPU)
			}
			if c.Resources.Requests.Memory != "" {
				fmt.Fprintf(w, "      memory:\t%s\n", c.Resources.Requests.Memory)
			}
		}
		if hasLimits {
			fmt.Fprintf(w, "    Limits:\n")
			if c.Resources.Limits.CPU != "" {
				fmt.Fprintf(w, "      cpu:\t%s\n", c.Resources.Limits.CPU)
			}
			if c.Resources.Limits.Memory != "" {
				fmt.Fprintf(w, "      memory:\t%s\n", c.Resources.Limits.Memory)
			}
		}

		if len(c.VolumeMounts) > 0 {
			fmt.Fprintf(w, "    Mounts:\n")
			for _, vm := range c.VolumeMounts {
				fmt.Fprintf(w, "      %s from %s (ro: %v)\n", vm.MountPath, vm.Name, vm.ReadOnly)
			}
		}
	}

	// Environment Variables
	if len(details.Environment) > 0 {
		fmt.Fprintf(w, "    Environment:\n")
		for _, env := range details.Environment {
			if env.Value != "" {
				fmt.Fprintf(w, "      %s:\t%s\n", env.Name, env.Value)
			} else if env.ValueFrom != "" {
				fmt.Fprintf(w, "      %s:\t%s\n", env.Name, env.ValueFrom)
			}
		}
	}
	if len(details.EnvironmentFrom) > 0 {
		fmt.Fprintf(w, "    Environment From:\n")
		for _, env := range details.EnvironmentFrom {
			fmt.Fprintf(w, "      %s\t%s\tOptional: %v\n", env.Name, env.Type, env.Optional)
		}
	}

	// Conditions
	if len(details.Conditions) > 0 {
		fmt.Fprintf(w, "Conditions:\n")
		fmt.Fprintf(w, "  Type\tStatus\tReason\n")
		fmt.Fprintf(w, "  ----\t------\t------\n")
		for _, c := range details.Conditions {
			fmt.Fprintf(w, "  %s\t%s\t%s\n", c.Type, c.Status, c.Reason)
		}
	}

	// Old ReplicaSets
	if len(details.OldReplicaSets) > 0 {
		fmt.Fprintf(w, "OldReplicaSets:\t")
		for i, rs := range details.OldReplicaSets {
			if i > 0 {
				fmt.Fprintf(w, ", ")
			}
			fmt.Fprintf(w, "%s (%s replicas created)", rs.Name, rs.ReplicasCreated)
		}
		fmt.Fprintf(w, "\n")
	}
	if details.NewReplicaSet.Name != "" {
		fmt.Fprintf(w, "NewReplicaSet:\t%s (%s replicas created)\n", details.NewReplicaSet.Name, details.NewReplicaSet.ReplicasCreated)
	}

	// Events
	if len(details.Events) > 0 {
		fmt.Fprintf(w, "Events:\n")
		fmt.Fprintf(w, "Type\tReason\tAge\tFrom\tMessage\n")
		fmt.Fprintf(w, "----\t------\t---\t----\t-------\n")
		for _, e := range details.Events {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				e.Type,
				e.Reason,
				e.Age.Round(time.Second),
				e.From,
				e.Message,
			)
		}
	}

	return nil
}
