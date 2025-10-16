// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package install

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func createTestArchive(t *testing.T, files map[string]string) string {
	t.Helper()

	tempFile, err := os.CreateTemp(t.TempDir(), "test-archive-*.tar.gz")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		closeErr := tempFile.Close()
		if closeErr != nil {
			t.Error(closeErr)
		}
	}()

	gzipWriter := gzip.NewWriter(tempFile)

	defer func() {
		closeErr := gzipWriter.Close()
		if closeErr != nil {
			t.Error(closeErr)
		}
	}()

	tarWriter := tar.NewWriter(gzipWriter)

	defer func() {
		closeErr := tarWriter.Close()
		if closeErr != nil {
			t.Error(closeErr)
		}
	}()

	for name, content := range files {
		header := createTarHeader(name, content)

		err := tarWriter.WriteHeader(header)
		if err != nil {
			t.Fatal(err)
		}

		_, err = tarWriter.Write([]byte(content))
		if err != nil {
			t.Fatal(err)
		}
	}

	return tempFile.Name()
}

func createTarHeader(name, content string) *tar.Header {
	return &tar.Header{
		Typeflag:   0,
		Name:       name,
		Linkname:   "",
		Size:       int64(len(content)),
		Mode:       0644,
		Uid:        0,
		Gid:        0,
		Uname:      "",
		Gname:      "",
		ModTime:    time.Time{},
		AccessTime: time.Time{},
		ChangeTime: time.Time{},
		Devmajor:   0,
		Devminor:   0,
		Xattrs:     nil,
		PAXRecords: nil,
		Format:     0,
	}
}

func TestInstallGo(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		setup     func(t *testing.T) (string, string)
		wantErr   bool
		checkFunc func(t *testing.T, installDir string)
	}{
		{
			name:      "success",
			setup:     setupSuccessTest,
			wantErr:   false,
			checkFunc: checkSuccessTest,
		},
		{
			name:      "invalid archive",
			setup:     setupInvalidArchiveTest,
			wantErr:   true,
			checkFunc: checkNoOp,
		},
		{
			name:      "archive not found",
			setup:     setupArchiveNotFoundTest,
			wantErr:   true,
			checkFunc: checkNoOp,
		},
	}

	runInstallTests(t, testCases)
}

func runInstallTests(t *testing.T, testCases []struct {
	name      string
	setup     func(t *testing.T) (string, string)
	wantErr   bool
	checkFunc func(t *testing.T, installDir string)
}) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			archivePath, installDir := testCase.setup(t)

			err := Go(archivePath, installDir)
			if (err != nil) != testCase.wantErr {
				t.Errorf("InstallGo() error = %v, wantErr %v", err, testCase.wantErr)
			}

			testCase.checkFunc(t, installDir)
		})
	}
}

func checkSuccessTest(t *testing.T, installDir string) {
	t.Helper()

	goBinary := filepath.Join(filepath.Dir(installDir), "go", "bin", "go")

	_, err := os.Stat(goBinary)
	if os.IsNotExist(err) {
		t.Errorf("go binary should exist")
	}

	readme := filepath.Join(filepath.Dir(installDir), "go", "README.md")

	_, err = os.Stat(readme)
	if os.IsNotExist(err) {
		t.Errorf("readme should exist")
	}
}

func checkNoOp(t *testing.T, _ string) {
	t.Helper()

	// No check needed
}

func setupSuccessTest(t *testing.T) (string, string) {
	t.Helper()

	files := map[string]string{
		"go/bin/go":    "fake go binary",
		"go/README.md": "readme content",
	}
	archivePath := createTestArchive(t, files)
	installDir := filepath.Join(t.TempDir(), "go")

	return archivePath, installDir
}

func setupInvalidArchiveTest(t *testing.T) (string, string) {
	t.Helper()

	tempFile, err := os.CreateTemp(t.TempDir(), "invalid-*.txt")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		closeErr := tempFile.Close()
		if closeErr != nil {
			t.Error(closeErr)
		}
	}()

	_, err = tempFile.WriteString("not a tar.gz")
	if err != nil {
		t.Fatal(err)
	}

	installDir := t.TempDir()

	return tempFile.Name(), installDir
}

func setupArchiveNotFoundTest(t *testing.T) (string, string) {
	t.Helper()

	installDir := t.TempDir()

	return "/nonexistent/archive.tar.gz", installDir
}

func TestPrepareInstallDir(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	installDir := filepath.Join(tempDir, "go")

	err := prepareInstallDir(installDir)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check that parent directory exists
	_, err = os.Stat(filepath.Dir(installDir))
	if os.IsNotExist(err) {
		t.Error("parent directory should exist")
	}
}
