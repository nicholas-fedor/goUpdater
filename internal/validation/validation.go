// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package validation provides comprehensive input validation utilities for goUpdater.
// It includes validation for file paths, archive paths, version strings, and directory paths
// to ensure security and reliability of user inputs.
package validation

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"golang.org/x/mod/semver"
)

// ValidateFilePath validates a file path for security and correctness.
// It checks for length limits, invalid characters, and prevents absolute paths.
func ValidateFilePath(path string) error {
	if path == "" {
		return ErrPathEmpty
	}

	if len(path) > MaxPathLength {
		return fmt.Errorf("file path too long (%d characters): %w", len(path), ErrPathTooLong)
	}

	if filepath.IsAbs(path) {
		return fmt.Errorf("absolute path not allowed: %s: %w", sanitizePath(path), ErrPathAbsolute)
	}

	// Check for dangerous characters that could be used for path manipulation
	if strings.Contains(path, "..") {
		return fmt.Errorf("path contains parent directory references: %s: %w", sanitizePath(path), ErrPathContainsInvalid)
	}

	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path contains null bytes: %s: %w", sanitizePath(path), ErrPathContainsInvalid)
	}

	// Check for backslashes (not allowed on Unix-like systems)
	if strings.Contains(path, "\\") {
		return fmt.Errorf("path contains backslashes: %s: %w", sanitizePath(path), ErrPathContainsInvalid)
	}

	// Ensure the path is valid UTF-8
	if !utf8.ValidString(path) {
		return fmt.Errorf("path contains invalid UTF-8: %s: %w", sanitizePath(path), ErrPathContainsInvalid)
	}

	return nil
}

// ValidateArchivePath validates an archive file path.
// It performs file path validation plus archive-specific checks.
func ValidateArchivePath(archivePath string) error {
	err := ValidateFilePath(archivePath)
	if err != nil {
		return fmt.Errorf("archive path validation failed: %w", err)
	}

	// Check if it has a valid archive extension
	if !strings.HasSuffix(strings.ToLower(archivePath), ".tar.gz") {
		return fmt.Errorf("archive must have .tar.gz extension: %s: %w", archivePath, ErrArchiveInvalid)
	}

	// Ensure the archive name follows Go naming conventions
	base := filepath.Base(archivePath)
	if !strings.HasPrefix(base, "go") {
		return fmt.Errorf("archive filename must start with 'go': %s: %w", base, ErrArchiveInvalid)
	}

	return nil
}

// ValidateVersionString validates a version string for correctness.
// It checks length, format, and semantic versioning compliance.
func ValidateVersionString(version string) error { //nolint:cyclop // detailed version string validation
	if version == "" {
		return ErrVersionEmpty
	}

	if len(version) > MaxVersionLength {
		return fmt.Errorf("version string too long (%d characters): %w", len(version), ErrVersionTooLong)
	}

	// Ensure the version is valid UTF-8
	if !utf8.ValidString(version) {
		return fmt.Errorf("version contains invalid UTF-8: %s: %w", version, ErrVersionInvalid)
	}

	// Check for dangerous characters
	if strings.Contains(version, "\x00") {
		return fmt.Errorf("version contains null bytes: %s: %w", version, ErrVersionInvalid)
	}

	if strings.Contains(version, "\n") || strings.Contains(version, "\r") {
		return fmt.Errorf("version contains newline characters: %s: %w", version, ErrVersionInvalid)
	}

	// Additional validation for Go version format
	if !strings.HasPrefix(version, "go") {
		return fmt.Errorf("version must start with 'go': %s: %w", version, ErrVersionInvalid)
	}

	// Extract the semantic version part after "go"
	semverPart := strings.TrimPrefix(version, "go")
	if semverPart == "" {
		return fmt.Errorf("version missing semantic version after 'go': %s: %w", version, ErrVersionInvalid)
	}

	// Check that the semantic version part contains at least one dot (required for major.minor.patch format)
	if !strings.Contains(semverPart, ".") {
		return fmt.Errorf("version does not follow semantic versioning: %s: %w", version, ErrVersionInvalid)
	}

	// Validate semantic versioning format
	// Allow both "v1.2.3" and "1.2.3" formats
	var semverVersion string
	if strings.HasPrefix(semverPart, "v") {
		semverVersion = semverPart
	} else {
		semverVersion = "v" + semverPart
	}

	if !semver.IsValid(semverVersion) {
		return fmt.Errorf("version does not follow semantic versioning: %s: %w", version, ErrVersionInvalid)
	}

	return nil
}

