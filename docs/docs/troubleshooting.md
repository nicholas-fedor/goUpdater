# Troubleshooting Guide

This guide addresses common issues and problems users might encounter when using goUpdater. Each section provides diagnostic steps, solutions, and prevention tips based on actual error handling patterns in the codebase.

## Permission Errors

### Problem: "Failed to obtain elevated privileges" or "Installation requires elevated privileges"

**Symptoms:**

- Commands fail with privilege-related errors
- Sudo request denied or failed
- Unable to write to system directories like `/usr/local/go`

**Diagnostic Steps:**

```bash
# Check if running as root
whoami

# Check current user privileges
id

# Verify sudo is available
which sudo
```

**Solutions:**

1. **Use sudo for system installations:**

   ```bash
   sudo goUpdater update
   sudo goUpdater install /path/to/archive.tar.gz
   ```

2. **Install to user-writable directory:**

   ```bash
   goUpdater update --install-dir ~/go
   export PATH=$HOME/go/bin:$PATH
   ```

3. **Fix sudo configuration:**

   ```bash
   # Ensure user is in sudo group
   sudo usermod -aG sudo $USER
   # Log out and back in for group changes to take effect
   ```

**Prevention Tips:**

- Always use `sudo` for default `/usr/local/go` installations
- Consider user-space installations (`~/go`) for development environments
- Test privilege escalation with `sudo -v` before running goUpdater

## Network Connection Issues

### Problem: "Failed to fetch version info" or "Download failed"

**Symptoms:**

- HTTP requests to `go.dev` fail
- Timeout errors during download
- Network-related error messages

**Diagnostic Steps:**

```bash
# Test basic connectivity
ping -c 3 go.dev

# Test HTTPS connectivity
curl -I https://go.dev/dl/

# Check DNS resolution
nslookup go.dev

# Verify firewall settings
sudo ufw status
```

**Solutions:**

1. **Check network connectivity:**

   ```bash
   # Test internet connection
   curl -s https://www.google.com > /dev/null && echo "Internet OK"

   # Test Go download site specifically
   curl -s https://go.dev/dl/?mode=json | head -c 100
   ```

2. **Configure proxy settings (if applicable):**

   ```bash
   export HTTP_PROXY=http://proxy.company.com:8080
   export HTTPS_PROXY=http://proxy.company.com:8080
   goUpdater download
   ```

3. **Use alternative DNS:**

   ```bash
   # Temporarily use Google DNS
   echo "nameserver 8.8.8.8" | sudo tee /etc/resolv.conf > /dev/null
   ```

**Prevention Tips:**

- Ensure stable internet connection before running updates
- Test connectivity to `go.dev` in restricted networks
- Consider offline installation workflows for air-gapped environments

## Checksum Verification Failures

### Problem: "Checksum mismatch" or "Checksum verification failed"

**Symptoms:**

- Download completes but verification fails
- SHA256 hash doesn't match expected value
- Archive appears corrupted

**Diagnostic Steps:**

```bash
# Check available disk space
df -h /tmp

# Verify downloaded file integrity
ls -la /tmp/go*.tar.gz

# Check if file is readable
head -c 100 /tmp/go*.tar.gz
```

**Solutions:**

1. **Re-download the archive:**

   ```bash
   # Clean up failed download
   rm -f /tmp/go*.tar.gz

   # Download fresh copy
   goUpdater download
   ```

2. **Check disk space:**

   ```bash
   # Ensure sufficient space (Go archives are ~500MB)
   df -h /tmp
   df -h $HOME

   # Clean up temporary files if needed
   sudo du -sh /tmp
   sudo rm -rf /tmp/* 2>/dev/null || true
   ```

3. **Verify system integrity:**

   ```bash
   # Check for disk errors
   sudo dmesg | grep -i error

   # Run filesystem check if suspicious
   sudo touch /forcefsck
   ```

**Prevention Tips:**

