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
  k8stool get pods
  k8stool get pods -n <namespace>
  ```
- ✅ Describe pods with comprehensive information
  ```bash
  k8stool describe pod <pod-name>
  k8stool describe pod <pod-name> -n <namespace>
  ```
- ✅ Show pod metrics (CPU/Memory usage)
  ```bash
  k8stool metrics <pod-name>
  k8stool metrics <pod-name> -n <namespace>
  ```
- ⬜ Execute commands in pods
  ```bash
  # Coming soon
  k8stool exec <pod-name> -- <command>
  ```
- ⬜ Stream pod logs
  ```bash
  # Coming soon
  k8stool logs <pod-name>
  k8stool logs <pod-name> -f  # follow logs
  ```
- ⬜ Port forwarding
  ```bash
  # Coming soon
  k8stool port-forward <pod-name> <local-port>:<pod-port>
  ```
- ⬜ Watch pod status changes
  ```bash
  # Coming soon
  k8stool get pods --watch
  ```
- ⬜ Delete/force delete pods
  ```bash
  # Coming soon
  k8stool delete pod <pod-name>
  k8stool delete pod <pod-name> --force
  ```

### Pod Information Display
- ✅ Colored status output
- ✅ Resource requests and limits
- ✅ Volume and mount information
- ✅ Node selector information
- ✅ Container details
- ⬜ Init container status
- ⬜ Pod conditions
- ⬜ Pod QoS class
- ⬜ Security context
- ⬜ Pod priority class

### Filtering and Sorting
- ⬜ Filter pods by labels
  ```bash
  # Coming soon
  k8stool get pods -l app=nginx
  ```
- ⬜ Filter pods by status
  ```bash
  # Coming soon
  k8stool get pods --status=Running
  ```
- ⬜ Sort by age/status/name
  ```bash
  # Coming soon
  k8stool get pods --sort-by=age
  k8stool get pods --sort-by=status
  k8stool get pods --sort-by=name
  ```
- ⬜ Namespace filtering
  ```bash
  # Coming soon
  k8stool get pods --all-namespaces
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
nginx-deployment-6799fc88d8-x7zv9   1/1      Running   0           3d         Deployment
redis-master-0                      1/1      Running   0           11d14h     StatefulSet
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
