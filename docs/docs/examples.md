# Usage Examples and Workflows

This document provides comprehensive, real-world examples and workflows for using goUpdater to manage Go installations on Linux systems. All examples are based on actual goUpdater functionality and include complete command sequences with expected outputs.

## Command Relationships and Workflows

### Full Update Process

The `update` command provides the simplest way to update Go, handling the entire process automatically:

```bash
sudo goUpdater update
```

This command internally performs:

1. Downloads the latest Go archive
2. Verifies the download with checksum
3. Uninstalls the current Go installation
4. Installs the new version
5. Verifies the installation

### Manual Installation Process

For more control, perform steps individually:

```bash
goUpdater download
```

```bash
sudo goUpdater install /tmp/go{version}.linux-amd64.tar.gz
```

```bash
goUpdater verify
```

### Custom Installation Directory

To install Go in a custom location:

```bash
sudo goUpdater update --install-dir /opt/go
```

```bash
sudo goUpdater install go1.25.3.linux-amd64.tar.gz --install-dir /opt/go
```

Ensure to update your PATH accordingly:

```bash
export PATH=/opt/go/bin:$PATH
```

## Common Workflows

### Routine Updates

The most common use case is keeping Go updated to the latest stable version. The `update` command handles the entire process automatically.

#### Automated Update Process

```bash
# Update Go to the latest stable version (requires sudo for system directories)
sudo goUpdater update
```

**Expected Output:**

```bash
Successfully updated Go in: /usr/local/go
```

**What happens internally:**

1. Downloads the latest Go archive for your platform
2. Verifies the download using SHA256 checksum
3. Uninstalls the current Go installation
4. Installs the new version
5. Verifies the installation is working

#### Update with Verbose Logging

```bash
# Enable detailed logging to see each step
sudo goUpdater --verbose update
```

**Expected Output:**

```bash
[INFO] Starting Go update process
[INFO] Detected platform: linux/amd64
[INFO] Downloading Go archive from https://go.dev/dl/go1.25.3.linux-amd64.tar.gz
[INFO] SHA256 checksum verification passed
[INFO] Uninstalling existing Go installation from /usr/local/go
[INFO] Installing Go to /usr/local/go
[INFO] Installation completed successfully
[INFO] Verifying Go installation
Successfully updated Go in: /usr/local/go
```

### Fresh Installations

For systems without Go installed, or when setting up new environments.

#### First-Time Installation

```bash
# Download the latest Go archive
goUpdater download

# Install from the downloaded archive (note the path from download output)
sudo goUpdater install /tmp/go1.25.3.linux-amd64.tar.gz

# Verify the installation
goUpdater verify
```

**Expected Output:**

```bash
# Download command
Successfully downloaded Go archive to: /tmp/go1.25.3.linux-amd64.tar.gz
SHA256 checksum: abc123def456...

# Install command
Successfully installed Go to: /usr/local/go

# Verify command
Go installation verified. Version: go1.25.3 linux/amd64
```

#### Automated Fresh Installation

```bash
# Use update command with auto-install flag (installs if Go not present)
sudo goUpdater update --auto-install
```

### Version-Specific Installations

While goUpdater focuses on the latest stable version, you can install specific versions by downloading the appropriate archive manually.

#### Installing a Specific Version

```bash
# Download a specific version archive manually
wget https://go.dev/dl/go1.24.1.linux-amd64.tar.gz

# Install the specific version
sudo goUpdater install go1.24.1.linux-amd64.tar.gz

# Verify the installation
goUpdater verify
```

**Expected Output:**

```bash
# Install command
Successfully installed Go to: /usr/local/go

# Verify command
Go installation verified. Version: go1.24.1 linux/amd64
```

## Advanced Scenarios

### Custom Directories

Install Go in non-standard locations for development environments, testing, or user-specific installations.

#### User-Specific Installation

```bash
# Install Go in user's home directory (no sudo required)
goUpdater update --install-dir ~/go

# Update PATH to include the custom installation
export PATH=$HOME/go/bin:$PATH

# Verify the installation
goUpdater verify --install-dir ~/go
```

