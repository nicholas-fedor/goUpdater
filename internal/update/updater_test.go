// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package update

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nicholas-fedor/goUpdater/internal/exec"
	mockExec "github.com/nicholas-fedor/goUpdater/internal/exec/mocks"
	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	mockFilesystem "github.com/nicholas-fedor/goUpdater/internal/filesystem/mocks"
	httpclient "github.com/nicholas-fedor/goUpdater/internal/http"
	mockUpdate "github.com/nicholas-fedor/goUpdater/internal/update/mocks"
)

// MatcherFunc is a helper for matching function arguments in mocks.
type MatcherFunc func(func() error) bool

func (m MatcherFunc) Matches(x interface{}) bool {
	if fn, ok := x.(func() error); ok {
		return m(fn)
	}

	return false
}

func (m MatcherFunc) String() string {
	return "MatcherFunc"
}

func TestNewUpdater(t *testing.T) {
	t.Parallel()
	mockFS := mockFilesystem.NewMockFileSystem(t)
	mockExecutor := mockExec.NewMockCommandExecutor(t)
	mockArchive := mockUpdate.NewMockArchiveDownloader(t)
	mockInstaller := mockUpdate.NewMockInstaller(t)
	mockUninstaller := mockUpdate.NewMockUninstaller(t)
	mockVerifier := mockUpdate.NewMockVerifier(t)
	mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
	mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

	tests := []struct {
		name string
		args struct {
			fileSystem        filesystem.FileSystem
			commandExecutor   exec.CommandExecutor
			archiveDownloader ArchiveDownloader
			installer         Installer
			uninstaller       Uninstaller
			verifier          Verifier
			privilegeManager  PrivilegeManager
			versionFetcher    VersionFetcher
		}
		want *Updater
	}{
		{
			name: "success - creates updater with all dependencies",
			args: struct {
				fileSystem        filesystem.FileSystem
				commandExecutor   exec.CommandExecutor
				archiveDownloader ArchiveDownloader
				installer         Installer
				uninstaller       Uninstaller
				verifier          Verifier
				privilegeManager  PrivilegeManager
				versionFetcher    VersionFetcher
			}{
				fileSystem:        mockFS,
				commandExecutor:   mockExecutor,
				archiveDownloader: mockArchive,
				installer:         mockInstaller,
				uninstaller:       mockUninstaller,
				verifier:          mockVerifier,
				privilegeManager:  mockPrivileges,
				versionFetcher:    mockVersionFetcher,
			},
			want: &Updater{
				fileSystem:        mockFS,
				commandExecutor:   mockExecutor,
				archiveDownloader: mockArchive,
				installer:         mockInstaller,
				uninstaller:       mockUninstaller,
				verifier:          mockVerifier,
				privilegeManager:  mockPrivileges,
				versionFetcher:    mockVersionFetcher,
			},
		},
		{
			name: "success - creates updater with nil dependencies",
			args: struct {
				fileSystem        filesystem.FileSystem
				commandExecutor   exec.CommandExecutor
				archiveDownloader ArchiveDownloader
				installer         Installer
				uninstaller       Uninstaller
				verifier          Verifier
				privilegeManager  PrivilegeManager
				versionFetcher    VersionFetcher
			}{
				fileSystem:        nil,
				commandExecutor:   nil,
				archiveDownloader: nil,
				installer:         nil,
				uninstaller:       nil,
				verifier:          nil,
				privilegeManager:  nil,
				versionFetcher:    nil,
			},
			want: &Updater{
				fileSystem:        nil,
				commandExecutor:   nil,
				archiveDownloader: nil,
				installer:         nil,
				uninstaller:       nil,
				verifier:          nil,
				privilegeManager:  nil,
				versionFetcher:    nil,
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := NewUpdater(testCase.args.fileSystem, testCase.args.commandExecutor, testCase.args.archiveDownloader,
				testCase.args.installer, testCase.args.uninstaller, testCase.args.verifier, testCase.args.privilegeManager,
				testCase.args.versionFetcher)
			assert.Equal(t, testCase.want, got)
		})
	}
}

func TestUpdater_Update_Success(t *testing.T) {
	t.Parallel()

	// Create mocks
	mockFS := mockFilesystem.NewMockFileSystem(t)
	mockExecutor := mockExec.NewMockCommandExecutor(t)
	mockArchive := mockUpdate.NewMockArchiveDownloader(t)
	mockInstaller := mockUpdate.NewMockInstaller(t)
	mockUninstaller := mockUpdate.NewMockUninstaller(t)
	mockVerifier := mockUpdate.NewMockVerifier(t)
	mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
	mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

	// Mock version check - Go is installed
	mockCmd := mockExec.NewMockCommand(t)
	mockExecutor.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(mockCmd)
	mockCmd.EXPECT().Output().Return([]byte("go version go1.20.0 linux/amd64"), nil)

	// Mock version fetcher
	mockVersionFetcher.EXPECT().GetLatestVersionInfo().Return(&httpclient.GoVersionInfo{Version: "1.21.0"}, nil)

	// Mock download
	mockArchive.EXPECT().GetLatest("/tmp/goUpdater-123").
		Return("/tmp/goUpdater-123/go1.21.0.tar.gz", "go1.21.0", nil).
		Once()

	// Mock filesystem operations
	mockFS.EXPECT().MkdirTemp("", "goUpdater-*").Return("/tmp/goUpdater-123", nil)
	mockFS.EXPECT().RemoveAll("/tmp/goUpdater-123").Return(nil)

	// Mock privilege elevation for uninstall
	mockPrivileges.EXPECT().ElevateAndExecute(mock.MatchedBy(func(x interface{}) bool {
		_, ok := x.(func() error)

		return ok
	})).RunAndReturn(func(operation func() error) error {
		return operation()
	})

	// Mock uninstaller
	mockUninstaller.EXPECT().Remove("/usr/local/go").Return(nil)

	// Mock installer
	mockInstaller.EXPECT().Extract("/tmp/goUpdater-123/go1.21.0.tar.gz",
		"/usr/local/go", "go1.20.0").Return(nil)

	// Mock verifier
	mockVerifier.EXPECT().Installation("/usr/local/go", "go1.21.0").Return(nil)

	// Create updater with mocks
	updater := NewUpdater(
		mockFS,
		mockExecutor,
		mockArchive,
		mockInstaller,
		mockUninstaller,
		mockVerifier,
		mockPrivileges,
		mockVersionFetcher,
	)

	err := updater.Update("/usr/local/go", false)
	require.NoError(t, err)
}

