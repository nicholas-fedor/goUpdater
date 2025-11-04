// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package verify

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nicholas-fedor/goUpdater/internal/exec"
	mockExec "github.com/nicholas-fedor/goUpdater/internal/exec/mocks"
	mockFilesystem "github.com/nicholas-fedor/goUpdater/internal/filesystem/mocks"
)

var (
	errPermissionDeniedTest = errors.New("permission denied")
	errCommandFailedTest    = errors.New("command failed")
	errExecFormatErrorTest  = errors.New("exec format error")
	errStatErrorTest        = errors.New("stat error")
	errUnderlyingErrorTest  = errors.New("underlying error")
)

// mockCommand is a mock implementation of exec.Command for testing.
type mockCommand struct {
	output []byte
	err    error
}

func (m *mockCommand) Output() ([]byte, error) {
	return m.output, m.err
}

func (m *mockCommand) Path() string {
	return ""
}

func (m *mockCommand) Args() []string {
	return []string{}
}

func newMockCmd(output []byte, err error) exec.Command {
	return &mockCommand{
		output: output,
		err:    err,
	}
}

func TestVerifier_GetInstalledVersionCore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		installDir     string
		mockStatErr    error
		mockCmdOutput  []byte
		mockCmdErr     error
		expectedResult string
		expectedError  string
	}{
		{
			name:           "successful version retrieval",
			installDir:     "/usr/local/go",
			mockStatErr:    nil,
			mockCmdOutput:  []byte("go version go1.21.0 linux/amd64\n"),
			mockCmdErr:     nil,
			expectedResult: "go1.21.0",
			expectedError:  "",
		},
		{
			name:           "go binary not found",
			installDir:     "/usr/local/go",
			mockStatErr:    &os.PathError{Op: "stat", Path: "/usr/local/go/bin/go", Err: os.ErrNotExist},
			mockCmdOutput:  nil,
			mockCmdErr:     nil,
			expectedResult: "",
			expectedError:  "",
		},
		{
			name:           "stat fails with other error",
			installDir:     "/usr/local/go",
			mockStatErr:    errPermissionDeniedTest,
			mockCmdOutput:  nil,
			mockCmdErr:     nil,
			expectedResult: "",
			expectedError:  "failed to check Go binary at /usr/local/go/bin/go: permission denied",
		},
		{
			name:           "command execution fails",
			installDir:     "/usr/local/go",
			mockStatErr:    nil,
			mockCmdOutput:  nil,
			mockCmdErr:     errCommandFailedTest,
			expectedResult: "",
			expectedError:  "failed to execute 'go version'",
		},
		{
			name:           "invalid go version output format",
			installDir:     "/usr/local/go",
			mockStatErr:    nil,
			mockCmdOutput:  []byte("invalid output"),
			mockCmdErr:     nil,
			expectedResult: "",
			expectedError:  "unexpected go version output format",
		},
		{
			name:           "go version output with insufficient parts",
			installDir:     "/usr/local/go",
			mockStatErr:    nil,
			mockCmdOutput:  []byte("go version"),
			mockCmdErr:     nil,
			expectedResult: "",
			expectedError:  "unexpected go version output format",
		},
		{
			name:           "go version output without 'go version' prefix",
			installDir:     "/usr/local/go",
			mockStatErr:    nil,
			mockCmdOutput:  []byte("some other output go1.21.0\n"),
			mockCmdErr:     nil,
			expectedResult: "",
			expectedError:  "unexpected go version output format",
		},
		{
			name:           "empty install directory",
			installDir:     "",
			mockStatErr:    &os.PathError{Op: "stat", Path: "bin/go", Err: os.ErrNotExist},
			mockCmdOutput:  nil,
			mockCmdErr:     nil,
			expectedResult: "",
			expectedError:  "",
		},
		{
			name:           "corrupted binary - command returns error",
			installDir:     "/usr/local/go",
			mockStatErr:    nil,
			mockCmdOutput:  nil,
			mockCmdErr:     errExecFormatErrorTest,
			expectedResult: "",
			expectedError:  "failed to execute 'go version'",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			mockFS := mockFilesystem.NewMockFileSystem(t)
			mockExecutor := mockExec.NewMockCommandExecutor(t)

			goBinPath := "/usr/local/go/bin/go" // hardcoded for simplicity, but in real code it's filepath.Join
			if testCase.installDir == "" {
				goBinPath = "bin/go"
			}

			switch {
			case testCase.mockStatErr != nil:
				mockFS.EXPECT().Stat(goBinPath).Return(nil, testCase.mockStatErr).Once()

				pathErr := &os.PathError{}
				if errors.As(testCase.mockStatErr, &pathErr) {
					mockFS.EXPECT().IsNotExist(testCase.mockStatErr).Return(true).Once()
				} else if testCase.mockStatErr.Error() == "permission denied" {
					mockFS.EXPECT().IsNotExist(testCase.mockStatErr).Return(false).Once()
				}
			default:
				mockFS.EXPECT().Stat(goBinPath).Return(nil, nil).Once()
				mockExecutor.EXPECT().CommandContext(context.Background(), goBinPath,
					[]string{"version"}).Return(newMockCmd(testCase.mockCmdOutput, testCase.mockCmdErr)).Once()
			}

			verifier := NewVerifier(mockFS, mockExecutor)
			result, err := verifier.GetInstalledVersionCore(testCase.installDir)

			if testCase.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
			}

			assert.Equal(t, testCase.expectedResult, result)
		})
	}
}

