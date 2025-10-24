// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package archive provides functionality for extracting and validating Go archives.
package archive

import (
	"path/filepath"
	"strings"

	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	"github.com/nicholas-fedor/goUpdater/internal/logger"
	"golang.org/x/mod/semver"
)

// Validate checks if the archive file exists and is a regular file.
// It returns an error if the archive path does not exist or is not a regular file.
func Validate(archivePath, destDir string) error {
	extractor := NewExtractor(&filesystem.OSFileSystem{}, &DefaultProcessor{})

	return extractor.Validate(archivePath, destDir)
}

// Extract extracts the tar.gz archive to the specified destination directory.
// It validates paths to prevent directory traversal attacks and limits the number of files.
func Extract(archivePath, destDir string) error {
	extractor := NewExtractor(&filesystem.OSFileSystem{}, &DefaultProcessor{})

	return extractor.Extract(archivePath, destDir)
}

// ExtractVersion extracts the Go version from an archive filename.
// It handles both full paths and filenames by extracting the base name.
// The function removes the .tar.gz extension if present, then parses the filename
// to extract the complete version string up to the platform part (e.g., "go1.25.2" from "go1.25.2.linux-amd64.tar.gz").
//
// The parsing logic assumes Go archive filenames follow the standard format:
// "go<major>.<minor>.<patch><suffix>.<platform>-<arch>.tar.gz"
//
// Parameters:
//   - filename: The archive filename or full path to parse
//
// Returns:
//   - The extracted complete version (e.g., "go1.25.2") if parsing succeeds
//   - The original filename (after basename extraction) as fallback if parsing fails
//
// Examples:
//   - "go1.21.0.linux-amd64.tar.gz" -> "go1.21.0"
//   - "/path/to/go1.20.0.darwin-amd64.tar.gz" -> "go1.20.0"
//   - "invalid-filename" -> "invalid-filename"
//
// cyclop:ignore
func ExtractVersion(filename string) string { //nolint:cyclop
	// Extract basename in case a full path is provided
	filename = filepath.Base(filename)
	logger.Debugf("ExtractVersion input filename: %s", filename)

	// Remove .tar.gz extension if present
	if len(filename) > 7 && strings.HasSuffix(filename, ".tar.gz") {
		filename = filename[:len(filename)-7]
	}

	logger.Debugf("ExtractVersion after extension removal: %s", filename)

	// Check for valid go prefix
	if !strings.HasPrefix(filename, "go") {
		logger.Debugf("ExtractVersion fallback: %s", filename)

		return filename
	}

	rest := filename[2:] // part after "go"
	if rest == "" {
		logger.Debugf("ExtractVersion fallback: %s", filename)

		return filename
	}

	// Check if the rest starts with a digit (valid version format)
	if rest[0] < '0' || rest[0] > '9' {
		logger.Debugf("ExtractVersion fallback: %s", filename)

		return filename
	}

	// Split into parts by "."
	parts := strings.Split(rest, ".")

	versionParts := make([]string, 0, len(parts))

	for _, part := range parts {
		// Stop at the first part containing "-" (platform part)
		if strings.Contains(part, "-") {
			break
		}

		versionParts = append(versionParts, part)
	}

	if len(versionParts) == 0 {
		logger.Debugf("ExtractVersion fallback: %s", filename)

		return filename
	}

	version := "go" + strings.Join(versionParts, ".")
	logger.Debugf("ExtractVersion extracted version: %s", version)

	// Validate the extracted version using semver
	if !semver.IsValid("v" + strings.TrimPrefix(version, "go")) {
		logger.Debugf("ExtractVersion invalid version: %s", version)

		return filename
	}

	return version
}

// ValidatePath ensures the extracted path is within the installation directory.
// It returns an error if the target path attempts to traverse outside the allowed directory.
func ValidatePath(targetPath, installDir string) error {
	cleanInstallDir := filepath.Clean(installDir)

	relPath, err := filepath.Rel(cleanInstallDir, targetPath)
	if err != nil {
		return &SecurityError{
			AttemptedPath: targetPath,
			Validation:    "path relativity check",
			Err:           err,
		}
	}

	if strings.HasPrefix(relPath, "..") || filepath.IsAbs(relPath) {
		return &SecurityError{
			AttemptedPath: targetPath,
			Validation:    "directory traversal prevention",
			Err:           ErrInvalidPath,
		}
	}

	return nil
}

// validateHeaderName checks if the tar header name is safe for extraction.
// It prevents directory traversal attacks by ensuring no absolute paths, parent directory references,
// or other dangerous patterns that could lead to path traversal vulnerabilities.
func validateHeaderName(headerName string) error {
	if filepath.IsAbs(headerName) {
		return &SecurityError{
			AttemptedPath: headerName,
			Validation:    "absolute path prevention",
			Err:           ErrInvalidPath,
		}
	}

	if strings.Contains(headerName, "..") {
		return &SecurityError{
			AttemptedPath: headerName,
			Validation:    "parent directory reference prevention",
			Err:           ErrInvalidPath,
		}
	}

	// Additional checks for other dangerous patterns
	if strings.Contains(headerName, "\\") {
		return &SecurityError{
			AttemptedPath: headerName,
			Validation:    "backslash prevention",
			Err:           ErrInvalidPath,
		}
	}

	// Check for null bytes which could be used for path manipulation
	if strings.Contains(headerName, "\x00") {
		return &SecurityError{
			AttemptedPath: headerName,
			Validation:    "null byte prevention",
			Err:           ErrInvalidPath,
		}
	}

	return nil
}
