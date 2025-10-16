<!-- markdownlint-disable -->
<img src=".github/assets/logo/goUpdater.svg" alt="goUpdater Logo" style="display: block; margin: 0 auto; width: 200px;" />

<h1 align="center">goUpdater</h1>
<!-- markdownlint-restore -->

A CLI tool that automates the installation, updating, and management of Go programming language installations on Linux.

goUpdater simplifies Go version management by providing commands to download, install, uninstall, update, and verify Go installations, including automatically handling platform detection automatically and  performing checksum verification.

The tool is designed to work with the official Go distribution from [go.dev](https://go.dev) and defaults to installing Go in `/usr/local/go`.

## Installation

1. Download the latest release from the [GitHub releases page](https://github.com/nicholas-fedor/goUpdater/releases):

   ```bash
   wget https://github.com/nicholas-fedor/goUpdater/releases/latest/download/goUpdater-linux-amd64
   ```

2. Make the binary executable:

   ```bash
   chmod +x goUpdater-linux-amd64
   ```

3. Verify the installation:

   ```bash
   ./goUpdater-linux-amd64 version
   ```

## Usage

1. Update Go to the latest version:

   ```bash
   sudo goUpdater update
   ```

2. Verify installation:

   ```bash
   goUpdater verify
   ```

For comprehensive documentation, including detailed commands, workflows, and troubleshooting, visit [goupdater.nickfedor.com](https://goupdater.nickfedor.com).

## License

This project is licensed under the AGPL-3.0-or-later License. See [LICENSE.md](LICENSE.md) for details.
