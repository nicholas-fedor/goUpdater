// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"net/url"
	"strings"
)

// SanitizeURLForError sanitizes URLs in error messages to prevent information disclosure.
// It removes query parameters and fragments that might contain sensitive information.
func SanitizeURLForError(rawURL string) string {
	if rawURL == "" {
		return "unknown"
	}

	// Parse the URL to remove sensitive parts
	parsed, err := url.Parse(rawURL)
	if err == nil {
		// Remove query parameters and fragments for security
		cleanedPath := parsed.Path
		if idx := strings.IndexAny(cleanedPath, "?#"); idx != -1 {
			cleanedPath = cleanedPath[:idx]
		}

		cleanURL := parsed.Scheme + "://" + parsed.Host + cleanedPath

		return cleanURL
	}

	return rawURL
}

// SanitizePathForError sanitizes file paths in error messages to prevent information disclosure.
// It removes absolute paths and sensitive directory information.
func SanitizePathForError(path string) string {
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
