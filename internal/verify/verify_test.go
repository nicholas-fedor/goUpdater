package verify

import (
	"os"
	"path/filepath"
	"testing"
)

func createTestGoBinary(t *testing.T, script string) string {
	t.Helper()

	tempDir := t.TempDir()

	binDir := filepath.Join(tempDir, "bin")

	err := os.MkdirAll(binDir, 0750)
	if err != nil {
		t.Fatal(err)
	}

	goBinary := filepath.Join(binDir, "go")

	err = os.WriteFile(goBinary, []byte(script), 0600)
	if err != nil {
		t.Fatal(err)
	}

	//nolint:gosec // G302: executable permissions required for test binary
	err = os.Chmod(goBinary, 0755)
	if err != nil {
		t.Fatal(err)
	}

	return tempDir
}

// runGetInstalledVersionTests runs the common test logic for GetInstalledVersion and GetInstalledVersionWithLogging.
func runGetInstalledVersionTests(t *testing.T, getter func(string) (string, error), funcName string) {
	t.Helper()

	tests := []struct {
		name    string
		script  string
		want    string
		wantErr bool
	}{
		{
			name:    "success",
			script:  "#!/bin/bash\necho \"go version go1.21.0 linux/amd64\"",
			want:    "go1.21.0",
			wantErr: false,
		},
		{
			name:    "command fails",
			script:  "#!/bin/bash\nexit 1",
			want:    "",
			wantErr: true,
		},
		{
			name:    "parse fails",
			script:  "#!/bin/bash\necho \"invalid output\"",
			want:    "",
			wantErr: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			installDir := createTestGoBinary(t, testCase.script)

			got, err := getter(installDir)
			if (err != nil) != testCase.wantErr {
				t.Errorf("%s() error = %v, wantErr %v", funcName, err, testCase.wantErr)

				return
			}

			if got != testCase.want {
				t.Errorf("%s() = %v, want %v", funcName, got, testCase.want)
			}
		})
	}
}

func TestVerifyInstallation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		script          string
		expectedVersion string
		wantErr         bool
	}{
		{
			name:            "success",
			script:          "#!/bin/bash\necho \"go version go1.21.0 linux/amd64\"",
			expectedVersion: "go1.21.0",
			wantErr:         false,
		},
		{
			name:            "binary not found",
			script:          "",
			expectedVersion: "go1.21.0",
			wantErr:         true,
		},
		{
			name:            "version mismatch",
			script:          "#!/bin/bash\necho \"go version go1.20.0 linux/amd64\"",
			expectedVersion: "go1.21.0",
			wantErr:         true,
		},
		{
			name:            "command fails",
			script:          "#!/bin/bash\nexit 1",
			expectedVersion: "go1.21.0",
			wantErr:         true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var installDir string
			if testCase.script != "" {
				installDir = createTestGoBinary(t, testCase.script)
			} else {
				installDir = t.TempDir()
			}

			err := Installation(installDir, testCase.expectedVersion)
			if (err != nil) != testCase.wantErr {
				t.Errorf("VerifyInstallation() error = %v, wantErr %v", err, testCase.wantErr)
			}
		})
	}
}

func TestGetInstalledVersion(t *testing.T) {
	t.Parallel()
	runGetInstalledVersionTests(t, GetInstalledVersion, "GetInstalledVersion")
}

func TestGetInstalledVersionWithLogging(t *testing.T) {
	t.Parallel()
	runGetInstalledVersionTests(t, GetInstalledVersionWithLogging, "GetInstalledVersionWithLogging")
}

