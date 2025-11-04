// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package download

import (
	"bufio"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/nicholas-fedor/goUpdater/internal/exec"
	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	httpclient "github.com/nicholas-fedor/goUpdater/internal/http"
	"github.com/nicholas-fedor/goUpdater/internal/logger"
	"github.com/nicholas-fedor/goUpdater/internal/privileges"
	"github.com/schollz/progressbar/v3"
)

// throttleDuration defines the update interval for the progress bar in milliseconds.
const throttleDuration = 100

// NewDownloader creates a new Downloader with the provided dependencies.
// It initializes the downloader with filesystem, HTTP client, command executor,
// and version fetcher interfaces for testing.
func NewDownloader(
	fileSystem filesystem.FileSystem,
	client HTTPClient,
	executor exec.CommandExecutor,
	versionFetcher VersionFetcher,
) *Downloader {
	return &Downloader{
		fs:             fileSystem,
		client:         client,
		executor:       executor,
		versionFetcher: versionFetcher,
	}
}

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
func (d *Downloader) getPlatformFile(version *httpclient.GoVersionInfo) (*httpclient.GoFileInfo, error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	logger.Debugf("Looking for archive for platform: %s/%s", goos, goarch)

	for _, file := range version.Files {
		if file.OS == goos && file.Arch == goarch && file.Kind == "archive" {
			logger.Debugf("Found matching archive: %s", file.Filename)

			return &file, nil
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

// createDestinationFile creates the destination file for writing the download.
// It uses the injected filesystem interface to create the file at the specified path.
func (d *Downloader) createDestinationFile(destPath string) (io.ReadWriteCloser, error) {
	logger.Debugf("Creating destination file: %s", destPath)

	out, err := d.fs.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	return out, nil
}

// downloadWithoutProgress copies data from the response body to the file without progress tracking.
// This method is used when the server doesn't provide content length information.
func (d *Downloader) downloadWithoutProgress(resp *http.Response, out io.ReadWriteCloser) error {
	logger.Debug("Content length unknown, falling back to simple copy")

	_, err := io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	logger.Debug("File download completed")

	return nil
}

// downloadWithProgress sets up a progress bar and copies data with progress tracking using buffered I/O.
// It creates a progress bar that shows download speed, ETA, and completion percentage.
// The method uses buffered I/O for better memory efficiency during large downloads.
func (d *Downloader) downloadWithProgress(resp *http.Response, out io.ReadWriteCloser, contentLength int64) error {
	logger.Debugf("Content length: %d bytes", contentLength)

	// Create progress bar with description
	bar := progressbar.NewOptions64(contentLength,
		progressbar.OptionSetDescription("Downloading Go archive"),
		progressbar.OptionSetWriter(os.Stderr), // Use stderr to avoid mixing with logs
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionShowCount(),
		progressbar.OptionThrottle(throttleDuration*time.Millisecond),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
	)

	logger.Debug("Starting download with progress tracking")

	// Create a progress reader that updates the bar
	progressReader := progressbar.NewReader(resp.Body, bar)

	// Use buffered I/O for better memory efficiency
	bufferedWriter := bufio.NewWriterSize(out, 64*1024) //nolint:mnd // 64KB buffer

	// Copy data with progress tracking and buffering
	_, err := io.Copy(bufferedWriter, &progressReader)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// Flush the buffer to ensure all data is written
	err = bufferedWriter.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush buffer: %w", err)
	}

	logger.Debug("File download completed with progress tracking")

	return nil
}

// downloadFile downloads a file from the given URL to the specified path with progress tracking.
// It displays download speed, ETA, and completion percentage using a progress bar.
// The method chooses between progress tracking and simple copy based on content length availability.
func (d *Downloader) downloadFile(url, destPath string) error {
	req, err := d.createDownloadRequest(context.Background(), url)
	if err != nil {
		return err
	}

	resp, err := d.executeDownloadRequest(req)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	out, err := d.createDestinationFile(destPath)
	if err != nil {
		return err
	}

	defer func() { _ = out.Close() }()

	// Get content length for progress bar
	contentLength := resp.ContentLength
	if contentLength <= 0 {
		return d.downloadWithoutProgress(resp, out)
	}

	return d.downloadWithProgress(resp, out, contentLength)
}

// verifyChecksum computes the SHA256 checksum of the file using streaming I/O and compares it to the expected value.
// This approach avoids loading the entire file into memory, making it suitable for large archives.
// It uses buffered reading for better performance with large files.
func (d *Downloader) verifyChecksum(filePath, expectedSha256 string) error {
	logger.Debugf("Verifying checksum for file: %s", filePath)

	file, err := d.fs.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	defer func() { _ = file.Close() }()

	hasher := sha256.New()

	logger.Debug("Computing SHA256 hash using streaming I/O")

	// Use buffered reader for better performance with large files
	bufferedReader := bufio.NewReaderSize(file, 64*1024) //nolint:mnd // 64KB buffer

	_, err = io.Copy(hasher, bufferedReader)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	actualSha256 := hex.EncodeToString(hasher.Sum(nil))
	logger.Debugf("Computed hash: %s", actualSha256)

	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(actualSha256), []byte(expectedSha256)) != 1 {
		return &ChecksumError{
			FilePath:       filePath,
			ExpectedSha256: expectedSha256,
			ActualSha256:   actualSha256,
			Err:            ErrChecksumMismatch,
		}
	}

	logger.Debug("Checksum verification passed")

	return nil
}
