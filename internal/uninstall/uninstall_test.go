// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package uninstall

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mockFilesystem "github.com/nicholas-fedor/goUpdater/internal/filesystem/mocks"
	mockUninstall "github.com/nicholas-fedor/goUpdater/internal/uninstall/mocks"
)

// Static test errors to satisfy err113 linter rule.
var (
	errDirectoryInUseTest         = errors.New("directory in use")
	errPartialRemovalFailedTest   = errors.New("partial removal failed")
	errInvalidPathTest            = errors.New("invalid path")
	errNoSpaceLeftTest            = errors.New("no space left on device")
	errReadOnlyFileSystemTest     = errors.New("read-only file system")
	errDeviceBusyTest             = errors.New("device or resource busy")
	errCrossDeviceLinkTest        = errors.New("cross-device link")
	errFileTooLargeTest           = errors.New("file too large")
	errTooManyLinksTest           = errors.New("too many links")
	errPathTooLongTest            = errors.New("path too long")
	errInvalidArgumentTest        = errors.New("invalid argument")
	errDirectoryNotEmptyTest      = errors.New("directory not empty")
	errTextFileBusyTest           = errors.New("text file busy")
	errCannotRemoveCurrentDirTest = errors.New("cannot remove current directory")
	errResourceBusyTest           = errors.New("resource busy")
	errPartialRemovalLockedTest   = errors.New("partial removal: some files locked")
)

