# Exec Commands

Commands for executing commands in containers.

## Execute Commands

```bash
k8stool exec <pod-name> [flags] -- <command> [args...]
```

### Flags
| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--namespace` | `-n` | Target namespace | `default` |
| `--container` | `-c` | Target container name | First container |
| `--interactive` | `-i` | Keep stdin open | `false` |
| `--tty` | `-t` | Allocate pseudo-TTY | `false` |

### Examples

Execute a command:
```bash
k8stool exec nginx-pod -- ls /app
```

Interactive shell:
```bash
k8stool exec nginx-pod -it -- /bin/bash
k8stool exec nginx-pod -it -- /bin/sh
```

Specify container in pod:
```bash
k8stool exec nginx-pod -c nginx -- ps aux
```

Run command with arguments:
```bash
k8stool exec nginx-pod -- curl localhost:8080/health
```

## Interactive Mode

When using the `-it` flags together:
1. Allocates a pseudo-TTY (`-t`)
2. Keeps stdin open (`-i`)
3. Provides an interactive shell session

Common interactive use cases:
- Debugging container issues
- Checking file contents
- Testing network connectivity
- Monitoring processes

## Command Execution

The command format after `--` is:
```bash
<command> [args...]
```

Examples:
- `ls /app`: List directory contents
- `/bin/bash`: Start interactive shell
- `ps aux`: List processes
- `cat /etc/config`: View file contents
- `curl localhost:8080`: Test HTTP endpoint

## Related Commands

- [Pods](pods.md): List and manage pods
- [Logs](logs.md): View container logs
- [Describe](describe.md): Get detailed pod information 