func TestVerifier_Installation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		installDir        string
		expectedVersion   string
		mockVersion       string
		mockVersionErr    error
		expectedError     string
		expectedErrorType interface{}
	}{
		{
			name:            "version matches",
			installDir:      "/usr/local/go",
			expectedVersion: "go1.21.0",
			mockVersion:     "go1.21.0",
			mockVersionErr:  nil,
			expectedError:   "",
		},
		{
			name:              "version does not match",
			installDir:        "/usr/local/go",
			expectedVersion:   "go1.21.0",
			mockVersion:       "go1.20.0",
			mockVersionErr:    nil,
			expectedError:     "verification failed: expected version go1.21.0, got go1.20.0",
			expectedErrorType: &VerificationError{},
		},
		{
			name:              "go not installed",
			installDir:        "/usr/local/go",
			expectedVersion:   "go1.21.0",
			mockVersion:       "",
			mockVersionErr:    nil,
			expectedError:     "verification failed: expected version go1.21.0 at /usr/local/go/bin/go, but Go is not installed",
			expectedErrorType: &VerificationError{},
		},
		{
			name:            "error getting version",
			installDir:      "/usr/local/go",
			expectedVersion: "go1.21.0",
			mockVersion:     "",
			mockVersionErr:  errCommandFailedTest,
			expectedError: "failed to get installed version: failed to check Go binary at " +
				"/usr/local/go/bin/go: command failed",
			expectedErrorType: nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			mockFS := mockFilesystem.NewMockFileSystem(t)
			mockExecutor := mockExec.NewMockCommandExecutor(t)

			// Mock GetInstalledVersionCore
			switch {
			case testCase.mockVersionErr != nil:
				mockFS.EXPECT().Stat("/usr/local/go/bin/go").Return(nil, testCase.mockVersionErr).Once()

				pathErr := &os.PathError{}
				if errors.As(testCase.mockVersionErr, &pathErr) {
					mockFS.EXPECT().IsNotExist(testCase.mockVersionErr).Return(true).Once()
				} else if testCase.mockVersionErr.Error() == "command failed" {
					mockFS.EXPECT().IsNotExist(testCase.mockVersionErr).Return(false).Once()
				}
			case testCase.mockVersion == "":
				mockFS.EXPECT().Stat("/usr/local/go/bin/go").Return(nil,
					&os.PathError{Op: "stat", Path: "/usr/local/go/bin/go", Err: os.ErrNotExist}).Once()
				mockFS.EXPECT().IsNotExist(mock.Anything).Return(true).Once()
			default:
				mockFS.EXPECT().Stat("/usr/local/go/bin/go").Return(nil, nil).Once()
				mockExecutor.EXPECT().CommandContext(context.Background(), "/usr/local/go/bin/go",
					[]string{"version"}).Return(newMockCmd([]byte("go version "+testCase.mockVersion+
					" linux/amd64\n"), nil)).Once()
			}

			verifier := NewVerifier(mockFS, mockExecutor)
			err := verifier.Installation(testCase.installDir, testCase.expectedVersion)

			if testCase.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				if testCase.expectedErrorType != nil {
					assert.ErrorAs(t, err, &testCase.expectedErrorType)
				}
			}
		})
	}
}

