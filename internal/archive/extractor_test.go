// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package archive

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	mockFilesystem "github.com/nicholas-fedor/goUpdater/internal/filesystem/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	errGzipError = errors.New("gzip error")
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
func (m *mockFileInfo) Sys() interface{}   { return nil }

// mockFile implements io.ReadWriteCloser for testing.
type mockFile struct {
	mock.Mock
}

func (m *mockFile) Read(p []byte) (int, error) {
	args := m.Called(p)

	return args.Int(0), args.Error(1)
}

func (m *mockFile) Write(p []byte) (int, error) {
	args := m.Called(p)

	return args.Int(0), args.Error(1)
}

func (m *mockFile) Close() error {
	args := m.Called()

	return args.Error(0) //nolint:wrapcheck
}

func (m *mockFile) EXPECT() *mockFileExpecter {
	return &mockFileExpecter{mock: &m.Mock}
}

type mockFileExpecter struct {
	mock *mock.Mock
}

func (e *mockFileExpecter) Read(p interface{}) *mock.Call {
	return e.mock.On("Read", p)
}

func (e *mockFileExpecter) Write(p interface{}) *mock.Call {
	return e.mock.On("Write", p)
}

func (e *mockFileExpecter) Close() *mock.Call {
	return e.mock.On("Close")
}

// mockProcessor implements Processor for testing.
type mockProcessor struct {
	mock.Mock
}

func (m *mockProcessor) EXPECT() *mockProcessorExpecter {
	return &mockProcessorExpecter{mock: &m.Mock}
}

type mockProcessorExpecter struct {
	mock *mock.Mock
}

func (e *mockProcessorExpecter) NewGzipReader(r interface{}) *mock.Call {
	return e.mock.On("NewGzipReader", r)
}

func (e *mockProcessorExpecter) NewTarReader(r interface{}) *mock.Call {
	return e.mock.On("NewTarReader", r)
}

func (m *mockProcessor) NewGzipReader(r io.Reader) (io.ReadCloser, error) {
	args := m.Called(r)

	err := args.Error(1)
	if args.Get(0) == nil {
		return nil, err //nolint:wrapcheck
	}

	if reader, ok := args.Get(0).(io.ReadCloser); ok {
		return reader, err //nolint:wrapcheck
	}

	return nil, errors.New("mock setup error") //nolint:err113
}

func (m *mockProcessor) NewTarReader(r io.Reader) TarReader {
	args := m.Called(r)

	if args.Get(0) == nil {
		return nil
	}

	if reader, ok := args.Get(0).(TarReader); ok {
		return reader
	}

	return nil
}

// mockTarReader implements TarReader for testing.
type mockTarReader struct {
	mock.Mock
}

func (m *mockTarReader) EXPECT() *mockTarReaderExpecter {
	return &mockTarReaderExpecter{mock: &m.Mock}
}

type mockTarReaderExpecter struct {
	mock *mock.Mock
}

func (e *mockTarReaderExpecter) Next() *mock.Call {
	return e.mock.On("Next")
}

func (e *mockTarReaderExpecter) Read(p interface{}) *mock.Call {
	return e.mock.On("Read", p)
}

func (m *mockTarReader) Next() (*tar.Header, error) {
	args := m.Called()

	err := args.Error(1)
	if args.Get(0) == nil {
		if err != nil {
			return nil, err //nolint:wrapcheck // mock returns raw error as configured
		}

		return nil, nil //nolint:nilnil // valid for tar.Reader.Next() EOF case
	}

	if header, ok := args.Get(0).(*tar.Header); ok {
		if err != nil {
			return header, err //nolint:wrapcheck // mock returns raw error as configured
		}

		return header, nil
	}

	return nil, errors.New("mock setup error") //nolint:err113
}

func (m *mockTarReader) Read(p []byte) (int, error) {
	args := m.Called(p)

	return args.Int(0), args.Error(1)
}

// mockGzipReader implements io.ReadCloser for testing.

type mockGzipReader struct {
	mock.Mock
}

func (m *mockGzipReader) EXPECT() *mockGzipReaderExpecter {
	return &mockGzipReaderExpecter{mock: &m.Mock}
}

type mockGzipReaderExpecter struct {
	mock *mock.Mock
}

func (e *mockGzipReaderExpecter) Read(p interface{}) *mock.Call {
	return e.mock.On("Read", p)
}

func (e *mockGzipReaderExpecter) Close() *mock.Call {
	return e.mock.On("Close")
}

func (m *mockGzipReader) Read(p []byte) (int, error) {
	args := m.Called(p)

	return args.Int(0), args.Error(1)
}

func (m *mockGzipReader) Close() error {
	args := m.Called()

	return args.Error(0) //nolint:wrapcheck
}

func TestNewExtractor(t *testing.T) {
	t.Parallel()

	mockFS := &mockFilesystem.MockFileSystem{}
	mockProcessor := &mockProcessor{}

	extractor := NewExtractor(mockFS, mockProcessor)

	assert.NotNil(t, extractor)
	assert.Equal(t, mockFS, extractor.fs)
	assert.Equal(t, mockProcessor, extractor.processor)
}

