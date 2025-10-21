// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package download provides functionality to download Go archives from official sources.
// It handles version checking, archive retrieval, and integrity verification.
package download

import (
	"fmt"
	"sync"
	"time"

	"github.com/nicholas-fedor/goUpdater/internal/exec"
	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	httpclient "github.com/nicholas-fedor/goUpdater/internal/http"
	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

// defaultVersionFetcher implements VersionFetcher using the version package.
type defaultVersionFetcher struct{}

// throttleDuration defines the update interval for the progress bar in milliseconds.
const throttleDuration = 100 // Progress bar update interval in milliseconds

// maxRetries defines the maximum number of retry attempts for failed downloads.
const maxRetries = 3

// retryDelay defines the base delay between retry attempts.
const retryDelay = 2 * time.Second

// versionCacheMutex protects concurrent access to versionCache operations.
// This ensures thread-safe access to the global version cache.
//
//nolint:gochecknoglobals // Required for thread-safe global cache protection
var versionCacheMutex sync.RWMutex

// versionCache provides thread-safe in-memory caching for version info to avoid redundant fetches within a single execution.
//
//nolint:gochecknoglobals,lll
var versionCache sync.Map

func (d *defaultVersionFetcher) GetLatestVersion() (*httpclient.GoVersionInfo, error) {
	info, err := httpclient.GetLatestVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}

	return info, nil
}

// Download downloads the latest Go version and handles display logic.
// It wraps the existing GetLatest functionality with appropriate logging.
// This function creates a new downloader instance and delegates to its GetLatest method.
func Download() (string, string, error) {
	logger.Debug("Starting download operation")

	downloader := NewDownloader(
		&filesystem.OSFileSystem{},
		&DefaultHTTPClient{Client: NewHTTPClient()},
		&exec.OSCommandExecutor{},
		&defaultVersionFetcher{},
	)

	path, checksum, err := downloader.GetLatest("")
	if err != nil {
		logger.Errorf("Error downloading Go archive: %v", err)

		return "", "", err
	}

	logger.Debugf("Download completed: path=%s, checksum=%s", path, checksum)

	return path, checksum, nil
}

// GetLatest downloads the latest stable Go archive for the current platform to the specified directory.
// If destDir is empty, it uses the system's temporary directory.
// It checks for existing archives in common user directories (~/Downloads, ~) and the destination directory,
// prioritizing user directories over the temp directory.
// It verifies the checksum of any found archives.
// If a valid archive exists, it skips the download.
// Otherwise, it downloads the archive to the destination directory and verifies the checksum.
// It returns the path to the file and its checksum, or an error.
// This is a convenience function that creates a new downloader instance.
func GetLatest(destDir string) (string, string, error) {
	downloader := NewDownloader(
		&filesystem.OSFileSystem{},
		&DefaultHTTPClient{Client: NewHTTPClient()},
		&exec.OSCommandExecutor{},
		&defaultVersionFetcher{},
	)

	return downloader.GetLatest(destDir)
}

// GetLatestVersionInfo fetches the latest stable Go version information from the official API.
// It returns the version info for the latest stable version or an error if not found.
// This function wraps the version package's GetLatestVersion with error handling and in-memory caching.
// The fetcher parameter allows for dependency injection, enabling proper testing with mocks.
func GetLatestVersionInfo(fetcher VersionFetcher) (*httpclient.GoVersionInfo, error) {
	// Check cache first with read lock for thread-safe access
	versionCacheMutex.RLock()

	if cached, ok := versionCache.Load("latest"); ok {
		versionCacheMutex.RUnlock()

		if info, ok := cached.(*httpclient.GoVersionInfo); ok {
			logger.Debug("Using cached version info")

			return info, nil
		}
	}

	versionCacheMutex.RUnlock()

	// Fetch from API if not cached
	info, err := fetcher.GetLatestVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}

	// Cache the result for future calls within this execution with write lock
	versionCacheMutex.Lock()
	versionCache.Store("latest", info)
	versionCacheMutex.Unlock()
	logger.Debug("Cached version info for future use")

	return info, nil
}

// RunDownload executes the download command logic.
// It downloads the latest Go version archive and handles any errors.
// This function is used by the CLI command to perform the download operation.
func RunDownload() error {
	downloader := NewDownloader(
		&filesystem.OSFileSystem{},
		&DefaultHTTPClient{Client: NewHTTPClient()},
		&exec.OSCommandExecutor{},
		&defaultVersionFetcher{},
	)

	_, _, err := downloader.GetLatest("")

	return err
}