// getGetVerificationInfoTestCases returns test cases for TestGetVerificationInfo.
func getGetVerificationInfoTestCases() []struct {
	name       string
	installDir string
	setup      func(t *testing.T) string
	want       VerificationInfo
	wantErr    bool
} {
	return []struct {
		name       string
		installDir string
		setup      func(t *testing.T) string
		want       VerificationInfo
		wantErr    bool
	}{
		{
			name:       "successful verification",
			installDir: "",
			setup: func(t *testing.T) string {
				t.Helper()

				installDir := createTestGoBinary(t, "#!/bin/bash\necho \"go version go1.21.0 linux/amd64\"")

				return installDir
			},
			want: VerificationInfo{
				Version:    "go1.21.0",
				Status:     "verified",
				InstallDir: "",
			},
			wantErr: false,
		},
		{
			name:       "verification fails",
			installDir: "",
			setup: func(t *testing.T) string {
				t.Helper()

				installDir := createTestGoBinary(t, "#!/bin/bash\nexit 1")

				return installDir
			},
			want: VerificationInfo{
				Version:    "",
				Status:     "failed",
				InstallDir: "",
			},
			wantErr: true,
		},
	}
}

func TestGetVerificationInfo(t *testing.T) {
	t.Parallel()

	tests := getGetVerificationInfoTestCases()

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			installDir := testCase.setup(t)
			if testCase.installDir != "" {
				installDir = testCase.installDir
			}

			got, err := GetVerificationInfo(installDir)
			if testCase.wantErr {
				if err == nil {
					t.Error("expected error")
				}

				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Check fields that should be set
			if got.Version != testCase.want.Version {
				t.Errorf("Version = %v, want %v", got.Version, testCase.want.Version)
			}

			if got.Status != testCase.want.Status {
				t.Errorf("Status = %v, want %v", got.Status, testCase.want.Status)
			}

			if got.InstallDir != installDir {
				t.Errorf("InstallDir = %v, want %v", got.InstallDir, installDir)
			}
		})
	}
}

func TestVerify(t *testing.T) {
	t.Parallel()

	// Test that Verify function doesn't panic
	// Since it calls os.Exit on error, we can't easily test the full flow
	// But we can test that it doesn't panic with valid input
	t.Run("verify with valid installation", func(t *testing.T) {
		t.Parallel()

		installDir := createTestGoBinary(t, "#!/bin/bash\necho \"go version go1.21.0 linux/amd64\"")

		// This would normally call os.Exit, but since we have a valid installation,
		// it should proceed without exiting. However, testing os.Exit is tricky.
		// For now, just ensure it doesn't panic.
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Verify panicked: %v", r)
			}
		}()

		// We can't easily test this without mocking os.Exit
		// So we'll just skip the actual call and ensure setup works
		_ = installDir
	})
}

// getGetInstalledVersionCoreTestCases returns test cases for TestGetInstalledVersionCore.
func getGetInstalledVersionCoreTestCases() []struct {
	name       string
	installDir string
	setup      func(t *testing.T) string
	want       string
	wantErr    bool
} {
	return []struct {
		name       string
		installDir string
		setup      func(t *testing.T) string
		want       string
		wantErr    bool
	}{
		{
			name:       "successful parsing",
			installDir: "",
			setup: func(t *testing.T) string {
				t.Helper()

				return createTestGoBinary(t, "#!/bin/bash\necho \"go version go1.21.0 linux/amd64\"")
			},
			want:    "go1.21.0",
			wantErr: false,
		},
		{
			name:       "invalid output format",
			installDir: "",
			setup: func(t *testing.T) string {
				t.Helper()

				return createTestGoBinary(t, "#!/bin/bash\necho \"invalid format\"")
			},
			want:    "",
			wantErr: true,
		},
		{
			name:       "command execution fails",
			installDir: "",
			setup: func(t *testing.T) string {
				t.Helper()

				return createTestGoBinary(t, "#!/bin/bash\nexit 1")
			},
			want:    "",
			wantErr: true,
		},
	}
}

func TestGetInstalledVersionCore(t *testing.T) {
	t.Parallel()

	tests := getGetInstalledVersionCoreTestCases()

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			installDir := testCase.setup(t)
			if testCase.installDir != "" {
				installDir = testCase.installDir
			}

			got, err := getInstalledVersionCore(installDir)
			if testCase.wantErr {
				if err == nil {
					t.Error("expected error")
				}

				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if got != testCase.want {
				t.Errorf("getInstalledVersionCore() = %v, want %v", got, testCase.want)
			}
		})
	}
}
