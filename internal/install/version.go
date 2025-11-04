package install

import (
	"fmt"
	"strings"

	"github.com/nicholas-fedor/goUpdater/internal/download"
	httpclient "github.com/nicholas-fedor/goUpdater/internal/http"
	"github.com/nicholas-fedor/goUpdater/internal/types"
	"golang.org/x/mod/semver"
)

func (d *defaultVersionFetcherImpl) GetLatestVersion() (*httpclient.GoVersionInfo, error) {
	info, err := httpclient.GetLatestVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}

	return info, nil
}

func (d *downloadServiceImpl) GetLatest(tempDir string) (string, string, error) {
	archivePath, checksum, err := d.downloader.GetLatest(tempDir)
	if err != nil {
		return "", "", fmt.Errorf("failed to download latest: %w", err)
	}

	return archivePath, checksum, nil
}

// GetLatestVersionInfo retrieves the latest version information from the download service.
// It returns a VersionInfo struct containing the version string on success.
// On error, it returns the zero value of VersionInfo and the error.
func (d *downloadServiceImpl) GetLatestVersionInfo() (types.VersionInfo, error) {
	info, err := download.GetLatestVersionInfo(&defaultVersionFetcherImpl{})
	if err != nil {
		return types.VersionInfo{}, fmt.Errorf("failed to get latest version info: %w", err)
	}

	return types.VersionInfo{Version: info.Version}, nil
}

func (v *versionServiceImpl) Compare(installedVersion, latestVersion string) (int, error) {
	result, err := compare(installedVersion, latestVersion)
	if err != nil {
		return 0, fmt.Errorf("failed to compare versions: %w", err)
	}

	return result, nil
}

// compare compares two Go version strings.
// It returns -1 if version1 < version2, 0 if version1 == version2, and 1 if version1 > version2.
// If either version is invalid, it returns an error.
// Both versions should be in the format "go1.x.x", "1.x.x", or "v1.x.x" (with or without "go" or "v" prefix).
func compare(version1, version2 string) (int, error) {
	if version1 == "" {
		return 0, fmt.Errorf("version1 cannot be empty: %w", errVersionParseError)
	}

	if version2 == "" {
		return 0, fmt.Errorf("version2 cannot be empty: %w", errVersionParseError)
	}

	// Normalize versions to semver: first strip "go" prefix if present,
	// then strip "v" prefix from remainder, finally prefix "v".
	// Examples: "go1.2.3" -> "v1.2.3", "v1.2.3" -> "v1.2.3", "1.2.3" -> "v1.2.3".
	version1Normalized := "v" + strings.TrimPrefix(strings.TrimPrefix(version1, "go"), "v")
	version2Normalized := "v" + strings.TrimPrefix(strings.TrimPrefix(version2, "go"), "v")

	if !semver.IsValid(version1Normalized) {
		return 0, fmt.Errorf("invalid version %s: %w", version1, errVersionParseError)
	}

	if !semver.IsValid(version2Normalized) {
		return 0, fmt.Errorf("invalid version %s: %w", version2, errVersionParseError)
	}

	return semver.Compare(version1Normalized, version2Normalized), nil
}
