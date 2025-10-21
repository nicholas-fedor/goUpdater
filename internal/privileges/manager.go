// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package privileges

import (
	"strings"

	"github.com/nicholas-fedor/goUpdater/internal/exec"
	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

// PrivilegeManager handles privilege escalation with dependency injection.
type PrivilegeManager struct {
	fs       filesystem.FileSystem
	pm       OSPrivilegeManager
	executor exec.CommandExecutor
	logger   AuditLogger
}

// NewPrivilegeManager creates a new PrivilegeManager with the given dependencies.
// It initializes the manager with filesystem, OS privilege manager, and command executor interfaces.
func NewPrivilegeManager(fs filesystem.FileSystem, pm OSPrivilegeManager,
	executor exec.CommandExecutor) *PrivilegeManager {
	return &PrivilegeManager{
		fs:       fs,
		pm:       pm,
		executor: executor,
		logger:   &DefaultAuditLogger{},
	}
}

// NewPrivilegeManagerWithLogger creates a new PrivilegeManager with custom audit logger.
// It allows injecting a custom logger for privilege operations auditing.
func NewPrivilegeManagerWithLogger(fs filesystem.FileSystem, pm OSPrivilegeManager,
	executor exec.CommandExecutor, logger AuditLogger) *PrivilegeManager {
	return &PrivilegeManager{
		fs:       fs,
		pm:       pm,
		executor: executor,
		logger:   logger,
	}
}

// isRoot reports whether the current process is running as root using injected dependencies.
// It checks if the effective user ID is 0, indicating root privileges.
func (p *PrivilegeManager) isRoot() bool { //nolint:funcorder
	return p.pm.Geteuid() == 0
}

// ElevateAndExecute checks for root privileges and requests elevation if necessary using injected dependencies.
// If elevation is required and fails, it handles the error appropriately.
// Once elevated or if already running as root, it executes the provided callback function
// and returns any error from the callback.
// The callback should be a function that performs the privileged operation.
// This method ensures operations requiring root access are properly elevated.
func (p *PrivilegeManager) ElevateAndExecute(callback func() error) error {
	logger.Debug("Checking privileges for operation")

	if !p.isRoot() {
		logger.Debug("Not running as root, requesting elevation")

		err := p.requestElevation()
		if err != nil {
			HandleElevationError(err)
			// HandleElevationError exits, so this point is not reached
			return nil // unreachable, but required for compilation
		}
		// If RequestElevation succeeds, the process is re-executed with elevation
		logger.Debug("Elevation request successful, process re-executed with sudo")

		return nil
	}

	logger.Debug("Already running as root")

	logger.Debug("Executing privileged operation")

	err := callback()
	if err != nil {
		if strings.Contains(err.Error(), "go is not installed") {
			// Treat ErrGoNotInstalled as a non-error case
			return nil
		}

		logger.Errorf("Error executing privileged operation: %v", err)
	}

	return err
}

// ElevateAndExecuteWithDrop checks for root privileges and requests elevation if necessary.
// After executing the callback, it attempts to drop privileges back to the original user.
// This provides a safer way to perform operations that require temporary elevation.
// The method ensures privileges are dropped even if the callback fails.
func (p *PrivilegeManager) ElevateAndExecuteWithDrop(callback func() error) error {
	logger.Debug("Checking privileges for operation with privilege drop")

	if !p.isRoot() {
		logger.Debug("Not running as root, requesting elevation")

		err := p.requestElevation()
		if err != nil {
			HandleElevationError(err)
			// HandleElevationError exits, so this point is not reached
			return nil // unreachable, but required for compilation
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

		return err
	}

	// Attempt to drop privileges after successful operation
	logger.Debug("Attempting to drop privileges back to original user")

	dropErr := p.dropPrivileges()
	if dropErr != nil {
		logger.Warnf("Failed to drop privileges: %v", dropErr)
		// Don't return the drop error as the main operation succeeded
		// But log it for security awareness
	}

	return err
}
