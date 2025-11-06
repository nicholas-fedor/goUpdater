// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package update

import (
	"os"

	"github.com/nicholas-fedor/goUpdater/internal/archive"
	"github.com/nicholas-fedor/goUpdater/internal/download"
	"github.com/nicholas-fedor/goUpdater/internal/exec"
	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	"github.com/nicholas-fedor/goUpdater/internal/install"
	"github.com/nicholas-fedor/goUpdater/internal/privileges"
	"github.com/nicholas-fedor/goUpdater/internal/uninstall"
	"github.com/nicholas-fedor/goUpdater/internal/verify"
)

// RunUpdate executes the update command logic.
// It performs a complete Go update workflow including privilege checking,
// version comparison, user prompts, and success/error messaging.
// updateDir is the directory where Go should be updated (e.g., "/usr/local/go").
// autoInstall enables automatic installation if Go is not present.
func RunUpdate(updateDir string, autoInstall bool) error {
	return WithPrivileges(updateDir, autoInstall)
}

// createDefaultUpdater creates a new Updater instance with default dependencies.
// It initializes all required components including filesystem, command executor,
// downloader, installer, uninstaller, verifier, privilege manager, and version fetcher.
// This helper function eliminates code duplication by providing a centralized
// way to create an Updater with standard production dependencies.
func createDefaultUpdater() *Updater {
	// Create service implementations for dependency injection
	archiveSvc := install.NewArchiveServiceImpl(
		archive.NewExtractor(&filesystem.OSFileSystem{}, &archive.DefaultProcessor{}),
	)
	downloadSvc := install.NewDownloadServiceImpl(download.NewDownloader(
		&filesystem.OSFileSystem{},
		download.NewDefaultHTTPClient(),
		&exec.OSCommandExecutor{},
		&install.DefaultVersionFetcherImpl{},
	))
	verifySvc := verify.NewVerifier(&filesystem.OSFileSystem{}, &exec.OSCommandExecutor{})
	versionSvc := &install.VersionServiceImpl{}
	privilegeSvc := privileges.NewPrivilegeManager(
		&filesystem.OSFileSystem{},
		privileges.OSPrivilegeManagerImpl{},
		&exec.OSCommandExecutor{},
	)

	installer := install.NewInstallerWithDeps(
		&filesystem.OSFileSystem{},
		archiveSvc,
		downloadSvc,
		verifySvc,
		versionSvc,
		privilegeSvc,
		os.Stdin,
	)

	return NewUpdater(
		&filesystem.OSFileSystem{},
		&exec.OSCommandExecutor{},
		download.NewDownloader(
			&filesystem.OSFileSystem{},
			download.NewDefaultHTTPClient(),
			&exec.OSCommandExecutor{},
			NewDefaultVersionFetcher(),
		),
		installer,
		uninstall.NewDefaultUninstaller(&filesystem.OSFileSystem{}),
		verifySvc,
		privileges.NewPrivilegeManager(
			&filesystem.OSFileSystem{},
			&privileges.OSPrivilegeManagerImpl{},
			&exec.OSCommandExecutor{},
		),
		NewDefaultVersionFetcher(),
	)
}

// Update performs a complete Go update: checks if Go is installed, compares versions,
// downloads the latest version if needed, removes the existing installation,
// installs the new version, verifies it, and logs success message.
// installDir is the directory where Go should be installed (e.g., "/usr/local/go").
// autoInstall enables automatic installation if Go is not present.
func Update(installDir string, autoInstall bool) error {
	updater := createDefaultUpdater()

	return updater.Update(installDir, autoInstall)
}

// WithPrivileges performs a complete Go update workflow including privilege checking,
// version comparison, user prompts, and success/error messaging.
// It wraps the existing update logic and handles all display/output logic.
// installDir is the directory where Go should be installed (e.g., "/usr/local/go").
// autoInstall enables automatic installation if Go is not present.
func WithPrivileges(installDir string, autoInstall bool) error {
	updater := createDefaultUpdater()

	return updater.UpdateWithPrivileges(installDir, autoInstall)
}
