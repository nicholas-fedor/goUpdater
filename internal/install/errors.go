// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package install

import (
	"errors"
	"fmt"
)

// ErrCreateInstallDir indicates failure to create the installation directory.
var ErrCreateInstallDir = errors.New("failed to create installation directory")

// ErrExtractArchive indicates failure to extract the archive.
var ErrExtractArchive = errors.New("failed to extract archive")

// ErrVerifyInstallation indicates failure in installation verification.
var ErrVerifyInstallation = errors.New("installation verification failed")

// ErrDownloadGo indicates failure to download Go.
var ErrDownloadGo = errors.New("failed to download Go")

// ErrPermissionDenied indicates permission denied during installation.
var ErrPermissionDenied = errors.New("permission denied")

// ErrElevationFailed indicates failure to elevate privileges.
var ErrElevationFailed = errors.New("elevation failed")

// ErrExtractionFailed indicates failure during archive extraction.
var ErrExtractionFailed = errors.New("extraction failed")

// ErrDirectoryCreationFailed indicates failure to create installation directory.
var ErrDirectoryCreationFailed = errors.New("permission denied")

// ErrVerificationFailed indicates failure during installation verification.
var ErrVerificationFailed = errors.New("verification failed")

// ErrInvalidArchive indicates the archive is invalid or corrupted.
var ErrInvalidArchive = errors.New("invalid archive")

// ErrExtractionError indicates an error occurred during extraction.
var ErrExtractionError = errors.New("extraction error")

// ErrTempDirFailed indicates failure to create temporary directory.
var ErrTempDirFailed = errors.New("temp dir failed")

// ErrDownloadFailed indicates failure to download the required files.
var ErrDownloadFailed = errors.New("download failed")

// ErrPathConflict indicates a conflict with existing paths.
var ErrPathConflict = errors.New("path conflict")

// ErrCleanupFailed indicates failure during cleanup operations.
var ErrCleanupFailed = errors.New("cleanup failed")

// InstallError represents installation workflow failures with contextual information.
// revive:disable:exported // This type name is intentionally descriptive for clarity
type InstallError struct {
	Phase     string // Installation phase (e.g., "prepare", "extract", "verify")
	FilePath  string // Relevant file path (archive or installation directory)
	Operation string // Operation being performed
	Err       error  // Underlying error
}

// Error implements the error interface for InstallError.
func (e *InstallError) Error() string {
	return fmt.Sprintf("install failed at %s phase: operation=%s path=%s: %v",
		e.Phase, e.Operation, e.FilePath, e.Err)
}

// Unwrap returns the underlying error for compatibility with errors.Is and errors.As.
func (e *InstallError) Unwrap() error {
	return e.Err
}