func TestUpdater_Update_NoUpdateNeeded(t *testing.T) {
	t.Parallel()

	// Create mocks
	mockFS := mockFilesystem.NewMockFileSystem(t)
	mockExecutor := mockExec.NewMockCommandExecutor(t)
	mockArchive := mockUpdate.NewMockArchiveDownloader(t)
	mockInstaller := mockUpdate.NewMockInstaller(t)
	mockUninstaller := mockUpdate.NewMockUninstaller(t)
	mockVerifier := mockUpdate.NewMockVerifier(t)
	mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
	mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

	// Mock version check - Go is installed
	mockCmd := mockExec.NewMockCommand(t)
	mockExecutor.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(mockCmd)
	mockCmd.EXPECT().Output().Return([]byte("go version go1.21.0 linux/amd64"), nil)

	// Mock version fetcher - same version
	mockVersionFetcher.EXPECT().GetLatestVersionInfo().Return(&httpclient.GoVersionInfo{Version: "1.21.0"}, nil)

	// Create updater with mocks
	updater := NewUpdater(
		mockFS,
		mockExecutor,
		mockArchive,
		mockInstaller,
		mockUninstaller,
		mockVerifier,
		mockPrivileges,
		mockVersionFetcher,
	)

	err := updater.Update("/usr/local/go", false)
	require.NoError(t, err)
}

func TestUpdater_Update_GoNotInstalled(t *testing.T) {
	t.Parallel()

	// Create mocks
	mockFS := mockFilesystem.NewMockFileSystem(t)
	mockExecutor := mockExec.NewMockCommandExecutor(t)
	mockArchive := mockUpdate.NewMockArchiveDownloader(t)
	mockInstaller := mockUpdate.NewMockInstaller(t)
	mockUninstaller := mockUpdate.NewMockUninstaller(t)
	mockVerifier := mockUpdate.NewMockVerifier(t)
	mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
	mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

	// Mock version check - Go not found
	mockCmd := mockExec.NewMockCommand(t)
	mockExecutor.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(mockCmd)
	mockCmd.EXPECT().Output().Return([]byte(""), ErrExecutableFileNotFound)

	// Create updater with mocks
	updater := NewUpdater(
		mockFS,
		mockExecutor,
		mockArchive,
		mockInstaller,
		mockUninstaller,
		mockVerifier,
		mockPrivileges,
		mockVersionFetcher,
	)

	err := updater.Update("/usr/local/go", false)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrGoNotInstalled)
}

func TestUpdater_Update_VersionFetchError(t *testing.T) {
	t.Parallel()

	// Create mocks
	mockFS := mockFilesystem.NewMockFileSystem(t)
	mockExecutor := mockExec.NewMockCommandExecutor(t)
	mockArchive := mockUpdate.NewMockArchiveDownloader(t)
	mockInstaller := mockUpdate.NewMockInstaller(t)
	mockUninstaller := mockUpdate.NewMockUninstaller(t)
	mockVerifier := mockUpdate.NewMockVerifier(t)
	mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
	mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

	// Mock version check - Go is installed
	mockCmd := mockExec.NewMockCommand(t)
	mockExecutor.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(mockCmd)
	mockCmd.EXPECT().Output().Return([]byte("go version go1.20.0 linux/amd64"), nil)

	// Mock version fetcher error
	mockVersionFetcher.EXPECT().GetLatestVersionInfo().Return(nil, ErrNetworkError)

	// Create updater with mocks
	updater := NewUpdater(
		mockFS,
		mockExecutor,
		mockArchive,
		mockInstaller,
		mockUninstaller,
		mockVerifier,
		mockPrivileges,
		mockVersionFetcher,
	)

	err := updater.Update("/usr/local/go", false)
	require.Error(t, err)
}

func TestUpdater_Update_DownloadError(t *testing.T) {
	t.Parallel()

	// Create mocks
	mockFS := mockFilesystem.NewMockFileSystem(t)
	mockExecutor := mockExec.NewMockCommandExecutor(t)
	mockArchive := mockUpdate.NewMockArchiveDownloader(t)
	mockInstaller := mockUpdate.NewMockInstaller(t)
	mockUninstaller := mockUpdate.NewMockUninstaller(t)
	mockVerifier := mockUpdate.NewMockVerifier(t)
	mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
	mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

	// Mock version check - Go is installed
	mockCmd := mockExec.NewMockCommand(t)
	mockExecutor.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(mockCmd)
	mockCmd.EXPECT().Output().Return([]byte("go version go1.20.0 linux/amd64"), nil)

	// Mock version fetcher
	mockVersionFetcher.EXPECT().GetLatestVersionInfo().Return(&httpclient.GoVersionInfo{Version: "1.21.0"}, nil)

	// Mock filesystem operations
	mockFS.EXPECT().MkdirTemp("", "goUpdater-*").Return("/tmp/goUpdater-123", nil)

	// Mock download error
	mockArchive.EXPECT().GetLatest("/tmp/goUpdater-123").Return("", "", ErrDownloadFailed)

	// Mock cleanup
	mockFS.EXPECT().RemoveAll("/tmp/goUpdater-123").Return(nil)

	// Create updater with mocks
	updater := NewUpdater(
		mockFS,
		mockExecutor,
		mockArchive,
		mockInstaller,
		mockUninstaller,
		mockVerifier,
		mockPrivileges,
		mockVersionFetcher,
	)

	err := updater.Update("/usr/local/go", false)
	require.Error(t, err)
}

