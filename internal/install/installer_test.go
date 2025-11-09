package install

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	mockFilesystem "github.com/nicholas-fedor/goUpdater/internal/filesystem/mocks"
	mockArchive "github.com/nicholas-fedor/goUpdater/internal/install/mocks"
	mockDownload "github.com/nicholas-fedor/goUpdater/internal/install/mocks"
	mockPrivileges "github.com/nicholas-fedor/goUpdater/internal/install/mocks"
	mockVerify "github.com/nicholas-fedor/goUpdater/internal/install/mocks"
	mockVersion "github.com/nicholas-fedor/goUpdater/internal/install/mocks"
	"github.com/nicholas-fedor/goUpdater/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewInstaller(t *testing.T) {
	type args struct {
		fileSystem filesystem.FileSystem
		reader     io.Reader
	}

	tests := []struct {
		name string
		args args
		want *Installer
	}{
		{
			name: "creates installer with provided dependencies",
			args: args{
				fileSystem: &mockFilesystem.MockFileSystem{},
				reader:     strings.NewReader(""),
			},
			want: &Installer{
				fs:               &mockFilesystem.MockFileSystem{},
				ArchiveService:   nil,
				downloadService:  nil,
				verifyService:    nil,
				versionService:   nil,
				privilegeService: nil,
				reader:           strings.NewReader(""),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewInstaller(tt.args.fileSystem, tt.args.reader)
			assert.NotNil(t, got)
			assert.Equal(t, tt.args.fileSystem, got.fs)
			assert.Equal(t, tt.args.reader, got.reader)
			assert.Nil(t, got.ArchiveService)
			assert.Nil(t, got.downloadService)
			assert.Nil(t, got.verifyService)
			assert.Nil(t, got.versionService)
			assert.Nil(t, got.privilegeService)
		})
	}
}

func TestNewInstallerWithDeps(t *testing.T) {
	type args struct {
		fileSystem   filesystem.FileSystem
		archiveSvc   ArchiveService
		downloadSvc  DownloadService
		verifySvc    VerifyService
		versionSvc   VersionService
		privilegeSvc PrivilegeService
		reader       io.Reader
	}

	tests := []struct {
		name string
		args args
		want *Installer
	}{
		{
			name: "creates installer with all dependencies injected",
			args: args{
				fileSystem:   &mockFilesystem.MockFileSystem{},
				archiveSvc:   &mockArchive.MockArchiveService{},
				downloadSvc:  &mockDownload.MockDownloadService{},
				verifySvc:    &mockVerify.MockVerifyService{},
				versionSvc:   &mockVersion.MockVersionService{},
				privilegeSvc: &mockPrivileges.MockPrivilegeService{},
				reader:       strings.NewReader(""),
			},
			want: &Installer{
				fs:               &mockFilesystem.MockFileSystem{},
				ArchiveService:   &mockArchive.MockArchiveService{},
				downloadService:  &mockDownload.MockDownloadService{},
				verifyService:    &mockVerify.MockVerifyService{},
				versionService:   &mockVersion.MockVersionService{},
				privilegeService: &mockPrivileges.MockPrivilegeService{},
				reader:           strings.NewReader(""),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewInstallerWithDeps(tt.args.fileSystem, tt.args.archiveSvc, tt.args.downloadSvc, tt.args.verifySvc, tt.args.versionSvc, tt.args.privilegeSvc, tt.args.reader)
			assert.NotNil(t, got)
			assert.Equal(t, tt.args.fileSystem, got.fs)
			assert.Equal(t, tt.args.archiveSvc, got.ArchiveService)
			assert.Equal(t, tt.args.downloadSvc, got.downloadService)
			assert.Equal(t, tt.args.verifySvc, got.verifyService)
			assert.Equal(t, tt.args.versionSvc, got.versionService)
			assert.Equal(t, tt.args.privilegeSvc, got.privilegeService)
			assert.Equal(t, tt.args.reader, got.reader)
		})
	}
}

