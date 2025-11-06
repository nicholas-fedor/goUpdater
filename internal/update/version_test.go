package update

import (
	"testing"

	httpclient "github.com/nicholas-fedor/goUpdater/internal/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultVersionFetcher_GetLatestVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func() *DefaultVersionFetcher
		want    *httpclient.GoVersionInfo
		wantErr bool
	}{
		{
			name: "success - returns version info",
			setup: func() *DefaultVersionFetcher {
				return &DefaultVersionFetcher{
					getLatestVersionFunc: func() (*httpclient.GoVersionInfo, error) {
						return &httpclient.GoVersionInfo{
							Version: "go1.21.0",
							Stable:  true,
							Files:   []httpclient.GoFileInfo{},
						}, nil
					},
				}
			},
			want: &httpclient.GoVersionInfo{
				Version: "go1.21.0",
				Stable:  true,
				Files:   []httpclient.GoFileInfo{},
			},
			wantErr: false,
		},
		{
			name: "error - http client fails",
			setup: func() *DefaultVersionFetcher {
				return &DefaultVersionFetcher{
					getLatestVersionFunc: func() (*httpclient.GoVersionInfo, error) {
						return nil, ErrNetworkError
					},
				}
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error - nil function",
			setup: func() *DefaultVersionFetcher {
				return &DefaultVersionFetcher{
					getLatestVersionFunc: nil,
				}
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			d := testCase.setup()

			got, err := d.GetLatestVersion()
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.want, got)
			}
		})
	}
}

func TestDefaultVersionFetcher_GetLatestVersionInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func() *DefaultVersionFetcher
		want    *httpclient.GoVersionInfo
		wantErr bool
	}{
		{
			name: "success - returns version info",
			setup: func() *DefaultVersionFetcher {
				return &DefaultVersionFetcher{
					getLatestVersionFunc: func() (*httpclient.GoVersionInfo, error) {
						return &httpclient.GoVersionInfo{
							Version: "go1.21.0",
							Stable:  true,
							Files:   []httpclient.GoFileInfo{},
						}, nil
					},
				}
			},
			want: &httpclient.GoVersionInfo{
				Version: "go1.21.0",
				Stable:  true,
				Files:   []httpclient.GoFileInfo{},
			},
			wantErr: false,
		},
		{
			name: "error - underlying getLatestVersion fails",
			setup: func() *DefaultVersionFetcher {
				return &DefaultVersionFetcher{
					getLatestVersionFunc: func() (*httpclient.GoVersionInfo, error) {
						return nil, ErrNetworkError
					},
				}
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error - nil function",
			setup: func() *DefaultVersionFetcher {
				return &DefaultVersionFetcher{
					getLatestVersionFunc: nil,
				}
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			d := testCase.setup()

			got, err := d.GetLatestVersionInfo()
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.want, got)
			}
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
			name:     "version1 equals version2",
			version1: "go1.21.0",
			version2: "go1.21.0",
			want:     0,
			wantErr:  false,
		},
		{
			name:     "version1 less than version2",
			version1: "go1.20.0",
			version2: "go1.21.0",
			want:     -1,
			wantErr:  false,
		},
		{
			name:     "version1 greater than version2",
			version1: "go1.22.0",
			version2: "go1.21.0",
			want:     1,
			wantErr:  false,
		},
		{
			name:     "version1 without go prefix equals version2",
			version1: "1.21.0",
			version2: "go1.21.0",
			want:     0,
			wantErr:  false,
		},
		{
			name:     "version1 with v prefix equals version2",
			version1: "v1.21.0",
			version2: "go1.21.0",
			want:     0,
			wantErr:  false,
		},
		{
			name:     "version1 empty",
			version1: "",
			version2: "go1.21.0",
			want:     0,
			wantErr:  true,
		},
		{
			name:     "version2 empty",
			version1: "go1.21.0",
			version2: "",
			want:     0,
			wantErr:  true,
		},
		{
			name:     "version1 invalid semver",
			version1: "invalid",
			version2: "go1.21.0",
			want:     0,
			wantErr:  true,
		},
		{
			name:     "version2 invalid semver",
			version1: "go1.21.0",
			version2: "invalid",
			want:     0,
			wantErr:  true,
		},
		{
			name:     "patch version comparison - less",
			version1: "go1.21.0",
			version2: "go1.21.1",
			want:     -1,
			wantErr:  false,
		},
		{
			name:     "patch version comparison - greater",
			version1: "go1.21.2",
			version2: "go1.21.1",
			want:     1,
			wantErr:  false,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := compare(testCase.version1, testCase.version2)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.want, got)
			}
		})
	}
}

func Test_needsUpdate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		installedVersion string
		latestVersionStr string
		want             bool
	}{
		{
			name:             "no installed version - needs update",
			installedVersion: "",
			latestVersionStr: "go1.21.0",
			want:             true,
		},
		{
			name:             "installed version equals latest - no update needed",
			installedVersion: "go1.21.0",
			latestVersionStr: "go1.21.0",
			want:             false,
		},
		{
			name:             "installed version greater than latest - no update needed",
			installedVersion: "go1.22.0",
			latestVersionStr: "go1.21.0",
			want:             false,
		},
		{
			name:             "installed version less than latest - needs update",
			installedVersion: "go1.20.0",
			latestVersionStr: "go1.21.0",
			want:             true,
		},
		{
			name:             "patch version update needed",
			installedVersion: "go1.21.0",
			latestVersionStr: "go1.21.1",
			want:             true,
		},
		{
			name:             "minor version update needed",
			installedVersion: "go1.20.0",
			latestVersionStr: "go1.21.0",
			want:             true,
		},
		{
			name:             "major version update needed",
			installedVersion: "go1.20.0",
			latestVersionStr: "go2.0.0",
			want:             true,
		},
		{
			name:             "installed version without go prefix",
			installedVersion: "1.21.0",
			latestVersionStr: "go1.21.0",
			want:             false,
		},
		{
			name:             "latest version without go prefix",
			installedVersion: "go1.21.0",
			latestVersionStr: "1.21.0",
			want:             false,
		},
		{
			name:             "invalid installed version - no update",
			installedVersion: "invalid",
			latestVersionStr: "go1.21.0",
			want:             false,
		},
		{
			name:             "invalid latest version - no update",
			installedVersion: "go1.21.0",
			latestVersionStr: "invalid",
			want:             false,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := needsUpdate(testCase.installedVersion, testCase.latestVersionStr)
			assert.Equal(t, testCase.want, got)
		})
	}
}
