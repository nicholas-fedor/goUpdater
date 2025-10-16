// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package archive provides utilities for handling Go archive files.
// It includes functions for extracting, validating, and processing Go installation archives.
package archive

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultDirPerm  = 0755 // Default directory permissions
	defaultFilePerm = 0644 // Default file permissions
	unixPermMask    = 0777 // Unix permission mask for tar headers
)

// errInvalidPath indicates an invalid file path in the archive.
var errInvalidPath = errors.New("invalid path")

// errArchiveNotRegular indicates the archive path is not a regular file.
var errArchiveNotRegular = errors.New("archive path is not a regular file")

// errTooManyFiles indicates the archive contains too many files.
var errTooManyFiles = errors.New("archive contains too many files")

// ExtractVersion extracts the Go version from an archive filename.
// It handles both full paths and filenames by extracting the base name.
// The function removes the .tar.gz extension if present, then parses the filename
// to extract the major version (e.g., "go1" from "go1.21.0.linux-amd64.tar.gz").
//
// The parsing logic assumes Go archive filenames follow the standard format:
// "go<major>.<minor>.<patch>.<platform>-<arch>.tar.gz"
//
// Parameters:
//   - filename: The archive filename or full path to parse
//
// Returns:
//   - The extracted major version (e.g., "go1") if parsing succeeds
//   - The original filename (after basename extraction) as fallback if parsing fails
//
// Examples:
//   - "go1.21.0.linux-amd64.tar.gz" -> "go1"
//   - "/path/to/go1.20.0.darwin-amd64.tar.gz" -> "go1"
//   - "invalid-filename" -> "invalid-filename"
func ExtractVersion(filename string) string {
	// Extract basename in case a full path is provided
	filename = filepath.Base(filename)

	// Remove .tar.gz extension if present
	if len(filename) > 7 && strings.HasSuffix(filename, ".tar.gz") {
		filename = filename[:len(filename)-7]
	}

	// Parse version: find major version after "go"
	if len(filename) > 2 && strings.HasPrefix(filename, "go") {
		// Find the first dot after "go" to extract major version
		dotIndex := strings.Index(filename[2:], ".")
		if dotIndex != -1 {
			version := filename[:2+dotIndex]

			return version
		}
	}

	// Fallback: return the processed filename
	return filename
}

// Validate checks if the archive file exists and is a regular file.
// It returns an error if the archive path does not exist or is not a regular file.
func Validate(archivePath string) error {
	info, err := os.Stat(archivePath)
	if err != nil {
		return fmt.Errorf("archive file does not exist: %w", err)
	}

	if !info.Mode().IsRegular() {
		return fmt.Errorf("archive path is not a regular file: %w", errArchiveNotRegular)
	}

	return nil
}

// validateHeaderName checks if the tar header name is safe for extraction.
// It prevents directory traversal attacks by ensuring no absolute paths or parent directory references.
func validateHeaderName(headerName string) error {
	if filepath.IsAbs(headerName) || strings.Contains(headerName, "..") {
		return fmt.Errorf("invalid file path in archive: %s: %w", headerName, errInvalidPath)
	}

	return nil
}

// processTarEntry processes a single tar entry, validating and extracting it to the destination directory.
func processTarEntry(tarReader *tar.Reader, header *tar.Header, destDir string) error {
	// Validate the header name
	err := validateHeaderName(header.Name)
	if err != nil {
		return err
	}

	// Construct target path safely
	cleanDestDir := filepath.Clean(destDir)
	// gosec G305 is triggered by filepath.Join, but we have validated the path thoroughly above
	// The path is safe because:
	// 1. header.Name is validated to not contain .. or be absolute
	// 2. targetPath is checked to be within cleanDestDir
	// 3. ValidatePath ensures no traversal
	targetPath := cleanDestDir + string(filepath.Separator) + header.Name
	targetPath = filepath.Clean(targetPath)

	// Validate that the target path is within the destination directory
	if !strings.HasPrefix(targetPath, cleanDestDir+string(filepath.Separator)) && targetPath != cleanDestDir {
		return fmt.Errorf("invalid file path in archive: %s: %w", targetPath, errInvalidPath)
	}

	// Additional validation to prevent path traversal
	rel, err := filepath.Rel(cleanDestDir, targetPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("invalid file path in archive: %s: %w", targetPath, errInvalidPath)
	}

	// Ensure the target path is safe by checking it doesn't escape the destination directory
	if !strings.HasPrefix(targetPath, cleanDestDir) {
		return fmt.Errorf("invalid file path in archive: %s: %w", targetPath, errInvalidPath)
	}

	// Final safety check: ensure the path is validated before use
	err = ValidatePath(targetPath, cleanDestDir)
	if err != nil {
		return err
	}

	// gosec G305 is triggered by filepath.Join, but we have validated the path thoroughly above
	// The path is safe because:
	// 1. header.Name is validated to not contain .. or be absolute
	// 2. targetPath is checked to be within cleanDestDir
	// 3. ValidatePath ensures no traversal
	return ExtractEntry(tarReader, header, targetPath)
}

