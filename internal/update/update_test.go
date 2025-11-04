// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package update

import (
	"errors"
	"testing"

	httpclient "github.com/nicholas-fedor/goUpdater/internal/http"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mockFilesystem "github.com/nicholas-fedor/goUpdater/internal/filesystem/mocks"
	mockUninstall "github.com/nicholas-fedor/goUpdater/internal/uninstall/mocks"
	mockUpdate "github.com/nicholas-fedor/goUpdater/internal/update/mocks"
)

// Static test errors to satisfy err113 linter rule.
var (
	errExitStatus1Test        = errors.New("exit status 1")
	errNetworkTimeoutTest     = errors.New("network timeout")
	errPermissionDeniedTest   = errors.New("permission denied")
	errExtractionFailedTest   = errors.New("extraction failed")
	errVerificationFailedTest = errors.New("verification failed")
	errUninstallFailedTest    = errors.New("uninstall failed")
	errElevationFailedTest    = errors.New("elevation failed")
)

// testUpdaterDeps holds all mocked dependencies for testing.
type testUpdaterDeps struct {
	updater         *Updater
	mockFS          *mockFilesystem.MockFileSystem
	mockCE          *mockUpdate.MockCommandExecutor
	mockVF          *mockUpdate.MockVersionFetcher
	mockAD          *mockUpdate.MockArchiveDownloader
	mockInstaller   *mockUpdate.MockInstaller
	mockUninstaller *mockUninstall.MockUninstaller
	mockVerifier    *mockUpdate.MockVerifier
	mockPM          *mockUpdate.MockPrivilegeManager
}

// newTestUpdater creates a new Updater with mocked dependencies for testing.
//
//nolint:thelper
func newTestUpdater(t *testing.T) *testUpdaterDeps {
	mockFS := mockFilesystem.NewMockFileSystem(t)
	mockCE := mockUpdate.NewMockCommandExecutor(t)
	mockVF := mockUpdate.NewMockVersionFetcher(t)
	mockAD := mockUpdate.NewMockArchiveDownloader(t)
	mockInstaller := mockUpdate.NewMockInstaller(t)
	mockUninstaller := mockUninstall.NewMockUninstaller(t)
	mockVerifier := mockUpdate.NewMockVerifier(t)
	mockPM := mockUpdate.NewMockPrivilegeManager(t)
	updater := &Updater{
		fileSystem:        mockFS,
		commandExecutor:   mockCE,
		versionFetcher:    mockVF,
		archiveDownloader: mockAD,
		installer:         mockInstaller,
		uninstaller:       mockUninstaller,
		verifier:          mockVerifier,
		privilegeManager:  mockPM,
	}

	return &testUpdaterDeps{
		updater:         updater,
		mockFS:          mockFS,
		mockCE:          mockCE,
		mockVF:          mockVF,
		mockAD:          mockAD,
		mockInstaller:   mockInstaller,
		mockUninstaller: mockUninstaller,
		mockVerifier:    mockVerifier,
		mockPM:          mockPM,
	}
}

// TestUpdateSuccessfulUpdate tests successful update when Go is installed and needs update.
func TestUpdateSuccessfulUpdate(t *testing.T) {
	t.Parallel()
	deps := newTestUpdater(t)
	updater := deps.updater
	mockFS := deps.mockFS
	mockCE := deps.mockCE
	mockVF := deps.mockVF
	mockAD := deps.mockAD
	mockInstaller := deps.mockInstaller
	mockUninstaller := deps.mockUninstaller
	mockVerifier := deps.mockVerifier

	// Mock checkInstallation - Go is installed
	mockCE.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(&mockExecCmd{
		output: []byte("go version go1.20.0 linux/amd64"),
		err:    nil,
		path:   "/usr/local/go/bin/go",
		args:   []string{"version"},
	})

	// Mock version fetcher
	versionInfo := &httpclient.GoVersionInfo{Version: "go1.21.0"}
	mockVF.EXPECT().GetLatestVersionInfo().Return(versionInfo, nil)

	// Mock filesystem operations for downloadLatest
	mockFS.EXPECT().MkdirTemp("", "goUpdater-*").Return("/tmp/goUpdater-123", nil)
	mockFS.EXPECT().RemoveAll("/tmp/goUpdater-123").Return(nil)

	// Mock archive downloader
	mockAD.EXPECT().GetLatest("/tmp/goUpdater-123").Return("/tmp/goUpdater-123/go1.21.0.tar.gz", "/tmp/goUpdater-123", nil)

	// Mock privilege manager
	mockPM := deps.mockPM
	mockPM.EXPECT().ElevateAndExecute(mock.Anything).RunAndReturn(
		func(fn func() error) error {
			return fn()
		},
	)

	// Mock uninstaller
	mockUninstaller.EXPECT().Remove("/usr/local/go").Return(nil)

	// Mock installer
	mockInstaller.EXPECT().Extract(mock.Anything, "/usr/local/go", "go1.20.0").Return(nil)

	// Mock verifier
	mockVerifier.EXPECT().Installation("/usr/local/go", "go1.21.0").Return(nil)

	err := updater.Update("/usr/local/go", false)

	require.NoError(t, err)
}

