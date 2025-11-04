// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package archive

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

const (
	defaultDirPerm  = 0755 // Default directory permissions
	defaultFilePerm = 0644 // Default file permissions
	unixPermMask    = 0777 // Unix permission mask for tar headers
)

// NewExtractor creates a new Extractor with the given dependencies.
func NewExtractor(fs filesystem.FileSystem, processor Processor) *Extractor {
	return &Extractor{
		fs:           fs,
		processor:    processor,
		maxFiles:     20000,             //nolint:mnd // 20000 is the standard limit for Go archives
		maxTotalSize: 500 * 1024 * 1024, //nolint:mnd // 500MB limit (Go archives are ~196MB)
		maxFileSize:  50 * 1024 * 1024,  //nolint:mnd // 50MB per file limit (largest Go file is ~20.6MB)
		bufferSize:   10 * 1024 * 1024,  //nolint:mnd // 10MB buffer size for testing large file chunks
	}
}

// Extract extracts the tar.gz archive to the specified destination directory.
// It validates paths to prevent directory traversal attacks and limits the number of files.
// The file count limit is set to 20,000 to accommodate legitimate Go archives while preventing zip bomb attacks.
//
//nolint:cyclop,funlen // complex archive extraction with security validations
func (e *Extractor) Extract(archivePath, destDir string) error {
	archivePath = filepath.Clean(archivePath)
	destDir = filepath.Clean(destDir)

	logger.Debugf("Extract archive: %s", archivePath)
	logger.Debugf("Extract destination: %s", destDir)

	err := e.Validate(archivePath, destDir)
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
	// Limit total extracted size to prevent zip bomb attacks
	// Limit individual file size to prevent zip bomb attacks

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
		if fileCount > e.maxFiles {
			return &ExtractionError{
				ArchivePath: archivePath,
				Destination: destDir,
				Context:     "validating file count",
				Err:         fmt.Errorf("archive contains too many files: %w", ErrTooManyFiles),
			}
		}

		// Check for zip bomb: extremely large files or excessive total size
		if header.Size > e.maxFileSize {
			return &ExtractionError{
				ArchivePath: archivePath,
				Destination: destDir,
				Context:     "validating file size",
				Err: fmt.Errorf("archive contains file too large: %s (%d bytes): %w",
					header.Name, header.Size, ErrFileTooLarge),
			}
		}

		totalSize += header.Size
		if totalSize > e.maxTotalSize {
			return &ExtractionError{
				ArchivePath: archivePath,
				Destination: destDir,
				Context:     "validating total size",
				Err:         fmt.Errorf("archive total size too large: %d bytes: %w", totalSize, ErrFileTooLarge),
			}
		}

		err = e.processTarEntry(tarReader, header, destDir, archivePath)
		if err != nil {
			return err
		}
	}

	logger.Debugf("Successfully extracted archive: %s", archivePath)

	return nil
}

// Validate checks if the archive file exists and is a regular file.
// It returns an error if the archive path does not exist or is not a regular file.
func (e *Extractor) Validate(archivePath, destDir string) error {
	logger.Debugf("Extractor.Validate: validating archive: %s", archivePath)
	logger.Debugf("Extractor.Validate: destination: %s", destDir)

	info, err := e.fs.Stat(archivePath)
	if err != nil {
		logger.Debugf("Extractor.Validate: Stat failed for %s: %v", archivePath, err)

		return &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "validating archive",
			Err: &ValidationError{
				FilePath: archivePath,
				Criteria: "file existence",
				Err:      err,
			},
		}
	}

	if !info.Mode().IsRegular() {
		logger.Debugf("Extractor.Validate: %s is not a regular file", archivePath)

		return &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "validating archive",
			Err: &ValidationError{
				FilePath: archivePath,
				Criteria: "regular file type",
				Err:      ErrArchiveNotRegular,
			},
		}
	}

	logger.Debugf("Extractor.Validate: validation successful for %s", archivePath)

	return nil
}

