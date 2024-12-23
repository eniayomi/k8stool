# k8stool

A command-line tool for managing Kubernetes resources with enhanced features and user-friendly output.

## Installation

### Using Homebrew (macOS/Linux)
```bash
brew tap eniayomi/k8stool
brew install k8stool
```

### Using Binary Releases
Download the latest binary for your platform from the [releases page](https://github.com/eniayomi/k8stool/releases).

#### Linux
```bash
curl -LO https://github.com/eniayomi/k8stool/releases/download/v0.0.5/k8stool_Linux_arm64.tar.gz
tar xzf k8stool_Linux_arm64.tar.gz
cd k8stool_Linux_arm64
chmod +x k8stool
sudo mv k8stool /usr/local/bin/k8stool
```

#### macOS
```bash
curl -LO https://github.com/eniayomi/k8stool/releases/download/v0.0.5/k8stool_Darwin_arm64.tar.gz
tar xzf k8stool_Darwin_arm64.tar.gz
cd k8stool_Darwin_arm64
chmod +x k8stool
sudo mv k8stool /usr/local/bin/k8stool
```

#### Windows
```bash
# Download the zip file
curl -LO https://github.com/eniayomi/k8stool/releases/download/v0.0.5/k8stool_Windows_x86_64.zip

# Extract the executable
Expand-Archive -Path k8stool_Windows_x86_64.zip -DestinationPath k8stool

# Move to a directory in your PATH (example using C:\Windows\System32)
move k8stool\k8stool.exe C:\Windows\System32\k8stool.exe
```

Alternatively, you can:
1. Download `k8stool_Windows_x86_64.zip` from the [releases page](https://github.com/eniayomi/k8stool/releases) or [directly](https://github.com/eniayomi/k8stool/releases/download/v0.0.5/k8stool_Windows_x86_64.zip)
2. Extract the ZIP file
3. Move `k8stool.exe` to a directory in your PATH

### From Source
```bash
git clone https://github.com/eniayomi/k8stool.git
cd k8stool
go install ./cmd/k8stool
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

### Resource Metrics
- ✅ View resource utilization metrics
  ```bash
  k8stool metrics pods                # Show metrics for all pods
  k8stool metrics nodes               # Show metrics for all nodes
  k8stool metrics <pod-name>          # Show metrics for specific pod
  k8stool get pods --metrics          # Show pod metrics in pod list
  k8stool get deploy --metrics        # Show aggregated deployment metrics
  ```
  - Real-time CPU and Memory usage
  - Supports both pods and deployments
  - Requires metrics-server installed

### Context Management
- ✅ Switch between Kubernetes contexts
  ```bash
  k8stool ctx current                # Show curret context
  k8stool context current            # Show curret context
  k8stool ctx <context-name>         # Switch to specific context
  k8stool context <context-name>     # Switch to specific context (long form)
  k8stool ctx switch                 # Interactive mode
  k8stool context switch             # Interactive mode
  ```
  - Interactive mode:
    - Lists all available contexts
    - Shows current context highlighted
    - Arrow key navigation
    - Enter to select context
  - Shows current context
  - Easy context switching
  - Displays associated clusters

### Namespace Management
- ✅ Switch between namespaces
  ```bash
  k8stool ns                         # show current namespace
  k8stool namespace                  # show current namespace
  k8stool ns <namespace-name>        # Switch to specific namespace
  k8stool namespace <namespace-name> # Switch to specific namespace (long form)
  k8stool ns -i                      # Switch to specific namespace
  k8stool namespace -i               # Switch to specific namespace (long form)
  ```
  - Interactive mode:
    - Lists all available namespaces
    - Shows current namespace highlighted
    - Arrow key navigation
    - Enter to select namespace
  - Shows current namespace
  - Easy namespace switching

## Requirements

- Kubernetes cluster with metrics-server installed (for metrics feature)
- kubectl configured with cluster access
- Go 1.21 or later

## Contributing

[contribution guidelines]

## License

This project is licensed under the MIT License - see the LICENSE file for details.
