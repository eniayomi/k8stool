# Deployment Commands

Commands for managing and viewing Kubernetes deployments.

## List Deployments

```bash
k8stool get deployments [flags]
k8stool get deploy [flags]    # Short alias
```

### Flags
| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--namespace` | `-n` | Target namespace | `default` |
| `--all-namespaces` | `-A` | List across all namespaces | `false` |
| `--metrics` | - | Show CPU/Memory usage | `false` |

### Examples

List deployments in default namespace:
```bash
k8stool get deployments
k8stool get deploy
```

List deployments in specific namespace:
```bash
k8stool get deploy -n kube-system
```

List deployments across all namespaces:
```bash
k8stool get deploy -A
k8stool get deploy --all-namespaces
```

Show resource usage:
```bash
k8stool get deploy --metrics
```

## Output

The output includes:

- Deployment name
- Ready replicas (ready/total)
- Up-to-date replicas
- Available replicas
- Age (smart formatting)
- CPU usage (if --metrics flag is used)
- Memory usage (if --metrics flag is used)
- Namespace (when listing across namespaces)

Example output:
```
NAME               READY   UP-TO-DATE   AVAILABLE   AGE    CPU    MEMORY
nginx-deployment   3/3     3            3           5d6h   30m    384Mi
redis-deployment   2/2     2            2           2h30m  100m   512Mi
```

## Related Commands

- [Pods](pods.md): List and manage pods
- [Events](events.md): View deployment events
- [Describe](describe.md): Get detailed deployment information 