// processTarEntry processes a single tar entry, validating and extracting it to the destination directory.
// It performs multiple layers of path validation including symlink resolution to prevent path traversal attacks.
//
//nolint:funlen // complex archive extraction with security validations
func (e *Extractor) processTarEntry(tarReader TarReader, header *tar.Header, destDir, archivePath string) error {
	logger.Debugf("Processing tar entry: %s", header.Name)
	// Validate the header name
	err := validateHeaderName(header.Name)
	if err != nil {
		return &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "validating header name",
			Err:         err,
		}
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
		return &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "validating target path within destination",
			Err:         fmt.Errorf("invalid file path in archive: %s: %w", targetPath, ErrInvalidPath),
		}
	}

	// Additional validation to prevent path traversal
	rel, err := filepath.Rel(cleanDestDir, targetPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "validating relative path",
			Err:         fmt.Errorf("invalid file path in archive: %s: %w", targetPath, ErrInvalidPath),
		}
	}

	// Ensure the target path is safe by checking it doesn't escape the destination directory
	if !strings.HasPrefix(targetPath, cleanDestDir) {
		return &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "validating target path safety",
			Err:         fmt.Errorf("invalid file path in archive: %s: %w", targetPath, ErrInvalidPath),
		}
	}

	// Final safety check: ensure the path is validated before use
	err = ValidatePath(targetPath, cleanDestDir)
	if err != nil {
		return &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "final path validation",
			Err:         err,
		}
	}

	// Validate resolved path to prevent symlink-based path traversal attacks
	// This accounts for previously extracted symlinks that could redirect the extraction path
	err = e.validateResolvedPath(targetPath, cleanDestDir)
	if err != nil {
		return &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "validating resolved path",
			Err:         err,
		}
	}

	// gosec G305 is triggered by filepath.Join, but we have validated the path thoroughly above
	// The path is safe because:
	// 1. header.Name is validated to not contain .. or be absolute
	// 2. targetPath is checked to be within cleanDestDir
	// 3. ValidatePath ensures no traversal
	// 4. Symlinks in the path are resolved and validated to stay within cleanDestDir
	err = e.extractEntry(tarReader, header, targetPath, cleanDestDir, cleanDestDir, archivePath)
	if err != nil {
		return err
	}

	return nil
}

// extractDirectory creates a directory with the specified permissions.
func (e *Extractor) extractDirectory(targetPath string, mode os.FileMode) error {
	// Create directory permissively, then set correct permissions
	err := e.fs.MkdirAll(targetPath, defaultDirPerm) // #nosec G301
	if err != nil {
		return fmt.Errorf("mkdirall failed: %w", err)
	}

	err = e.fs.Chmod(targetPath, mode)
	if err != nil {
		return fmt.Errorf("chmod failed: %w", err)
	}

	return nil
}

// extractRegularFile extracts a regular file from the tar reader using buffered I/O for better performance.
func (e *Extractor) extractRegularFile(tarReader TarReader, targetPath string, mode os.FileMode) error {
	targetPath = filepath.Clean(targetPath)

	// Ensure parent directory exists
	err := e.fs.MkdirAll(filepath.Dir(targetPath), defaultDirPerm) // #nosec G301
	if err != nil {
		return fmt.Errorf("mkdirall failed: %w", err)
	}

	// Create file permissively, then set correct permissions
	file, err := e.fs.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, defaultFilePerm) // #nosec G302
	if err != nil {
		return fmt.Errorf("open file failed: %w", err)
	}

	// Use buffered copy with configurable buffer for better I/O performance
	buffer := make([]byte, e.bufferSize)

	_, err = io.CopyBuffer(file, tarReader, buffer)
	if err != nil {
		_ = file.Close()

		return fmt.Errorf("copy buffer failed: %w", err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("close failed: %w", err)
	}

	// Set correct permissions
	err = e.fs.Chmod(targetPath, mode)
	if err != nil {
		return fmt.Errorf("chmod failed: %w", err)
	}

	return nil
}

// extractSymlink creates a symlink after validating the linkname.
func (e *Extractor) extractSymlink(targetPath, linkname, baseDir, destDir, archivePath string) error {
	// Validate the linkname to prevent symlink attacks
	err := e.validateLinkname(linkname, baseDir, destDir)
	if err != nil {
		return &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "validating symlink target",
			Err:         err,
		}
	}

	// Create symlink
	err = e.fs.Symlink(linkname, targetPath)
	if err != nil {
		return fmt.Errorf("symlink failed: %w", err)
	}

	return nil
}

// extractHardLink creates a hard link after validating the linkname.
func (e *Extractor) extractHardLink(targetPath, linkname, baseDir, destDir, archivePath string) error {
	// Validate the linkname to prevent hard link attacks
	err := e.validateLinkname(linkname, baseDir, destDir)
	if err != nil {
		return &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "validating linkname for hard link",
			Err:         err,
		}
	}

	// Create hard link
	err = e.fs.Link(linkname, targetPath)
	if err != nil {
		return &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "extracting hard link",
			Err:         err,
		}
	}

	return nil
}

// extractEntry extracts a single entry from the tar archive.
// It handles directories, regular files, symlinks, and hard links, preserving permissions from the tar header.
// Files and directories are created permissively then chmod to the correct permissions from header.Mode & 0777.
func (e *Extractor) extractEntry(
	tarReader TarReader,
	header *tar.Header,
	targetPath, baseDir, destDir, archivePath string,
) error {
	// Extract permissions from tar header, masking to standard Unix permissions
	mode := os.FileMode(header.Mode & unixPermMask) // #nosec G115

	switch header.Typeflag {
	case tar.TypeDir:
		err := e.extractDirectory(targetPath, mode)
		if err != nil {
			return &ExtractionError{
				ArchivePath: archivePath,
				Destination: destDir,
				Context:     "extracting directory",
				Err:         err,
			}
		}

	case tar.TypeReg:
		err := e.extractRegularFile(tarReader, targetPath, mode)
		if err != nil {
			return &ExtractionError{
				ArchivePath: archivePath,
				Destination: destDir,
				Context:     "extracting file",
				Err:         err,
			}
		}

	case tar.TypeSymlink:
		err := e.extractSymlink(targetPath, header.Linkname, baseDir, destDir, archivePath)
		if err != nil {
			return &ExtractionError{
				ArchivePath: archivePath,
				Destination: destDir,
				Context:     "extracting symlink",
				Err:         err,
			}
		}

	case tar.TypeLink:
		err := e.extractHardLink(targetPath, header.Linkname, baseDir, destDir, archivePath)
		if err != nil {
			return &ExtractionError{
				ArchivePath: archivePath,
				Destination: destDir,
				Context:     "extracting hard link",
				Err:         err,
			}
		}

	default:
		// Skip unsupported entry types (e.g., character devices, block devices)
		return nil
	}

	return nil
}