// TestUpdateNoUpdateNeeded tests no update needed when versions match.
func TestUpdateNoUpdateNeeded(t *testing.T) {
	t.Parallel()
	deps := newTestUpdater(t)
	updater := deps.updater
	mockCE := deps.mockCE
	mockVF := deps.mockVF

	// Mock checkInstallation - Go is installed with latest version
	mockCE.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(&mockExecCmd{
		output: []byte("go version go1.21.0 linux/amd64"),
		err:    nil,
		path:   "/usr/local/go/bin/go",
		args:   []string{"version"},
	})

	// Mock version fetcher
	versionInfo := &httpclient.GoVersionInfo{Version: "go1.21.0"}
	mockVF.EXPECT().GetLatestVersionInfo().Return(versionInfo, nil)

	err := updater.Update("/usr/local/go", false)

	require.NoError(t, err)
}

// TestUpdateAutoInstall tests auto install when Go not installed.
func TestUpdateAutoInstall(t *testing.T) {
	t.Parallel()
	deps := newTestUpdater(t)
	updater := deps.updater
	mockFS := deps.mockFS
	mockCE := deps.mockCE
	mockVF := deps.mockVF
	mockAD := deps.mockAD
	mockInstaller := deps.mockInstaller
	mockVerifier := deps.mockVerifier

	// Mock checkInstallation - Go not installed
	mockCE.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(&mockExecCmd{
		output: nil,
		err:    errExitStatus1Test,
		path:   "/usr/local/go/bin/go",
		args:   []string{"version"},
	})

	// Mock version fetcher
	versionInfo := &httpclient.GoVersionInfo{Version: "go1.21.0"}
	mockVF.EXPECT().GetLatestVersionInfo().Return(versionInfo, nil)

	// Mock filesystem operations for downloadLatest
	mockFS.EXPECT().MkdirTemp("", "goUpdater-*").Return("/tmp/goUpdater-123", nil)
	mockFS.EXPECT().RemoveAll("/tmp/goUpdater-123").Return(nil)

	// Mock archive downloader
	mockAD.EXPECT().GetLatest("/tmp/goUpdater-123").Return("/tmp/goUpdater-123/go1.21.0.tar.gz", "/tmp/goUpdater-123", nil)

	// Mock installer (no uninstall needed for fresh install)
	mockInstaller.EXPECT().Extract(mock.Anything, "/usr/local/go", "").Return(nil)

	// Mock verifier
	mockVerifier.EXPECT().Installation("/usr/local/go", "go1.21.0").Return(nil)

	err := updater.Update("/usr/local/go", true)

	require.NoError(t, err)
}

// TestUpdateGoNotInstalledNoAutoInstall tests error when Go not installed and autoInstall false.
func TestUpdateGoNotInstalledNoAutoInstall(t *testing.T) {
	t.Parallel()
	deps := newTestUpdater(t)
	updater := deps.updater
	mockCE := deps.mockCE

	// Mock checkInstallation - Go not installed
	mockCE.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(&mockExecCmd{
		output: nil,
		err:    errExitStatus1Test,
		path:   "/usr/local/go/bin/go",
		args:   []string{"version"},
	})

	err := updater.Update("/usr/local/go", false)

	require.Error(t, err)
	require.Equal(t, ErrGoNotInstalled, err)
}