**Expected Output:**

```bash
Successfully updated Go in: /home/user/go
Go installation verified. Version: go1.25.3 linux/amd64
```

#### Development Environment Setup

```bash
# Create a development directory structure
mkdir -p ~/dev/go-versions

# Install multiple versions for testing
sudo goUpdater install go1.24.1.linux-amd64.tar.gz --install-dir ~/dev/go-versions/go1.24.1
sudo goUpdater install go1.25.3.linux-amd64.tar.gz --install-dir ~/dev/go-versions/go1.25.3

# Switch between versions by updating PATH
export PATH=$HOME/dev/go-versions/go1.25.3/bin:$PATH
go version
```

### Offline Installations

For air-gapped systems or when network access is limited.

#### Prepare Archives for Offline Installation

```bash
# On a connected system, download archives for offline use
goUpdater download
# Archive saved to /tmp/go1.25.3.linux-amd64.tar.gz

# Transfer the archive to the offline system via USB drive, etc.
cp /tmp/go1.25.3.linux-amd64.tar.gz /media/usb/

# On the offline system, install from the transferred archive
sudo goUpdater install /media/usb/go1.25.3.linux-amd64.tar.gz
goUpdater verify
```

**Expected Output:**

```bash
Successfully installed Go to: /usr/local/go
Go installation verified. Version: go1.25.3 linux/amd64
```

#### Offline Update Process

```bash
# Download on connected system
goUpdater download

# Transfer archive to offline system
scp /tmp/go1.25.3.linux-amd64.tar.gz offline-server:/tmp/

# On offline system: uninstall current, install new
sudo goUpdater uninstall
sudo goUpdater install /tmp/go1.25.3.linux-amd64.tar.gz
goUpdater verify
```

### CI/CD Integration

Integrate goUpdater into automated build and deployment pipelines.

#### GitHub Actions Workflow

```yaml
name: Go Update and Test
on:
  schedule:
    # Run weekly on Mondays
    - cron: '0 2 * * 1'

jobs:
  update-go:
    runs-on: ubuntu-latest
    steps:
      - name: Download goUpdater
        run: |
          wget https://github.com/nicholas-fedor/goUpdater/releases/latest/download/goUpdater-linux-amd64
          chmod +x goUpdater-linux-amd64
          sudo mv goUpdater-linux-amd64 /usr/local/bin/goUpdater

      - name: Update Go
        run: sudo goUpdater update

      - name: Verify Go installation
        run: goUpdater verify

      - name: Run tests with updated Go
        run: go test ./...
```

#### Docker Multi-Stage Build

```dockerfile
# Multi-stage Docker build with goUpdater
FROM ubuntu:20.04 AS downloader
RUN apt-get update && apt-get install -y wget
RUN wget https://github.com/nicholas-fedor/goUpdater/releases/latest/download/goUpdater-linux-amd64
RUN chmod +x goUpdater-linux-amd64

FROM golang:1.24 AS builder
COPY --from=downloader goUpdater-linux-amd64 /usr/local/bin/goUpdater
RUN goUpdater download
RUN sudo goUpdater install /tmp/go*.tar.gz --install-dir /usr/local/go
RUN go version
# Continue with your build process
```

#### Jenkins Pipeline

```groovy
pipeline {
    agent any
    stages {
        stage('Setup Go Environment') {
            steps {
                sh '''
                    # Download and setup goUpdater
                    wget https://github.com/nicholas-fedor/goUpdater/releases/latest/download/goUpdater-linux-amd64
                    chmod +x goUpdater-linux-amd64
                    sudo mv goUpdater-linux-amd64 /usr/local/bin/goUpdater

                    # Update Go
                    sudo goUpdater update

                    # Verify installation
                    goUpdater verify
                '''
            }
        }
        stage('Build') {
            steps {
                sh 'go build ./cmd/...'
            }
        }
        stage('Test') {
            steps {
                sh 'go test ./...'
            }
        }
    }
}
```