func TestUpdater_Update_UninstallError(t *testing.T) {
	t.Parallel()

	// Create mocks
	mockFS := mockFilesystem.NewMockFileSystem(t)
	mockExecutor := mockExec.NewMockCommandExecutor(t)
	mockArchive := mockUpdate.NewMockArchiveDownloader(t)
	mockInstaller := mockUpdate.NewMockInstaller(t)
	mockUninstaller := mockUpdate.NewMockUninstaller(t)
	mockVerifier := mockUpdate.NewMockVerifier(t)
	mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
	mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

	// Mock version check - Go is installed
	mockCmd := mockExec.NewMockCommand(t)
	mockExecutor.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(mockCmd)
	mockCmd.EXPECT().Output().Return([]byte("go version go1.20.0 linux/amd64"), nil)

	// Mock version fetcher
	mockVersionFetcher.EXPECT().GetLatestVersionInfo().Return(&httpclient.GoVersionInfo{Version: "1.21.0"}, nil)

	// Mock download
	mockArchive.EXPECT().GetLatest("/tmp/goUpdater-123").
		Return("/tmp/goUpdater-123/go1.21.0.tar.gz", "go1.21.0", nil).
		Once()

	// Mock filesystem operations
	mockFS.EXPECT().MkdirTemp("", "goUpdater-*").Return("/tmp/goUpdater-123", nil)
	mockFS.EXPECT().RemoveAll("/tmp/goUpdater-123").Return(nil)

	// Mock privilege elevation for uninstall error
	mockPrivileges.EXPECT().ElevateAndExecute(mock.MatchedBy(func(x interface{}) bool {
		_, ok := x.(func() error)

		return ok
	})).RunAndReturn(func(operation func() error) error {
		return operation()
	})

	mockUninstaller.EXPECT().Remove("/usr/local/go").Return(ErrUninstallFailed)

	// Create updater with mocks
	updater := NewUpdater(
		mockFS,
		mockExecutor,
		mockArchive,
		mockInstaller,
		mockUninstaller,
		mockVerifier,
		mockPrivileges,
		mockVersionFetcher,
	)

	err := updater.Update("/usr/local/go", false)
	require.Error(t, err)
}

func TestUpdater_Update_InstallError(t *testing.T) {
	t.Parallel()

	// Create mocks
	mockFS := mockFilesystem.NewMockFileSystem(t)
	mockExecutor := mockExec.NewMockCommandExecutor(t)
	mockArchive := mockUpdate.NewMockArchiveDownloader(t)
	mockInstaller := mockUpdate.NewMockInstaller(t)
	mockUninstaller := mockUpdate.NewMockUninstaller(t)
	mockVerifier := mockUpdate.NewMockVerifier(t)
	mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
	mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

	// Mock version check - Go is installed
	mockCmd := mockExec.NewMockCommand(t)
	mockExecutor.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(mockCmd)
	mockCmd.EXPECT().Output().Return([]byte("go version go1.20.0 linux/amd64"), nil)

	// Mock version fetcher
	mockVersionFetcher.EXPECT().GetLatestVersionInfo().Return(&httpclient.GoVersionInfo{Version: "1.21.0"}, nil)

	// Mock download
	mockArchive.EXPECT().GetLatest("/tmp/goUpdater-123").
		Return("/tmp/goUpdater-123/go1.21.0.tar.gz", "go1.21.0", nil).
		Once()

	// Mock filesystem operations
	mockFS.EXPECT().MkdirTemp("", "goUpdater-*").Return("/tmp/goUpdater-123", nil)
	mockFS.EXPECT().RemoveAll("/tmp/goUpdater-123").Return(nil)

	// Mock privilege elevation for uninstall
	mockPrivileges.EXPECT().ElevateAndExecute(mock.MatchedBy(func(x interface{}) bool {
		_, ok := x.(func() error)

		return ok
	})).RunAndReturn(func(operation func() error) error {
		return operation()
	})

	mockUninstaller.EXPECT().Remove("/usr/local/go").Return(nil)
	mockInstaller.EXPECT().
		Extract("/tmp/goUpdater-123/go1.21.0.tar.gz", "/usr/local/go", "go1.20.0").
		Return(ErrInstallFailed)

	// Create updater with mocks
	updater := NewUpdater(
		mockFS,
		mockExecutor,
		mockArchive,
		mockInstaller,
		mockUninstaller,
		mockVerifier,
		mockPrivileges,
		mockVersionFetcher,
	)

	err := updater.Update("/usr/local/go", false)
	require.Error(t, err)
}

