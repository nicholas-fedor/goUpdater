// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package uninstall

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUninstallGo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(t *testing.T) string
		wantErr   bool
		checkFunc func(t *testing.T, installDir string)
	}{
		{
			name:    "success",
			setup:   setupSuccessUninstallTest,
			wantErr: false,
			checkFunc: func(t *testing.T, installDir string) {
				t.Helper()

				_, err := os.Stat(installDir)
				if !os.IsNotExist(err) {
					t.Errorf("directory should be removed")
				}
			},
		},
		{
			name:    "directory not found",
			setup:   setupDirectoryNotFoundTest,
			wantErr: true,
			checkFunc: func(t *testing.T, _ string) {
				t.Helper()

				// No check needed
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			installDir := testCase.setup(t)

			err := Remove(installDir)
			if (err != nil) != testCase.wantErr {
				t.Errorf("UninstallGo() error = %v, wantErr %v", err, testCase.wantErr)
			}

			testCase.checkFunc(t, installDir)
		})
	}
}

func setupSuccessUninstallTest(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()

	installDir := filepath.Join(tempDir, "go")

	err := os.MkdirAll(installDir, 0750)
	if err != nil {
		t.Fatal(err)
	}

	// Create some files
	file := filepath.Join(installDir, "test.txt")

	err = os.WriteFile(file, []byte("test"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	return installDir
}

func setupDirectoryNotFoundTest(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()

	return filepath.Join(tempDir, "nonexistent")
}

// runRemoveTest executes a single test case for the Remove function.
// It handles setup, execution, error checking, and verification.
func runRemoveTest(t *testing.T, testCase struct {
	name        string
	installDir  string
	setup       func(t *testing.T, dir string)
	wantErr     bool
	expectedErr string
}) {
	t.Helper()

	var installDir string
	if testCase.installDir == "" {
		installDir = filepath.Join(t.TempDir(), "go")
	} else {
		installDir = testCase.installDir
	}

	testCase.setup(t, installDir)

	err := Remove(installDir)
	if testCase.wantErr {
		if err == nil {
			t.Error("expected error")
		} else if !strings.Contains(err.Error(), testCase.expectedErr) {
			t.Errorf("expected error containing %q, got %q", testCase.expectedErr, err.Error())
		}
	} else if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify directory is removed for successful cases
	if !testCase.wantErr {
		_, err = os.Stat(installDir)
		if !os.IsNotExist(err) {
			t.Error("directory should have been removed")
		}
	}
}

// getRemoveTestCases returns the test cases for TestRemove.
func getRemoveTestCases() []struct {
	name        string
	installDir  string
	setup       func(t *testing.T, dir string)
	wantErr     bool
	expectedErr string
} {
	return []struct {
		name        string
		installDir  string
		setup       func(t *testing.T, dir string)
		wantErr     bool
		expectedErr string
	}{
		{
			name:        "successful removal",
			installDir:  "",
			setup:       setupSuccessfulRemoval,
			wantErr:     false,
			expectedErr: "",
		},
		{
			name:        "directory not found",
			installDir:  "",
			setup:       func(_ *testing.T, _ string) {}, // no setup
			wantErr:     true,
			expectedErr: "installation not found",
		},
		{
			name:        "empty directory",
			installDir:  "",
			setup:       setupEmptyDirectory,
			wantErr:     false,
			expectedErr: "",
		},
		{
			name:        "nested directories",
			installDir:  "",
			setup:       setupNestedDirectories,
			wantErr:     false,
			expectedErr: "",
		},
	}
}

// setupSuccessfulRemoval creates a directory with a test file.
func setupSuccessfulRemoval(t *testing.T, dir string) {
	t.Helper()

	err := os.MkdirAll(dir, 0750)
	if err != nil {
		t.Fatal(err)
	}

	file := filepath.Join(dir, "test.txt")

	err = os.WriteFile(file, []byte("test"), 0600)
	if err != nil {
		t.Fatal(err)
	}
}

// setupEmptyDirectory creates an empty directory.
func setupEmptyDirectory(t *testing.T, dir string) {
	t.Helper()

	err := os.MkdirAll(dir, 0750)
	if err != nil {
		t.Fatal(err)
	}
}

// setupNestedDirectories creates nested directories with a file.
func setupNestedDirectories(t *testing.T, dir string) {
	t.Helper()

	nestedDir := filepath.Join(dir, "bin")

	err := os.MkdirAll(nestedDir, 0750)
	if err != nil {
		t.Fatal(err)
	}

	file := filepath.Join(nestedDir, "go")

	err = os.WriteFile(file, []byte("binary"), 0600)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRemove(t *testing.T) {
	t.Parallel()

	tests := getRemoveTestCases()

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			runRemoveTest(t, testCase)
		})
	}
}
