// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package archive

import (
	"errors"
	"fmt"
)

// ErrInvalidPath indicates an invalid file path in the archive.
var ErrInvalidPath = errors.New("invalid path")

// ErrArchiveNotRegular indicates the archive path is not a regular file.
var ErrArchiveNotRegular = errors.New("archive path is not a regular file")

// ErrTooManyFiles indicates the archive contains too many files.
var ErrTooManyFiles = errors.New("archive contains too many files")

// ExtractionError represents archive extraction failures with contextual information.
type ExtractionError struct {
	ArchivePath string
	Destination string
	Context     string
	Err         error
}

// SecurityError represents path traversal and security validation failures.
type SecurityError struct {
	AttemptedPath string
	Validation    string
	Err           error
}

// ValidationError represents archive validation failures with file details and criteria.
type ValidationError struct {
	FilePath string
	Criteria string
	Err      error
}

// Error implements the error interface for ExtractionError.
// Note: This method sanitizes paths to prevent information disclosure.
func (e *ExtractionError) Error() string {
	sanitizedArchive := sanitizePathForError(e.ArchivePath)
	sanitizedDest := sanitizePathForError(e.Destination)

	return fmt.Sprintf("extraction failed: archive=%s dest=%s context=%s",
		sanitizedArchive, sanitizedDest, e.Context)
}

// Unwrap returns the underlying error for compatibility with errors.Is and errors.As.
func (e *ExtractionError) Unwrap() error {
	return e.Err
}

// Error implements the error interface for SecurityError.
// Note: This method sanitizes paths to prevent information disclosure.
func (e *SecurityError) Error() string {
	sanitizedPath := sanitizePathForError(e.AttemptedPath)

	return fmt.Sprintf("security error: path=%s validation=%s", sanitizedPath, e.Validation)
}

// Unwrap returns the underlying error for compatibility with errors.Is and errors.As.
func (e *SecurityError) Unwrap() error {
	return e.Err
}

// Error implements the error interface for ValidationError.
// Note: This method does not expose sensitive file system information to prevent information disclosure.
func (e *ValidationError) Error() string {
	// Sanitize file path to prevent information disclosure
	sanitizedPath := sanitizePathForError(e.FilePath)

	return fmt.Sprintf("validation error: file=%s criteria=%s", sanitizedPath, e.Criteria)
}

// Unwrap returns the underlying error for compatibility with errors.Is and errors.As.
func (e *ValidationError) Unwrap() error {
	return e.Err
}

// sanitizePathForError sanitizes file paths in error messages to prevent information disclosure.
// It removes absolute paths and sensitive directory information.
func sanitizePathForError(path string) string {
	if path == "" {
		return "unknown"
	}

	// For security, only show the filename, not the full path
	// This prevents leaking directory structure information
	if lastSlash := findLastSlash(path); lastSlash >= 0 {
		return path[lastSlash+1:]
	}

	return path
}

// findLastSlash finds the last path separator in a string.
func findLastSlash(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '/' || s[i] == '\\' {
			return i
		}
	}

	return -1
}
