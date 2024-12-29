# Pod Commands

Commands for managing and viewing Kubernetes pods.

## List Pods

```bash
k8stool get pods [flags]
k8stool get po [flags]      # Short alias
```

### Flags
| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--namespace` | `-n` | Target namespace | `default` |
| `--all-namespaces` | `-A` | List across all namespaces | `false` |
| `--selector` | `-l` | Label selector | - |
| `--status` | `-s` | Filter by status | - |
| `--sort` | - | Sort by (age\|name\|status) | - |
| `--reverse` | - | Reverse sort order | `false` |
| `--metrics` | - | Show CPU/Memory usage | `false` |

### Examples

List pods in default namespace:
```bash
k8stool get pods
```

List pods across all namespaces:
```bash
k8stool get pods -A
k8stool get pods --all-namespaces
```

Filter pods by label:
```bash
k8stool get pods -l app=nginx
k8stool get pods --selector app=nginx,env=prod
```

Filter by status:
```bash
k8stool get pods -s Running
k8stool get pods --status Pending
```

Sort pods:
```bash
k8stool get pods --sort age         # Sort by age (oldest first)
k8stool get pods --sort age --reverse # Sort by age (newest first)
k8stool get pods --sort name        # Sort by name
k8stool get pods --sort status      # Sort by status
```

Show resource usage:
```bash
k8stool get pods --metrics          # Show CPU/Memory usage
```

## Output

The output includes:

- Pod name
- Ready status (running containers/total containers)
- Status (color-coded)
  - Running: Green
  - Pending: Yellow
  - Failed: Red
- Age (smart formatting)
- CPU usage (if --metrics flag is used)
- Memory usage (if --metrics flag is used)
- Namespace (when listing across namespaces)

Example output:
```
NAME                     READY   STATUS    AGE    CPU    MEMORY
nginx-6799fc88d8-abc12   1/1     Running   5d6h   10m    128Mi
redis-7d8594697c-def34   1/1     Running   2h30m  50m    256Mi
```

## Related Commands

- [Logs](logs.md): View pod logs
- [Port Forward](port-forward.md): Forward ports to pods
- [Exec](exec.md): Execute commands in pods
- [Events](events.md): View pod events
- [Describe](describe.md): Get detailed pod information 