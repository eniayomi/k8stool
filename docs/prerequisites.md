# Prerequisites

Before using K8sTool, ensure you have the following:

## Required
- A Kubernetes cluster
- cluster access configured
- Proper permissions to interact with your cluster

## Optional
- **metrics-server** installed in your cluster
  - Only needed if you plan to use resource metrics features (CPU and Memory usage)
  - Required for commands using `--metrics` flag
  - Required for `k8stool metrics` commands

