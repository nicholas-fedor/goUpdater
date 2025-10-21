// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package update provides orchestration logic for updating Go installations.
// It coordinates downloading, installing, and verifying Go updates.
package update

import (
	"fmt"

	"github.com/nicholas-fedor/goUpdater/internal/download"
	"github.com/nicholas-fedor/goUpdater/internal/exec"
	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	httpclient "github.com/nicholas-fedor/goUpdater/internal/http"
	"github.com/nicholas-fedor/goUpdater/internal/install"
	"github.com/nicholas-fedor/goUpdater/internal/privileges"
	"github.com/nicholas-fedor/goUpdater/internal/uninstall"
	"github.com/nicholas-fedor/goUpdater/internal/verify"
)

// Update performs a complete Go update: checks if Go is installed, compares versions,
// downloads the latest version if needed, removes the existing installation,
// installs the new version, verifies it, and logs success message.
// installDir is the directory where Go should be installed (e.g., "/usr/local/go").
// autoInstall enables automatic installation if Go is not present.
func Update(installDir string, autoInstall bool) error {
	updater := NewUpdater(
		&filesystem.OSFileSystem{},
		&exec.OSCommandExecutor{},
		download.NewDownloader(
			&filesystem.OSFileSystem{},
			&download.DefaultHTTPClient{Client: download.NewHTTPClient()},
			&exec.OSCommandExecutor{},
			&DefaultVersionFetcher{},
		),
		&install.Installer{},
		uninstall.NewDefaultUninstaller(&filesystem.OSFileSystem{}),
		&verify.Verifier{},
		privileges.NewPrivilegeManager(
			&filesystem.OSFileSystem{},
			&privileges.OSPrivilegeManagerImpl{},
			&exec.OSCommandExecutor{},
		),
	)

	return updater.Update(installDir, autoInstall)
}

// WithPrivileges performs a complete Go update workflow including privilege checking,
// version comparison, user prompts, and success/error messaging.
// It wraps the existing update logic and handles all display/output logic.
// installDir is the directory where Go should be installed (e.g., "/usr/local/go").
// autoInstall enables automatic installation if Go is not present.
func WithPrivileges(installDir string, autoInstall bool) error {
	updater := NewUpdater(
		&filesystem.OSFileSystem{},
		&exec.OSCommandExecutor{},
		download.NewDownloader(
			&filesystem.OSFileSystem{},
			&download.DefaultHTTPClient{Client: download.NewHTTPClient()},
			&exec.OSCommandExecutor{},
			&DefaultVersionFetcher{},
		),
		&install.Installer{},
		uninstall.NewDefaultUninstaller(&filesystem.OSFileSystem{}),
		&verify.Verifier{},
		privileges.NewPrivilegeManager(
			&filesystem.OSFileSystem{},
			&privileges.OSPrivilegeManagerImpl{},
			&exec.OSCommandExecutor{},
		),
	)

	return updater.UpdateWithPrivileges(installDir, autoInstall)
}

// RunUpdate executes the update command logic.
// It performs a complete Go update workflow including privilege checking,
// version comparison, user prompts, and success/error messaging.
// updateDir is the directory where Go should be updated (e.g., "/usr/local/go").
// autoInstall enables automatic installation if Go is not present.
func RunUpdate(updateDir string, autoInstall bool) error {
	return WithPrivileges(updateDir, autoInstall)
}

// DefaultVersionFetcher implements VersionFetcher using the version package.
type DefaultVersionFetcher struct{}

// GetLatestVersion fetches the latest stable Go version information from the official API.
// It returns the version info for the current platform or an error if not found.
// This function maintains backward compatibility by using the default HTTP client.
func (d *DefaultVersionFetcher) GetLatestVersion() (*httpclient.GoVersionInfo, error) {
	info, err := httpclient.GetLatestVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}

	return info, nil
}

// GetLatestVersionInfo fetches the latest Go version information.
// It returns the version info for the current platform or an error if not found.
// This function maintains backward compatibility by using the default HTTP client.
func (d *DefaultVersionFetcher) GetLatestVersionInfo() (*download.GoVersionInfo, error) {
	info, err := d.GetLatestVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}

	return &download.GoVersionInfo{
		Version: info.Version,
		Stable:  info.Stable,
		Files:   convertGoFileInfos(info.Files),
	}, nil
}

// convertGoFileInfos converts httpclient.GoFileInfo slice to download.GoFileInfo slice.
func convertGoFileInfos(files []httpclient.GoFileInfo) []download.GoFileInfo {
	result := make([]download.GoFileInfo, len(files))
	for i, file := range files {
		result[i] = download.GoFileInfo{
			Filename: file.Filename,
			OS:       file.OS,
			Arch:     file.Arch,
			Version:  file.Version,
			Sha256:   file.Sha256,
			Size:     file.Size,
			Kind:     file.Kind,
		}
	}

	return result
}
