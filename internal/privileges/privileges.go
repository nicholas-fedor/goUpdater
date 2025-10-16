// Package privileges provides functions to detect privileges and request elevation using sudo.
// It handles privilege escalation for system operations that require root access.
package privileges

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

// IsRoot reports whether the current process is running as root.
func IsRoot() bool {
	return os.Geteuid() == 0
}

// RequestElevation re-executes the current process with sudo if not already running as root.
func RequestElevation() error {
	logger.Debug("Checking if elevation is needed")
	// Check if already running as root
	if IsRoot() {
		logger.Debug("Already running as root, no elevation needed")

		return nil
	}

	logger.Debug("Requesting elevation via sudo")

	// Get the path to the current executable
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	logger.Debugf("Executable path: %s", exePath)

	// Resolve any symlinks to get the actual executable path
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	logger.Debugf("Resolved executable path: %s", exePath)

	// Prepare the command arguments: sudo followed by the executable and original args
	args := append([]string{"sudo", exePath}, os.Args[1:]...)
	logger.Debugf("Sudo command args: %v", args)

	// Use syscall.Exec to replace the current process entirely with sudo
	// This is necessary for sudo to work properly and maintain the process environment
	// gosec: G204 - Subprocess launched with variable is acceptable here as we control the args
	logger.Debug("Executing with sudo")

	err = syscall.Exec("/usr/bin/sudo", args, os.Environ()) //nolint:gosec
	if err != nil {
		return fmt.Errorf("failed to execute with sudo: %w", err)
	}

	// This point should never be reached if syscall.Exec succeeds
	return nil
}

// HandleElevationError logs and exits with an error message for privilege elevation failures.
func HandleElevationError(err error) {
	logger.Errorf("Error: Failed to obtain elevated privileges: %v", err)
	logger.Error("Installation requires elevated privileges. Please run with sudo or as root.")
	os.Exit(1)
}

// ElevateAndExecute checks for root privileges and requests elevation if necessary.
// If elevation is required and fails, it handles the error appropriately.
// Once elevated or if already running as root, it executes the provided callback function
// and returns any error from the callback.
// The callback should be a function that performs the privileged operation.
func ElevateAndExecute(callback func() error) error {
	logger.Debug("Checking privileges for operation")

	if !IsRoot() {
		logger.Debug("Not running as root, requesting elevation")

		err := RequestElevation()
		if err != nil {
			HandleElevationError(err)
			// HandleElevationError exits, so this point is not reached
		}
		// If RequestElevation succeeds, the process is re-executed with elevation
		logger.Debug("Elevation request successful, process re-executed with sudo")

		return nil
	}

	logger.Debug("Already running as root")

	logger.Debug("Executing privileged operation")

	err := callback()
	if err != nil {
		logger.Errorf("Error executing privileged operation: %v", err)
	}

	return err
}

// RequestSudo is deprecated. Use RequestElevation instead.
func RequestSudo() error {
	return RequestElevation()
}
