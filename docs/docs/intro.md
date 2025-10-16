---
sidebar_position: 1
---

# Getting Started with goUpdater

Welcome to goUpdater! This guide will help you get started with installing and using goUpdater, a powerful CLI tool for managing Go programming language installations on Linux systems.

## What is goUpdater?

goUpdater is a command-line tool that simplifies the installation, updating, and management of Go programming language versions on Linux systems. It automates the complex process of downloading, installing, and verifying Go installations, making it easy for developers to keep their Go environment up-to-date.

The tool works with official Go distributions from [go.dev](https://go.dev) and follows standard Go installation conventions, defaulting to installing Go in `/usr/local/go`.

## Key Benefits

- **Automated Management**: Handles the entire Go installation lifecycle with single commands
- **Security First**: Performs SHA256 checksum verification on all downloads
- **Platform Detection**: Automatically detects your system architecture and downloads the appropriate Go version
- **Privilege Handling**: Securely escalates privileges when needed using syscall-based sudo execution
- **Zero Dependencies**: Download and run the binary immediately without additional setup
- **Error Handling**: Comprehensive error handling with clear, actionable messages

## System Requirements

Before installing goUpdater, ensure your system meets these requirements:

- **Operating System**: Linux (Ubuntu, CentOS, Fedora, etc.)
- **Go Version**: Go 1.25.2 or later (only required if building from source)
- **Internet Connection**: Required for downloading Go archives
- **Permissions**: Ability to run `sudo` for system-wide installations (installs to `/usr/local/go` by default)
- **Storage**: Sufficient disk space for Go installation (approximately 500MB)

## Installation

goUpdater can be installed using two methods. The recommended approach is downloading the pre-built binary, which requires no additional setup.

### Method 1: Download Pre-built Binary (Recommended)

This is the easiest way to get started. Download the latest release directly from GitHub and run it immediately.

1. **Download the binary**:

   ```bash
   wget https://github.com/nicholas-fedor/goUpdater/releases/latest/download/goUpdater-linux-amd64
   ```

2. **Make it executable**:

   ```bash
   chmod +x goUpdater-linux-amd64
   ```

3. **Run it directly** (no installation required):

   ```bash
   ./goUpdater-linux-amd64 --help
   ```

The binary is self-contained and handles all privilege escalation automatically when needed.

### Method 2: Build from Source

If you prefer to build goUpdater yourself or need to customize it:

1. **Clone the repository**:

   ```bash
   git clone https://github.com/nicholas-fedor/goUpdater.git
   cd goUpdater
   ```

2. **Build the binary**:

   ```bash
   go build -o goUpdater .
   ```

3. **(Optional) Install system-wide**:

   ```bash
   sudo mv goUpdater /usr/local/bin/
   ```

After building, you can run `./goUpdater` from the project directory or `/usr/local/bin/goUpdater` if you installed it system-wide.

## Quick Start Guide

Once installed, you can start using goUpdater immediately. Here's how to perform common tasks:

### Update Go to the Latest Version

The simplest way to update Go is using the `update` command, which handles everything automatically:

```bash
sudo ./goUpdater-linux-amd64 update
```

This command will:

- Download the latest stable Go version
- Verify the download integrity
- Uninstall any existing Go installation
- Install the new version
- Verify the installation

**Expected output**:

```text
Successfully updated Go in: /usr/local/go
```

### Install a Specific Go Version

For more control, you can download and install manually:

1. **Download the latest Go archive**:

   ```bash
   ./goUpdater-linux-amd64 download
   ```

   **Output**:

   ```text
   Successfully downloaded Go archive to: /tmp/go1.25.3.linux-amd64.tar.gz
   SHA256 checksum: abc123...
   ```

2. **Install from the downloaded archive**:

   ```bash
   sudo ./goUpdater-linux-amd64 install /tmp/go1.25.3.linux-amd64.tar.gz
   ```

   **Output**:

   ```text
   Successfully installed Go to: /usr/local/go
   ```

### Custom Installation Directory

To install Go in a different location (useful for user-specific installations):

```bash
sudo ./goUpdater-linux-amd64 update --install-dir /opt/go
```

Remember to update your PATH:

```bash
export PATH=/opt/go/bin:$PATH
```

## Verification Steps

After installation, verify that Go is working correctly:

1. **Verify the installation**:

   ```bash
   ./goUpdater-linux-amd64 verify
   ```

   **Expected output**:

   ```text
   Go installation verified. Version: go1.25.3 linux/amd64
   ```

2. **Test Go directly**:

   ```bash
   /usr/local/go/bin/go version
   ```

   You should see output similar to:

   ```text
   go version go1.25.3 linux/amd64
   ```

3. **Check your PATH**:

   Ensure `/usr/local/go/bin` is in your PATH by running:

   ```bash
   echo $PATH
   ```

   If it's not there, add it to your shell profile (e.g., `~/.bashrc` or `~/.zshrc`):

   ```bash
   export PATH=/usr/local/go/bin:$PATH
   ```

## Next Steps

Now that you have goUpdater installed and working:

- **Learn More**: Explore the [Command Reference](commands.md) for detailed information about all available commands
- **Advanced Usage**: Check out [Workflow Examples](examples.md) for common usage patterns
- **Troubleshooting**: Visit the [Troubleshooting Guide](troubleshooting.md) if you encounter any issues
- **Contributing**: See the [GitHub Repository](https://github.com/nicholas-fedor/goUpdater) for contribution guidelines

For the latest updates and releases, visit the [goUpdater GitHub Repository](https://github.com/nicholas-fedor/goUpdater).
