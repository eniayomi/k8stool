# Port Forward Commands

Commands for forwarding local ports to pods and deployments.

## Usage

```bash
k8stool port-forward (pod|deployment) NAME [LOCAL_PORT:]REMOTE_PORT [...[LOCAL_PORT_N:]REMOTE_PORT_N] [flags]
k8stool pf (pod|deployment) NAME [LOCAL_PORT:]REMOTE_PORT [...[LOCAL_PORT_N:]REMOTE_PORT_N] [flags]    # Short alias
```

### Flags
| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--namespace` | `-n` | Target namespace | Current namespace |
| `--interactive` | `-i` | Interactive mode | `false` |
| `--address` | - | Local address to bind to | `localhost` |
| `--protocol` | - | Protocol to use (tcp or udp) | `tcp` |

### Examples

Forward a single port:
```bash
k8stool port-forward pod nginx 8080:80
k8stool pf pod nginx 8080:80

# Forward to deployment
k8stool port-forward deployment nginx 8080:80
k8stool pf deploy nginx 8080:80
```

Forward multiple ports:
```bash
k8stool port-forward pod nginx 8080:80 9090:90
k8stool pf pod nginx 8080:80 9090:90
```

Use UDP protocol:
```bash
k8stool port-forward pod nginx 8080:80 --protocol=udp
```

Interactive mode:
```bash
k8stool port-forward -i
k8stool pf -i
```

## Interactive Mode Features

The interactive mode provides a guided experience with:

1. Resource Selection
   - Choose between pod or deployment
   - List of available resources
   - Arrow key navigation
   - Current resource highlighted

2. Port Selection
   - Shows available container ports
   - Displays container names and protocols
   - Easy selection with arrow keys

3. Local Port Configuration
   - Option to specify custom local port
   - Default to same as remote port
   - Validation for port number range (1-65535)

Interactive mode steps:
1. Select resource type (pod/deployment)
2. Choose specific resource from list
3. Select container port to forward
4. Optionally specify local port
5. Automatic port forward setup

## Port Format

The port format is:
```
[LOCAL_PORT:]REMOTE_PORT
```

Examples:
- `8080:80`: Forward local port 8080 to container port 80
- `80`: Forward local port 80 to container port 80 (same port)
- Multiple ports: `8080:80 9090:90`

## Related Commands

- [Pods](pods.md): List and manage pods
- [Deployments](deployments.md): List and manage deployments
- [Describe](describe.md): Get detailed resource information 