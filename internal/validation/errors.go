// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package validation

import "errors"

// Common validation errors.
var (
	// ErrPathTooLong is returned when the path exceeds the maximum allowed length.
	ErrPathTooLong = errors.New("path exceeds maximum allowed length")
	// ErrPathContainsInvalid is returned when the path contains invalid characters.
	ErrPathContainsInvalid = errors.New("path contains invalid characters")
	// ErrPathAbsolute is returned when absolute paths are not allowed.
	ErrPathAbsolute = errors.New("absolute paths are not allowed")
	// ErrPathEmpty is returned when the path cannot be empty.
	ErrPathEmpty = errors.New("path cannot be empty")
	// ErrVersionTooLong is returned when the version string exceeds the maximum allowed length.
	ErrVersionTooLong = errors.New("version string exceeds maximum allowed length")
	// ErrVersionInvalid is returned when the version string is not valid.
	ErrVersionInvalid = errors.New("version string is not valid")
	// ErrVersionEmpty is returned when the version string cannot be empty.
	ErrVersionEmpty = errors.New("version string cannot be empty")
	// ErrArchiveInvalid is returned when the archive path is not valid.
	ErrArchiveInvalid = errors.New("archive path is not valid")
	// ErrDirectoryInvalid is returned when the directory path is not valid.
	ErrDirectoryInvalid = errors.New("directory path is not valid")
)
