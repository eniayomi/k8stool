# Context Commands

The context command allows you to view, switch, and manage Kubernetes contexts. These commands work without requiring cluster access, as they only interact with your kubeconfig file.

## Usage

```bash
k8stool context [command]    # Long form
k8stool ctx [command]       # Short form
```

## Available Commands

### Show Current Context
```bash
k8stool ctx current
```
Shows details about the current context including:
- Context name
- Cluster
- User
- Namespace (if set)

### List Contexts
```bash
k8stool ctx list
k8stool ctx ls
```
Lists all available contexts with their details. The current context is marked with an asterisk (*).
Output includes:
- Context name
- Cluster name
- User
- Namespace (if set)
- Active status (*)

### Switch Context
Direct switch:
```bash
k8stool ctx switch <context-name>
```
Switch to a specific context directly.

Interactive switch:
```bash
k8stool ctx switch
k8stool ctx switch -i
```
Opens an interactive menu to select and switch contexts. Features:
- Shows current context with "(current)" suffix
- Uses colored output for better visibility
- Shows 10 contexts at a time
- Uses emoji indicators for selection

## Interactive Mode Features

The interactive mode provides:
1. List of all available contexts
2. Current context highlighted
3. Arrow key navigation
4. Enter to select context

Example output:
```
Select context:
  dev-cluster
👉 production (current)
  staging
  minikube
```

## Output

The output includes:
- Context name
- Cluster name
- User
- Namespace (if set)
- Active status (*)

Example output when showing current context:
```
Current context: production
Cluster: prod-cluster-1
User: admin
Namespace: default
```

## Related Commands

- [Namespace](namespace.md): Manage namespace selection 