package install

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/nicholas-fedor/goUpdater/internal/archive"
	mockArchive "github.com/nicholas-fedor/goUpdater/internal/archive/mocks"
	mockFilesystem "github.com/nicholas-fedor/goUpdater/internal/filesystem/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockFileInfo implements os.FileInfo for testing.
type mockFileInfo struct {
	name string
	mode os.FileMode
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return m.mode.IsDir() }
func (m *mockFileInfo) Sys() any           { return nil }

// mockFileWriter implements io.ReadWriteCloser for testing file operations.
type mockFileWriter struct {
	mock.Mock
}

type mockFileWriter_Expecter struct {
	mock *mock.Mock
}

func (m *mockFileWriter) EXPECT() *mockFileWriter_Expecter {
	return &mockFileWriter_Expecter{mock: &m.Mock}
}

func (e *mockFileWriter_Expecter) Write(p interface{}) *mock.Call {
	return e.mock.On("Write", p)
}

func (e *mockFileWriter_Expecter) Close() *mock.Call {
	return e.mock.On("Close")
}

func (m *mockFileWriter) Read(_ []byte) (int, error) {
	return 0, nil
}

func (m *mockFileWriter) Write(p []byte) (int, error) {
	args := m.Called(p)

	return args.Int(0), args.Error(1)
}

func (m *mockFileWriter) Close() error {
	args := m.Called()

	return args.Error(0)
}

// mockGzipReader implements io.ReadCloser for testing gzip readers.
type mockGzipReader struct {
	reader *bytes.Reader
}

func (m *mockGzipReader) Read(p []byte) (int, error) {
	n, err := m.reader.Read(p)

	return n, err
}

func (m *mockGzipReader) Close() error {
	return nil
}

