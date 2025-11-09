package install

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/nicholas-fedor/goUpdater/internal/download"
	mockDownload "github.com/nicholas-fedor/goUpdater/internal/download/mocks"
	mockExec "github.com/nicholas-fedor/goUpdater/internal/exec/mocks"
	mockFilesystem "github.com/nicholas-fedor/goUpdater/internal/filesystem/mocks"
	httpclient "github.com/nicholas-fedor/goUpdater/internal/http"
	"github.com/nicholas-fedor/goUpdater/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockReadCloser implements io.ReadCloser for testing.
type mockReadCloser struct {
	io.Reader
}

func (m *mockReadCloser) Close() error {
	return nil
}

func TestDefaultVersionFetcherImpl_GetLatestVersion(t *testing.T) {
	t.Parallel()

	d := &DefaultVersionFetcherImpl{}

	got, err := d.GetLatestVersion()
	if err != nil {
		assert.Error(t, err)
		assert.Nil(t, got)
	} else {
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.IsType(t, &httpclient.GoVersionInfo{}, got)
	}
}

func TestDownloadServiceImpl_GetLatest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		tempDir    string
		setupMocks func(*mockFilesystem.MockFileSystem, *mockDownload.MockHTTPClient, *mockExec.MockCommandExecutor, *mockDownload.MockVersionFetcher)
		want       string
		want1      string
		wantErr    bool
	}{
		{
			name:    "successful download",
			tempDir: "/tmp",
			setupMocks: func(mfs *mockFilesystem.MockFileSystem, mhc *mockDownload.MockHTTPClient, mec *mockExec.MockCommandExecutor, mvf *mockDownload.MockVersionFetcher) {
				versionInfo := &httpclient.GoVersionInfo{
					Version: "go1.21.0",
					Stable:  true,
					Files: []httpclient.GoFileInfo{
						{
							Filename: "go1.21.0.linux-amd64.tar.gz",
							OS:       "linux",
							Arch:     "amd64",
							Version:  "go1.21.0",
							Sha256:   "26df8a783491d87eb6bae3f16ae0b588dab6b83150c791ebf9d2406bf94a8999",
							Size:     100,
							Kind:     "archive",
						},
					},
				}
				mvf.EXPECT().GetLatestVersion().Return(versionInfo, nil).Once()

				mfs.EXPECT().UserHomeDir().Return("/home/user", nil).Once()

				// Search directories: isReadableDir calls Stat on directories
				mfs.EXPECT().Stat("/home/user/Downloads").Return(&mockFileInfo{name: "/home/user/Downloads", mode: 0755}, nil).Once()
				mfs.EXPECT().Stat("/home/user").Return(&mockFileInfo{name: "/home/user", mode: 0755}, nil).Once()

				// checkExistingArchive calls Stat on directories and files
				mfs.EXPECT().Stat("/home/user/Downloads").Return(&mockFileInfo{name: "/home/user/Downloads", mode: 0755}, nil).Once()
				mfs.EXPECT().Stat("/home/user/Downloads/go1.21.0.linux-amd64.tar.gz").Return(nil, fmt.Errorf("%w", os.ErrNotExist)).Once()
				mfs.EXPECT().IsNotExist(fmt.Errorf("%w", os.ErrNotExist)).Return(true).Once()

				mfs.EXPECT().Stat("/home/user").Return(&mockFileInfo{name: "/home/user", mode: 0755}, nil).Once()
				mfs.EXPECT().Stat("/home/user/go1.21.0.linux-amd64.tar.gz").Return(nil, fmt.Errorf("%w", os.ErrNotExist)).Once()
				mfs.EXPECT().IsNotExist(fmt.Errorf("%w", os.ErrNotExist)).Return(true).Once()

				mfs.EXPECT().Stat("/tmp/go1.21.0.linux-amd64.tar.gz").Return(nil, fmt.Errorf("%w", os.ErrNotExist)).Once()
				mfs.EXPECT().IsNotExist(fmt.Errorf("%w", os.ErrNotExist)).Return(true).Once()

				reader, writer, err := os.Pipe()
				require.NoError(t, err)
				t.Cleanup(func() { reader.Close(); writer.Close() })

				mfs.EXPECT().Create("/tmp/go1.21.0.linux-amd64.tar.gz").Return(writer, nil).Once()
				mfs.EXPECT().Open("/tmp/go1.21.0.linux-amd64.tar.gz").Return(reader, nil).Once()

				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       &mockReadCloser{bytes.NewReader([]byte("mock response"))},
					Header:     make(http.Header),
					Request:    &http.Request{URL: &url.URL{Scheme: "https"}},
				}
				mhc.EXPECT().Do(mock.AnythingOfType("*http.Request")).Return(resp, nil).Once()
			},
			want:    "/tmp/go1.21.0.linux-amd64.tar.gz",
			want1:   "26df8a783491d87eb6bae3f16ae0b588dab6b83150c791ebf9d2406bf94a8999",
			wantErr: false,
		},
		{
			name:    "download failure",
			tempDir: "/tmp",
			setupMocks: func(mfs *mockFilesystem.MockFileSystem, mhc *mockDownload.MockHTTPClient, mec *mockExec.MockCommandExecutor, mvf *mockDownload.MockVersionFetcher) {
				mfs.EXPECT().TempDir().Return("/tmp").Once()
				mvf.EXPECT().GetLatestVersion().Return(nil, errors.New("version fetch error")).Once()
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mfs := &mockFilesystem.MockFileSystem{}
			mhc := &mockDownload.MockHTTPClient{}
			mec := &mockExec.MockCommandExecutor{}
			mvf := &mockDownload.MockVersionFetcher{}

			tt.setupMocks(mfs, mhc, mec, mvf)

			downloader := download.NewDownloader(mfs, mhc, mec, mvf)
			d := NewDownloadServiceImpl(downloader)

			got, got1, err := d.GetLatest(tt.tempDir)
			if tt.wantErr {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want1, got1)
		})
	}
}

