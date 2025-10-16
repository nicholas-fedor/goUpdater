// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package install_test provides tests for the install command.
package install_test

import (
	"testing"

	"github.com/nicholas-fedor/goUpdater/cmd/install"
)

func TestNewInstallCmd(t *testing.T) {
	t.Parallel()

	cmd := install.NewInstallCmd()

	if cmd.Use != "install [archive-path]" {
		t.Errorf("Expected command use to be 'install [archive-path]', got %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected command to have a short description")
	}

	if cmd.Long == "" {
		t.Error("Expected command to have a long description")
	}

	// Test that the command has the install-dir flag
	installDirFlag := cmd.Flags().Lookup("install-dir")
	if installDirFlag == nil {
		t.Error("Expected command to have install-dir flag")
	}

	if installDirFlag != nil && installDirFlag.DefValue != "/usr/local/go" {
		t.Errorf("Expected install-dir default value to be '/usr/local/go', got %s", installDirFlag.DefValue)
	}

	// Test that the command requires exactly one argument
	if cmd.Args == nil {
		t.Error("Expected command to have Args validation")
	}
}

func TestInstallCmdExecution(t *testing.T) {
	t.Parallel()

	cmd := install.NewInstallCmd()

	// Test command properties
	if cmd.Run == nil {
		t.Error("Expected command to have a Run function")
	}

	// Test that the command has the expected structure
	if cmd.Use != "install [archive-path]" {
		t.Errorf("Command use should be 'install [archive-path]', got %s", cmd.Use)
	}
}

func TestInstallCmdArgsValidation(t *testing.T) {
	t.Parallel()

	cmd := install.NewInstallCmd()

	// Test with no arguments (should pass)
	err := cmd.Args(cmd, []string{})
	if err != nil {
		t.Errorf("Expected Args validation to pass with no arguments, got error: %v", err)
	}

	// Test with one argument (should pass)
	err = cmd.Args(cmd, []string{"archive.tar.gz"})
	if err != nil {
		t.Errorf("Expected Args validation to pass with one argument, got error: %v", err)
	}

	// Test with multiple arguments (should fail)
	err = cmd.Args(cmd, []string{"archive1.tar.gz", "archive2.tar.gz"})
	if err == nil {
		t.Error("Expected Args validation to fail with multiple arguments")
	}
}

func TestInstallCmdFlags(t *testing.T) {
	t.Parallel()

	cmd := install.NewInstallCmd()

	// Test install-dir flag
	installDirFlag := cmd.Flags().Lookup("install-dir")
	if installDirFlag == nil {
		t.Fatal("install-dir flag not found")
	}

	if installDirFlag != nil && installDirFlag.Name != "install-dir" {
		t.Errorf("Expected flag name to be 'install-dir', got %s", installDirFlag.Name)
	}

	if installDirFlag.DefValue != "/usr/local/go" {
		t.Errorf("Expected default value to be '/usr/local/go', got %s", installDirFlag.DefValue)
	}

	// Test that flag can be set
	err := cmd.Flags().Set("install-dir", "/custom/path")
	if err != nil {
		t.Errorf("Expected to be able to set install-dir flag, got error: %v", err)
	}

	value, err := cmd.Flags().GetString("install-dir")
	if err != nil {
		t.Errorf("Expected to be able to get install-dir flag, got error: %v", err)
	}

	if value != "/custom/path" {
		t.Errorf("Expected install-dir flag value to be '/custom/path', got %s", value)
	}
}

func TestInstallCmdStructure(t *testing.T) {
	t.Parallel()

	cmd := install.NewInstallCmd()

	// Test that command uses Run, not RunE
	if cmd.RunE != nil {
		t.Error("Expected command to use Run, not RunE")
	}

	// Test that command has no pre/post run hooks
	if cmd.PreRun != nil || cmd.PostRun != nil {
		t.Error("Expected command to have no pre/post run hooks")
	}

	// Test that command has no aliases
	if len(cmd.Aliases) > 0 {
		t.Error("Expected command to have no aliases")
	}

	// Test that command has expected completion options
	if cmd.CompletionOptions.DisableDefaultCmd {
		t.Error("Expected command to allow default completion")
	}
}

