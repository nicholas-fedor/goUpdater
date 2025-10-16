// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package version_test provides tests for the version command.
package version_test

import (
	"os"
	"strings"
	"testing"

	"github.com/nicholas-fedor/goUpdater/cmd/version"
	versionpkg "github.com/nicholas-fedor/goUpdater/internal/version"
)

func TestNewVersionCmd(t *testing.T) {
	t.Parallel()

	cmd := version.NewVersionCmd()

	// Test basic command properties
	if cmd.Use != "version" {
		t.Errorf("Expected command use to be 'version', got %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected command to have a short description")
	}

	if cmd.Long == "" {
		t.Error("Expected command to have a long description")
	}

	// Test that expected flags are defined
	flags := cmd.Flags()
	if flags.Lookup("format") == nil {
		t.Error("Expected format flag to be defined")
	}

	if flags.Lookup("json") == nil {
		t.Error("Expected json flag to be defined")
	}

	if flags.Lookup("short") == nil {
		t.Error("Expected short flag to be defined")
	}

	if flags.Lookup("verbose") == nil {
		t.Error("Expected verbose flag to be defined")
	}
}

func TestVersionCmdStructure(t *testing.T) {
	t.Parallel()

	cmd := version.NewVersionCmd()

	// Test that command has expected structure
	if cmd.RunE != nil {
		t.Error("Expected command to use Run, not RunE")
	}

	if cmd.Run == nil {
		t.Error("Expected command to have a Run function")
	}

	// Test that command has no pre/post run hooks
	if cmd.PreRun != nil || cmd.PostRun != nil {
		t.Error("Expected command to have no pre/post run hooks")
	}

	// Test that command has no aliases
	if len(cmd.Aliases) > 0 {
		t.Error("Expected command to have no aliases")
	}

	// Test that command accepts no arguments
	if cmd.Args != nil {
		t.Error("Expected command to accept no arguments")
	}
}

func TestVersionCmdFlags(t *testing.T) {
	t.Parallel()

	cmd := version.NewVersionCmd()
	flags := cmd.Flags()

	// Test format flag properties
	formatFlag := flags.Lookup("format")
	if formatFlag == nil {
		t.Fatal("format flag not found")
	}

	if formatFlag.DefValue != "" {
		t.Errorf("Expected format flag default value to be empty, got %s", formatFlag.DefValue)
	}

	// Test json flag properties
	jsonFlag := flags.Lookup("json")
	if jsonFlag == nil {
		t.Fatal("json flag not found")
	}

	const defaultFlagValue = "false"

	if jsonFlag.DefValue != defaultFlagValue {
		t.Errorf("Expected json flag default value to be false, got %s", jsonFlag.DefValue)
	}

	// Test short flag properties
	shortFlag := flags.Lookup("short")
	if shortFlag == nil {
		t.Fatal("short flag not found")
	}

	if shortFlag.DefValue != defaultFlagValue {
		t.Errorf("Expected short flag default value to be false, got %s", shortFlag.DefValue)
	}

	// Test verbose flag properties
	verboseFlag := flags.Lookup("verbose")
	if verboseFlag == nil {
		t.Fatal("verbose flag not found")
	}

	if verboseFlag.DefValue != defaultFlagValue {
		t.Errorf("Expected verbose flag default value to be false, got %s", verboseFlag.DefValue)
	}
}

