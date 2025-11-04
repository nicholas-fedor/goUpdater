// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package install

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mockFilesystem "github.com/nicholas-fedor/goUpdater/internal/filesystem/mocks"
	mockInstall "github.com/nicholas-fedor/goUpdater/internal/install/mocks"
	"github.com/nicholas-fedor/goUpdater/internal/types"
)

func TestInstall_InstallLatestVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		installDir          string
		archivePath         string
		mockSetup           func(*mockInstall.MockPrivilegeService, *mockInstall.MockVerifyService)
		expectedError       string
		expectPrivilegeCall bool
	}{
		{
			name:        "successful latest version install",
			installDir:  "/usr/local/go",
			archivePath: "",
			mockSetup: func(mockPriv *mockInstall.MockPrivilegeService, mockVer *mockInstall.MockVerifyService) {
				mockVer.EXPECT().GetInstalledVersion("/usr/local/go").Return("", nil).Once()
				mockPriv.EXPECT().ElevateAndExecute(mock.AnythingOfType("func() error")).Return(nil).Once()
			},
			expectedError:       "",
			expectPrivilegeCall: true,
		},
		{
			name:        "privilege elevation failure",
			installDir:  "/usr/local/go",
			archivePath: "",
			mockSetup: func(mockPriv *mockInstall.MockPrivilegeService, mockVer *mockInstall.MockVerifyService) {
				mockVer.EXPECT().GetInstalledVersion("/usr/local/go").Return("", nil).Once()
				mockPriv.EXPECT().ElevateAndExecute(mock.AnythingOfType("func() error")).Return(ErrPermissionDenied).Once()
			},
			expectedError:       "failed to install latest Go: permission denied",
			expectPrivilegeCall: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// Setup mocks
			mockPriv := mockInstall.NewMockPrivilegeService(t)
			mockVer := mockInstall.NewMockVerifyService(t)
			testCase.mockSetup(mockPriv, mockVer)

			// Create installer with mocks
			installer := NewInstallerWithDeps(
				mockFilesystem.NewMockFileSystem(t),
				nil, // archiveService not needed for this test
				nil, // downloadService not needed for this test
				mockVer,
				nil, // versionService not needed for this test
				mockPriv,
				nil, // reader not needed for this test
			)

			// Execute
			err := installer.Install(testCase.installDir, testCase.archivePath)

			// Assert
			if testCase.expectedError == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
			}
		})
	}
}

func TestInstall_InstallFromArchive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		installDir          string
		archivePath         string
		mockSetup           func(*mockInstall.MockPrivilegeService, *mockInstall.MockVerifyService)
		expectedError       string
		expectPrivilegeCall bool
	}{
		{
			name:        "successful archive install",
			installDir:  "/usr/local/go",
			archivePath: "/tmp/go1.21.0.linux-amd64.tar.gz",
			mockSetup: func(mockPriv *mockInstall.MockPrivilegeService, _ *mockInstall.MockVerifyService) {
				mockPriv.EXPECT().ElevateAndExecute(mock.AnythingOfType("func() error")).Return(nil).Once()
			},
			expectedError:       "",
			expectPrivilegeCall: true,
		},
		{
			name:        "privilege elevation failure for archive",
			installDir:  "/usr/local/go",
			archivePath: "/tmp/go1.21.0.linux-amd64.tar.gz",
			mockSetup: func(mockPriv *mockInstall.MockPrivilegeService, _ *mockInstall.MockVerifyService) {
				mockPriv.EXPECT().ElevateAndExecute(mock.AnythingOfType("func() error")).Return(ErrElevationFailed).Once()
			},
			expectedError:       "failed to install Go from archive: elevation failed",
			expectPrivilegeCall: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// Setup mocks
			mockPriv := mockInstall.NewMockPrivilegeService(t)
			mockVer := mockInstall.NewMockVerifyService(t)
			testCase.mockSetup(mockPriv, mockVer)

			// Create installer with mocks
			installer := NewInstallerWithDeps(
				mockFilesystem.NewMockFileSystem(t),
				nil, // archiveService not needed for this test
				nil, // downloadService not needed for this test
				mockVer,
				nil, // versionService not needed for this test
				mockPriv,
				nil, // reader not needed for this test
			)

			// Execute
			err := installer.Install(testCase.installDir, testCase.archivePath)

			// Assert
			if testCase.expectedError == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
			}
		})
	}
}

