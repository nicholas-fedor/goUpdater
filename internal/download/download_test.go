// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package download

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	mockDownload "github.com/nicholas-fedor/goUpdater/internal/download/mocks"
	mockExec "github.com/nicholas-fedor/goUpdater/internal/exec/mocks"
	mockFilesystem "github.com/nicholas-fedor/goUpdater/internal/filesystem/mocks"
	httpclient "github.com/nicholas-fedor/goUpdater/internal/http"
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

// mockFileInfo implements os.FileInfo for testing.
type mockFileInfo struct {
	mode  os.FileMode
	isDir bool
}

func (m *mockFileInfo) Name() string       { return "mock" }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }

func TestDownload(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupMocks func(
			*mockFilesystem.MockFileSystem,
			*mockDownload.MockHTTPClient,
			*mockExec.MockCommandExecutor,
			*mockDownload.MockVersionFetcher,
		)
		expectedPath   string
		expectedSha256 string
		expectError    bool
	}{
		{
			name: "successful download",
			setupMocks: func(
				mfs *mockFilesystem.MockFileSystem,
				mhc *mockDownload.MockHTTPClient,
				_ *mockExec.MockCommandExecutor,
				mvf *mockDownload.MockVersionFetcher,
			) {
				// Setup version fetcher
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
							Size:     100000000,
							Kind:     "archive",
						},
					},
				}
				mvf.EXPECT().GetLatestVersion().Return(versionInfo, nil).Once()

				// Setup filesystem
				mfs.EXPECT().TempDir().Return("/tmp").Once()
				mfs.EXPECT().UserHomeDir().Return("/home/user", nil).Once()

				// Mock Stat calls for directory readability checks in GetSearchDirectories
				mfs.EXPECT().Stat("/home/user/Downloads").Return(&mockFileInfo{isDir: true}, nil).Once()
				mfs.EXPECT().Stat("/home/user").Return(&mockFileInfo{isDir: true}, nil).Once()

				// No existing archive found
				mfs.EXPECT().
					Stat("/home/user/Downloads/go1.21.0.linux-amd64.tar.gz").
					Return(nil, fmt.Errorf("%w", ErrNotFound)).
					Once()
				mfs.EXPECT().IsNotExist(fmt.Errorf("%w", ErrNotFound)).Return(true).Once()
				mfs.EXPECT().Stat("/home/user/go1.21.0.linux-amd64.tar.gz").Return(nil, fmt.Errorf("%w", ErrNotFound)).Once()
				mfs.EXPECT().IsNotExist(fmt.Errorf("%w", ErrNotFound)).Return(true).Once()
				mfs.EXPECT().Stat("/tmp/go1.21.0.linux-amd64.tar.gz").Return(nil, fmt.Errorf("%w", ErrNotFound)).Once()
				mfs.EXPECT().IsNotExist(fmt.Errorf("%w", ErrNotFound)).Return(true).Once()

				// Download setup - use os.Pipe for valid file handles
				reader, writer, err := os.Pipe()
				if err != nil {
					panic("failed to create pipe")
				}

				mfs.EXPECT().Create("/tmp/go1.21.0.linux-amd64.tar.gz").Return(writer, nil).Once()
				mfs.EXPECT().Open("/tmp/go1.21.0.linux-amd64.tar.gz").Return(reader, nil).Once()

				// Mock HTTP client - return a proper HTTP response with headers
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Security-Policy":   []string{"default-src 'self'"},
						"X-Content-Type-Options":    []string{"nosniff"},
						"X-Frame-Options":           []string{"DENY"},
						"X-XSS-Protection":          []string{"1; mode=block"},
						"Strict-Transport-Security": []string{"max-age=31536000"},
					},
					Body:    &mockReadCloser{bytes.NewReader([]byte("mock response"))},
					Request: &http.Request{URL: &url.URL{Scheme: "https"}},
				}
				mhc.EXPECT().Do(mock.AnythingOfType("*http.Request")).Return(resp, nil).Once()
			},
			expectedPath:   "/tmp/go1.21.0.linux-amd64.tar.gz",
			expectedSha256: "26df8a783491d87eb6bae3f16ae0b588dab6b83150c791ebf9d2406bf94a8999",
			expectError:    false,
		},
		{
			name: "version fetch failure",
			setupMocks: func(
				mfs *mockFilesystem.MockFileSystem,
				_ *mockDownload.MockHTTPClient,
				_ *mockExec.MockCommandExecutor,
				mvf *mockDownload.MockVersionFetcher,
			) {
				mfs.EXPECT().TempDir().Return("/tmp").Once()
				mvf.EXPECT().GetLatestVersion().Return(nil, fmt.Errorf("%w", ErrVersionFetchFailed)).Once()
			},
			expectError: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			mfs := mockFilesystem.NewMockFileSystem(t)
			mhc := mockDownload.NewMockHTTPClient(t)
			mec := mockExec.NewMockCommandExecutor(t)
			mvf := mockDownload.NewMockVersionFetcher(t)

			testCase.setupMocks(mfs, mhc, mec, mvf)

			downloader := NewDownloader(mfs, mhc, mec, mvf)
			path, sha256, err := downloader.GetLatest("")

			if testCase.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.expectedPath, path)
				require.Equal(t, testCase.expectedSha256, sha256)
			}
		})
	}
}
