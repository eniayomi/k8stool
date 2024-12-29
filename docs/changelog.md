# Changelog

All notable changes to K8sTool will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.0] - 2024-12-28

### Added
- Pod Management
  - List pods with detailed status
  - Color-coded status (Running: green, Pending: yellow, Failed: red)
  - Smart age formatting
  - Filter by labels and status
  - Sort by age, name, and status
  - Show CPU/Memory usage with metrics

- Pod/Deployment Events
  - Show resource events
  - Color-coded event types
  - Support for both pods and deployments

- Pod Logs
  - View and follow container logs
  - Show previous container logs
  - Tail specific number of lines
  - Time-based filtering
  - Container-specific logs

- Port Forwarding
  - Forward ports to pods
  - Multiple port forwarding
  - Interactive mode with pod selection

- Deployment Management
  - List deployments with status
  - Cross-namespace viewing
  - Detailed deployment information

- Resource Metrics
  - Pod and deployment metrics
  - Real-time CPU and Memory usage

- Context Management
  - Switch between contexts
  - Interactive context switching
  - Current context display

- Namespace Management
  - Switch between namespaces
  - Interactive namespace switching
  - Current namespace display

### Notes
- Initial release with core functionality
- Requires metrics-server for metrics features 