func TestDirectExtract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		archivePath   string
		installDir    string
		mockSetup     func(*mockInstall.MockArchiveService, *mockInstall.MockVerifyService, *mockFilesystem.MockFileSystem)
		expectedError string
	}{
		{
			name:        "successful direct extraction",
			archivePath: "/tmp/go1.21.0.linux-amd64.tar.gz",
			installDir:  "/usr/local/go",
			mockSetup: func(
				mockArch *mockInstall.MockArchiveService,
				mockVer *mockInstall.MockVerifyService,
				mockFS *mockFilesystem.MockFileSystem,
			) {
				mockFS.EXPECT().MkdirAll("/usr/local", os.FileMode(directoryPermissions)).Return(nil).Once()
				mockArch.EXPECT().Extract("/tmp/go1.21.0.linux-amd64.tar.gz", "/usr/local").Return(nil).Once()
				mockArch.EXPECT().ExtractVersion("/tmp/go1.21.0.linux-amd64.tar.gz").Return("1.21.0").Once()
				mockVer.EXPECT().Installation("/usr/local/go", "1.21.0").Return(nil).Once()
			},
			expectedError: "",
		},
		{
			name:        "extraction failure",
			archivePath: "/tmp/go1.21.0.linux-amd64.tar.gz",
			installDir:  "/usr/local/go",
			mockSetup: func(
				mockArch *mockInstall.MockArchiveService,
				_ *mockInstall.MockVerifyService,
				mockFS *mockFilesystem.MockFileSystem,
			) {
				mockFS.EXPECT().MkdirAll("/usr/local", os.FileMode(directoryPermissions)).Return(nil).Once()
				mockArch.EXPECT().Extract("/tmp/go1.21.0.linux-amd64.tar.gz", "/usr/local").Return(ErrExtractionFailed).Once()
			},
			expectedError: "install failed at extract phase",
		},
		{
			name:        "directory creation failure",
			archivePath: "/tmp/go1.21.0.linux-amd64.tar.gz",
			installDir:  "/usr/local/go",
			mockSetup: func(
				_ *mockInstall.MockArchiveService,
				_ *mockInstall.MockVerifyService,
				mockFS *mockFilesystem.MockFileSystem,
			) {
				mockFS.EXPECT().MkdirAll("/usr/local", os.FileMode(directoryPermissions)).Return(ErrDirectoryCreationFailed).Once()
			},
			expectedError: "install failed at prepare phase",
		},
		{
			name:        "verification failure",
			archivePath: "/tmp/go1.21.0.linux-amd64.tar.gz",
			installDir:  "/usr/local/go",
			mockSetup: func(
				mockArch *mockInstall.MockArchiveService,
				mockVer *mockInstall.MockVerifyService,
				mockFS *mockFilesystem.MockFileSystem,
			) {
				mockFS.EXPECT().MkdirAll("/usr/local", os.FileMode(directoryPermissions)).Return(nil).Once()
				mockArch.EXPECT().Extract("/tmp/go1.21.0.linux-amd64.tar.gz", "/usr/local").Return(nil).Once()
				mockArch.EXPECT().ExtractVersion("/tmp/go1.21.0.linux-amd64.tar.gz").Return("1.21.0").Once()
				mockVer.EXPECT().Installation("/usr/local/go", "1.21.0").Return(ErrVerificationFailed).Once()
			},
			expectedError: "install failed at verify phase",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// Setup mocks
			mockArch := mockInstall.NewMockArchiveService(t)
			mockVer := mockInstall.NewMockVerifyService(t)
			mockFS := mockFilesystem.NewMockFileSystem(t)
			testCase.mockSetup(mockArch, mockVer, mockFS)

			// Create installer with mocks
			installer := NewInstallerWithDeps(
				mockFS,
				mockArch,
				nil, // downloadService not needed
				mockVer,
				nil, // versionService not needed
				nil, // privilegeService not needed
				nil, // reader not needed
			)

			// Execute
			err := installer.DirectExtract(testCase.archivePath, testCase.installDir)

			// Assert
			if testCase.expectedError == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
			}
		})
	}
}

