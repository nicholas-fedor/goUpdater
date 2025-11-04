// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package install

import (
	"io"

	"github.com/nicholas-fedor/goUpdater/internal/archive"
	"github.com/nicholas-fedor/goUpdater/internal/download"
	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	"github.com/nicholas-fedor/goUpdater/internal/types"
)

// ArchiveService defines the interface for archive operations.
type ArchiveService interface {
	Validate(archivePath, destDir string) error
	Extract(archivePath, destDir string) error
	ExtractVersion(archivePath string) string
}

// DownloadService defines the interface for download operations.
type DownloadService interface {
	GetLatest(tempDir string) (archivePath, checksum string, err error)
	GetLatestVersionInfo() (versionInfo types.VersionInfo, err error)
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
	reader           io.Reader // reader provides input reading functionality, typically os.Stdin in production
}

// downloadServiceImpl implements DownloadService using download package functions.
type downloadServiceImpl struct {
	downloader *download.Downloader
}

// archiveServiceImpl implements ArchiveService using archive package functions.
type archiveServiceImpl struct {
	extractor *archive.Extractor
}

// defaultVersionFetcherImpl implements VersionFetcher for download service.
type defaultVersionFetcherImpl struct{}

// versionServiceImpl implements VersionService using version package functions.
type versionServiceImpl struct{}
