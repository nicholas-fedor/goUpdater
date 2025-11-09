// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package install

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstallError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		e    *InstallError
		want string
	}{
		{
			name: "basic error with all fields",
			e: &InstallError{
				Phase:     "extract",
				FilePath:  "/tmp/archive.tar.gz",
				Operation: "extract",
				Err:       ErrExtractArchive,
			},
			want: "install failed at extract phase: operation=extract path=/tmp/archive.tar.gz: failed to extract archive",
		},
		{
			name: "error with empty phase",
			e: &InstallError{
				Phase:     "",
				FilePath:  "/usr/local/go",
				Operation: "create_dir",
				Err:       ErrCreateInstallDir,
			},
			want: "install failed at  phase: operation=create_dir path=/usr/local/go: failed to create installation directory",
		},
		{
			name: "error with empty operation",
			e: &InstallError{
				Phase:     "verify",
				FilePath:  "/usr/local/go/bin/go",
				Operation: "",
				Err:       ErrVerifyInstallation,
			},
			want: "install failed at verify phase: operation= path=/usr/local/go/bin/go: installation verification failed",
		},
		{
			name: "error with empty path",
			e: &InstallError{
				Phase:     "download",
				FilePath:  "",
				Operation: "download",
				Err:       ErrDownloadGo,
			},
			want: "install failed at download phase: operation=download path=: failed to download Go",
		},
		{
			name: "error with nil underlying error",
			e: &InstallError{
				Phase:     "prepare",
				FilePath:  "/tmp",
				Operation: "temp_dir",
				Err:       nil,
			},
			want: "install failed at prepare phase: operation=temp_dir path=/tmp: <nil>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.e.Error()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInstallError_Unwrap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		e    *InstallError
		want error
	}{
		{
			name: "unwrap with underlying error",
			e: &InstallError{
				Phase:     "extract",
				FilePath:  "/tmp/archive.tar.gz",
				Operation: "extract",
				Err:       ErrExtractArchive,
			},
			want: ErrExtractArchive,
		},
		{
			name: "error with permission denied",
			e: &InstallError{
				Phase:     "install",
				FilePath:  "/usr/local/go",
				Operation: "write",
				Err:       ErrPermissionDenied,
			},
			want: ErrPermissionDenied,
		},
		{
			name: "error with network error",
			e: &InstallError{
				Phase:     "download",
				FilePath:  "https://golang.org/dl/go1.21.0.linux-amd64.tar.gz",
				Operation: "http_get",
				Err:       ErrNetworkError,
			},
			want: ErrNetworkError,
		},
		{
			name: "error with invalid archive",
			e: &InstallError{
				Phase:     "extract",
				FilePath:  "/tmp/corrupted.tar.gz",
				Operation: "validate",
				Err:       ErrInvalidArchive,
			},
			want: ErrInvalidArchive,
		},
		{
			name: "error with cleanup failure",
			e: &InstallError{
				Phase:     "cleanup",
				FilePath:  "/tmp/go-install-temp",
				Operation: "remove",
				Err:       ErrCleanupFailed,
			},
			want: ErrCleanupFailed,
		},
		{
			name: "error with path conflict",
			e: &InstallError{
				Phase:     "prepare",
				FilePath:  "/usr/local/go",
				Operation: "check",
				Err:       ErrPathConflict,
			},
			want: ErrPathConflict,
		},
		{
			name: "unwrap with predefined error",
			e: &InstallError{
				Phase:     "verify",
				FilePath:  "/usr/local/go/bin/go",
				Operation: "verify",
				Err:       ErrVerifyInstallation,
			},
			want: ErrVerifyInstallation,
		},
		{
			name: "unwrap with nil error",
			e: &InstallError{
				Phase:     "prepare",
				FilePath:  "/tmp",
				Operation: "temp_dir",
				Err:       nil,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.e.Unwrap()
			assert.Equal(t, tt.want, got)
		})
	}
}
