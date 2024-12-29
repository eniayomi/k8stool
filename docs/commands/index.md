# Command Reference

K8sTool provides a comprehensive set of commands for managing Kubernetes resources. The commands are organized into the following categories:

## Resource Management

Commands for managing Kubernetes resources:

- [Pods](pods.md): List, filter, and manage pods
- [Deployments](deployments.md): Work with deployments
- [Events](events.md): View and monitor resource events
- [Describe](describe.md): Get detailed information about resources

## Operations

Commands for operational tasks:

- [Logs](logs.md): View and follow container logs
- [Port Forward](port-forward.md): Forward ports to pods
- [Exec](exec.md): Execute commands in containers

## Cluster Management

Commands for managing cluster access:

- [Context](context.md): Switch between Kubernetes contexts
- [Namespace](namespace.md): Manage namespace selection

## Monitoring

Commands for monitoring resources:

- [Metrics](metrics.md): View resource utilization metrics

## Global Flags

These flags can be used with any command:

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--namespace` | `-n` | Target namespace | `default` |
| `--all-namespaces` | `-A` | List across all namespaces | `false` |
| `--help` | `-h` | Show help for command | - |

## Output Features

- Color-coded status for resources
  - Pod status (Running: green, Pending: yellow, Failed: red)
  - Event types (Normal: green, Warning: yellow)
- Smart age formatting (2y3d, 3M15d, 5d6h, 2h30m, 45m, 30s)
- Interactive mode for context and namespace switching 