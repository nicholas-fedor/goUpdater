// Package download provides functionality to download Go archives from official sources.
// It handles version checking, archive retrieval, and integrity verification.
package download

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/nicholas-fedor/goUpdater/internal/logger"
	"github.com/schollz/progressbar/v3"
)

// throttleDuration defines the update interval for the progress bar in milliseconds.
const throttleDuration = 100 // Progress bar update interval in milliseconds

// errUnexpectedStatus indicates an unexpected HTTP status code.
var errUnexpectedStatus = errors.New("unexpected status")

// errNoStableVersion indicates no stable Go version was found.
var errNoStableVersion = errors.New("no stable version found")

// errNoArchive indicates no archive was found for the platform.
var errNoArchive = errors.New("no archive")

// errDownloadFailed indicates the download failed.
var errDownloadFailed = errors.New("download failed")

// errChecksumMismatch indicates a checksum mismatch.
var errChecksumMismatch = errors.New("checksum mismatch")

// GoVersionInfo represents the structure of a Go version from the official API.
type GoVersionInfo struct {
	Version string       `json:"version"`
	Stable  bool         `json:"stable"`
	Files   []goFileInfo `json:"files"`
}

// goFileInfo represents a file in a Go version.
type goFileInfo struct {
	Filename string `json:"filename"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Version  string `json:"version"`
	Sha256   string `json:"sha256"`
	Size     int    `json:"size"`
	Kind     string `json:"kind"`
}

// Download downloads the latest Go version and handles display logic.
// It wraps the existing GetLatest functionality with appropriate logging.
func Download() (string, string, error) {
	logger.Debug("Starting download operation")

	path, checksum, err := GetLatest("")
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
func GetLatest(destDir string) (string, string, error) {
	if destDir == "" {
		destDir = os.TempDir()
		logger.Debugf("Using temporary directory: %s", destDir)
	}

	logger.Debugf("Starting download of latest Go archive to: %s", destDir)

	version, err := getLatestVersion()
	if err != nil {
		return "", "", fmt.Errorf("failed to get latest version: %w", err)
	}

	file, err := getPlatformFile(version)
	if err != nil {
		return "", "", fmt.Errorf("failed to get platform file: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("failed to get home directory: %w", err)
	}

	// Check for existing archives in user directories first, then destination directory
	searchDirs := getSearchDirectories(home, destDir)
	for _, dir := range searchDirs {
		candidatePath := filepath.Join(dir, file.Filename)
		if checkExistingArchive(candidatePath, file.Sha256) {
			logger.Infof("Valid Go archive already exists at %s", candidatePath)
			logger.Infof("SHA256 checksum: %s...", file.Sha256[:12])

			return candidatePath, file.Sha256, nil
		}
	}

	url := "https://go.dev/dl/" + file.Filename
	destPath := filepath.Join(destDir, file.Filename)

	err = downloadAndVerify(url, destPath, file.Sha256)
	if err != nil {
		return "", "", err
	}

	logger.Infof("Successfully downloaded Go archive to: %s", destPath)
	logger.Infof("SHA256 checksum: %s...", file.Sha256[:12])

	return destPath, file.Sha256, nil
}

// GetLatestVersionInfo fetches the latest stable Go version information from the official API.
// It returns the version info for the latest stable version or an error if not found.
func GetLatestVersionInfo() (*GoVersionInfo, error) {
	return getLatestVersion()
}

// getLatestVersion fetches the latest stable Go version information from the official API.
// It returns the version info for the current platform or an error if not found.
func getLatestVersion() (*GoVersionInfo, error) {
	logger.Debug("Fetching latest Go version information from official API")

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://go.dev/dl/?mode=json", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch version info: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d: %w", resp.StatusCode, errUnexpectedStatus)
	}

	var versions []GoVersionInfo

	err = json.NewDecoder(resp.Body).Decode(&versions)
	if err != nil {
		return nil, fmt.Errorf("failed to decode version info: %w", err)
	}

	// Find the latest stable version
	for _, v := range versions {
		if v.Stable {
			logger.Debugf("Found stable version: %s", v.Version)

			return &v, nil
		}
	}

	return nil, errNoStableVersion
}

// getPlatformFile finds the archive file for the current platform from the version info.
func getPlatformFile(version *GoVersionInfo) (*goFileInfo, error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	logger.Debugf("Looking for archive for platform: %s/%s", goos, goarch)

	for _, file := range version.Files {
		if file.OS == goos && file.Arch == goarch && file.Kind == "archive" {
			logger.Debugf("Found matching archive: %s", file.Filename)

			return &file, nil
		}
	}

	return nil, fmt.Errorf("no archive found for %s/%s: %w", goos, goarch, errNoArchive)
}

// getSearchDirectories determines the directories to search for existing archives.
// When running with elevated privileges, it includes both the elevated user's directories
// and the original user's directories to find user-downloaded archives.
// It only includes directories that are readable to maintain security.
func getSearchDirectories(elevatedHome, destDir string) []string {
	var searchDirs []string

	logger.Debugf("Building search directories: elevatedHome=%s, destDir=%s, isElevated=%t",
		elevatedHome, destDir, isElevated())

	// Always include elevated user's directories
	if isReadableDir(filepath.Join(elevatedHome, "Downloads")) {
		searchDirs = append(searchDirs, filepath.Join(elevatedHome, "Downloads"))
		logger.Debugf("Added elevated user's Downloads directory: %s", filepath.Join(elevatedHome, "Downloads"))
	} else {
		logger.Debugf("Elevated user's Downloads directory not readable: %s", filepath.Join(elevatedHome, "Downloads"))
	}

	if isReadableDir(elevatedHome) {
		searchDirs = append(searchDirs, elevatedHome)
		logger.Debugf("Added elevated user's home directory: %s", elevatedHome)
	} else {
		logger.Debugf("Elevated user's home directory not readable: %s", elevatedHome)
	}

	// Check if running with elevated privileges and add original user's directories
	if isElevated() {
		addOriginalUserDirs(&searchDirs)
	}

	// Always include destination directory
	searchDirs = append(searchDirs, destDir)
	logger.Debugf("Added destination directory: %s", destDir)

	logger.Debugf("Final search directories: %v", searchDirs)

	return searchDirs
}

// addOriginalUserDirs adds the original user's directories to the search list if available.
func addOriginalUserDirs(searchDirs *[]string) {
	originalHome := getOriginalUserHome()
	logger.Debugf("Original user home directory: %s", originalHome)

	if originalHome == "" {
		logger.Debug("No original user home directory found")

		return
	}

	if isReadableDir(filepath.Join(originalHome, "Downloads")) {
		*searchDirs = append(*searchDirs, filepath.Join(originalHome, "Downloads"))
		logger.Debugf("Added original user's Downloads directory: %s", filepath.Join(originalHome, "Downloads"))
	} else {
		logger.Debugf("Original user's Downloads directory not readable: %s", filepath.Join(originalHome, "Downloads"))
	}

	if isReadableDir(originalHome) {
		*searchDirs = append(*searchDirs, originalHome)
		logger.Debugf("Added original user's home directory: %s", originalHome)
	} else {
		logger.Debugf("Original user's home directory not readable: %s", originalHome)
	}
}

// isElevated checks if the process is running with elevated privileges via sudo.
// It returns true if the SUDO_USER environment variable is set, indicating
// the process was started with sudo by a different user.
func isElevated() bool {
	return os.Getenv("SUDO_USER") != ""
}

// getOriginalUserHome retrieves the original user's home directory from the SUDO_USER environment variable.
func getOriginalUserHome() string {
	sudoUser := os.Getenv("SUDO_USER")
	logger.Debugf("SUDO_USER environment variable: %s", sudoUser)

	if sudoUser == "" {
		logger.Debug("SUDO_USER environment variable is empty")

		return ""
	}

	originalUser, err := user.Lookup(sudoUser)
	if err != nil {
		logger.Debugf("Failed to lookup user %s: %v", sudoUser, err)

		return ""
	}

	logger.Debugf("Original user home directory resolved: %s", originalUser.HomeDir)

	return originalUser.HomeDir
}

// isReadableDir checks if a directory exists and is readable.
func isReadableDir(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}

	return info.IsDir()
}

// checkExistingArchive checks if the archive already exists at the given path and verifies its checksum.
// It returns true if the archive exists and is valid, false otherwise.
// If the archive exists but checksum is invalid, it removes the file.
func checkExistingArchive(destPath, expectedSha256 string) bool {
	logger.Debugf("Checking if archive already exists at: %s", destPath)

	_, err := os.Stat(destPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Debug("Archive does not exist")

			return false
		}

		logger.Debugf("Failed to check existing file: %v", err)

		return false
	}

	logger.Debug("Archive exists, verifying checksum")

	err = verifyChecksum(destPath, expectedSha256)
	if err != nil {
		logger.Debug("Existing archive checksum verification failed, removing invalid file")

		_ = os.Remove(destPath)

		return false
	}

	logger.Debug("Existing archive checksum verification successful")

	return true
}

// downloadAndVerify downloads the file from the given URL to the destination path and verifies its checksum.
// It removes the file if verification fails.
func downloadAndVerify(url, destPath, expectedSha256 string) error {
	logger.Debugf("Downloading from URL: %s to %s", url, destPath)

	err := downloadFile(url, destPath)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	logger.Debug("Download completed, verifying checksum")

	err = verifyChecksum(destPath, expectedSha256)
	if err != nil {
		logger.Debug("Checksum verification failed, cleaning up")

		_ = os.Remove(destPath) // Clean up on verification failure

		return fmt.Errorf("checksum verification failed: %w", err)
	}

	logger.Debug("Checksum verification successful")

	return nil
}

// createDownloadRequest creates an HTTP GET request for the given URL with context.
func createDownloadRequest(url string) (*http.Request, error) {
	logger.Debugf("Creating HTTP request for: %s", url)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return req, nil
}

// executeDownloadRequest executes the HTTP request and returns the response.
// It ensures the response body is closed on error.
func executeDownloadRequest(req *http.Request) (*http.Response, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer func() { _ = resp.Body.Close() }()

		return nil, fmt.Errorf("download failed with status: %d: %w", resp.StatusCode, errDownloadFailed)
	}

	return resp, nil
}

// createDestinationFile creates the destination file for writing the download.
// It includes a gosec comment for the linter.
func createDestinationFile(destPath string) (*os.File, error) {
	logger.Debugf("Creating destination file: %s", destPath)

	out, err := os.Create(destPath) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	// gosec: G304 - Potential file inclusion via variable is acceptable here as we control the destPath

	return out, nil
}

// downloadWithoutProgress copies data from the response body to the file without progress tracking.
func downloadWithoutProgress(resp *http.Response, out *os.File) error {
	logger.Debug("Content length unknown, falling back to simple copy")

	_, err := io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	logger.Debug("File download completed")

	return nil
}

// downloadWithProgress sets up a progress bar and copies data with progress tracking.
func downloadWithProgress(resp *http.Response, out *os.File, contentLength int64) error {
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

	// Copy data with progress tracking
	_, err := io.Copy(out, &progressReader)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	logger.Debug("File download completed with progress tracking")

	return nil
}

// downloadFile downloads a file from the given URL to the specified path with progress tracking.
// It displays download speed, ETA, and completion percentage using a progress bar.
func downloadFile(url, destPath string) error {
	req, err := createDownloadRequest(url)
	if err != nil {
		return err
	}

	resp, err := executeDownloadRequest(req)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	out, err := createDestinationFile(destPath)
	if err != nil {
		return err
	}

	defer func() { _ = out.Close() }()

	// Get content length for progress bar
	contentLength := resp.ContentLength
	if contentLength <= 0 {
		return downloadWithoutProgress(resp, out)
	}

	return downloadWithProgress(resp, out, contentLength)
}

// verifyChecksum computes the SHA256 checksum of the file and compares it to the expected value.
func verifyChecksum(filePath, expectedSha256 string) error {
	logger.Debugf("Verifying checksum for file: %s", filePath)

	file, err := os.Open(filePath) //nolint:gosec
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	defer func() { _ = file.Close() }()

	// gosec: G304 - Potential file inclusion via variable is acceptable here as we control the filePath

	hasher := sha256.New()

	logger.Debug("Computing SHA256 hash")

	_, err = io.Copy(hasher, file)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	actualSha256 := hex.EncodeToString(hasher.Sum(nil))
	logger.Debugf("Computed hash: %s", actualSha256)

	if !strings.EqualFold(actualSha256, expectedSha256) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s: %w", expectedSha256, actualSha256, errChecksumMismatch)
	}

	logger.Debug("Checksum verification passed")

	return nil
}