func TestArchiveServiceImpl_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		archivePath string
		destDir     string
		wantErr     bool
	}{
		{
			name:        "successful validation",
			archivePath: "valid.tar.gz",
			destDir:     "/tmp/dest",
			wantErr:     false,
		},
		{
			name:        "validation failure",
			archivePath: "",
			destDir:     "/tmp/dest",
			wantErr:     true,
		},
		{
			name:        "archive file not found",
			archivePath: "invalid.tar.gz",
			destDir:     "/tmp/dest",
			wantErr:     true,
		},
		{
			name:        "empty destination directory",
			archivePath: "archive.tar.gz",
			destDir:     "",
			wantErr:     true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockFS := mockFilesystem.NewMockFileSystem(t)
			mockProcessor := mockArchive.NewMockProcessor(t)

			// Setup mocks for extractor creation
			// Only set up Stat expectations when both archivePath and destDir are non-empty,
			// as that's when the method actually calls Stat
			if testCase.archivePath != "" && testCase.destDir != "" {
				mockFS.EXPECT().Stat(testCase.destDir).
					Return(&mockFileInfo{name: testCase.destDir, mode: 0755 | os.ModeDir}, nil).Once()

				if testCase.name == "archive file not found" {
					mockFS.EXPECT().Stat(testCase.archivePath).
						Return(nil, fmt.Errorf("%w", os.ErrNotExist)).Once()
				} else {
					mockFS.EXPECT().Stat(testCase.archivePath).
						Return(&mockFileInfo{name: testCase.archivePath, mode: 0644}, nil).Once()
				}
			}

			extractor := archive.NewExtractor(mockFS, mockProcessor)

			a := &ArchiveServiceImpl{
				extractor: extractor,
			}

			err := a.Validate(testCase.archivePath, testCase.destDir)

			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArchiveServiceImpl_Extract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		archivePath string
		destDir     string
		wantErr     bool
	}{
		{
			name:        "successful extraction",
			archivePath: "archive.tar.gz",
			destDir:     "/tmp/dest",
			wantErr:     false,
		},
		{
			name:        "extraction failure",
			archivePath: "corrupt.tar.gz",
			destDir:     "/tmp/dest",
			wantErr:     true,
		},
		{
			name:        "extraction with different archive formats",
			archivePath: "archive.zip",
			destDir:     "/tmp/dest",
			wantErr:     true,
		},
		{
			name:        "extraction to non-existent directory",
			archivePath: "archive.tar.gz",
			destDir:     "/non/existent",
			wantErr:     true,
		},
		{
			name:        "extraction with empty archive path",
			archivePath: "",
			destDir:     "/tmp/dest",
			wantErr:     true,
		},
		{
			name:        "extraction with empty destination directory",
			archivePath: "archive.tar.gz",
			destDir:     "",
			wantErr:     true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockFS := mockFilesystem.NewMockFileSystem(t)
			mockProcessor := mockArchive.NewMockProcessor(t)

			// Setup mocks for extractor creation
			if testCase.destDir != "" {
				if testCase.name == "extraction to non-existent directory" {
					mockFS.EXPECT().Stat(testCase.destDir).
						Return(nil, fmt.Errorf("%w", os.ErrNotExist)).Once()
				} else {
					mockFS.EXPECT().Stat(testCase.destDir).
						Return(&mockFileInfo{name: testCase.destDir, mode: 0755 | os.ModeDir}, nil).Once()
				}
			}

			if testCase.archivePath != "" && testCase.destDir != "" && testCase.name != "extraction to non-existent directory" {
				mockFS.EXPECT().Stat(testCase.archivePath).
					Return(&mockFileInfo{name: testCase.archivePath, mode: 0644}, nil).Once()
			}

			// Special case for empty archive path: after cleaning, archivePath becomes "."
			// and Validate calls Stat(".") which should return a directory info
			if testCase.name == "extraction with empty archive path" {
				mockFS.EXPECT().Stat(".").
					Return(&mockFileInfo{name: ".", mode: 0755 | os.ModeDir}, nil).Once()
			}

			// Special case for empty destination directory: after cleaning, destDir becomes "."
			// and Validate calls Stat(".") which should return an error to fail validation
			if testCase.name == "extraction with empty destination directory" {
				mockFS.EXPECT().Stat(".").
					Return(nil, errors.New("destination directory is empty")).Once()
			}

			// Setup mocks for extraction
			if !testCase.wantErr {
				// For successful extraction, mock the Open call with a valid reader
				reader, writer, err := os.Pipe()
				require.NoError(t, err)
				t.Cleanup(func() { reader.Close(); writer.Close() })

				// Write some mock gzip data
				gzipData := []byte{0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03,
					0x4b, 0x4c, 0x4a, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
				writer.Write(gzipData)
				writer.Close()

				mockFS.EXPECT().Open(testCase.archivePath).Return(reader, nil).Once()

				// Mock processor methods
				mockGzipReader := &mockGzipReader{reader: bytes.NewReader(gzipData)}
				mockProcessor.EXPECT().NewGzipReader(reader).Return(mockGzipReader, nil).Once()

				mockTarReader := &mockArchive.MockTarReader{}
				mockProcessor.EXPECT().NewTarReader(mockGzipReader).Return(mockTarReader).Once()

				// Mock tar reader to return a file header, then EOF
				fileHeader := &tar.Header{
					Name: "test.txt",
					Mode: 0644,
					Size: 12, // "Hello World!" is 12 bytes
				}
				mockTarReader.EXPECT().Next().Return(fileHeader, nil).Once()
				mockTarReader.EXPECT().Next().Return((*tar.Header)(nil), io.EOF).Once()

				// Mock Read for the file content
				fileContent := []byte("Hello World!")
				mockTarReader.EXPECT().Read(mock.Anything).RunAndReturn(func(b []byte) (int, error) {
					return copy(b, fileContent), nil
				}).Once()
				mockTarReader.EXPECT().Read(mock.Anything).Return(0, io.EOF).Maybe()

				// Mock filesystem operations for file extraction
				mockFile := &mockFileWriter{}

				mockFS.EXPECT().MkdirAll(mock.Anything, mock.Anything).Return(nil).Maybe()
				mockFS.EXPECT().OpenFile(mock.Anything, mock.Anything, mock.Anything).Return(mockFile, nil).Maybe()
				mockFile.EXPECT().Write(mock.AnythingOfType("[]uint8")).Return(12, nil).Maybe()
				mockFile.EXPECT().Close().Return(nil).Maybe()
				mockFS.EXPECT().Chmod(mock.Anything, mock.Anything).Return(nil).Maybe()
			} else if testCase.name == "extraction failure" {
				// For extraction failure, mock the Open call to return an error
				mockFS.EXPECT().Open(testCase.archivePath).Return(nil, errors.New("mock open error")).Once()
			} else if testCase.name == "extraction with different archive formats" {
				// For .zip file, mock the Open call with zip data that will fail at gzip level
				reader, writer, err := os.Pipe()
				require.NoError(t, err)
				t.Cleanup(func() { reader.Close(); writer.Close() })

				// Write some mock zip header data
				zipData := []byte{0x50, 0x4b, 0x03, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
				writer.Write(zipData)
				writer.Close()

				mockFS.EXPECT().Open(testCase.archivePath).Return(reader, nil).Once()

				// Mock NewGzipReader to fail for zip data
				mockProcessor.EXPECT().NewGzipReader(reader).Return(nil, errors.New("gzip: invalid header")).Once()
			}

			extractor := archive.NewExtractor(mockFS, mockProcessor)

			a := &ArchiveServiceImpl{
				extractor: extractor,
			}

			err := a.Extract(testCase.archivePath, testCase.destDir)

			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArchiveServiceImpl_ExtractVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		archivePath string
		expected    string
	}{
		{
			name:        "go1.21.0 linux amd64",
			archivePath: "go1.21.0.linux-amd64.tar.gz",
			expected:    "go1.21.0",
		},
		{
			name:        "go1.20.0 darwin amd64",
			archivePath: "/path/to/go1.20.0.darwin-amd64.tar.gz",
			expected:    "go1.20.0",
		},
		{
			name:        "go1.21.0-rc1 linux amd64",
			archivePath: "go1.21.0-rc1.linux-amd64.tar.gz",
			expected:    "go1.21.0",
		},
		{
			name:        "invalid filename",
			archivePath: "invalid-filename",
			expected:    "invalid-filename",
		},
		{
			name:        "empty string",
			archivePath: "",
			expected:    "",
		},
		{
			name:        "filename without go prefix",
			archivePath: "some-other-archive.tar.gz",
			expected:    "some-other-archive.tar.gz",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			a := &ArchiveServiceImpl{}
			result := a.ExtractVersion(testCase.archivePath)
			assert.Equal(t, testCase.expected, result)
		})
	}
}