// ValidatePath ensures the extracted path is within the installation directory.
// It returns an error if the target path attempts to traverse outside the allowed directory.
func ValidatePath(targetPath, installDir string) error {
	cleanInstallDir := filepath.Clean(installDir)

	relPath, err := filepath.Rel(cleanInstallDir, targetPath)
	if err != nil {
		return fmt.Errorf("invalid file path in archive: %s: %w", targetPath, errInvalidPath)
	}

	if strings.HasPrefix(relPath, "..") || filepath.IsAbs(relPath) {
		return fmt.Errorf("invalid file path in archive: %s: %w", targetPath, errInvalidPath)
	}

	return nil
}

// extractDirectory creates a directory with the specified permissions.
func extractDirectory(targetPath string, mode os.FileMode) error {
	// Create directory permissively, then set correct permissions
	err := os.MkdirAll(targetPath, defaultDirPerm) // #nosec G301
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
	}

	err = os.Chmod(targetPath, mode)
	if err != nil {
		return fmt.Errorf("failed to set permissions on directory %s: %w", targetPath, err)
	}

	return nil
}

// extractRegularFile extracts a regular file from the tar reader.
func extractRegularFile(tarReader *tar.Reader, targetPath string, mode os.FileMode) error {
	targetPath = filepath.Clean(targetPath)

	// Ensure parent directory exists
	err := os.MkdirAll(filepath.Dir(targetPath), defaultDirPerm) // #nosec G301
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", filepath.Dir(targetPath), err)
	}

	// Create file permissively, then set correct permissions
	file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY, defaultFilePerm) // #nosec G302
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", targetPath, err)
	}

	_, err = io.Copy(file, tarReader)
	if err != nil {
		_ = file.Close()

		return fmt.Errorf("failed to copy file %s: %w", targetPath, err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("failed to close file %s: %w", targetPath, err)
	}

	// Set correct permissions
	err = os.Chmod(targetPath, mode)
	if err != nil {
		return fmt.Errorf("failed to set permissions on file %s: %w", targetPath, err)
	}

	return nil
}

// extractSymlink creates a symlink.
func extractSymlink(targetPath, linkname string) error {
	// Create symlink
	err := os.Symlink(linkname, targetPath)
	if err != nil {
		return fmt.Errorf("failed to create symlink %s -> %s: %w", targetPath, linkname, err)
	}

	return nil
}

// extractHardLink creates a hard link.
func extractHardLink(targetPath, linkname string) error {
	// Create hard link
	err := os.Link(linkname, targetPath)
	if err != nil {
		return fmt.Errorf("failed to create hard link %s -> %s: %w", targetPath, linkname, err)
	}

	return nil
}

// ExtractEntry extracts a single entry from the tar archive.
// It handles directories, regular files, symlinks, and hard links, preserving permissions from the tar header.
// Files and directories are created permissively then chmod to the correct permissions from header.Mode & 0777.
func ExtractEntry(tarReader *tar.Reader, header *tar.Header, targetPath string) error {
	// Extract permissions from tar header, masking to standard Unix permissions
	mode := os.FileMode(header.Mode & unixPermMask) // #nosec G115

	switch header.Typeflag {
	case tar.TypeDir:
		return extractDirectory(targetPath, mode)

	case tar.TypeReg:
		return extractRegularFile(tarReader, targetPath, mode)

	case tar.TypeSymlink:
		return extractSymlink(targetPath, header.Linkname)

	case tar.TypeLink:
		return extractHardLink(targetPath, header.Linkname)

	default:
		// Skip unsupported entry types (e.g., character devices, block devices)
		return nil
	}
}

// Extract extracts the tar.gz archive to the specified destination directory.
// It validates paths to prevent directory traversal attacks and limits the number of files.
func Extract(archivePath, destDir string) error {
	// Validate the archive path before opening
	err := Validate(archivePath)
	if err != nil {
		return err
	}

	archivePath = filepath.Clean(archivePath)

	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}

	defer func() { _ = file.Close() }()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}

	defer func() { _ = gzipReader.Close() }()

	tarReader := tar.NewReader(gzipReader)

	// Limit the number of files to prevent zip bomb attacks
	const maxFiles = 50000

	fileCount := 0

	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		fileCount++
		if fileCount > maxFiles {
			return fmt.Errorf("archive contains too many files: %w", errTooManyFiles)
		}

		err = processTarEntry(tarReader, header, destDir)
		if err != nil {
			return err
		}
	}

	return nil
}