func TestUpdater_Update_VerificationError(t *testing.T) {
	t.Parallel()

	// Create mocks
	mockFS := mockFilesystem.NewMockFileSystem(t)
	mockExecutor := mockExec.NewMockCommandExecutor(t)
	mockArchive := mockUpdate.NewMockArchiveDownloader(t)
	mockInstaller := mockUpdate.NewMockInstaller(t)
	mockUninstaller := mockUpdate.NewMockUninstaller(t)
	mockVerifier := mockUpdate.NewMockVerifier(t)
	mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
	mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

	// Mock version check - Go is installed
	mockCmd := mockExec.NewMockCommand(t)
	mockExecutor.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(mockCmd)
	mockCmd.EXPECT().Output().Return([]byte("go version go1.20.0 linux/amd64"), nil)

	// Mock version fetcher
	mockVersionFetcher.EXPECT().GetLatestVersionInfo().Return(&httpclient.GoVersionInfo{Version: "1.21.0"}, nil)

	// Mock download
	mockArchive.EXPECT().GetLatest("/tmp/goUpdater-123").
		Return("/tmp/goUpdater-123/go1.21.0.tar.gz", "go1.21.0", nil).Once()

	// Mock filesystem operations
	mockFS.EXPECT().MkdirTemp("", "goUpdater-*").Return("/tmp/goUpdater-123", nil)
	mockFS.EXPECT().RemoveAll("/tmp/goUpdater-123").Return(nil)

	// Mock privilege elevation for uninstall
	mockPrivileges.EXPECT().ElevateAndExecute(mock.MatchedBy(func(x interface{}) bool {
		_, ok := x.(func() error)

		return ok
	})).RunAndReturn(func(operation func() error) error {
		return operation()
	})

	mockUninstaller.EXPECT().Remove("/usr/local/go").Return(nil)
	mockInstaller.EXPECT().Extract("/tmp/goUpdater-123/go1.21.0.tar.gz", "/usr/local/go", "go1.20.0").Return(nil)
	mockVerifier.EXPECT().Installation("/usr/local/go", "go1.21.0").Return(ErrVerificationFailed)

	// Create updater with mocks
	updater := NewUpdater(
		mockFS,
		mockExecutor,
		mockArchive,
		mockInstaller,
		mockUninstaller,
		mockVerifier,
		mockPrivileges,
		mockVersionFetcher,
	)

	err := updater.Update("/usr/local/go", false)
	require.Error(t, err)
}

