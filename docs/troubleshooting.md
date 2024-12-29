# Troubleshooting Guide

This guide helps you diagnose and resolve common issues you might encounter while using K8sTool.

## Common Issues

### Connection Problems

#### Unable to Connect to Cluster
```bash
Error: Unable to connect to the server: dial tcp: lookup kubernetes.default.svc: no such host
```

**Solutions:**
1. Verify your kubeconfig file is correctly configured
2. Check network connectivity to the cluster
3. Ensure cluster certificates are valid and not expired

#### Context Switching Issues
```bash
Error: no context exists with name "invalid-context"
```

**Solutions:**
1. List available contexts: `k8stool config get-contexts`
2. Verify context name spelling
3. Check if kubeconfig file is properly formatted

### Authentication Issues

#### Invalid Credentials
```bash
Error: invalid credentials: token has expired
```

**Solutions:**
1. Refresh your authentication token
2. Check if your certificates are valid
3. Verify RBAC permissions for your user

#### Permission Denied
```bash
Error: pods is forbidden: User "user" cannot list resource "pods" in API group "" in namespace "default"
```

**Solutions:**
1. Verify your RBAC permissions
2. Check namespace access rights
3. Request necessary permissions from cluster admin

### Resource Management

#### Pod Creation Failures
```bash
Error: 0/3 nodes are available: 3 Insufficient memory
```

**Solutions:**
1. Check cluster resource availability
2. Verify resource requests and limits
3. Consider scaling cluster or optimizing resource usage

#### Deployment Issues
```bash
Error: deployment "app" exceeded its progress deadline
```

**Solutions:**
1. Check pod events: `k8stool describe pod <pod-name>`
2. Verify container image availability
3. Check resource constraints and pod scheduling

## Diagnostic Commands

### System Status
```bash
# Check tool version
k8stool version

# Verify configuration
k8stool config view

# Test cluster connectivity
k8stool cluster-info
```

### Resource Inspection
```bash
# Get detailed pod information
k8stool describe pod <pod-name> -n <namespace>

# Check pod logs
k8stool logs <pod-name> -n <namespace>

# View events
k8stool get events -n <namespace>
```


## Getting Help

If you're still experiencing issues:

1. **Documentation**
   - Check the [FAQ](faq.md)

2. **Community Support**
   - [GitHub Issues](https://github.com/eniayomi/k8stool/issues)
   - [GitHub Discussions](https://github.com/eniayomi/k8stool/discussions)

3. **Bug Reports**
   - Include tool version: `k8stool version`
   - Describe reproduction steps 