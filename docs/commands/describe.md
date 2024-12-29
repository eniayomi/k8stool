# Describe Commands

Commands for getting detailed information about Kubernetes resources.

## Describe Resources

```bash
k8stool describe <resource-type> <resource-name> [flags]
k8stool desc <resource-type> <resource-name> [flags]    # Short alias
```

### Resource Types
- `pods` (or `po`): Pod details
- `deployments` (or `deploy`): Deployment details
- `services` (or `svc`): Service details
- `nodes` (or `no`): Node details

### Flags
| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--namespace` | `-n` | Target namespace | `default` |

### Examples

Describe a pod:
```bash
k8stool describe pod nginx-pod
k8stool desc po nginx-pod
```

Describe a deployment:
```bash
k8stool describe deployment nginx
k8stool desc deploy nginx
```

Describe a service:
```bash
k8stool describe service nginx-service
k8stool desc svc nginx-service
```

Describe a node:
```bash
k8stool describe node node-1
k8stool desc no node-1
```

## Output

The output includes detailed information about the resource, formatted for readability with color-coding for important fields.

### Pod Description
- Basic Information
  - Name, Namespace, Node
  - Labels, Annotations
- Status
  - Phase, Conditions
  - IP Addresses
- Containers
  - Image, Ports
  - Resource Requests/Limits
  - Environment Variables
- Events
  - Recent events related to the pod

### Deployment Description
- Basic Information
  - Name, Namespace
  - Labels, Annotations
- Spec
  - Replicas
  - Strategy
  - Selector
- Status
  - Available Replicas
  - Conditions
- Events
  - Recent events related to the deployment

### Service Description
- Basic Information
  - Name, Namespace
  - Type, IP
- Ports
  - Port mappings
  - Target ports
- Endpoints
  - Pod IPs and Ports
- Events
  - Recent events related to the service

### Node Description
- Basic Information
  - Name, Labels
  - Architecture, OS
- Status
  - Conditions
  - Capacity
  - Allocatable Resources
- System Info
  - Kernel Version
  - Container Runtime
- Events
  - Recent events related to the node

## Related Commands

- [Events](events.md): View resource events
- [Pods](pods.md): List and manage pods
- [Deployments](deployments.md): List and manage deployments 