## Troubleshooting

### Common Issues and Solutions

#### Permission Denied Errors

**Problem:** Installation fails with permission errors.

**Solution:** Use sudo for system directories.

```bash
# Instead of:
goUpdater update
# Use:
sudo goUpdater update
```

**Expected Output After Fix:**

```bash
Successfully updated Go in: /usr/local/go
```

#### Network Connection Issues

**Problem:** Download fails due to network problems.

**Solution:** Check connectivity and retry.

```bash
# Check network connectivity
ping -c 3 go.dev

# Retry download
goUpdater download
```

**Expected Output:**

```bash
PING go.dev (142.250.72.142) 56(84) bytes of data.
64 bytes from lga25s65-in-f14.1e100.net (142.250.72.142): icmp_seq=1 ttl=118 time=12.3 ms
# ... successful pings

Successfully downloaded Go archive to: /tmp/go1.25.3.linux-amd64.tar.gz
SHA256 checksum: abc123def456...
```

#### Insufficient Disk Space

**Problem:** Installation fails due to lack of space.

**Solution:** Check available space and clean up if needed.

```bash
# Check disk space
df -h /usr/local

# If space is low, clean up old archives
rm -f /tmp/go*.tar.gz

# Then retry installation
sudo goUpdater update
```

#### Corrupted Archive Files

**Problem:** SHA256 checksum verification fails.

**Solution:** Re-download the archive.

```bash
# Remove corrupted archive
rm /tmp/go1.25.3.linux-amd64.tar.gz

# Download fresh copy
goUpdater download

# Retry installation
sudo goUpdater install /tmp/go1.25.3.linux-amd64.tar.gz
```

#### PATH Not Updated

**Problem:** `go` command not found after installation.

**Solution:** Update PATH environment variable.

```bash
# Add Go to PATH (add to ~/.bashrc for persistence)
export PATH=/usr/local/go/bin:$PATH

# Test the fix
go version
```

**Expected Output:**

```bash
go version go1.25.3 linux/amd64
```

#### Existing Go Processes Blocking Uninstall

**Problem:** Uninstall fails because Go processes are running.

**Solution:** Stop running processes before uninstalling.

```bash
# Find running Go processes
pgrep -f go

# Stop the processes (replace PID with actual process ID)
kill 12345

# Then uninstall
sudo goUpdater uninstall
```

## Best Practices and Tips

### Command-Specific Best Practices

#### 1. Use the Update Command for Routine Updates

```bash
sudo goUpdater update
```

This ensures a complete, verified update process.

#### 2. Verify Installations After Manual Operations

Always verify after installation or update:

```bash
goUpdater verify
```

#### 3. Use Custom Directories for Testing

For development or testing, use custom installation directories:

```bash
sudo goUpdater update --install-dir /opt/go-dev
```

#### 4. Backup Before Major Changes

While goUpdater doesn't create backups, manually backup existing installations if needed:

```bash
sudo cp -r /usr/local/go /usr/local/go.backup
```

#### 5. Check Available Space

Ensure sufficient disk space before operations:

```bash
df -h /usr/local
```

#### 6. Use Verbose Mode for Troubleshooting

Enable verbose logging for detailed operation information:

```bash
goUpdater --verbose download
```

#### 7. Update PATH After Custom Installations

When using custom installation directories, update your PATH:

```bash
export PATH=/opt/go/bin:$PATH
```

#### 8. Regular Verification

Periodically verify your Go installation:

```bash
@weekly goUpdater verify
```

### 1. Regular Updates

Keep Go updated for security patches and new features:

```bash
# Set up weekly cron job for automatic updates
echo "0 2 * * 1 /usr/local/bin/goUpdater update" | sudo tee /etc/cron.d/go-update
```

### 2. Backup Before Major Changes

While goUpdater doesn't create backups, manually backup when needed:

```bash
# Backup existing installation before update
sudo cp -r /usr/local/go /usr/local/go.backup.$(date +%Y%m%d)

# Update Go
sudo goUpdater update
```

### 3. Use Custom Directories for Development

