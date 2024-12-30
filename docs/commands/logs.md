# Logs Commands

Commands for viewing container logs from pods and deployments.

## Usage

```bash
k8stool logs (pod|deployment)/(name) [flags]
k8stool logs (pod|deployment) [name] [flags]
```

## Available Commands

### View Pod Logs
```bash
# Using slash format
k8stool logs pod/nginx-pod
k8stool logs po/nginx-pod

# Using space format
k8stool logs pod nginx-pod
k8stool logs po nginx-pod
```

### View Deployment Logs
```bash
# Using slash format
k8stool logs deployment/nginx
k8stool logs deploy/nginx

# Using space format
k8stool logs deployment nginx
k8stool logs deploy nginx
```

### Flags
| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--namespace` | `-n` | Target namespace | Current namespace |
| `--container` | `-c` | Print logs of this container | First container |
| `--follow` | `-f` | Follow log output | `false` |
| `--previous` | `-p` | Print logs of previous instance | `false` |
| `--tail` | `-t` | Lines of recent log file to display | `-1` (all) |
| `--since` | - | Show logs since duration (e.g. 1h, 5m, 30s) | - |
| `--since-time` | - | Show logs since specific time (RFC3339) | - |
| `--all-containers` | `-a` | Get logs from all containers (deployment only) | `false` |

### Examples

View pod logs:
```bash
# Basic pod logs
k8stool logs pod/nginx-pod
k8stool logs pod nginx-pod

# Follow log output
k8stool logs pod/nginx-pod -f

# View previous container logs
k8stool logs pod/nginx-pod -p

# Show specific number of lines
k8stool logs pod/nginx-pod --tail 100

# View logs from specific container
k8stool logs pod/nginx-pod -c nginx

# Time-based log filtering
k8stool logs pod/nginx-pod --since 1h
k8stool logs pod/nginx-pod --since 5m
k8stool logs pod/nginx-pod --since-time "2024-01-20T15:04:05Z"
```

View deployment logs:
```bash
# Basic deployment logs
k8stool logs deployment/nginx
k8stool logs deploy nginx

# View logs from all containers
k8stool logs deployment/nginx -a

# Follow logs from specific container
k8stool logs deployment/nginx -c nginx -f

# Show recent logs
k8stool logs deployment/nginx --tail 50
```

## Output

The output includes:
- Container logs with timestamps (if enabled)
- Color-coded log levels (if log format supports it)
- Real-time streaming (with --follow)
- Previous container logs (with --previous)

## Related Commands

- [Pods](pods.md): List and manage pods
- [Deployments](deployments.md): List and manage deployments
- [Describe](describe.md): Get detailed resource information 