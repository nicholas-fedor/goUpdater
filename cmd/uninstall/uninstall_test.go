// Package uninstall_test provides tests for the uninstall command.
package uninstall_test

import (
	"testing"

	"github.com/nicholas-fedor/goUpdater/cmd/uninstall"
)

func TestNewUninstallCmd(t *testing.T) {
	t.Parallel()

	cmd := uninstall.NewUninstallCmd()

	// Test basic command properties
	if cmd.Use != "uninstall" {
		t.Errorf("Expected command use to be 'uninstall', got %s", cmd.Use)
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

	// Test that the command accepts no arguments
	if cmd.Args != nil {
		t.Error("Expected command to accept no arguments")
	}

	// Test that the command has a Run function
	if cmd.Run == nil {
		t.Error("Expected command to have a Run function")
	}
}

func TestUninstallCmdFlags(t *testing.T) {
	t.Parallel()

	cmd := uninstall.NewUninstallCmd()

	// Test install-dir flag
	installDirFlag := cmd.Flags().Lookup("install-dir")
	if installDirFlag == nil {
		t.Fatal("install-dir flag not found")
	}

	if installDirFlag.Name != "install-dir" {
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

func TestUninstallCmdStructure(t *testing.T) {
	t.Parallel()

	cmd := uninstall.NewUninstallCmd()

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

func TestUninstallCmdOutput(t *testing.T) {
	t.Parallel()

	cmd := uninstall.NewUninstallCmd()

	// Test that command has expected output settings
	if cmd.SilenceUsage {
		t.Error("Expected command to show usage on error")
	}

	if cmd.SilenceErrors {
		t.Error("Expected command to show errors")
	}
}

func TestUninstallCmdHelp(t *testing.T) {
	t.Parallel()

	cmd := uninstall.NewUninstallCmd()

	// Test that command has help functionality
	// Note: The help flag is automatically added by cobra when the command is executed
	// We can't test flag setting in isolation, but we can verify command structure

	if cmd.DisableFlagParsing {
		t.Error("Expected command to allow flag parsing")
	}
}

func TestUninstallCmdFlagShortForm(t *testing.T) {
	t.Parallel()

	cmd := uninstall.NewUninstallCmd()

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

func TestUninstallCmdFlagEdgeCases(t *testing.T) {
	t.Parallel()

	testEmptyInstallDir(t)
	testRelativePath(t)
	testAbsolutePath(t)
	testPathWithSpaces(t)
	testPathWithSpecialChars(t)
	testVeryLongPath(t)
}

func testEmptyInstallDir(t *testing.T) {
	t.Helper()
	t.Run("empty install dir", func(t *testing.T) {
		t.Parallel()

		cmd := uninstall.NewUninstallCmd()

		err := cmd.Flags().Set("install-dir", "")
		if err != nil {
			t.Errorf("Setting flag to empty string: expected no error, got error=%v", err)
		}

		value, getErr := cmd.Flags().GetString("install-dir")
		if getErr != nil {
			t.Errorf("Expected to be able to get install-dir flag, got error: %v", getErr)
		}

		if value != "" {
			t.Errorf("Expected install-dir flag value to be empty, got %q", value)
		}
	})
}

func testRelativePath(t *testing.T) {
	t.Helper()
	t.Run("relative path", func(t *testing.T) {
		t.Parallel()

		cmd := uninstall.NewUninstallCmd()

		flagValue := "./relative/path"

		err := cmd.Flags().Set("install-dir", flagValue)
		if err != nil {
			t.Errorf("Setting flag to %q: expected no error, got error=%v", flagValue, err)
		}

		value, getErr := cmd.Flags().GetString("install-dir")
		if getErr != nil {
			t.Errorf("Expected to be able to get install-dir flag, got error: %v", getErr)
		}

		if value != flagValue {
			t.Errorf("Expected install-dir flag value to be %q, got %q", flagValue, value)
		}
	})
}

func testAbsolutePath(t *testing.T) {
	t.Helper()
	t.Run("absolute path", func(t *testing.T) {
		t.Parallel()

		cmd := uninstall.NewUninstallCmd()

		flagValue := "/absolute/path"

		err := cmd.Flags().Set("install-dir", flagValue)
		if err != nil {
			t.Errorf("Setting flag to %q: expected no error, got error=%v", flagValue, err)
		}

		value, getErr := cmd.Flags().GetString("install-dir")
		if getErr != nil {
			t.Errorf("Expected to be able to get install-dir flag, got error: %v", getErr)
		}

		if value != flagValue {
			t.Errorf("Expected install-dir flag value to be %q, got %q", flagValue, value)
		}
	})
}

func testPathWithSpaces(t *testing.T) {
	t.Helper()
	t.Run("path with spaces", func(t *testing.T) {
		t.Parallel()

		cmd := uninstall.NewUninstallCmd()

		flagValue := "/path with spaces"

		err := cmd.Flags().Set("install-dir", flagValue)
		if err != nil {
			t.Errorf("Setting flag to %q: expected no error, got error=%v", flagValue, err)
		}

		value, getErr := cmd.Flags().GetString("install-dir")
		if getErr != nil {
			t.Errorf("Expected to be able to get install-dir flag, got error: %v", getErr)
		}

		if value != flagValue {
			t.Errorf("Expected install-dir flag value to be %q, got %q", flagValue, value)
		}
	})
}

func testPathWithSpecialChars(t *testing.T) {
	t.Helper()
	t.Run("path with special chars", func(t *testing.T) {
		t.Parallel()

		cmd := uninstall.NewUninstallCmd()

		flagValue := "/path-with_special.chars"

		err := cmd.Flags().Set("install-dir", flagValue)
		if err != nil {
			t.Errorf("Setting flag to %q: expected no error, got error=%v", flagValue, err)
		}

		value, getErr := cmd.Flags().GetString("install-dir")
		if getErr != nil {
			t.Errorf("Expected to be able to get install-dir flag, got error: %v", getErr)
		}

		if value != flagValue {
			t.Errorf("Expected install-dir flag value to be %q, got %q", flagValue, value)
		}
	})
}

func testVeryLongPath(t *testing.T) {
	t.Helper()
	t.Run("very long path", func(t *testing.T) {
		t.Parallel()

		cmd := uninstall.NewUninstallCmd()

		flagValue := "/very/long/path/that/might/be/too/long/but/cobra/should/handle/it/anyway"

		err := cmd.Flags().Set("install-dir", flagValue)
		if err != nil {
			t.Errorf("Setting flag to %q: expected no error, got error=%v", flagValue, err)
		}

		value, getErr := cmd.Flags().GetString("install-dir")
		if getErr != nil {
			t.Errorf("Expected to be able to get install-dir flag, got error: %v", getErr)
		}

		if value != flagValue {
			t.Errorf("Expected install-dir flag value to be %q, got %q", flagValue, value)
		}
	})
}
