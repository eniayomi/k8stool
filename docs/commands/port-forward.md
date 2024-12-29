# Port Forward Commands

Commands for forwarding local ports to pods.

## Forward Ports

```bash
k8stool port-forward <resource-type> <name> <ports...> [flags]
k8stool pf <resource-type> <name> <ports...> [flags]    # Short alias
```

### Flags
| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--namespace` | `-n` | Target namespace | `default` |
| `--interactive` | `-i` | Interactive mode | `false` |

### Examples

Forward a single port:
```bash
k8stool port-forward pod nginx 8080:80
k8stool pf pod nginx 8080:80
```

Forward multiple ports:
```bash
k8stool port-forward pod nginx 8080:80 9090:90
k8stool pf pod nginx 8080:80 9090:90
```

Interactive mode:
```bash
k8stool port-forward -i
k8stool pf -i
```

## Interactive Mode Features

The interactive mode provides:
1. Pod selection from a list
2. Available container port viewing
3. Auto-setup port forwarding

Interactive mode steps:
1. Lists all available pods
2. Shows current pod highlighted
3. Arrow key navigation
4. Enter to select pod
5. Shows available ports
6. Configures port forwarding automatically

## Port Format

The port format is:
```
<local-port>:<pod-port>
```

Examples:
- `8080:80`: Forward local 8080 to pod's 80
- `5432:5432`: Forward same port number
- Multiple ports: `8080:80 9090:90`

## Related Commands

- [Pods](pods.md): List and manage pods
- [Describe](describe.md): Get detailed pod information 