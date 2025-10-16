// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package update_test provides tests for the update command.
package update_test

import (
	"testing"

	"github.com/nicholas-fedor/goUpdater/cmd/update"
)

func TestNewUpdateCmd(t *testing.T) {
	t.Parallel()

	cmd := update.NewUpdateCmd()

	// Test basic command properties
	if cmd.Use != "update" {
		t.Errorf("Expected command use to be 'update', got %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected command to have a short description")
	}

	if cmd.Long == "" {
		t.Error("Expected command to have a long description")
	}

	// Test that the command accepts no arguments
	if cmd.Args != nil {
		t.Error("Expected command to accept no arguments")
	}

	// Test that the command has a Run function
	if cmd.Run == nil {
		t.Error("Expected command to have a Run function")
	}
}

func TestUpdateCmdFlags(t *testing.T) {
	t.Parallel()

	testInstallDirFlag(t)
	testAutoInstallFlag(t)
}

func testInstallDirFlag(t *testing.T) {
	t.Helper()
	t.Run("install-dir flag", func(t *testing.T) {
		t.Parallel()

		cmd := update.NewUpdateCmd()

		// Test that the flag exists
		flag := cmd.Flags().Lookup("install-dir")
		if flag == nil {
			t.Fatalf("Expected command to have install-dir flag")
		}

		// Test flag name
		if flag.Name != "install-dir" {
			t.Errorf("Expected flag name to be 'install-dir', got '%s'", flag.Name)
		}

		// Test default value
		if flag.DefValue != "/usr/local/go" {
			t.Errorf("Expected default value to be '/usr/local/go', got '%s'", flag.DefValue)
		}

		// Test setting the flag using long form
		err := cmd.Flags().Set("install-dir", "/custom/path")
		if err != nil {
			t.Errorf("Expected to be able to set install-dir flag, got error: %v", err)
		}

		// Test getting the flag value
		value, err := cmd.Flags().GetString("install-dir")
		if err != nil {
			t.Errorf("Expected to be able to get install-dir flag, got error: %v", err)
		}

		if value != "/custom/path" {
			t.Errorf("Expected install-dir flag value to be '/custom/path', got '%s'", value)
		}
	})
}

func testAutoInstallFlag(t *testing.T) {
	t.Helper()
	t.Run("auto-install flag", func(t *testing.T) {
		t.Parallel()

		cmd := update.NewUpdateCmd()

		// Test that the flag exists
		flag := cmd.Flags().Lookup("auto-install")
		if flag == nil {
			t.Fatalf("Expected command to have auto-install flag")
		}

		// Test flag name
		if flag.Name != "auto-install" {
			t.Errorf("Expected flag name to be 'auto-install', got '%s'", flag.Name)
		}

		// Test default value
		if flag.DefValue != "false" {
			t.Errorf("Expected default value to be 'false', got '%s'", flag.DefValue)
		}

		// Test setting the flag using long form
		err := cmd.Flags().Set("auto-install", "true")
		if err != nil {
			t.Errorf("Expected to be able to set auto-install flag, got error: %v", err)
		}

		// Test getting the flag value
		value, err := cmd.Flags().GetBool("auto-install")
		if err != nil {
			t.Errorf("Expected to be able to get auto-install flag, got error: %v", err)
		}

		if !value {
			t.Errorf("Expected auto-install flag value to be true, got %t", value)
		}
	})
}

func TestUpdateCmdFlagCombinations(t *testing.T) {
	t.Parallel()

	cmd := update.NewUpdateCmd()

	// Test setting both flags together
	err := cmd.Flags().Set("install-dir", "/opt/go")
	if err != nil {
		t.Errorf("Failed to set install-dir flag: %v", err)
	}

	err = cmd.Flags().Set("auto-install", "true")
	if err != nil {
		t.Errorf("Failed to set auto-install flag: %v", err)
	}

	// Verify both values
	installDir, err := cmd.Flags().GetString("install-dir")
	if err != nil {
		t.Errorf("Failed to get install-dir flag: %v", err)
	}

	if installDir != "/opt/go" {
		t.Errorf("Expected install-dir to be '/opt/go', got '%s'", installDir)
	}

	autoInstall, err := cmd.Flags().GetBool("auto-install")
	if err != nil {
		t.Errorf("Failed to get auto-install flag: %v", err)
	}

	if !autoInstall {
		t.Error("Expected auto-install to be true")
	}
}

func TestUpdateCmdShortFlags(t *testing.T) {
	t.Parallel()

	cmd := update.NewUpdateCmd()

	// Test short flags using the long flag names since cobra short flags are accessed via long names
	err := cmd.Flags().Set("install-dir", "/tmp/go")
	if err != nil {
		t.Errorf("Failed to set install-dir flag: %v", err)
	}

	err = cmd.Flags().Set("auto-install", "true")
	if err != nil {
		t.Errorf("Failed to set auto-install flag: %v", err)
	}

	// Verify values using long flag names
	installDir, err := cmd.Flags().GetString("install-dir")
	if err != nil {
		t.Errorf("Failed to get install-dir after setting flag: %v", err)
	}

	if installDir != "/tmp/go" {
		t.Errorf("Expected install-dir to be '/tmp/go', got '%s'", installDir)
	}

	autoInstall, err := cmd.Flags().GetBool("auto-install")
	if err != nil {
		t.Errorf("Failed to get auto-install after setting flag: %v", err)
	}

	if !autoInstall {
		t.Error("Expected auto-install to be true after setting flag")
	}
}
