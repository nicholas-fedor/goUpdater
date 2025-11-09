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
	archiveSvc := &ArchiveServiceImpl{
		extractor: archive.NewExtractor(&filesystem.OSFileSystem{}, &archive.DefaultProcessor{}),
	}
	downloadSvc := &DownloadServiceImpl{
		downloader: download.NewDownloader(
			&filesystem.OSFileSystem{},
			download.NewDefaultHTTPClient(),
			&exec.OSCommandExecutor{},
			&DefaultVersionFetcherImpl{},
		),
	}
	verifySvc := verify.NewVerifier(&filesystem.OSFileSystem{}, &exec.OSCommandExecutor{})
	versionSvc := &VersionServiceImpl{}
	privilegeSvc := privileges.NewPrivilegeManager(
		&filesystem.OSFileSystem{},
		privileges.OSPrivilegeManagerImpl{},
		&exec.OSCommandExecutor{},
	)

	return RunInstallWithDeps(
		&filesystem.OSFileSystem{},
		archiveSvc,
		downloadSvc,
		verifySvc,
		versionSvc,
		privilegeSvc,
		installDir,
		archivePath,
	)
}

// RunInstallWithDeps executes the install command logic with injected dependencies.
// This function is testable and allows for dependency injection in unit tests.
// installDir is the directory where Go should be installed (e.g., "/usr/local/go").
// archivePath is the path to the archive file to install from, or empty to install the latest version.
func RunInstallWithDeps(
	fs filesystem.FileSystem,
	archiveSvc ArchiveService,
	downloadSvc DownloadService,
	verifySvc VerifyService,
	versionSvc VersionService,
	privilegeSvc PrivilegeService,
	installDir, archivePath string,
) error {
	installer := NewInstallerWithDeps(
		fs,
		archiveSvc,
		downloadSvc,
		verifySvc,
		versionSvc,
		privilegeSvc,
		os.Stdin,
	)

	return installer.Install(installDir, archivePath)
}