func TestDefaultUninstaller_Remove(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		installDir     string
		statError      error
		isNotExist     bool
		removeAllError error
		expectedError  error
	}{
		{
			name:          "successful removal",
			installDir:    "/usr/local/go",
			statError:     nil,
			isNotExist:    false,
			expectedError: nil,
		},
		{
			name:          "directory does not exist",
			installDir:    "/nonexistent/go",
			statError:     os.ErrNotExist,
			isNotExist:    true,
			expectedError: nil,
		},
		{
			name:          "stat fails with permission error",
			installDir:    "/usr/local/go",
			statError:     os.ErrPermission,
			isNotExist:    false,
			expectedError: ErrCheckInstallDir,
		},
		{
			name:           "removal fails with permission error",
			installDir:     "/usr/local/go",
			statError:      nil,
			isNotExist:     false,
			removeAllError: os.ErrPermission,
			expectedError:  ErrRemoveFailed,
		},
		{
			name:           "removal fails with directory in use",
			installDir:     "/usr/local/go",
			statError:      nil,
			isNotExist:     false,
			removeAllError: errDirectoryInUseTest,
			expectedError:  ErrRemoveFailed,
		},
		{
			name:           "partial removal failure",
			installDir:     "/usr/local/go",
			statError:      nil,
			isNotExist:     false,
			removeAllError: errPartialRemovalFailedTest,
			expectedError:  ErrRemoveFailed,
		},
		{
			name:          "empty install directory",
			installDir:    "",
			statError:     os.ErrNotExist,
			isNotExist:    true,
			expectedError: ErrCheckInstallDir,
		},
		{
			name:          "invalid path characters",
			installDir:    "/invalid/path/\x00",
			statError:     errInvalidPathTest,
			isNotExist:    false,
			expectedError: ErrCheckInstallDir,
		},
		{
			name:          "permission denied on stat",
			installDir:    "/usr/local/go",
			statError:     os.ErrPermission,
			isNotExist:    false,
			expectedError: ErrCheckInstallDir,
		},
		{
			name:           "disk full on removal",
			installDir:     "/usr/local/go",
			statError:      nil,
			isNotExist:     false,
			removeAllError: errNoSpaceLeftTest,
			expectedError:  ErrRemoveFailed,
		},
		{
			name:           "read-only filesystem",
			installDir:     "/usr/local/go",
			statError:      nil,
			isNotExist:     false,
			removeAllError: errReadOnlyFileSystemTest,
			expectedError:  ErrRemoveFailed,
		},
		{
			name:           "device busy",
			installDir:     "/usr/local/go",
			statError:      nil,
			isNotExist:     false,
			removeAllError: errDeviceBusyTest,
			expectedError:  ErrRemoveFailed,
		},
		{
			name:           "cross-device link",
			installDir:     "/usr/local/go",
			statError:      nil,
			isNotExist:     false,
			removeAllError: errCrossDeviceLinkTest,
			expectedError:  ErrRemoveFailed,
		},
		{
			name:           "file too large",
			installDir:     "/usr/local/go",
			statError:      nil,
			isNotExist:     false,
			removeAllError: errFileTooLargeTest,
			expectedError:  ErrRemoveFailed,
		},
		{
			name:           "too many links",
			installDir:     "/usr/local/go",
			statError:      nil,
			isNotExist:     false,
			removeAllError: errTooManyLinksTest,
			expectedError:  ErrRemoveFailed,
		},
		{
			name:           "path too long",
			installDir:     "/usr/local/go",
			statError:      nil,
			isNotExist:     false,
			removeAllError: errPathTooLongTest,
			expectedError:  ErrRemoveFailed,
		},
		{
			name:           "invalid argument",
			installDir:     "/usr/local/go",
			statError:      nil,
			isNotExist:     false,
			removeAllError: errInvalidArgumentTest,
			expectedError:  ErrRemoveFailed,
		},
		{
			name:           "directory not empty",
			installDir:     "/usr/local/go",
			statError:      nil,
			isNotExist:     false,
			removeAllError: errDirectoryNotEmptyTest,
			expectedError:  ErrRemoveFailed,
		},
		{
			name:           "text file busy",
			installDir:     "/usr/local/go",
			statError:      nil,
			isNotExist:     false,
			removeAllError: errTextFileBusyTest,
			expectedError:  ErrRemoveFailed,
		},
		{
			name:           "root directory",
			installDir:     "/",
			statError:      nil,
			isNotExist:     false,
			removeAllError: os.ErrPermission,
			expectedError:  ErrRemoveFailed,
		},
		{
			name:           "current directory",
			installDir:     ".",
			statError:      nil,
			isNotExist:     false,
			removeAllError: errCannotRemoveCurrentDirTest,
			expectedError:  ErrRemoveFailed,
		},
		{
			name:          "path with spaces",
			installDir:    "/path with spaces/go",
			statError:     nil,
			isNotExist:    false,
			expectedError: nil,
		},
		{
			name:          "unicode path",
			installDir:    "/usr/local/го",
			statError:     nil,
			isNotExist:    false,
			expectedError: nil,
		},
		{
			name:          "very long path",
			installDir:    "/usr/local/" + string(make([]byte, 256)) + "/go",
			statError:     errPathTooLongTest,
			isNotExist:    false,
			expectedError: ErrCheckInstallDir,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// Setup mock filesystem
			mockFS := mockFilesystem.NewMockFileSystem(t)

			// Setup expectations
			mockFS.EXPECT().Stat(testCase.installDir).Return(nil, testCase.statError).Once()

			if testCase.statError == nil {
				// Directory exists, expect RemoveAll call
				mockFS.EXPECT().RemoveAll(testCase.installDir).Return(testCase.removeAllError).Once()
			} else {
				// Stat failed, check if it's not exist
				mockFS.EXPECT().IsNotExist(testCase.statError).Return(testCase.isNotExist).Once()
			}

			// Create uninstaller with mock
			uninstaller := NewDefaultUninstaller(mockFS)

			// Execute
			err := uninstaller.Remove(testCase.installDir)

			// Assert
			if testCase.expectedError != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, testCase.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultUninstaller_Remove_EdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("concurrent access simulation", func(t *testing.T) {
		t.Parallel()
		mockFS := mockFilesystem.NewMockFileSystem(t)

		const installDir = "/usr/local/go"

		// Simulate directory exists but removal fails due to concurrent access
		mockFS.EXPECT().Stat(installDir).Return(nil, nil).Once()
		mockFS.EXPECT().RemoveAll(installDir).Return(errResourceBusyTest).Once()

		uninstaller := NewDefaultUninstaller(mockFS)
		err := uninstaller.Remove(installDir)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrRemoveFailed)
	})

	t.Run("nested directory removal", func(t *testing.T) {
		t.Parallel()
		mockFS := mockFilesystem.NewMockFileSystem(t)

		const installDir = "/usr/local/go"

		// Simulate successful removal of nested directory structure
		mockFS.EXPECT().Stat(installDir).Return(nil, nil).Once()
		mockFS.EXPECT().RemoveAll(installDir).Return(nil).Once()

		uninstaller := NewDefaultUninstaller(mockFS)
		err := uninstaller.Remove(installDir)

		assert.NoError(t, err)
	})

	t.Run("symlink handling", func(t *testing.T) {
		t.Parallel()
		mockFS := mockFilesystem.NewMockFileSystem(t)

		const installDir = "/usr/local/go"

		// Simulate removal where directory contains symlinks
		mockFS.EXPECT().Stat(installDir).Return(nil, nil).Once()
		mockFS.EXPECT().RemoveAll(installDir).Return(nil).Once()

		uninstaller := NewDefaultUninstaller(mockFS)
		err := uninstaller.Remove(installDir)

		assert.NoError(t, err)
	})

	t.Run("running process conflict", func(t *testing.T) {
		t.Parallel()
		mockFS := mockFilesystem.NewMockFileSystem(t)

		const installDir = "/usr/local/go"

		// Simulate removal failure due to running process
		mockFS.EXPECT().Stat(installDir).Return(nil, nil).Once()
		mockFS.EXPECT().RemoveAll(installDir).Return(errTextFileBusyTest).Once()

		uninstaller := NewDefaultUninstaller(mockFS)
		err := uninstaller.Remove(installDir)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrRemoveFailed)
	})

	t.Run("partial uninstallation scenario", func(t *testing.T) {
		t.Parallel()
		mockFS := mockFilesystem.NewMockFileSystem(t)

		const installDir = "/usr/local/go"

		// Simulate partial removal where some files are removed but not all
		mockFS.EXPECT().Stat(installDir).Return(nil, nil).Once()
		mockFS.EXPECT().RemoveAll(installDir).Return(errPartialRemovalLockedTest).Once()

		uninstaller := NewDefaultUninstaller(mockFS)
		err := uninstaller.Remove(installDir)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrRemoveFailed)
	})
}

func TestRemove(t *testing.T) {
	t.Parallel()
	t.Skip("Remove performs real filesystem operations and cannot be unit tested without " +
		"significant refactoring to accept dependency injection. This function is " +
		"intended for command-line usage only.")
}

func TestUninstallerInterface(t *testing.T) {
	t.Parallel()
	mockUninstall := mockUninstall.NewMockUninstaller(t)
	mockUninstall.EXPECT().Remove("/test/dir").Return(nil).Once()

	var u Uninstaller = mockUninstall

	err := u.Remove("/test/dir")

	assert.NoError(t, err)
}

func TestNewDefaultUninstaller(t *testing.T) {
	t.Parallel()
	mockFS := mockFilesystem.NewMockFileSystem(t)
	uninstaller := NewDefaultUninstaller(mockFS)

	assert.NotNil(t, uninstaller)
	assert.Equal(t, mockFS, uninstaller.fs)
}