- Ensure at least 1GB free space in download directory
- Avoid downloading to full filesystems
- Re-download immediately if checksum fails (don't retry corrupted files)

## Disk Space Problems

### Problem: "No space left on device" or extraction failures

**Symptoms:**

- Archive extraction fails
- "Failed to create directory" errors
- Installation stops midway

**Diagnostic Steps:**

```bash
# Check disk space in installation directory
df -h /usr/local

# Check inode availability
df -i /usr/local

# Estimate Go installation size
du -sh /usr/local/go 2>/dev/null || echo "Go not installed"
```

**Solutions:**

1. **Free up disk space:**

   ```bash
   # Check largest directories
   sudo du -h /usr/local | sort -hr | head -10

   # Clean package cache
   sudo apt clean  # Ubuntu/Debian
   sudo yum clean all  # RHEL/CentOS

   # Remove old Go versions
   sudo rm -rf /usr/local/go.old*
   ```

2. **Use alternative installation directory:**

   ```bash
   # Install to larger filesystem
   df -h  # Find suitable location
   sudo goUpdater install archive.tar.gz --install-dir /opt/go
   ```

3. **Clean temporary files:**

   ```bash
   # Clear system temp
   sudo rm -rf /tmp/*
   sudo rm -rf /var/tmp/*

   # Clear user temp
   rm -rf ~/tmp/* 2>/dev/null || true
   ```

**Prevention Tips:**

- Check available space before installation: `df -h /usr/local`
- Go installations require approximately 500MB free space
- Monitor disk usage in automated environments

## Archive Validation Errors

### Problem: "Archive file does not exist" or "Archive path is not a regular file"

**Symptoms:**

- Archive validation fails
- File not found errors
- Path-related validation errors

**Diagnostic Steps:**

```bash
# Verify archive exists and is readable
ls -la /path/to/go*.tar.gz

# Check file type
file /path/to/go*.tar.gz

# Verify archive integrity
tar -tzf /path/to/go*.tar.gz > /dev/null && echo "Archive OK"
```

**Solutions:**

1. **Verify archive path:**

   ```bash
   # Use absolute paths
   goUpdater install /tmp/go1.21.0.linux-amd64.tar.gz

   # Check if file exists
   ls -la /tmp/go1.21.0.linux-amd64.tar.gz
   ```

2. **Download fresh archive:**

   ```bash
   # Re-run download command
   goUpdater download

   # Find the downloaded file
   ls -la /tmp/go*.tar.gz
   ```

3. **Check file permissions:**

   ```bash
   # Ensure archive is readable
   chmod 644 /path/to/archive.tar.gz
   ```

**Prevention Tips:**

- Use absolute paths when specifying archive locations
- Verify archive exists before running install commands
- Don't move or delete archives between download and install steps

## Installation Verification Failures

### Problem: "Go binary not found" or "Version mismatch"

**Symptoms:**

- Installation appears successful but verification fails
- `go version` command not found
- Version doesn't match expected

**Diagnostic Steps:**

```bash
# Check if Go binary exists
ls -la /usr/local/go/bin/go

# Test Go binary directly
/usr/local/go/bin/go version

# Check PATH
echo $PATH | tr ':' '\n' | grep go

# Verify installation directory
ls -la /usr/local/go/
```

**Solutions:**

1. **Update PATH:**

   ```bash
   # Add Go to PATH
   export PATH=/usr/local/go/bin:$PATH

   # For permanent change, add to ~/.bashrc
   echo 'export PATH=/usr/local/go/bin:$PATH' >> ~/.bashrc
   source ~/.bashrc
   ```

2. **Re-verify installation:**

   ```bash
   # Run verification command
   goUpdater verify

   # Or check manually
   /usr/local/go/bin/go version
   ```

3. **Reinstall if verification fails:**

   ```bash
   # Complete reinstall
   sudo goUpdater update --auto-install
   ```

**Prevention Tips:**

- Always run `goUpdater verify` after installation
- Ensure `/usr/local/go/bin` is in PATH for system installations
- Test `go version` command after installation

## Privilege Escalation Failures

### Problem: "Failed to execute with sudo" or process exits unexpectedly

**Symptoms:**

- Sudo request fails
- Process terminates without completing
- Elevation errors in logs

**Diagnostic Steps:**

```bash
# Test sudo access
sudo -v

# Check sudo configuration
sudo -l

# Verify user can run sudo
sudo whoami
```

**Solutions:**

1. **Fix sudo configuration:**

   ```bash
   # Ensure user has sudo privileges
   sudo visudo
   # Add line: username ALL=(ALL) NOPASSWD: ALL

   # Or add to sudo group
   sudo usermod -aG sudo $USER
   ```

2. **Use direct root access:**

   ```bash
   # Switch to root user
   sudo su -
   goUpdater update
   ```

3. **Check sudo timeout:**

   ```bash
   # Refresh sudo timestamp
   sudo -v
   goUpdater update
   ```

**Prevention Tips:**

- Test sudo access before running privileged commands
- Use `sudo -v` to refresh sudo credentials
- Consider passwordless sudo for automated environments

## File System Permission Issues

### Problem: "Failed to create directory" or "Permission denied"

**Symptoms:**

- Directory creation fails
- File write operations fail
- Installation cannot proceed

**Diagnostic Steps:**

```bash
# Check directory permissions
ls -ld /usr/local

# Check user permissions
id

# Test write access
touch /usr/local/test && rm /usr/local/test && echo "Write OK"
```

**Solutions:**

1. **Fix directory permissions:**

   ```bash
   # Ensure parent directory is writable
   sudo chown -R root:root /usr/local
   sudo chmod 755 /usr/local
   ```

2. **Use sudo for system directories:**

   ```bash
   # Always use sudo for /usr/local installations
   sudo goUpdater install archive.tar.gz
   ```

3. **Choose user-writable location:**

   ```bash
   # Install to home directory
   goUpdater install archive.tar.gz --install-dir ~/go
   ```

**Prevention Tips:**

- Use `sudo` for all system directory operations
- Check directory permissions before installation
- Consider user-space installations for restricted environments

## When to Seek Help

### Community Support

If you encounter issues not covered in this guide:

1. **Gather diagnostic information:**

   ```bash
   # Collect system information
   uname -a
   go version 2>/dev/null || echo "Go not installed"
   goUpdater version

   # Check recent logs
   journalctl -u goUpdater 2>/dev/null || echo "No systemd logs"
   ```

2. **Report issues with:**
   - Complete error messages
   - System information (`uname -a`)
   - goUpdater version (`goUpdater version`)
   - Steps to reproduce the issue

### Professional Support

For enterprise deployments or critical production issues:

- Contact system administrators
- Review organizational policies for software installation
- Consider commercial support options

## Best Practices

### 1. Pre-Installation Checks

```bash
# Verify system requirements
goUpdater --verbose version
df -h /usr/local
ping -c 1 go.dev
```

### 2. Use Verbose Mode for Debugging

```bash
# Enable detailed logging
goUpdater --verbose update
```

### 3. Backup Existing Installations

```bash
# Create backup before major changes
sudo cp -r /usr/local/go /usr/local/go.backup.$(date +%Y%m%d)
```

### 4. Test in Non-Production First

```bash
# Use custom directory for testing
goUpdater update --install-dir ~/test-go
```

### 5. Monitor Disk Space

```bash
# Regular space checks
df -h /usr/local /tmp
```

### 6. Keep System Updated

```bash
# Update package managers
sudo apt update && sudo apt upgrade  # Ubuntu/Debian
sudo yum update  # RHEL/CentOS
```

This troubleshooting guide covers the most common issues encountered with goUpdater, based on error handling patterns in the codebase. Following these guidelines should resolve most problems and help maintain reliable Go installations.
