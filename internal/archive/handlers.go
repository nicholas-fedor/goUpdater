// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package archive

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

const (
	defaultDirPerm  = 0755 // Default directory permissions
	defaultFilePerm = 0644 // Default file permissions
	unixPermMask    = 0777 // Unix permission mask for tar headers
	goBinaryPerm    = 0755 // Permissions for extracted go binary
)

// processTarEntry processes a single tar entry, validating and extracting it to the destination directory.
// It performs multiple layers of path validation including symlink resolution to prevent path traversal attacks.
func (e *Extractor) processTarEntry(tarReader TarReader, header *TarHeader, destDir string) error {
	logger.Debugf("Processing tar entry: %s", header.Name)
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
		return fmt.Errorf("invalid file path in archive: %s: %w", targetPath, ErrInvalidPath)
	}

	// Additional validation to prevent path traversal
	rel, err := filepath.Rel(cleanDestDir, targetPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("invalid file path in archive: %s: %w", targetPath, ErrInvalidPath)
	}

	// Ensure the target path is safe by checking it doesn't escape the destination directory
	if !strings.HasPrefix(targetPath, cleanDestDir) {
		return fmt.Errorf("invalid file path in archive: %s: %w", targetPath, ErrInvalidPath)
	}

	// Final safety check: ensure the path is validated before use
	err = ValidatePath(targetPath, cleanDestDir)
	if err != nil {
		return err
	}

	// Validate resolved path to prevent symlink-based path traversal attacks
	// This accounts for previously extracted symlinks that could redirect the extraction path
	err = e.validateResolvedPath(targetPath, cleanDestDir)
	if err != nil {
		return err
	}

	// gosec G305 is triggered by filepath.Join, but we have validated the path thoroughly above
	// The path is safe because:
	// 1. header.Name is validated to not contain .. or be absolute
	// 2. targetPath is checked to be within cleanDestDir
	// 3. ValidatePath ensures no traversal
	// 4. Symlinks in the path are resolved and validated to stay within cleanDestDir
	return e.extractEntry(tarReader, header, targetPath, cleanDestDir, cleanDestDir)
}

// extractDirectory creates a directory with the specified permissions.
func (e *Extractor) extractDirectory(targetPath string, mode os.FileMode) error {
	// Create directory permissively, then set correct permissions
	err := e.fs.MkdirAll(targetPath, defaultDirPerm) // #nosec G301
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
	}

	err = e.fs.Chmod(targetPath, mode)
	if err != nil {
		return fmt.Errorf("failed to set permissions on directory %s: %w", targetPath, err)
	}

	return nil
}

// extractRegularFile extracts a regular file from the tar reader using buffered I/O for better performance.
func (e *Extractor) extractRegularFile(tarReader TarReader, targetPath string, mode os.FileMode) error {
	targetPath = filepath.Clean(targetPath)

	// Ensure parent directory exists
	err := e.fs.MkdirAll(filepath.Dir(targetPath), defaultDirPerm) // #nosec G301
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", filepath.Dir(targetPath), err)
	}

	// Create file permissively, then set correct permissions
	file, err := e.fs.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY, defaultFilePerm) // #nosec G302
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", targetPath, err)
	}

	// Use buffered copy with 64KB buffer for better I/O performance
	buffer := make([]byte, 64*1024) //nolint:mnd // 64KB buffer size is a reasonable constant for I/O operations

	_, err = io.CopyBuffer(file, tarReader, buffer)
	if err != nil {
		_ = file.Close()

		return fmt.Errorf("failed to copy file %s: %w", targetPath, err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("failed to close file %s: %w", targetPath, err)
	}

	// Set correct permissions
	err = e.fs.Chmod(targetPath, mode)
	if err != nil {
		return fmt.Errorf("failed to set permissions on file %s: %w", targetPath, err)
	}

	return nil
}

// extractSymlink creates a symlink after validating the linkname.
func (e *Extractor) extractSymlink(targetPath, linkname, baseDir, destDir string) error {
	// Validate the linkname to prevent symlink attacks
	err := e.validateLinkname(linkname, baseDir, destDir)
	if err != nil {
		return err
	}

	// Create symlink
	err = e.fs.Symlink(linkname, targetPath)
	if err != nil {
		return fmt.Errorf("failed to create symlink %s -> %s: %w", targetPath, linkname, err)
	}

	return nil
}

// extractHardLink creates a hard link after validating the linkname.
func (e *Extractor) extractHardLink(targetPath, linkname, baseDir, destDir string) error {
	// Validate the linkname to prevent hard link attacks
	err := e.validateLinkname(linkname, baseDir, destDir)
	if err != nil {
		return err
	}

	// Create hard link
	err = e.fs.Link(linkname, targetPath)
	if err != nil {
		return fmt.Errorf("failed to create hard link %s -> %s: %w", targetPath, linkname, err)
	}

	return nil
}

// extractEntry extracts a single entry from the tar archive.
// It handles directories, regular files, symlinks, and hard links, preserving permissions from the tar header.
// Files and directories are created permissively then chmod to the correct permissions from header.Mode & 0777.
func (e *Extractor) extractEntry(
	tarReader TarReader,
	header *TarHeader,
	targetPath, baseDir, destDir string,
) error {
	// Extract permissions from tar header, masking to standard Unix permissions
	mode := os.FileMode(header.Mode & unixPermMask) // #nosec G115

	switch header.Typeflag {
	case tar.TypeDir:
		return e.extractDirectory(targetPath, mode)

	case tar.TypeReg:
		return e.extractRegularFile(tarReader, targetPath, mode)

	case tar.TypeSymlink:
		return e.extractSymlink(targetPath, header.Linkname, baseDir, destDir)

	case tar.TypeLink:
		return e.extractHardLink(targetPath, header.Linkname, baseDir, destDir)

	default:
		// Skip unsupported entry types (e.g., character devices, block devices)
		return nil
	}
}