func TestExtract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		archivePath   string
		installDir    string
		checksum      string
		mockSetup     func(*mockInstall.MockArchiveService, *mockFilesystem.MockFileSystem)
		expectedError string
	}{
		{
			name:        "successful extraction with validation",
			archivePath: "/tmp/go1.21.0.linux-amd64.tar.gz",
			installDir:  "/usr/local/go",
			checksum:    "abc123",
			mockSetup: func(mockArch *mockInstall.MockArchiveService, mockFS *mockFilesystem.MockFileSystem) {
				mockArch.EXPECT().Validate("/tmp/go1.21.0.linux-amd64.tar.gz", "/usr/local").Return(nil).Once()
				mockFS.EXPECT().MkdirAll("/usr/local", os.FileMode(directoryPermissions)).Return(nil).Once()
				mockArch.EXPECT().Extract("/tmp/go1.21.0.linux-amd64.tar.gz", "/usr/local").Return(nil).Once()
				mockArch.EXPECT().ExtractVersion("/tmp/go1.21.0.linux-amd64.tar.gz").Return("1.21.0").Once()
			},
			expectedError: "",
		},
		{
			name:        "validation failure",
			archivePath: "/tmp/go1.21.0.linux-amd64.tar.gz",
			installDir:  "/usr/local/go",
			checksum:    "abc123",
			mockSetup: func(mockArch *mockInstall.MockArchiveService, _ *mockFilesystem.MockFileSystem) {
				mockArch.EXPECT().Validate("/tmp/go1.21.0.linux-amd64.tar.gz", "/usr/local").Return(ErrInvalidArchive).Once()
			},
			expectedError: "failed to validate archive",
		},
		{
			name:        "extraction failure after validation",
			archivePath: "/tmp/go1.21.0.linux-amd64.tar.gz",
			installDir:  "/usr/local/go",
			checksum:    "abc123",
			mockSetup: func(
				mockArch *mockInstall.MockArchiveService,
				mockFS *mockFilesystem.MockFileSystem,
			) {
				mockArch.EXPECT().Validate("/tmp/go1.21.0.linux-amd64.tar.gz", "/usr/local").Return(nil).Once()
				mockFS.EXPECT().MkdirAll("/usr/local", os.FileMode(directoryPermissions)).Return(nil).Once()
				mockArch.EXPECT().Extract("/tmp/go1.21.0.linux-amd64.tar.gz", "/usr/local").Return(ErrExtractionError).Once()
			},
			expectedError: "install failed at extract phase",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// Setup mocks
			mockArch := mockInstall.NewMockArchiveService(t)
			mockFS := mockFilesystem.NewMockFileSystem(t)
			testCase.mockSetup(mockArch, mockFS)

			// Create installer with mocks
			installer := NewInstallerWithDeps(
				mockFS,
				mockArch,
				nil, // downloadService not needed
				nil, // verifyService not needed
				nil, // versionService not needed
				nil, // privilegeService not needed
				nil, // reader not needed
			)

			// Execute
			err := installer.Extract(testCase.archivePath, testCase.installDir, testCase.checksum)

			// Assert
			if testCase.expectedError == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
			}
		})
	}
}