func TestVerifier_GetVerificationInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		installDir     string
		mockVersion    string
		mockVersionErr error
		expectedInfo   VerificationInfo
		expectedError  string
	}{
		{
			name:           "successful verification info",
			installDir:     "/usr/local/go",
			mockVersion:    "go1.21.0",
			mockVersionErr: nil,
			expectedInfo: VerificationInfo{
				InstallDir: "/usr/local/go",
				Version:    "go1.21.0",
				Status:     "Verified",
			},
			expectedError: "",
		},
		{
			name:           "go not installed",
			installDir:     "/usr/local/go",
			mockVersion:    "",
			mockVersionErr: nil,
			expectedInfo: VerificationInfo{
				InstallDir: "/usr/local/go",
				Version:    "",
				Status:     "Not installed or not found",
			},
			expectedError: "",
		},
		{
			name:           "error getting version",
			installDir:     "/usr/local/go",
			mockVersion:    "",
			mockVersionErr: errStatErrorTest,
			expectedInfo: VerificationInfo{
				InstallDir: "/usr/local/go",
				Version:    "",
				Status:     "Failed to get version",
			},
			expectedError: "failed to get installed version: failed to check Go binary at /usr/local/go/bin/go: stat error",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			mockFS := mockFilesystem.NewMockFileSystem(t)
			mockExecutor := mockExec.NewMockCommandExecutor(t)

			// Mock GetInstalledVersionCore
			switch {
			case testCase.mockVersionErr != nil:
				mockFS.EXPECT().Stat("/usr/local/go/bin/go").Return(nil, testCase.mockVersionErr).Once()

				pathErr := &os.PathError{}
				if errors.As(testCase.mockVersionErr, &pathErr) {
					mockFS.EXPECT().IsNotExist(testCase.mockVersionErr).Return(true).Once()
				} else if testCase.mockVersionErr.Error() == "stat error" {
					mockFS.EXPECT().IsNotExist(testCase.mockVersionErr).Return(false).Once()
				}
			case testCase.mockVersion == "":
				mockFS.EXPECT().Stat("/usr/local/go/bin/go").Return(nil,
					&os.PathError{Op: "stat", Path: "/usr/local/go/bin/go", Err: os.ErrNotExist}).Once()
				mockFS.EXPECT().IsNotExist(mock.Anything).Return(true).Once()
			default:
				mockFS.EXPECT().Stat("/usr/local/go/bin/go").Return(nil, nil).Once()
				mockExecutor.EXPECT().CommandContext(context.Background(), "/usr/local/go/bin/go",
					[]string{"version"}).Return(newMockCmd([]byte("go version "+testCase.mockVersion+
					" linux/amd64\n"), nil)).Once()
			}

			verifier := NewVerifier(mockFS, mockExecutor)
			info, err := verifier.GetVerificationInfo(testCase.installDir)

			if testCase.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
			}

			assert.Equal(t, testCase.expectedInfo, info)
		})
	}
}

