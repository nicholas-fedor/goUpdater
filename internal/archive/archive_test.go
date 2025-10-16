// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package archive

import (
	"path/filepath"
	"testing"
)

const goVersionPrefix = "go1"

// getStandardExtractVersionTestCases returns standard test cases for TestExtractVersion.
func getStandardExtractVersionTestCases() []struct {
	name     string
	filename string
	expected string
} {
	return []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "standard linux archive",
			filename: "go1.21.0.linux-amd64.tar.gz",
			expected: goVersionPrefix,
		},
		{
			name:     "standard darwin archive",
			filename: "go1.20.0.darwin-amd64.tar.gz",
			expected: goVersionPrefix,
		},
		{
			name:     "standard windows archive",
			filename: "go1.19.0.windows-amd64.zip",
			expected: goVersionPrefix,
		},
		{
			name:     "full path linux archive",
			filename: "/path/to/go1.21.0.linux-amd64.tar.gz",
			expected: goVersionPrefix,
		},
		{
			name:     "full path windows archive",
			filename: "C:\\path\\to\\go1.21.0.windows-amd64.zip",
			expected: "C:\\path\\to\\go1.21.0.windows-amd64.zip", // filepath.Base on Windows gives full path for this format
		},
		{
			name:     "without extension",
			filename: "go1.21.0.linux-amd64",
			expected: goVersionPrefix,
		},
		{
			name:     "major version only",
			filename: "go1.21",
			expected: goVersionPrefix,
		},
		{
			name:     "no version after go",
			filename: "go",
			expected: "go",
		},
		{
			name:     "no dot after go",
			filename: "go1",
			expected: goVersionPrefix,
		},
		{
			name:     "invalid prefix",
			filename: "invalid-filename.tar.gz",
			expected: "invalid-filename",
		},
	}
}

// getEdgeExtractVersionTestCases returns edge case test cases for TestExtractVersion.
func getEdgeExtractVersionTestCases() []struct {
	name     string
	filename string
	expected string
} {
	return []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "empty string",
			filename: "",
			expected: ".",
		},
		{
			name:     "only extension",
			filename: ".tar.gz",
			expected: ".tar.gz",
		},
		{
			name:     "version with multiple dots",
			filename: "go1.21.0.linux-amd64.tar.gz",
			expected: goVersionPrefix,
		},
		{
			name:     "version with patch level",
			filename: "go1.21.5.linux-amd64.tar.gz",
			expected: goVersionPrefix,
		},
		{
			name:     "version with rc",
			filename: "go1.21rc1.linux-amd64.tar.gz",
			expected: goVersionPrefix,
		},
		{
			name:     "version with beta",
			filename: "go1.21beta1.linux-amd64.tar.gz",
			expected: goVersionPrefix,
		},
		{
			name:     "different major version",
			filename: "go2.0.0.linux-amd64.tar.gz",
			expected: "go2",
		},
		{
			name:     "just go with version",
			filename: "go1.21.0",
			expected: goVersionPrefix,
		},
		{
			name:     "filename with spaces",
			filename: "go 1.21.0 linux amd64.tar.gz",
			expected: "go 1",
		},
		{
			name:     "filename with special characters",
			filename: "go1.21.0-linux-amd64.tar.gz",
			expected: goVersionPrefix,
		},
	}
}

// getExtractVersionTestCases returns all test cases for TestExtractVersion.
func getExtractVersionTestCases() []struct {
	name     string
	filename string
	expected string
} {
	var allCases []struct {
		name     string
		filename string
		expected string
	}

	allCases = append(allCases, getStandardExtractVersionTestCases()...)
	allCases = append(allCases, getEdgeExtractVersionTestCases()...)

	return allCases
}

func TestExtractVersion(t *testing.T) {
	t.Parallel()

	tests := getExtractVersionTestCases()

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := ExtractVersion(testCase.filename)
			if result != testCase.expected {
				t.Errorf("ExtractVersion(%q) = %q, want %q", testCase.filename, result, testCase.expected)
			}
		})
	}
}

func TestExtractVersion_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("basename extraction", func(t *testing.T) {
		t.Parallel()

		fullPath := filepath.Join("some", "path", "go1.21.0.linux-amd64.tar.gz")
		result := ExtractVersion(fullPath)

		expected := goVersionPrefix
		if result != expected {
			t.Errorf("ExtractVersion with path %q = %q, want %q", fullPath, result, expected)
		}
	})

	t.Run("case sensitivity", func(t *testing.T) {
		t.Parallel()

		// Test with uppercase GO
		result := ExtractVersion("GO1.21.0.linux-amd64.tar.gz")

		expected := "GO1.21.0.linux-amd64" // Function doesn't remove .tar.gz for non-lowercase 'go' prefix
		if result != expected {
			t.Errorf("ExtractVersion with uppercase = %q, want %q", result, expected)
		}
	})

	t.Run("minimal valid version", func(t *testing.T) {
		t.Parallel()

		result := ExtractVersion("go1.0")

		expected := goVersionPrefix
		if result != expected {
			t.Errorf("ExtractVersion minimal = %q, want %q", result, expected)
		}
	})

	t.Run("very long version string", func(t *testing.T) {
		t.Parallel()

		longVersion := "go1.21.0.1.2.3.4.5.linux-amd64.tar.gz"
		result := ExtractVersion(longVersion)

		expected := goVersionPrefix
		if result != expected {
			t.Errorf("ExtractVersion long version = %q, want %q", result, expected)
		}
	})
}

func BenchmarkExtractVersion(b *testing.B) {
	testCases := []string{
		"go1.21.0.linux-amd64.tar.gz",
		"/full/path/to/go1.20.0.darwin-amd64.tar.gz",
		"go1.19.0.windows-amd64.zip",
		"invalid-filename",
		"go1.21.0",
	}

	b.ResetTimer()

	for range b.N {
		for _, tc := range testCases {
			ExtractVersion(tc)
		}
	}
}
