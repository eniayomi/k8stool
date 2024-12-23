# k8stool

A command-line tool for managing Kubernetes resources with enhanced features and user-friendly output.


## Installation

[Add installation instructions here]

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
  - Color-coded status (Running: green, Pending: yellow, Failed: red)
  - Namespace column only shows when listing across namespaces
  - Smart age formatting (2y3d, 3M15d, 5d6h, 2h30m, 45m, 30s)

### Pod/Deployment Events
- ✅ Show resource events
  ```bash
  k8stool get events pod nginx-pod         # Show events for a pod
  k8stool get ev pod nginx-pod             # Short alias for events
  k8stool get events deployment nginx      # Show events for a deployment
  k8stool get ev deploy nginx              # Short alias
  ```
  - Color-coded event types (Normal: green, Warning: yellow)
  - Supports both pods and deployments

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

### Port Forwarding
- ✅ Forward ports to pods
  ```bash
  k8stool port-forward pod nginx 8080:80          # Forward local 8080 to pod 80
  k8stool pf pod nginx 8080:80                    # Short alias
  k8stool port-forward pod nginx 8080:80 9090:90  # Multiple ports
  k8stool port-forward -i                         # Interactive mode
  ```
  - Interactive mode features:
    - Select pod from a list
    - View and select available container ports
    - Auto-setup port forwarding

### Deployment Management
- ✅ List deployments with status
  ```bash
  k8stool get deployments             # List deployments in default namespace
  k8stool get deploy                  # Short alias
  k8stool get deploy -n kube-system   # List in specific namespace
  k8stool get deploy -A               # List in all namespaces
  ```
  - Namespace column only shows when listing across namespaces
  - Detailed deployment information (replicas, status, age)
  - Smart age formatting (2y3d, 3M15d, 5d6h, 2h30m, 45m, 30s)
  - View deployment events with describe command
  ```bash
  k8stool describe deployment nginx   # Show detailed deployment info with events
  k8stool describe deploy nginx       # Short alias
  ```

## Requirements

- Kubernetes cluster with metrics-server installed (for metrics feature)
- kubectl configured with cluster access
- Go 1.19 or later

## Contributing

[contribution guidelines]

## License

This project is licensed under the MIT License - see the LICENSE file for details.
