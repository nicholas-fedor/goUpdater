// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	mockHTTP "github.com/nicholas-fedor/goUpdater/internal/http/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errNetworkError = errors.New("network error")

func TestGetLatestVersion(t *testing.T) {
	t.Parallel()

	got, err := GetLatestVersion()
	require.NoError(t, err)
	assert.NotNil(t, got)
	assert.True(t, got.Stable)
	assert.NotEmpty(t, got.Version)
	assert.NotEmpty(t, got.Files)
}

func Test_getLatestVersionWithClient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupMock func(*mockHTTP.MockClient)
		want      *GoVersionInfo
		wantErr   bool
	}{
		{
			name: "success with stable version",
			setupMock: func(mockClient *mockHTTP.MockClient) {
				versionInfo := []GoVersionInfo{
					{
						Version: "go1.20.0",
						Stable:  false,
						Files:   []GoFileInfo{},
					},
					{
						Version: "go1.21.0",
						Stable:  true,
						Files: []GoFileInfo{
							{
								Filename: "go1.21.0.linux-amd64.tar.gz",
								OS:       "linux",
								Arch:     "amd64",
								Version:  "go1.21.0",
								Sha256:   "hash",
								Size:     100,
								Kind:     "archive",
							},
						},
					},
				}
				jsonData, err := json.Marshal(versionInfo)
				require.NoError(t, err)

				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(jsonData)),
					Header:     make(http.Header),
				}
				mockClient.EXPECT().Do(mock.Anything).Return(resp, nil)
			},
			want: &GoVersionInfo{
				Version: "go1.21.0",
				Stable:  true,
				Files: []GoFileInfo{
					{
						Filename: "go1.21.0.linux-amd64.tar.gz",
						OS:       "linux",
						Arch:     "amd64",
						Version:  "go1.21.0",
						Sha256:   "hash",
						Size:     100,
						Kind:     "archive",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "http client error",
			setupMock: func(mockClient *mockHTTP.MockClient) {
				mockClient.EXPECT().Do(mock.Anything).Return(nil, errNetworkError)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "unexpected status code",
			setupMock: func(mockClient *mockHTTP.MockClient) {
				resp := &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(strings.NewReader("")),
					Header:     make(http.Header),
				}
				mockClient.EXPECT().Do(mock.Anything).Return(resp, nil)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid json response",
			setupMock: func(mockClient *mockHTTP.MockClient) {
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader("invalid json")),
					Header:     make(http.Header),
				}
				mockClient.EXPECT().Do(mock.Anything).Return(resp, nil)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "no stable version found",
			setupMock: func(mockClient *mockHTTP.MockClient) {
				versionInfo := []GoVersionInfo{
					{
						Version: "go1.20.0",
						Stable:  false,
						Files:   []GoFileInfo{},
					},
					{
						Version: "go1.21.0",
						Stable:  false,
						Files:   []GoFileInfo{},
					},
				}
				jsonData, err := json.Marshal(versionInfo)
				require.NoError(t, err)

				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(jsonData)),
					Header:     make(http.Header),
				}
				mockClient.EXPECT().Do(mock.Anything).Return(resp, nil)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "empty version list",
			setupMock: func(mockClient *mockHTTP.MockClient) {
				versionInfo := []GoVersionInfo{}
				jsonData, err := json.Marshal(versionInfo)
				require.NoError(t, err)

				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(jsonData)),
					Header:     make(http.Header),
				}
				mockClient.EXPECT().Do(mock.Anything).Return(resp, nil)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "stable version at beginning",
			setupMock: func(mockClient *mockHTTP.MockClient) {
				versionInfo := []GoVersionInfo{
					{
						Version: "go1.21.0",
						Stable:  true,
						Files: []GoFileInfo{
							{
								Filename: "go1.21.0.linux-amd64.tar.gz",
								OS:       "linux",
								Arch:     "amd64",
								Version:  "go1.21.0",
								Sha256:   "hash",
								Size:     100,
								Kind:     "archive",
							},
						},
					},
					{
						Version: "go1.22.0",
						Stable:  false,
						Files:   []GoFileInfo{},
					},
				}
				jsonData, err := json.Marshal(versionInfo)
				require.NoError(t, err)

				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(jsonData)),
					Header:     make(http.Header),
				}
				mockClient.EXPECT().Do(mock.Anything).Return(resp, nil)
			},
			want: &GoVersionInfo{
				Version: "go1.21.0",
				Stable:  true,
				Files: []GoFileInfo{
					{
						Filename: "go1.21.0.linux-amd64.tar.gz",
						OS:       "linux",
						Arch:     "amd64",
						Version:  "go1.21.0",
						Sha256:   "hash",
						Size:     100,
						Kind:     "archive",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockClient := mockHTTP.NewMockClient(t)
			testCase.setupMock(mockClient)

			got, err := getLatestVersionWithClient(mockClient)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.want, got)
			}
		})
	}
}