func TestVerifier_GetInstalledVersionWithLogging(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		installDir     string
		mockStatErr    error
		mockCmdOutput  []byte
		mockCmdErr     error
		expectedResult string
		expectedError  string
	}{
		{
			name:           "successful version retrieval",
			installDir:     "/usr/local/go",
			mockStatErr:    nil,
			mockCmdOutput:  []byte("go version go1.21.0 linux/amd64\n"),
			mockCmdErr:     nil,
			expectedResult: "go1.21.0",
			expectedError:  "",
		},
		{
			name:           "go binary not found",
			installDir:     "/usr/local/go",
			mockStatErr:    &os.PathError{Op: "stat", Path: "/usr/local/go/bin/go", Err: os.ErrNotExist},
			mockCmdOutput:  nil,
			mockCmdErr:     nil,
			expectedResult: "",
			expectedError:  "",
		},
		{
			name:           "stat fails with other error",
			installDir:     "/usr/local/go",
			mockStatErr:    errPermissionDeniedTest,
			mockCmdOutput:  nil,
			mockCmdErr:     nil,
			expectedResult: "",
			expectedError:  "failed to check Go binary at /usr/local/go/bin/go: permission denied",
		},
		{
			name:           "command execution fails",
			installDir:     "/usr/local/go",
			mockStatErr:    nil,
			mockCmdOutput:  nil,
			mockCmdErr:     errCommandFailedTest,
			expectedResult: "",
			expectedError:  "failed to execute 'go version' at /usr/local/go/bin/go",
		},
		{
			name:           "invalid go version output format",
			installDir:     "/usr/local/go",
			mockStatErr:    nil,
			mockCmdOutput:  []byte("invalid output"),
			mockCmdErr:     nil,
			expectedResult: "",
			expectedError:  "unexpected go version output format",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			mockFS := mockFilesystem.NewMockFileSystem(t)
			mockExecutor := mockExec.NewMockCommandExecutor(t)

			goBinPath := "/usr/local/go/bin/go"

			switch {
			case testCase.mockStatErr != nil:
				mockFS.EXPECT().Stat(goBinPath).Return(nil, testCase.mockStatErr).Once()

				pathErr := &os.PathError{}
				if errors.As(testCase.mockStatErr, &pathErr) {
					mockFS.EXPECT().IsNotExist(testCase.mockStatErr).Return(true).Once()
				} else if testCase.mockStatErr.Error() == "permission denied" {
					mockFS.EXPECT().IsNotExist(testCase.mockStatErr).Return(false).Once()
				}
			default:
				mockFS.EXPECT().Stat(goBinPath).Return(nil, nil).Once()
				mockExecutor.EXPECT().CommandContext(context.Background(), goBinPath,
					[]string{"version"}).Return(newMockCmd(testCase.mockCmdOutput, testCase.mockCmdErr)).Once()
			}

			verifier := NewVerifier(mockFS, mockExecutor)
			result, err := verifier.GetInstalledVersionWithLogging(testCase.installDir)

			if testCase.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
			}

			assert.Equal(t, testCase.expectedResult, result)
		})
	}
}

func TestVerifier_displayVerificationInfo(t *testing.T) {
	t.Parallel()
	// Since displayVerificationInfo uses logger, and we can't easily capture output,
	// we'll just ensure it doesn't panic with various inputs
	mockFS := mockFilesystem.NewMockFileSystem(t)
	mockExecutor := mockExec.NewMockCommandExecutor(t)

	verifier := NewVerifier(mockFS, mockExecutor)

	testCases := []VerificationInfo{
		{InstallDir: "/usr/local/go", Version: "go1.21.0", Status: "Verified"},
		{InstallDir: "", Version: "", Status: "Not installed"},
		{InstallDir: "/opt/go", Version: "go1.20.0", Status: "Verified"},
	}

	for _, info := range testCases {
		// Just call it to ensure no panic
		verifier.displayVerificationInfo(info)
	}
}

func TestInstallation(t *testing.T) {
	t.Parallel()
	mockFS := mockFilesystem.NewMockFileSystem(t)
	mockExecutor := mockExec.NewMockCommandExecutor(t)

	// Mock the OS filesystem and executor
	mockFS.EXPECT().Stat("/usr/local/go/bin/go").Return(nil, nil).Once()
	mockExecutor.EXPECT().CommandContext(context.Background(), "/usr/local/go/bin/go",
		[]string{"version"}).Return(newMockCmd([]byte("go version go1.21.0 linux/amd64"), nil)).Once()

	t.Skip("Skipping global function test to avoid host operations")
}

func TestGetInstalledVersion(t *testing.T) {
	t.Parallel()
	t.Skip("Skipping global function test to avoid host operations")
}