Isolate development environments:

```bash
# Development installation
sudo goUpdater update --install-dir /opt/go-dev

# Add to PATH only when needed
export PATH=/opt/go-dev/bin:$PATH
```

### 4. Verify After Every Operation

Always verify installations:

```bash
# Make verification part of your workflow
sudo goUpdater update && goUpdater verify
```

### 5. Monitor Disk Space

Ensure sufficient space before operations:

```bash
# Check space before update
df -h /usr/local && sudo goUpdater update
```

### 6. Use Verbose Mode for Debugging

Enable detailed logging when troubleshooting:

```bash
# Verbose update for debugging
sudo goUpdater --verbose update
```

### 7. Clean Up Old Archives

Remove downloaded archives after installation:

```bash
# Clean up after successful installation
goUpdater download && sudo goUpdater install /tmp/go*.tar.gz && rm /tmp/go*.tar.gz
```

### 8. Test in Staging First

For production systems, test updates in staging:

```bash
# Test update in staging environment
ssh staging-server "sudo goUpdater update --install-dir /opt/go-test && goUpdater verify --install-dir /opt/go-test"
```

## Integration Examples

### Scripts

#### Automated Go Update Script

```bash
#!/bin/bash
# go-update.sh - Automated Go update script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check if running as root or with sudo
if [[ $EUID -eq 0 ]]; then
    SUDO=""
else
    SUDO="sudo"
fi

log "Starting automated Go update process"

# Check current Go version
if command -v go &> /dev/null; then
    CURRENT_VERSION=$(go version | awk '{print $3}')
    log "Current Go version: $CURRENT_VERSION"
else
    warning "Go not currently installed"
fi

# Update Go
log "Updating Go to latest version..."
if $SUDO goUpdater update; then
    log "Go update completed successfully"
else
    error "Go update failed"
    exit 1
fi

# Verify installation
log "Verifying Go installation..."
if goUpdater verify; then
    NEW_VERSION=$(go version | awk '{print $3}')
    log "Go verification successful. New version: $NEW_VERSION"
else
    error "Go verification failed"
    exit 1
fi

log "Go update process completed successfully"
```

#### Version Comparison Script

```bash
#!/bin/bash
# check-go-version.sh - Check if Go needs updating

LATEST_VERSION=$(curl -s https://go.dev/VERSION?m=text | head -1 | sed 's/go//')
CURRENT_VERSION=$(go version 2>/dev/null | awk '{print $3}' | sed 's/go//')

if [ -z "$CURRENT_VERSION" ]; then
    echo "Go is not installed"
    exit 1
fi

if [ "$LATEST_VERSION" != "$CURRENT_VERSION" ]; then
    echo "Go update available: $CURRENT_VERSION → $LATEST_VERSION"
    echo "Run: sudo goUpdater update"
else
    echo "Go is up to date: $CURRENT_VERSION"
fi
```

### Automation

#### Ansible Playbook for Go Management

```yaml
---
# ansible/playbooks/go-setup.yml
- name: Setup Go environment
  hosts: all
  become: yes
  tasks:
    - name: Download goUpdater
      get_url:
        url: "https://github.com/nicholas-fedor/goUpdater/releases/latest/download/goUpdater-linux-amd64"
        dest: "/usr/local/bin/goUpdater"
        mode: '0755'

    - name: Update Go to latest version
      command: "/usr/local/bin/goUpdater update"
      register: go_update
      changed_when: "'Successfully updated' in go_update.stdout"

    - name: Verify Go installation
      command: "/usr/local/bin/goUpdater verify"
      register: go_verify
      failed_when: go_verify.rc != 0

    - name: Ensure Go is in PATH
      lineinfile:
        path: "/etc/environment"
        line: "PATH=/usr/local/go/bin:$PATH"
        state: present

- name: Go development environment
  hosts: developers
  become: yes
  tasks:
    - name: Install Go in development directory
      command: "/usr/local/bin/goUpdater update --install-dir /opt/go-dev"

    - name: Create Go workspace
      file:
        path: "/home/{{ ansible_user }}/go"
        state: directory
        owner: "{{ ansible_user }}"
        group: "{{ ansible_user }}"
        mode: '0755'
```

