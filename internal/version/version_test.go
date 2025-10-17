// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package version

import (
	"os"
	"strings"
	"testing"
)

// resetGlobals resets all global version variables to empty strings for testing.
func resetGlobals() {
	versionMutex.Lock()

	version = ""
	commit = ""
	date = ""
	goVersion = ""
	platform = ""

	versionMutex.Unlock()
}

//nolint:funlen
func TestGetVersionInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFunc func()
		expected  VersionInfo
	}{
		{
			name: "all ldflags set",
			setupFunc: func() {
				resetGlobals()
				SetVersion("1.2.3")
				SetCommit("abc123")
				SetDate("2023-10-01T12:00:00Z")
				SetGoVersion("go1.21.0")
				SetPlatform("linux/amd64")
			},
			expected: VersionInfo{
				Version:   "1.2.3",
				Commit:    "abc123",
				Date:      "2023-10-01T12:00:00Z",
				GoVersion: "go1.21.0",
				Platform:  "linux/amd64",
			},
		},
		{
			name: "version empty defaults to dev",
			setupFunc: func() {
				resetGlobals()
				SetCommit("def456")
				SetDate("2023-10-02T13:00:00Z")
			},
			expected: VersionInfo{
				Version:   "dev",
				Commit:    "def456",
				Date:      "2023-10-02T13:00:00Z",
				GoVersion: "",
				Platform:  "",
			},
		},
		{
			name: "missing commit and date",
			setupFunc: func() {
				resetGlobals()
				SetVersion("2.0.0")
				SetGoVersion("go1.22.0")
			},
			expected: VersionInfo{
				Version:   "2.0.0",
				Commit:    "",
				Date:      "",
				GoVersion: "go1.22.0",
				Platform:  "",
			},
		},
		{
			name: "invalid date format",
			setupFunc: func() {
				resetGlobals()
				SetVersion("1.0.0")
				SetDate("invalid-date")
			},
			expected: VersionInfo{
				Version:   "1.0.0",
				Commit:    "",
				Date:      "invalid-date",
				GoVersion: "",
				Platform:  "",
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testCase.setupFunc()

			result := GetVersionInfo()
			if result != testCase.expected {
				t.Errorf("GetVersionInfo() = %v, want %v", result, testCase.expected)
			}
		})
	}
}

// TestGetVersionInfoWithDebugFallback tests fallback to debug.ReadBuildInfo
// This is a simplified test since mocking debug.ReadBuildInfo is complex.
func TestGetVersionInfoWithDebugFallback(t *testing.T) {
	t.Parallel()
	resetGlobals()

	SetVersion("1.0.0")
	// commit and date are empty, so it should try debug.ReadBuildInfo
	// In a real scenario, this would depend on build info availability
	result := GetVersionInfo()

	const expectedVersion = "1.0.0"
	if result.Version != expectedVersion {
		t.Errorf("Expected version %s, got %s", expectedVersion, result.Version)
	}
	// Note: Testing actual debug.ReadBuildInfo fallback would require build tags
	// or more complex mocking, which is beyond basic unit test scope
}

// TestSetterFunctions tests all setter functions.
func TestSetterFunctions(t *testing.T) {
	t.Parallel()
	resetGlobals()

	SetVersion("1.0.0")
	SetCommit("abc123")
	SetDate("2023-10-01T12:00:00Z")
	SetGoVersion("go1.21.0")
	SetPlatform("linux/amd64")

	info := GetVersionInfo()
	if info.Version != "1.0.0" {
		t.Errorf("SetVersion failed: got %s, want 1.0.0", info.Version)
	}

	if info.Commit != "abc123" {
		t.Errorf("SetCommit failed: got %s, want abc123", info.Commit)
	}

	if info.Date != "2023-10-01T12:00:00Z" {
		t.Errorf("SetDate failed: got %s, want 2023-10-01T12:00:00Z", info.Date)
	}

	if info.GoVersion != "go1.21.0" {
		t.Errorf("SetGoVersion failed: got %s, want go1.21.0", info.GoVersion)
	}

	if info.Platform != "linux/amd64" {
		t.Errorf("SetPlatform failed: got %s, want linux/amd64", info.Platform)
	}
}

