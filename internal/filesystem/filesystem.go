// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package filesystem

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"
)

// Stat returns a FileInfo describing the named file.
func (fs *OSFileSystem) Stat(name string) (os.FileInfo, error) {
	info, err := os.Stat(name)
	if err != nil {
		return nil, &FileOperationError{Path: name, Operation: "stat", Permissions: 0, Extra: "", Err: err}
	}

	return info, nil
}

// Open opens the named file for reading.
func (fs *OSFileSystem) Open(name string) (io.ReadWriteCloser, error) {
	// #nosec G304 -- name is validated by caller
	file, err := os.Open(name)
	if err != nil {
		return nil, &FileOperationError{Path: name, Operation: "open", Permissions: 0, Extra: "", Err: err}
	}

	return file, nil
}

// Create creates the named file with mode 0666 (before umask), truncating it if it already exists.
func (fs *OSFileSystem) Create(name string) (io.ReadWriteCloser, error) {
	// #nosec G304 -- name is validated by caller
	file, err := os.Create(name)
	if err != nil {
		return nil, &FileOperationError{Path: name, Operation: "create", Permissions: 0, Extra: "", Err: err}
	}

	return file, nil
}

// RemoveAll removes path and any children it contains.
func (fs *OSFileSystem) RemoveAll(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return &FileOperationError{Path: path, Operation: "removeAll", Permissions: 0, Extra: "", Err: err}
	}

	return nil
}

// MkdirAll creates a directory named path, along with any necessary parents, and returns nil, or else returns an error.
func (fs *OSFileSystem) MkdirAll(path string, perm os.FileMode) error {
	err := os.MkdirAll(path, perm)
	if err != nil {
		return &FileOperationError{Path: path, Operation: "mkdirAll", Permissions: perm, Extra: "", Err: err}
	}

	return nil
}

// Chmod changes the mode of the named file to mode.
func (fs *OSFileSystem) Chmod(name string, mode os.FileMode) error {
	err := os.Chmod(name, mode)
	if err != nil {
		return &FileOperationError{Path: name, Operation: "chmod", Permissions: mode, Extra: "", Err: err}
	}

	return nil
}

// UserHomeDir returns the home directory of the current user.
func (fs *OSFileSystem) UserHomeDir() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", &FileOperationError{Path: "", Operation: "userHomeDir", Permissions: 0, Extra: "", Err: err}
	}

	return dir, nil
}

// TempDir returns the default directory to use for temporary files.
func (fs *OSFileSystem) TempDir() string {
	return os.TempDir()
}

// MkdirTemp creates a new temporary directory in the directory dir and returns the pathname of the new directory.
func (fs *OSFileSystem) MkdirTemp(dir, pattern string) (string, error) {
	tempDir, err := os.MkdirTemp(dir, pattern)
	if err != nil {
		return "", &FileOperationError{Path: dir, Operation: "mkdirTemp", Permissions: 0, Extra: "", Err: err}
	}

	return tempDir, nil
}

// Lstat returns a FileInfo describing the named file without following symbolic links.
func (fs *OSFileSystem) Lstat(name string) (os.FileInfo, error) {
	info, err := os.Lstat(name)
	if err != nil {
		return nil, &FileOperationError{Path: name, Operation: "lstat", Permissions: 0, Extra: "", Err: err}
	}

	return info, nil
}

// EvalSymlinks returns the path name after the evaluation of any symbolic links.
func (fs *OSFileSystem) EvalSymlinks(path string) (string, error) {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", &FileOperationError{Path: path, Operation: "evalSymlinks", Permissions: 0, Extra: "", Err: err}
	}

	return resolved, nil
}

// Symlink creates newname as a symbolic link to oldname.
func (fs *OSFileSystem) Symlink(oldname, newname string) error {
	err := os.Symlink(oldname, newname)
	if err != nil {
		return &FileOperationError{Path: oldname, Operation: "symlink", Permissions: 0, Extra: newname, Err: err}
	}

	return nil
}

// Link creates newname as a hard link to oldname.
func (fs *OSFileSystem) Link(oldname, newname string) error {
	err := os.Link(oldname, newname)
	if err != nil {
		return &FileOperationError{Path: oldname, Operation: "link", Permissions: 0, Extra: newname, Err: err}
	}

	return nil
}

// OpenFile is the generalized open call; most users will use Open or Create instead.
func (fs *OSFileSystem) OpenFile(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
	// #nosec G304 -- name is validated by caller
	file, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, &FileOperationError{Path: name, Operation: "openFile", Permissions: perm, Extra: "", Err: err}
	}

	return file, nil
}

// IsNotExist reports whether the given error is an os.IsNotExist error.
// It uses errors.Is to properly unwrap wrapped errors.
func (fs *OSFileSystem) IsNotExist(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}

// ReadBuildInfo reads build information from the runtime.
func (r *OSDebugInfoReader) ReadBuildInfo() (*debug.BuildInfo, bool) {
	return debug.ReadBuildInfo()
}

// Parse parses a time string using the specified layout.
func (p *OSTimeParser) Parse(layout, value string) (time.Time, error) {
	parsed, err := time.Parse(layout, value)
	if err != nil {
		return time.Time{}, &FileOperationError{Path: value, Operation: "parse", Permissions: 0, Extra: layout, Err: err}
	}

	return parsed, nil
}

// Format formats a time using the specified layout.
func (p *OSTimeParser) Format(t time.Time, layout string) string {
	return t.Format(layout)
}

// NewEncoder creates a new JSON encoder for the given writer.
func (e *OSJSONEncoder) NewEncoder(w io.Writer) *json.Encoder {
	return json.NewEncoder(w)
}

// Encode encodes the given value to JSON and writes it to the underlying writer.
func (e *OSJSONEncoder) Encode(v interface{}) error {
	return fmt.Errorf("failed to encode JSON: %w", e.encoder.Encode(v))
}

// SetIndent sets the indentation for the encoder.
func (e *OSJSONEncoder) SetIndent(prefix, indent string) {
	e.encoder.SetIndent(prefix, indent)
}

// SetEscapeHTML sets whether HTML characters should be escaped.
func (e *OSJSONEncoder) SetEscapeHTML(on bool) {
	e.encoder.SetEscapeHTML(on)
}

// Fprintf writes formatted output to the writer.
func (w *OSErrorWriter) Fprintf(writer io.Writer, format string, a ...any) (int, error) {
	n, err := fmt.Fprintf(writer, format, a...)
	if err != nil {
		return n, &FileOperationError{Path: "", Operation: "fprintf", Permissions: 0, Extra: "", Err: err}
	}

	return n, nil
}
