// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package archive

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExtract tests the Extract function.
// Note: This test focuses on the wrapper function behavior and validation.
// Full extraction testing is done in TestExtractor_Extract.
func TestExtract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		archivePath string
		destDir     string
		expectError bool
	}{
		{
			name:        "validation fails",
			archivePath: "/path/to/invalid.tar.gz",
			destDir:     "/dest/dir",
			expectError: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			t.Logf("TestExtract: starting test case %s", testCase.name)

			// The Extract function uses real filesystem, so we can't mock it directly.
			// We test the validation failure case which should fail before any filesystem operations.
			err := Extract(testCase.archivePath, testCase.destDir)

			if testCase.expectError {
				require.Error(t, err)

				var targetErr interface{}
				assert.ErrorAs(t, err, &targetErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
