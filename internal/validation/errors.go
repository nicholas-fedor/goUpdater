// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package validation

import "errors"

// Common validation errors.
var (
	ErrPathTooLong         = errors.New("path exceeds maximum allowed length")
	ErrPathContainsInvalid = errors.New("path contains invalid characters")
	ErrPathAbsolute        = errors.New("absolute paths are not allowed")
	ErrPathEmpty           = errors.New("path cannot be empty")
	ErrVersionTooLong      = errors.New("version string exceeds maximum allowed length")
	ErrVersionInvalid      = errors.New("version string is not valid")
	ErrVersionEmpty        = errors.New("version string cannot be empty")
	ErrArchiveInvalid      = errors.New("archive path is not valid")
	ErrDirectoryInvalid    = errors.New("directory path is not valid")
)
