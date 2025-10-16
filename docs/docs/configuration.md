# Configuration

This document covers all aspects of configuring goUpdater, including command-line options, build-time configuration, and best practices for secure and effective usage.

## Command-Line Flags

goUpdater supports several command-line flags that control its behavior and output.

### Global Flags

These flags are available across all commands:

#### `--verbose`, `-v`

Enable verbose logging for detailed operation information.

**Usage:**

```bash
goUpdater --verbose update
goUpdater -v verify
```

**Example output:**

```text
[DEBUG] Starting Go version check
[DEBUG] Current Go version: go1.21.0
[DEBUG] Latest Go version: go1.21.5
[INFO] Downloading Go 1.21.5...
```

### Command-Specific Flags

#### `--install-dir`, `-d` (install, update, uninstall, verify commands)

Specify a custom installation directory for Go. The default is `/usr/local/go`.

**Usage:**

```bash
goUpdater install archive.tar.gz --install-dir /opt/go
goUpdater update --install-dir ~/go
goUpdater verify --install-dir /custom/go/path
```

**Short form:**

```bash
goUpdater update -d /opt/go
```

## Environment Variables

goUpdater does not currently use environment variables for configuration. All configuration is handled through command-line flags and build-time settings.

## Build-Time Configuration

goUpdater supports build-time configuration through Go linker flags (ldflags) to inject version information and other build metadata into the binary.

### Version Information Variables

The following variables are set at build time using ldflags:

- `version`: The application version (e.g., "1.0.0")
- `commit`: Git commit hash of the build
- `date`: Build timestamp in RFC3339 format
- `goVersion`: Go version used to build the binary
- `platform`: Target platform in format "os/arch" (e.g., "linux/amd64")

### Makefile Build Configuration

When building with `make build`, the following ldflags are applied:

```bash
-ldflags "-X github.com/nicholas-fedor/goUpdater/internal/version.goVersion=$(go version | cut -d' ' -f3) -X github.com/nicholas-fedor/goUpdater/internal/version.platform=$(go env GOOS)/$(go env GOARCH)"
```

This injects the Go version and platform information into the binary.

### GoReleaser Configuration

For release builds, GoReleaser uses more comprehensive ldflags:

```yaml
ldflags:
  - -s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}} -X main.builtBy=goreleaser -X github.com/nicholas-fedor/goUpdater/internal/version.goVersion={{.Env.GO_VERSION}} -X github.com/nicholas-fedor/goUpdater/internal/version.platform={{.Runtime.Goos}}/{{.Runtime.Goarch}}
```

The `-s -w` flags strip debugging information and symbol tables to reduce binary size.

### Custom Build Configuration

To build with custom version information:

```bash
go build -ldflags "-X github.com/nicholas-fedor/goUpdater/internal/version.version=1.2.3 -X github.com/nicholas-fedor/goUpdater/internal/version.commit=abc123" -o goUpdater .
```

## Custom Installation Directories

### Directory Selection

When using `--install-dir`, consider:

- **System-wide installations**: Use `/usr/local/go` (default) or `/opt/go`
- **User-specific installations**: Use `~/go` or `$HOME/go`
- **Development environments**: Use project-specific directories like `~/dev/go-versions/go1.21`

### Implications

#### PATH Configuration

After installing to a custom directory, ensure the Go binary is in your PATH:

```bash
# For system-wide installation
export PATH=/opt/go/bin:$PATH

# For user installation
export PATH=$HOME/go/bin:$PATH
```

Add the export to your shell profile (`.bashrc`, `.zshrc`, etc.) for persistence.

#### Permissions

- **System directories** (`/usr/local/go`, `/opt/go`): Require root/sudo privileges
- **User directories** (`~/go`): No special privileges needed
- **Custom directories**: Ensure write permissions for the installing user

#### Multiple Versions

Custom directories enable side-by-side installations:

```bash
# Install multiple versions
sudo goUpdater install go1.20.12.linux-amd64.tar.gz --install-dir /opt/go-1.20
sudo goUpdater install go1.21.5.linux-amd64.tar.gz --install-dir /opt/go-1.21

# Switch between versions
export PATH=/opt/go-1.21/bin:$PATH  # Use Go 1.21
export PATH=/opt/go-1.20/bin:$PATH  # Use Go 1.20
```

## Configuration Best Practices

### Security Considerations

1. **Privilege Management**: Use `sudo` only when necessary. Prefer user-specific installations when possible.

2. **Directory Permissions**: Ensure installation directories have appropriate permissions:

   ```bash
   # For system installations
   sudo chown -R root:root /usr/local/go
   sudo chmod -R 755 /usr/local/go

   # For user installations
   chmod -R 755 ~/go
   ```

3. **Archive Verification**: Always verify downloaded archives before installation, especially from untrusted sources.

### Operational Best Practices

1. **Backup Existing Installations**: Before updating, backup your current Go installation:

   ```bash
   sudo cp -r /usr/local/go /usr/local/go.backup
   ```

2. **Test in Staging**: Use custom directories for testing updates:

   ```bash
   goUpdater update --install-dir ~/test-go
   ~/test-go/bin/go version  # Verify the installation
   ```

3. **Version Pinning**: For production environments, consider pinning specific versions:

   ```bash
   goUpdater install go1.21.5.linux-amd64.tar.gz --install-dir /opt/go-1.21.5
   ```

4. **Regular Updates**: Keep Go updated for security patches and performance improvements.

### Troubleshooting Configuration Issues

1. **Permission Denied**: Use `sudo` for system directories or choose a user-writable location.

2. **PATH Issues**: Verify Go binaries are accessible:

   ```bash
   which go
   go version
   ```

3. **Version Conflicts**: Check for multiple Go installations:

   ```bash
   whereis go
   ls -la /usr/local/go/bin/go
   ```

4. **Build Information**: Use verbose logging to diagnose issues:

    ```bash
    goUpdater --verbose version
    ```

## Security Features

goUpdater implements several security measures to ensure safe Go installations:

### Download Integrity Verification

- All downloads are verified using SHA256 checksums from official Go sources
- Invalid or corrupted archives are automatically removed and re-downloaded
- Network operations are performed with user privileges when possible

### Privilege Escalation

- Uses secure syscall-based privilege escalation via sudo
- Only requests elevated privileges when necessary for system modifications
- Maintains separation between download (user privileges) and installation (elevated privileges)

### Archive Search and Validation

- Automatically searches user directories (~/Downloads, ~) for existing archives before downloading
- Validates checksums of found archives to ensure integrity
- Prevents unnecessary downloads when valid archives already exist
