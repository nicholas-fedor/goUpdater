// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package archive provides utilities for handling Go archive files.
// It includes functions for extracting, validating, and processing Go installation archives.
package archive

import (
	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
)

// Validate checks if the archive file exists and is a regular file.
// It returns an error if the archive path does not exist or is not a regular file.
func Validate(archivePath string) error {
	extractor := NewExtractor(&filesystem.OSFileSystem{}, &DefaultProcessor{})

	return extractor.Validate(archivePath)
}

// Extract extracts the tar.gz archive to the specified destination directory.
// It validates paths to prevent directory traversal attacks and limits the number of files.
func Extract(archivePath, destDir string) error {
	extractor := NewExtractor(&filesystem.OSFileSystem{}, &DefaultProcessor{})

	return extractor.Extract(archivePath, destDir)
}