// validateEvalSymlinks performs common symlink evaluation and validation logic.
// It calls EvalSymlinks, cleans the result, checks the destDir prefix, and returns
// appropriate SecurityError with the provided validationType for EvalSymlinks errors
// and "symlink chain destination check" for invalid path cases.
func (e *Extractor) validateEvalSymlinks(targetPath, destDir, validationType string) error {
	evaled, err := e.fs.EvalSymlinks(targetPath)
	if err != nil {
		return &SecurityError{
			AttemptedPath: targetPath,
			Validation:    validationType,
			Err:           err,
		}
	}

	evaled = filepath.Clean(evaled)
	if !strings.HasPrefix(evaled, destDir+string(filepath.Separator)) && evaled != destDir {
		return &SecurityError{
			AttemptedPath: targetPath,
			Validation:    "symlink chain destination check",
			Err:           ErrInvalidPath,
		}
	}

	return nil
}

// validateSymlinkChain validates that a symlink chain resolves within the destination directory.
func (e *Extractor) validateSymlinkChain(resolved, destDir string) error {
	return e.validateEvalSymlinks(resolved, destDir, "symlink chain resolution")
}

// validateResolvedPath resolves any symlinks in the target path and validates
// that the resolved path stays within the destination directory.
func (e *Extractor) validateResolvedPath(targetPath, destDir string) error {
	return e.validateEvalSymlinks(targetPath, destDir, "resolved path destination check")
}

// validateLinkname validates the linkname for symlinks and hard links to prevent symlink attacks.
// It ensures the linkname does not contain absolute paths, ".." sequences, backslashes, or null bytes,
// resolves the path relative to the base directory and checks that the resolved path stays within the destination directory.
// Additionally, if the resolved path exists and is a symlink, it resolves any symlink chains and verifies
// the final resolved path is within the destination directory. It also prevents links to sensitive system files.
//
//nolint:lll,cyclop,funlen
func (e *Extractor) validateLinkname(linkname, baseDir, destDir string) error {
	if filepath.IsAbs(linkname) {
		return &SecurityError{
			AttemptedPath: linkname,
			Validation:    "absolute path prevention",
			Err:           ErrInvalidPath,
		}
	}

	if strings.Contains(linkname, "..") {
		return &SecurityError{
			AttemptedPath: linkname,
			Validation:    "parent directory reference prevention",
			Err:           ErrInvalidPath,
		}
	}

	if strings.Contains(linkname, "\\") {
		return &SecurityError{
			AttemptedPath: linkname,
			Validation:    "backslash prevention",
			Err:           ErrInvalidPath,
		}
	}

	if strings.Contains(linkname, "\x00") {
		return &SecurityError{
			AttemptedPath: linkname,
			Validation:    "null byte prevention",
			Err:           ErrInvalidPath,
		}
	}

	resolved := filepath.Join(baseDir, linkname)
	resolved = filepath.Clean(resolved)

	// Check if resolved is within destDir
	if !strings.HasPrefix(resolved, destDir+string(filepath.Separator)) && resolved != destDir {
		return &SecurityError{
			AttemptedPath: linkname,
			Validation:    "linkname destination check",
			Err:           ErrInvalidPath,
		}
	}

	// Additional validation using filepath.Rel
	rel, err := filepath.Rel(destDir, resolved)
	if err != nil || strings.HasPrefix(rel, "..") {
		return &SecurityError{
			AttemptedPath: linkname,
			Validation:    "relative path validation",
			Err:           ErrInvalidPath,
		}
	}

	// Prevent links to sensitive system files
	sensitivePaths := []string{"/etc", "/usr", "/bin", "/sbin", "/dev", "/proc", "/sys", "/root", "/home"}
	for _, sensitive := range sensitivePaths {
		if strings.HasPrefix(resolved, sensitive) {
			return &SecurityError{
				AttemptedPath: linkname,
				Validation:    "sensitive system path prevention",
				Err:           ErrInvalidPath,
			}
		}
	}

	// If the resolved path exists and is a symlink, resolve the symlink chain
	info, err := e.fs.Lstat(resolved)
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
		err = e.validateSymlinkChain(resolved, destDir)
		if err != nil {
			return &SecurityError{
				AttemptedPath: linkname,
				Validation:    "symlink chain validation",
				Err:           err,
			}
		}
	}

	return nil
}
