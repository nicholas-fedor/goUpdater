// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package version

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	mockFilesystem "github.com/nicholas-fedor/goUpdater/internal/filesystem/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewDisplayManager(t *testing.T) {
	t.Parallel()
	mockReader := mockFilesystem.NewMockDebugInfoReader(t)
	mockParser := mockFilesystem.NewMockTimeParser(t)
	mockEncoder := mockFilesystem.NewMockJSONEncoder(t)
	mockWriter := mockFilesystem.NewMockErrorWriter(t)

	tests := []struct {
		name string
		args struct {
			reader  filesystem.DebugInfoReader
			parser  filesystem.TimeParser
			encoder filesystem.JSONEncoder
			writer  filesystem.ErrorWriter
		}
		want *DisplayManager
	}{
		{
			name: "creates display manager with valid dependencies",
			args: struct {
				reader  filesystem.DebugInfoReader
				parser  filesystem.TimeParser
				encoder filesystem.JSONEncoder
				writer  filesystem.ErrorWriter
			}{
				reader:  mockReader,
				parser:  mockParser,
				encoder: mockEncoder,
				writer:  mockWriter,
			},
			want: &DisplayManager{
				reader:  mockReader,
				parser:  mockParser,
				encoder: mockEncoder,
				writer:  mockWriter,
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := NewDisplayManager(testCase.args.reader, testCase.args.parser, testCase.args.encoder, testCase.args.writer)
			assert.Equal(t, testCase.want, got)
		})
	}
}

func TestDisplayManager_DisplayDefault(t *testing.T) {
	t.Parallel()
	mockReader := mockFilesystem.NewMockDebugInfoReader(t)
	mockParser := mockFilesystem.NewMockTimeParser(t)
	mockEncoder := mockFilesystem.NewMockJSONEncoder(t)
	mockWriter := mockFilesystem.NewMockErrorWriter(t)

	displayManager := NewDisplayManager(mockReader, mockParser, mockEncoder, mockWriter)

	tests := []struct {
		name       string
		info       Info
		wantWriter string
	}{
		{
			name: "displays version when present",
			info: Info{
				Version: "1.0.0",
			},
			wantWriter: "goUpdater 1.0.0\n",
		},
		{
			name: "displays default when version empty",
			info: Info{
				Version: "",
			},
			wantWriter: "goUpdater\n",
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			writer := &bytes.Buffer{}
			displayManager.DisplayDefault(writer, testCase.info)
			assert.Equal(t, testCase.wantWriter, writer.String())
		})
	}
}

func TestDisplayManager_DisplayShort(t *testing.T) {
	t.Parallel()
	mockReader := mockFilesystem.NewMockDebugInfoReader(t)
	mockParser := mockFilesystem.NewMockTimeParser(t)
	mockEncoder := mockFilesystem.NewMockJSONEncoder(t)
	mockWriter := mockFilesystem.NewMockErrorWriter(t)

	displayManager := NewDisplayManager(mockReader, mockParser, mockEncoder, mockWriter)

	tests := []struct {
		name       string
		info       Info
		wantWriter string
	}{
		{
			name: "displays version only",
			info: Info{
				Version: "1.0.0",
			},
			wantWriter: "1.0.0\n",
		},
		{
			name: "displays empty version",
			info: Info{
				Version: "",
			},
			wantWriter: "\n",
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			writer := &bytes.Buffer{}
			displayManager.DisplayShort(writer, testCase.info)
			assert.Equal(t, testCase.wantWriter, writer.String())
		})
	}
}

func TestDisplayManager_DisplayVerbose(t *testing.T) {
	t.Parallel()
	mockReader := mockFilesystem.NewMockDebugInfoReader(t)
	mockParser := mockFilesystem.NewMockTimeParser(t)
	mockEncoder := mockFilesystem.NewMockJSONEncoder(t)
	mockWriter := mockFilesystem.NewMockErrorWriter(t)

	displayManager := NewDisplayManager(mockReader, mockParser, mockEncoder, mockWriter)

	tests := []struct {
		name       string
		info       Info
		wantWriter string
	}{
		{
			name: "displays all fields when present",
			info: Info{
				Version:   "1.0.0",
				Commit:    "abc123",
				Date:      "2023-01-01",
				GoVersion: "go1.21",
				Platform:  "linux/amd64",
			},
			wantWriter: "goUpdater\n" +
				"├─ Version: 1.0.0\n" +
				"├─ Commit: abc123\n" +
				"├─ Date: 2023-01-01\n" +
				"├─ Go Version: go1.21\n" +
				"└─ Platform: linux/amd64\n",
		},
		{
			name: "displays only available fields",
			info: Info{
				Version: "1.0.0",
				Commit:  "abc123",
			},
			wantWriter: "goUpdater\n├─ Version: 1.0.0\n├─ Commit: abc123\n",
		},
		{
			name:       "displays empty when no fields",
			info:       Info{},
			wantWriter: "goUpdater\n",
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			writer := &bytes.Buffer{}
			displayManager.DisplayVerbose(writer, testCase.info)
			assert.Equal(t, testCase.wantWriter, writer.String())
		})
	}
}

func TestDisplayManager_DisplayJSON(t *testing.T) {
	t.Parallel()
	mockReader := mockFilesystem.NewMockDebugInfoReader(t)
	mockParser := mockFilesystem.NewMockTimeParser(t)
	mockEncoder := mockFilesystem.NewMockJSONEncoder(t)
	mockWriter := mockFilesystem.NewMockErrorWriter(t)

	displayManager := NewDisplayManager(mockReader, mockParser, mockEncoder, mockWriter)

	tests := []struct {
		name       string
		info       Info
		wantWriter string
		wantErr    bool
		setupMocks func()
	}{
		{
			name: "encodes info successfully",
			info: Info{
				Version:   "1.0.0",
				Commit:    "abc123",
				Date:      "2023-01-01",
				GoVersion: "go1.21",
				Platform:  "linux/amd64",
			},
			wantWriter: `{"version":"1.0.0","commit":"abc123","date":"2023-01-01","goVersion":"go1.21","platform":"linux/amd64"}
`,
			wantErr: false,
			setupMocks: func() {
				mockEncoder.EXPECT().NewEncoder(mock.Anything).RunAndReturn(json.NewEncoder).Once()
			},
		},
		{
			name: "handles encoding error",
			info: Info{
				Version: "1.0.0",
			},
			wantWriter: "",
			wantErr:    true,
			setupMocks: func() {
				// Mock encoder that returns nil to simulate error
				mockEncoder.EXPECT().NewEncoder(mock.Anything).Return((*json.Encoder)(nil)).Once()
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if testCase.setupMocks != nil {
				testCase.setupMocks()
			}

			writer := &bytes.Buffer{}

			err := displayManager.DisplayJSON(writer, testCase.info)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.wantWriter, writer.String())
			}
		})
	}
}
