// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package archive

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	errUnderlyingExtraction = errors.New("underlying extraction error")
	errUnderlyingSecurity   = errors.New("underlying security error")
	errUnderlyingValidation = errors.New("underlying validation error")
)

func TestExtractionError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		archivePath string
		destination string
		context     string
		expectedMsg string
	}{
		{
			name:        "basic extraction error",
			archivePath: "/path/to/archive.tar.gz",
			destination: "/dest/dir",
			context:     "failed to extract",
			expectedMsg: "extraction failed: archive=archive.tar.gz dest=dir context=failed to extract",
		},
		{
			name:        "extraction error with empty paths",
			archivePath: "",
			destination: "",
			context:     "no archive found",
			expectedMsg: "extraction failed: archive=unknown dest=unknown context=no archive found",
		},
		{
			name:        "extraction error with relative paths",
			archivePath: "archive.zip",
			destination: "dest",
			context:     "permission denied",
			expectedMsg: "extraction failed: archive=archive.zip dest=dest context=permission denied",
		},
		{
			name:        "extraction error with windows paths",
			archivePath: "C:\\path\\to\\archive.exe",
			destination: "C:\\dest\\dir",
			context:     "invalid format",
			expectedMsg: "extraction failed: archive=archive.exe dest=dir context=invalid format",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			e := &ExtractionError{
				ArchivePath: testCase.archivePath,
				Destination: testCase.destination,
				Context:     testCase.context,
			}
			assert.Equal(t, testCase.expectedMsg, e.Error())
		})
	}
}

func TestExtractionError_Unwrap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		underlying  error
		expectedErr error
	}{
		{
			name:        "unwrap with underlying error",
			underlying:  errUnderlyingExtraction,
			expectedErr: errUnderlyingExtraction,
		},
		{
			name:        "unwrap with nil error",
			underlying:  nil,
			expectedErr: nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			e := &ExtractionError{Err: testCase.underlying}
			assert.Equal(t, testCase.expectedErr, e.Unwrap())
		})
	}
}

func TestSecurityError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		attemptedPath string
		validation    string
		expectedMsg   string
	}{
		{
			name:          "basic security error",
			attemptedPath: "/path/to/../../../etc/passwd",
			validation:    "path traversal detected",
			expectedMsg:   "security error: path=passwd validation=path traversal detected",
		},
		{
			name:          "security error with empty path",
			attemptedPath: "",
			validation:    "empty path not allowed",
			expectedMsg:   "security error: path=unknown validation=empty path not allowed",
		},
		{
			name:          "security error with relative path",
			attemptedPath: "../config.yml",
			validation:    "relative path outside allowed directory",
			expectedMsg:   "security error: path=config.yml validation=relative path outside allowed directory",
		},
		{
			name:          "security error with windows path",
			attemptedPath: "C:\\Windows\\System32\\cmd.exe",
			validation:    "executable not allowed",
			expectedMsg:   "security error: path=cmd.exe validation=executable not allowed",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			e := &SecurityError{
				AttemptedPath: testCase.attemptedPath,
				Validation:    testCase.validation,
			}
			assert.Equal(t, testCase.expectedMsg, e.Error())
		})
	}
}

func TestSecurityError_Unwrap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		underlying  error
		expectedErr error
	}{
		{
			name:        "unwrap with underlying error",
			underlying:  errUnderlyingSecurity,
			expectedErr: errUnderlyingSecurity,
		},
		{
			name:        "unwrap with nil error",
			underlying:  nil,
			expectedErr: nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			e := &SecurityError{Err: testCase.underlying}
			assert.Equal(t, testCase.expectedErr, e.Unwrap())
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		filePath    string
		criteria    string
		expectedMsg string
	}{
		{
			name:        "basic validation error",
			filePath:    "/path/to/malformed.zip",
			criteria:    "invalid archive format",
			expectedMsg: "validation error: file=malformed.zip criteria=invalid archive format",
		},
		{
			name:        "validation error with empty path",
			filePath:    "",
			criteria:    "file cannot be empty",
			expectedMsg: "validation error: file=unknown criteria=file cannot be empty",
		},
		{
			name:        "validation error with relative path",
			filePath:    "archive.rar",
			criteria:    "unsupported format",
			expectedMsg: "validation error: file=archive.rar criteria=unsupported format",
		},
		{
			name:        "validation error with windows path",
			filePath:    "C:\\temp\\corrupt.tar.gz",
			criteria:    "corrupted archive",
			expectedMsg: "validation error: file=corrupt.tar.gz criteria=corrupted archive",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			e := &ValidationError{
				FilePath: testCase.filePath,
				Criteria: testCase.criteria,
			}
			assert.Equal(t, testCase.expectedMsg, e.Error())
		})
	}
}

func TestValidationError_Unwrap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		underlying  error
		expectedErr error
	}{
		{
			name:        "unwrap with underlying error",
			underlying:  errUnderlyingValidation,
			expectedErr: errUnderlyingValidation,
		},
		{
			name:        "unwrap with nil error",
			underlying:  nil,
			expectedErr: nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			e := &ValidationError{Err: testCase.underlying}
			assert.Equal(t, testCase.expectedErr, e.Unwrap())
		})
	}
}

func Test_sanitizePathForError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "empty path",
			path:     "",
			expected: "unknown",
		},
		{
			name:     "filename only",
			path:     "file.txt",
			expected: "file.txt",
		},
		{
			name:     "unix path",
			path:     "/home/user/file.txt",
			expected: "file.txt",
		},
		{
			name:     "windows path",
			path:     "C:\\Users\\user\\file.txt",
			expected: "file.txt",
		},
		{
			name:     "path with multiple slashes",
			path:     "/a/b/c/d/file.txt",
			expected: "file.txt",
		},
		{
			name:     "path ending with slash",
			path:     "/path/to/dir/",
			expected: "",
		},
		{
			name:     "path with backslashes",
			path:     "path\\to\\file.txt",
			expected: "file.txt",
		},
		{
			name:     "path with mixed slashes",
			path:     "/path\\to/file.txt",
			expected: "file.txt",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := sanitizePathForError(testCase.path)
			assert.Equal(t, testCase.expected, result)
		})
	}
}

func Test_findLastSlash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "empty string",
			input:    "",
			expected: -1,
		},
		{
			name:     "no slashes",
			input:    "filename.txt",
			expected: -1,
		},
		{
			name:     "unix path",
			input:    "/path/to/file.txt",
			expected: 8, // position of '/' before 'file.txt'
		},
		{
			name:     "windows path",
			input:    "C:\\path\\to\\file.txt",
			expected: 10, // position of '\' before 'file.txt'
		},
		{
			name:     "mixed slashes",
			input:    "/path\\to/file.txt",
			expected: 8, // position of '/' before 'file.txt'
		},
		{
			name:     "ends with slash",
			input:    "/path/to/dir/",
			expected: 12, // position of last '/'
		},
		{
			name:     "only slashes",
			input:    "///",
			expected: 2, // position of last '/'
		},
		{
			name:     "backslash only",
			input:    "path\\to\\file",
			expected: 7, // position of '\' before 'file'
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := findLastSlash(testCase.input)
			assert.Equal(t, testCase.expected, result)
		})
	}
}
