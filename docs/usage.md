# Usage Guide

This guide provides practical examples for common use cases and scenarios you might encounter while using K8sTool.

## Pod Management

### Viewing and Filtering Pods
```bash
# List pods in default namespace
k8stool get pods

# List pods in all namespaces
k8stool get pods -A

# Filter pods by label
k8stool get pods -l app=nginx

# Filter by status
k8stool get pods -s Running

# Sort pods by different criteria
k8stool get pods --sort age         # Sort by age
k8stool get pods --sort name        # Sort by name
k8stool get pods --sort status      # Sort by status
k8stool get pods --sort age --reverse # Newest first
```

### Working with Pod Logs
```bash
# View pod logs
k8stool logs nginx-pod

# Follow log output in real-time
k8stool logs nginx-pod -f

# View previous container logs
k8stool logs nginx-pod -p

# Show specific number of lines
k8stool logs nginx-pod --tail 100

# View logs from specific container
k8stool logs nginx-pod -c nginx

# Time-based log filtering
k8stool logs nginx-pod --since 1h
k8stool logs nginx-pod --since 5m
k8stool logs nginx-pod --since-time "2024-01-20T15:04:05Z"
```

## Resource Monitoring

### Pod Metrics
```bash
# View pod metrics
k8stool get pods --metrics

# View metrics for specific pods
k8stool metrics pods
k8stool metrics <pod-name>

# View node metrics
k8stool metrics nodes
```

### Event Monitoring
```bash
# View pod events
k8stool get events pod nginx-pod
k8stool get ev pod nginx-pod    # Short form

# View deployment events
k8stool get events deployment nginx
k8stool get ev deploy nginx     # Short form
```

## Deployment Management

### Basic Operations
```bash
# List deployments
k8stool get deployments
k8stool get deploy    # Short form

# List deployments in specific namespace
k8stool get deploy -n kube-system

# List deployments across all namespaces
k8stool get deploy -A

# View deployment details
k8stool describe deployment nginx
k8stool describe deploy nginx    # Short form
```

## Port Forwarding

### Basic Port Forwarding
```bash
# Forward single port
k8stool port-forward pod nginx 8080:80
k8stool pf pod nginx 8080:80    # Short form

# Forward multiple ports
k8stool port-forward pod nginx 8080:80 9090:90

# Interactive port forwarding
k8stool port-forward -i
```

## Context and Namespace Management

### Working with Contexts
```bash
# View current context
k8stool ctx current
k8stool context current    # Long form

# Switch context
k8stool ctx my-context
k8stool context my-context    # Long form

# Interactive context switching
k8stool ctx switch
k8stool context switch    # Long form
```

### Managing Namespaces
```bash
# View current namespace
k8stool ns
k8stool namespace    # Long form

# Switch namespace
k8stool ns production
k8stool namespace production    # Long form

# Interactive namespace switching
k8stool ns -i
k8stool namespace -i    # Long form
```

## Common Workflows

### Application Monitoring
```bash
# View application pods with metrics
k8stool get pods -l app=myapp --metrics

# Monitor pod events
k8stool get events pod myapp-pod

# Follow application logs
k8stool logs -l app=myapp -f
```

### Development Workflow
```bash
# Forward application ports
k8stool port-forward pod myapp 8080:80

# Switch context to development
k8stool ctx dev-cluster

# Switch to development namespace
k8stool ns development

# Monitor application
k8stool get pods -l app=myapp -w
```

## Best Practices

1. **Resource Organization**
   - Use consistent labels for filtering
   - Leverage namespaces for isolation
   - Use short aliases for common commands

2. **Monitoring**
   - Regularly check pod metrics
   - Monitor important events
   - Use real-time log following for debugging

3. **Context Management**
   - Use interactive mode for context switching
   - Verify current context before operations
   - Keep namespaces organized

4. **Port Forwarding**
   - Use interactive mode for better UX
   - Forward multiple ports when needed
   - Clean up port forwards when done
