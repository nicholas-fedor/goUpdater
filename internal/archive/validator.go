// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package archive

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

// Validate checks if the archive file exists and is a regular file.
// It returns an error if the archive path does not exist or is not a regular file.
func (e *Extractor) Validate(archivePath string) error {
	logger.Debugf("Extractor.Validate: validating archive: %s", archivePath)

	info, err := e.fs.Stat(archivePath)
	if err != nil {
		logger.Debugf("Extractor.Validate: Stat failed for %s: %v", archivePath, err)

		return &ValidationError{
			FilePath: archivePath,
			Criteria: "file existence",
			Err:      err,
		}
	}

	if !info.Mode().IsRegular() {
		logger.Debugf("Extractor.Validate: %s is not a regular file", archivePath)

		return &ValidationError{
			FilePath: archivePath,
			Criteria: "regular file type",
			Err:      ErrArchiveNotRegular,
		}
	}

	logger.Debugf("Extractor.Validate: validation successful for %s", archivePath)

	return nil
}

// ValidatePath ensures the extracted path is within the installation directory.
// It returns an error if the target path attempts to traverse outside the allowed directory.
func ValidatePath(targetPath, installDir string) error {
	cleanInstallDir := filepath.Clean(installDir)

	relPath, err := filepath.Rel(cleanInstallDir, targetPath)
	if err != nil {
		return &SecurityError{
			AttemptedPath: targetPath,
			Validation:    "path relativity check",
			Err:           err,
		}
	}

	if strings.HasPrefix(relPath, "..") || filepath.IsAbs(relPath) {
		return &SecurityError{
			AttemptedPath: targetPath,
			Validation:    "directory traversal prevention",
			Err:           ErrInvalidPath,
		}
	}

	return nil
}

// validateHeaderName checks if the tar header name is safe for extraction.
// It prevents directory traversal attacks by ensuring no absolute paths, parent directory references,
// or other dangerous patterns that could lead to path traversal vulnerabilities.
func validateHeaderName(headerName string) error {
	if filepath.IsAbs(headerName) {
		return &SecurityError{
			AttemptedPath: headerName,
			Validation:    "absolute path prevention",
			Err:           ErrInvalidPath,
		}
	}

	if strings.Contains(headerName, "..") {
		return &SecurityError{
			AttemptedPath: headerName,
			Validation:    "parent directory reference prevention",
			Err:           ErrInvalidPath,
		}
	}

	// Additional checks for other dangerous patterns
	if strings.Contains(headerName, "\\") {
		return &SecurityError{
			AttemptedPath: headerName,
			Validation:    "backslash prevention",
			Err:           ErrInvalidPath,
		}
	}

	// Check for null bytes which could be used for path manipulation
	if strings.Contains(headerName, "\x00") {
		return &SecurityError{
			AttemptedPath: headerName,
			Validation:    "null byte prevention",
			Err:           ErrInvalidPath,
		}
	}

	return nil
}

// validateSymlinkChain validates that a symlink chain resolves within the destination directory.
func (e *Extractor) validateSymlinkChain(resolved, destDir string) error {
	evaled, err := e.fs.EvalSymlinks(resolved)
	if err != nil {
		return &SecurityError{
			AttemptedPath: resolved,
			Validation:    "symlink chain resolution",
			Err:           err,
		}
	}

	evaled = filepath.Clean(evaled)
	if !strings.HasPrefix(evaled, destDir+string(filepath.Separator)) && evaled != destDir {
		return &SecurityError{
			AttemptedPath: evaled,
			Validation:    "symlink chain destination check",
			Err:           ErrInvalidPath,
		}
	}

	return nil
}

// validateResolvedPath resolves any symlinks in the target path and validates
// that the resolved path stays within the destination directory.
func (e *Extractor) validateResolvedPath(targetPath, destDir string) error {
	resolved, err := e.fs.EvalSymlinks(targetPath)
	if err == nil {
		resolved = filepath.Clean(resolved)
		if !strings.HasPrefix(resolved, destDir+string(filepath.Separator)) && resolved != destDir {
			return &SecurityError{
				AttemptedPath: resolved,
				Validation:    "resolved path destination check",
				Err:           ErrInvalidPath,
			}
		}
	}

	return nil
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
