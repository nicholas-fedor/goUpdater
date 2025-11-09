// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package install

import (
	"context"
	"testing"
	"time"

	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	mockFilesystem "github.com/nicholas-fedor/goUpdater/internal/filesystem/mocks"
	mockArchive "github.com/nicholas-fedor/goUpdater/internal/install/mocks"
	mockDownload "github.com/nicholas-fedor/goUpdater/internal/install/mocks"
	mockPrivileges "github.com/nicholas-fedor/goUpdater/internal/install/mocks"
	mockVerify "github.com/nicholas-fedor/goUpdater/internal/install/mocks"
	mockVersion "github.com/nicholas-fedor/goUpdater/internal/install/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRunInstall(t *testing.T) {
	t.Parallel()

	// Note: RunInstall uses real dependencies and performs actual filesystem operations.
	// This test is intentionally minimal as it would require host filesystem access.
	// Comprehensive testing is done via RunInstallWithDeps using mocks.
	t.Run("RunInstall_RequiresHostAccess", func(t *testing.T) {
		t.Parallel()

		// This test documents that RunInstall performs real operations
		// and should not be used in unit tests
		installDir := "/tmp/test-go-install"
		archivePath := "/tmp/test-archive.tar.gz"

		// We expect this to fail in a test environment since it tries to access real filesystem
		err := RunInstall(installDir, archivePath)
		// The exact error depends on the environment, but it should fail
		assert.Error(t, err, "RunInstall should fail in test environment due to real filesystem operations")
	})
}