// TestGetterFunctions tests Get, GetCommit, GetDate functions.
func TestGetterFunctions(t *testing.T) {
	t.Parallel()
	resetGlobals()

	SetVersion("2.1.0")
	SetCommit("def789")
	SetDate("2023-10-02T14:00:00Z")

	if got := Get(); got != "2.1.0" {
		t.Errorf("Get() = %s, want 2.1.0", got)
	}

	if got := GetCommit(); got != "def789" {
		t.Errorf("GetCommit() = %s, want def789", got)
	}

	if got := GetDate(); got != "2023-10-02T14:00:00Z" {
		t.Errorf("GetDate() = %s, want 2023-10-02T14:00:00Z", got)
	}
}

// TestDisplay tests the Display function by capturing stdout.
//
//nolint:funlen
func TestDisplay(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFunc func()
		expected  []string // substrings that should be in output
	}{
		{
			name: "full information",
			setupFunc: func() {
				resetGlobals()
				SetVersion("1.2.3")
				SetCommit("abc123")
				SetDate("2023-10-01T12:00:00Z")
				SetGoVersion("go1.21.0")
				SetPlatform("linux/amd64")
			},
			expected: []string{
				"goUpdater 1.2.3",
				"├─ Commit: abc123",
				"├─ Built: October 1, 2023 at 12:00 PM UTC",
				"├─ Go version: go1.21.0",
				"└─ Platform: linux/amd64",
			},
		},
		{
			name: "minimal information",
			setupFunc: func() {
				resetGlobals()
				SetVersion("dev")
			},
			expected: []string{
				"goUpdater dev",
			},
		},
		{
			name: "invalid date format",
			setupFunc: func() {
				resetGlobals()
				SetVersion("1.0.0")
				SetDate("invalid-date")
				SetCommit("xyz789")
			},
			expected: []string{
				"goUpdater 1.0.0",
				"├─ Commit: xyz789",
				"├─ Built: invalid-date",
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testCase.setupFunc()

			// Capture stdout
			oldStdout := os.Stdout
			reader, writer, _ := os.Pipe()
			os.Stdout = writer
			_ = oldStdout

			Display(writer)

			_ = writer.Close()

			_ = oldStdout
			os.Stdout = oldStdout

			buf := make([]byte, 1024)
			n, _ := reader.Read(buf)
			output := string(buf[:n])

			for _, expected := range testCase.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Display() output missing expected substring: %s\nOutput: %s", expected, output)
				}
			}
		})
	}
}

// TestTimeParsing tests time parsing and formatting in Display.
func TestTimeParsing(t *testing.T) {
	t.Parallel()

	resetGlobals()

	SetVersion("1.0.0")
	SetDate("2023-12-25T15:30:45Z")
	SetCommit("test-commit")

	// Capture stdout
	oldStdout := os.Stdout
	reader, writer, _ := os.Pipe()
	os.Stdout = writer
	_ = oldStdout

	Display(writer)

	_ = writer.Close()

	_ = oldStdout
	os.Stdout = oldStdout

	buf := make([]byte, 1024)
	n, _ := reader.Read(buf)
	output := string(buf[:n])

	expected := "December 25, 2023 at 3:30 PM UTC"
	if !strings.Contains(output, expected) {
		t.Errorf("Display() time formatting failed. Expected: %s, Output: %s", expected, output)
	}
}

// TestEdgeCases tests various edge cases.
func TestEdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("empty version defaults to dev", func(t *testing.T) {
		t.Parallel()
		resetGlobals()

		if got := Get(); got != "dev" {
			t.Errorf("Get() with empty version = %s, want dev", got)
		}
	})

	t.Run("invalid time in date field", func(t *testing.T) {
		t.Parallel()
		resetGlobals()

		SetDate("not-a-date")

		info := GetVersionInfo()
		if info.Date != "not-a-date" {
			t.Errorf("Expected invalid date to be preserved, got %s", info.Date)
		}
	})

	t.Run("debug build info fallback simulation", func(t *testing.T) {
		t.Parallel()
		resetGlobals()

		SetVersion("1.0.0")
		// In a real scenario without ldflags for commit/date,
		// it would try debug.ReadBuildInfo, but we can't easily test that
		info := GetVersionInfo()
		if info.Version != "1.0.0" {
			t.Errorf("Version should be 1.0.0, got %s", info.Version)
		}
	})
}

// TestConcurrentAccess tests that getInfo is thread-safe.

func TestConcurrentAccess(t *testing.T) {
	t.Parallel()
	resetGlobals()

	SetVersion("1.0.0")
	SetCommit("test")

	done := make(chan bool, 2)

	go func() {
		_ = GetVersionInfo()

		done <- true
	}()

	go func() {
		_ = GetVersionInfo()

		done <- true
	}()

	<-done
	<-done
	// If no panic or race condition, test passes
}