// TestUpdateNetworkFailure tests network failure during version fetch.
func TestUpdateNetworkFailure(t *testing.T) {
	t.Parallel()
	deps := newTestUpdater(t)
	updater := deps.updater
	mockCE := deps.mockCE
	mockVF := deps.mockVF

	// Mock checkInstallation - Go is installed
	mockCE.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(&mockExecCmd{
		output: []byte("go version go1.20.0 linux/amd64"),
		err:    nil,
		path:   "/usr/local/go/bin/go",
		args:   []string{"version"},
	})

	// Mock version fetcher failure
	mockVF.EXPECT().GetLatestVersionInfo().Return(nil, errNetworkTimeoutTest)

	err := updater.Update("/usr/local/go", false)

	require.Error(t, err)
	require.Contains(t, err.Error(), "fetch_version")
}

// TestUpdateUninstallPermissionDenied tests permission denied during uninstall.
func TestUpdateUninstallPermissionDenied(t *testing.T) {
	t.Parallel()
	deps := newTestUpdater(t)
	updater := deps.updater
	mockFS := deps.mockFS
	mockCE := deps.mockCE
	mockVF := deps.mockVF
	mockAD := deps.mockAD
	mockPM := deps.mockPM

	// Mock checkInstallation - Go is installed
	mockCE.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(&mockExecCmd{
		output: []byte("go version go1.20.0 linux/amd64"),
		err:    nil,
		path:   "/usr/local/go/bin/go",
		args:   []string{"version"},
	})

	// Mock version fetcher
	versionInfo := &httpclient.GoVersionInfo{Version: "go1.21.0"}
	mockVF.EXPECT().GetLatestVersionInfo().Return(versionInfo, nil)

	// Mock filesystem operations for downloadLatest
	mockFS.EXPECT().MkdirTemp("", "goUpdater-*").Return("/tmp/goUpdater-123", nil)
	mockFS.EXPECT().RemoveAll("/tmp/goUpdater-123").Return(nil)

	// Mock archive downloader
	mockAD.EXPECT().GetLatest("/tmp/goUpdater-123").Return("/tmp/goUpdater-123/go1.21.0.tar.gz", "/tmp/goUpdater-123", nil)

	// Mock privilege manager
	mockPM.EXPECT().ElevateAndExecute(mock.Anything).Return(errPermissionDeniedTest)

	err := updater.Update("/usr/local/go", false)

	require.Error(t, err)
	require.Contains(t, err.Error(), "remove_existing")
}

// TestUpdateInstallationFailure tests installation failure.
func TestUpdateInstallationFailure(t *testing.T) {
	t.Parallel()
	deps := newTestUpdater(t)
	updater := deps.updater
	mockFS := deps.mockFS
	mockCE := deps.mockCE
	mockVF := deps.mockVF
	mockAD := deps.mockAD
	mockInstaller := deps.mockInstaller
	mockUninstaller := deps.mockUninstaller
	mockPM := deps.mockPM

	// Mock checkInstallation - Go is installed
	mockCE.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(&mockExecCmd{
		output: []byte("go version go1.20.0 linux/amd64"),
		err:    nil,
		path:   "/usr/local/go/bin/go",
		args:   []string{"version"},
	})

	// Mock version fetcher
	versionInfo := &httpclient.GoVersionInfo{Version: "go1.21.0"}
	mockVF.EXPECT().GetLatestVersionInfo().Return(versionInfo, nil)

	// Mock filesystem operations for downloadLatest
	mockFS.EXPECT().MkdirTemp("", "goUpdater-*").Return("/tmp/goUpdater-123", nil)
	mockFS.EXPECT().RemoveAll("/tmp/goUpdater-123").Return(nil)

	// Mock archive downloader
	mockAD.EXPECT().GetLatest("/tmp/goUpdater-123").Return("/tmp/goUpdater-123/go1.21.0.tar.gz", "/tmp/goUpdater-123", nil)

	// Mock privilege manager
	mockPM.EXPECT().ElevateAndExecute(mock.Anything).RunAndReturn(
		func(fn func() error) error {
			return fn()
		},
	)

	// Mock uninstaller
	mockUninstaller.EXPECT().Remove("/usr/local/go").Return(nil)

	// Mock installer failure
	mockInstaller.EXPECT().Extract(mock.Anything, "/usr/local/go", "go1.20.0").Return(errExtractionFailedTest)

	err := updater.Update("/usr/local/go", false)

	require.Error(t, err)
	require.Contains(t, err.Error(), "extract_archive")
}