// ValidateDirectoryPath validates a directory path.
// It performs file path validation plus directory-specific checks.
func ValidateDirectoryPath(dirPath string) error {
	err := ValidateFilePath(dirPath)
	if err != nil {
		return fmt.Errorf("directory path validation failed: %w", err)
	}

	// Check for directory-specific invalid patterns
	if strings.Contains(dirPath, "*") || strings.Contains(dirPath, "?") {
		return fmt.Errorf("directory path contains wildcards: %s: %w", dirPath, ErrDirectoryInvalid)
	}

	// Ensure it doesn't end with a file separator in a way that suggests it's a file
	cleanPath := filepath.Clean(dirPath)
	if strings.HasSuffix(cleanPath, ".tar.gz") || strings.HasSuffix(cleanPath, ".exe") {
		return fmt.Errorf("directory path appears to be a file: %s: %w", dirPath, ErrDirectoryInvalid)
	}

	return nil
}

// SanitizePath sanitizes a path by cleaning it and removing potentially dangerous elements.
// This should be used in conjunction with validation, not as a replacement.
func SanitizePath(path string) string {
	// Clean the path to resolve any . or .. elements
	cleaned := filepath.Clean(path)

	// Remove any leading ./ or .\ to normalize relative paths
	cleaned = strings.TrimPrefix(cleaned, "./")
	cleaned = strings.TrimPrefix(cleaned, ".\\")

	return cleaned
}

// sanitizePath returns a sanitized representation of the path for error messages.
// It returns the basename to provide context without exposing full filesystem paths.
func sanitizePath(path string) string {
	return filepath.Base(path)
}

// IsValidPathChar checks if a character is valid for use in file paths.
// This is a helper function for more complex validation logic.
func IsValidPathChar(r rune) bool { //nolint:cyclop // character validation logic
	// Allow alphanumeric, common path separators, dots, hyphens, underscores
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '/' || r == '\\' || r == '.' || r == '-' || r == '_'
}

// ValidatePathCharacters validates that all characters in a path are safe.
// This provides additional security beyond basic validation.
func ValidatePathCharacters(path string) error {
	for _, r := range path {
		if !IsValidPathChar(r) {
			return fmt.Errorf("path contains invalid character '%c': %s: %w", r, path, ErrPathContainsInvalid)
		}
	}

	return nil
}

// ValidateArchiveName validates an archive filename against Go archive naming conventions.
func ValidateArchiveName(filename string) error {
	if filename == "" {
		return fmt.Errorf("archive filename cannot be empty: %w", ErrArchiveInvalid)
	}

	// Must start with "go"
	if !strings.HasPrefix(filename, "go") {
		return fmt.Errorf("archive filename must start with 'go': %s: %w", filename, ErrArchiveInvalid)
	}

	// Must end with .tar.gz
	if !strings.HasSuffix(filename, ".tar.gz") {
		return fmt.Errorf("archive filename must end with '.tar.gz': %s: %w", filename, ErrArchiveInvalid)
	}

	// Extract version part
	versionPart := strings.TrimSuffix(strings.TrimPrefix(filename, "go"), ".tar.gz")
	if versionPart == "" {
		return fmt.Errorf("archive filename missing version part: %s: %w", filename, ErrArchiveInvalid)
	}

	// Validate version part contains platform info
	parts := strings.Split(versionPart, ".")
	if len(parts) < minVersionParts {
		return fmt.Errorf("archive filename version part too short: %s: %w", versionPart, ErrArchiveInvalid)
	}

	// Check for platform separator
	if !strings.Contains(versionPart, "-") {
		return fmt.Errorf("archive filename missing platform separator: %s: %w", versionPart, ErrArchiveInvalid)
	}

	return nil
}