func TestInstaller_Install(t *testing.T) {
	tests := []struct {
		name        string
		installDir  string
		archivePath string
		setupMocks  func(*mockVerify.MockVerifyService, *mockPrivileges.MockPrivilegeService)
		wantErr     bool
	}{
		{
			name:        "installs latest version when no archive provided and Go not installed",
			installDir:  "/usr/local/go",
			archivePath: "",
			setupMocks: func(verifySvc *mockVerify.MockVerifyService, privilegeSvc *mockPrivileges.MockPrivilegeService) {
				verifySvc.EXPECT().GetInstalledVersion("/usr/local/go").Return("", nil)
				privilegeSvc.EXPECT().ElevateAndExecute(mock.AnythingOfType("func() error")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "returns no error when Go already installed",
			installDir:  "/usr/local/go",
			archivePath: "",
			setupMocks: func(verifySvc *mockVerify.MockVerifyService, privilegeSvc *mockPrivileges.MockPrivilegeService) {
				verifySvc.EXPECT().GetInstalledVersion("/usr/local/go").Return("1.21.0", nil)
			},
			wantErr: false,
		},
		{
			name:        "returns error when checking existing installation fails",
			installDir:  "/usr/local/go",
			archivePath: "",
			setupMocks: func(verifySvc *mockVerify.MockVerifyService, privilegeSvc *mockPrivileges.MockPrivilegeService) {
				verifySvc.EXPECT().GetInstalledVersion("/usr/local/go").Return("", errors.New("check failed"))
			},
			wantErr: true,
		},
		{
			name:        "installs from archive when archive path provided",
			installDir:  "/usr/local/go",
			archivePath: "/tmp/go.tar.gz",
			setupMocks: func(verifySvc *mockVerify.MockVerifyService, privilegeSvc *mockPrivileges.MockPrivilegeService) {
				privilegeSvc.EXPECT().ElevateAndExecute(mock.AnythingOfType("func() error")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "returns error when archive installation fails",
			installDir:  "/usr/local/go",
			archivePath: "/tmp/go.tar.gz",
			setupMocks: func(verifySvc *mockVerify.MockVerifyService, privilegeSvc *mockPrivileges.MockPrivilegeService) {
				privilegeSvc.EXPECT().ElevateAndExecute(mock.AnythingOfType("func() error")).Return(errors.New("installation failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockVerifySvc := &mockVerify.MockVerifyService{}
			mockPrivilegeSvc := &mockPrivileges.MockPrivilegeService{}

			installer := &Installer{
				fs:               &mockFilesystem.MockFileSystem{},
				verifyService:    mockVerifySvc,
				privilegeService: mockPrivilegeSvc,
				reader:           strings.NewReader(""),
			}

			if tt.setupMocks != nil {
				tt.setupMocks(mockVerifySvc, mockPrivilegeSvc)
			}

			err := installer.Install(tt.installDir, tt.archivePath)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInstaller_DirectExtract(t *testing.T) {
	tests := []struct {
		name        string
		archivePath string
		installDir  string
		setupMocks  func(*mockFilesystem.MockFileSystem, *mockArchive.MockArchiveService, *mockVerify.MockVerifyService)
		wantErr     bool
	}{
		{
			name:        "successfully extracts archive directly",
			archivePath: "/tmp/go.tar.gz",
			installDir:  "/usr/local/go",
			setupMocks: func(fs *mockFilesystem.MockFileSystem, archiveSvc *mockArchive.MockArchiveService, verifySvc *mockVerify.MockVerifyService) {
				fs.EXPECT().MkdirAll("/usr/local", mock.AnythingOfType("fs.FileMode")).Return(nil)
				archiveSvc.EXPECT().Extract("/tmp/go.tar.gz", "/usr/local").Return(nil)
				archiveSvc.EXPECT().ExtractVersion("/tmp/go.tar.gz").Return("1.21.0")
				verifySvc.EXPECT().Installation("/usr/local/go", "1.21.0").Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "returns error when directory preparation fails",
			archivePath: "/tmp/go.tar.gz",
			installDir:  "/usr/local/go",
			setupMocks: func(fs *mockFilesystem.MockFileSystem, archiveSvc *mockArchive.MockArchiveService, verifySvc *mockVerify.MockVerifyService) {
				fs.EXPECT().MkdirAll("/usr/local", mock.AnythingOfType("fs.FileMode")).Return(errors.New("mkdir failed"))
			},
			wantErr: true,
		},
		{
			name:        "returns error when archive extraction fails",
			archivePath: "/tmp/go.tar.gz",
			installDir:  "/usr/local/go",
			setupMocks: func(fs *mockFilesystem.MockFileSystem, archiveSvc *mockArchive.MockArchiveService, verifySvc *mockVerify.MockVerifyService) {
				fs.EXPECT().MkdirAll("/usr/local", mock.AnythingOfType("fs.FileMode")).Return(nil)
				archiveSvc.EXPECT().Extract("/tmp/go.tar.gz", "/usr/local").Return(errors.New("extract failed"))
			},
			wantErr: true,
		},
		{
			name:        "returns error when verification fails",
			archivePath: "/tmp/go.tar.gz",
			installDir:  "/usr/local/go",
			setupMocks: func(fs *mockFilesystem.MockFileSystem, archiveSvc *mockArchive.MockArchiveService, verifySvc *mockVerify.MockVerifyService) {
				fs.EXPECT().MkdirAll("/usr/local", mock.AnythingOfType("fs.FileMode")).Return(nil)
				archiveSvc.EXPECT().Extract("/tmp/go.tar.gz", "/usr/local").Return(nil)
				archiveSvc.EXPECT().ExtractVersion("/tmp/go.tar.gz").Return("1.21.0")
				verifySvc.EXPECT().Installation("/usr/local/go", "1.21.0").Return(errors.New("verification failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockFS := &mockFilesystem.MockFileSystem{}
			mockArchiveSvc := &mockArchive.MockArchiveService{}
			mockVerifySvc := &mockVerify.MockVerifyService{}

			installer := &Installer{
				fs:             mockFS,
				ArchiveService: mockArchiveSvc,
				verifyService:  mockVerifySvc,
			}

			if tt.setupMocks != nil {
				tt.setupMocks(mockFS, mockArchiveSvc, mockVerifySvc)
			}

			err := installer.DirectExtract(tt.archivePath, tt.installDir)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInstaller_Extract(t *testing.T) {
	tests := []struct {
		name        string
		archivePath string
		installDir  string
		checksum    string
		setupMocks  func(*mockFilesystem.MockFileSystem, *mockArchive.MockArchiveService)
		wantErr     bool
	}{
		{
			name:        "successfully extracts archive with validation",
			archivePath: "/tmp/go.tar.gz",
			installDir:  "/usr/local/go",
			checksum:    "abc123",
			setupMocks: func(fs *mockFilesystem.MockFileSystem, archiveSvc *mockArchive.MockArchiveService) {
				archiveSvc.EXPECT().Validate("/tmp/go.tar.gz", "/usr/local").Return(nil)
				fs.EXPECT().MkdirAll("/usr/local", mock.AnythingOfType("fs.FileMode")).Return(nil)
				archiveSvc.EXPECT().Extract("/tmp/go.tar.gz", "/usr/local").Return(nil)
				archiveSvc.EXPECT().ExtractVersion("/tmp/go.tar.gz").Return("1.21.0")
			},
			wantErr: false,
		},
		{
			name:        "returns error when archive validation fails",
			archivePath: "/tmp/go.tar.gz",
			installDir:  "/usr/local/go",
			checksum:    "abc123",
			setupMocks: func(fs *mockFilesystem.MockFileSystem, archiveSvc *mockArchive.MockArchiveService) {
				archiveSvc.EXPECT().Validate("/tmp/go.tar.gz", "/usr/local").Return(errors.New("validation failed"))
			},
			wantErr: true,
		},
		{
			name:        "returns error when directory preparation fails",
			archivePath: "/tmp/go.tar.gz",
			installDir:  "/usr/local/go",
			checksum:    "abc123",
			setupMocks: func(fs *mockFilesystem.MockFileSystem, archiveSvc *mockArchive.MockArchiveService) {
				archiveSvc.EXPECT().Validate("/tmp/go.tar.gz", "/usr/local").Return(nil)
				fs.EXPECT().MkdirAll("/usr/local", mock.AnythingOfType("fs.FileMode")).Return(errors.New("mkdir failed"))
			},
			wantErr: true,
		},
		{
			name:        "returns error when archive extraction fails",
			archivePath: "/tmp/go.tar.gz",
			installDir:  "/usr/local/go",
			checksum:    "abc123",
			setupMocks: func(fs *mockFilesystem.MockFileSystem, archiveSvc *mockArchive.MockArchiveService) {
				archiveSvc.EXPECT().Validate("/tmp/go.tar.gz", "/usr/local").Return(nil)
				fs.EXPECT().MkdirAll("/usr/local", mock.AnythingOfType("fs.FileMode")).Return(nil)
				archiveSvc.EXPECT().Extract("/tmp/go.tar.gz", "/usr/local").Return(errors.New("extract failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockFS := &mockFilesystem.MockFileSystem{}
			mockArchiveSvc := &mockArchive.MockArchiveService{}

			installer := &Installer{
				fs:             mockFS,
				ArchiveService: mockArchiveSvc,
			}

			if tt.setupMocks != nil {
				tt.setupMocks(mockFS, mockArchiveSvc)
			}

			err := installer.Extract(tt.archivePath, tt.installDir, tt.checksum)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInstaller_ExtractWithVerification(t *testing.T) {
	tests := []struct {
		name        string
		archivePath string
		installDir  string
		checksum    string
		setupMocks  func(*mockArchive.MockArchiveService, *mockVerify.MockVerifyService)
		wantErr     bool
	}{
		{
			name:        "successfully extracts and verifies archive",
			archivePath: "/tmp/go.tar.gz",
			installDir:  "/usr/local/go",
			checksum:    "abc123",
			setupMocks: func(archiveSvc *mockArchive.MockArchiveService, verifySvc *mockVerify.MockVerifyService) {
				archiveSvc.EXPECT().ExtractVersion("/tmp/go.tar.gz").Return("1.21.0").Times(2)
				// Extract is called internally and will be mocked in the Extract method
				verifySvc.EXPECT().Installation("/usr/local/go", "1.21.0").Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "returns error when extraction fails",
			archivePath: "/tmp/go.tar.gz",
			installDir:  "/usr/local/go",
			checksum:    "abc123",
			setupMocks: func(archiveSvc *mockArchive.MockArchiveService, verifySvc *mockVerify.MockVerifyService) {
				archiveSvc.EXPECT().ExtractVersion("/tmp/go.tar.gz").Return("1.21.0")
				archiveSvc.EXPECT().Validate("/tmp/go.tar.gz", "/usr/local").Return(errors.New("validation failed"))
			},
			wantErr: true,
		},
		{
			name:        "returns error when verification fails",
			archivePath: "/tmp/go.tar.gz",
			installDir:  "/usr/local/go",
			checksum:    "abc123",
			setupMocks: func(archiveSvc *mockArchive.MockArchiveService, verifySvc *mockVerify.MockVerifyService) {
				archiveSvc.EXPECT().ExtractVersion("/tmp/go.tar.gz").Return("1.21.0").Times(2)
				verifySvc.EXPECT().Installation("/usr/local/go", "1.21.0").Return(errors.New("verification failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockFS := &mockFilesystem.MockFileSystem{}
			mockArchiveSvc := &mockArchive.MockArchiveService{}
			mockVerifySvc := &mockVerify.MockVerifyService{}

			installer := &Installer{
				fs:             mockFS,
				ArchiveService: mockArchiveSvc,
				verifyService:  mockVerifySvc,
			}

			// Setup mocks for the Extract call that happens internally
			if tt.name == "successfully extracts and verifies archive" {
				mockArchiveSvc.EXPECT().Validate("/tmp/go.tar.gz", "/usr/local").Return(nil)
				mockFS.EXPECT().MkdirAll("/usr/local", mock.AnythingOfType("fs.FileMode")).Return(nil)
				mockArchiveSvc.EXPECT().Extract("/tmp/go.tar.gz", "/usr/local").Return(nil)
			}

			if tt.name == "returns error when verification fails" {
				mockArchiveSvc.EXPECT().Validate("/tmp/go.tar.gz", "/usr/local").Return(nil)
				mockFS.EXPECT().MkdirAll("/usr/local", mock.AnythingOfType("fs.FileMode")).Return(nil)
				mockArchiveSvc.EXPECT().Extract("/tmp/go.tar.gz", "/usr/local").Return(nil)
			}

			if tt.setupMocks != nil {
				tt.setupMocks(mockArchiveSvc, mockVerifySvc)
			}

			err := installer.ExtractWithVerification(tt.archivePath, tt.installDir, tt.checksum)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInstaller_Latest(t *testing.T) {
	tests := []struct {
		name       string
		installDir string
		setupMocks func(*mockFilesystem.MockFileSystem, *mockDownload.MockDownloadService)
		wantErr    bool
	}{
		{
			name:       "successfully installs latest version",
			installDir: "/usr/local/go",
			setupMocks: func(fs *mockFilesystem.MockFileSystem, downloadSvc *mockDownload.MockDownloadService) {
				fs.EXPECT().MkdirTemp("", "goUpdater-install-*").Return("/tmp/temp", nil)
				fs.EXPECT().RemoveAll("/tmp/temp").Return(nil)
				downloadSvc.EXPECT().GetLatest("/tmp/temp").Return("/tmp/go.tar.gz", "checksum", nil)
				// ExtractWithVerification will be called internally
			},
			wantErr: false,
		},
		{
			name:       "returns error when temp directory creation fails",
			installDir: "/usr/local/go",
			setupMocks: func(fs *mockFilesystem.MockFileSystem, downloadSvc *mockDownload.MockDownloadService) {
				fs.EXPECT().MkdirTemp("", "goUpdater-install-*").Return("", errors.New("mkdir temp failed"))
			},
			wantErr: true,
		},
		{
			name:       "returns error when download fails",
			installDir: "/usr/local/go",
			setupMocks: func(fs *mockFilesystem.MockFileSystem, downloadSvc *mockDownload.MockDownloadService) {
				fs.EXPECT().MkdirTemp("", "goUpdater-install-*").Return("/tmp/temp", nil)
				fs.EXPECT().RemoveAll("/tmp/temp").Return(nil)
				downloadSvc.EXPECT().GetLatest("/tmp/temp").Return("", "", errors.New("download failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockFS := &mockFilesystem.MockFileSystem{}
			mockDownloadSvc := &mockDownload.MockDownloadService{}
			mockArchiveSvc := &mockArchive.MockArchiveService{}
			mockVerifySvc := &mockVerify.MockVerifyService{}

			installer := &Installer{
				fs:              mockFS,
				downloadService: mockDownloadSvc,
				ArchiveService:  mockArchiveSvc,
				verifyService:   mockVerifySvc,
			}

			// Setup mocks for ExtractWithVerification if needed
			if tt.name == "successfully installs latest version" {
				mockArchiveSvc.EXPECT().ExtractVersion("/tmp/go.tar.gz").Return("1.21.0")
				mockArchiveSvc.EXPECT().Validate("/tmp/go.tar.gz", "/usr/local").Return(nil)
				mockFS.EXPECT().MkdirAll("/usr/local", mock.AnythingOfType("fs.FileMode")).Return(nil)
				mockArchiveSvc.EXPECT().Extract("/tmp/go.tar.gz", "/usr/local").Return(nil)
				mockVerifySvc.EXPECT().Installation("/usr/local/go", "1.21.0").Return(nil)
			}

			if tt.setupMocks != nil {
				tt.setupMocks(mockFS, mockDownloadSvc)
			}

			err := installer.Latest(tt.installDir)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInstaller_HandleExistingInstallation(t *testing.T) {
	tests := []struct {
		name             string
		installDir       string
		installedVersion string
		userInput        string
		setupMocks       func(*mockDownload.MockDownloadService, *mockVersion.MockVersionService)
		wantErr          bool
	}{
		{
			name:             "handles existing installation when version is up to date",
			installDir:       "/usr/local/go",
			installedVersion: "1.21.0",
			setupMocks: func(downloadSvc *mockDownload.MockDownloadService, versionSvc *mockVersion.MockVersionService) {
				downloadSvc.EXPECT().GetLatestVersionInfo().Return(types.VersionInfo{Version: "go1.21.0"}, nil)
				versionSvc.EXPECT().Compare("1.21.0", "1.21.0").Return(0, nil)
			},
			wantErr: false,
		},
		{
			name:             "handles existing installation when version is newer",
			installDir:       "/usr/local/go",
			installedVersion: "1.22.0",
			setupMocks: func(downloadSvc *mockDownload.MockDownloadService, versionSvc *mockVersion.MockVersionService) {
				downloadSvc.EXPECT().GetLatestVersionInfo().Return(types.VersionInfo{Version: "go1.21.0"}, nil)
				versionSvc.EXPECT().Compare("1.22.0", "1.21.0").Return(1, nil)
			},
			wantErr: false,
		},
		{
			name:             "prompts for update when version is older",
			installDir:       "/usr/local/go",
			installedVersion: "1.20.0",
			userInput:        "y\n",
			setupMocks: func(downloadSvc *mockDownload.MockDownloadService, versionSvc *mockVersion.MockVersionService) {
				downloadSvc.EXPECT().GetLatestVersionInfo().Return(types.VersionInfo{Version: "go1.21.0"}, nil)
				versionSvc.EXPECT().Compare("1.20.0", "1.21.0").Return(-1, nil)
			},
			wantErr: false,
		},
		{
			name:             "cancels update when user says no",
			installDir:       "/usr/local/go",
			installedVersion: "1.20.0",
			userInput:        "n\n",
			setupMocks: func(downloadSvc *mockDownload.MockDownloadService, versionSvc *mockVersion.MockVersionService) {
				downloadSvc.EXPECT().GetLatestVersionInfo().Return(types.VersionInfo{Version: "go1.21.0"}, nil)
				versionSvc.EXPECT().Compare("1.20.0", "1.21.0").Return(-1, nil)
			},
			wantErr: false,
		},
		{
			name:             "returns error when fetching latest version fails",
			installDir:       "/usr/local/go",
			installedVersion: "1.20.0",
			setupMocks: func(downloadSvc *mockDownload.MockDownloadService, versionSvc *mockVersion.MockVersionService) {
				downloadSvc.EXPECT().GetLatestVersionInfo().Return(types.VersionInfo{}, errors.New("fetch failed"))
			},
			wantErr: true,
		},
		{
			name:             "returns error when version comparison fails",
			installDir:       "/usr/local/go",
			installedVersion: "1.20.0",
			setupMocks: func(downloadSvc *mockDownload.MockDownloadService, versionSvc *mockVersion.MockVersionService) {
				downloadSvc.EXPECT().GetLatestVersionInfo().Return(types.VersionInfo{Version: "go1.21.0"}, nil)
				versionSvc.EXPECT().Compare("1.20.0", "1.21.0").Return(0, errors.New("compare failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockDownloadSvc := &mockDownload.MockDownloadService{}
			mockVersionSvc := &mockVersion.MockVersionService{}

			var reader io.Reader = strings.NewReader(tt.userInput)
			if tt.userInput == "" {
				reader = strings.NewReader("")
			}

			installer := &Installer{
				downloadService: mockDownloadSvc,
				versionService:  mockVersionSvc,
				reader:          reader,
			}

			if tt.setupMocks != nil {
				tt.setupMocks(mockDownloadSvc, mockVersionSvc)
			}

			err := installer.HandleExistingInstallation(tt.installDir, tt.installedVersion)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInstaller_prepareInstallDir(t *testing.T) {
	tests := []struct {
		name       string
		installDir string
		setupMocks func(*mockFilesystem.MockFileSystem)
		wantErr    bool
	}{
		{
			name:       "successfully prepares installation directory",
			installDir: "/usr/local/go",
			setupMocks: func(fs *mockFilesystem.MockFileSystem) {
				fs.EXPECT().MkdirAll("/usr/local", mock.AnythingOfType("fs.FileMode")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:       "returns error when directory creation fails",
			installDir: "/usr/local/go",
			setupMocks: func(fs *mockFilesystem.MockFileSystem) {
				fs.EXPECT().MkdirAll("/usr/local", mock.AnythingOfType("fs.FileMode")).Return(errors.New("mkdir failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockFS := &mockFilesystem.MockFileSystem{}

			installer := &Installer{
				fs: mockFS,
			}

			if tt.setupMocks != nil {
				tt.setupMocks(mockFS)
			}

			err := installer.prepareInstallDir(tt.installDir)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