func TestExtractWithVerification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		archivePath   string
		installDir    string
		checksum      string
		mockSetup     func(*mockInstall.MockArchiveService, *mockInstall.MockVerifyService, *mockFilesystem.MockFileSystem)
		expectedError string
	}{
		{
			name:        "successful extraction with verification",
			archivePath: "/tmp/go1.21.0.linux-amd64.tar.gz",
			installDir:  "/usr/local/go",
			checksum:    "abc123",
			mockSetup: func(
				mockArch *mockInstall.MockArchiveService,
				mockVer *mockInstall.MockVerifyService,
				mockFS *mockFilesystem.MockFileSystem,
			) {
				mockArch.EXPECT().ExtractVersion("/tmp/go1.21.0.linux-amd64.tar.gz").Return("1.21.0").Once()
				mockArch.EXPECT().Validate("/tmp/go1.21.0.linux-amd64.tar.gz", "/usr/local").Return(nil).Once()
				mockFS.EXPECT().MkdirAll("/usr/local", os.FileMode(directoryPermissions)).Return(nil).Once()
				mockArch.EXPECT().Extract("/tmp/go1.21.0.linux-amd64.tar.gz", "/usr/local").Return(nil).Once()
				mockArch.EXPECT().ExtractVersion("/tmp/go1.21.0.linux-amd64.tar.gz").Return("1.21.0").Once()
				mockVer.EXPECT().Installation("/usr/local/go", "1.21.0").Return(nil).Once()
			},
			expectedError: "",
		},
		{
			name:        "verification failure",
			archivePath: "/tmp/go1.21.0.linux-amd64.tar.gz",
			installDir:  "/usr/local/go",
			checksum:    "abc123",
			mockSetup: func(
				mockArch *mockInstall.MockArchiveService,
				mockVer *mockInstall.MockVerifyService,
				mockFS *mockFilesystem.MockFileSystem,
			) {
				mockArch.EXPECT().ExtractVersion("/tmp/go1.21.0.linux-amd64.tar.gz").Return("1.21.0").Once()
				mockArch.EXPECT().Validate("/tmp/go1.21.0.linux-amd64.tar.gz", "/usr/local").Return(nil).Once()
				mockFS.EXPECT().MkdirAll("/usr/local", os.FileMode(directoryPermissions)).Return(nil).Once()
				mockArch.EXPECT().Extract("/tmp/go1.21.0.linux-amd64.tar.gz", "/usr/local").Return(nil).Once()
				mockArch.EXPECT().ExtractVersion("/tmp/go1.21.0.linux-amd64.tar.gz").Return("1.21.0").Once()
				mockVer.EXPECT().Installation("/usr/local/go", "1.21.0").Return(ErrVerificationFailed).Once()
			},
			expectedError: "install failed at verify phase",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// Setup mocks
			mockArch := mockInstall.NewMockArchiveService(t)
			mockVer := mockInstall.NewMockVerifyService(t)
			mockFS := mockFilesystem.NewMockFileSystem(t)
			testCase.mockSetup(mockArch, mockVer, mockFS)

			// Create installer with mocks
			installer := NewInstallerWithDeps(
				mockFS,
				mockArch,
				nil, // downloadService not needed
				mockVer,
				nil, // versionService not needed
				nil, // privilegeService not needed
				nil, // reader not needed
			)

			// Execute
			err := installer.ExtractWithVerification(testCase.archivePath, testCase.installDir, testCase.checksum)

			// Assert
			if testCase.expectedError == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
			}
		})
	}
}

func TestLatest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		installDir string
		mockSetup  func(
			*mockInstall.MockDownloadService,
			*mockInstall.MockArchiveService,
			*mockInstall.MockVerifyService,
			*mockFilesystem.MockFileSystem,
		)
		expectedError string
	}{
		{
			name:       "successful latest version download and install",
			installDir: "/usr/local/go",
			mockSetup: func(
				mockDown *mockInstall.MockDownloadService,
				mockArch *mockInstall.MockArchiveService,
				mockVer *mockInstall.MockVerifyService,
				mockFS *mockFilesystem.MockFileSystem,
			) {
				mockFS.EXPECT().MkdirTemp("", "goUpdater-install-*").Return("/tmp/goUpdater-install-123", nil).Once()
				mockFS.EXPECT().RemoveAll("/tmp/goUpdater-install-123").Return(nil).Once()
				mockDown.EXPECT().GetLatest("/tmp/goUpdater-install-123").Return(
					"/tmp/goUpdater-install-123/go.tar.gz",
					"checksum123",
					nil,
				).Once()
				mockArch.EXPECT().ExtractVersion("/tmp/goUpdater-install-123/go.tar.gz").Return("1.21.0").Once()
				mockArch.EXPECT().Validate("/tmp/goUpdater-install-123/go.tar.gz", "/usr/local").Return(nil).Once()
				mockFS.EXPECT().MkdirAll("/usr/local", os.FileMode(directoryPermissions)).Return(nil).Once()
				mockArch.EXPECT().Extract("/tmp/goUpdater-install-123/go.tar.gz", "/usr/local").Return(nil).Once()
				mockArch.EXPECT().ExtractVersion("/tmp/goUpdater-install-123/go.tar.gz").Return("1.21.0").Once()
				mockVer.EXPECT().Installation("/usr/local/go", "1.21.0").Return(nil).Once()
			},
			expectedError: "",
		},
		{
			name:       "temp directory creation failure",
			installDir: "/usr/local/go",
			mockSetup: func(
				_ *mockInstall.MockDownloadService,
				_ *mockInstall.MockArchiveService,
				_ *mockInstall.MockVerifyService,
				mockFS *mockFilesystem.MockFileSystem,
			) {
				mockFS.EXPECT().MkdirTemp("", "goUpdater-install-*").Return("", ErrTempDirFailed).Once()
			},
			expectedError: "install failed at prepare phase",
		},
		{
			name:       "download failure",
			installDir: "/usr/local/go",
			mockSetup: func(
				mockDown *mockInstall.MockDownloadService,
				_ *mockInstall.MockArchiveService,
				_ *mockInstall.MockVerifyService,
				mockFS *mockFilesystem.MockFileSystem,
			) {
				mockFS.EXPECT().MkdirTemp("", "goUpdater-install-*").Return("/tmp/goUpdater-install-123", nil).Once()
				mockFS.EXPECT().RemoveAll("/tmp/goUpdater-install-123").Return(nil).Once()
				mockDown.EXPECT().GetLatest("/tmp/goUpdater-install-123").Return("", "", ErrDownloadFailed).Once()
			},
			expectedError: "install failed at download phase",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// Setup mocks
			mockDown := mockInstall.NewMockDownloadService(t)
			mockArch := mockInstall.NewMockArchiveService(t)
			mockVer := mockInstall.NewMockVerifyService(t)
			mockFS := mockFilesystem.NewMockFileSystem(t)
			testCase.mockSetup(mockDown, mockArch, mockVer, mockFS)

			// Create installer with mocks
			installer := NewInstallerWithDeps(
				mockFS,
				mockArch,
				mockDown,
				mockVer,
				nil, // versionService not needed
				nil, // privilegeService not needed
				nil, // reader not needed
			)

			// Execute
			err := installer.Latest(testCase.installDir)

			// Assert
			if testCase.expectedError == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
			}
		})
	}
}

func TestPrepareInstallDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		installDir    string
		mockSetup     func(*mockFilesystem.MockFileSystem)
		expectedError string
	}{
		{
			name:       "successful directory preparation",
			installDir: "/usr/local/go",
			mockSetup: func(mockFS *mockFilesystem.MockFileSystem) {
				mockFS.EXPECT().MkdirAll("/usr/local", os.FileMode(directoryPermissions)).Return(nil).Once()
			},
			expectedError: "",
		},
		{
			name:       "directory creation failure",
			installDir: "/usr/local/go",
			mockSetup: func(mockFS *mockFilesystem.MockFileSystem) {
				mockFS.EXPECT().MkdirAll("/usr/local", os.FileMode(directoryPermissions)).Return(ErrDirectoryCreationFailed).Once()
			},
			expectedError: "install failed at prepare phase",
		},
		{
			name:       "nested directory creation",
			installDir: "/opt/custom/go",
			mockSetup: func(mockFS *mockFilesystem.MockFileSystem) {
				mockFS.EXPECT().MkdirAll("/opt/custom", os.FileMode(directoryPermissions)).Return(nil).Once()
			},
			expectedError: "",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// Setup mocks
			mockFS := mockFilesystem.NewMockFileSystem(t)
			testCase.mockSetup(mockFS)

			// Create installer with mocks
			installer := NewInstallerWithDeps(
				mockFS,
				nil, // archiveService not needed
				nil, // downloadService not needed
				nil, // verifyService not needed
				nil, // versionService not needed
				nil, // privilegeService not needed
				nil, // reader not needed
			)

			// Execute
			err := installer.prepareInstallDir(testCase.installDir)

			// Assert
			if testCase.expectedError == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
			}
		})
	}
}

func TestRunCommand(t *testing.T) {
	t.Parallel()
	// RunCommand uses real filesystem operations and cannot be unit tested
	// without significant refactoring to accept dependency injection.
	// This function is intended for command-line usage only.
	t.Skip("RunCommand performs real filesystem operations and cannot be unit tested")
}

// Test edge cases and error scenarios.
func TestInstall_EdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("privilege service nil", func(t *testing.T) {
		t.Parallel()
		installer := NewInstallerWithDeps(
			mockFilesystem.NewMockFileSystem(t),
			nil, nil, nil, nil, nil, nil,
		)

		// This will panic because privilegeService is nil
		assert.Panics(t, func() {
			_ = installer.Install("/usr/local/go", "")
		})
	})
}