func TestUpdater_UpdateWithPrivileges(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func() *Updater
		args  struct {
			installDir  string
			autoInstall bool
		}
		wantErr bool
		errType error
	}{
		{
			name: "success - update with privileges when Go is installed",
			setup: func() *Updater {
				mockFS := mockFilesystem.NewMockFileSystem(t)
				mockExecutor := mockExec.NewMockCommandExecutor(t)
				mockArchive := mockUpdate.NewMockArchiveDownloader(t)
				mockInstaller := mockUpdate.NewMockInstaller(t)
				mockUninstaller := mockUpdate.NewMockUninstaller(t)
				mockVerifier := mockUpdate.NewMockVerifier(t)
				mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
				mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

				mockCmd := mockExec.NewMockCommand(t)
				mockExecutor.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(mockCmd)
				mockCmd.EXPECT().Output().Return([]byte("go version go1.20.0 linux/amd64"), nil)

				mockPrivileges.EXPECT().ElevateAndExecute(mock.MatchedBy(func(x interface{}) bool {
					_, ok := x.(func() error)

					return ok
				})).RunAndReturn(func(operation func() error) error {
					// Simulate the Update call within ElevateAndExecute
					mockVersionFetcher.EXPECT().GetLatestVersionInfo().Return(&httpclient.GoVersionInfo{Version: "1.21.0"}, nil)
					mockFS.EXPECT().MkdirTemp("", "goUpdater-*").Return("/tmp/goUpdater-123", nil)
					mockArchive.EXPECT().GetLatest("/tmp/goUpdater-123").Return("/tmp/goUpdater-123/go1.21.0.tar.gz", "go1.21.0", nil)
					mockFS.EXPECT().RemoveAll("/tmp/goUpdater-123").Return(nil)

					mockPrivileges.EXPECT().ElevateAndExecute(mock.MatchedBy(func(x interface{}) bool {
						_, ok := x.(func() error)

						return ok
					})).RunAndReturn(func(operation func() error) error {
						return operation()
					})

					mockUninstaller.EXPECT().Remove("/usr/local/go").Return(nil)
					mockInstaller.EXPECT().Extract("/tmp/goUpdater-123/go1.21.0.tar.gz", "/usr/local/go", "go1.20.0").Return(nil)
					mockVerifier.EXPECT().Installation("/usr/local/go", "go1.21.0").Return(nil)

					return operation()
				})

				return NewUpdater(mockFS, mockExecutor, mockArchive, mockInstaller,
					mockUninstaller, mockVerifier, mockPrivileges, mockVersionFetcher)
			},
			args: struct {
				installDir  string
				autoInstall bool
			}{
				installDir:  "/usr/local/go",
				autoInstall: false,
			},
			wantErr: false,
		},
		{
			name: "error - Go not installed and autoInstall false",
			setup: func() *Updater {
				mockFS := mockFilesystem.NewMockFileSystem(t)
				mockExecutor := mockExec.NewMockCommandExecutor(t)
				mockArchive := mockUpdate.NewMockArchiveDownloader(t)
				mockInstaller := mockUpdate.NewMockInstaller(t)
				mockUninstaller := mockUpdate.NewMockUninstaller(t)
				mockVerifier := mockUpdate.NewMockVerifier(t)
				mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
				mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

				mockCmd := mockExec.NewMockCommand(t)
				mockExecutor.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(mockCmd)
				mockCmd.EXPECT().Output().Return([]byte(""), ErrExecutableFileNotFound)

				return NewUpdater(mockFS, mockExecutor, mockArchive, mockInstaller, mockUninstaller,
					mockVerifier, mockPrivileges, mockVersionFetcher)
			},
			args: struct {
				installDir  string
				autoInstall bool
			}{
				installDir:  "/usr/local/go",
				autoInstall: false,
			},
			wantErr: true,
			errType: ErrGoNotInstalled,
		},
		{
			name: "error - privilege elevation fails",
			setup: func() *Updater {
				mockFS := mockFilesystem.NewMockFileSystem(t)
				mockExecutor := mockExec.NewMockCommandExecutor(t)
				mockArchive := mockUpdate.NewMockArchiveDownloader(t)
				mockInstaller := mockUpdate.NewMockInstaller(t)
				mockUninstaller := mockUpdate.NewMockUninstaller(t)
				mockVerifier := mockUpdate.NewMockVerifier(t)
				mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
				mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

				mockCmd := mockExec.NewMockCommand(t)
				mockExecutor.EXPECT().
					CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).
					Return(mockCmd).Once()
				mockCmd.EXPECT().Output().Return([]byte("go version go1.20.0 linux/amd64"), nil).Once()

				mockPrivileges.EXPECT().ElevateAndExecute(mock.MatchedBy(func(x interface{}) bool {
					_, ok := x.(func() error)

					return ok
				})).Return(ErrElevationFailed)

				return NewUpdater(mockFS, mockExecutor, mockArchive, mockInstaller, mockUninstaller,
					mockVerifier, mockPrivileges, mockVersionFetcher)
			},
			args: struct {
				installDir  string
				autoInstall bool
			}{
				installDir:  "/usr/local/go",
				autoInstall: false,
			},
			wantErr: true,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			u := testCase.setup()

			err := u.UpdateWithPrivileges(testCase.args.installDir, testCase.args.autoInstall)
			if testCase.wantErr {
				require.Error(t, err)

				if testCase.errType != nil {
					require.ErrorIs(t, err, testCase.errType)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUpdater_checkAndPrepare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func() *Updater
		args  struct {
			installDir  string
			autoInstall bool
		}
		want    string
		want1   string
		wantErr bool
		errType error
	}{
		{
			name: "success - Go installed and version fetch succeeds",
			setup: func() *Updater {
				mockFS := mockFilesystem.NewMockFileSystem(t)
				mockExecutor := mockExec.NewMockCommandExecutor(t)
				mockArchive := mockUpdate.NewMockArchiveDownloader(t)
				mockInstaller := mockUpdate.NewMockInstaller(t)
				mockUninstaller := mockUpdate.NewMockUninstaller(t)
				mockVerifier := mockUpdate.NewMockVerifier(t)
				mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
				mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

				mockCmd := mockExec.NewMockCommand(t)
				mockExecutor.EXPECT().
					CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).
					Return(mockCmd).Once()
				mockCmd.EXPECT().Output().Return([]byte("go version go1.20.0 linux/amd64"), nil).Once()

				mockVersionFetcher.EXPECT().GetLatestVersionInfo().Return(&httpclient.GoVersionInfo{Version: "1.21.0"}, nil)

				return NewUpdater(mockFS, mockExecutor, mockArchive, mockInstaller, mockUninstaller,
					mockVerifier, mockPrivileges, mockVersionFetcher)
			},
			args: struct {
				installDir  string
				autoInstall bool
			}{
				installDir:  "/usr/local/go",
				autoInstall: false,
			},
			want:    "go1.20.0",
			want1:   "go1.21.0",
			wantErr: false,
		},
		{
			name: "success - Go not installed and autoInstall true",
			setup: func() *Updater {
				mockFS := mockFilesystem.NewMockFileSystem(t)
				mockExecutor := mockExec.NewMockCommandExecutor(t)
				mockArchive := mockUpdate.NewMockArchiveDownloader(t)
				mockInstaller := mockUpdate.NewMockInstaller(t)
				mockUninstaller := mockUpdate.NewMockUninstaller(t)
				mockVerifier := mockUpdate.NewMockVerifier(t)
				mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
				mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

				mockCmd := mockExec.NewMockCommand(t)
				mockExecutor.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(mockCmd)
				mockCmd.EXPECT().Output().Return([]byte(""), ErrExecutableFileNotFound)

				mockVersionFetcher.EXPECT().GetLatestVersionInfo().Return(&httpclient.GoVersionInfo{Version: "go1.21.0"}, nil)

				return NewUpdater(mockFS, mockExecutor, mockArchive, mockInstaller, mockUninstaller,
					mockVerifier, mockPrivileges, mockVersionFetcher)
			},
			args: struct {
				installDir  string
				autoInstall bool
			}{
				installDir:  "/usr/local/go",
				autoInstall: true,
			},
			want:    "",
			want1:   "go1.21.0",
			wantErr: false,
		},
		{
			name: "error - checkInstallation fails",
			setup: func() *Updater {
				mockFS := mockFilesystem.NewMockFileSystem(t)
				mockExecutor := mockExec.NewMockCommandExecutor(t)
				mockArchive := mockUpdate.NewMockArchiveDownloader(t)
				mockInstaller := mockUpdate.NewMockInstaller(t)
				mockUninstaller := mockUpdate.NewMockUninstaller(t)
				mockVerifier := mockUpdate.NewMockVerifier(t)
				mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
				mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

				mockCmd := mockExec.NewMockCommand(t)
				mockExecutor.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(mockCmd)
				mockCmd.EXPECT().Output().Return([]byte(""), ErrExecutableFileNotFound)

				return NewUpdater(mockFS, mockExecutor, mockArchive, mockInstaller,
					mockUninstaller, mockVerifier, mockPrivileges, mockVersionFetcher)
			},
			args: struct {
				installDir  string
				autoInstall bool
			}{
				installDir:  "/usr/local/go",
				autoInstall: false,
			},
			wantErr: true,
			errType: ErrGoNotInstalled,
		},
		{
			name: "error - version fetcher is nil",
			setup: func() *Updater {
				mockFS := mockFilesystem.NewMockFileSystem(t)
				mockExecutor := mockExec.NewMockCommandExecutor(t)
				mockArchive := mockUpdate.NewMockArchiveDownloader(t)
				mockInstaller := mockUpdate.NewMockInstaller(t)
				mockUninstaller := mockUpdate.NewMockUninstaller(t)
				mockVerifier := mockUpdate.NewMockVerifier(t)
				mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)

				mockCmd := mockExec.NewMockCommand(t)
				mockExecutor.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(mockCmd)
				mockCmd.EXPECT().Output().Return([]byte("go version go1.20.0 linux/amd64"), nil)

				return NewUpdater(
					mockFS,
					mockExecutor,
					mockArchive,
					mockInstaller,
					mockUninstaller,
					mockVerifier,
					mockPrivileges,
					nil,
				)
			},
			args: struct {
				installDir  string
				autoInstall bool
			}{
				installDir:  "/usr/local/go",
				autoInstall: false,
			},
			wantErr: true,
			errType: ErrVersionFetcherNil,
		},
		{
			name: "error - version fetch fails",
			setup: func() *Updater {
				mockFS := mockFilesystem.NewMockFileSystem(t)
				mockExecutor := mockExec.NewMockCommandExecutor(t)
				mockArchive := mockUpdate.NewMockArchiveDownloader(t)
				mockInstaller := mockUpdate.NewMockInstaller(t)
				mockUninstaller := mockUpdate.NewMockUninstaller(t)
				mockVerifier := mockUpdate.NewMockVerifier(t)
				mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
				mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

				mockCmd := mockExec.NewMockCommand(t)
				mockExecutor.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(mockCmd)
				mockCmd.EXPECT().Output().Return([]byte("go version go1.20.0 linux/amd64"), nil)

				mockVersionFetcher.EXPECT().GetLatestVersionInfo().Return(nil, ErrNetworkError)

				return NewUpdater(mockFS, mockExecutor, mockArchive, mockInstaller,
					mockUninstaller, mockVerifier, mockPrivileges, mockVersionFetcher)
			},
			args: struct {
				installDir  string
				autoInstall bool
			}{
				installDir:  "/usr/local/go",
				autoInstall: false,
			},
			wantErr: true,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			u := testCase.setup()

			got, got1, err := u.checkAndPrepare(testCase.args.installDir, testCase.args.autoInstall)
			if testCase.wantErr {
				require.Error(t, err)

				if testCase.errType != nil {
					require.ErrorIs(t, err, testCase.errType)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.want, got)
				assert.Equal(t, testCase.want1, got1)
			}
		})
	}
}

