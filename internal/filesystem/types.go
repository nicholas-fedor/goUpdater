// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package filesystem provides filesystem interface for dependency injection.
package filesystem

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"time"
)

// FileSystem abstracts filesystem operations for testing.
//
//nolint:interfacebloat // FileSystem interface requires all os package methods for full abstraction
type FileSystem interface {
	Stat(name string) (os.FileInfo, error)
	Open(name string) (io.ReadWriteCloser, error)
	Create(name string) (io.ReadWriteCloser, error)
	RemoveAll(path string) error
	MkdirAll(path string, perm os.FileMode) error
	Chmod(name string, mode os.FileMode) error
	UserHomeDir() (string, error)
	TempDir() string
	MkdirTemp(dir, pattern string) (string, error)
	Lstat(name string) (os.FileInfo, error)
	EvalSymlinks(path string) (string, error)
	Symlink(oldname, newname string) error
	Link(oldname, newname string) error
	OpenFile(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error)
	IsNotExist(err error) bool
}

// OSFileSystem implements FileSystem using the standard os package.
type OSFileSystem struct{}

// OSDebugInfoReader implements DebugInfoReader using runtime/debug.
type OSDebugInfoReader struct{}

// OSTimeParser implements TimeParser using time package.
type OSTimeParser struct{}

// OSJSONEncoder implements JSONEncoder using json.Encoder.
type OSJSONEncoder struct {
	encoder *json.Encoder
}

// OSErrorWriter implements ErrorWriter using fmt.Fprintf.
type OSErrorWriter struct{}

// DebugInfoReader provides debug build information reading.
type DebugInfoReader interface {
	ReadBuildInfo() (*debug.BuildInfo, bool)
}

// TimeParser provides time parsing functionality.
type TimeParser interface {
	Parse(layout, value string) (time.Time, error)
	Format(t time.Time, layout string) string
}

// JSONEncoder provides JSON encoding functionality.
type JSONEncoder interface {
	NewEncoder(w io.Writer) *json.Encoder
}

// ErrorWriter provides error output functionality.
type ErrorWriter interface {
	Fprintf(w io.Writer, format string, a ...interface{}) (int, error)
}

// NewOSJSONEncoder creates a new JSON encoder for the given writer.
func NewOSJSONEncoder(w io.Writer) *OSJSONEncoder {
	return &OSJSONEncoder{
		encoder: json.NewEncoder(w),
	}
}
