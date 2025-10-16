# Command Reference

This document provides comprehensive reference documentation for all goUpdater CLI commands. goUpdater is a tool for automating Go programming language installation, updating, and management on Linux systems.

## Global Options

All commands support the following global options:

### `--verbose`, `-v`

Enable verbose logging for detailed operation information.

Enable verbose logging for the update command:

```bash
goUpdater --verbose update
```

Enable verbose logging for the verify command:

```bash
goUpdater -v verify
```

### `--install-dir`

Specify a custom installation directory (default: `/usr/local/go`). This option is available for commands that interact with Go installations.

Specify custom installation directory for the update command:

```bash
goUpdater --install-dir /opt/go update
```

## Commands

### `update`

Updates Go to the latest stable version by performing a complete update cycle: downloading the latest archive, uninstalling the current version, installing the new version, and verifying the installation. The `--auto-install` flag enables automatic installation of Go if no existing installation is detected, making it suitable for initial setup scenarios.

#### Syntax

```bash
goUpdater update [flags]
```

#### Flags

- `--install-dir`, `-d` string: Directory where Go should be updated (default "/usr/local/go")
- `--auto-install`, `-a`: Automatically install Go if not present (default false)

#### Examples

Update Go to the latest stable version:

```bash
sudo goUpdater update
```

Update Go in a custom installation directory:

```bash
sudo goUpdater update --install-dir /opt/go
```

Update Go with auto-install enabled:

```bash
sudo goUpdater update --auto-install
```

#### Expected Output

```bash
Successfully updated Go in: /usr/local/go
```

#### Error Cases

- Returns exit code 1 if update fails
- Requires sudo privileges for system directories
- Fails if network connection is unavailable for downloading

### `download`

