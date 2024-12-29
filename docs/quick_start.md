# Quick Start Guide

This guide will help you get started with K8sTool quickly.

## Prerequisites

Before installing K8sTool, ensure you have:
1. Kubernetes cluster configured
2. `kubectl` installed and configured
3. Optional: metrics-server installed (for resource metrics)

See [Prerequisites](prerequisites.md) for detailed requirements.

## Installation

```bash
# Download latest release
curl -LO https://github.com/k8stool/k8stool/releases/latest/download/k8stool

# Make executable
chmod +x k8stool

# Move to PATH
sudo mv k8stool /usr/local/bin/
```

## Basic Usage

### View Pods
```bash
k8stool get pods
k8stool get pods -A    # All namespaces
```
See [Pods](commands/pods.md) for more details.

### View Logs
```bash
k8stool logs nginx-pod
k8stool logs nginx-pod -f    # Follow logs
```
See [Logs](commands/logs.md) for more details.

### Forward Ports
```bash
k8stool port-forward pod nginx 8080:80
k8stool pf -i    # Interactive mode
```
See [Port Forward](commands/port-forward.md) for more details.

### View Metrics
```bash
k8stool metrics pods
k8stool metrics nodes
```
See [Metrics](commands/metrics.md) for more details.

### Switch Context/Namespace
```bash
k8stool ctx switch    # Interactive context switch
k8stool ns -i        # Interactive namespace switch
```
See [Context](commands/context.md) and [Namespace](commands/namespace.md) for more details.

## Next Steps

- Read the command documentation for detailed usage:
  - [Pods](commands/pods.md)
  - [Logs](commands/logs.md)
  - [Deployments](commands/deployments.md)
  - [Events](commands/events.md)
  - [Port Forward](commands/port-forward.md)
  - [Context](commands/context.md)
  - [Namespace](commands/namespace.md)
  - [Metrics](commands/metrics.md)
  - [Describe](commands/describe.md)
  - [Exec](commands/exec.md)
- Check the [FAQ](faq.md) for common questions
- Join our community for support
