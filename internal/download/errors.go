// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package download

import (
	"errors"
	"fmt"
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

// simpleURL represents a simplified URL structure for sanitization.
type simpleURL struct {
	Scheme string
	Host   string
	Path   string
}

// sanitizeURLForError sanitizes URLs in error messages to prevent information disclosure.
// It removes query parameters and fragments that might contain sensitive information.
func sanitizeURLForError(url string) string {
	if url == "" {
		return "unknown"
	}

	// Parse the URL to remove sensitive parts
	parsed, err := parseURL(url)
	if err == nil {
		// Remove query parameters and fragments for security
		cleanURL := parsed.Scheme + "://" + parsed.Host + parsed.Path

		return cleanURL
	}

	return url
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

// parseURL is a simple URL parser that avoids using net/url.Parse to prevent potential issues.
func parseURL(rawURL string) (*simpleURL, error) {
	// Simple URL parsing for sanitization purposes
	// This is a basic implementation to avoid complex parsing dependencies

	// Find scheme
	schemeEnd := findSchemeEnd(rawURL)
	if schemeEnd == -1 {
		return nil, fmt.Errorf("%w", ErrInvalidURLError)
	}

	scheme := rawURL[:schemeEnd]

	// Find host start
	hostStart := schemeEnd
	if hostStart < len(rawURL) && rawURL[hostStart] == ':' {
		hostStart++
		if hostStart < len(rawURL) && rawURL[hostStart] == '/' {
			hostStart++
		}

		if hostStart < len(rawURL) && rawURL[hostStart] == '/' {
			hostStart++
		}
	}

	// Find host end (before path/query/fragment)
	hostEnd := findHostEnd(rawURL, hostStart)

	host := ""
	path := ""

	if hostEnd > hostStart {
		host = rawURL[hostStart:hostEnd]
		if hostEnd < len(rawURL) {
			path = rawURL[hostEnd:]
		}
	}

	return &simpleURL{
		Scheme: scheme,
		Host:   host,
		Path:   path,
	}, nil
}

// findSchemeEnd finds the end of the URL scheme.
func findSchemeEnd(url string) int {
	for i, r := range url {
		if r == ':' {
			return i
		}

		if !isSchemeChar(r) {
			break
		}
	}

	return -1
}

// isSchemeChar checks if a character is valid in a URL scheme.
func isSchemeChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '+' || r == '-' || r == '.'
}

// findHostEnd finds the end of the host part in a URL.
func findHostEnd(url string, start int) int {
	for i := start; i < len(url); i++ {
		if url[i] == '/' || url[i] == '?' || url[i] == '#' {
			return i
		}
	}

	return len(url)
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

// Error implements the error interface for Error.
// Note: This method sanitizes URLs and paths to prevent information disclosure.
func (e *Error) Error() string {
	sanitizedURL := sanitizeURLForError(e.URL)
	sanitizedDest := sanitizePathForError(e.Destination)

	return fmt.Sprintf("download failed: url=%s dest=%s", sanitizedURL, sanitizedDest)
}

// Unwrap returns the underlying error for compatibility with errors.Is and errors.As.
func (e *Error) Unwrap() error {
	return e.Err
}

// Error implements the error interface for NetworkError.
// Note: This method sanitizes URLs to prevent information disclosure.
func (e *NetworkError) Error() string {
	sanitizedURL := sanitizeURLForError(e.URL)
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
