// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package version

import (
	"runtime/debug"
	"testing"

	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	mockFilesystem "github.com/nicholas-fedor/goUpdater/internal/filesystem/mocks"
	"github.com/stretchr/testify/require"
)

func TestGetVersionInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		reader     filesystem.DebugInfoReader
		want       Info
		setupMocks func(*mockFilesystem.MockDebugInfoReader)
	}{
		{
			name:   "returns default info when no ldflags set",
			reader: mockFilesystem.NewMockDebugInfoReader(t),
			want: Info{
				Version:   "dev",
				Commit:    "",
				Date:      "",
				GoVersion: "",
				Platform:  "linux/amd64",
			},
			setupMocks: func(mockReader *mockFilesystem.MockDebugInfoReader) {
				mockReader.EXPECT().ReadBuildInfo().Return(nil, false).Once()
			},
		},
		{
			name:   "populates from build info when available",
			reader: mockFilesystem.NewMockDebugInfoReader(t),
			want: Info{
				Version:   "dev",
				Commit:    "abc123",
				Date:      "2023-01-01",
				GoVersion: "go1.21",
				Platform:  "linux/amd64",
			},
			setupMocks: func(mockReader *mockFilesystem.MockDebugInfoReader) {
				buildInfo := &debug.BuildInfo{
					GoVersion: "go1.21",
					Settings: []debug.BuildSetting{
						{Key: "vcs.revision", Value: "abc123"},
						{Key: "vcs.time", Value: "2023-01-01"},
					},
				}
				mockReader.EXPECT().ReadBuildInfo().Return(buildInfo, true).Once()
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if testCase.setupMocks != nil {
				if mockReader, ok := testCase.reader.(*mockFilesystem.MockDebugInfoReader); ok {
					testCase.setupMocks(mockReader)
				}
			}

			got := GetVersionInfo(testCase.reader)
			require.Equal(t, testCase.want, got)
		})
	}
}

func TestGetClientVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want string
	}{
		{
			name: "returns dev when no version set",
			want: "dev",
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := GetClientVersion()
			require.Equal(t, testCase.want, got)
		})
	}
}

func Test_populateFromBuildInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		info       *Info
		reader     filesystem.DebugInfoReader
		expected   Info
		setupMocks func(*mockFilesystem.MockDebugInfoReader)
	}{
		{
			name: "populates missing fields from build info",
			info: &Info{
				Version:   "1.0.0",
				Commit:    "",
				Date:      "",
				GoVersion: "",
				Platform:  "",
			},
			reader: mockFilesystem.NewMockDebugInfoReader(t),
			expected: Info{
				Version:   "1.0.0",
				Commit:    "abc123",
				Date:      "2023-01-01",
				GoVersion: "go1.21",
				Platform:  "",
			},
			setupMocks: func(mockReader *mockFilesystem.MockDebugInfoReader) {
				buildInfo := &debug.BuildInfo{
					GoVersion: "go1.21",
					Settings: []debug.BuildSetting{
						{Key: "vcs.revision", Value: "abc123"},
						{Key: "vcs.time", Value: "2023-01-01"},
					},
				}
				mockReader.EXPECT().ReadBuildInfo().Return(buildInfo, true).Once()
			},
		},
		{
			name: "does not overwrite existing fields",
			info: &Info{
				Version:   "1.0.0",
				Commit:    "existing",
				Date:      "existing",
				GoVersion: "",
				Platform:  "",
			},
			reader: mockFilesystem.NewMockDebugInfoReader(t),
			expected: Info{
				Version:   "1.0.0",
				Commit:    "existing",
				Date:      "existing",
				GoVersion: "go1.21",
				Platform:  "",
			},
			setupMocks: func(mockReader *mockFilesystem.MockDebugInfoReader) {
				buildInfo := &debug.BuildInfo{
					GoVersion: "go1.21",
					Settings: []debug.BuildSetting{
						{Key: "vcs.revision", Value: "abc123"},
						{Key: "vcs.time", Value: "2023-01-01"},
					},
				}
				mockReader.EXPECT().ReadBuildInfo().Return(buildInfo, true).Once()
			},
		},
		{
			name: "returns early when all fields populated",
			info: &Info{
				Version:   "1.0.0",
				Commit:    "existing",
				Date:      "existing",
				GoVersion: "existing",
				Platform:  "",
			},
			reader: mockFilesystem.NewMockDebugInfoReader(t),
			expected: Info{
				Version:   "1.0.0",
				Commit:    "existing",
				Date:      "existing",
				GoVersion: "existing",
				Platform:  "",
			},
			setupMocks: func(_ *mockFilesystem.MockDebugInfoReader) {
				// Should not be called
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if testCase.setupMocks != nil {
				if mockReader, ok := testCase.reader.(*mockFilesystem.MockDebugInfoReader); ok {
					testCase.setupMocks(mockReader)
				}
			}

			populateFromBuildInfo(testCase.info, testCase.reader)
			require.Equal(t, testCase.expected, *testCase.info)
		})
	}
}
