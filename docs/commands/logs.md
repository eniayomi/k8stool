# Log Commands

Commands for viewing and following container logs.

## View Logs

```bash
k8stool logs <pod-name> [flags]
```

### Flags
| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--follow` | `-f` | Stream logs in real-time | `false` |
| `--previous` | `-p` | Show previous container logs | `false` |
| `--container` | `-c` | Specific container name | - |
| `--tail` | - | Number of lines to show | `all` |
| `--since` | - | Show logs since duration | - |
| `--since-time` | - | Show logs since timestamp | - |
| `--namespace` | `-n` | Target namespace | `default` |

### Examples

View pod logs:
```bash
k8stool logs nginx-pod
```

Follow log output in real-time:
```bash
k8stool logs nginx-pod -f
k8stool logs nginx-pod --follow
```

View previous container logs:
```bash
k8stool logs nginx-pod -p
k8stool logs nginx-pod --previous
```

Show specific number of lines:
```bash
k8stool logs nginx-pod --tail 100
```

View logs from specific container:
```bash
k8stool logs nginx-pod -c nginx
k8stool logs nginx-pod --container nginx
```

Time-based filtering:
```bash
k8stool logs nginx-pod --since 1h      # Last hour
k8stool logs nginx-pod --since 5m      # Last 5 minutes
k8stool logs nginx-pod --since-time "2024-01-20T15:04:05Z"
```

## Duration Format

The `--since` flag accepts various duration formats:
- `h`: Hours (e.g., `1h`, `24h`)
- `m`: Minutes (e.g., `5m`, `30m`)
- `s`: Seconds (e.g., `30s`, `90s`)

## Timestamp Format

The `--since-time` flag accepts RFC3339 format:
```bash
YYYY-MM-DDTHH:MM:SSZ
# Example: 2024-01-20T15:04:05Z
```

## Related Commands

- [Pods](pods.md): List and manage pods
- [Events](events.md): View pod events
- [Describe](describe.md): Get detailed pod information 