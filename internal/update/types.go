// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package update

import (
	"github.com/nicholas-fedor/goUpdater/internal/exec"
	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	httpclient "github.com/nicholas-fedor/goUpdater/internal/http"
)

// CommandExecutor defines the interface for executing commands.
type CommandExecutor = exec.CommandExecutor

// VersionFetcher defines the interface for fetching version information.
// Canonical ownership: httpclient.GoVersionInfo.
type VersionFetcher interface {
	// GetLatestVersionInfo fetches the latest Go version information.
	GetLatestVersionInfo() (*httpclient.GoVersionInfo, error)
}

// ArchiveDownloader defines the interface for downloading Go archives.
type ArchiveDownloader interface {
	// GetLatest downloads the latest Go archive to the specified directory.
	GetLatest(destDir string) (string, string, error)
}

// Installer defines the interface for installing Go.
type Installer interface {
	// Extract extracts the Go archive to the specified directory.
	Extract(archivePath, installDir, version string) error
}

// Uninstaller defines the interface for uninstalling Go.
type Uninstaller interface {
	// Remove removes the Go installation from the specified directory.
	Remove(installDir string) error
}

// Verifier defines the interface for verifying Go installations.
type Verifier interface {
	// Installation verifies that Go is properly installed in the specified directory.
	Installation(installDir, version string) error
}

// PrivilegeManager defines the interface for managing privilege elevation.
type PrivilegeManager interface {
	// ElevateAndExecute executes the given function with elevated privileges.
	ElevateAndExecute(fn func() error) error
}

// Updater handles Go installation updates with dependency injection.
type Updater struct {
	fileSystem        filesystem.FileSystem
	commandExecutor   CommandExecutor
	versionFetcher    VersionFetcher
	archiveDownloader ArchiveDownloader
	installer         Installer
	uninstaller       Uninstaller
	verifier          Verifier
	privilegeManager  PrivilegeManager
}

// DefaultVersionFetcher implements VersionFetcher using the version package.
type DefaultVersionFetcher struct {
	getLatestVersionFunc func() (*httpclient.GoVersionInfo, error)
}

// NewDefaultVersionFetcher creates a new DefaultVersionFetcher with the default HTTP client.
func NewDefaultVersionFetcher() *DefaultVersionFetcher {
	return &DefaultVersionFetcher{
		getLatestVersionFunc: httpclient.GetLatestVersion,
	}
}
