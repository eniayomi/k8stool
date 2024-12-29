# Frequently Asked Questions

## General

### What is K8sTool?
K8sTool is a command-line tool that simplifies common Kubernetes operations with an intuitive interface and enhanced output formatting.

### What are the main features?
- Pod management (see [Pods](commands/pods.md))
- Log viewing (see [Logs](commands/logs.md))
- Port forwarding (see [Port Forward](commands/port-forward.md))
- Resource metrics (see [Metrics](commands/metrics.md))
- Context/namespace management (see [Context](commands/context.md) and [Namespace](commands/namespace.md))

### How do I get started?
Check out our [Quick Start Guide](quick_start.md) for installation and basic usage.

## Installation

### What are the prerequisites?
See [Prerequisites](prerequisites.md) for detailed requirements.

### How do I install K8sTool?
Follow the installation steps in our [Quick Start Guide](quick_start.md).

## Usage

### How do I view pod logs?
See the [Logs](commands/logs.md) command documentation.

### How do I forward ports?
See the [Port Forward](commands/port-forward.md) command documentation.

### How do I view resource metrics?
See the [Metrics](commands/metrics.md) command documentation. Note that this requires metrics-server to be installed in your cluster.

### How do I switch contexts/namespaces?
See the [Context](commands/context.md) and [Namespace](commands/namespace.md) command documentation.

## Troubleshooting

### Why can't I see resource metrics?
Make sure metrics-server is installed in your cluster. See [Prerequisites](prerequisites.md) for details.

### How do I report issues?
Please open an issue on our GitHub repository with:
1. K8sTool version (`k8stool --version`)
2. Kubernetes version
3. Steps to reproduce
4. Expected vs actual behavior
