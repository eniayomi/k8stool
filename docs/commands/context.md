# Context Commands

Commands for managing Kubernetes contexts.

## Context Operations

```bash
k8stool ctx [command]
k8stool context [command]    # Long form
```

### Commands
| Command | Description |
|---------|-------------|
| `current` | Show current context |
| `<context-name>` | Switch to specific context |
| `switch` | Interactive context selection |

### Examples

Show current context:
```bash
k8stool ctx current
k8stool context current
```

Switch to specific context:
```bash
k8stool ctx production
k8stool context production
```

Interactive context switching:
```bash
k8stool ctx switch
k8stool context switch
```

## Interactive Mode Features

The interactive mode provides:
1. List of all available contexts
2. Current context highlighted
3. Arrow key navigation
4. Enter to select context

Example output:
```
Select Kubernetes context:
  dev-cluster
> production
  staging
  minikube
```

## Output

The output includes:
- Context name
- Cluster name
- User
- Namespace (if set)

Example output when showing current context:
```
Current context: production
Cluster: prod-cluster-1
User: admin
Namespace: default
```

## Related Commands

- [Namespace](namespace.md): Manage namespace selection 