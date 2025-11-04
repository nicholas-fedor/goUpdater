// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package download provides functionality to download Go archives from official sources.
// It handles version checking, archive retrieval, and integrity verification.
package download

import (
	"github.com/nicholas-fedor/goUpdater/internal/exec"
	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

// newDefaultDownloader creates a new downloader with default dependencies.
// It configures the downloader with OS filesystem, default HTTP client,
// OS command executor, and default version fetcher.
func newDefaultDownloader() *Downloader {
	return NewDownloader(
		&filesystem.OSFileSystem{},
		NewDefaultHTTPClient(),
		&exec.OSCommandExecutor{},
		&defaultVersionFetcher{},
	)
}

// RunDownload executes the download command logic.
// It downloads the latest Go version archive and handles any errors.
// This function is used by the CLI command to perform the download operation.
func RunDownload() error {
	downloader := newDefaultDownloader()

	_, _, err := downloader.GetLatest("")

	return err
}

// Download downloads the latest Go version and handles display logic.
// It wraps the existing GetLatest functionality with appropriate logging.
// This function creates a new downloader instance and delegates to its GetLatest method.
func Download() (string, string, error) {
	logger.Debug("Starting download operation")

	downloader := newDefaultDownloader()

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
	downloader := newDefaultDownloader()

	return downloader.GetLatest(destDir)
}