func TestDirectExtract_EdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("empty archive path", func(t *testing.T) {
		t.Parallel()
		mockArch := mockInstall.NewMockArchiveService(t)
		mockVer := mockInstall.NewMockVerifyService(t)
		mockFS := mockFilesystem.NewMockFileSystem(t)

		mockFS.EXPECT().MkdirAll(".", os.FileMode(directoryPermissions)).Return(nil).Once()
		mockArch.EXPECT().Extract("", ".").Return(nil).Once()
		mockArch.EXPECT().ExtractVersion("").Return("").Once()
		mockVer.EXPECT().Installation("", "").Return(nil).Once()

		installer := NewInstallerWithDeps(mockFS, mockArch, nil, mockVer, nil, nil, nil)

		err := installer.DirectExtract("", "")
		assert.NoError(t, err)
	})

	t.Run("archive extraction with path conflicts", func(t *testing.T) {
		t.Parallel()
		mockArch := mockInstall.NewMockArchiveService(t)
		mockVer := mockInstall.NewMockVerifyService(t)
		mockFS := mockFilesystem.NewMockFileSystem(t)

		mockFS.EXPECT().MkdirAll("/existing", os.FileMode(directoryPermissions)).Return(nil).Once()
		mockArch.EXPECT().Extract("/tmp/go.tar.gz", "/existing").Return(ErrPathConflict).Once()

		installer := NewInstallerWithDeps(mockFS, mockArch, nil, mockVer, nil, nil, nil)

		err := installer.DirectExtract("/tmp/go.tar.gz", "/existing/go")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "extract")
	})
}

func TestLatest_EdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("temp directory cleanup failure", func(t *testing.T) {
		t.Parallel()
		mockDown := mockInstall.NewMockDownloadService(t)
		mockArch := mockInstall.NewMockArchiveService(t)
		mockVer := mockInstall.NewMockVerifyService(t)
		mockFS := mockFilesystem.NewMockFileSystem(t)

		mockFS.EXPECT().MkdirTemp("", "goUpdater-install-*").Return("/tmp/test", nil).Once()
		mockFS.EXPECT().RemoveAll("/tmp/test").Return(ErrCleanupFailed).Once()
		mockDown.EXPECT().GetLatest("/tmp/test").Return("", "", ErrDownloadFailed).Once()

		installer := NewInstallerWithDeps(mockFS, mockArch, mockDown, mockVer, nil, nil, nil)

		err := installer.Latest("/usr/local/go")
		require.Error(t, err)
		// Even with cleanup failure, the download error should be returned
		assert.Contains(t, err.Error(), "download")
	})
}

