// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package download

import (
	"fmt"
	"path/filepath"
	"runtime"

	httpclient "github.com/nicholas-fedor/goUpdater/internal/http"
	"github.com/nicholas-fedor/goUpdater/internal/logger"
	"github.com/nicholas-fedor/goUpdater/internal/privileges"
)

// GetLatest downloads the latest stable Go archive for the current platform to the specified directory.
// If destDir is empty, it uses the system's temporary directory.
// It checks for existing archives in common user directories (~/Downloads, ~) and the destination directory,
// prioritizing user directories over the temp directory.
// It verifies the checksum of any found archives.
// If a valid archive exists, it skips the download.
// Otherwise, it downloads the archive to the destination directory and verifies the checksum.
// It returns the path to the file and its checksum, or an error.
func (d *Downloader) GetLatest(destDir string) (string, string, error) {
	if destDir == "" {
		destDir = d.fs.TempDir()
		logger.Debugf("Using temporary directory: %s", destDir)
	}

	logger.Debugf("Starting download of latest Go archive to: %s", destDir)

	version, err := d.getLatestVersion()
	if err != nil {
		return "", "", fmt.Errorf("failed to get latest version: %w", err)
	}

	file, err := d.getPlatformFile(version)
	if err != nil {
		return "", "", fmt.Errorf("failed to get platform file: %w", err)
	}

	home, err := d.fs.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("failed to get home directory: %w", err)
	}

	// Check for existing archives in user directories first, then destination directory
	searchDirs := privileges.GetSearchDirectories(home, destDir, d.fs)
	for _, dir := range searchDirs {
		candidatePath := filepath.Join(dir, file.Filename)
		if d.checkExistingArchive(candidatePath, file.Sha256) {
			logger.Infof("Valid Go archive already exists at %s", candidatePath)
			logger.Infof("SHA256 checksum: %s...", file.Sha256[:12])

			return candidatePath, file.Sha256, nil
		}
	}

	url := "https://go.dev/dl/" + file.Filename
	destPath := filepath.Join(destDir, file.Filename)

	err = d.downloadAndVerify(url, destPath, file.Sha256)
	if err != nil {
		return "", "", &Error{
			URL:         url,
			Destination: destPath,
			Err:         err,
		}
	}

	logger.Infof("Successfully downloaded Go archive to: %s", destPath)
	logger.Infof("SHA256 checksum: %s...", file.Sha256[:12])

	return destPath, file.Sha256, nil
}

// getLatestVersion fetches the latest stable Go version information from the official API.
// It returns the version info for the current platform or an error if not found.
// This method delegates to the injected version fetcher.
func (d *Downloader) getLatestVersion() (*httpclient.GoVersionInfo, error) {
	info, err := d.versionFetcher.GetLatestVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}

	return info, nil
}

// getPlatformFile finds the archive file for the current platform from the version info.
// It iterates through the available files and returns the one matching the current OS, architecture, and archive kind.
func (d *Downloader) getPlatformFile(version *httpclient.GoVersionInfo) (*GoFileInfo, error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	logger.Debugf("Looking for archive for platform: %s/%s", goos, goarch)

	for _, file := range version.Files {
		if file.OS == goos && file.Arch == goarch && file.Kind == "archive" {
			logger.Debugf("Found matching archive: %s", file.Filename)

			return &GoFileInfo{
				Filename: file.Filename,
				OS:       file.OS,
				Arch:     file.Arch,
				Version:  file.Version,
				Sha256:   file.Sha256,
				Size:     file.Size,
				Kind:     file.Kind,
			}, nil
		}
	}

	return nil, fmt.Errorf("no archive found for %s/%s: %w", goos, goarch, ErrNoArchive)
}

// checkExistingArchive checks if the archive already exists at the given path and verifies its checksum.
// It returns true if the archive exists and is valid, false otherwise.
// If the archive exists but checksum is invalid, it removes the file to prevent using corrupted archives.
func (d *Downloader) checkExistingArchive(destPath, expectedSha256 string) bool {
	logger.Debugf("Checking if archive already exists at: %s", destPath)

	_, err := d.fs.Stat(destPath)
	if err != nil {
		if d.fs.IsNotExist(err) {
			logger.Debug("Archive does not exist")

			return false
		}

		logger.Debugf("Failed to check existing file: %v", err)

		return false
	}

	logger.Debug("Archive exists, verifying checksum")

	err = d.verifyChecksum(destPath, expectedSha256)
	if err != nil {
		logger.Debug("Existing archive checksum verification failed, removing invalid file")

		_ = d.fs.RemoveAll(destPath)

		return false
	}

	logger.Debug("Existing archive checksum verification successful")

	return true
}

// downloadAndVerify downloads the file from the given URL to the destination path and verifies its checksum.
// It removes the file if verification fails to prevent using corrupted downloads.
func (d *Downloader) downloadAndVerify(url, destPath, expectedSha256 string) error {
	logger.Debugf("Downloading from URL: %s to %s", url, destPath)

	err := d.downloadFile(url, destPath)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	logger.Debug("Download completed, verifying checksum")

	err = d.verifyChecksum(destPath, expectedSha256)
	if err != nil {
		logger.Debug("Checksum verification failed, cleaning up")

		_ = d.fs.RemoveAll(destPath) // Clean up on verification failure

		return fmt.Errorf("checksum verification failed: %w", err)
	}

	logger.Debug("Checksum verification successful")

	return nil
}
