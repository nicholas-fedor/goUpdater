// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package filesystem

import (
	"errors"
	"fmt"
	"os"
)

// ErrStatFile indicates failure to stat a file.
var ErrStatFile = errors.New("failed to stat file")

// ErrOpenFile indicates failure to open a file.
var ErrOpenFile = errors.New("failed to open file")

// ErrCreateFile indicates failure to create a file.
var ErrCreateFile = errors.New("failed to create file")

// ErrRemoveAll indicates failure to remove all.
var ErrRemoveAll = errors.New("failed to remove all")

// ErrCreateDir indicates failure to create a directory.
var ErrCreateDir = errors.New("failed to create directory")

// ErrChangeMode indicates failure to change mode.
var ErrChangeMode = errors.New("failed to change mode")

// ErrGetHomeDir indicates failure to get home directory.
var ErrGetHomeDir = errors.New("failed to get home directory")

// ErrCreateTempDir indicates failure to create temporary directory.
var ErrCreateTempDir = errors.New("failed to create temporary directory")

// ErrLstatFile indicates failure to lstat a file.
var ErrLstatFile = errors.New("failed to lstat file")

// ErrEvalSymlinks indicates failure to evaluate symlinks.
var ErrEvalSymlinks = errors.New("failed to evaluate symlinks")

// ErrCreateSymlink indicates failure to create a symlink.
var ErrCreateSymlink = errors.New("failed to create symlink")

// ErrCreateHardLink indicates failure to create a hard link.
var ErrCreateHardLink = errors.New("failed to create hard link")

// ErrParseTime indicates failure to parse time.
var ErrParseTime = errors.New("failed to parse time")

// ErrWriteOutput indicates failure to write output.
var ErrWriteOutput = errors.New("failed to write output")

// ErrNoSpaceLeft indicates no space left on device.
var ErrNoSpaceLeft = errors.New("no space left on device")

// ErrHomeNotSet indicates HOME environment variable not set.
var ErrHomeNotSet = errors.New("HOME not set")

// ErrTooManyLinks indicates too many symbolic links.
var ErrTooManyLinks = errors.New("too many links")

// ErrCrossDeviceLink indicates cross-device link error.
var ErrCrossDeviceLink = errors.New("cross-device link")

// ErrSomeOther indicates some other error.
var ErrSomeOther = errors.New("some other error")

// ErrDiskFull indicates disk is full.
var ErrDiskFull = errors.New("disk full")

// ErrWrite indicates write error.
var ErrWrite = errors.New("write error")

// ErrSome indicates some error.
var ErrSome = errors.New("some error")

// ErrQuotaExceeded indicates disk quota exceeded.
var ErrQuotaExceeded = errors.New("disk quota exceeded")

// FileOperationError represents file operation failures with contextual information.
type FileOperationError struct {
	Path        string
	Operation   string
	Permissions os.FileMode
	Extra       string
	Err         error
}

// Error implements the error interface for FileOperationError.
//
//nolint:cyclop // cyclomatic complexity is acceptable for error formatting
func (e *FileOperationError) Error() string {
	switch e.Operation {
	case "stat":
		return fmt.Sprintf("failed to stat file %q: %v", e.Path, e.Err)
	case "open":
		return fmt.Sprintf("failed to open file %q: %v", e.Path, e.Err)
	case "create":
		return fmt.Sprintf("failed to create file %q: %v", e.Path, e.Err)
	case "removeAll":
		return fmt.Sprintf("failed to remove all at path %q: %v", e.Path, e.Err)
	case "mkdirAll":
		return fmt.Sprintf("failed to create directory %q: %v", e.Path, e.Err)
	case "chmod":
		return fmt.Sprintf("failed to change mode of %q: %v", e.Path, e.Err)
	case "userHomeDir":
		return fmt.Sprintf("failed to get user home directory: %v", e.Err)
	case "mkdirTemp":
		return fmt.Sprintf("failed to create temporary directory in %q: %v", e.Path, e.Err)
	case "lstat":
		return fmt.Sprintf("failed to lstat file %q: %v", e.Path, e.Err)
	case "evalSymlinks":
		return fmt.Sprintf("failed to evaluate symlinks for %q: %v", e.Path, e.Err)
	case "symlink":
		return fmt.Sprintf("failed to create symlink from %q to %q: %v", e.Path, e.Extra, e.Err)
	case "link":
		return fmt.Sprintf("failed to create hard link from %q to %q: %v", e.Path, e.Extra, e.Err)
	case "openFile":
		return fmt.Sprintf("failed to open file %q: %v", e.Path, e.Err)
	case "parse":
		return fmt.Sprintf("failed to parse time %q with layout %q: %v", e.Path, e.Extra, e.Err)
	case "fprintf":
		return fmt.Sprintf("failed to write formatted output: %v", e.Err)
	default:
		return fmt.Sprintf("file operation failed: %v", e.Err)
	}
}

// Unwrap returns the underlying error for compatibility with errors.Is and errors.As.
func (e *FileOperationError) Unwrap() error {
	return e.Err
}