// TestUpdateVerificationFailure tests verification failure.
func TestUpdateVerificationFailure(t *testing.T) {
	t.Parallel()
	deps := newTestUpdater(t)
	updater := deps.updater
	mockFS := deps.mockFS
	mockCE := deps.mockCE
	mockVF := deps.mockVF
	mockAD := deps.mockAD
	mockInstaller := deps.mockInstaller
	mockUninstaller := deps.mockUninstaller
	mockVerifier := deps.mockVerifier
	mockPM := deps.mockPM

	// Mock checkInstallation - Go is installed
	mockCE.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(&mockExecCmd{
		output: []byte("go version go1.20.0 linux/amd64"),
		err:    nil,
		path:   "/usr/local/go/bin/go",
		args:   []string{"version"},
	})

	// Mock version fetcher
	versionInfo := &httpclient.GoVersionInfo{Version: "go1.21.0"}
	mockVF.EXPECT().GetLatestVersionInfo().Return(versionInfo, nil)

	// Mock filesystem operations for downloadLatest
	mockFS.EXPECT().MkdirTemp("", "goUpdater-*").Return("/tmp/goUpdater-123", nil)
	mockFS.EXPECT().RemoveAll("/tmp/goUpdater-123").Return(nil)

	// Mock archive downloader
	mockAD.EXPECT().GetLatest("/tmp/goUpdater-123").Return("/tmp/goUpdater-123/go1.21.0.tar.gz", "/tmp/goUpdater-123", nil)

	// Mock privilege manager
	mockPM.EXPECT().ElevateAndExecute(mock.Anything).RunAndReturn(
		func(fn func() error) error {
			return fn()
		},
	)

	// Mock uninstaller
	mockUninstaller.EXPECT().Remove("/usr/local/go").Return(nil)

	// Mock installer
	mockInstaller.EXPECT().Extract(mock.Anything, "/usr/local/go", "go1.20.0").Return(nil)

	// Mock verifier failure
	mockVerifier.EXPECT().Installation("/usr/local/go", "go1.21.0").Return(errVerificationFailedTest)

	err := updater.Update("/usr/local/go", false)

	require.Error(t, err)
	require.Contains(t, err.Error(), "check_installation")
}

// TestUpdateInvalidVersionFormat tests invalid version format.
func TestUpdateInvalidVersionFormat(t *testing.T) {
	t.Parallel()
	deps := newTestUpdater(t)
	updater := deps.updater
	mockCE := deps.mockCE

	// Mock checkInstallation - invalid version output
	mockCE.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(&mockExecCmd{
		output: []byte("go version invalid linux/amd64"),
		err:    nil,
		path:   "/usr/local/go/bin/go",
		args:   []string{"version"},
	})

	err := updater.Update("/usr/local/go", false)

	require.Error(t, err)
	require.Contains(t, err.Error(), "unable to parse version")
}

// TestUpdateEmptyInstallDirectory tests empty install directory.
func TestUpdateEmptyInstallDirectory(t *testing.T) {
	t.Parallel()
	deps := newTestUpdater(t)
	updater := deps.updater
	mockCE := deps.mockCE

	// Empty install dir should cause issues in checkInstallation
	mockCE.EXPECT().CommandContext(mock.Anything, "bin/go", []string{"version"}).Return(&mockExecCmd{
		output: nil,
		err:    errExitStatus1Test,
		path:   "bin/go",
		args:   []string{"version"},
	})

	err := updater.Update("", false)

	require.Error(t, err)
	require.Equal(t, ErrGoNotInstalled, err)
}

