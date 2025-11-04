// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package install provides functionality to install Go from downloaded archives.
// It handles archive extraction, installation verification, and user interaction.
package install

import (
	"os"

	"github.com/nicholas-fedor/goUpdater/internal/archive"
	"github.com/nicholas-fedor/goUpdater/internal/download"
	"github.com/nicholas-fedor/goUpdater/internal/exec"
	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	"github.com/nicholas-fedor/goUpdater/internal/privileges"
	"github.com/nicholas-fedor/goUpdater/internal/verify"
)

// RunInstall executes the install command logic.
// It installs Go to the specified directory, either from the latest version or from a provided archive.
// installDir is the directory where Go should be installed (e.g., "/usr/local/go").
// archivePath is the path to the archive file to install from, or empty to install the latest version.
// WARNING: This function performs real filesystem operations and should not be used in tests.
func RunInstall(installDir, archivePath string) error {
	// Create service implementations for dependency injection
	archiveSvc := &archiveServiceImpl{
		extractor: archive.NewExtractor(&filesystem.OSFileSystem{}, &archive.DefaultProcessor{}),
	}
	downloadSvc := &downloadServiceImpl{
		downloader: download.NewDownloader(
			&filesystem.OSFileSystem{},
			download.NewDefaultHTTPClient(),
			&exec.OSCommandExecutor{},
			&defaultVersionFetcherImpl{},
		),
	}
	verifySvc := verify.NewVerifier(&filesystem.OSFileSystem{}, &exec.OSCommandExecutor{})
	versionSvc := &versionServiceImpl{}
	privilegeSvc := privileges.NewPrivilegeManager(
		&filesystem.OSFileSystem{},
		privileges.OSPrivilegeManagerImpl{},
		&exec.OSCommandExecutor{},
	)

	installer := NewInstallerWithDeps(
		&filesystem.OSFileSystem{},
		archiveSvc,
		downloadSvc,
		verifySvc,
		versionSvc,
		privilegeSvc,
		os.Stdin,
	)

	return installer.Install(installDir, archivePath)
}
