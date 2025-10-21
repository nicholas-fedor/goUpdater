// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package archive

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"golang.org/x/mod/semver"

	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

// NewExtractor creates a new Extractor with the given dependencies.
func NewExtractor(fs filesystem.FileSystem, processor Processor) *Extractor {
	return &Extractor{
		fs:        fs,
		processor: processor,
	}
}

// Extract extracts the tar.gz archive to the specified destination directory.
// It validates paths to prevent directory traversal attacks and limits the number of files.
// The file count limit is set to 20,000 to accommodate legitimate Go archives while preventing zip bomb attacks.
//
//nolint:cyclop,funlen // complex archive extraction with security validations
func (e *Extractor) Extract(archivePath, destDir string) error {
	archivePath = filepath.Clean(archivePath)
	logger.Debugf("Extract archive: %s", archivePath)

	err := e.Validate(archivePath)
	if err != nil {
		return err
	}

	file, err := e.fs.Open(archivePath)
	if err != nil {
		return &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "opening archive file",
			Err:         err,
		}
	}

	defer func() { _ = file.Close() }()

	gzipReader, err := e.processor.NewGzipReader(file)
	if err != nil {
		return &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "creating gzip reader",
			Err:         err,
		}
	}

	defer func() { _ = gzipReader.Close() }()

	tarReader := e.processor.NewTarReader(gzipReader)

	// Limit the number of files to prevent zip bomb attacks
	// Set to 20,000 to accommodate Go archives which contain ~16k files with generous buffer
	const maxFiles = 20000
	// Limit total extracted size to prevent zip bomb attacks
	const maxTotalSize = 500 * 1024 * 1024 // 500MB limit (Go archives are ~196MB)
	// Limit individual file size to prevent zip bomb attacks
	const maxFileSize = 50 * 1024 * 1024 // 50MB per file limit (largest Go file is ~20.6MB)

	fileCount := 0
	totalSize := int64(0)

	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return &ExtractionError{
				ArchivePath: archivePath,
				Destination: destDir,
				Context:     "reading tar header",
				Err:         err,
			}
		}

		fileCount++
		if fileCount > maxFiles {
			return fmt.Errorf("archive contains too many files: %w", ErrTooManyFiles)
		}

		// Check for zip bomb: extremely large files or excessive total size
		if header.Size > maxFileSize {
			return fmt.Errorf("archive contains file too large: %s (%d bytes): %w", header.Name, header.Size, ErrTooManyFiles)
		}

		totalSize += header.Size
		if totalSize > maxTotalSize {
			return fmt.Errorf("archive total size too large: %d bytes: %w", totalSize, ErrTooManyFiles)
		}

		err = e.processTarEntry(tarReader, header, destDir)
		if err != nil {
			return err
		}
	}

	logger.Debugf("Successfully extracted archive: %s", archivePath)

	return nil
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
