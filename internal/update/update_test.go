// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package update

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mockExec "github.com/nicholas-fedor/goUpdater/internal/exec/mocks"
	mockFilesystem "github.com/nicholas-fedor/goUpdater/internal/filesystem/mocks"
	httpclient "github.com/nicholas-fedor/goUpdater/internal/http"
	mockUpdate "github.com/nicholas-fedor/goUpdater/internal/update/mocks"
)

func Test_createDefaultUpdater(t *testing.T) {
	t.Parallel()
	t.Run("creates updater with default dependencies", func(t *testing.T) {
		t.Parallel()

		updater := createDefaultUpdater()

		require.NotNil(t, updater)
		assert.NotNil(t, updater.fileSystem)
		assert.NotNil(t, updater.commandExecutor)
		assert.NotNil(t, updater.versionFetcher)
		assert.NotNil(t, updater.archiveDownloader)
		assert.NotNil(t, updater.installer)
		assert.NotNil(t, updater.uninstaller)
		assert.NotNil(t, updater.verifier)
		assert.NotNil(t, updater.privilegeManager)
	})
}

func TestUpdate_UpdateWithPrivileges(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		installDir  string
		autoInstall bool
		setupMocks  func(
			*mockFilesystem.MockFileSystem,
			*mockUpdate.MockCommandExecutor,
			*mockUpdate.MockArchiveDownloader,
			*mockUpdate.MockInstaller,
			*mockUpdate.MockUninstaller,
			*mockUpdate.MockVerifier,
			*mockUpdate.MockPrivilegeManager,
			*mockUpdate.MockVersionFetcher,
		)
		wantErr bool
	}{
		{
			name:        "successful update with privileges",
			installDir:  "/usr/local/go",
			autoInstall: false,
			setupMocks: func(
				_ *mockFilesystem.MockFileSystem,
				exec *mockUpdate.MockCommandExecutor,
				_ *mockUpdate.MockArchiveDownloader,
				_ *mockUpdate.MockInstaller,
				_ *mockUpdate.MockUninstaller,
				_ *mockUpdate.MockVerifier,
				privileges *mockUpdate.MockPrivilegeManager,
				_ *mockUpdate.MockVersionFetcher,
			) {
				// Mock version check - Go is installed
				cmd := mockExec.NewMockCommand(t)
				cmd.EXPECT().Output().Return([]byte("go version go1.20.0 linux/amd64"), nil).Once()
				exec.EXPECT().CommandContext(context.Background(), "/usr/local/go/bin/go", []string{"version"}).Return(cmd).Once()

				// Mock privilege elevation for the entire update operation
				privileges.EXPECT().ElevateAndExecute(mock.AnythingOfType("func() error")).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:        "go not installed with autoInstall false",
			installDir:  "/usr/local/go",
			autoInstall: false,
			setupMocks: func(
				_ *mockFilesystem.MockFileSystem,
				exec *mockUpdate.MockCommandExecutor,
				_ *mockUpdate.MockArchiveDownloader,
				_ *mockUpdate.MockInstaller,
				_ *mockUpdate.MockUninstaller,
				_ *mockUpdate.MockVerifier,
				_ *mockUpdate.MockPrivilegeManager,
				_ *mockUpdate.MockVersionFetcher,
			) {
				// Mock version check - Go not found
				cmd := mockExec.NewMockCommand(t)
				cmd.EXPECT().Output().Return(nil, ErrGoNotFound).Once()
				exec.EXPECT().CommandContext(context.Background(), "/usr/local/go/bin/go", []string{"version"}).Return(cmd).Once()
			},
			wantErr: true,
		},
		{
			name:        "privilege elevation error",
			installDir:  "/usr/local/go",
			autoInstall: false,
			setupMocks: func(
				_ *mockFilesystem.MockFileSystem,
				exec *mockUpdate.MockCommandExecutor,
				_ *mockUpdate.MockArchiveDownloader,
				_ *mockUpdate.MockInstaller,
				_ *mockUpdate.MockUninstaller,
				_ *mockUpdate.MockVerifier,
				privileges *mockUpdate.MockPrivilegeManager,
				_ *mockUpdate.MockVersionFetcher,
			) {
				// Mock version check - Go is installed
				cmd := mockExec.NewMockCommand(t)
				cmd.EXPECT().Output().Return([]byte("go version go1.20.0 linux/amd64"), nil).Once()
				exec.EXPECT().CommandContext(context.Background(), "/usr/local/go/bin/go", []string{"version"}).Return(cmd).Once()

				// Mock privilege elevation error
				privileges.EXPECT().ElevateAndExecute(mock.AnythingOfType("func() error")).Return(ErrElevationFailed).Once()
			},
			wantErr: true,
		},
		{
			name:        "go not installed with autoInstall true",
			installDir:  "/usr/local/go",
			autoInstall: true,
			setupMocks: func(
				_ *mockFilesystem.MockFileSystem,
				exec *mockUpdate.MockCommandExecutor,
				_ *mockUpdate.MockArchiveDownloader,
				_ *mockUpdate.MockInstaller,
				_ *mockUpdate.MockUninstaller,
				_ *mockUpdate.MockVerifier,
				privileges *mockUpdate.MockPrivilegeManager,
				_ *mockUpdate.MockVersionFetcher,
			) {
				// Mock version check - Go not found, but autoInstall is true
				cmd := mockExec.NewMockCommand(t)
				cmd.EXPECT().Output().Return(nil, ErrGoNotFound).Once()
				exec.EXPECT().CommandContext(context.Background(), "/usr/local/go/bin/go", []string{"version"}).Return(cmd).Once()

				// Mock privilege elevation for the entire update operation (which will install)
				privileges.EXPECT().ElevateAndExecute(mock.AnythingOfType("func() error")).Return(nil).Once()
			},
			wantErr: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create mocks
			mockFS := mockFilesystem.NewMockFileSystem(t)
			mockCommandExecutor := mockUpdate.NewMockCommandExecutor(t)
			mockDownloader := mockUpdate.NewMockArchiveDownloader(t)
			mockInstaller := mockUpdate.NewMockInstaller(t)
			mockUninstaller := mockUpdate.NewMockUninstaller(t)
			mockVerifier := mockUpdate.NewMockVerifier(t)
			mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
			mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

			// Setup mocks
			if testCase.setupMocks != nil {
				testCase.setupMocks(mockFS, mockCommandExecutor, mockDownloader, mockInstaller,
					mockUninstaller, mockVerifier, mockPrivileges, mockVersionFetcher)
			}

			// Create updater with mocks
			updater := NewUpdater(
				mockFS,
				mockCommandExecutor,
				mockDownloader,
				mockInstaller,
				mockUninstaller,
				mockVerifier,
				mockPrivileges,
				mockVersionFetcher,
			)

			// Test UpdateWithPrivileges method on mocked updater
			err := updater.UpdateWithPrivileges(testCase.installDir, testCase.autoInstall)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdate_Success(t *testing.T) {
	t.Parallel()

	// Create mocks
	mockFS := mockFilesystem.NewMockFileSystem(t)
	execMock := mockUpdate.NewMockCommandExecutor(t)
	mockDownloader := mockUpdate.NewMockArchiveDownloader(t)
	mockInstaller := mockUpdate.NewMockInstaller(t)
	mockUninstaller := mockUpdate.NewMockUninstaller(t)
	mockVerifier := mockUpdate.NewMockVerifier(t)
	mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
	mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

	// Mock version check - Go is installed
	cmd := mockExec.NewMockCommand(t)
	cmd.EXPECT().Output().Return([]byte("go version go1.20.0 linux/amd64"), nil).Once()
	execMock.EXPECT().CommandContext(context.Background(), "/usr/local/go/bin/go", []string{"version"}).Return(cmd).Once()

	// Mock version fetcher
	mockVersionFetcher.EXPECT().GetLatestVersionInfo().Return(&httpclient.GoVersionInfo{Version: "go1.21.0"}, nil).Once()

	// Mock download
	mockDownloader.EXPECT().GetLatest("/tmp/goUpdater-test").
		Return("/tmp/goUpdater-test/go1.21.0.tar.gz", "go1.21.0", nil).
		Once()

	// Mock filesystem operations
	mockFS.EXPECT().MkdirTemp("", "goUpdater-*").Return("/tmp/goUpdater-test", nil).Once()
	mockFS.EXPECT().RemoveAll("/tmp/goUpdater-test").Return(nil).Once()

	// Mock privilege elevation for uninstall
	mockPrivileges.EXPECT().ElevateAndExecute(mock.AnythingOfType("func() error")).Return(nil).Once()

	// Mock installer
	mockInstaller.EXPECT().Extract("/tmp/goUpdater-test/go1.21.0.tar.gz",
		"/usr/local/go", "go1.20.0").Return(nil).Once()

	// Mock verifier
	mockVerifier.EXPECT().Installation("/usr/local/go", "go1.21.0").Return(nil).Once()

	// Create updater with mocks
	updater := NewUpdater(
		mockFS,
		execMock,
		mockDownloader,
		mockInstaller,
		mockUninstaller,
		mockVerifier,
		mockPrivileges,
		mockVersionFetcher,
	)

	err := updater.Update("/usr/local/go", false)
	assert.NoError(t, err)
}

func TestUpdate_NoUpdateNeeded(t *testing.T) {
	t.Parallel()

	// Create mocks
	mockFS := mockFilesystem.NewMockFileSystem(t)
	execMock := mockUpdate.NewMockCommandExecutor(t)
	mockDownloader := mockUpdate.NewMockArchiveDownloader(t)
	mockInstaller := mockUpdate.NewMockInstaller(t)
	mockUninstaller := mockUpdate.NewMockUninstaller(t)
	mockVerifier := mockUpdate.NewMockVerifier(t)
	mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
	mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

	// Mock version check - Go is installed
	cmd := mockExec.NewMockCommand(t)
	cmd.EXPECT().Output().Return([]byte("go version go1.21.0 linux/amd64"), nil).Once()
	execMock.EXPECT().CommandContext(context.Background(), "/usr/local/go/bin/go", []string{"version"}).Return(cmd).Once()

	// Mock version fetcher - same version
	mockVersionFetcher.EXPECT().GetLatestVersionInfo().Return(&httpclient.GoVersionInfo{Version: "go1.21.0"}, nil).Once()

	// Create updater with mocks
	updater := NewUpdater(
		mockFS,
		execMock,
		mockDownloader,
		mockInstaller,
		mockUninstaller,
		mockVerifier,
		mockPrivileges,
		mockVersionFetcher,
	)

	err := updater.Update("/usr/local/go", false)
	assert.NoError(t, err)
}

func TestUpdate_GoNotInstalled(t *testing.T) {
	t.Parallel()

	// Create mocks
	mockFS := mockFilesystem.NewMockFileSystem(t)
	execMock := mockUpdate.NewMockCommandExecutor(t)
	mockDownloader := mockUpdate.NewMockArchiveDownloader(t)
	mockInstaller := mockUpdate.NewMockInstaller(t)
	mockUninstaller := mockUpdate.NewMockUninstaller(t)
	mockVerifier := mockUpdate.NewMockVerifier(t)
	mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
	mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

	// Mock version check - Go not found
	cmd := mockExec.NewMockCommand(t)
	cmd.EXPECT().Output().Return(nil, ErrGoNotFound).Once()
	execMock.EXPECT().CommandContext(context.Background(), "/usr/local/go/bin/go", []string{"version"}).Return(cmd).Once()

	// Create updater with mocks
	updater := NewUpdater(
		mockFS,
		execMock,
		mockDownloader,
		mockInstaller,
		mockUninstaller,
		mockVerifier,
		mockPrivileges,
		mockVersionFetcher,
	)

	err := updater.Update("/usr/local/go", false)
	assert.Error(t, err)
}

func TestUpdate_VersionFetchError(t *testing.T) {
	t.Parallel()

	// Create mocks
	mockFS := mockFilesystem.NewMockFileSystem(t)
	execMock := mockUpdate.NewMockCommandExecutor(t)
	mockDownloader := mockUpdate.NewMockArchiveDownloader(t)
	mockInstaller := mockUpdate.NewMockInstaller(t)
	mockUninstaller := mockUpdate.NewMockUninstaller(t)
	mockVerifier := mockUpdate.NewMockVerifier(t)
	mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
	mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

	// Mock version check - Go is installed
	cmd := mockExec.NewMockCommand(t)
	cmd.EXPECT().Output().Return([]byte("go version go1.20.0 linux/amd64"), nil).Once()
	execMock.EXPECT().CommandContext(context.Background(), "/usr/local/go/bin/go", []string{"version"}).Return(cmd).Once()

	// Mock version fetcher error
	mockVersionFetcher.EXPECT().GetLatestVersionInfo().Return(nil, ErrNetworkError).Once()

	// Create updater with mocks
	updater := NewUpdater(
		mockFS,
		execMock,
		mockDownloader,
		mockInstaller,
		mockUninstaller,
		mockVerifier,
		mockPrivileges,
		mockVersionFetcher,
	)

	err := updater.Update("/usr/local/go", false)
	assert.Error(t, err)
}

func TestUpdate_DownloadError(t *testing.T) {
	t.Parallel()

	// Create mocks
	mockFS := mockFilesystem.NewMockFileSystem(t)
	execMock := mockUpdate.NewMockCommandExecutor(t)
	mockDownloader := mockUpdate.NewMockArchiveDownloader(t)
	mockInstaller := mockUpdate.NewMockInstaller(t)
	mockUninstaller := mockUpdate.NewMockUninstaller(t)
	mockVerifier := mockUpdate.NewMockVerifier(t)
	mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
	mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

	// Mock version check - Go is installed
	cmd := mockExec.NewMockCommand(t)
	cmd.EXPECT().Output().Return([]byte("go version go1.20.0 linux/amd64"), nil).Once()
	execMock.EXPECT().CommandContext(context.Background(), "/usr/local/go/bin/go", []string{"version"}).Return(cmd).Once()

	// Mock version fetcher
	mockVersionFetcher.EXPECT().GetLatestVersionInfo().Return(&httpclient.GoVersionInfo{Version: "go1.21.0"}, nil).Once()

	// Mock filesystem operations
	mockFS.EXPECT().MkdirTemp("", "goUpdater-*").Return("/tmp/goUpdater-test", nil).Once()

	// Mock download error
	mockDownloader.EXPECT().GetLatest("/tmp/goUpdater-test").Return("", "", ErrDownloadFailed).Once()

	// Mock cleanup
	mockFS.EXPECT().RemoveAll("/tmp/goUpdater-test").Return(nil).Once()

	// Create updater with mocks
	updater := NewUpdater(
		mockFS,
		execMock,
		mockDownloader,
		mockInstaller,
		mockUninstaller,
		mockVerifier,
		mockPrivileges,
		mockVersionFetcher,
	)

	err := updater.Update("/usr/local/go", false)
	assert.Error(t, err)
}

func TestUpdate_UninstallError(t *testing.T) {
	t.Parallel()

	// Create mocks
	mockFS := mockFilesystem.NewMockFileSystem(t)
	execMock := mockUpdate.NewMockCommandExecutor(t)
	mockDownloader := mockUpdate.NewMockArchiveDownloader(t)
	mockInstaller := mockUpdate.NewMockInstaller(t)
	mockUninstaller := mockUpdate.NewMockUninstaller(t)
	mockVerifier := mockUpdate.NewMockVerifier(t)
	mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
	mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

	// Mock version check - Go is installed
	cmd := mockExec.NewMockCommand(t)
	cmd.EXPECT().Output().Return([]byte("go version go1.20.0 linux/amd64"), nil).Once()
	execMock.EXPECT().CommandContext(context.Background(), "/usr/local/go/bin/go", []string{"version"}).Return(cmd).Once()

	// Mock version fetcher
	mockVersionFetcher.EXPECT().GetLatestVersionInfo().Return(&httpclient.GoVersionInfo{Version: "go1.21.0"}, nil).Once()

	// Mock download
	mockDownloader.EXPECT().GetLatest("/tmp/goUpdater-test").
		Return("/tmp/goUpdater-test/go1.21.0.tar.gz", "go1.21.0", nil).
		Once()

	// Mock filesystem operations
	mockFS.EXPECT().MkdirTemp("", "goUpdater-*").Return("/tmp/goUpdater-test", nil).Once()
	mockFS.EXPECT().RemoveAll("/tmp/goUpdater-test").Return(nil).Once()

	// Mock privilege elevation for uninstall error
	mockPrivileges.EXPECT().ElevateAndExecute(mock.AnythingOfType("func() error")).Return(ErrUninstallFailed).Once()

	// Create updater with mocks
	updater := NewUpdater(
		mockFS,
		execMock,
		mockDownloader,
		mockInstaller,
		mockUninstaller,
		mockVerifier,
		mockPrivileges,
		mockVersionFetcher,
	)

	err := updater.Update("/usr/local/go", false)
	assert.Error(t, err)
}

func TestUpdate_InstallError(t *testing.T) {
	t.Parallel()

	// Create mocks
	mockFS := mockFilesystem.NewMockFileSystem(t)
	execMock := mockUpdate.NewMockCommandExecutor(t)
	mockDownloader := mockUpdate.NewMockArchiveDownloader(t)
	mockInstaller := mockUpdate.NewMockInstaller(t)
	mockUninstaller := mockUpdate.NewMockUninstaller(t)
	mockVerifier := mockUpdate.NewMockVerifier(t)
	mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
	mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

	// Mock version check - Go is installed
	cmd := mockExec.NewMockCommand(t)
	cmd.EXPECT().Output().Return([]byte("go version go1.20.0 linux/amd64"), nil).Once()
	execMock.EXPECT().CommandContext(context.Background(), "/usr/local/go/bin/go", []string{"version"}).Return(cmd).Once()

	// Mock version fetcher
	mockVersionFetcher.EXPECT().GetLatestVersionInfo().Return(&httpclient.GoVersionInfo{Version: "go1.21.0"}, nil).Once()

	// Mock download
	mockDownloader.EXPECT().GetLatest("/tmp/goUpdater-test").
		Return("/tmp/goUpdater-test/go1.21.0.tar.gz", "go1.21.0", nil).
		Once()

	// Mock filesystem operations
	mockFS.EXPECT().MkdirTemp("", "goUpdater-*").Return("/tmp/goUpdater-test", nil).Once()
	mockFS.EXPECT().RemoveAll("/tmp/goUpdater-test").Return(nil).Once()

	// Mock privilege elevation for uninstall
	mockPrivileges.EXPECT().ElevateAndExecute(mock.AnythingOfType("func() error")).Return(nil).Once()

	// Mock installer error
	mockInstaller.EXPECT().Extract("/tmp/goUpdater-test/go1.21.0.tar.gz",
		"/usr/local/go", "go1.20.0").Return(ErrInstallFailed).Once()

	// Create updater with mocks
	updater := NewUpdater(
		mockFS,
		execMock,
		mockDownloader,
		mockInstaller,
		mockUninstaller,
		mockVerifier,
		mockPrivileges,
		mockVersionFetcher,
	)

	err := updater.Update("/usr/local/go", false)
	assert.Error(t, err)
}

func TestUpdate_VerificationError(t *testing.T) {
	t.Parallel()

	// Create mocks
	mockFS := mockFilesystem.NewMockFileSystem(t)
	execMock := mockUpdate.NewMockCommandExecutor(t)
	mockDownloader := mockUpdate.NewMockArchiveDownloader(t)
	mockInstaller := mockUpdate.NewMockInstaller(t)
	mockUninstaller := mockUpdate.NewMockUninstaller(t)
	mockVerifier := mockUpdate.NewMockVerifier(t)
	mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
	mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

	// Mock version check - Go is installed
	cmd := mockExec.NewMockCommand(t)
	cmd.EXPECT().Output().Return([]byte("go version go1.20.0 linux/amd64"), nil).Once()
	execMock.EXPECT().CommandContext(context.Background(), "/usr/local/go/bin/go", []string{"version"}).Return(cmd).Once()

	// Mock version fetcher
	mockVersionFetcher.EXPECT().GetLatestVersionInfo().Return(&httpclient.GoVersionInfo{Version: "go1.21.0"}, nil).Once()

	// Mock download
	mockDownloader.EXPECT().GetLatest("/tmp/goUpdater-test").
		Return("/tmp/goUpdater-test/go1.21.0.tar.gz", "go1.21.0", nil).Once()

	// Mock filesystem operations
	mockFS.EXPECT().MkdirTemp("", "goUpdater-*").Return("/tmp/goUpdater-test", nil).Once()
	mockFS.EXPECT().RemoveAll("/tmp/goUpdater-test").Return(nil).Once()

	// Mock privilege elevation for uninstall
	mockPrivileges.EXPECT().ElevateAndExecute(mock.AnythingOfType("func() error")).Return(nil).Once()

	// Mock installer
	mockInstaller.EXPECT().Extract("/tmp/goUpdater-test/go1.21.0.tar.gz",
		"/usr/local/go", "go1.20.0").Return(nil).Once()

	// Mock verifier error
	mockVerifier.EXPECT().Installation("/usr/local/go", "go1.21.0").Return(ErrVerificationFailed).Once()

	// Create updater with mocks
	updater := NewUpdater(
		mockFS,
		execMock,
		mockDownloader,
		mockInstaller,
		mockUninstaller,
		mockVerifier,
		mockPrivileges,
		mockVersionFetcher,
	)

	err := updater.Update("/usr/local/go", false)
	assert.Error(t, err)
}

func TestUpdateWithPrivileges_Success(t *testing.T) {
	t.Parallel()

	// Create mocks
	mockFS := mockFilesystem.NewMockFileSystem(t)
	execMock := mockUpdate.NewMockCommandExecutor(t)
	mockDownloader := mockUpdate.NewMockArchiveDownloader(t)
	mockInstaller := mockUpdate.NewMockInstaller(t)
	mockUninstaller := mockUpdate.NewMockUninstaller(t)
	mockVerifier := mockUpdate.NewMockVerifier(t)
	mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
	mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

	// Mock version check - Go is installed
	cmd := mockExec.NewMockCommand(t)
	cmd.EXPECT().Output().Return([]byte("go version go1.20.0 linux/amd64"), nil).Once()
	execMock.EXPECT().CommandContext(context.Background(), "/usr/local/go/bin/go", []string{"version"}).Return(cmd).Once()

	// Mock privilege elevation for the entire update operation
	mockPrivileges.EXPECT().ElevateAndExecute(mock.AnythingOfType("func() error")).Return(nil).Once()

	// Create updater with mocks
	updater := NewUpdater(
		mockFS,
		execMock,
		mockDownloader,
		mockInstaller,
		mockUninstaller,
		mockVerifier,
		mockPrivileges,
		mockVersionFetcher,
	)

	// Test UpdateWithPrivileges method on mocked updater
	err := updater.UpdateWithPrivileges("/usr/local/go", false)
	assert.NoError(t, err)
}

func TestUpdateWithPrivileges_GoNotInstalled(t *testing.T) {
	t.Parallel()

	// Create mocks
	mockFS := mockFilesystem.NewMockFileSystem(t)
	execMock := mockUpdate.NewMockCommandExecutor(t)
	mockDownloader := mockUpdate.NewMockArchiveDownloader(t)
	mockInstaller := mockUpdate.NewMockInstaller(t)
	mockUninstaller := mockUpdate.NewMockUninstaller(t)
	mockVerifier := mockUpdate.NewMockVerifier(t)
	mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
	mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

	// Mock version check - Go not found
	cmd := mockExec.NewMockCommand(t)
	cmd.EXPECT().Output().Return(nil, ErrGoNotFound).Once()
	execMock.EXPECT().CommandContext(context.Background(), "/usr/local/go/bin/go", []string{"version"}).Return(cmd).Once()

	// Create updater with mocks
	updater := NewUpdater(
		mockFS,
		execMock,
		mockDownloader,
		mockInstaller,
		mockUninstaller,
		mockVerifier,
		mockPrivileges,
		mockVersionFetcher,
	)

	// Test UpdateWithPrivileges method on mocked updater
	err := updater.UpdateWithPrivileges("/usr/local/go", false)
	assert.Error(t, err)
}

func TestUpdateWithPrivileges_ElevationError(t *testing.T) {
	t.Parallel()

	// Create mocks
	mockFS := mockFilesystem.NewMockFileSystem(t)
	execMock := mockUpdate.NewMockCommandExecutor(t)
	mockDownloader := mockUpdate.NewMockArchiveDownloader(t)
	mockInstaller := mockUpdate.NewMockInstaller(t)
	mockUninstaller := mockUpdate.NewMockUninstaller(t)
	mockVerifier := mockUpdate.NewMockVerifier(t)
	mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
	mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

	// Mock version check - Go is installed
	cmd := mockExec.NewMockCommand(t)
	cmd.EXPECT().Output().Return([]byte("go version go1.20.0 linux/amd64"), nil).Once()
	execMock.EXPECT().CommandContext(context.Background(), "/usr/local/go/bin/go", []string{"version"}).Return(cmd).Once()

	// Mock privilege elevation error
	mockPrivileges.EXPECT().ElevateAndExecute(mock.AnythingOfType("func() error")).Return(ErrElevationFailed).Once()

	// Create updater with mocks
	updater := NewUpdater(
		mockFS,
		execMock,
		mockDownloader,
		mockInstaller,
		mockUninstaller,
		mockVerifier,
		mockPrivileges,
		mockVersionFetcher,
	)

	// Test UpdateWithPrivileges method on mocked updater
	err := updater.UpdateWithPrivileges("/usr/local/go", false)
	assert.Error(t, err)
}

func TestUpdateWithPrivileges_AutoInstall(t *testing.T) {
	t.Parallel()

	// Create mocks
	mockFS := mockFilesystem.NewMockFileSystem(t)
	execMock := mockUpdate.NewMockCommandExecutor(t)
	mockDownloader := mockUpdate.NewMockArchiveDownloader(t)
	mockInstaller := mockUpdate.NewMockInstaller(t)
	mockUninstaller := mockUpdate.NewMockUninstaller(t)
	mockVerifier := mockUpdate.NewMockVerifier(t)
	mockPrivileges := mockUpdate.NewMockPrivilegeManager(t)
	mockVersionFetcher := mockUpdate.NewMockVersionFetcher(t)

	// Mock version check - Go not found, but autoInstall is true
	cmd := mockExec.NewMockCommand(t)
	cmd.EXPECT().Output().Return(nil, ErrGoNotFound).Once()
	execMock.EXPECT().CommandContext(context.Background(), "/usr/local/go/bin/go", []string{"version"}).Return(cmd).Once()

	// Mock privilege elevation for the entire update operation (which will install)
	mockPrivileges.EXPECT().ElevateAndExecute(mock.AnythingOfType("func() error")).Return(nil).Once()

	// Create updater with mocks
	updater := NewUpdater(
		mockFS,
		execMock,
		mockDownloader,
		mockInstaller,
		mockUninstaller,
		mockVerifier,
		mockPrivileges,
		mockVersionFetcher,
	)

	// Test UpdateWithPrivileges method on mocked updater
	err := updater.UpdateWithPrivileges("/usr/local/go", true)
	assert.NoError(t, err)
}