// TestNeedsUpdate tests the needsUpdate function with comprehensive version comparison scenarios.
func TestNeedsUpdate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		installedVersion string
		latestVersion    string
		expected         bool
		expectError      bool
	}{
		{
			name:             "no installed version requires update",
			installedVersion: "",
			latestVersion:    "v1.21.0",
			expected:         true,
		},
		{
			name:             "installed version older than latest",
			installedVersion: "go1.20.0",
			latestVersion:    "v1.21.0",
			expected:         true,
		},
		{
			name:             "installed version same as latest",
			installedVersion: "go1.21.0",
			latestVersion:    "v1.21.0",
			expected:         false,
		},
		{
			name:             "installed version newer than latest",
			installedVersion: "go1.22.0",
			latestVersion:    "v1.21.0",
			expected:         false,
		},
		{
			name:             "version comparison with go prefix",
			installedVersion: "go1.20.0",
			latestVersion:    "v1.21.0",
			expected:         true,
		},
		{
			name:             "version comparison with v prefix",
			installedVersion: "v1.20.0",
			latestVersion:    "v1.21.0",
			expected:         true,
		},
		{
			name:             "patch version comparison",
			installedVersion: "go1.21.0",
			latestVersion:    "v1.21.1",
			expected:         true,
		},
		{
			name:             "complex version comparison",
			installedVersion: "go1.21.5",
			latestVersion:    "v1.21.10",
			expected:         true,
		},
		{
			name:             "invalid installed version",
			installedVersion: "invalid",
			latestVersion:    "v1.21.0",
			expected:         false,
			expectError:      true,
		},
		{
			name:             "invalid latest version",
			installedVersion: "go1.20.0",
			latestVersion:    "invalid",
			expected:         false,
			expectError:      true,
		},
		{
			name:             "empty latest version",
			installedVersion: "go1.20.0",
			latestVersion:    "",
			expected:         false,
			expectError:      true,
		},
		{
			name:             "pre-release versions",
			installedVersion: "go1.21.0",
			latestVersion:    "v1.21.0-rc.1",
			expected:         false,
		},
		{
			name:             "build metadata comparison",
			installedVersion: "go1.21.0+build.1",
			latestVersion:    "v1.21.0+build.2",
			expected:         false,
		},
		{
			name:             "major version difference",
			installedVersion: "go1.20.0",
			latestVersion:    "v2.0.0",
			expected:         true,
		},
		{
			name:             "minor version difference",
			installedVersion: "go1.20.0",
			latestVersion:    "v1.22.0",
			expected:         true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := needsUpdate(testCase.installedVersion, testCase.latestVersion)

			if testCase.expectError {
				// In the current implementation, errors are logged but function returns false
				require.False(t, result)
			} else {
				require.Equal(t, testCase.expected, result)
			}
		})
	}
}

// TestUpdater_checkInstallation tests the checkInstallation method with various scenarios.
func TestUpdater_checkInstallation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		installDir     string
		autoInstall    bool
		setupMocks     func(*mockUpdate.MockCommandExecutor)
		expectedResult string
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:        "successful installation check",
			installDir:  "/usr/local/go",
			autoInstall: false,
			setupMocks: func(ce *mockUpdate.MockCommandExecutor) {
				ce.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(&mockExecCmd{
					output: []byte("go version go1.21.0 linux/amd64"),
					err:    nil,
					path:   "/usr/local/go/bin/go",
					args:   []string{"version"},
				})
			},
			expectedResult: "go1.21.0",
			expectedError:  false,
		},
		{
			name:        "Go not installed, autoInstall true",
			installDir:  "/usr/local/go",
			autoInstall: true,
			setupMocks: func(ce *mockUpdate.MockCommandExecutor) {
				ce.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(&mockExecCmd{
					output: nil,
					err:    errExitStatus1Test,
					path:   "/usr/local/go/bin/go",
					args:   []string{"version"},
				})
			},
			expectedResult: "",
			expectedError:  false,
		},
		{
			name:        "Go not installed, autoInstall false",
			installDir:  "/usr/local/go",
			autoInstall: false,
			setupMocks: func(ce *mockUpdate.MockCommandExecutor) {
				ce.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(&mockExecCmd{
					output: nil,
					err:    errExitStatus1Test,
					path:   "/usr/local/go/bin/go",
					args:   []string{"version"},
				})
			},
			expectedResult: "",
			expectedError:  true,
			expectedErrMsg: "go is not installed",
		},
		{
			name:        "invalid version output format",
			installDir:  "/usr/local/go",
			autoInstall: false,
			setupMocks: func(ce *mockUpdate.MockCommandExecutor) {
				ce.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(&mockExecCmd{
					output: []byte("invalid output format"),
					err:    nil,
					path:   "/usr/local/go/bin/go",
					args:   []string{"version"},
				})
			},
			expectedResult: "",
			expectedError:  true,
			expectedErrMsg: "unable to parse version from output: invalid output format",
		},
		{
			name:        "invalid semver version",
			installDir:  "/usr/local/go",
			autoInstall: false,
			setupMocks: func(ce *mockUpdate.MockCommandExecutor) {
				ce.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(&mockExecCmd{
					output: []byte("go version invalid-version linux/amd64"),
					err:    nil,
					path:   "/usr/local/go/bin/go",
					args:   []string{"version"},
				})
			},
			expectedResult: "",
			expectedError:  true,
			expectedErrMsg: "invalid version format: invalid-version",
		},
		{
			name:        "command execution error",
			installDir:  "/usr/local/go",
			autoInstall: false,
			setupMocks: func(ce *mockUpdate.MockCommandExecutor) {
				ce.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(&mockExecCmd{
					output: nil,
					err:    errExitStatus1Test,
					path:   "/usr/local/go/bin/go",
					args:   []string{"version"},
				})
			},
			expectedResult: "",
			expectedError:  true,
			expectedErrMsg: "go is not installed",
		},
		{
			name:        "empty install directory",
			installDir:  "",
			autoInstall: false,
			setupMocks: func(ce *mockUpdate.MockCommandExecutor) {
				ce.EXPECT().CommandContext(mock.Anything, "bin/go", []string{"version"}).Return(&mockExecCmd{
					output: nil,
					err:    errExitStatus1Test,
					path:   "bin/go",
					args:   []string{"version"},
				})
			},
			expectedResult: "",
			expectedError:  true,
			expectedErrMsg: "go is not installed",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			mockCE := mockUpdate.NewMockCommandExecutor(t)

			if testCase.setupMocks != nil {
				testCase.setupMocks(mockCE)
			}

			updater := &Updater{
				commandExecutor: mockCE,
			}

			result, err := updater.checkInstallation(testCase.installDir, testCase.autoInstall)

			if testCase.expectedError {
				require.Error(t, err)

				if testCase.expectedErrMsg != "" {
					require.Contains(t, err.Error(), testCase.expectedErrMsg)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.expectedResult, result)
			}
		})
	}
}