func TestRunInstallWithDeps(t *testing.T) {
	t.Parallel()

	// Set a short timeout for all tests to prevent hanging
	timeout := 15 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)

	type args struct {
		fs           filesystem.FileSystem
		archiveSvc   ArchiveService
		downloadSvc  DownloadService
		verifySvc    VerifyService
		versionSvc   VersionService
		privilegeSvc PrivilegeService
		installDir   string
		archivePath  string
	}

	tests := []struct {
		name    string
		args    args
		setup   func(*mockArchive.MockArchiveService, *mockDownload.MockDownloadService, *mockVerify.MockVerifyService, *mockVersion.MockVersionService, *mockPrivileges.MockPrivilegeService) //nolint:lll
		wantErr bool
	}{
		{
			name: "SuccessfulLatestInstall_NoExistingInstallation",
			args: args{
				fs:          &mockFilesystem.MockFileSystem{},
				installDir:  "/usr/local/go",
				archivePath: "",
			},
			setup: func(
				_ *mockArchive.MockArchiveService,
				_ *mockDownload.MockDownloadService,
				mockVerifySvc *mockVerify.MockVerifyService,
				_ *mockVersion.MockVersionService,
				mockPrivilegeSvc *mockPrivileges.MockPrivilegeService,
			) {
				// No existing installation
				mockVerifySvc.EXPECT().GetInstalledVersion("/usr/local/go").Return("", nil).Once()
				// Privilege elevation succeeds
				mockPrivilegeSvc.EXPECT().ElevateAndExecute(mock.MatchedBy(func(fn func() error) bool {
					return fn != nil
				})).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "SuccessfulArchiveInstall",
			args: args{
				fs:          &mockFilesystem.MockFileSystem{},
				installDir:  "/usr/local/go",
				archivePath: "/tmp/go1.21.0.linux-amd64.tar.gz",
			},
			setup: func(
				_ *mockArchive.MockArchiveService,
				_ *mockDownload.MockDownloadService,
				_ *mockVerify.MockVerifyService,
				_ *mockVersion.MockVersionService,
				mockPrivilegeSvc *mockPrivileges.MockPrivilegeService,
			) {
				// Privilege elevation succeeds
				mockPrivilegeSvc.EXPECT().ElevateAndExecute(mock.MatchedBy(func(fn func() error) bool {
					return fn != nil
				})).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "AlreadyInstalled_NoAction",
			args: args{
				fs:          &mockFilesystem.MockFileSystem{},
				installDir:  "/usr/local/go",
				archivePath: "",
			},
			setup: func(_ *mockArchive.MockArchiveService, _ *mockDownload.MockDownloadService, mockVerifySvc *mockVerify.MockVerifyService, _ *mockVersion.MockVersionService, _ *mockPrivileges.MockPrivilegeService) {
				// Go is already installed
				mockVerifySvc.EXPECT().GetInstalledVersion("/usr/local/go").Return("1.20.0", nil).Once()
			},
			wantErr: false,
		},
		{
			name: "CheckExistingInstallation_Failure",
			args: args{
				fs:          &mockFilesystem.MockFileSystem{},
				installDir:  "/usr/local/go",
				archivePath: "",
			},
			setup: func(
				_ *mockArchive.MockArchiveService,
				_ *mockDownload.MockDownloadService,
				mockVerifySvc *mockVerify.MockVerifyService,
				_ *mockVersion.MockVersionService,
				_ *mockPrivileges.MockPrivilegeService,
			) {
				// Check existing installation fails
				mockVerifySvc.EXPECT().GetInstalledVersion("/usr/local/go").Return("", ErrNetworkError).Once()
			},
			wantErr: true,
		},
		{
			name: "LatestInstall_PrivilegeElevation_Failure",
			args: args{
				fs:          &mockFilesystem.MockFileSystem{},
				installDir:  "/usr/local/go",
				archivePath: "",
			},
			setup: func(
				_ *mockArchive.MockArchiveService,
				_ *mockDownload.MockDownloadService,
				mockVerifySvc *mockVerify.MockVerifyService,
				_ *mockVersion.MockVersionService,
				mockPrivilegeSvc *mockPrivileges.MockPrivilegeService,
			) {
				// No existing installation
				mockVerifySvc.EXPECT().GetInstalledVersion("/usr/local/go").Return("", nil).Once()
				// Privilege elevation fails
				mockPrivilegeSvc.EXPECT().ElevateAndExecute(mock.MatchedBy(func(fn func() error) bool {
					return fn != nil
				})).Return(ErrPermissionDenied).Once()
			},
			wantErr: true,
		},
		{
			name: "ArchiveInstall_PrivilegeElevation_Failure",
			args: args{
				fs:          &mockFilesystem.MockFileSystem{},
				installDir:  "/usr/local/go",
				archivePath: "/tmp/go1.21.0.linux-amd64.tar.gz",
			},
			setup: func(
				_ *mockArchive.MockArchiveService,
				_ *mockDownload.MockDownloadService,
				_ *mockVerify.MockVerifyService,
				_ *mockVersion.MockVersionService,
				mockPrivilegeSvc *mockPrivileges.MockPrivilegeService,
			) {
				// Privilege elevation fails
				mockPrivilegeSvc.EXPECT().ElevateAndExecute(mock.MatchedBy(func(fn func() error) bool {
					return fn != nil
				})).Return(ErrPermissionDenied).Once()
			},
			wantErr: true,
		},
		{
			name: "EmptyInstallDir",
			args: args{
				fs:          &mockFilesystem.MockFileSystem{},
				installDir:  "",
				archivePath: "/tmp/go1.21.0.linux-amd64.tar.gz",
			},
			setup: func(
				_ *mockArchive.MockArchiveService,
				_ *mockDownload.MockDownloadService,
				_ *mockVerify.MockVerifyService,
				_ *mockVersion.MockVersionService,
				mockPrivilegeSvc *mockPrivileges.MockPrivilegeService,
			) {
				// Privilege elevation succeeds
				mockPrivilegeSvc.EXPECT().ElevateAndExecute(mock.MatchedBy(func(fn func() error) bool {
					return fn != nil
				})).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "InstallDirWithSpaces",
			args: args{
				fs:          &mockFilesystem.MockFileSystem{},
				installDir:  "/usr/local/go with spaces",
				archivePath: "/tmp/go1.21.0.linux-amd64.tar.gz",
			},
			setup: func(
				_ *mockArchive.MockArchiveService,
				_ *mockDownload.MockDownloadService,
				_ *mockVerify.MockVerifyService,
				_ *mockVersion.MockVersionService,
				mockPrivilegeSvc *mockPrivileges.MockPrivilegeService,
			) {
				// Privilege elevation succeeds
				mockPrivilegeSvc.EXPECT().ElevateAndExecute(mock.MatchedBy(func(fn func() error) bool {
					return fn != nil
				})).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "ArchivePathWithSpecialCharacters",
			args: args{
				fs:          &mockFilesystem.MockFileSystem{},
				installDir:  "/usr/local/go",
				archivePath: "/tmp/go1.21.0.linux-amd64.tar.gz.backup",
			},
			setup: func(
				_ *mockArchive.MockArchiveService,
				_ *mockDownload.MockDownloadService,
				_ *mockVerify.MockVerifyService,
				_ *mockVersion.MockVersionService,
				mockPrivilegeSvc *mockPrivileges.MockPrivilegeService,
			) {
				// Privilege elevation succeeds
				mockPrivilegeSvc.EXPECT().ElevateAndExecute(mock.AnythingOfType("func() error")).Return(nil).Once()
			},
			wantErr: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			select {
			case <-ctx.Done():
				t.Fatal("Test timed out")
			default:
			}

			// Create mocks
			mockArchiveSvc := mockArchive.NewMockArchiveService(t)
			mockDownloadSvc := mockDownload.NewMockDownloadService(t)
			mockVerifySvc := mockVerify.NewMockVerifyService(t)
			mockVersionSvc := mockVersion.NewMockVersionService(t)
			mockPrivilegeSvc := mockPrivileges.NewMockPrivilegeService(t)

			// Setup expectations
			if testCase.setup != nil {
				testCase.setup(mockArchiveSvc, mockDownloadSvc, mockVerifySvc, mockVersionSvc, mockPrivilegeSvc)
			}

			// Assign mocks to args
			testCase.args.archiveSvc = mockArchiveSvc
			testCase.args.downloadSvc = mockDownloadSvc
			testCase.args.verifySvc = mockVerifySvc
			testCase.args.versionSvc = mockVersionSvc
			testCase.args.privilegeSvc = mockPrivilegeSvc

			err := RunInstallWithDeps(testCase.args.fs, testCase.args.archiveSvc, testCase.args.downloadSvc, testCase.args.verifySvc, testCase.args.versionSvc, testCase.args.privilegeSvc, testCase.args.installDir, testCase.args.archivePath) //nolint:lll
			if testCase.wantErr {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Expected no error but got: %v", err)
			}
		})
	}
}
