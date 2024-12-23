# k8stool

A command-line tool for managing Kubernetes resources with enhanced features and user-friendly output.

## Installation

```bash
go install github.com/eniayomi/k8stool
```

## Features

### Pod Management
- ✅ List pods with detailed status
  ```bash
  k8stool get pods                    # List pods in default namespace
  k8stool get po                      # Short alias for pods
  k8stool get pods -n kube-system     # List pods in specific namespace
  k8stool get pods -A                 # List pods in all namespaces
  k8stool get pods --metrics          # Show CPU/Memory usage
  k8stool get pods -l app=nginx       # Filter by labels
  k8stool get pods -s Running         # Filter by status
  k8stool get pods --sort age         # Sort by age (oldest first)
  k8stool get pods --sort age --reverse # Sort by age (newest first)
  k8stool get pods --sort name        # Sort by name
  k8stool get pods --sort status      # Sort by status
  ```

### Pod Events
- ✅ Show pod events
  ```bash
  k8stool get events nginx-pod        # Show events for a pod
  k8stool get ev nginx-pod            # Short alias for events
  ```

### Pod Logs
- ✅ View and follow container logs
  ```bash
  k8stool logs nginx-pod              # View pod logs
  k8stool logs nginx-pod -f           # Follow log output
  k8stool logs nginx-pod -p           # Show previous container logs
  k8stool logs nginx-pod --tail 100   # Show last 100 lines
  k8stool logs nginx-pod -c nginx     # Show specific container logs
  k8stool logs nginx-pod --since 1h   # Show logs from last hour
  k8stool logs nginx-pod --since 5m   # Show logs from last 5 minutes
  k8stool logs nginx-pod --since-time "2024-01-20T15:04:05Z"  # Show logs since specific time
  ```

### Pod Exec
- ✅ Execute commands in containers
  ```bash
  k8stool exec nginx-pod              # Start a shell (tries bash, falls back to sh)
  k8stool exec nginx-pod bash         # Start bash shell
  k8stool exec nginx-pod sh           # Start sh shell
  k8stool exec nginx-pod -- ls /app   # Run specific command
  k8stool exec nginx-pod -c nginx     # Execute in specific container
  ```

### Deployment Management
- ✅ List deployments with status
  ```bash
  k8stool get deployments             # List deployments in default namespace
  k8stool get deploy                  # Short alias
  k8stool get deploy -n kube-system   # List in specific namespace
  k8stool get deploy -A               # List in all namespaces
  ```


## Usage Examples

### List Pods
```bash
# List pods in default namespace
k8stool get pods

# List pods in specific namespace
k8stool get pods -n kube-system
```

### Describe Pod
```bash
# Describe a pod
k8stool describe pod my-pod-name

# Describe a pod in specific namespace
k8stool describe pod my-pod-name -n my-namespace
```

### Pod Metrics
```bash
# Get pod metrics
k8stool metrics my-pod-name

# Get pod metrics in specific namespace
k8stool metrics my-pod-name -n my-namespace
```

## Output Examples

### Pod List
```
NAME                                READY    STATUS    RESTARTS    AGE        CONTROLLER
nginx-deployment-6799fc88d8-x7zv9   1/1      Running   0           2y18d      Deployment
redis-master-0                      1/1      Running   0           5d2h       StatefulSet
```

### Pod List with Metrics
```bash
NAME                     READY    STATUS    RESTARTS    CPU     MEMORY    AGE     CONTROLLER
nginx-deployment-x7zv9   1/1      Running   0          125m    64Mi      2y18d   Deployment
redis-master-0           1/1      Running   0          250m    128Mi     5d2h    StatefulSet
```

### Deployment List
```bash
NAME               READY   UP-TO-DATE   AVAILABLE   AGE
nginx-deployment   3/3     3            3           2y18d
redis-deployment   2/2     2            2           5d2h
```

### Pod Description
```
Pod Details:
  Name:            nginx-deployment-6799fc88d8-x7zv9
  Namespace:       default
  Node:            worker-1
  Status:          Running
  IP:              10.244.1.12
  Created:         2023-04-20 15:04:05
  Node-Selectors:  <none>

Containers:
  • nginx:
      Image:         nginx:1.14.2
      State:         Running
      Ready:         true
      Restart Count: 0

      Resources:
        Requests:
          CPU:    100m
          Memory: 128Mi
        Limits:
          CPU:    200m
          Memory: 256Mi
```

### Pod Metrics
```
Pod Metrics:
  Name:      nginx-deployment-6799fc88d8-x7zv9
  Namespace: default
  CPU:       125m
  Memory:    64Mi

Container Metrics:
  • nginx:
      CPU:    100m
      Memory: 45Mi
```

## Requirements

- Kubernetes cluster with metrics-server installed (for metrics feature)
- kubectl configured with cluster access
- Go 1.19 or later

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