Downloads the latest stable Go version archive for the current platform to a temporary directory and verifies its integrity using SHA256 checksum. The command automatically searches for existing archives in common user directories (user's Downloads directory and home directory) before downloading, prioritizing user-downloaded archives over temporary directory downloads. During download, a progress bar displays download speed, estimated time of arrival (ETA), and completion percentage.

#### Syntax

```bash
goUpdater download
```

#### Flags

None

#### Examples

Download the latest Go archive:

```bash
goUpdater download
```

#### Expected Output

```bash
Successfully downloaded Go archive to: /tmp/go{version}.linux-amd64.tar.gz
SHA256 checksum: abc123...
```

#### Error Cases

- Returns exit code 1 if download fails
- Fails if network connection is unavailable
- Fails if checksum verification fails

### `install`

Installs Go either by automatically downloading and installing the latest stable version, or from a specified archive file.

#### Syntax

```bash
goUpdater install [archive-path] [flags]
```

#### Arguments

- `archive-path`: Path to the Go archive file (optional - if not provided, downloads the latest version)

#### Flags

- `--install-dir`, `-d` string: Directory to install Go (default "/usr/local/go")

#### Examples

Install the latest Go version automatically:

```bash
sudo goUpdater install
```

Install Go from a specific archive file:

```bash
sudo goUpdater install /tmp/go{version}.linux-amd64.tar.gz
```

Install Go to a custom directory:

```bash
sudo goUpdater install go{version}.linux-amd64.tar.gz --install-dir /opt/go
```

#### Expected Output

```bash
Successfully installed Go to: /usr/local/go
```

#### Error Cases

- Returns exit code 1 if installation fails
- Requires sudo privileges for system directories
- Fails if archive file is invalid or corrupted

### `uninstall`

Removes the Go installation from the specified directory.

#### Syntax

```bash
goUpdater uninstall [flags]
```

#### Flags

- `--install-dir`, `-d` string: Directory from which to uninstall Go (default "/usr/local/go")

#### Examples

Uninstall Go from the default directory:

```bash
sudo goUpdater uninstall
```

Uninstall Go from a custom directory:

```bash
sudo goUpdater uninstall --install-dir /opt/go
```

#### Expected Output

```bash
Successfully uninstalled Go from: /usr/local/go
```

#### Error Cases

- Returns exit code 1 if uninstallation fails
- Requires sudo privileges for system directories
- Fails if Go is not installed in the specified directory

### `verify`

Verifies that Go is properly installed by checking the version of the installed Go binary.

#### Syntax

```bash
goUpdater verify [flags]
```

#### Flags

- `--install-dir`, `-d` string: Directory to verify Go installation (default "/usr/local/go")

#### Examples

Verify the Go installation in the default directory:

```bash
goUpdater verify
```

Verify the Go installation in a custom directory:

```bash
goUpdater verify --install-dir /opt/go
```

#### Expected Output

```bash
Go installation verified. Version: go{version} linux/amd64
```

#### Error Cases

- Returns exit code 1 if verification fails
- Fails if Go binary is not found
- Fails if Go version cannot be determined

### `version`

Displays detailed version information of goUpdater including version, commit hash, build date, Go version, and platform.

#### Syntax

```bash
goUpdater version [flags]
```

#### Flags

- `--format` string: Output format: default, short, verbose, json
- `--json`: Output in JSON format (shorthand for --format=json)
- `--short`: Output only version number (shorthand for --format=short)
- `--verbose`: Output all available information (shorthand for --format=verbose)

#### Examples

Display default version information:

```bash
goUpdater version
```

Display only the version number:

```bash
goUpdater version --short
```

Display verbose version information:

```bash
goUpdater version --verbose
```

Display version information in JSON format:

```bash
goUpdater version --json
```

#### Expected Output (Default Format)

```
goUpdater {version}
├─ Commit: {commit-hash}
├─ Built: {build-date}
├─ Go version: go{go-version}
└─ Platform: linux/amd64
```

#### Expected Output (JSON Format)

```json
{
  "version": "{version}",
  "commit": "{commit-hash}",
  "date": "{build-date}",
  "goVersion": "go{go-version}",
  "platform": "linux/amd64"
}
```

### `completion`

Generates shell completion scripts for bash, zsh, fish, and PowerShell to enable tab completion for goUpdater commands.

#### Syntax

```bash
goUpdater completion [bash|zsh|fish|powershell]
```

#### Arguments

- `bash`: Generate bash completion script
- `zsh`: Generate zsh completion script
- `fish`: Generate fish completion script
- `powershell`: Generate PowerShell completion script

#### Examples

Generate bash completion script:

```bash
goUpdater completion bash > goUpdater-completion.bash
```

To load bash completion:

```bash
source goUpdater-completion.bash
```

Or to make it permanent, add to your ~/.bashrc:

```bash
echo 'source <(goUpdater completion bash)' >> ~/.bashrc
```

Generate zsh completion script:

```bash
goUpdater completion zsh > _goUpdater
```

To load zsh completion:

```bash
source _goUpdater
```

Or to make it permanent, add to your ~/.zshrc:

```bash
echo 'source <(goUpdater completion zsh)' >> ~/.zshrc
```

Generate fish completion script:

```bash
goUpdater completion fish > goUpdater.fish
```

To load fish completion, add to ~/.config/fish/completions/:

```bash
goUpdater completion fish > ~/.config/fish/completions/goUpdater.fish
```

Generate PowerShell completion script:

```bash
goUpdater completion powershell > goUpdater.ps1
```

To load PowerShell completion, add to your PowerShell profile:

```powershell
. /path/to/goUpdater.ps1
```

#### Expected Output

The command outputs the completion script to stdout.

#### Error Cases

- Returns exit code 1 if invalid shell is specified
- Requires the shell to be installed for the completion to work

## Error Handling and Exit Codes

goUpdater uses standard exit codes to indicate success or failure:

- **Exit Code 0**: Success - Command completed successfully
- **Exit Code 1**: General error - Command failed due to various reasons

### Common Error Scenarios

1. **Permission Errors**: Installing to system directories requires sudo privileges

    ```bash
    sudo goUpdater update
    ```

2. **Network Errors**: Download commands require internet connectivity

    ```bash
    ping go.dev
    ```

3. **File System Errors**: Ensure sufficient disk space and write permissions

    ```bash
    # Check disk space
    df -h
    ```

4. **Archive Corruption**: Downloads are verified with SHA256 checksums

    ```bash
    goUpdater download
    ```

### Privilege Escalation

goUpdater uses secure syscall-based privilege escalation. When elevated privileges are needed, it will automatically request sudo access. The tool handles all privilege escalation transparently, ensuring that downloads and installations are performed with appropriate security measures. All network operations are conducted with the original user's privileges when possible, while system modifications require elevated privileges.

This command reference covers all goUpdater CLI functionality with detailed syntax, examples, and operational guidance for effective Go version management.