func TestGetVerificationInfo(t *testing.T) {
	t.Parallel()
	t.Skip("Skipping global function test to avoid host operations")
}

func TestGetInstalledVersionWithLogging(t *testing.T) {
	t.Parallel()
	t.Skip("Skipping global function test to avoid host operations")
}

func TestVerify(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		installDir     string
		mockVersion    string
		mockVersionErr error
		expectedError  string
	}{
		{
			name:           "successful verification",
			installDir:     "/usr/local/go",
			mockVersion:    "go1.21.0",
			mockVersionErr: nil,
			expectedError:  "",
		},
		{
			name:           "verification fails",
			installDir:     "/usr/local/go",
			mockVersion:    "",
			mockVersionErr: errStatErrorTest,
			expectedError:  "failed to get installed version: failed to check Go binary at /usr/local/go/bin/go: stat error",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			mockFS := mockFilesystem.NewMockFileSystem(t)
			mockExecutor := mockExec.NewMockCommandExecutor(t)

			// Mock GetVerificationInfo
			switch {
			case testCase.mockVersionErr != nil:
				mockFS.EXPECT().Stat("/usr/local/go/bin/go").Return(nil, testCase.mockVersionErr).Once()

				pathErr := &os.PathError{}
				if errors.As(testCase.mockVersionErr, &pathErr) {
					mockFS.EXPECT().IsNotExist(testCase.mockVersionErr).Return(true).Once()
				} else if testCase.mockVersionErr.Error() == "stat error" {
					mockFS.EXPECT().IsNotExist(testCase.mockVersionErr).Return(false).Once()
				}
			case testCase.mockVersion == "":
				mockFS.EXPECT().Stat("/usr/local/go/bin/go").Return(nil,
					&os.PathError{Op: "stat", Path: "/usr/local/go/bin/go", Err: os.ErrNotExist}).Once()
				mockFS.EXPECT().IsNotExist(mock.Anything).Return(true).Once()
			default:
				mockFS.EXPECT().Stat("/usr/local/go/bin/go").Return(nil, nil).Once()
				mockExecutor.EXPECT().CommandContext(context.Background(), "/usr/local/go/bin/go",
					[]string{"version"}).Return(newMockCmd([]byte("go version "+testCase.mockVersion+
					" linux/amd64\n"), nil)).Once()
			}

			verifier := NewVerifier(mockFS, mockExecutor)
			err := verifier.Verify(testCase.installDir)

			if testCase.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
			}
		})
	}
}

func TestRunVerify(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		verifyDir     string
		expectedError string
	}{
		{
			name:          "successful verification",
			verifyDir:     "/usr/local/go",
			expectedError: "",
		},
		{
			name:          "verification fails",
			verifyDir:     "/nonexistent/go",
			expectedError: "",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := RunVerify(testCase.verifyDir)

			if testCase.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
			}
		})
	}
}

func TestVerificationError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      *VerificationError
		expected string
	}{
		{
			name: "not installed error",
			err: &VerificationError{
				ExpectedVersion: "go1.21.0",
				ActualVersion:   "",
				BinaryPath:      "/usr/local/go/bin/go",
				Err:             nil,
			},
			expected: "verification failed: expected version go1.21.0 at /usr/local/go/bin/go, but Go is not installed",
		},
		{
			name: "version mismatch error",
			err: &VerificationError{
				ExpectedVersion: "go1.21.0",
				ActualVersion:   "go1.20.0",
				BinaryPath:      "/usr/local/go/bin/go",
				Err:             nil,
			},
			expected: "verification failed: expected version go1.21.0, got go1.20.0 at /usr/local/go/bin/go",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, testCase.expected, testCase.err.Error())
		})
	}
}

func TestVerificationError_Unwrap(t *testing.T) {
	t.Parallel()

	err := &VerificationError{
		ExpectedVersion: "go1.21.0",
		ActualVersion:   "go1.20.0",
		BinaryPath:      "/usr/local/go/bin/go",
		Err:             errUnderlyingErrorTest,
	}

	assert.Equal(t, errUnderlyingErrorTest, err.Unwrap())
}
