# Namespace Commands

The namespace command allows you to view, switch, and manage Kubernetes namespaces.

## Usage

```bash
k8stool namespace [command]    # Long form
k8stool ns [command]          # Short form
```

## Available Commands

### Show Current Namespace
```bash
k8stool ns current
```
Shows the currently active namespace.

### List Namespaces
```bash
k8stool ns list
k8stool ns ls
```
Lists all available namespaces with their status. Output includes:
- Namespace name
- Status
- Active status (*)

### Switch Namespace
Direct switch:
```bash
k8stool ns switch <namespace-name>
```
Switch to a different namespace directly.

Interactive switch:
```bash
k8stool ns switch
k8stool ns switch -i
```
Opens an interactive menu to select and switch namespaces. Features:
- Shows current namespace with "(current)" suffix
- Uses colored output for better visibility
- Shows 10 namespaces at a time
- Uses emoji indicators for selection

## Interactive Mode Features

The interactive mode provides:
1. List of all available namespaces
2. Current namespace highlighted
3. Arrow key navigation
4. Enter to select namespace

Example output:
```
Select namespace:
  default
👉 production (current)
  kube-system
  monitoring
```

## Output

The output includes:
- Namespace name
- Status (color-coded)
- Active status (*)

Example output when showing current namespace:
```
Current namespace: production
```

## Related Commands

- [Context](context.md): Manage context selection