#### Puppet Manifest for Go Management

```puppet
# puppet/manifests/go.pp
class go_manager {
    $go_install_dir = '/usr/local/go'
    $go_updater_url = 'https://github.com/nicholas-fedor/goUpdater/releases/latest/download/goUpdater-linux-amd64'
    $go_updater_path = '/usr/local/bin/goUpdater'

    # Download goUpdater
    exec { 'download_goUpdater':
        command => "/usr/bin/wget -O ${go_updater_path} ${go_updater_url}",
        creates => $go_updater_path,
        require => Package['wget'],
    }

    # Make goUpdater executable
    file { $go_updater_path:
        ensure  => file,
        mode    => '0755',
        require => Exec['download_goUpdater'],
    }

    # Update Go
    exec { 'update_go':
        command => "${go_updater_path} update",
        onlyif  => "${go_updater_path} verify",
        require => File[$go_updater_path],
    }

    # Ensure Go is in system PATH
    file_line { 'go_path':
        path    => '/etc/environment',
        line    => "PATH=${go_install_dir}/bin:\$PATH",
        require => Exec['update_go'],
    }
}
```

### Package Managers

#### Nix Package for goUpdater

```nix
# nix/packages/goUpdater.nix
{ stdenv, fetchurl, autoPatchelfHook }:

stdenv.mkDerivation rec {
  pname = "goUpdater";
  version = "1.0.0";

  src = fetchurl {
    url = "https://github.com/nicholas-fedor/goUpdater/releases/download/v${version}/goUpdater-linux-amd64";
    sha256 = "abc123..."; # Replace with actual SHA256
  };

  nativeBuildInputs = [ autoPatchelfHook ];

  dontUnpack = true;
  dontBuild = true;
  dontConfigure = true;

  installPhase = ''
    install -Dm755 $src $out/bin/goUpdater
  '';

  meta = with stdenv.lib; {
    description = "CLI tool for automating Go installation and updates";
    homepage = "https://github.com/nicholas-fedor/goUpdater";
    license = licenses.agpl3;
    platforms = platforms.linux;
  };
}
```

#### Homebrew Formula (macOS)

```ruby
# homebrew/Formula/goUpdater.rb
class GoUpdater < Formula
  desc "CLI tool for automating Go installation and updates"
  homepage "https://github.com/nicholas-fedor/goUpdater"
  url "https://github.com/nicholas-fedor/goUpdater/releases/download/v1.0.0/goUpdater-linux-amd64"
  sha256 "abc123..." # Replace with actual SHA256
  license "AGPL-3.0-or-later"

  def install
    bin.install "goUpdater-linux-amd64" => "goUpdater"
  end

  test do
    system "#{bin}/goUpdater", "version"
  end
end
```

#### Debian Package Creation

```bash
# Create Debian package structure
mkdir -p goUpdater-deb/DEBIAN
mkdir -p goUpdater-deb/usr/local/bin

# Download goUpdater
wget -O goUpdater-deb/usr/local/bin/goUpdater \
  https://github.com/nicholas-fedor/goUpdater/releases/latest/download/goUpdater-linux-amd64
chmod +x goUpdater-deb/usr/local/bin/goUpdater

# Create control file
cat > goUpdater-deb/DEBIAN/control << EOF
Package: goUpdater
Version: 1.0.0
Section: devel
Priority: optional
Architecture: amd64
Depends: ca-certificates
Maintainer: Your Name <your.email@example.com>
Description: CLI tool for automating Go installation and updates
 Automates the installation, updating, and management of Go programming
 language installations on Linux systems.
EOF

# Build the package
dpkg-deb --build goUpdater-deb

# Install the package
sudo dpkg -i goUpdater-deb.deb
```

These examples demonstrate how goUpdater can be integrated into various automation and deployment scenarios, from simple scripts to enterprise-grade configuration management systems.
