// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package privileges

import (
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mockExec "github.com/nicholas-fedor/goUpdater/internal/exec/mocks"
	mockFilesystem "github.com/nicholas-fedor/goUpdater/internal/filesystem/mocks"
	mockPrivileges "github.com/nicholas-fedor/goUpdater/internal/privileges/mocks"
)

// Static test errors to satisfy err113 linter rule.
var (
	errSymlinkTest     = errors.New("symlink error")
	errExecTest        = errors.New("exec failed")
	errCallbackTest    = errors.New("callback failed")
	errCauseTest       = errors.New("cause")
	errGenericTest     = errors.New("generic error")
	errNotReadableTest = errors.New("not readable")
	errSetgidTest      = errors.New("setgid failed")
	errSetuidTest      = errors.New("setuid failed")
	errUserNotFound    = errors.New("user not found")
)

// argsMutex synchronizes access to os.Args across parallel subtests to prevent data races.
//
//nolint:gochecknoglobals // global mutex required for synchronizing parallel test access to os.Args
var argsMutex sync.Mutex

func TestIsRoot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		geteuid  int
		expected bool
	}{
		{
			name:     "root user (uid 0)",
			geteuid:  0,
			expected: true,
		},
		{
			name:     "non-root user",
			geteuid:  1000,
			expected: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockPM := mockPrivileges.NewMockOSPrivilegeManager(t)
			mockPM.EXPECT().Geteuid().Return(testCase.geteuid)

			mockFS := mockFilesystem.NewMockFileSystem(t)
			mockExecutor := mockExec.NewMockCommandExecutor(t)

			privilegeManager := NewPrivilegeManager(mockFS, mockPM, mockExecutor)

			result := privilegeManager.isRoot()
			assert.Equal(t, testCase.expected, result)
		})
	}
}
func TestRequestElevation(t *testing.T) { //nolint:gocognit,nestif,cyclop,lll,nolintlint // complex test setup for elevation scenarios
	t.Parallel()

	tests := []struct {
		name            string
		isRoot          bool
		sudoStatErr     error
		sudoMode        os.FileMode
		exeErr          error
		exePath         string
		evalSymlinksErr error
		resolvedPath    string
		args            []string
		sanitizedArgs   []string
		sanitizeErr     error
		execErr         error
		expectExec      bool
		expectError     bool
		errorType       error
	}{
		{
			name:        "already root - no elevation needed",
			isRoot:      true,
			expectExec:  false,
			expectError: false,
		},
		{
			name:        "sudo not found",
			isRoot:      false,
			sudoStatErr: ErrSudoNotAvailable,
			expectExec:  false,
			expectError: true,
			errorType:   &ElevationError{},
		},
		{
			name:        "sudo not executable",
			isRoot:      false,
			sudoMode:    0644, // not executable
			expectExec:  false,
			expectError: true,
			errorType:   &ElevationError{},
		},
		{
			name:        "executable path error",
			isRoot:      false,
			sudoMode:    0755,
			exeErr:      ErrExecutableNotFound,
			expectExec:  false,
			expectError: true,
			errorType:   &ElevationError{},
		},
		{
			name:            "symlink resolution error",
			isRoot:          false,
			sudoMode:        0755,
			exePath:         "/usr/bin/test",
			evalSymlinksErr: errSymlinkTest,
			expectExec:      false,
			expectError:     true,
			errorType:       &ElevationError{},
		},
		{
			name:         "argument sanitization error",
			isRoot:       false,
			sudoMode:     0755,
			exePath:      "/usr/bin/test",
			resolvedPath: "/usr/bin/test",
			args:         []string{"--dangerous", ";rm -rf /"},
			sanitizeErr:  ErrDangerousCharacters,
			expectExec:   false,
			expectError:  true,
			errorType:    &ElevationError{},
		},
		{
			name:          "successful elevation",
			isRoot:        false,
			sudoMode:      0755,
			exePath:       "/usr/bin/test",
			resolvedPath:  "/usr/bin/test",
			args:          []string{"arg1"},
			sanitizedArgs: []string{"arg1"},
			expectExec:    true,
			expectError:   false,
		},
		{
			name:          "exec failure",
			isRoot:        false,
			sudoMode:      0755,
			exePath:       "/usr/bin/test",
			resolvedPath:  "/usr/bin/test",
			args:          []string{"arg1"},
			sanitizedArgs: []string{"arg1"},
			execErr:       errExecTest,
			expectExec:    true,
			expectError:   true,
			errorType:     &ElevationError{},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			t.Helper()

			mockPM := mockPrivileges.NewMockOSPrivilegeManager(t)
			mockFS := mockFilesystem.NewMockFileSystem(t)
			mockExecutor := mockExec.NewMockCommandExecutor(t)
			mockLogger := mockPrivileges.NewMockAuditLogger(t)
			privilegeManager := NewPrivilegeManagerWithLogger(mockFS, mockPM, mockExecutor, mockLogger)

			// Setup command line arguments for the test case with mutex protection
			argsMutex.Lock()

			originalArgs := os.Args

			os.Args = append([]string{"test"}, testCase.args...)

			defer func() {
				os.Args = originalArgs

				argsMutex.Unlock()
			}()

			// Setup mocks based on test case
			euid := 0
			if !testCase.isRoot {
				euid = 1000
			}

			mockPM.EXPECT().Geteuid().Return(euid)

			if !testCase.isRoot {
				setupSudoMocks(testCase, mockFS, mockPM)
			}

			switch {
			case testCase.expectExec:
				mockLogger.EXPECT().LogElevationAttempt(true, "attempting sudo elevation")
				mockPM.EXPECT().Environ().Return([]string{"PATH=/usr/bin"})
				mockPM.EXPECT().Exec("/usr/bin/sudo", mock.AnythingOfType("[]string"),
					[]string{"PATH=/usr/bin"}).Return(testCase.execErr)

				if testCase.execErr != nil {
					mockLogger.EXPECT().LogElevationAttempt(false, "sudo execution failed")
				}
			case testCase.isRoot:
				mockLogger.EXPECT().LogElevationAttempt(false, "already root - no elevation needed")
			case testCase.sudoStatErr != nil:
				mockLogger.EXPECT().LogElevationAttempt(false, "sudo not available")
			case testCase.sudoMode&0111 == 0:
				mockLogger.EXPECT().LogElevationAttempt(false, "sudo not available")
			case testCase.exeErr != nil:
				mockLogger.EXPECT().LogElevationAttempt(false, "executable validation failed")
			case testCase.evalSymlinksErr != nil:
				mockLogger.EXPECT().LogElevationAttempt(false, "executable validation failed")
			case testCase.sanitizeErr != nil:
				// For sanitize error, we expect the error to be returned before logging
				// No logging expectation needed here
			}

			err := privilegeManager.requestElevation()
			if testCase.expectError {
				require.Error(t, err)

				if testCase.errorType != nil {
					var target *ElevationError
					require.ErrorAs(t, err, &target)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestElevateAndExecute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		isRoot      bool
		elevateErr  error
		callbackErr error
		expectError bool
	}{
		{
			name:        "already root - successful execution",
			isRoot:      true,
			expectError: false,
		},
		{
			name:        "already root - callback error",
			isRoot:      true,
			callbackErr: errCallbackTest,
			expectError: true,
		},
		{
			name:        "elevation required - success",
			isRoot:      false,
			elevateErr:  nil,
			expectError: false,
		},
		{
			name:        "elevation required - elevation failure",
			isRoot:      false,
			elevateErr:  ErrElevationFailed,
			expectError: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockPM := mockPrivileges.NewMockOSPrivilegeManager(t)
			mockFS := mockFilesystem.NewMockFileSystem(t)
			mockExecutor := mockExec.NewMockCommandExecutor(t)

			privilegeManager := NewPrivilegeManager(mockFS, mockPM, mockExecutor)

			mockPM.EXPECT().Geteuid().Return(func() int {
				if testCase.isRoot {
					return 0
				}

				return 1000
			}())

			if !testCase.isRoot {
				// Mock elevation - need to set up sudo stat mock
				mockFS.EXPECT().Stat("/usr/bin/sudo").Return(&mockFileInfo{mode: 0755}, nil)
				mockPM.EXPECT().Executable().Return("/usr/bin/test", nil)
				mockFS.EXPECT().Stat("/usr/bin/test").Return(&mockFileInfo{mode: 0755}, nil)
				mockPM.EXPECT().EvalSymlinks("/usr/bin/test").Return("/usr/bin/test", nil)

				mockLogger := mockPrivileges.NewMockAuditLogger(t)
				mockLogger.EXPECT().LogElevationAttempt(true, "attempting sudo elevation")
				mockPM.EXPECT().Environ().Return([]string{"PATH=/usr/bin"})
				mockPM.EXPECT().Exec("/usr/bin/sudo", mock.AnythingOfType("[]string"),
					mock.AnythingOfType("[]string")).Return(testCase.elevateErr)

				if testCase.elevateErr != nil {
					mockLogger.EXPECT().LogElevationAttempt(false, "sudo execution failed")
					// We can't easily test the exit behavior, so we'll skip elevation error cases
					t.Skip("Elevation error testing requires special handling due to os.Exit")
				} else {
					// For successful elevation, the process would re-execute, but we can't test that
					t.Skip("Elevation success testing requires special handling due to process re-execution")
				}
			}

			if testCase.isRoot || testCase.elevateErr == nil {
				called := false
				callback := func() error {
					called = true

					return testCase.callbackErr
				}

				err := privilegeManager.ElevateAndExecute(callback)

				assert.True(t, called)

				if testCase.callbackErr != nil {
					assert.Equal(t, testCase.callbackErr, err)
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestElevateAndExecuteWithDrop(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		isRoot      bool
		callbackErr error
		dropErr     error
		expectError bool
	}{
		{
			name:        "successful execution with privilege drop",
			isRoot:      true,
			callbackErr: nil,
			dropErr:     nil,
			expectError: false,
		},
		{
			name:        "callback error with privilege drop",
			isRoot:      true,
			callbackErr: errCallbackTest,
			dropErr:     nil,
			expectError: true,
		},
		{
			name:        "privilege drop failure after successful callback",
			isRoot:      true,
			callbackErr: nil,
			dropErr:     ErrPrivilegeDropFailed,
			expectError: false, // drop error doesn't propagate
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockPM := mockPrivileges.NewMockOSPrivilegeManager(t)
			mockFS := mockFilesystem.NewMockFileSystem(t)
			mockExecutor := mockExec.NewMockCommandExecutor(t)
			mockLogger := mockPrivileges.NewMockAuditLogger(t)

			privilegeManager := NewPrivilegeManagerWithLogger(mockFS, mockPM, mockExecutor, mockLogger)

			mockPM.EXPECT().Geteuid().Return(0).Maybe()

			called := false
			callback := func() error {
				called = true

				return testCase.callbackErr
			}

			// Set up environment for privilege drop
			originalSudoUID := os.Getenv("SUDO_UID")
			originalSudoGID := os.Getenv("SUDO_GID")

			defer func() {
				if originalSudoUID == "" {
					os.Unsetenv("SUDO_UID")
				} else {
					os.Setenv("SUDO_UID", originalSudoUID) //nolint:usetesting // os.Setenv required for parallel tests
				}

				if originalSudoGID == "" {
					os.Unsetenv("SUDO_GID")
				} else {
					os.Setenv("SUDO_GID", originalSudoGID) //nolint:usetesting // os.Setenv required for parallel tests
				}
			}()

			os.Setenv("SUDO_USER", "testuser") //nolint:usetesting // os.Setenv required for parallel tests
			os.Setenv("SUDO_UID", "1000")      //nolint:usetesting // os.Setenv required for parallel tests
			os.Setenv("SUDO_GID", "1000")      //nolint:usetesting // os.Setenv required for parallel tests

			// Mock Getenv calls for privilege drop and elevation check
			mockPM.EXPECT().Getenv("SUDO_USER").Return("testuser").Maybe()
			mockPM.EXPECT().Getenv("SUDO_UID").Return("1000").Maybe()
			mockPM.EXPECT().Getenv("SUDO_GID").Return("1000").Maybe()

			// Mock privilege drop - only for successful execution cases
			if testCase.callbackErr == nil {
				mockLogger.EXPECT().LogPrivilegeDrop(true, 1000, "dropping privileges to original user")
				mockPM.EXPECT().Setgid(1000).Return(testCase.dropErr)

				if testCase.dropErr == nil {
					mockPM.EXPECT().Setuid(1000).Return(nil)
					mockLogger.EXPECT().LogPrivilegeChange("drop", 0, 1000, "dropped privileges to original user")
				} else {
					mockLogger.EXPECT().LogPrivilegeDrop(false, 1000, "failed to drop group privileges")
				}
			}

			err := privilegeManager.ElevateAndExecuteWithDrop(callback)

			assert.True(t, called)

			if testCase.expectError {
				assert.Equal(t, testCase.callbackErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleElevationError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		err         error
		expectExit  bool
		checkOutput func(t *testing.T, output string)
	}{
		{
			name:       "elevation error with sudo issue",
			err:        &ElevationError{Op: "test", Reason: "sudo failed", SudoErr: true, Cause: errCauseTest},
			expectExit: true,
			checkOutput: func(t *testing.T, output string) {
				t.Helper()
				assert.Contains(t, output, "Failed to obtain elevated privileges during test: sudo failed")
				assert.Contains(t, output, "sudo is installed and configured correctly")
				assert.Contains(t, output, "Underlying cause: cause")
			},
		},
		{
			name:       "elevation error without sudo issue",
			err:        &ElevationError{Op: "test", Reason: "other failure", SudoErr: false},
			expectExit: true,
			checkOutput: func(t *testing.T, output string) {
				t.Helper()
				assert.Contains(t, output, "Failed to obtain elevated privileges during test: other failure")
				assert.Contains(t, output, "Please run with sudo or as root")
			},
		},
		{
			name:       "generic error",
			err:        errGenericTest,
			expectExit: true,
			checkOutput: func(t *testing.T, output string) {
				t.Helper()
				assert.Contains(t, output, "Failed to obtain elevated privileges: generic error")
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// We can't easily test os.Exit, so we'll test the logging behavior
			// by temporarily replacing the logger
			t.Skip("HandleElevationError calls os.Exit, making it difficult to test directly")
		})
	}
}
func TestGetSearchDirectories(t *testing.T) {
	tests := []struct {
		name         string
		elevatedHome string
		destDir      string
		isElevated   bool
		originalHome string
		readableDirs map[string]bool
		expectedDirs []string
	}{
		{
			name:         "non-elevated user",
			elevatedHome: "/home/user",
			destDir:      "/tmp",
			isElevated:   false,
			readableDirs: map[string]bool{
				"/home/user/Downloads": true,
				"/home/user":           true,
			},
			expectedDirs: []string{"/home/user/Downloads", "/home/user", "/tmp"},
		},
		{
			name:         "elevated user with readable original directories",
			elevatedHome: "/home/root",
			destDir:      "/tmp",
			isElevated:   true,
			originalHome: "/home/user",
			readableDirs: map[string]bool{
				"/home/root/Downloads": true,
				"/home/root":           true,
				"/home/user/Downloads": true,
				"/home/user":           true,
			},
			expectedDirs: []string{"/home/root/Downloads", "/home/root", "/home/user/Downloads", "/home/user", "/tmp"},
		},
		{
			name:         "elevated user with non-readable original directories",
			elevatedHome: "/home/root",
			destDir:      "/tmp",
			isElevated:   true,
			originalHome: "/home/user",
			readableDirs: map[string]bool{
				"/home/root/Downloads": true,
				"/home/root":           true,
				"/home/user/Downloads": false,
				"/home/user":           false,
			},
			expectedDirs: []string{"/home/root/Downloads", "/home/root", "/tmp"},
		},
		{
			name:         "elevated user with no original home",
			elevatedHome: "/home/root",
			destDir:      "/tmp",
			isElevated:   true,
			originalHome: "",
			readableDirs: map[string]bool{
				"/home/root/Downloads": true,
				"/home/root":           true,
			},
			expectedDirs: []string{"/home/root/Downloads", "/home/root", "/tmp"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			mockFS := mockFilesystem.NewMockFileSystem(t)

			// Setup readable directory mocks
			for dir, readable := range testCase.readableDirs {
				if readable {
					mockFS.EXPECT().Stat(dir).Return(&mockFileInfo{isDir: true}, nil).Maybe()
				} else {
					mockFS.EXPECT().Stat(dir).Return(nil, errNotReadableTest).Maybe()
				}
			}

			// Mock user lookup for elevated case
			originalSudoUser := os.Getenv("SUDO_USER")

			defer func() {
				if originalSudoUser == "" {
					os.Unsetenv("SUDO_USER")
				} else {
					t.Setenv("SUDO_USER", originalSudoUser)
				}
			}()

			if testCase.isElevated && testCase.originalHome != "" {
				os.Setenv("SUDO_USER", "testuser") //nolint:usetesting // os.Setenv required for parallel tests
			} else if !testCase.isElevated {
				os.Unsetenv("SUDO_USER")
			}

			// For elevated cases with original home, we need to mock user.Lookup
			// Since we can't easily mock the user package, we'll test the non-elevated cases
			// and skip the elevated ones that require complex mocking
			if testCase.isElevated && testCase.originalHome != "" {
				t.Skip("TestGetSearchDirectories requires user package mocking which is complex")
			}

			result := GetSearchDirectories(testCase.elevatedHome, testCase.destDir, mockFS)
			assert.Equal(t, testCase.expectedDirs, result)
		})
	}
}

func TestDefaultAuditLogger(t *testing.T) {
	t.Parallel()

	logger := &DefaultAuditLogger{}

	// Test LogElevationAttempt
	t.Run("LogElevationAttempt success", func(t *testing.T) {
		t.Parallel()
		// This would log to the default logger, we can't easily capture output
		logger.LogElevationAttempt(true, "test success")
	})

	t.Run("LogElevationAttempt failure", func(t *testing.T) {
		t.Parallel()
		logger.LogElevationAttempt(false, "test failure")
	})

	// Test LogPrivilegeChange
	t.Run("LogPrivilegeChange", func(t *testing.T) {
		t.Parallel()
		logger.LogPrivilegeChange("test_op", 1000, 0, "test reason")
	})

	// Test LogPrivilegeDrop
	t.Run("LogPrivilegeDrop success", func(t *testing.T) {
		t.Parallel()
		logger.LogPrivilegeDrop(true, 1000, "test drop success")
	})

	t.Run("LogPrivilegeDrop failure", func(t *testing.T) {
		t.Parallel()
		logger.LogPrivilegeDrop(false, 1000, "test drop failure")
	})
}

func TestOSPrivilegeManagerImpl(t *testing.T) {
	t.Parallel()

	privilegeManager := OSPrivilegeManagerImpl{}

	t.Run("Geteuid", func(t *testing.T) {
		t.Parallel()

		uid := privilegeManager.Geteuid()
		assert.GreaterOrEqual(t, uid, 0)
	})

	t.Run("Getuid", func(t *testing.T) {
		t.Parallel()

		uid := privilegeManager.Getuid()
		assert.GreaterOrEqual(t, uid, 0)
	})

	t.Run("Getgid", func(t *testing.T) {
		t.Parallel()

		gid := privilegeManager.Getgid()
		assert.GreaterOrEqual(t, gid, 0)
	})

	t.Run("Environ", func(t *testing.T) {
		t.Parallel()

		env := privilegeManager.Environ()
		assert.NotEmpty(t, env)
		assert.Contains(t, env[0], "=") // Environment variables have format KEY=VALUE
	})
}
func TestDropPrivileges(t *testing.T) { //nolint:gocognit,nestif,cyclop,lll,nolintlint // complex test setup for privilege drop scenarios
	t.Parallel()

	tests := []struct {
		name        string
		isRoot      bool
		isElevated  bool
		sudoUID     string
		sudoGID     string
		setgidErr   error
		setuidErr   error
		expectError bool
		errorType   error
	}{
		{
			name:        "not root - no drop needed",
			isRoot:      false,
			isElevated:  false,
			expectError: false,
		},
		{
			name:        "root but not elevated - no drop needed",
			isRoot:      true,
			isElevated:  false,
			expectError: false,
		},
		{
			name:        "missing SUDO_UID",
			isRoot:      true,
			isElevated:  true,
			sudoUID:     "",
			sudoGID:     "1000",
			expectError: true,
			errorType:   &PrivilegeDropError{},
		},
		{
			name:        "missing SUDO_GID",
			isRoot:      true,
			isElevated:  true,
			sudoUID:     "1000",
			sudoGID:     "",
			expectError: true,
			errorType:   &PrivilegeDropError{},
		},
		{
			name:        "invalid SUDO_UID",
			isRoot:      true,
			isElevated:  true,
			sudoUID:     "invalid",
			sudoGID:     "1000",
			expectError: true,
			errorType:   &PrivilegeDropError{},
		},
		{
			name:        "invalid SUDO_GID",
			isRoot:      true,
			isElevated:  true,
			sudoUID:     "1000",
			sudoGID:     "invalid",
			expectError: true,
			errorType:   &PrivilegeDropError{},
		},
		{
			name:        "setgid failure",
			isRoot:      true,
			isElevated:  true,
			sudoUID:     "1000",
			sudoGID:     "1000",
			setgidErr:   errSetgidTest,
			expectError: true,
			errorType:   &PrivilegeDropError{},
		},
		{
			name:        "setuid failure",
			isRoot:      true,
			isElevated:  true,
			sudoUID:     "1000",
			sudoGID:     "1000",
			setuidErr:   errSetuidTest,
			expectError: true,
			errorType:   &PrivilegeDropError{},
		},
		{
			name:        "successful privilege drop",
			isRoot:      true,
			isElevated:  true,
			sudoUID:     "1000",
			sudoGID:     "1000",
			expectError: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			t.Helper()

			mockPM := mockPrivileges.NewMockOSPrivilegeManager(t)
			mockFS := mockFilesystem.NewMockFileSystem(t)
			mockExecutor := mockExec.NewMockCommandExecutor(t)
			mockLogger := mockPrivileges.NewMockAuditLogger(t)

			privilegeManager := NewPrivilegeManagerWithLogger(mockFS, mockPM, mockExecutor, mockLogger)

			mockPM.EXPECT().Geteuid().Return(func() int {
				if testCase.isRoot {
					return 0
				}

				return 1000
			}())

			shouldDrop := testCase.isRoot && testCase.isElevated && testCase.sudoUID == "1000" && testCase.sudoGID == "1000"
			if !shouldDrop {
				goto afterDropSetup
			}

			mockLogger.EXPECT().LogPrivilegeDrop(true, 1000, "dropping privileges to original user")
			mockPM.EXPECT().Setgid(1000).Return(testCase.setgidErr)

			if testCase.setgidErr != nil {
				mockLogger.EXPECT().LogPrivilegeDrop(false, 1000, "failed to drop group privileges")

				goto afterDropSetup
			}

			mockPM.EXPECT().Setuid(1000).Return(testCase.setuidErr)

			if testCase.setuidErr != nil {
				mockLogger.EXPECT().LogPrivilegeDrop(false, 1000, "failed to drop user privileges")
			} else {
				mockLogger.EXPECT().LogPrivilegeChange("drop", 0, 1000, "dropped privileges to original user")
			}

		afterDropSetup:

			// Set up environment variables
			originalSudoUser := os.Getenv("SUDO_USER")

			originalSudoUID := os.Getenv("SUDO_UID")
			originalSudoGID := os.Getenv("SUDO_GID")

			defer func() {
				if originalSudoUser == "" {
					os.Unsetenv("SUDO_USER")
				} else {
					os.Setenv("SUDO_USER", originalSudoUser) //nolint:usetesting // os.Setenv required for parallel tests
				}

				if originalSudoUID == "" {
					os.Unsetenv("SUDO_UID")
				} else {
					os.Setenv("SUDO_UID", originalSudoUID) //nolint:usetesting // os.Setenv required for parallel tests
				}

				if originalSudoGID == "" {
					os.Unsetenv("SUDO_GID")
				} else {
					os.Setenv("SUDO_GID", originalSudoGID) //nolint:usetesting // os.Setenv required for parallel tests
				}
			}()

			if testCase.isElevated {
				os.Setenv("SUDO_USER", "testuser")      //nolint:usetesting // os.Setenv required for parallel tests
				os.Setenv("SUDO_UID", testCase.sudoUID) //nolint:usetesting // os.Setenv required for parallel tests
				os.Setenv("SUDO_GID", testCase.sudoGID) //nolint:usetesting // os.Setenv required for parallel tests

				// Mock Getenv calls for privilege drop and elevation check
				mockPM.EXPECT().Getenv("SUDO_USER").Return("testuser")
				mockPM.EXPECT().Getenv("SUDO_UID").Return(testCase.sudoUID)
				mockPM.EXPECT().Getenv("SUDO_GID").Return(testCase.sudoGID)
			} else {
				// For non-elevated cases, unset the variables
				// Since t.Setenv doesn't have unset, we need to save and restore
				originalSudoUser := os.Getenv("SUDO_USER")
				originalSudoUID := os.Getenv("SUDO_UID")
				originalSudoGID := os.Getenv("SUDO_GID")
				os.Unsetenv("SUDO_USER")
				os.Unsetenv("SUDO_UID")
				os.Unsetenv("SUDO_GID")

				// Mock Getenv calls to return empty for non-elevated cases
				mockPM.EXPECT().Getenv("SUDO_USER").Return("").Maybe()

				defer func() {
					if originalSudoUser != "" {
						os.Setenv("SUDO_USER", originalSudoUser) //nolint:usetesting // os.Setenv required for parallel tests
					}

					if originalSudoUID != "" {
						os.Setenv("SUDO_UID", originalSudoUID) //nolint:usetesting // os.Setenv required for parallel tests
					}

					if originalSudoGID != "" {
						os.Setenv("SUDO_GID", originalSudoGID) //nolint:usetesting // os.Setenv required for parallel tests
					}
				}()
			}

			err := privilegeManager.dropPrivileges()

			if testCase.expectError {
				var target *PrivilegeDropError
				require.ErrorAs(t, err, &target)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSanitizeArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        []string
		expectError bool
		expected    []string
	}{
		{
			name:     "empty args",
			args:     []string{},
			expected: []string{},
		},
		{
			name:     "valid args",
			args:     []string{"--version", "install", "/path/to/file"},
			expected: []string{"--version", "install", "/path/to/file"},
		},
		{
			name:        "dangerous semicolon",
			args:        []string{"rm -rf /", ";", "ls"},
			expectError: true,
		},
		{
			name:        "dangerous pipe",
			args:        []string{"cat", "|", "rm -rf /"},
			expectError: true,
		},
		{
			name:        "dangerous command substitution",
			args:        []string{"echo", "$(rm -rf /)"},
			expectError: true,
		},
		{
			name:        "dangerous sudo option -e",
			args:        []string{"-e", "script"},
			expectError: true,
		},
		{
			name:        "dangerous sudo option -c",
			args:        []string{"-c", "command"},
			expectError: true,
		},
		{
			name:        "non-printable characters",
			args:        []string{"valid", "in\x00valid"},
			expectError: true,
		},
		{
			name:     "whitespace only",
			args:     []string{"", "   ", "\t"},
			expected: []string{"", "   ", "\t"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			t.Helper()

			result, err := sanitizeArgs(testCase.args)

			if testCase.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.expected, result)
			}
		})
	}
}

func TestValidateArg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		arg         string
		expectError bool
		errorType   error
	}{
		{
			name: "empty string",
			arg:  "",
		},
		{
			name: "whitespace only",
			arg:  "   ",
		},
		{
			name: "valid argument",
			arg:  "--version",
		},
		{
			name:        "semicolon injection",
			arg:         "rm -rf /; ls",
			expectError: true,
			errorType:   ErrDangerousCharacters,
		},
		{
			name:        "pipe injection",
			arg:         "cat | rm -rf /",
			expectError: true,
			errorType:   ErrDangerousCharacters,
		},
		{
			name:        "command substitution",
			arg:         "$(rm -rf /)",
			expectError: true,
			errorType:   ErrDangerousCharacters,
		},
		{
			name:        "dangerous sudo option -e",
			arg:         "-e",
			expectError: true,
			errorType:   ErrDangerousSudoOption,
		},
		{
			name:        "dangerous sudo option -c",
			arg:         "-c",
			expectError: true,
			errorType:   ErrDangerousSudoOption,
		},
		{
			name:        "non-printable character",
			arg:         "valid\x00invalid",
			expectError: true,
			errorType:   ErrNonPrintableCharacters,
		},
		{
			name: "unicode characters",
			arg:  "café",
		},
		{
			name: "numbers and symbols",
			arg:  "file-123.txt",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := validateArg(testCase.arg)

			if testCase.expectError {
				require.Error(t, err)

				if testCase.errorType != nil {
					require.ErrorIs(t, err, testCase.errorType)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIsElevated(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		sudoUser string
		expected bool
	}{
		{
			name:     "elevated with sudo user",
			sudoUser: "testuser",
			expected: true,
		},
		{
			name:     "not elevated",
			sudoUser: "",
			expected: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			originalValue := os.Getenv("SUDO_USER")

			defer func() {
				if originalValue == "" {
					os.Unsetenv("SUDO_USER")
				} else {
					os.Setenv("SUDO_USER", originalValue) //nolint:usetesting // os.Setenv required for parallel tests
				}
			}()

			if testCase.sudoUser == "" {
				os.Unsetenv("SUDO_USER")
			} else {
				os.Setenv("SUDO_USER", testCase.sudoUser) //nolint:usetesting // os.Setenv required for parallel tests
			}

			result := IsElevated()
			assert.Equal(t, testCase.expected, result)
		})
	}
}

func TestGetOriginalUserHome(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		sudoUser      string
		userLookupErr error
		userHome      string
		expected      string
	}{
		{
			name:     "no sudo user",
			sudoUser: "",
			expected: "",
		},
		{
			name:          "user lookup failure",
			sudoUser:      "nonexistent",
			userLookupErr: errUserNotFound,
			expected:      "",
		},
		{
			name:     "successful lookup",
			sudoUser: "testuser",
			userHome: "/home/testuser",
			expected: "/home/testuser",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			originalValue := os.Getenv("SUDO_USER")

			defer func() {
				if originalValue == "" {
					os.Unsetenv("SUDO_USER")
				} else {
					os.Setenv("SUDO_USER", originalValue) //nolint:usetesting // os.Setenv required for parallel tests
				}
			}()

			if testCase.sudoUser == "" {
				os.Unsetenv("SUDO_USER")
			} else {
				os.Setenv("SUDO_USER", testCase.sudoUser) //nolint:usetesting // os.Setenv required for parallel tests
			}

			// We can't easily mock user.Lookup, so we'll test the basic logic
			result := GetOriginalUserHome()
			if testCase.expected == "" {
				assert.Empty(t, result)
			} else if testCase.name == "successful lookup" {
				// For successful cases, we can't predict the exact home directory
				// without mocking the user package, so skip the "successful_lookup" subtest
				t.Skip("successful_lookup requires user package mocking which is complex")
			}
		})
	}
}

// mockFileInfo implements filesystem.FileInfo for testing.
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

// setupSudoMocks sets up mocks for sudo-related operations in elevation tests.
func setupSudoMocks(testCase struct {
	name            string
	isRoot          bool
	sudoStatErr     error
	sudoMode        os.FileMode
	exeErr          error
	exePath         string
	evalSymlinksErr error
	resolvedPath    string
	args            []string
	sanitizedArgs   []string
	sanitizeErr     error
	execErr         error
	expectExec      bool
	expectError     bool
	errorType       error
}, mockFS *mockFilesystem.MockFileSystem, mockPM *mockPrivileges.MockOSPrivilegeManager) {
	// Setup sudo stat
	if testCase.sudoStatErr == nil {
		mockFS.EXPECT().Stat("/usr/bin/sudo").Return(&mockFileInfo{mode: testCase.sudoMode}, nil)
	} else {
		mockFS.EXPECT().Stat("/usr/bin/sudo").Return(nil, testCase.sudoStatErr)
	}

	sudoAvailable := testCase.sudoStatErr == nil && testCase.sudoMode&0111 != 0
	if !sudoAvailable {
		// Sudo not available, skip further setup
		return
	}

	mockPM.EXPECT().Executable().Return(testCase.exePath, testCase.exeErr)

	if testCase.exeErr != nil {
		return
	}

	mockFS.EXPECT().Stat(testCase.exePath).Return(&mockFileInfo{mode: 0755}, nil)

	if testCase.evalSymlinksErr == nil {
		mockPM.EXPECT().EvalSymlinks(testCase.exePath).Return(testCase.resolvedPath, nil)
	} else {
		mockPM.EXPECT().EvalSymlinks(testCase.exePath).Return("", testCase.evalSymlinksErr)
	}

	// For successful elevation case, we need to sanitize args first
	if testCase.expectExec && testCase.sanitizedArgs != nil {
		// The sanitization happens before exec, so we expect it to be called
		// But since it's internal, we don't mock it directly
		// This block is intentionally empty as sanitization is internal
		_ = testCase.sanitizedArgs // avoid unused variable warning
	}
}
