# Event Commands

Commands for viewing Kubernetes events.

## View Events

```bash
k8stool get events <resource-type> <resource-name> [flags]
k8stool get ev <resource-type> <resource-name> [flags]    # Short alias
```

### Flags
| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--namespace` | `-n` | Target namespace | `default` |
| `--type` | - | Filter by event type (Normal/Warning) | - |

### Examples

View pod events:
```bash
k8stool get events pod nginx-pod
k8stool get ev pod nginx-pod
```

View deployment events:
```bash
k8stool get events deployment nginx
k8stool get ev deploy nginx
```

Filter by event type:
```bash
k8stool get events pod nginx-pod --type Warning
k8stool get events deployment nginx --type Normal
```

## Output

The output includes:

- Last Seen (smart formatting)
- Type (color-coded)
  - Normal: Green
  - Warning: Yellow
- Reason
- Object
- Message

Example output:
```
LAST SEEN   TYPE     REASON      OBJECT                MESSAGE
2m          Normal   Scheduled   pod/nginx-pod         Successfully assigned default/nginx-pod to node-1
30s         Warning  Failed      pod/nginx-pod         Error: ImagePullBackOff
```

## Related Commands

- [Pods](pods.md): List and manage pods
- [Deployments](deployments.md): List and manage deployments
- [Describe](describe.md): Get detailed resource information 