func TestVersionCmdExecutionDefault(t *testing.T) {
	t.Parallel()
	// Reset version info
	resetVersionGlobals()

	versionpkg.SetVersion("1.2.3")
	versionpkg.SetCommit("abc123")
	versionpkg.SetDate("2023-10-01T12:00:00Z")
	versionpkg.SetGoVersion("go1.21.0")
	versionpkg.SetPlatform("linux/amd64")

	cmd := version.NewVersionCmd()

	// Capture stdout
	oldStdout := os.Stdout
	reader, writer, _ := os.Pipe()
	os.Stdout = writer

	// Execute command with default format
	cmd.Run(cmd, []string{})

	_ = writer.Close()

	os.Stdout = oldStdout

	buf := make([]byte, 1024)
	n, _ := reader.Read(buf)
	output := string(buf[:n])

	// Check expected output for default format
	expected := []string{
		"goUpdater 1.2.3",
		"├─ Commit: abc123",
		"├─ Built: October 1, 2023 at 12:00 PM UTC",
		"├─ Go version: go1.21.0",
		"└─ Platform: linux/amd64",
	}

	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Default format output missing expected substring: %s\nOutput: %s", exp, output)
		}
	}
}

func TestVersionCmdExecutionShort(t *testing.T) {
	t.Parallel()
	resetVersionGlobals()
	versionpkg.SetVersion("1.2.3")

	cmd := version.NewVersionCmd()

	// Set short flag
	_ = cmd.Flags().Set("short", "true")

	// Capture stdout
	oldStdout := os.Stdout
	reader, writer, _ := os.Pipe()
	os.Stdout = writer

	cmd.Run(cmd, []string{})

	_ = writer.Close()

	os.Stdout = oldStdout

	buf := make([]byte, 1024)
	n, _ := reader.Read(buf)
	output := string(buf[:n])

	expected := "1.2.3\n"
	if output != expected {
		t.Errorf("Short format output = %q, want %q", output, expected)
	}
}

func TestVersionCmdExecutionVerbose(t *testing.T) {
	t.Parallel()
	resetVersionGlobals()
	versionpkg.SetVersion("1.2.3")
	versionpkg.SetCommit("abc123")
	versionpkg.SetDate("2023-10-01T12:00:00Z")
	versionpkg.SetGoVersion("go1.21.0")
	versionpkg.SetPlatform("linux/amd64")

	cmd := version.NewVersionCmd()

	// Set verbose flag
	_ = cmd.Flags().Set("verbose", "true")

	// Capture stdout
	oldStdout := os.Stdout
	reader, writer, _ := os.Pipe()
	os.Stdout = writer

	cmd.Run(cmd, []string{})

	_ = writer.Close()

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
			t.Errorf("Verbose format output missing expected line: %s\nOutput: %s", expected, output)
		}
	}
}

func TestVersionCmdExecutionJSON(t *testing.T) {
	t.Parallel()
	resetVersionGlobals()
	versionpkg.SetVersion("1.2.3")
	versionpkg.SetCommit("abc123")

	cmd := version.NewVersionCmd()

	// Set json flag
	_ = cmd.Flags().Set("json", "true")

	// Capture stdout
	oldStdout := os.Stdout
	reader, writer, _ := os.Pipe()
	os.Stdout = writer

	cmd.Run(cmd, []string{})

	_ = writer.Close()

	os.Stdout = oldStdout

	buf := make([]byte, 1024)
	n, _ := reader.Read(buf)
	output := string(buf[:n])

	// Check if output contains expected JSON fields
	if !strings.Contains(output, `"version": "1.2.3"`) {
		t.Errorf("JSON output missing version field: %s", output)
	}

	if !strings.Contains(output, `"commit": "abc123"`) {
		t.Errorf("JSON output missing commit field: %s", output)
	}

	// Verify it's valid JSON by checking for braces and quotes
	if !strings.Contains(output, "{") || !strings.Contains(output, "}") {
		t.Errorf("Output does not appear to be valid JSON: %s", output)
	}
}

func TestVersionCmdExecutionFormatFlag(t *testing.T) {
	t.Parallel()
	resetVersionGlobals()
	versionpkg.SetVersion("1.2.3")

	cmd := version.NewVersionCmd()

	// Set format flag to short
	_ = cmd.Flags().Set("format", "short")

	// Capture stdout
	oldStdout := os.Stdout
	reader, writer, _ := os.Pipe()
	os.Stdout = writer

	cmd.Run(cmd, []string{})

	_ = writer.Close()

	os.Stdout = oldStdout

	buf := make([]byte, 1024)
	n, _ := reader.Read(buf)
	output := string(buf[:n])

	expected := "1.2.3\n"
	if output != expected {
		t.Errorf("Format flag 'short' output = %q, want %q", output, expected)
	}
}