func TestUpdater_checkInstallation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func() *Updater
		args  struct {
			installDir  string
			autoInstall bool
		}
		want    string
		wantErr bool
		errType error
	}{
		{
			name: "success - Go installed with valid version",
			setup: func() *Updater {
				mockFS := mockFilesystem.NewMockFileSystem(t)
				mockExecutor := mockExec.NewMockCommandExecutor(t)
				mockArchive := mockUpdate.NewMockArchiveDownloader(t)
				mockInstaller := mockUpdate.NewMockInstaller(t)
				mockUninstaller := mockUpdate.NewMockUninstaller(t)
				mockVerifier := mockUpdate.NewMockVerifier(t)
				mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
				mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

				mockCmd := mockExec.NewMockCommand(t)
				mockExecutor.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(mockCmd)
				mockCmd.EXPECT().Output().Return([]byte("go version go1.21.0 linux/amd64"), nil)

				return NewUpdater(
					mockFS,
					mockExecutor,
					mockArchive,
					mockInstaller,
					mockUninstaller,
					mockVerifier,
					mockPrivileges,
					mockVersionFetcher,
				)
			},
			args: struct {
				installDir  string
				autoInstall bool
			}{
				installDir:  "/usr/local/go",
				autoInstall: false,
			},
			want:    "go1.21.0",
			wantErr: false,
		},
		{
			name: "success - Go not installed and autoInstall true",
			setup: func() *Updater {
				mockFS := mockFilesystem.NewMockFileSystem(t)
				mockExecutor := mockExec.NewMockCommandExecutor(t)
				mockArchive := mockUpdate.NewMockArchiveDownloader(t)
				mockInstaller := mockUpdate.NewMockInstaller(t)
				mockUninstaller := mockUpdate.NewMockUninstaller(t)
				mockVerifier := mockUpdate.NewMockVerifier(t)
				mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
				mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

				mockCmd := mockExec.NewMockCommand(t)
				mockExecutor.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(mockCmd)
				mockCmd.EXPECT().Output().Return([]byte(""), ErrExecutableFileNotFound)

				return NewUpdater(
					mockFS,
					mockExecutor,
					mockArchive,
					mockInstaller,
					mockUninstaller,
					mockVerifier,
					mockPrivileges,
					mockVersionFetcher,
				)
			},
			args: struct {
				installDir  string
				autoInstall bool
			}{
				installDir:  "/usr/local/go",
				autoInstall: true,
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "error - Go not installed and autoInstall false",
			setup: func() *Updater {
				mockFS := mockFilesystem.NewMockFileSystem(t)
				mockExecutor := mockExec.NewMockCommandExecutor(t)
				mockArchive := mockUpdate.NewMockArchiveDownloader(t)
				mockInstaller := mockUpdate.NewMockInstaller(t)
				mockUninstaller := mockUpdate.NewMockUninstaller(t)
				mockVerifier := mockUpdate.NewMockVerifier(t)
				mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
				mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

				mockCmd := mockExec.NewMockCommand(t)
				mockExecutor.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(mockCmd)
				mockCmd.EXPECT().Output().Return([]byte(""), ErrExecutableFileNotFound)

				return NewUpdater(
					mockFS,
					mockExecutor,
					mockArchive,
					mockInstaller,
					mockUninstaller,
					mockVerifier,
					mockPrivileges,
					mockVersionFetcher,
				)
			},
			args: struct {
				installDir  string
				autoInstall bool
			}{
				installDir:  "/usr/local/go",
				autoInstall: false,
			},
			wantErr: true,
			errType: ErrGoNotInstalled,
		},
		{
			name: "error - invalid version format",
			setup: func() *Updater {
				mockFS := mockFilesystem.NewMockFileSystem(t)
				mockExecutor := mockExec.NewMockCommandExecutor(t)
				mockArchive := mockUpdate.NewMockArchiveDownloader(t)
				mockInstaller := mockUpdate.NewMockInstaller(t)
				mockUninstaller := mockUpdate.NewMockUninstaller(t)
				mockVerifier := mockUpdate.NewMockVerifier(t)
				mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
				mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

				mockCmd := mockExec.NewMockCommand(t)
				mockExecutor.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(mockCmd)
				mockCmd.EXPECT().Output().Return([]byte("go version invalid-version linux/amd64"), nil)

				return NewUpdater(
					mockFS,
					mockExecutor,
					mockArchive,
					mockInstaller,
					mockUninstaller,
					mockVerifier,
					mockPrivileges,
					mockVersionFetcher,
				)
			},
			args: struct {
				installDir  string
				autoInstall bool
			}{
				installDir:  "/usr/local/go",
				autoInstall: false,
			},
			wantErr: true,
			errType: ErrUnableToParseVersion,
		},
		{
			name: "error - malformed version output",
			setup: func() *Updater {
				mockFS := mockFilesystem.NewMockFileSystem(t)
				mockExecutor := mockExec.NewMockCommandExecutor(t)
				mockArchive := mockUpdate.NewMockArchiveDownloader(t)
				mockInstaller := mockUpdate.NewMockInstaller(t)
				mockUninstaller := mockUpdate.NewMockUninstaller(t)
				mockVerifier := mockUpdate.NewMockVerifier(t)
				mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
				mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

				mockCmd := mockExec.NewMockCommand(t)
				mockExecutor.EXPECT().CommandContext(mock.Anything, "/usr/local/go/bin/go", []string{"version"}).Return(mockCmd)
				mockCmd.EXPECT().Output().Return([]byte("invalid output format"), nil)

				return NewUpdater(
					mockFS,
					mockExecutor,
					mockArchive,
					mockInstaller,
					mockUninstaller,
					mockVerifier,
					mockPrivileges,
					mockVersionFetcher,
				)
			},
			args: struct {
				installDir  string
				autoInstall bool
			}{
				installDir:  "/usr/local/go",
				autoInstall: false,
			},
			wantErr: true,
			errType: ErrUnableToParseVersion,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			u := testCase.setup()

			got, err := u.checkInstallation(testCase.args.installDir, testCase.args.autoInstall)
			if testCase.wantErr {
				require.Error(t, err)

				if testCase.errType != nil {
					require.ErrorIs(t, err, testCase.errType)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.want, got)
			}
		})
	}
}

