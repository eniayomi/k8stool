# Welcome to K8sTool

K8sTool is a versatile and efficient command-line tool designed to simplify Kubernetes resource management. It empowers users with intuitive commands and advanced features, making Kubernetes management seamless.

## Key Features

### Pod Management
- List pods with detailed status and filtering
- Color-coded status (Running: green, Pending: yellow, Failed: red)
- Smart age formatting (2y3d, 3M15d, 5d6h, 2h30m, 45m, 30s)
- Filter by labels, status, and namespace
- Sort by age, name, or status

### Pod/Deployment Events
- Show resource events with color coding
- Normal events in green, Warning events in yellow
- Support for both pods and deployments
- Easy event filtering and monitoring

### Pod Logs
- View and follow container logs
- Show previous container logs
- Tail specific number of lines
- Filter by time duration
- Container-specific log viewing
- Time-based log filtering

### Port Forwarding
- Forward ports to pods with simple commands
- Support for multiple port forwarding
- Interactive mode with:
  - Pod selection from list
  - Available container port viewing
  - Auto-setup port forwarding

### Deployment Management
- List deployments with status
- View across namespaces
- Detailed deployment information
- View deployment events

### Resource Metrics
- View resource utilization metrics
- Real-time CPU and Memory usage
- Support for both pods and deployments
- Requires metrics-server installed

### Context Management
- Switch between Kubernetes contexts
- Show current context
- Interactive context switching with:
  - List of available contexts
  - Current context highlighting
  - Arrow key navigation

### Namespace Management
- Switch between namespaces
- Show current namespace
- Interactive namespace switching with:
  - List of available namespaces
  - Current namespace highlighting
  - Arrow key navigation

## Getting Started

1. Check the [Prerequisites](prerequisites.md) for installation
2. Follow our [Installation Guide](installation.md)
3. Try the [Quick Start Tutorial](quick_start.md)

## Community and Support

- 👥 Join our [GitHub Discussions](https://github.com/eniayomi/k8stool/discussions)
- 🐛 Report issues on [GitHub Issues](https://github.com/eniayomi/k8stool/issues)
- 📖 Contribute to the [Documentation](contributing.md)

## Latest Release

Current Version: `v0.1.0`

Check out our [Release Notes](changelog.md) for the latest updates and improvements.