// TestUpdater_downloadLatest tests the downloadLatest method.
func TestUpdater_downloadLatest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupMocks     func(*mockFilesystem.MockFileSystem, *mockUpdate.MockArchiveDownloader)
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name: "successful download",
			setupMocks: func(fs *mockFilesystem.MockFileSystem, ad *mockUpdate.MockArchiveDownloader) {
				fs.EXPECT().MkdirTemp("", "goUpdater-*").Return("/tmp/goUpdater-123", nil)
				ad.EXPECT().GetLatest("/tmp/goUpdater-123").Return("/tmp/goUpdater-123/go.tar.gz", "/tmp/goUpdater-123", nil)
			},
			expectedError: false,
		},
		{
			name: "temp directory creation fails",
			setupMocks: func(fs *mockFilesystem.MockFileSystem, _ *mockUpdate.MockArchiveDownloader) {
				fs.EXPECT().MkdirTemp("", "goUpdater-*").Return("", errPermissionDeniedTest)
			},
			expectedError:  true,
			expectedErrMsg: "create_temp_dir",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			mockFS := mockFilesystem.NewMockFileSystem(t)
			mockCE := mockUpdate.NewMockCommandExecutor(t)
			mockVF := mockUpdate.NewMockVersionFetcher(t)
			mockAD := mockUpdate.NewMockArchiveDownloader(t)
			mockInstaller := mockUpdate.NewMockInstaller(t)
			mockUninstaller := mockUninstall.NewMockUninstaller(t)
			mockVerifier := mockUpdate.NewMockVerifier(t)
			mockPM := mockUpdate.NewMockPrivilegeManager(t)

			if testCase.setupMocks != nil {
				testCase.setupMocks(mockFS, mockAD)
			}

			updater := &Updater{
				fileSystem:        mockFS,
				commandExecutor:   mockCE,
				versionFetcher:    mockVF,
				archiveDownloader: mockAD,
				installer:         mockInstaller,
				uninstaller:       mockUninstaller,
				verifier:          mockVerifier,
				privilegeManager:  mockPM,
			}

			_, _, err := updater.downloadLatest()

			if testCase.expectedError {
				require.Error(t, err)

				if testCase.expectedErrMsg != "" {
					require.Contains(t, err.Error(), testCase.expectedErrMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestUpdater_performUpdate tests the performUpdate method.
func TestUpdater_performUpdate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		archivePath      string
		installDir       string
		installedVersion string
		setupMocks       func(*mockUpdate.MockInstaller, *mockUninstall.MockUninstaller, *mockUpdate.MockPrivilegeManager)
		expectedError    bool
		expectedErrMsg   string
	}{
		{
			name:             "successful update with existing installation",
			archivePath:      "/tmp/go1.21.0.tar.gz",
			installDir:       "/usr/local/go",
			installedVersion: "go1.20.0",
			setupMocks: func(installer *mockUpdate.MockInstaller,
				uninstaller *mockUninstall.MockUninstaller, privilegeManager *mockUpdate.MockPrivilegeManager) {
				privilegeManager.EXPECT().ElevateAndExecute(mock.Anything).RunAndReturn(
					func(fn func() error) error {
						return fn()
					},
				)
				uninstaller.EXPECT().Remove("/usr/local/go").Return(nil)
				installer.EXPECT().Extract("/tmp/go1.21.0.tar.gz", "/usr/local/go", "go1.20.0").Return(nil)
			},
			expectedError: false,
		},
		{
			name:             "successful fresh installation",
			archivePath:      "/tmp/go1.21.0.tar.gz",
			installDir:       "/usr/local/go",
			installedVersion: "",
			setupMocks: func(installer *mockUpdate.MockInstaller,
				_ *mockUninstall.MockUninstaller, _ *mockUpdate.MockPrivilegeManager) {
				installer.EXPECT().Extract("/tmp/go1.21.0.tar.gz", "/usr/local/go", "").Return(nil)
			},
			expectedError: false,
		},
		{
			name:             "uninstall failure",
			archivePath:      "/tmp/go1.21.0.tar.gz",
			installDir:       "/usr/local/go",
			installedVersion: "go1.20.0",
			setupMocks: func(_ *mockUpdate.MockInstaller, uninstaller *mockUninstall.MockUninstaller,
				privilegeManager *mockUpdate.MockPrivilegeManager) {
				privilegeManager.EXPECT().ElevateAndExecute(mock.Anything).RunAndReturn(
					func(fn func() error) error {
						return fn()
					},
				)
				uninstaller.EXPECT().Remove("/usr/local/go").Return(errUninstallFailedTest)
			},
			expectedError:  true,
			expectedErrMsg: "remove_existing",
		},
		{
			name:             "privilege elevation failure",
			archivePath:      "/tmp/go1.21.0.tar.gz",
			installDir:       "/usr/local/go",
			installedVersion: "go1.20.0",
			setupMocks: func(_ *mockUpdate.MockInstaller, _ *mockUninstall.MockUninstaller,
				privilegeManager *mockUpdate.MockPrivilegeManager) {
				privilegeManager.EXPECT().ElevateAndExecute(mock.Anything).Return(errElevationFailedTest)
			},
			expectedError:  true,
			expectedErrMsg: "remove_existing",
		},
		{
			name:             "installation failure",
			archivePath:      "/tmp/go1.21.0.tar.gz",
			installDir:       "/usr/local/go",
			installedVersion: "",
			setupMocks: func(installer *mockUpdate.MockInstaller, _ *mockUninstall.MockUninstaller,
				_ *mockUpdate.MockPrivilegeManager) {
				installer.EXPECT().Extract("/tmp/go1.21.0.tar.gz", "/usr/local/go", "").Return(errExtractionFailedTest)
			},
			expectedError:  true,
			expectedErrMsg: "extract_archive",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			mockInstaller := mockUpdate.NewMockInstaller(t)
			mockUninstaller := mockUninstall.NewMockUninstaller(t)
			mockPM := mockUpdate.NewMockPrivilegeManager(t)

			if testCase.setupMocks != nil {
				testCase.setupMocks(mockInstaller, mockUninstaller, mockPM)
			}

			updater := &Updater{
				installer:        mockInstaller,
				uninstaller:      mockUninstaller,
				privilegeManager: mockPM,
			}

			err := updater.performUpdate(testCase.archivePath, testCase.installDir, testCase.installedVersion)

			if testCase.expectedError {
				require.Error(t, err)

				if testCase.expectedErrMsg != "" {
					require.Contains(t, err.Error(), testCase.expectedErrMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// mockExecCmd implements exec.ExecCommand interface for testing.
type mockExecCmd struct {
	output []byte
	err    error
	path   string
	args   []string
}

func (m *mockExecCmd) Output() ([]byte, error) {
	return m.output, m.err
}

func (m *mockExecCmd) Path() string {
	return m.path
}

func (m *mockExecCmd) Args() []string {
	return m.args
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