func TestUpdater_downloadLatest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func() *Updater
		want    string
		want1   string
		wantErr bool
	}{
		{
			name: "success - download succeeds",
			setup: func() *Updater {
				mockFS := mockFilesystem.NewMockFileSystem(t)
				mockExecutor := mockExec.NewMockCommandExecutor(t)
				mockArchive := mockUpdate.NewMockArchiveDownloader(t)
				mockInstaller := mockUpdate.NewMockInstaller(t)
				mockUninstaller := mockUpdate.NewMockUninstaller(t)
				mockVerifier := mockUpdate.NewMockVerifier(t)
				mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
				mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

				mockFS.EXPECT().MkdirTemp("", "goUpdater-*").Return("/tmp/goUpdater-123", nil)
				mockArchive.EXPECT().GetLatest("/tmp/goUpdater-123").Return("/tmp/goUpdater-123/go1.21.0.tar.gz", "go1.21.0", nil)

				return NewUpdater(
					mockFS,
					mockExecutor,
					mockArchive,
					mockInstaller,
					mockUninstaller,
					mockVerifier,
					mockPrivileges,
					mockVersionFetcher,
				)
			},
			want:    "/tmp/goUpdater-123/go1.21.0.tar.gz",
			want1:   "/tmp/goUpdater-123",
			wantErr: false,
		},
		{
			name: "error - mkdir temp fails",
			setup: func() *Updater {
				mockFS := mockFilesystem.NewMockFileSystem(t)
				mockExecutor := mockExec.NewMockCommandExecutor(t)
				mockArchive := mockUpdate.NewMockArchiveDownloader(t)
				mockInstaller := mockUpdate.NewMockInstaller(t)
				mockUninstaller := mockUpdate.NewMockUninstaller(t)
				mockVerifier := mockUpdate.NewMockVerifier(t)
				mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
				mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

				mockFS.EXPECT().MkdirTemp("", "goUpdater-*").Return("", ErrMkdirTempFailed)

				return NewUpdater(
					mockFS,
					mockExecutor,
					mockArchive,
					mockInstaller,
					mockUninstaller,
					mockVerifier,
					mockPrivileges,
					mockVersionFetcher,
				)
			},
			wantErr: true,
		},
		{
			name: "error - archive download fails",
			setup: func() *Updater {
				mockFS := mockFilesystem.NewMockFileSystem(t)
				mockExecutor := mockExec.NewMockCommandExecutor(t)
				mockArchive := mockUpdate.NewMockArchiveDownloader(t)
				mockInstaller := mockUpdate.NewMockInstaller(t)
				mockUninstaller := mockUpdate.NewMockUninstaller(t)
				mockVerifier := mockUpdate.NewMockVerifier(t)
				mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
				mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

				mockFS.EXPECT().MkdirTemp("", "goUpdater-*").Return("/tmp/goUpdater-123", nil)
				mockArchive.EXPECT().GetLatest("/tmp/goUpdater-123").Return("", "", ErrDownloadFailed)
				mockFS.EXPECT().RemoveAll("/tmp/goUpdater-123").Return(nil)

				return NewUpdater(
					mockFS,
					mockExecutor,
					mockArchive,
					mockInstaller,
					mockUninstaller,
					mockVerifier,
					mockPrivileges,
					mockVersionFetcher,
				)
			},
			wantErr: true,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			u := testCase.setup()

			got, got1, err := u.downloadLatest()
			if testCase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.want, got)
				assert.Equal(t, testCase.want1, got1)
			}
		})
	}
}

