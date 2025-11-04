// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package download

import (
	"errors"
	"fmt"

	httpValidation "github.com/nicholas-fedor/goUpdater/internal/http"
)

// ErrNoArchive indicates no archive was found for the platform.
var ErrNoArchive = errors.New("no archive")

// ErrDownloadFailed indicates the download failed.
var ErrDownloadFailed = errors.New("download failed")

// ErrChecksumMismatch indicates a checksum mismatch.
var ErrChecksumMismatch = errors.New("checksum mismatch")

// ErrInvalidURL indicates that the provided URL is invalid or insecure.
var ErrInvalidURL = errors.New("invalid or insecure URL")

// ErrEmptyURL indicates that an empty URL was provided.
var ErrEmptyURL = errors.New("empty URL provided")

// ErrInvalidURLScheme indicates that the URL scheme is not allowed.
var ErrInvalidURLScheme = errors.New("invalid URL scheme")

// ErrInvalidURLHost indicates that the URL host is invalid or not allowed.
var ErrInvalidURLHost = errors.New("invalid or disallowed URL host")

// ErrDirectoryTraversal indicates that the URL contains directory traversal sequences.
var ErrDirectoryTraversal = errors.New("URL contains directory traversal sequences")

// ErrSecurityHeaderValidation indicates security header validation warnings.
var ErrSecurityHeaderValidation = errors.New("security header validation warnings")

// ErrInvalidURLError indicates a general URL parsing error.
var ErrInvalidURLError = errors.New("invalid URL format")

// ErrNotFound indicates that a file or resource was not found.
var ErrNotFound = errors.New("not found")

// ErrVersionFetchFailed indicates that fetching the version information failed.
var ErrVersionFetchFailed = errors.New("version fetch failed")

// Error represents a general download failure with contextual information.
type Error struct {
	URL         string
	Destination string
	Err         error
}

// NetworkError represents HTTP/network specific failures with status codes and response details.
type NetworkError struct {
	StatusCode int
	URL        string
	Response   string
	Err        error
}

// ChecksumError represents checksum verification failures with expected vs actual values.
type ChecksumError struct {
	FilePath       string
	ExpectedSha256 string
	ActualSha256   string
	Err            error
}

// Error implements the error interface for Error.
// Note: This method sanitizes URLs and paths to prevent information disclosure.
func (e *Error) Error() string {
	sanitizedURL := httpValidation.SanitizeURLForError(e.URL)
	sanitizedDest := httpValidation.SanitizePathForError(e.Destination)

	return fmt.Sprintf("download failed: url=%s dest=%s", sanitizedURL, sanitizedDest)
}

// Unwrap returns the underlying error for compatibility with errors.Is and errors.As.
func (e *Error) Unwrap() error {
	return e.Err
}

// Error implements the error interface for NetworkError.
// Note: This method sanitizes URLs to prevent information disclosure.
func (e *NetworkError) Error() string {
	sanitizedURL := httpValidation.SanitizeURLForError(e.URL)
	if e.Response != "" {
		return fmt.Sprintf("network error: status=%d url=%s response=%s", e.StatusCode, sanitizedURL, e.Response)
	}

	return fmt.Sprintf("network error: status=%d url=%s", e.StatusCode, sanitizedURL)
}

// Unwrap returns the underlying error for compatibility with errors.Is and errors.As.
func (e *NetworkError) Unwrap() error {
	return e.Err
}

// Error implements the error interface for ChecksumError.
func (e *ChecksumError) Error() string {
	return fmt.Sprintf("checksum mismatch for %s: expected %s, got %s", e.FilePath, e.ExpectedSha256, e.ActualSha256)
}

// Unwrap returns the underlying error for compatibility with errors.Is and errors.As.
func (e *ChecksumError) Unwrap() error {
	return e.Err
}