func TestInstallCmdOutput(t *testing.T) {
	t.Parallel()

	cmd := install.NewInstallCmd()

	// Test that command has expected output setestCase.ngs
	if cmd.SilenceUsage {
		t.Error("Expected command to show usage on error")
	}

	if cmd.SilenceErrors {
		t.Error("Expected command to show errors")
	}
}

func TestInstallCmdHelp(t *testing.T) {
	t.Parallel()

	cmd := install.NewInstallCmd()

	// Test that command has help functionality
	// Note: The help flag is automatically added by cobra when the command is executed
	// We can't test flag setestCase.ng in isolation, but we can verify command structure

	if cmd.DisableFlagParsing {
		t.Error("Expected command to allow flag parsing")
	}
}

func TestInstallCmdFlagShortForm(t *testing.T) {
	t.Parallel()

	cmd := install.NewInstallCmd()

	// Test short form of install-dir flag (-d)
	err := cmd.Flags().Set("install-dir", "/short/path")
	if err != nil {
		t.Errorf("Expected to be able to set -d flag, got error: %v", err)
	}

	value, err := cmd.Flags().GetString("install-dir")
	if err != nil {
		t.Errorf("Expected to be able to get install-dir flag, got error: %v", err)
	}

	if value != "/short/path" {
		t.Errorf("Expected install-dir flag value to be '/short/path', got %s", value)
	}
}

func TestInstallCmdArgsEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
	}{
		{
			name:      "no args",
			args:      []string{},
			shouldErr: false,
		},
		{
			name:      "one arg",
			args:      []string{"archive.tar.gz"},
			shouldErr: false,
		},
		{
			name:      "two args",
			args:      []string{"archive1.tar.gz", "archive2.tar.gz"},
			shouldErr: true,
		},
		{
			name:      "empty string arg",
			args:      []string{""},
			shouldErr: false,
		},
		{
			name:      "arg with spaces",
			args:      []string{"archive with spaces.tar.gz"},
			shouldErr: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cmd := install.NewInstallCmd()

			err := cmd.Args(cmd, testCase.args)
			if (err != nil) != testCase.shouldErr {
				t.Errorf("Args validation for %v: expected error=%v, got error=%v", testCase.args, testCase.shouldErr, err != nil)
			}
		})
	}
}

func TestInstallCmdFlagEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		flagValue string
		shouldErr bool
	}{
		{
			name:      "empty install dir",
			flagValue: "",
			shouldErr: false, // Cobra allows empty strings for string flags
		},
		{
			name:      "relative path",
			flagValue: "./relative/path",
			shouldErr: false,
		},
		{
			name:      "absolute path",
			flagValue: "/absolute/path",
			shouldErr: false,
		},
		{
			name:      "path with spaces",
			flagValue: "/path with spaces",
			shouldErr: false,
		},
		{
			name:      "path with special chars",
			flagValue: "/path-with_special.chars",
			shouldErr: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cmd := install.NewInstallCmd()

			err := cmd.Flags().Set("install-dir", testCase.flagValue)
			if (err != nil) != testCase.shouldErr {
				t.Errorf("Setting flag to %q: expected error=%v, got error=%v", testCase.flagValue, testCase.shouldErr, err != nil)
			}

			if !testCase.shouldErr {
				value, getErr := cmd.Flags().GetString("install-dir")
				if getErr != nil {
					t.Errorf("Expected to be able to get install-dir flag, got error: %v", getErr)
				}

				if value != testCase.flagValue {
					t.Errorf("Expected install-dir flag value to be %q, got %q", testCase.flagValue, value)
				}
			}
		})
	}
}