func TestUpdater_performUpdate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func() *Updater
		args  struct {
			archivePath      string
			installDir       string
			installedVersion string
		}
		wantErr bool
	}{
		{
			name: "success - update with existing installation",
			setup: func() *Updater {
				mockFS := mockFilesystem.NewMockFileSystem(t)
				mockExecutor := mockExec.NewMockCommandExecutor(t)
				mockArchive := mockUpdate.NewMockArchiveDownloader(t)
				mockInstaller := mockUpdate.NewMockInstaller(t)
				mockUninstaller := mockUpdate.NewMockUninstaller(t)
				mockVerifier := mockUpdate.NewMockVerifier(t)
				mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
				mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

				mockPrivileges.EXPECT().ElevateAndExecute(mock.MatchedBy(func(x interface{}) bool {
					_, ok := x.(func() error)

					return ok
				})).RunAndReturn(func(operation func() error) error {
					return operation()
				})

				mockUninstaller.EXPECT().Remove("/usr/local/go").Return(nil)
				mockInstaller.EXPECT().Extract("/tmp/go1.21.0.tar.gz", "/usr/local/go", "go1.20.0").Return(nil)

				return NewUpdater(
					mockFS,
					mockExecutor,
					mockArchive,
					mockInstaller,
					mockUninstaller,
					mockVerifier,
					mockPrivileges,
					mockVersionFetcher,
				)
			},
			args: struct {
				archivePath      string
				installDir       string
				installedVersion string
			}{
				archivePath:      "/tmp/go1.21.0.tar.gz",
				installDir:       "/usr/local/go",
				installedVersion: "go1.20.0",
			},
			wantErr: false,
		},
		{
			name: "success - fresh install without existing version",
			setup: func() *Updater {
				mockFS := mockFilesystem.NewMockFileSystem(t)
				mockExecutor := mockExec.NewMockCommandExecutor(t)
				mockArchive := mockUpdate.NewMockArchiveDownloader(t)
				mockInstaller := mockUpdate.NewMockInstaller(t)
				mockUninstaller := mockUpdate.NewMockUninstaller(t)
				mockVerifier := mockUpdate.NewMockVerifier(t)
				mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
				mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

				mockInstaller.EXPECT().Extract("/tmp/go1.21.0.tar.gz", "/usr/local/go", "").Return(nil)

				return NewUpdater(
					mockFS,
					mockExecutor,
					mockArchive,
					mockInstaller,
					mockUninstaller,
					mockVerifier,
					mockPrivileges,
					mockVersionFetcher,
				)
			},
			args: struct {
				archivePath      string
				installDir       string
				installedVersion string
			}{
				archivePath:      "/tmp/go1.21.0.tar.gz",
				installDir:       "/usr/local/go",
				installedVersion: "",
			},
			wantErr: false,
		},
		{
			name: "error - uninstall fails",
			setup: func() *Updater {
				mockFS := mockFilesystem.NewMockFileSystem(t)
				mockExecutor := mockExec.NewMockCommandExecutor(t)
				mockArchive := mockUpdate.NewMockArchiveDownloader(t)
				mockInstaller := mockUpdate.NewMockInstaller(t)
				mockUninstaller := mockUpdate.NewMockUninstaller(t)
				mockVerifier := mockUpdate.NewMockVerifier(t)
				mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
				mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

				mockPrivileges.EXPECT().ElevateAndExecute(mock.MatchedBy(func(x interface{}) bool {
					_, ok := x.(func() error)

					return ok
				})).RunAndReturn(func(operation func() error) error {
					return operation()
				})

				mockUninstaller.EXPECT().Remove("/usr/local/go").Return(ErrUninstallFailed)

				return NewUpdater(
					mockFS,
					mockExecutor,
					mockArchive,
					mockInstaller,
					mockUninstaller,
					mockVerifier,
					mockPrivileges,
					mockVersionFetcher,
				)
			},
			args: struct {
				archivePath      string
				installDir       string
				installedVersion string
			}{
				archivePath:      "/tmp/go1.21.0.tar.gz",
				installDir:       "/usr/local/go",
				installedVersion: "go1.20.0",
			},
			wantErr: true,
		},
		{
			name: "error - install fails",
			setup: func() *Updater {
				mockFS := mockFilesystem.NewMockFileSystem(t)
				mockExecutor := mockExec.NewMockCommandExecutor(t)
				mockArchive := mockUpdate.NewMockArchiveDownloader(t)
				mockInstaller := mockUpdate.NewMockInstaller(t)
				mockUninstaller := mockUpdate.NewMockUninstaller(t)
				mockVerifier := mockUpdate.NewMockVerifier(t)
				mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
				mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

				mockPrivileges.EXPECT().ElevateAndExecute(mock.MatchedBy(func(x interface{}) bool {
					_, ok := x.(func() error)

					return ok
				})).RunAndReturn(func(operation func() error) error {
					return operation()
				})

				mockUninstaller.EXPECT().Remove("/usr/local/go").Return(nil)
				mockInstaller.EXPECT().Extract("/tmp/go1.21.0.tar.gz", "/usr/local/go", "go1.20.0").Return(ErrInstallFailed)

				return NewUpdater(
					mockFS,
					mockExecutor,
					mockArchive,
					mockInstaller,
					mockUninstaller,
					mockVerifier,
					mockPrivileges,
					mockVersionFetcher,
				)
			},
			args: struct {
				archivePath      string
				installDir       string
				installedVersion string
			}{
				archivePath:      "/tmp/go1.21.0.tar.gz",
				installDir:       "/usr/local/go",
				installedVersion: "go1.20.0",
			},
			wantErr: true,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			u := testCase.setup()

			err := u.performUpdate(testCase.args.archivePath, testCase.args.installDir, testCase.args.installedVersion)
			if testCase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