func TestExtractor_Extract(t *testing.T) { //nolint:maintidx
	t.Parallel()

	tests := []struct {
		name        string
		setupMocks  func(*mockFilesystem.MockFileSystem, *mockProcessor, *mockTarReader)
		wantErr     bool
		expectedErr error
		isWrapped   bool
	}{
		{
			name: "successful extraction",
			setupMocks: func(filesystem *mockFilesystem.MockFileSystem, processor *mockProcessor, _ *mockTarReader) {
				filesystem.EXPECT().Stat("archive.tar.gz").Return(&mockFileInfo{name: "archive.tar.gz", mode: 0644}, nil)

				mockFile := &mockFile{}
				filesystem.EXPECT().Open("archive.tar.gz").Return(mockFile, nil)
				mockFile.EXPECT().Close().Return(nil)
				processor.EXPECT().NewGzipReader(mockFile).Return(nil, errGzipError)
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp/dest",
				Context:     "creating gzip reader",
				Err:         errGzipError,
			},
			isWrapped: true,
		},
		{
			name: "validation failure",
			setupMocks: func(filesystem *mockFilesystem.MockFileSystem, _ *mockProcessor, _ *mockTarReader) {
				filesystem.EXPECT().Stat("archive.tar.gz").Return(nil, os.ErrNotExist)
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp/dest",
				Context:     "validating archive",
				Err:         &ValidationError{FilePath: "archive.tar.gz", Criteria: "file existence", Err: os.ErrNotExist},
			},
			isWrapped: true,
		},
		{
			name: "file open failure",
			setupMocks: func(filesystem *mockFilesystem.MockFileSystem, _ *mockProcessor, _ *mockTarReader) {
				filesystem.EXPECT().Stat("archive.tar.gz").Return(&mockFileInfo{name: "archive.tar.gz", mode: 0644}, nil)
				filesystem.EXPECT().Open("archive.tar.gz").Return(nil, os.ErrPermission)
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp/dest",
				Context:     "opening archive file",
				Err:         os.ErrPermission,
			},
			isWrapped: true,
		},
		{
			name: "gzip reader creation failure",
			setupMocks: func(filesystem *mockFilesystem.MockFileSystem, processor *mockProcessor, _ *mockTarReader) {
				filesystem.EXPECT().Stat("archive.tar.gz").Return(&mockFileInfo{name: "archive.tar.gz", mode: 0644}, nil)

				mockFile := &mockFile{}
				filesystem.EXPECT().Open("archive.tar.gz").Return(mockFile, nil)
				mockFile.EXPECT().Close().Return(nil)
				processor.EXPECT().NewGzipReader(mockFile).Return(nil, errGzipError)
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp/dest",
				Context:     "creating gzip reader",
				Err:         errGzipError,
			},
		},
		{
			name: "tar header read failure",
			setupMocks: func(filesystem *mockFilesystem.MockFileSystem, processor *mockProcessor, tarReader *mockTarReader) {
				filesystem.EXPECT().Stat("archive.tar.gz").Return(&mockFileInfo{name: "archive.tar.gz", mode: 0644}, nil)

				mockFile := &mockFile{}
				filesystem.EXPECT().Open("archive.tar.gz").Return(mockFile, nil)
				mockFile.EXPECT().Close().Return(nil)

				gzipReader := &mockGzipReader{}
				processor.EXPECT().NewGzipReader(mockFile).Return(gzipReader, nil)
				gzipReader.EXPECT().Close().Return(nil)
				processor.EXPECT().NewTarReader(gzipReader).Return(tarReader)
				tarReader.EXPECT().Next().Return(nil, errGzipError)
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp/dest",
				Context:     "reading tar header",
				Err:         errGzipError,
			},
		},
		{
			name: "too many files",
			setupMocks: func(filesystem *mockFilesystem.MockFileSystem, processor *mockProcessor, tarReader *mockTarReader) {
				filesystem.EXPECT().Stat("archive.tar.gz").Return(&mockFileInfo{name: "archive.tar.gz", mode: 0644}, nil)

				archiveFile := &mockFile{}
				filesystem.EXPECT().Open("archive.tar.gz").Return(archiveFile, nil)
				archiveFile.EXPECT().Close().Return(nil)

				gzipReader := &mockGzipReader{}
				processor.EXPECT().NewGzipReader(archiveFile).Return(gzipReader, nil)
				gzipReader.EXPECT().Close().Return(nil)
				processor.EXPECT().NewTarReader(gzipReader).Return(tarReader)

				for fileIndex := range 6 {
					tarReader.EXPECT().Next().Return(&tar.Header{
						Name:     fmt.Sprintf("file%d", fileIndex),
						Typeflag: tar.TypeReg,
						Mode:     0644,
					}, nil)
				}

				for fileIndex := range 5 {
					filePath := fmt.Sprintf("/tmp/dest/file%d", fileIndex)
					filesystem.EXPECT().EvalSymlinks(filePath).Return(filePath, nil)
					filesystem.EXPECT().MkdirAll("/tmp/dest", os.FileMode(0755)).Return(nil)

					mockFile2 := &mockFile{}
					filesystem.EXPECT().OpenFile(filePath,
						os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(0644)).Return(mockFile2, nil)
					mockFile2.EXPECT().Close().Return(nil)
					filesystem.EXPECT().Chmod(filePath, os.FileMode(0644)).Return(nil)
				}

				tarReader.EXPECT().Read(mock.Anything).Return(0, io.EOF).Times(5)
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp/dest",
				Context:     "validating file count",
				Err:         fmt.Errorf("archive contains too many files: %w", ErrTooManyFiles),
			},
			isWrapped: true,
		},
		{
			name: "file too large",
			setupMocks: func(filesystem *mockFilesystem.MockFileSystem, processor *mockProcessor, tarReader *mockTarReader) {
				filesystem.EXPECT().Stat("archive.tar.gz").Return(&mockFileInfo{name: "archive.tar.gz", mode: 0644}, nil)

				mockFile := &mockFile{}
				filesystem.EXPECT().Open("archive.tar.gz").Return(mockFile, nil)
				mockFile.EXPECT().Close().Return(nil)

				gzipReader := &mockGzipReader{}
				processor.EXPECT().NewGzipReader(mockFile).Return(gzipReader, nil)
				gzipReader.EXPECT().Close().Return(nil)
				processor.EXPECT().NewTarReader(gzipReader).Return(tarReader)
				tarReader.EXPECT().Next().Return(&tar.Header{
					Name: "largefile", Size: 60 * 1024 * 1024, Typeflag: tar.TypeReg,
				}, nil).Once()
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp/dest",
				Context:     "validating file size",
				Err:         fmt.Errorf("archive contains file too large: largefile (62914560 bytes): %w", ErrFileTooLarge),
			},
			isWrapped: true,
		},
		{
			name: "total size too large",
			setupMocks: func(filesystem *mockFilesystem.MockFileSystem, processor *mockProcessor, tarReader *mockTarReader) {
				filesystem.EXPECT().Stat("archive.tar.gz").Return(&mockFileInfo{name: "archive.tar.gz", mode: 0644}, nil)

				archiveFile := &mockFile{}
				filesystem.EXPECT().Open("archive.tar.gz").Return(archiveFile, nil)
				archiveFile.EXPECT().Close().Return(nil)

				gzipReader := &mockGzipReader{}
				processor.EXPECT().NewGzipReader(archiveFile).Return(gzipReader, nil)
				gzipReader.EXPECT().Close().Return(nil)
				processor.EXPECT().NewTarReader(gzipReader).Return(tarReader)
				tarReader.EXPECT().Next().Return(&tar.Header{
					Name: "file", Size: 1 * 1024 * 1024, Typeflag: tar.TypeReg, Mode: 0644,
				}, nil).Times(11)
				filesystem.EXPECT().EvalSymlinks("/tmp/dest/file").Return("/tmp/dest/file", nil).Times(11)
				filesystem.EXPECT().MkdirAll("/tmp/dest", os.FileMode(0755)).Return(nil).Times(10)

				fileMock := &mockFile{}
				filesystem.EXPECT().OpenFile("/tmp/dest/file",
					os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(0644)).Return(fileMock, nil).Times(10)
				tarReader.EXPECT().Read(mock.Anything).Return(1*1024*1024, nil).Times(11)
				tarReader.EXPECT().Read(mock.Anything).Return(0, io.EOF).Times(11)
				fileMock.EXPECT().Write(mock.Anything).Return(1*1024*1024, nil).Times(11)
				fileMock.EXPECT().Close().Return(nil).Times(10)
				filesystem.EXPECT().Chmod("/tmp/dest/file", os.FileMode(0644)).Return(nil).Times(10)
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp/dest",
				Context:     "validating total size",
				Err:         fmt.Errorf("archive total size too large: 11534336 bytes: %w", ErrFileTooLarge),
			},
			isWrapped: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockFS := &mockFilesystem.MockFileSystem{}
			mockProcessor := &mockProcessor{}
			mockTarReader := &mockTarReader{}

			testCase.setupMocks(mockFS, mockProcessor, mockTarReader)

			extractor := &Extractor{
				fs:           mockFS,
				processor:    mockProcessor,
				maxFiles:     5,
				maxTotalSize: 10 * 1024 * 1024, // 10MB for testing
				maxFileSize:  1 * 1024 * 1024,  // 1MB per file for testing
				bufferSize:   1 * 1024 * 1024,  // 1MB buffer for testing
			}

			// Set maxFiles high enough for tests that need to hit other limits
			if testCase.name == "file too large" || testCase.name == "total size too large" {
				extractor.maxFiles = 100
			}

			err := extractor.Extract("archive.tar.gz", "/tmp/dest")

			if !testCase.wantErr {
				require.NoError(t, err)

				return
			}

			require.Error(t, err)

			if testCase.expectedErr == nil {
				return
			}

			if testCase.isWrapped {
				var extractionErr *ExtractionError
				require.ErrorAs(t, err, &extractionErr)
				require.Equal(t, testCase.expectedErr, extractionErr)
			} else {
				require.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestExtractor_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		archivePath string
		setupMocks  func(*mockFilesystem.MockFileSystem)
		wantErr     bool
		expectedErr error
		isWrapped   bool
	}{
		{
			name:        "valid regular file",
			archivePath: "archive.tar.gz",
			setupMocks: func(fs *mockFilesystem.MockFileSystem) {
				fs.EXPECT().Stat("archive.tar.gz").Return(&mockFileInfo{name: "archive.tar.gz", mode: 0644}, nil)
			},
			wantErr:   false,
			isWrapped: false,
		},
		{
			name:        "file does not exist",
			archivePath: "nonexistent.tar.gz",
			setupMocks: func(fs *mockFilesystem.MockFileSystem) {
				fs.EXPECT().Stat("nonexistent.tar.gz").Return(nil, os.ErrNotExist)
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "nonexistent.tar.gz",
				Destination: "/tmp/dest",
				Context:     "validating archive",
				Err:         &ValidationError{FilePath: "nonexistent.tar.gz", Criteria: "file existence", Err: os.ErrNotExist},
			},
			isWrapped: true,
		},
		{
			name:        "path is directory",
			archivePath: "directory",
			setupMocks: func(fs *mockFilesystem.MockFileSystem) {
				fs.EXPECT().Stat("directory").Return(&mockFileInfo{name: "directory", mode: os.ModeDir}, nil)
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "directory",
				Destination: "/tmp/dest",
				Context:     "validating archive",
				Err:         &ValidationError{FilePath: "directory", Criteria: "regular file type", Err: ErrArchiveNotRegular},
			},
			isWrapped: true,
		},
		{
			name:        "stat error",
			archivePath: "error.tar.gz",
			setupMocks: func(fs *mockFilesystem.MockFileSystem) {
				fs.EXPECT().Stat("error.tar.gz").Return(nil, os.ErrPermission)
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "error.tar.gz",
				Destination: "/tmp/dest",
				Context:     "validating archive",
				Err:         &ValidationError{FilePath: "error.tar.gz", Criteria: "file existence", Err: os.ErrPermission},
			},
			isWrapped: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockFS := &mockFilesystem.MockFileSystem{}
			mockProcessor := &mockProcessor{}

			testCase.setupMocks(mockFS)

			extractor := &Extractor{
				fs:           mockFS,
				processor:    mockProcessor,
				maxFiles:     5,
				maxTotalSize: 10 * 1024 * 1024, // 10MB for testing
				maxFileSize:  1 * 1024 * 1024,  // 1MB per file for testing
				bufferSize:   1 * 1024 * 1024,  // 1MB buffer for testing
			}

			err := extractor.Validate(testCase.archivePath, "/tmp/dest")

			if !testCase.wantErr {
				require.NoError(t, err)

				return
			}

			require.Error(t, err)

			if testCase.expectedErr == nil {
				return
			}

			if testCase.isWrapped {
				var extractionErr *ExtractionError
				require.ErrorAs(t, err, &extractionErr)
				require.Equal(t, testCase.expectedErr, extractionErr)
			} else {
				require.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestExtractor_processTarEntry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		header      *tar.Header
		destDir     string
		setupMocks  func(*mockFilesystem.MockFileSystem, *mockTarReader)
		wantErr     bool
		expectedErr error
		isWrapped   bool
	}{
		{
			name:    "invalid header name with absolute path",
			header:  &tar.Header{Name: "/absolute/path"},
			destDir: "/tmp/dest",
			setupMocks: func(_ *mockFilesystem.MockFileSystem, _ *mockTarReader) {
				// No mocks needed for header validation failure
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp/dest",
				Context:     "validating header name",
				Err: &SecurityError{
					AttemptedPath: "/absolute/path",
					Validation:    "absolute path prevention",
					Err:           ErrInvalidPath,
				},
			},
			isWrapped: true,
		},
		{
			name:    "invalid header name with parent directory",
			header:  &tar.Header{Name: "../escape"},
			destDir: "/tmp/dest",
			setupMocks: func(_ *mockFilesystem.MockFileSystem, _ *mockTarReader) {
				// No mocks needed for header validation failure
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp/dest",
				Context:     "validating header name",
				Err: &SecurityError{
					AttemptedPath: "../escape",
					Validation:    "parent directory reference prevention",
					Err:           ErrInvalidPath,
				},
			},
			isWrapped: true,
		},
		{
			name:    "invalid header name with backslash",
			header:  &tar.Header{Name: "file\\with\\backslash"},
			destDir: "/tmp/dest",
			setupMocks: func(_ *mockFilesystem.MockFileSystem, _ *mockTarReader) {
				// No mocks needed for header validation failure
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp/dest",
				Context:     "validating header name",
				Err: &SecurityError{
					AttemptedPath: "file\\with\\backslash",
					Validation:    "backslash prevention",
					Err:           ErrInvalidPath,
				},
			},
			isWrapped: true,
		},
		{
			name:    "invalid header name with null byte",
			header:  &tar.Header{Name: "file\x00null"},
			destDir: "/tmp/dest",
			setupMocks: func(_ *mockFilesystem.MockFileSystem, _ *mockTarReader) {
				// No mocks needed for header validation failure
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp/dest",
				Context:     "validating header name",
				Err:         &SecurityError{AttemptedPath: "file\x00null", Validation: "null byte prevention", Err: ErrInvalidPath},
			},
			isWrapped: true,
		},
		{
			name:    "path traversal outside destination",
			header:  &tar.Header{Name: "../../../outside"},
			destDir: "/tmp/dest",
			setupMocks: func(_ *mockFilesystem.MockFileSystem, _ *mockTarReader) {
				// No mocks needed for header validation failure
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp/dest",
				Context:     "validating header name",
				Err: &SecurityError{
					AttemptedPath: "../../../outside",
					Validation:    "parent directory reference prevention",
					Err:           ErrInvalidPath,
				},
			},
			isWrapped: true,
		},
		{
			name:    "successful directory extraction",
			header:  &tar.Header{Name: "dir/", Typeflag: tar.TypeDir, Mode: 0755},
			destDir: "/tmp/dest",
			setupMocks: func(filesystem *mockFilesystem.MockFileSystem, _ *mockTarReader) {
				filesystem.EXPECT().EvalSymlinks("/tmp/dest/dir").Return("/tmp/dest/dir", nil)
				filesystem.EXPECT().MkdirAll("/tmp/dest/dir", os.FileMode(0755)).Return(nil)
				filesystem.EXPECT().Chmod("/tmp/dest/dir", os.FileMode(0755)).Return(nil)
			},
			wantErr:   false,
			isWrapped: false,
		},
		{
			name:    "successful regular file extraction",
			header:  &tar.Header{Name: "file.txt", Typeflag: tar.TypeReg, Mode: 0644, Size: 100},
			destDir: "/tmp/dest",
			setupMocks: func(filesystem *mockFilesystem.MockFileSystem, tarReader *mockTarReader) {
				filesystem.EXPECT().EvalSymlinks("/tmp/dest/file.txt").Return("/tmp/dest/file.txt", nil)
				filesystem.EXPECT().MkdirAll("/tmp/dest", os.FileMode(0755)).Return(nil)

				mockFile := &mockFile{}
				filesystem.EXPECT().OpenFile("/tmp/dest/file.txt",
					os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(0644)).Return(mockFile, nil)
				tarReader.EXPECT().Read(mock.Anything).Return(100, nil).Once()
				tarReader.EXPECT().Read(mock.Anything).Return(0, io.EOF).Once()
				mockFile.EXPECT().Write(mock.Anything).Return(100, nil).Once()
				mockFile.EXPECT().Close().Return(nil)
				filesystem.EXPECT().Chmod("/tmp/dest/file.txt", os.FileMode(0644)).Return(nil)
			},
			wantErr:   false,
			isWrapped: false,
		},
		{
			name:    "successful symlink extraction",
			header:  &tar.Header{Name: "link", Typeflag: tar.TypeSymlink, Mode: 0777, Linkname: "target"},
			destDir: "/tmp/dest",
			setupMocks: func(filesystem *mockFilesystem.MockFileSystem, _ *mockTarReader) {
				filesystem.EXPECT().EvalSymlinks("/tmp/dest/link").Return("/tmp/dest/link", nil)
				filesystem.EXPECT().Lstat("/tmp/dest/target").Return(&mockFileInfo{name: "target", mode: 0644}, nil)
				filesystem.EXPECT().EvalSymlinks("/tmp/dest/target").Return("/tmp/dest/target", nil)
				filesystem.EXPECT().Symlink("target", "/tmp/dest/link").Return(nil)
			},
			wantErr:   false,
			isWrapped: false,
		},
		{
			name:    "successful hard link extraction",
			header:  &tar.Header{Name: "hardlink", Typeflag: tar.TypeLink, Mode: 0777, Linkname: "target"},
			destDir: "/tmp/dest",
			setupMocks: func(filesystem *mockFilesystem.MockFileSystem, _ *mockTarReader) {
				filesystem.EXPECT().EvalSymlinks("/tmp/dest/hardlink").Return("/tmp/dest/hardlink", nil)
				filesystem.EXPECT().Lstat("/tmp/dest/target").Return(&mockFileInfo{name: "target", mode: 0644}, nil)
				filesystem.EXPECT().EvalSymlinks("/tmp/dest/target").Return("/tmp/dest/target", nil)
				filesystem.EXPECT().Link("target", "/tmp/dest/hardlink").Return(nil)
			},
			wantErr:   false,
			isWrapped: false,
		},
		{
			name:    "unsupported entry type",
			header:  &tar.Header{Name: "device", Typeflag: tar.TypeChar, Mode: 0777},
			destDir: "/tmp/dest",
			setupMocks: func(filesystem *mockFilesystem.MockFileSystem, _ *mockTarReader) {
				filesystem.EXPECT().EvalSymlinks("/tmp/dest/device").Return("/tmp/dest/device", nil)
				// No extraction for unsupported types
			},
			wantErr:   false,
			isWrapped: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockFS := &mockFilesystem.MockFileSystem{}
			mockProcessor := &mockProcessor{}
			mockTarReader := &mockTarReader{}

			testCase.setupMocks(mockFS, mockTarReader)

			extractor := &Extractor{
				fs:           mockFS,
				processor:    mockProcessor,
				maxFiles:     5,
				maxTotalSize: 10 * 1024 * 1024, // 10MB for testing
				maxFileSize:  1 * 1024 * 1024,  // 1MB per file for testing
				bufferSize:   1 * 1024 * 1024,  // 1MB buffer for testing
			}

			err := extractor.processTarEntry(mockTarReader, testCase.header, testCase.destDir, "archive.tar.gz")

			if !testCase.wantErr {
				require.NoError(t, err)

				return
			}

			require.Error(t, err)

			if testCase.expectedErr == nil {
				return
			}

			if testCase.isWrapped {
				var extractionErr *ExtractionError
				require.ErrorAs(t, err, &extractionErr)
				require.Equal(t, testCase.expectedErr, extractionErr)
			} else {
				require.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestExtractor_extractDirectory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		targetPath  string
		mode        os.FileMode
		setupMocks  func(*mockFilesystem.MockFileSystem)
		wantErr     bool
		expectedErr error
		isWrapped   bool
	}{
		{
			name:       "successful directory creation",
			targetPath: "/tmp/testdir",
			mode:       0755,
			setupMocks: func(fs *mockFilesystem.MockFileSystem) {
				fs.EXPECT().MkdirAll("/tmp/testdir", os.FileMode(0755)).Return(nil)
				fs.EXPECT().Chmod("/tmp/testdir", os.FileMode(0755)).Return(nil)
			},
			wantErr:   false,
			isWrapped: false,
		},
		{
			name:       "mkdirall failure",
			targetPath: "/tmp/testdir",
			mode:       0755,
			setupMocks: func(fs *mockFilesystem.MockFileSystem) {
				fs.EXPECT().MkdirAll("/tmp/testdir", os.FileMode(0755)).Return(os.ErrPermission)
			},
			wantErr:     true,
			expectedErr: fmt.Errorf("mkdirall failed: %w", os.ErrPermission),
		},
		{
			name:       "chmod failure",
			targetPath: "/tmp/testdir",
			mode:       0755,
			setupMocks: func(fs *mockFilesystem.MockFileSystem) {
				fs.EXPECT().MkdirAll("/tmp/testdir", os.FileMode(0755)).Return(nil)
				fs.EXPECT().Chmod("/tmp/testdir", os.FileMode(0755)).Return(os.ErrPermission)
			},
			wantErr:     true,
			expectedErr: fmt.Errorf("chmod failed: %w", os.ErrPermission),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockFS := &mockFilesystem.MockFileSystem{}
			mockProcessor := &mockProcessor{}

			testCase.setupMocks(mockFS)

			extractor := &Extractor{
				fs:           mockFS,
				processor:    mockProcessor,
				maxFiles:     5,
				maxTotalSize: 10 * 1024 * 1024, // 10MB for testing
				maxFileSize:  1 * 1024 * 1024,  // 1MB per file for testing
				bufferSize:   1 * 1024 * 1024,  // 1MB buffer for testing
			}

			// Set maxFiles high enough for tests that need to hit other limits
			if testCase.name == "file too large" || testCase.name == "total size too large" {
				extractor.maxFiles = 100
			}

			err := extractor.extractDirectory(testCase.targetPath, testCase.mode)

			if testCase.wantErr {
				require.Error(t, err)

				if testCase.expectedErr != nil {
					require.Equal(t, testCase.expectedErr, err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestExtractor_extractRegularFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		tarReader   *mockTarReader
		targetPath  string
		mode        os.FileMode
		setupMocks  func(*mockFilesystem.MockFileSystem, *mockTarReader)
		wantErr     bool
		expectedErr error
		isWrapped   bool
	}{
		{
			name:       "successful file extraction",
			targetPath: "/tmp/testfile.txt",
			mode:       0644,
			setupMocks: func(filesystem *mockFilesystem.MockFileSystem, tarReader *mockTarReader) {
				filesystem.EXPECT().MkdirAll("/tmp", os.FileMode(0755)).Return(nil)

				mockFile := &mockFile{}
				filesystem.EXPECT().OpenFile("/tmp/testfile.txt",
					os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(0644)).Return(mockFile, nil)
				tarReader.EXPECT().Read(mock.Anything).Return(100, nil).Once()
				tarReader.EXPECT().Read(mock.Anything).Return(0, io.EOF).Once()
				mockFile.EXPECT().Write(mock.Anything).Return(100, nil).Once()
				mockFile.EXPECT().Close().Return(nil)
				filesystem.EXPECT().Chmod("/tmp/testfile.txt", os.FileMode(0644)).Return(nil)
			},
			wantErr:   false,
			isWrapped: false,
		},
		{
			name:       "mkdirall failure",
			targetPath: "/tmp/testfile.txt",
			mode:       0644,
			setupMocks: func(filesystem *mockFilesystem.MockFileSystem, _ *mockTarReader) {
				filesystem.EXPECT().MkdirAll("/tmp", os.FileMode(0755)).Return(os.ErrPermission)
			},
			wantErr:     true,
			expectedErr: fmt.Errorf("mkdirall failed: %w", os.ErrPermission),
		},
		{
			name:       "open file failure",
			targetPath: "/tmp/testfile.txt",
			mode:       0644,
			setupMocks: func(filesystem *mockFilesystem.MockFileSystem, _ *mockTarReader) {
				filesystem.EXPECT().MkdirAll("/tmp", os.FileMode(0755)).Return(nil)
				filesystem.EXPECT().OpenFile("/tmp/testfile.txt",
					os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(0644)).Return(nil, os.ErrPermission)
			},
			wantErr:     true,
			expectedErr: fmt.Errorf("open file failed: %w", os.ErrPermission),
		},
		{
			name:       "read failure",
			targetPath: "/tmp/testfile.txt",
			mode:       0644,
			setupMocks: func(filesystem *mockFilesystem.MockFileSystem, tarReader *mockTarReader) {
				filesystem.EXPECT().MkdirAll("/tmp", os.FileMode(0755)).Return(nil)

				mockFile := &mockFile{}
				filesystem.EXPECT().OpenFile("/tmp/testfile.txt",
					os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(0644)).Return(mockFile, nil)
				tarReader.EXPECT().Read(mock.Anything).Return(0, io.ErrUnexpectedEOF).Once()
				mockFile.EXPECT().Close().Return(nil)
			},
			wantErr:     true,
			expectedErr: fmt.Errorf("copy buffer failed: %w", io.ErrUnexpectedEOF),
		},
		{
			name:       "close failure",
			targetPath: "/tmp/testfile.txt",
			mode:       0644,
			setupMocks: func(filesystem *mockFilesystem.MockFileSystem, tarReader *mockTarReader) {
				filesystem.EXPECT().MkdirAll("/tmp", os.FileMode(0755)).Return(nil)

				mockFile := &mockFile{}
				filesystem.EXPECT().OpenFile("/tmp/testfile.txt",
					os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(0644)).Return(mockFile, nil)
				tarReader.EXPECT().Read(mock.Anything).Return(100, nil).Once()
				tarReader.EXPECT().Read(mock.Anything).Return(0, io.EOF).Once()
				mockFile.EXPECT().Write(mock.Anything).Return(100, nil).Once()
				mockFile.EXPECT().Close().Return(os.ErrPermission)
			},
			wantErr:     true,
			expectedErr: fmt.Errorf("close failed: %w", os.ErrPermission),
		},
		{
			name:       "chmod failure",
			targetPath: "/tmp/testfile.txt",
			mode:       0644,
			setupMocks: func(filesystem *mockFilesystem.MockFileSystem, tarReader *mockTarReader) {
				filesystem.EXPECT().MkdirAll("/tmp", os.FileMode(0755)).Return(nil)

				mockFile := &mockFile{}
				filesystem.EXPECT().OpenFile("/tmp/testfile.txt",
					os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(0644)).Return(mockFile, nil)
				tarReader.EXPECT().Read(mock.Anything).Return(100, nil).Once()
				tarReader.EXPECT().Read(mock.Anything).Return(0, io.EOF).Once()
				mockFile.EXPECT().Write(mock.Anything).Return(100, nil).Once()
				mockFile.EXPECT().Close().Return(nil)
				filesystem.EXPECT().Chmod("/tmp/testfile.txt", os.FileMode(0644)).Return(os.ErrPermission)
			},
			wantErr:     true,
			expectedErr: fmt.Errorf("chmod failed: %w", os.ErrPermission),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockFS := &mockFilesystem.MockFileSystem{}
			mockProcessor := &mockProcessor{}
			mockTarReader := &mockTarReader{}

			testCase.setupMocks(mockFS, mockTarReader)

			extractor := &Extractor{
				fs:           mockFS,
				processor:    mockProcessor,
				maxFiles:     5,
				maxTotalSize: 10 * 1024 * 1024, // 10MB for testing
				maxFileSize:  1 * 1024 * 1024,  // 1MB per file for testing
				bufferSize:   1 * 1024 * 1024,  // 1MB buffer for testing
			}

			err := extractor.extractRegularFile(mockTarReader, testCase.targetPath, testCase.mode)

			if testCase.wantErr {
				require.Error(t, err)

				if testCase.expectedErr != nil {
					require.Equal(t, testCase.expectedErr, err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestExtractor_extractSymlink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		targetPath  string
		linkname    string
		baseDir     string
		destDir     string
		setupMocks  func(*mockFilesystem.MockFileSystem)
		wantErr     bool
		expectedErr error
		isWrapped   bool
	}{
		{
			name:       "successful symlink creation",
			targetPath: "/tmp/link",
			linkname:   "target",
			baseDir:    "/tmp",
			destDir:    "/tmp",
			setupMocks: func(filesystem *mockFilesystem.MockFileSystem) {
				filesystem.EXPECT().Lstat("/tmp/target").Return(&mockFileInfo{name: "target", mode: 0644}, nil)
				filesystem.EXPECT().EvalSymlinks("/tmp/target").Return("/tmp/target", nil)
				filesystem.EXPECT().Symlink("target", "/tmp/link").Return(nil)
			},
			wantErr:   false,
			isWrapped: false,
		},
		{
			name:       "linkname validation failure - absolute path",
			targetPath: "/tmp/link",
			linkname:   "/absolute/path",
			baseDir:    "/tmp",
			destDir:    "/tmp",
			setupMocks: func(_ *mockFilesystem.MockFileSystem) {
				// No mocks needed for validation failure
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp",
				Context:     "validating symlink target",
				Err: &SecurityError{
					AttemptedPath: "/absolute/path",
					Validation:    "absolute path prevention",
					Err:           ErrInvalidPath,
				},
			},
		},
		{
			name:       "linkname validation failure - parent directory",
			targetPath: "/tmp/link",
			linkname:   "../escape",
			baseDir:    "/tmp",
			destDir:    "/tmp",
			setupMocks: func(_ *mockFilesystem.MockFileSystem) {
				// No mocks needed for validation failure
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp",
				Context:     "validating symlink target",
				Err: &SecurityError{
					AttemptedPath: "../escape",
					Validation:    "parent directory reference prevention",
					Err:           ErrInvalidPath,
				},
			},
		},
		{
			name:       "linkname validation failure - backslash",
			targetPath: "/tmp/link",
			linkname:   "file\\with\\backslash",
			baseDir:    "/tmp",
			destDir:    "/tmp",
			setupMocks: func(_ *mockFilesystem.MockFileSystem) {
				// No mocks needed for validation failure
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp",
				Context:     "validating symlink target",
				Err: &SecurityError{
					AttemptedPath: "file\\with\\backslash",
					Validation:    "backslash prevention",
					Err:           ErrInvalidPath,
				},
			},
		},
		{
			name:       "linkname validation failure - null byte",
			targetPath: "/tmp/link",
			linkname:   "file\x00null",
			baseDir:    "/tmp",
			destDir:    "/tmp",
			setupMocks: func(_ *mockFilesystem.MockFileSystem) {
				// No mocks needed for validation failure
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp",
				Context:     "validating symlink target",
				Err: &SecurityError{
					AttemptedPath: "file\x00null",
					Validation:    "null byte prevention",
					Err:           ErrInvalidPath,
				},
			},
		},
		{
			name:       "linkname validation failure - outside destination",
			targetPath: "/tmp/link",
			linkname:   "../../../outside",
			baseDir:    "/tmp",
			destDir:    "/tmp",
			setupMocks: func(_ *mockFilesystem.MockFileSystem) {
				// No mocks needed for validation failure
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp",
				Context:     "validating symlink target",
				Err: &SecurityError{
					AttemptedPath: "../../../outside",
					Validation:    "parent directory reference prevention",
					Err:           ErrInvalidPath,
				},
			},
		},
		{
			name:       "linkname validation failure - sensitive path",
			targetPath: "/tmp/link",
			linkname:   "etc/passwd",
			baseDir:    "/",
			destDir:    "/tmp",
			setupMocks: func(_ *mockFilesystem.MockFileSystem) {
				// No mocks needed for validation failure
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp",
				Context:     "validating symlink target",
				Err: &SecurityError{
					AttemptedPath: "etc/passwd",
					Validation:    "linkname destination check",
					Err:           ErrInvalidPath,
				},
			},
			isWrapped: true,
		},
		{
			name:       "symlink chain validation failure",
			targetPath: "/tmp/link",
			linkname:   "symlink_target",
			baseDir:    "/tmp",
			destDir:    "/tmp",
			setupMocks: func(fs *mockFilesystem.MockFileSystem) {
				fs.EXPECT().Lstat("/tmp/symlink_target").Return(&mockFileInfo{name: "symlink_target", mode: os.ModeSymlink}, nil)
				fs.EXPECT().EvalSymlinks("/tmp/symlink_target").Return("/outside/destination", nil)
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp",
				Context:     "validating symlink target",
				Err: &SecurityError{
					AttemptedPath: "symlink_target",
					Validation:    "symlink chain validation",
					Err: &SecurityError{
						AttemptedPath: "/tmp/symlink_target",
						Validation:    "symlink chain destination check",
						Err:           ErrInvalidPath,
					},
				},
			},
			isWrapped: true,
		},
		{
			name:       "symlink creation failure",
			targetPath: "/tmp/link",
			linkname:   "target",
			baseDir:    "/tmp",
			destDir:    "/tmp",
			setupMocks: func(fs *mockFilesystem.MockFileSystem) {
				fs.EXPECT().Lstat("/tmp/target").Return(&mockFileInfo{name: "target", mode: 0644}, nil)
				fs.EXPECT().EvalSymlinks("/tmp/target").Return("/tmp/target", nil)
				fs.EXPECT().Symlink("target", "/tmp/link").Return(os.ErrPermission)
			},
			wantErr:     true,
			expectedErr: fmt.Errorf("symlink failed: %w", os.ErrPermission),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockFS := &mockFilesystem.MockFileSystem{}
			mockProcessor := &mockProcessor{}

			testCase.setupMocks(mockFS)

			extractor := &Extractor{
				fs:           mockFS,
				processor:    mockProcessor,
				maxFiles:     5,
				maxTotalSize: 10 * 1024 * 1024, // 10MB for testing
				maxFileSize:  1 * 1024 * 1024,  // 1MB per file for testing
				bufferSize:   1 * 1024 * 1024,  // 1MB buffer for testing
			}

			err := extractor.extractSymlink(testCase.targetPath, testCase.linkname,
				testCase.baseDir, testCase.destDir, "archive.tar.gz")

			if testCase.wantErr {
				require.Error(t, err)

				if testCase.expectedErr != nil {
					require.Equal(t, testCase.expectedErr, err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestExtractor_extractHardLink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		targetPath  string
		linkname    string
		baseDir     string
		destDir     string
		setupMocks  func(*mockFilesystem.MockFileSystem)
		wantErr     bool
		expectedErr error
		isWrapped   bool
	}{
		{
			name:       "successful hard link creation",
			targetPath: "/tmp/hardlink",
			linkname:   "target",
			baseDir:    "/tmp",
			destDir:    "/tmp",
			setupMocks: func(fs *mockFilesystem.MockFileSystem) {
				fs.EXPECT().Lstat("/tmp/target").Return(&mockFileInfo{name: "target", mode: 0644}, nil)
				fs.EXPECT().EvalSymlinks("/tmp/target").Return("/tmp/target", nil)
				fs.EXPECT().Link("target", "/tmp/hardlink").Return(nil)
			},
			wantErr:   false,
			isWrapped: false,
		},
		{
			name:       "linkname validation failure - absolute path",
			targetPath: "/tmp/hardlink",
			linkname:   "/absolute/path",
			baseDir:    "/tmp",
			destDir:    "/tmp",
			setupMocks: func(_ *mockFilesystem.MockFileSystem) {
				// No mocks needed for validation failure
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp",
				Context:     "validating linkname for hard link",
				Err: &SecurityError{
					AttemptedPath: "/absolute/path",
					Validation:    "absolute path prevention",
					Err:           ErrInvalidPath,
				},
			},
			isWrapped: true,
		},
		{
			name:       "linkname validation failure - parent directory",
			targetPath: "/tmp/hardlink",
			linkname:   "../escape",
			baseDir:    "/tmp",
			destDir:    "/tmp",
			setupMocks: func(_ *mockFilesystem.MockFileSystem) {
				// No mocks needed for validation failure
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp",
				Context:     "validating linkname for hard link",
				Err: &SecurityError{
					AttemptedPath: "../escape",
					Validation:    "parent directory reference prevention",
					Err:           ErrInvalidPath,
				},
			},
			isWrapped: true,
		},
		{
			name:       "linkname validation failure - backslash",
			targetPath: "/tmp/hardlink",
			linkname:   "file\\with\\backslash",
			baseDir:    "/tmp",
			destDir:    "/tmp",
			setupMocks: func(_ *mockFilesystem.MockFileSystem) {
				// No mocks needed for validation failure
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp",
				Context:     "validating linkname for hard link",
				Err: &SecurityError{
					AttemptedPath: "file\\with\\backslash",
					Validation:    "backslash prevention",
					Err:           ErrInvalidPath,
				},
			},
			isWrapped: true,
		},
		{
			name:       "linkname validation failure - null byte",
			targetPath: "/tmp/hardlink",
			linkname:   "file\x00null",
			baseDir:    "/tmp",
			destDir:    "/tmp",
			setupMocks: func(_ *mockFilesystem.MockFileSystem) {
				// No mocks needed for validation failure
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp",
				Context:     "validating linkname for hard link",
				Err: &SecurityError{
					AttemptedPath: "file\x00null",
					Validation:    "null byte prevention",
					Err:           ErrInvalidPath,
				},
			},
			isWrapped: true,
		},
		{
			name:       "linkname validation failure - outside destination",
			targetPath: "/tmp/hardlink",
			linkname:   "../../../outside",
			baseDir:    "/tmp",
			destDir:    "/tmp",
			setupMocks: func(_ *mockFilesystem.MockFileSystem) {
				// No mocks needed for validation failure
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp",
				Context:     "validating linkname for hard link",
				Err: &SecurityError{
					AttemptedPath: "../../../outside",
					Validation:    "parent directory reference prevention",
					Err:           ErrInvalidPath,
				},
			},
			isWrapped: true,
		},
		{
			name:       "linkname validation failure - sensitive path",
			targetPath: "/tmp/hardlink",
			linkname:   "etc/passwd",
			baseDir:    "/",
			destDir:    "/tmp",
			setupMocks: func(_ *mockFilesystem.MockFileSystem) {
				// No mocks needed for validation failure
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp",
				Context:     "validating linkname for hard link",
				Err: &SecurityError{
					AttemptedPath: "etc/passwd",
					Validation:    "linkname destination check",
					Err:           ErrInvalidPath,
				},
			},
			isWrapped: true,
		},
		{
			name:       "symlink chain validation failure",
			targetPath: "/tmp/hardlink",
			linkname:   "symlink_target",
			baseDir:    "/tmp",
			destDir:    "/tmp",
			setupMocks: func(fs *mockFilesystem.MockFileSystem) {
				fs.EXPECT().Lstat("/tmp/symlink_target").Return(&mockFileInfo{name: "symlink_target", mode: os.ModeSymlink}, nil)
				fs.EXPECT().EvalSymlinks("/tmp/symlink_target").Return("/outside/destination", nil)
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp",
				Context:     "validating linkname for hard link",
				Err: &SecurityError{
					AttemptedPath: "symlink_target",
					Validation:    "symlink chain validation",
					Err: &SecurityError{
						AttemptedPath: "/tmp/symlink_target",
						Validation:    "symlink chain destination check",
						Err:           ErrInvalidPath,
					},
				},
			},
			isWrapped: true,
		},
		{
			name:       "hard link creation failure",
			targetPath: "/tmp/hardlink",
			linkname:   "target",
			baseDir:    "/tmp",
			destDir:    "/tmp",
			setupMocks: func(fs *mockFilesystem.MockFileSystem) {
				fs.EXPECT().Lstat("/tmp/target").Return(&mockFileInfo{name: "target", mode: 0644}, nil)
				fs.EXPECT().EvalSymlinks("/tmp/target").Return("/tmp/target", nil)
				fs.EXPECT().Link("target", "/tmp/hardlink").Return(os.ErrPermission)
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp",
				Context:     "extracting hard link",
				Err:         os.ErrPermission,
			},
			isWrapped: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockFS := &mockFilesystem.MockFileSystem{}
			mockProcessor := &mockProcessor{}

			testCase.setupMocks(mockFS)

			extractor := &Extractor{
				fs:           mockFS,
				processor:    mockProcessor,
				maxFiles:     5,
				maxTotalSize: 10 * 1024 * 1024, // 10MB for testing
				maxFileSize:  1 * 1024 * 1024,  // 1MB per file for testing
				bufferSize:   1 * 1024 * 1024,  // 1MB buffer for testing
			}

			err := extractor.extractHardLink(testCase.targetPath, testCase.linkname, testCase.baseDir,
				testCase.destDir, "archive.tar.gz")

			if !testCase.wantErr {
				require.NoError(t, err)

				return
			}

			require.Error(t, err)

			if testCase.expectedErr == nil {
				return
			}

			if testCase.isWrapped {
				var extractionErr *ExtractionError
				require.ErrorAs(t, err, &extractionErr)
				require.Equal(t, testCase.expectedErr, extractionErr)
			} else {
				require.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestExtractor_extractEntry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		header      *tar.Header
		targetPath  string
		baseDir     string
		destDir     string
		setupMocks  func(*mockFilesystem.MockFileSystem, *mockTarReader)
		wantErr     bool
		expectedErr error
		isWrapped   bool
	}{
		{
			name:       "extract directory",
			header:     &tar.Header{Name: "dir/", Typeflag: tar.TypeDir, Mode: 0755},
			targetPath: "/tmp/dir/",
			baseDir:    "/tmp",
			destDir:    "/tmp",
			setupMocks: func(filesystem *mockFilesystem.MockFileSystem, _ *mockTarReader) {
				filesystem.EXPECT().MkdirAll("/tmp/dir/", os.FileMode(0755)).Return(nil)
				filesystem.EXPECT().Chmod("/tmp/dir/", os.FileMode(0755)).Return(nil)
			},
			wantErr:   false,
			isWrapped: false,
		},
		{
			name:       "extract regular file",
			header:     &tar.Header{Name: "file.txt", Typeflag: tar.TypeReg, Mode: 0644, Size: 100},
			targetPath: "/tmp/file.txt",
			baseDir:    "/tmp",
			destDir:    "/tmp",
			setupMocks: func(filesystem *mockFilesystem.MockFileSystem, tarReader *mockTarReader) {
				filesystem.EXPECT().MkdirAll("/tmp", os.FileMode(0755)).Return(nil)

				mockFile := &mockFile{}
				filesystem.EXPECT().OpenFile("/tmp/file.txt",
					os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(0644)).Return(mockFile, nil)
				tarReader.EXPECT().Read(mock.Anything).Return(100, nil).Once()
				tarReader.EXPECT().Read(mock.Anything).Return(0, io.EOF).Once()
				mockFile.EXPECT().Write(mock.Anything).Return(100, nil).Once()
				mockFile.EXPECT().Close().Return(nil)
				filesystem.EXPECT().Chmod("/tmp/file.txt", os.FileMode(0644)).Return(nil)
			},
			wantErr:   false,
			isWrapped: false,
		},
		{
			name:       "extract symlink",
			header:     &tar.Header{Name: "link", Typeflag: tar.TypeSymlink, Mode: 0777, Linkname: "target"},
			targetPath: "/tmp/link",
			baseDir:    "/tmp",
			destDir:    "/tmp",
			setupMocks: func(filesystem *mockFilesystem.MockFileSystem, _ *mockTarReader) {
				filesystem.EXPECT().Lstat("/tmp/target").Return(&mockFileInfo{name: "target", mode: 0644}, nil)
				filesystem.EXPECT().EvalSymlinks("/tmp/target").Return("/tmp/target", nil)
				filesystem.EXPECT().Symlink("target", "/tmp/link").Return(nil)
			},
			wantErr:   false,
			isWrapped: false,
		},
		{
			name:       "extract hard link",
			header:     &tar.Header{Name: "hardlink", Typeflag: tar.TypeLink, Mode: 0777, Linkname: "target"},
			targetPath: "/tmp/hardlink",
			baseDir:    "/tmp",
			destDir:    "/tmp",
			setupMocks: func(filesystem *mockFilesystem.MockFileSystem, _ *mockTarReader) {
				filesystem.EXPECT().Lstat("/tmp/target").Return(&mockFileInfo{name: "target", mode: 0644}, nil)
				filesystem.EXPECT().EvalSymlinks("/tmp/target").Return("/tmp/target", nil)
				filesystem.EXPECT().Link("target", "/tmp/hardlink").Return(nil)
			},
			wantErr:   false,
			isWrapped: false,
		},
		{
			name:       "unsupported entry type",
			header:     &tar.Header{Name: "device", Typeflag: tar.TypeChar, Mode: 0777},
			targetPath: "/tmp/device",
			baseDir:    "/tmp",
			destDir:    "/tmp",
			setupMocks: func(_ *mockFilesystem.MockFileSystem, _ *mockTarReader) {
				// No extraction for unsupported types
			},
			wantErr:   false,
			isWrapped: false,
		},
		{
			name:       "directory extraction failure",
			header:     &tar.Header{Name: "dir/", Typeflag: tar.TypeDir, Mode: 0755},
			targetPath: "/tmp/dir/",
			baseDir:    "/tmp",
			destDir:    "/tmp",
			setupMocks: func(filesystem *mockFilesystem.MockFileSystem, _ *mockTarReader) {
				filesystem.EXPECT().MkdirAll("/tmp/dir/", os.FileMode(0755)).Return(os.ErrPermission)
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp",
				Context:     "extracting directory",
				Err:         fmt.Errorf("mkdirall failed: %w", os.ErrPermission),
			},
			isWrapped: true,
		},
		{
			name:       "file extraction failure",
			header:     &tar.Header{Name: "file.txt", Typeflag: tar.TypeReg, Mode: 0644, Size: 100},
			targetPath: "/tmp/file.txt",
			baseDir:    "/tmp",
			destDir:    "/tmp",
			setupMocks: func(filesystem *mockFilesystem.MockFileSystem, _ *mockTarReader) {
				filesystem.EXPECT().MkdirAll("/tmp", os.FileMode(0755)).Return(nil)
				filesystem.EXPECT().OpenFile("/tmp/file.txt",
					os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(0644)).Return(nil, os.ErrPermission)
			},
			wantErr: true,
			expectedErr: &ExtractionError{
				ArchivePath: "archive.tar.gz",
				Destination: "/tmp",
				Context:     "extracting file",
				Err:         fmt.Errorf("open file failed: %w", os.ErrPermission),
			},
			isWrapped: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockFS := &mockFilesystem.MockFileSystem{}
			mockProcessor := &mockProcessor{}
			mockTarReader := &mockTarReader{}

			testCase.setupMocks(mockFS, mockTarReader)

			extractor := &Extractor{
				fs:           mockFS,
				processor:    mockProcessor,
				maxFiles:     5,
				maxTotalSize: 10 * 1024 * 1024, // 10MB for testing
				maxFileSize:  1 * 1024 * 1024,  // 1MB per file for testing
				bufferSize:   1 * 1024 * 1024,  // 1MB buffer for testing
			}

			err := extractor.extractEntry(mockTarReader, testCase.header, testCase.targetPath,
				testCase.baseDir, testCase.destDir,
				"archive.tar.gz")

			if !testCase.wantErr {
				require.NoError(t, err)

				return
			}

			require.Error(t, err)

			if testCase.expectedErr == nil {
				return
			}

			if testCase.isWrapped {
				var extractionErr *ExtractionError
				require.ErrorAs(t, err, &extractionErr)
				require.Equal(t, testCase.expectedErr, extractionErr)
			} else {
				require.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestExtractor_validateSymlinkChain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		resolved    string
		destDir     string
		setupMocks  func(*mockFilesystem.MockFileSystem)
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "valid symlink chain within destination",
			resolved: "/tmp/dest/valid",
			destDir:  "/tmp/dest",
			setupMocks: func(fs *mockFilesystem.MockFileSystem) {
				fs.EXPECT().EvalSymlinks("/tmp/dest/valid").Return("/tmp/dest/valid", nil)
			},
			wantErr: false,
		},
		{
			name:     "symlink chain resolves outside destination",
			resolved: "/tmp/dest/link",
			destDir:  "/tmp/dest",
			setupMocks: func(fs *mockFilesystem.MockFileSystem) {
				fs.EXPECT().EvalSymlinks("/tmp/dest/link").Return("/outside/destination", nil)
			},
			wantErr: true,
			expectedErr: &SecurityError{
				AttemptedPath: "/tmp/dest/link",
				Validation:    "symlink chain destination check",
				Err:           ErrInvalidPath,
			},
		},
		{
			name:     "symlink chain resolves to destination root",
			resolved: "/tmp/dest",
			destDir:  "/tmp/dest",
			setupMocks: func(fs *mockFilesystem.MockFileSystem) {
				fs.EXPECT().EvalSymlinks("/tmp/dest").Return("/tmp/dest", nil)
			},
			wantErr: false,
		},
		{
			name:     "eval symlinks failure",
			resolved: "/tmp/dest/link",
			destDir:  "/tmp/dest",
			setupMocks: func(fs *mockFilesystem.MockFileSystem) {
				fs.EXPECT().EvalSymlinks("/tmp/dest/link").Return("", os.ErrNotExist)
			},
			wantErr: true,
			expectedErr: &SecurityError{
				AttemptedPath: "/tmp/dest/link",
				Validation:    "symlink chain resolution",
				Err:           os.ErrNotExist,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockFS := &mockFilesystem.MockFileSystem{}
			mockProcessor := &mockProcessor{}

			testCase.setupMocks(mockFS)

			extractor := &Extractor{
				fs:        mockFS,
				processor: mockProcessor,
			}

			err := extractor.validateSymlinkChain(testCase.resolved, testCase.destDir)

			if testCase.wantErr {
				require.Error(t, err)

				if testCase.expectedErr != nil {
					var securityErr *SecurityError
					require.ErrorAs(t, err, &securityErr)
					require.Equal(t, testCase.expectedErr, securityErr)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestExtractor_validateResolvedPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		targetPath  string
		destDir     string
		setupMocks  func(*mockFilesystem.MockFileSystem)
		wantErr     bool
		expectedErr error
	}{
		{
			name:       "path resolves within destination",
			targetPath: "/tmp/dest/file",
			destDir:    "/tmp/dest",
			setupMocks: func(fs *mockFilesystem.MockFileSystem) {
				fs.EXPECT().EvalSymlinks("/tmp/dest/file").Return("/tmp/dest/file", nil)
			},
			wantErr: false,
		},
		{
			name:       "path resolves outside destination",
			targetPath: "/tmp/dest/link",
			destDir:    "/tmp/dest",
			setupMocks: func(fs *mockFilesystem.MockFileSystem) {
				fs.EXPECT().EvalSymlinks("/tmp/dest/link").Return("/outside/destination", nil)
			},
			wantErr: true,
			expectedErr: &SecurityError{
				AttemptedPath: "/tmp/dest/link",
				Validation:    "symlink chain destination check",
				Err:           ErrInvalidPath,
			},
		},
		{
			name:       "path resolves to destination root",
			targetPath: "/tmp/dest",
			destDir:    "/tmp/dest",
			setupMocks: func(fs *mockFilesystem.MockFileSystem) {
				fs.EXPECT().EvalSymlinks("/tmp/dest").Return("/tmp/dest", nil)
			},
			wantErr: false,
		},
		{
			name:       "eval symlinks failure - path doesn't exist",
			targetPath: "/tmp/dest/nonexistent",
			destDir:    "/tmp/dest",
			setupMocks: func(fs *mockFilesystem.MockFileSystem) {
				fs.EXPECT().EvalSymlinks("/tmp/dest/nonexistent").Return("", os.ErrNotExist)
			},
			wantErr: true,
			expectedErr: &SecurityError{
				AttemptedPath: "/tmp/dest/nonexistent",
				Validation:    "resolved path destination check",
				Err:           os.ErrNotExist,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockFS := &mockFilesystem.MockFileSystem{}
			mockProcessor := &mockProcessor{}

			testCase.setupMocks(mockFS)

			extractor := &Extractor{
				fs:        mockFS,
				processor: mockProcessor,
			}

			err := extractor.validateResolvedPath(testCase.targetPath, testCase.destDir)

			if testCase.wantErr {
				require.Error(t, err)

				if testCase.expectedErr != nil {
					require.Equal(t, testCase.expectedErr, err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestExtractor_validateLinkname(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		linkname    string
		baseDir     string
		destDir     string
		setupMocks  func(*mockFilesystem.MockFileSystem)
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "valid relative linkname",
			linkname: "target",
			baseDir:  "/tmp",
			destDir:  "/tmp",
			setupMocks: func(fs *mockFilesystem.MockFileSystem) {
				fs.EXPECT().Lstat("/tmp/target").Return(&mockFileInfo{name: "target", mode: 0644}, nil)
				fs.EXPECT().EvalSymlinks("/tmp/target").Return("/tmp/target", nil)
			},
			wantErr: false,
		},
		{
			name:     "absolute path linkname",
			linkname: "/absolute/path",
			baseDir:  "/tmp",
			destDir:  "/tmp",
			setupMocks: func(_ *mockFilesystem.MockFileSystem) {
				// No mocks needed for validation failure
			},
			wantErr: true,
			expectedErr: &SecurityError{
				AttemptedPath: "/absolute/path",
				Validation:    "absolute path prevention",
				Err:           ErrInvalidPath,
			},
		},
		{
			name:     "parent directory reference",
			linkname: "../escape",
			baseDir:  "/tmp",
			destDir:  "/tmp",
			setupMocks: func(_ *mockFilesystem.MockFileSystem) {
				// No mocks needed for validation failure
			},
			wantErr: true,
			expectedErr: &SecurityError{
				AttemptedPath: "../escape",
				Validation:    "parent directory reference prevention",
				Err:           ErrInvalidPath,
			},
		},
		{
			name:     "backslash in linkname",
			linkname: "file\\with\\backslash",
			baseDir:  "/tmp",
			destDir:  "/tmp",
			setupMocks: func(_ *mockFilesystem.MockFileSystem) {
				// No mocks needed for validation failure
			},
			wantErr: true,
			expectedErr: &SecurityError{
				AttemptedPath: "file\\with\\backslash",
				Validation:    "backslash prevention",
				Err:           ErrInvalidPath,
			},
		},
		{
			name:     "null byte in linkname",
			linkname: "file\x00null",
			baseDir:  "/tmp",
			destDir:  "/tmp",
			setupMocks: func(_ *mockFilesystem.MockFileSystem) {
				// No mocks needed for validation failure
			},
			wantErr:     true,
			expectedErr: &SecurityError{AttemptedPath: "file\x00null", Validation: "null byte prevention", Err: ErrInvalidPath},
		},
		{
			name:     "linkname resolves outside destination",
			linkname: "../../../outside",
			baseDir:  "/tmp",
			destDir:  "/tmp",
			setupMocks: func(_ *mockFilesystem.MockFileSystem) {
				// No mocks needed for validation failure
			},
			wantErr: true,
			expectedErr: &SecurityError{
				AttemptedPath: "../../../outside",
				Validation:    "parent directory reference prevention",
				Err:           ErrInvalidPath,
			},
		},
		{
			name:     "sensitive system path",
			linkname: "etc/passwd",
			baseDir:  "/",
			destDir:  "/tmp",
			setupMocks: func(_ *mockFilesystem.MockFileSystem) {
				// No mocks needed for validation failure
			},
			wantErr: true,
			expectedErr: &SecurityError{
				AttemptedPath: "etc/passwd",
				Validation:    "linkname destination check",
				Err:           ErrInvalidPath,
			},
		},
		{
			name:     "symlink chain validation",
			linkname: "symlink_target",
			baseDir:  "/tmp",
			destDir:  "/tmp",
			setupMocks: func(fs *mockFilesystem.MockFileSystem) {
				fs.EXPECT().Lstat("/tmp/symlink_target").Return(&mockFileInfo{name: "symlink_target", mode: os.ModeSymlink}, nil)
				fs.EXPECT().EvalSymlinks("/tmp/symlink_target").Return("/outside/destination", nil)
			},
			wantErr: true,
			expectedErr: &SecurityError{
				AttemptedPath: "symlink_target",
				Validation:    "symlink chain validation",
				Err: &SecurityError{
					AttemptedPath: "/tmp/symlink_target",
					Validation:    "symlink chain destination check",
					Err:           ErrInvalidPath,
				},
			},
		},
		{
			name:     "valid symlink chain",
			linkname: "valid_target",
			baseDir:  "/tmp",
			destDir:  "/tmp",
			setupMocks: func(fs *mockFilesystem.MockFileSystem) {
				fs.EXPECT().Lstat("/tmp/valid_target").Return(&mockFileInfo{name: "valid_target", mode: os.ModeSymlink}, nil)
				fs.EXPECT().EvalSymlinks("/tmp/valid_target").Return("/tmp/valid_target", nil)
			},
			wantErr: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockFS := &mockFilesystem.MockFileSystem{}
			mockProcessor := &mockProcessor{}

			testCase.setupMocks(mockFS)

			extractor := &Extractor{
				fs:        mockFS,
				processor: mockProcessor,
			}

			err := extractor.validateLinkname(testCase.linkname, testCase.baseDir, testCase.destDir)

			if testCase.wantErr {
				require.Error(t, err)

				if testCase.expectedErr != nil {
					require.ErrorIs(t, err, testCase.expectedErr)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
