// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package download_test provides tests for the download command.
package download_test

import (
	"testing"

	"github.com/nicholas-fedor/goUpdater/cmd/download"
)

func TestNewDownloadCmd(t *testing.T) {
	t.Parallel()

	cmd := download.NewDownloadCmd()

	// Test basic command properties
	if cmd.Use != "download" {
		t.Errorf("Expected command use to be 'download', got %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected command to have a short description")
	}

	if cmd.Long == "" {
		t.Error("Expected command to have a long description")
	}

	// Test that no custom flags are defined
	// The download command should not have any custom flags
	flags := cmd.Flags()
	if flags.Lookup("verbose") != nil {
		t.Error("Expected no verbose flag on download command")
	}

	if flags.Lookup("install-dir") != nil {
		t.Error("Expected no install-dir flag on download command")
	}
}

func TestDownloadCmdExecution(t *testing.T) {
	t.Parallel()

	cmd := download.NewDownloadCmd()

	// Test command properties
	if cmd.Run == nil {
		t.Error("Expected command to have a Run function")
	}

	// Test that command has expected properties
	if cmd.Use != "download" {
		t.Errorf("Command use should be 'download', got %s", cmd.Use)
	}

	// Test that command has no arguments (download takes no args)
	if cmd.Args != nil {
		t.Error("Expected command to accept no arguments")
	}
}

func TestDownloadCmdStructure(t *testing.T) {
	t.Parallel()

	cmd := download.NewDownloadCmd()

	// Test that command has expected structure
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
}

func TestDownloadCmdArgs(t *testing.T) {
	t.Parallel()

	cmd := download.NewDownloadCmd()

	// Test that command accepts no arguments (Args is nil)
	if cmd.Args != nil {
		t.Error("Expected command to accept no arguments")
	}

	// Since Args is nil, we can't call cmd.Args() - cobra will handle argument validation
	// The command should accept zero arguments by default when Args is nil
}

func TestDownloadCmdFlags(t *testing.T) {
	t.Parallel()

	cmd := download.NewDownloadCmd()

	// Test that no custom flags are defined (only inherited ones)
	flags := cmd.Flags()

	// Note: Cobra commands automatically inherit help flag, but it's not added until the command is executed
	// We can't test for the help flag existence in this context

	// Test that no custom flags are defined
	if flags.Lookup("verbose") != nil {
		t.Error("Expected no verbose flag on download command")
	}

	if flags.Lookup("version") != nil {
		t.Error("Expected no version flag on download command")
	}

	if flags.Lookup("install-dir") != nil {
		t.Error("Expected no install-dir flag on download command")
	}
}

func TestDownloadCmdOutput(t *testing.T) {
	t.Parallel()

	cmd := download.NewDownloadCmd()

	// Test that command has expected output settings
	if cmd.SilenceUsage {
		t.Error("Expected command to show usage on error")
	}

	if cmd.SilenceErrors {
		t.Error("Expected command to show errors")
	}
}

func TestDownloadCmdHelp(t *testing.T) {
	t.Parallel()

	cmd := download.NewDownloadCmd()

	// Test that command has help functionality
	// Note: The help flag is automatically added by cobra when the command is executed
	// We can't test flag setting in isolation, but we can verify command structure

	if cmd.DisableFlagParsing {
		t.Error("Expected command to allow flag parsing")
	}
}
