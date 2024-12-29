# Metrics Commands

Commands for viewing resource metrics. Requires metrics-server to be installed in the cluster.

## View Metrics

```bash
k8stool metrics <resource-type> [resource-name] [flags]
```

### Resource Types
- `pods` (or `po`): Pod metrics
- `nodes` (or `no`): Node metrics

### Flags
| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--namespace` | `-n` | Target namespace | `default` |
| `--all-namespaces` | `-A` | List across all namespaces | `false` |
| `--sort` | - | Sort by (cpu\|memory) | - |
| `--reverse` | - | Reverse sort order | `false` |

### Examples

View pod metrics:
```bash
k8stool metrics pods
k8stool metrics po
```

View specific pod metrics:
```bash
k8stool metrics pod nginx-pod
```

View node metrics:
```bash
k8stool metrics nodes
k8stool metrics no
```

Sort by resource usage:
```bash
k8stool metrics pods --sort cpu
k8stool metrics pods --sort memory --reverse
```

## Output

### Pod Metrics
- Pod name
- CPU usage
- Memory usage
- CPU request/limit
- Memory request/limit
- Namespace (when listing across namespaces)

Example pod metrics output:
```
NAME           CPU(cores)   MEMORY(bytes)   CPU%    MEMORY%
nginx-pod      10m         128Mi           5%      25%
redis-pod      50m         256Mi           25%     50%
```

### Node Metrics
- Node name
- CPU usage
- Memory usage
- CPU capacity
- Memory capacity
- CPU%
- Memory%

Example node metrics output:
```
NAME       CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
node-1     1200m       60%    4Gi             50%
node-2     800m        40%    3Gi             37%
```

## Prerequisites

The metrics command requires:
1. metrics-server installed in the cluster
2. Proper RBAC permissions to access metrics

## Related Commands

- [Pods](pods.md): List pods with metrics
- [Deployments](deployments.md): List deployments with metrics 