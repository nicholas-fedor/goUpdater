// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package install

import (
	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
)

// ArchiveService defines the interface for archive operations.
type ArchiveService interface {
	Validate(archivePath string) error
	Extract(archivePath, destDir string) error
	ExtractVersion(archivePath string) string
}

// DownloadService defines the interface for download operations.
type DownloadService interface {
	GetLatest(tempDir string) (archivePath, checksum string, err error)
	GetLatestVersionInfo() (versionInfo struct{ Version string }, err error)
}

// VerifyService defines the interface for verification operations.
type VerifyService interface {
	Installation(installDir, expectedVersion string) error
	GetInstalledVersion(installDir string) (string, error)
}

// VersionService defines the interface for version comparison operations.
type VersionService interface {
	Compare(installedVersion, latestVersion string) (int, error)
}

// PrivilegeService defines the interface for privilege elevation operations.
type PrivilegeService interface {
	ElevateAndExecute(fn func() error) error
}

// Installer handles Go installation operations with dependency injection for all external services.
type Installer struct {
	fs               filesystem.FileSystem
	archiveService   ArchiveService
	downloadService  DownloadService
	verifyService    VerifyService
	versionService   VersionService
	privilegeService PrivilegeService
}

// NewInstaller creates a new Installer with the provided dependencies.
func NewInstaller(fileSystem filesystem.FileSystem) *Installer {
	return &Installer{
		fs:               fileSystem,
		archiveService:   nil,
		downloadService:  nil,
		verifyService:    nil,
		versionService:   nil,
		privilegeService: nil,
	}
}

// NewInstallerWithDeps creates a new Installer with all dependencies injected.
func NewInstallerWithDeps(
	fileSystem filesystem.FileSystem,
	archiveSvc ArchiveService,
	downloadSvc DownloadService,
	verifySvc VerifyService,
	versionSvc VersionService,
	privilegeSvc PrivilegeService,
) *Installer {
	return &Installer{
		fs:               fileSystem,
		archiveService:   archiveSvc,
		downloadService:  downloadSvc,
		verifyService:    verifySvc,
		versionService:   versionSvc,
		privilegeService: privilegeSvc,
	}
}