func TestDownloadServiceImpl_GetLatestVersionInfo(t *testing.T) {
	t.Parallel()

	d := &DownloadServiceImpl{}

	got, err := d.GetLatestVersionInfo()
	if err != nil {
		assert.Error(t, err)
		assert.Equal(t, types.VersionInfo{}, got)
	} else {
		assert.NoError(t, err)
		assert.IsType(t, types.VersionInfo{}, got)
		assert.NotEmpty(t, got.Version)
	}
}

func TestVersionServiceImpl_Compare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		installedVersion string
		latestVersion    string
		want             int
		wantErr          bool
	}{
		{
			name:             "installed older than latest",
			installedVersion: "go1.20.0",
			latestVersion:    "go1.21.0",
			want:             -1,
			wantErr:          false,
		},
		{
			name:             "installed newer than latest",
			installedVersion: "go1.22.0",
			latestVersion:    "go1.21.0",
			want:             1,
			wantErr:          false,
		},
		{
			name:             "versions equal",
			installedVersion: "go1.21.0",
			latestVersion:    "go1.21.0",
			want:             0,
			wantErr:          false,
		},
		{
			name:             "invalid installed version",
			installedVersion: "",
			latestVersion:    "go1.21.0",
			wantErr:          true,
		},
		{
			name:             "invalid latest version",
			installedVersion: "go1.21.0",
			latestVersion:    "",
			wantErr:          true,
		},
		{
			name:             "invalid version format",
			installedVersion: "invalid",
			latestVersion:    "go1.21.0",
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v := &VersionServiceImpl{}

			got, err := v.Compare(tt.installedVersion, tt.latestVersion)
			if tt.wantErr {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_compare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		version1 string
		version2 string
		want     int
		wantErr  bool
	}{
		{
			name:     "version1 older than version2",
			version1: "go1.20.0",
			version2: "go1.21.0",
			want:     -1,
			wantErr:  false,
		},
		{
			name:     "version1 newer than version2",
			version1: "go1.22.0",
			version2: "go1.21.0",
			want:     1,
			wantErr:  false,
		},
		{
			name:     "versions equal",
			version1: "go1.21.0",
			version2: "go1.21.0",
			want:     0,
			wantErr:  false,
		},
		{
			name:     "empty version1",
			version1: "",
			version2: "go1.21.0",
			wantErr:  true,
		},
		{
			name:     "empty version2",
			version1: "go1.21.0",
			version2: "",
			wantErr:  true,
		},
		{
			name:     "invalid version1",
			version1: "invalid",
			version2: "go1.21.0",
			wantErr:  true,
		},
		{
			name:     "invalid version2",
			version1: "go1.21.0",
			version2: "invalid",
			wantErr:  true,
		},
		{
			name:     "version without go prefix",
			version1: "1.20.0",
			version2: "1.21.0",
			want:     -1,
			wantErr:  false,
		},
		{
			name:     "version with v prefix",
			version1: "v1.20.0",
			version2: "v1.21.0",
			want:     -1,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := compare(tt.version1, tt.version2)
			if tt.wantErr {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