func TestHandleExistingInstallation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		installedVersion string
		userInput        string
		mockSetup        func(*mockInstall.MockDownloadService, *mockInstall.MockVersionService)
		expectedError    string
	}{
		{
			name:             "user accepts update",
			installedVersion: "go1.20.0",
			userInput:        "y\n",
			mockSetup: func(mockDown *mockInstall.MockDownloadService, mockVer *mockInstall.MockVersionService) {
				mockDown.EXPECT().GetLatestVersionInfo().Return(types.VersionInfo{Version: "go1.21.0"}, nil).Once()
				mockVer.EXPECT().Compare("go1.20.0", "1.21.0").Return(-1, nil).Once()
			},
			expectedError: "",
		},
		{
			name:             "user accepts update with yes",
			installedVersion: "go1.20.0",
			userInput:        "yes\n",
			mockSetup: func(mockDown *mockInstall.MockDownloadService, mockVer *mockInstall.MockVersionService) {
				mockDown.EXPECT().GetLatestVersionInfo().Return(types.VersionInfo{Version: "go1.21.0"}, nil).Once()
				mockVer.EXPECT().Compare("go1.20.0", "1.21.0").Return(-1, nil).Once()
			},
			expectedError: "",
		},
		{
			name:             "user rejects update",
			installedVersion: "go1.20.0",
			userInput:        "n\n",
			mockSetup: func(mockDown *mockInstall.MockDownloadService, mockVer *mockInstall.MockVersionService) {
				mockDown.EXPECT().GetLatestVersionInfo().Return(types.VersionInfo{Version: "go1.21.0"}, nil).Once()
				mockVer.EXPECT().Compare("go1.20.0", "1.21.0").Return(-1, nil).Once()
			},
			expectedError: "",
		},
		{
			name:             "user rejects update with no",
			installedVersion: "go1.20.0",
			userInput:        "no\n",
			mockSetup: func(mockDown *mockInstall.MockDownloadService, mockVer *mockInstall.MockVersionService) {
				mockDown.EXPECT().GetLatestVersionInfo().Return(types.VersionInfo{Version: "go1.21.0"}, nil).Once()
				mockVer.EXPECT().Compare("go1.20.0", "1.21.0").Return(-1, nil).Once()
			},
			expectedError: "",
		},
		{
			name:             "already up to date",
			installedVersion: "go1.21.0",
			userInput:        "",
			mockSetup: func(mockDown *mockInstall.MockDownloadService, mockVer *mockInstall.MockVersionService) {
				mockDown.EXPECT().GetLatestVersionInfo().Return(types.VersionInfo{Version: "go1.21.0"}, nil).Once()
				mockVer.EXPECT().Compare("go1.21.0", "1.21.0").Return(0, nil).Once()
			},
			expectedError: "",
		},
		{
			name:             "newer version installed",
			installedVersion: "go1.22.0",
			userInput:        "",
			mockSetup: func(mockDown *mockInstall.MockDownloadService, mockVer *mockInstall.MockVersionService) {
				mockDown.EXPECT().GetLatestVersionInfo().Return(types.VersionInfo{Version: "go1.21.0"}, nil).Once()
				mockVer.EXPECT().Compare("go1.22.0", "1.21.0").Return(1, nil).Once()
			},
			expectedError: "",
		},
		{
			name:             "fetch latest version fails",
			installedVersion: "go1.20.0",
			userInput:        "",
			mockSetup: func(mockDown *mockInstall.MockDownloadService, _ *mockInstall.MockVersionService) {
				mockDown.EXPECT().GetLatestVersionInfo().Return(types.VersionInfo{}, ErrNetworkError).Once()
			},
			expectedError: "failed to fetch latest Go version info: network error",
		},
		{
			name:             "version comparison fails",
			installedVersion: "go1.20.0",
			userInput:        "",
			mockSetup: func(mockDown *mockInstall.MockDownloadService, mockVer *mockInstall.MockVersionService) {
				mockDown.EXPECT().GetLatestVersionInfo().Return(types.VersionInfo{Version: "go1.21.0"}, nil).Once()
				mockVer.EXPECT().Compare("go1.20.0", "1.21.0").Return(0, ErrComparisonError).Once()
			},
			expectedError: "failed to compare versions: comparison error",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// Setup mocks
			mockDown := mockInstall.NewMockDownloadService(t)
			mockVer := mockInstall.NewMockVersionService(t)
			testCase.mockSetup(mockDown, mockVer)

			// Create installer with mock reader
			reader := strings.NewReader(testCase.userInput)
			installer := NewInstallerWithDeps(
				mockFilesystem.NewMockFileSystem(t),
				nil, // archiveService not needed
				mockDown,
				nil, // verifyService not needed
				mockVer,
				nil, // privilegeService not needed
				reader,
			)

			// Execute
			err := installer.HandleExistingInstallation("", testCase.installedVersion)

			// Assert
			if testCase.expectedError == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
			}
		})
	}
}

func TestCompare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		version1 string
		version2 string
		expected int
		wantErr  bool
	}{
		{
			name:     "equal versions",
			version1: "go1.21.0",
			version2: "v1.21.0",
			expected: 0,
			wantErr:  false,
		},
		{
			name:     "version1 less than version2",
			version1: "go1.20.0",
			version2: "v1.21.0",
			expected: -1,
			wantErr:  false,
		},
		{
			name:     "version1 greater than version2",
			version1: "go1.22.0",
			version2: "v1.21.0",
			expected: 1,
			wantErr:  false,
		},
		{
			name:     "patch version difference",
			version1: "1.21.0",
			version2: "1.21.1",
			expected: -1,
			wantErr:  false,
		},
		{
			name:     "empty version1",
			version1: "",
			version2: "v1.21.0",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "empty version2",
			version1: "go1.21.0",
			version2: "",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "invalid version1",
			version1: "invalid",
			version2: "v1.21.0",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "invalid version2",
			version1: "go1.21.0",
			version2: "invalid",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "major version difference",
			version1: "go2.0.0",
			version2: "v1.21.0",
			expected: 1,
			wantErr:  false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result, err := compare(testCase.version1, testCase.version2)

			if testCase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.expected, result)
			}
		})
	}
}