// BenchmarkGetVersionInfo benchmarks the GetVersionInfo function.
func BenchmarkGetVersionInfo(b *testing.B) {
	resetGlobals()

	SetVersion("1.0.0")
	SetCommit("bench")
	SetDate("2023-01-01T00:00:00Z")

	for range b.N {
		_ = GetVersionInfo()
	}
}

// getBasicCompareTestCases returns basic test cases for TestCompare.
func getBasicCompareTestCases() []struct {
	name     string
	v1       string
	v2       string
	expected int
} {
	return []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		{
			name:     "equal versions",
			v1:       "1.21.0",
			v2:       "1.21.0",
			expected: 0,
		},
		{
			name:     "v1 greater than v2",
			v1:       "1.22.0",
			v2:       "1.21.0",
			expected: 1,
		},
		{
			name:     "v1 less than v2",
			v1:       "1.20.0",
			v2:       "1.21.0",
			expected: -1,
		},
		{
			name:     "missing patch version",
			v1:       "1.21",
			v2:       "1.21.0",
			expected: 0,
		},
		{
			name:     "different major versions",
			v1:       "2.0.0",
			v2:       "1.21.0",
			expected: 1,
		},
	}
}

// getAdvancedCompareTestCases returns advanced test cases for TestCompare.
func getAdvancedCompareTestCases() []struct {
	name     string
	v1       string
	v2       string
	expected int
} {
	return []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		{
			name:     "different minor versions",
			v1:       "1.20.0",
			v2:       "1.21.0",
			expected: -1,
		},
		{
			name:     "different patch versions",
			v1:       "1.21.1",
			v2:       "1.21.0",
			expected: 1,
		},
		{
			name:     "longer version string",
			v1:       "1.21.0.1",
			v2:       "1.21.0",
			expected: 1,
		},
		{
			name:     "shorter version string",
			v1:       "1.21",
			v2:       "1.21.0.1",
			expected: -1,
		},
	}
}

// getCompareTestCases returns all test cases for TestCompare.
func getCompareTestCases() []struct {
	name     string
	v1       string
	v2       string
	expected int
} {
	var allCases []struct {
		name     string
		v1       string
		v2       string
		expected int
	}

	allCases = append(allCases, getBasicCompareTestCases()...)
	allCases = append(allCases, getAdvancedCompareTestCases()...)

	return allCases
}

func TestCompare(t *testing.T) {
	t.Parallel()

	tests := getCompareTestCases()

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := Compare(testCase.v1, testCase.v2)
			if result != testCase.expected {
				t.Errorf("Compare(%s, %s) = %d, want %d", testCase.v1, testCase.v2, result, testCase.expected)
			}
		})
	}
}

// getFlagDetermineFormatTestCases returns flag-based test cases for TestDetermineFormat.
func getFlagDetermineFormatTestCases() []struct {
	name        string
	format      string
	jsonFlag    bool
	shortFlag   bool
	verboseFlag bool
	expected    outputFormat
} {
	return []struct {
		name        string
		format      string
		jsonFlag    bool
		shortFlag   bool
		verboseFlag bool
		expected    outputFormat
	}{
		{
			name:        "json flag",
			format:      "",
			jsonFlag:    true,
			shortFlag:   false,
			verboseFlag: false,
			expected:    formatJSON,
		},
		{
			name:        "short flag",
			format:      "",
			jsonFlag:    false,
			shortFlag:   true,
			verboseFlag: false,
			expected:    formatShort,
		},
		{
			name:        "verbose flag",
			format:      "",
			jsonFlag:    false,
			shortFlag:   false,
			verboseFlag: true,
			expected:    formatVerbose,
		},
	}
}

// getFormatStringDetermineFormatTestCases returns format string test cases for TestDetermineFormat.
func getFormatStringDetermineFormatTestCases() []struct {
	name        string
	format      string
	jsonFlag    bool
	shortFlag   bool
	verboseFlag bool
	expected    outputFormat
} {
	return []struct {
		name        string
		format      string
		jsonFlag    bool
		shortFlag   bool
		verboseFlag bool
		expected    outputFormat
	}{
		{
			name:        "format short",
			format:      "short",
			jsonFlag:    false,
			shortFlag:   false,
			verboseFlag: false,
			expected:    formatShort,
		},
		{
			name:        "format verbose",
			format:      "verbose",
			jsonFlag:    false,
			shortFlag:   false,
			verboseFlag: false,
			expected:    formatVerbose,
		},
		{
			name:        "format json",
			format:      "json",
			jsonFlag:    false,
			shortFlag:   false,
			verboseFlag: false,
			expected:    formatJSON,
		},
		{
			name:        "default format",
			format:      "unknown",
			jsonFlag:    false,
			shortFlag:   false,
			verboseFlag: false,
			expected:    formatDefault,
		},
		{
			name:        "empty format",
			format:      "",
			jsonFlag:    false,
			shortFlag:   false,
			verboseFlag: false,
			expected:    formatDefault,
		},
	}
}

