# Installation Guide

This guide covers all available methods to install K8sTool on different platforms.

## Prerequisites

Before installing K8sTool, ensure you have:
- Kubernetes cluster access
- `kubectl` installed and configured
- Proper permissions to interact with your cluster

## Installation Methods

### 1. Using Homebrew (macOS/Linux)

The recommended method for macOS and Linux users:

```bash
# Add the K8sTool tap
brew tap eniayomi/k8stool

# Install K8sTool
brew install k8stool
```

### 2. Binary Releases

Download pre-compiled binaries for your platform from our [releases page](https://github.com/eniayomi/k8stool/releases).

#### Linux (arm64)
```bash
# Download the latest release
curl -LO https://github.com/eniayomi/k8stool/releases/download/v0.0.5/k8stool_Linux_arm64.tar.gz

# Extract the archive
tar xzf k8stool_Linux_arm64.tar.gz
cd k8stool_Linux_arm64

# Make it executable and move to PATH
chmod +x k8stool
sudo mv k8stool /usr/local/bin/k8stool
```

#### macOS (arm64)
```bash
# Download the latest release
curl -LO https://github.com/eniayomi/k8stool/releases/download/v0.0.5/k8stool_Darwin_arm64.tar.gz

# Extract the archive
tar xzf k8stool_Darwin_arm64.tar.gz
cd k8stool_Darwin_arm64

# Make it executable and move to PATH
chmod +x k8stool
sudo mv k8stool /usr/local/bin/k8stool
```

#### Windows
Option 1 - Using PowerShell:
```powershell
# Download the zip file
curl -LO https://github.com/eniayomi/k8stool/releases/download/v0.0.5/k8stool_Windows_x86_64.zip

# Extract the executable
Expand-Archive -Path k8stool_Windows_x86_64.zip -DestinationPath k8stool

# Move to a directory in your PATH
move k8stool\k8stool.exe C:\Windows\System32\k8stool.exe
```

Option 2 - Manual Installation:
1. Download `k8stool_Windows_x86_64.zip` from the [releases page](https://github.com/eniayomi/k8stool/releases)
2. Extract the ZIP file
3. Move `k8stool.exe` to a directory in your PATH

### 3. Building from Source

For developers who want to build from source:

```bash
# Clone the repository
git clone https://github.com/eniayomi/k8stool.git

# Change to the project directory
cd k8stool

# Install using Go
go install ./cmd/k8stool
```

Requirements for building from source:
- Go 1.19 or later
- Git

## Verifying the Installation

After installation, verify K8sTool is working correctly:

```bash
# Check the version
k8stool version

# Test cluster connectivity
k8stool cluster-info
```

## Configuration

After installation:

1. K8sTool will automatically use your existing kubeconfig file (`~/.kube/config`)

## Upgrading

### Homebrew
```bash
brew upgrade k8stool
```

### Binary Installation
Download and install the new version following the same steps as the initial installation.

### From Source
```bash
cd k8stool
git pull
go install ./cmd/k8stool
```

## Troubleshooting Installation

If you encounter issues during installation:

1. **Permission Denied**
   ```bash
   sudo chown -R $USER /usr/local/bin
   ```

2. **Binary Not Found**
   - Ensure the installation directory is in your PATH
   - Try logging out and back in

3. **Version Mismatch**
   - Remove old versions before upgrading
   - Clear Go module cache if building from source

For more help, see our [Troubleshooting Guide](troubleshooting.md)
