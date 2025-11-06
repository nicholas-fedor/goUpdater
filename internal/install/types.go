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
	ArchiveService   ArchiveService
	downloadService  DownloadService
	verifyService    VerifyService
	versionService   VersionService
	privilegeService PrivilegeService
	reader           io.Reader // reader provides input reading functionality, typically os.Stdin in production
}

// DownloadServiceImpl implements DownloadService using download package functions.
type DownloadServiceImpl struct {
	downloader *download.Downloader
}

// NewDownloadServiceImpl creates a new DownloadServiceImpl.
func NewDownloadServiceImpl(downloader *download.Downloader) *DownloadServiceImpl {
	return &DownloadServiceImpl{downloader: downloader}
}

// ArchiveServiceImpl implements ArchiveService using archive package functions.
type ArchiveServiceImpl struct {
	extractor *archive.Extractor
}

// NewArchiveServiceImpl creates a new ArchiveServiceImpl.
func NewArchiveServiceImpl(extractor *archive.Extractor) *ArchiveServiceImpl {
	return &ArchiveServiceImpl{extractor: extractor}
}

// DefaultVersionFetcherImpl implements VersionFetcher for download service.
type DefaultVersionFetcherImpl struct{}

// NewDefaultVersionFetcherImpl creates a new DefaultVersionFetcherImpl.
func NewDefaultVersionFetcherImpl() *DefaultVersionFetcherImpl {
	return &DefaultVersionFetcherImpl{}
}

// VersionServiceImpl implements VersionService using version package functions.
type VersionServiceImpl struct{}

// NewVersionServiceImpl creates a new VersionServiceImpl.
func NewVersionServiceImpl() *VersionServiceImpl {
	return &VersionServiceImpl{}
}
