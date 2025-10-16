<!-- markdownlint-disable -->
<div align="center">
<img src=".github/assets/logo/goUpdater.svg" alt="goUpdater Logo" width="200px;" />
<h1>goUpdater</h1>

[![Latest Version](https://img.shields.io/github/tag/nicholas-fedor/goupdater.svg)](https://github.com/nicholas-fedor/goupdater/releases)
[![CircleCI](https://dl.circleci.com/status-badge/img/gh/nicholas-fedor/goupdater/tree/main.svg?style=shield)](https://dl.circleci.com/status-badge/redirect/gh/nicholas-fedor/goupdater/tree/main)
[![Codecov](https://codecov.io/gh/nicholas-fedor/goupdater/branch/main/graph/badge.svg)](https://codecov.io/gh/nicholas-fedor/goupdater)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/ffbca83bd14d48669260bb9bb38668a8)](https://www.codacy.com/gh/nicholas-fedor/goupdater/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=nicholas-fedor/goupdater&amp;utm_campaign=Badge_Grade)
[![GoDoc](https://godoc.org/github.com/nicholas-fedor/goupdater?status.svg)](https://godoc.org/github.com/nicholas-fedor/goupdater)
[![Go Report Card](https://goreportcard.com/badge/github.com/nicholas-fedor/goupdater)](https://goreportcard.com/report/github.com/nicholas-fedor/goupdater)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/nicholas-fedor/goupdater)
[![License](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
</div>

<!-- markdownlint-restore -->

A CLI tool that automates the installation, updating, and management of Go programming language installations on Linux.

goUpdater simplifies Go version management by providing commands to download, install, uninstall, update, and verify Go installations, including automatically handling platform detection automatically and  performing checksum verification.

The tool is designed to work with the official Go distribution from [go.dev](https://go.dev) and defaults to installing Go in `/usr/local/go`.

## Installation

1. Download the latest release from the [GitHub releases page](https://github.com/nicholas-fedor/goUpdater/releases):

   ```bash
   wget https://github.com/nicholas-fedor/goUpdater/releases/latest/download/goUpdater
   ```

2. Make the binary executable:

   ```bash
   chmod +x goUpdater
   ```

3. Verify the installation:

   ```bash
   ./goUpdater version
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
