// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package update

import (
	"fmt"
	"strings"

	"golang.org/x/mod/semver"

	"github.com/nicholas-fedor/goUpdater/internal/logger"
	"github.com/nicholas-fedor/goUpdater/internal/version"
)

// compare compares two Go version strings.
// It returns -1 if version1 < version2, 0 if version1 == version2, and 1 if version1 > version2.
// If either version is invalid, it returns an error.
// Both versions should be in the format "go1.x.x", "1.x.x", or "v1.x.x" (with or without "go" or "v" prefix).
func compare(version1, version2 string) (int, error) {
	if version1 == "" {
		return 0, fmt.Errorf("version1 cannot be empty: %w", version.ErrVersionParseError)
	}

	if version2 == "" {
		return 0, fmt.Errorf("version2 cannot be empty: %w", version.ErrVersionParseError)
	}

	// Normalize versions by removing "go" and "v" prefixes, then adding "v" prefix for semver
	version1Normalized := "v" + strings.TrimPrefix(strings.TrimPrefix(version1, "go"), "v")
	version2Normalized := "v" + strings.TrimPrefix(strings.TrimPrefix(version2, "go"), "v")

	if !semver.IsValid(version1Normalized) {
		return 0, fmt.Errorf("invalid version %s: %w", version1, version.ErrVersionParseError)
	}

	if !semver.IsValid(version2Normalized) {
		return 0, fmt.Errorf("invalid version %s: %w", version2, version.ErrVersionParseError)
	}

	return semver.Compare(version1Normalized, version2Normalized), nil
}

// needsUpdate determines if an update is required based on semantic version comparison.
// Returns true if no version is installed or if the installed version is older than the latest.
// Uses semantic versioning to properly handle version comparisons like "1.21.0" vs "1.21.10".
func needsUpdate(installedVersion, latestVersionStr string) bool {
	if installedVersion == "" {
		return true
	}

	// Compare versions using the local compare function which handles "go" prefixes properly
	result, err := compare(installedVersion, latestVersionStr)
	if err != nil {
		logger.Errorf("Error comparing versions: %v", err)

		return false
	}

	if result >= 0 {
		logger.Infof("Latest Go version (%s) already installed.", latestVersionStr)

		return false
	}

	logger.Infof("Updating Go from %s to %s", installedVersion, latestVersionStr)

	return true
}
