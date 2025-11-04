// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package validation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateVersionString(t *testing.T) { //nolint:maintidx // comprehensive test with many edge cases
	t.Parallel()

	tests := []struct {
		name        string
		version     string
		expectError bool
		errorType   error
	}{
		// Valid version strings
		{
			name:        "valid standard version",
			version:     "go1.21.0",
			expectError: false,
		},
		{
			name:        "valid version with patch",
			version:     "go1.20.7",
			expectError: false,
		},
		{
			name:        "valid version with pre-release",
			version:     "go1.21.0-rc.1",
			expectError: false,
		},
		{
			name:        "valid version with build metadata",
			version:     "go1.21.0+build.1",
			expectError: false,
		},
		{
			name:        "valid version with pre-release and build",
			version:     "go1.21.0-rc.1+build.1",
			expectError: false,
		},
		{
			name:        "valid major version only",
			version:     "go1.0.0",
			expectError: false,
		},
		{
			name:        "valid version with higher major",
			version:     "go2.0.0",
			expectError: false,
		},
		{
			name:        "valid version with long patch",
			version:     "go1.21.123",
			expectError: false,
		},
		{
			name:        "valid version with zero patch",
			version:     "go1.21.0",
			expectError: false,
		},
		{
			name:        "valid version with pre-release alpha",
			version:     "go1.22.0-alpha.1",
			expectError: false,
		},
		{
			name:        "valid version with pre-release beta",
			version:     "go1.22.0-beta.2",
			expectError: false,
		},
		{
			name:        "valid version with complex pre-release",
			version:     "go1.22.0-alpha.1.beta.2.rc.3",
			expectError: false,
		},
		{
			name:        "valid version with build metadata only",
			version:     "go1.21.0+20230101",
			expectError: false,
		},
		{
			name:        "valid version with complex build metadata",
			version:     "go1.21.0+build.1.sha.abcdef123456",
			expectError: false,
		},
		{
			name:        "valid version with numeric pre-release",
			version:     "go1.21.0-1",
			expectError: false,
		},
		{
			name:        "valid version with mixed pre-release",
			version:     "go1.21.0-alpha.1.2.3",
			expectError: false,
		},

		// Invalid: empty or missing required parts
		{
			name:        "empty string",
			version:     "",
			expectError: true,
			errorType:   ErrVersionEmpty,
		},
		{
			name:        "missing go prefix",
			version:     "1.21.0",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "only go prefix",
			version:     "go",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "go with empty semver",
			version:     "go",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "go with space",
			version:     "go ",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "go with only dot",
			version:     "go.",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "go with incomplete version",
			version:     "go1.",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "go with incomplete version 2",
			version:     "go1.21.",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},

		// Invalid: length constraints
		{
			name:        "version too long",
			version:     "go" + strings.Repeat("1", MaxVersionLength-1), // 257 chars total
			expectError: true,
			errorType:   ErrVersionTooLong,
		},
		{
			name:        "version exactly at max length",
			version:     "go" + strings.Repeat("1", MaxVersionLength-2), // 256 chars total
			expectError: true,                                           // Length ok, semver invalid
		},

		// Invalid: UTF-8 issues
		{
			name:        "invalid UTF-8 sequence",
			version:     "go1.21.0\xFF",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "incomplete UTF-8 sequence",
			version:     "go1.21.0\xC2", // Incomplete 2-byte sequence
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "null byte in version",
			version:     "go1.21.0\x00",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "null byte at start",
			version:     "\x00go1.21.0",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},

		// Invalid: newline characters
		{
			name:        "newline character LF",
			version:     "go1.21.0\n",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "newline character CR",
			version:     "go1.21.0\r",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "newline character CRLF",
			version:     "go1.21.0\r\n",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "newline in middle",
			version:     "go1.21\n.0",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},

		// Invalid: semantic versioning violations
		{
			name:        "invalid semver - leading zero",
			version:     "go01.21.0",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "invalid semver - negative major",
			version:     "go-1.21.0",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "invalid semver - negative minor",
			version:     "go1.-21.0",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "invalid semver - negative patch",
			version:     "go1.21.-0",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "invalid semver - empty major",
			version:     "go.21.0",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "invalid semver - empty minor",
			version:     "go1..0",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "invalid semver - empty patch",
			version:     "go1.21.",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "invalid semver - too many dots",
			version:     "go1.21.0.1",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "invalid semver - letters in numeric parts",
			version:     "go1a.21.0",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "invalid semver - spaces in version",
			version:     "go1.21 .0",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "invalid semver - special characters",
			version:     "go1.21.0@",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "invalid semver - pre-release without version",
			version:     "go-alpha.1",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "invalid semver - build metadata without version",
			version:     "go+build.1",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "invalid semver - pre-release with invalid chars",
			version:     "go1.21.0-alpha_1",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "invalid semver - build metadata with invalid chars",
			version:     "go1.21.0+build_1",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "invalid semver - pre-release starting with zero",
			version:     "go1.21.0-01",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "invalid semver - empty pre-release identifier",
			version:     "go1.21.0-alpha..1",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "invalid semver - empty build identifier",
			version:     "go1.21.0+build..1",
			expectError: true,
			errorType:   ErrVersionInvalid,
		},

		// Edge cases
		{
			name: "version with maximum reasonable length",
			version: "go1.21.0-alpha.1.beta.2.rc.3.delta.4.echo.5.foxtrot.6." +
				"golf.7.hotel.8.india.9.juliet.10.kilo.11.lima.12.mike.13." +
				"november.14.oscar.15.papa.16.quebec.17.romeo.18.sierra.19." +
				"tango.20.uniform.21.victor.22.whiskey.23.xray.24.yankee.25.zulu.26",
			expectError: false,
		},
		{
			name:        "version with unicode in build metadata",
			version:     "go1.21.0+build.🚀",
			expectError: true, // semver doesn't allow unicode in build metadata
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "version with unicode in pre-release",
			version:     "go1.21.0-alpha.🚀",
			expectError: true, // semver doesn't allow unicode in pre-release
			errorType:   ErrVersionInvalid,
		},
		{
			name:        "version with only numbers in pre-release",
			version:     "go1.21.0-123.456.789",
			expectError: false,
		},
		{
			name:        "version with mixed alphanumeric pre-release",
			version:     "go1.21.0-alpha1.beta2.rc3",
			expectError: false,
		},
		{
			name:        "version with single character pre-release",
			version:     "go1.21.0-a",
			expectError: false,
		},
		{
			name:        "version with single digit pre-release",
			version:     "go1.21.0-1",
			expectError: false,
		},
		{
			name:        "version with very long pre-release",
			version:     "go1.21.0-" + strings.Repeat("a", 200),
			expectError: false,
		},
		{
			name:        "version with very long build metadata",
			version:     "go1.21.0+" + strings.Repeat("a", 200),
			expectError: false,
		},
		{
			name:        "version with maximum numeric values",
			version:     "go999999999.999999999.999999999",
			expectError: false,
		},
		{
			name:        "version with minimum numeric values",
			version:     "go0.0.0",
			expectError: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateVersionString(testCase.version)

			if testCase.expectError {
				require.Error(t, err)

				if testCase.errorType != nil {
					require.ErrorIs(t, err, testCase.errorType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateVersionString_UTF8EdgeCases tests specific UTF-8 edge cases.
func TestValidateVersionString_UTF8EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version string
		valid   bool
	}{
		{
			name:    "valid ASCII",
			version: "go1.21.0",
			valid:   true,
		},
		{
			name:    "valid UTF-8 with accents",
			version: "go1.21.0-café",
			valid:   false, // semver doesn't allow unicode in pre-release
		},
		{
			name:    "invalid UTF-8 - overlong encoding",
			version: "go1.21.0\xC0\xAF", // Overlong encoding of '/'
			valid:   false,
		},
		{
			name:    "invalid UTF-8 - surrogate half",
			version: "go1.21.0\xED\xA0\x80", // U+D800 surrogate
			valid:   false,
		},
		{
			name:    "invalid UTF-8 - invalid continuation",
			version: "go1.21.0\xC2\x00", // Invalid continuation byte
			valid:   false,
		},
		{
			name:    "valid multi-byte UTF-8",
			version: "go1.21.0+café", // In build metadata
			valid:   false,           // semver doesn't allow unicode in build metadata
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateVersionString(testCase.version)
			if testCase.valid {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

// TestValidateVersionString_LengthBoundary tests length boundary conditions.
func TestValidateVersionString_LengthBoundary(t *testing.T) {
	t.Parallel()
	// Test exactly at max length (should pass length check)
	maxLengthVersion := "go" + strings.Repeat("1", MaxVersionLength-2) // 256 chars total
	err := ValidateVersionString(maxLengthVersion)
	// Length check passes, but semver validation may or may not pass depending on the string
	// We just ensure it's not a length error
	if err != nil {
		require.NotErrorIs(t, err, ErrVersionTooLong)
	}

	// Test over max length
	overLengthVersion := "go" + strings.Repeat("1", MaxVersionLength-1) // 257 chars total
	err = ValidateVersionString(overLengthVersion)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrVersionTooLong)
}

// TestValidateVersionString_Security tests security-related validations.
func TestValidateVersionString_Security(t *testing.T) {
	t.Parallel()

	dangerousVersions := []string{
		"go1.21.0\x00",         // Null byte
		"go1.21.0\n",           // Newline
		"go1.21.0\r",           // Carriage return
		"go1.21.0\r\n",         // CRLF
		"go1.21.0\x00\x00",     // Multiple null bytes
		"go1.21.0\n\r",         // Mixed newlines
		"\x00go1.21.0",         // Null byte at start
		"go1.21.0\x00extra",    // Null byte in middle
		"go1.21.0\xC0\xAF",     // Invalid UTF-8 that could be dangerous
		"go1.21.0\xED\xA0\x80", // Surrogate half
	}

	for _, version := range dangerousVersions {
		t.Run("dangerous_"+version, func(t *testing.T) {
			t.Parallel()

			err := ValidateVersionString(version)
			assert.Error(t, err, "Version %q should be rejected for security reasons", version)
		})
	}
}

// TestValidateVersionString_GoSpecific tests Go-specific version format requirements.
func TestValidateVersionString_GoSpecific(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version string
		valid   bool
	}{
		{
			name:    "standard go version",
			version: "go1.21.0",
			valid:   true,
		},
		{
			name:    "go version with pre-release",
			version: "go1.21.0-rc.1",
			valid:   true,
		},
		{
			name:    "missing go prefix",
			version: "1.21.0",
			valid:   false,
		},
		{
			name:    "uppercase GO prefix",
			version: "GO1.21.0",
			valid:   false,
		},
		{
			name:    "mixed case Go prefix",
			version: "Go1.21.0",
			valid:   false,
		},
		{
			name:    "go prefix with space",
			version: "go 1.21.0",
			valid:   false,
		},
		{
			name:    "go prefix with underscore",
			version: "go_1.21.0",
			valid:   false,
		},
		{
			name:    "go prefix with dash",
			version: "go-1.21.0",
			valid:   false,
		},
		{
			name:    "go prefix with dot",
			version: "go.1.21.0",
			valid:   false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateVersionString(testCase.version)
			if testCase.valid {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