func TestVersionCmdFlagPrecedence(t *testing.T) {
	t.Parallel()
	resetVersionGlobals()
	versionpkg.SetVersion("1.2.3")

	cmd := version.NewVersionCmd()

	// Set format to verbose but json flag to true - json should take precedence
	_ = cmd.Flags().Set("format", "verbose")
	_ = cmd.Flags().Set("json", "true")

	// Capture stdout
	oldStdout := os.Stdout
	reader, writer, _ := os.Pipe()
	os.Stdout = writer

	cmd.Run(cmd, []string{})

	_ = writer.Close()

	os.Stdout = oldStdout

	buf := make([]byte, 1024)
	n, _ := reader.Read(buf)
	output := string(buf[:n])

	// Should output JSON, not verbose
	if !strings.Contains(output, `"version": "1.2.3"`) {
		t.Errorf("JSON flag should take precedence over format flag. Output: %s", output)
	}

	if strings.Contains(output, "Version: 1.2.3") {
		t.Errorf("Verbose format should not be used when json flag is set. Output: %s", output)
	}
}

func TestVersionCmdInvalidFormat(t *testing.T) {
	t.Parallel()
	resetVersionGlobals()
	versionpkg.SetVersion("1.2.3")
	versionpkg.SetCommit("abc123")

	cmd := version.NewVersionCmd()

	// Set invalid format - should default to default format
	_ = cmd.Flags().Set("format", "invalid")

	// Capture stdout
	oldStdout := os.Stdout
	reader, writer, _ := os.Pipe()
	os.Stdout = writer

	cmd.Run(cmd, []string{})

	_ = writer.Close()

	os.Stdout = oldStdout

	buf := make([]byte, 1024)
	n, _ := reader.Read(buf)
	output := string(buf[:n])

	// Should use default format
	if !strings.Contains(output, "goUpdater 1.2.3") {
		t.Errorf("Invalid format should default to default format. Output: %s", output)
	}
}

func TestVersionCmdMinimalInfo(t *testing.T) {
	t.Parallel()
	resetVersionGlobals()
	versionpkg.SetVersion("dev")

	cmd := version.NewVersionCmd()

	// Capture stdout
	oldStdout := os.Stdout
	reader, writer, _ := os.Pipe()
	os.Stdout = writer

	cmd.Run(cmd, []string{})

	_ = writer.Close()

	os.Stdout = oldStdout

	buf := make([]byte, 1024)
	n, _ := reader.Read(buf)
	output := string(buf[:n])

	expected := "goUpdater dev\n"
	if output != expected {
		t.Errorf("Minimal info output = %q, want %q", output, expected)
	}
}

func TestVersionCmdEmptyVersion(t *testing.T) {
	t.Parallel()
	resetVersionGlobals()
	// Version not set - should default to "dev"

	cmd := version.NewVersionCmd()

	// Capture stdout
	oldStdout := os.Stdout
	reader, writer, _ := os.Pipe()
	os.Stdout = writer

	cmd.Run(cmd, []string{})

	_ = writer.Close()

	os.Stdout = oldStdout

	buf := make([]byte, 1024)
	n, _ := reader.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "goUpdater dev") {
		t.Errorf("Empty version should default to dev. Output: %s", output)
	}
}

// resetVersionGlobals resets all version global variables for testing.
func resetVersionGlobals() {
	versionpkg.SetVersion("")
	versionpkg.SetCommit("")
	versionpkg.SetDate("")
	versionpkg.SetGoVersion("")
	versionpkg.SetPlatform("")
}
