# Namespace Commands

Commands for managing Kubernetes namespace selection.

## Namespace Operations

```bash
k8stool ns [command]
k8stool namespace [command]    # Long form
```

### Commands
| Command | Description |
|---------|-------------|
| (no args) | Show current namespace |
| `<namespace-name>` | Switch to specific namespace |
| `-i, --interactive` | Interactive namespace selection |

### Examples

Show current namespace:
```bash
k8stool ns
k8stool namespace
```

Switch to specific namespace:
```bash
k8stool ns production
k8stool namespace production
```

Interactive namespace selection:
```bash
k8stool ns -i
k8stool namespace --interactive
```

## Interactive Mode Features

The interactive mode provides:
1. List of all available namespaces
2. Current namespace highlighted
3. Arrow key navigation
4. Enter to select namespace

Example output:
```
Select Kubernetes namespace:
  default
  kube-system
> production
  staging
```

## Output

The output includes:
- Namespace name
- Status (Active)
- Age

Example output when showing current namespace:
```
Current namespace: production
Status: Active
Age: 30d
```

## Related Commands

- [Context](context.md): Manage context selection
- [Pods](pods.md): List pods in namespace
- [Deployments](deployments.md): List deployments in namespace 