// getDetermineFormatTestCases returns all test cases for TestDetermineFormat.
func getDetermineFormatTestCases() []struct {
	name        string
	format      string
	jsonFlag    bool
	shortFlag   bool
	verboseFlag bool
	expected    outputFormat
} {
	var allCases []struct {
		name        string
		format      string
		jsonFlag    bool
		shortFlag   bool
		verboseFlag bool
		expected    outputFormat
	}

	allCases = append(allCases, getFlagDetermineFormatTestCases()...)
	allCases = append(allCases, getFormatStringDetermineFormatTestCases()...)

	return allCases
}

func TestDetermineFormat(t *testing.T) {
	t.Parallel()

	tests := getDetermineFormatTestCases()

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := determineFormat(testCase.format, testCase.jsonFlag, testCase.shortFlag, testCase.verboseFlag)
			if result != testCase.expected {
				t.Errorf("determineFormat(%s, %v, %v, %v) = %v, want %v",
					testCase.format, testCase.jsonFlag, testCase.shortFlag, testCase.verboseFlag, result, testCase.expected)
			}
		})
	}
}

func TestDisplayShort(t *testing.T) {
	t.Parallel()
	resetGlobals()
	SetVersion("1.2.3")

	// Capture stdout
	oldStdout := os.Stdout
	reader, writer, _ := os.Pipe()
	os.Stdout = writer
	_ = oldStdout

	displayShort(os.Stdout, GetVersionInfo())

	_ = writer.Close()
	_ = oldStdout
	os.Stdout = oldStdout

	buf := make([]byte, 1024)
	n, _ := reader.Read(buf)
	output := string(buf[:n])

	expected := "1.2.3\n"
	if output != expected {
		t.Errorf("displayShort() output = %q, want %q", output, expected)
	}
}

func TestDisplayVerbose(t *testing.T) {
	t.Parallel()
	resetGlobals()
	SetVersion("1.2.3")
	SetCommit("abc123")
	SetDate("2023-10-01T12:00:00Z")
	SetGoVersion("go1.21.0")
	SetPlatform("linux/amd64")

	// Capture stdout
	oldStdout := os.Stdout
	reader, writer, _ := os.Pipe()
	os.Stdout = writer
	_ = oldStdout

	displayVerbose(os.Stdout, GetVersionInfo())

	_ = writer.Close()
	_ = oldStdout
	os.Stdout = oldStdout

	buf := make([]byte, 1024)
	n, _ := reader.Read(buf)
	output := string(buf[:n])

	expectedLines := []string{
		"Version: 1.2.3",
		"Commit: abc123",
		"Built: October 1, 2023 at 12:00 PM UTC",
		"Go version: go1.21.0",
		"Platform: linux/amd64",
	}

	for _, expected := range expectedLines {
		if !strings.Contains(output, expected) {
			t.Errorf("displayVerbose() output missing expected line: %s\nOutput: %s", expected, output)
		}
	}
}

func TestDisplayJSON(t *testing.T) {
	t.Parallel()
	resetGlobals()
	SetVersion("1.2.3")
	SetCommit("abc123")

	// Capture stdout
	oldStdout := os.Stdout
	reader, writer, _ := os.Pipe()
	os.Stdout = writer
	_ = oldStdout

	displayJSON(os.Stdout, GetVersionInfo())

	_ = writer.Close()
	_ = oldStdout
	os.Stdout = oldStdout

	buf := make([]byte, 1024)
	n, _ := reader.Read(buf)
	output := string(buf[:n])

	// Check if it's valid JSON and contains expected fields
	if !strings.Contains(output, `"version": "1.2.3"`) {
		t.Errorf("displayJSON() output missing version field: %s", output)
	}

	if !strings.Contains(output, `"commit": "abc123"`) {
		t.Errorf("displayJSON() output missing commit field: %s", output)
	}
}
