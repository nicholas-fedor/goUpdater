// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package privileges

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/nicholas-fedor/goUpdater/internal/exec"
	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

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
			return HandleElevationError(err)
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
			return HandleElevationError(err)
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

// isRoot reports whether the current process is running as root using injected dependencies.
// It checks if the effective user ID is 0, indicating root privileges.
func (p *PrivilegeManager) isRoot() bool {
	return p.pm.Geteuid() == 0
}

// requestElevation re-executes the current process with sudo if not already running as root.
// It validates sudo availability and executable integrity before attempting elevation.
// The method uses syscall.Exec to replace the current process entirely with sudo.
// This ensures proper environment preservation and sudo functionality.
//
//nolint:funlen // Complex elevation logic requires multiple validation steps
func (p *PrivilegeManager) requestElevation() error {
	logger.Debug("Checking if privilege elevation is needed")
	// Check if already running as root
	if p.isRoot() {
		logger.Debug("Already running as root, no elevation needed")
		p.logger.LogElevationAttempt(false, "already root - no elevation needed")

		return nil
	}

	logger.Debug("Initiating privilege elevation via sudo")

	// Validate sudo availability
	logger.Debug("Validating sudo availability")

	err := p.validateSudoAvailability()
	if err != nil {
		p.logger.LogElevationAttempt(false, "sudo not available")

		return err
	}

	logger.Debug("Sudo availability validated successfully")

	// Validate executable
	logger.Debug("Validating executable integrity")

	exePath, err := p.validateExecutable()
	if err != nil {
		p.logger.LogElevationAttempt(false, "executable validation failed")

		return err
	}

	logger.Debug("Executable validation completed successfully")

	// Sanitize command-line arguments to prevent privilege escalation
	logger.Debug("Sanitizing command-line arguments")

	sanitizedArgs, err := sanitizeArgs(p.pm.Args()[1:])
	if err != nil {
		return &ElevationError{
			Op:      "sanitize_args",
			Reason:  "argument sanitization failed",
			Cause:   err,
			SudoErr: false,
		}
	}

	logger.Debug("Command-line arguments sanitized successfully")

	// Prepare the command arguments: sudo followed by the executable and sanitized args
	args := append([]string{"sudo", exePath}, sanitizedArgs...)

	logger.Debug("Sudo command prepared for execution")

	// Log elevation attempt
	p.logger.LogElevationAttempt(true, "attempting sudo elevation")

	// Use syscall.Exec to replace the current process entirely with sudo
	// This is necessary for sudo to work properly and maintain the process environment
	// gosec: G204 - Subprocess launched with variable is acceptable here as we control the args
	logger.Debug("Executing process with elevated privileges")

	err = p.pm.Exec("/usr/bin/sudo", args, p.pm.Environ())
	if err != nil {
		elevErr := &ElevationError{
			Op:      "execute_sudo",
			Reason:  "sudo execution failed",
			Cause:   err,
			SudoErr: true,
		}

		p.logger.LogElevationAttempt(false, "sudo execution failed")

		return elevErr
	}

	// This point should never be reached if syscall.Exec succeeds
	return nil
}

// validateSudoAvailability checks if sudo is available and executable.
// It verifies that the sudo binary exists at the standard location and has execute permissions.
func (p *PrivilegeManager) validateSudoAvailability() error {
	// Check if sudo exists and is executable
	info, err := p.fs.Stat("/usr/bin/sudo")
	if err != nil {
		return &ElevationError{
			Op:      "validate_sudo",
			Reason:  "sudo binary not found at /usr/bin/sudo",
			Cause:   err,
			SudoErr: true,
		}
	}

	if info.Mode()&0111 == 0 {
		return &ElevationError{
			Op:      "validate_sudo",
			Reason:  "sudo binary is not executable",
			Cause:   nil,
			SudoErr: true,
		}
	}

	return nil
}

// validateExecutable checks if the current executable is valid for elevation.
// It ensures the executable exists, is accessible, and resolves any symlinks for proper sudo execution.
func (p *PrivilegeManager) validateExecutable() (string, error) {
	exePath, err := p.pm.Executable()
	if err != nil {
		return "", &ElevationError{
			Op:      "validate_executable",
			Reason:  "failed to get executable path",
			Cause:   err,
			SudoErr: false,
		}
	}

	// Check if executable exists and is accessible
	info, err := p.fs.Stat(exePath)
	if err != nil {
		return "", &ElevationError{
			Op:      "validate_executable",
			Reason:  "executable not accessible: " + exePath,
			Cause:   err,
			SudoErr: false,
		}
	}

	if info.Mode()&0111 == 0 {
		return "", &ElevationError{
			Op:      "validate_executable",
			Reason:  "executable is not executable: " + exePath,
			Cause:   nil,
			SudoErr: false,
		}
	}

	// Resolve symlinks
	resolvedPath, err := p.pm.EvalSymlinks(exePath)
	if err != nil {
		return "", &ElevationError{
			Op:      "validate_executable",
			Reason:  "failed to resolve executable symlinks",
			Cause:   err,
			SudoErr: false,
		}
	}

	return resolvedPath, nil
}

// dropPrivileges drops privileges back to the original user after performing elevated operations.
// It only attempts to drop privileges if the process is currently running as root and was elevated via sudo.
// The method follows the security practice of dropping group privileges before user privileges.
//
//nolint:funlen
func (p *PrivilegeManager) dropPrivileges() error {
	// Only drop privileges if we are currently running as root and were elevated via sudo
	if !p.isRoot() || !isElevated(p.pm) {
		return nil // Not elevated, nothing to drop
	}

	// Get the original user's UID and GID from SUDO_UID and SUDO_GID
	sudoUID := p.pm.Getenv("SUDO_UID")
	sudoGID := p.pm.Getenv("SUDO_GID")

	logger.Debugf("Parsed SUDO_UID='%s', SUDO_GID='%s'", sudoUID, sudoGID)

	if sudoUID == "" || sudoGID == "" {
		return &PrivilegeDropError{
			TargetUID: 0, // We don't know the target
			Reason:    "SUDO_UID or SUDO_GID environment variables not set",
			Cause:     nil,
		}
	}

	// Parse the UIDs/GIDs
	originalUID, err := strconv.Atoi(sudoUID)
	if err != nil {
		return &PrivilegeDropError{
			TargetUID: 0,
			Reason:    "invalid SUDO_UID value",
			Cause:     err,
		}
	}

	originalGID, err := strconv.Atoi(sudoGID)
	if err != nil {
		return &PrivilegeDropError{
			TargetUID: originalUID,
			Reason:    "invalid SUDO_GID value",
			Cause:     err,
		}
	}

	logger.Debugf("Parsed UID=%d, GID=%d", originalUID, originalGID)

	logger.Debugf("Calling LogPrivilegeDrop with success=true, uid=%d, message='dropping privileges to original user'", //nolint:lll
		originalUID)
	// Log the privilege drop attempt
	p.logger.LogPrivilegeDrop(true, originalUID, "dropping privileges to original user")

	logger.Debugf("Calling Setgid with gid=%d", originalGID)
	// Drop group privileges first (recommended practice)
	err = p.pm.Setgid(originalGID)
	if err != nil {
		dropErr := &PrivilegeDropError{
			TargetUID: originalUID,
			Reason:    "failed to set GID",
			Cause:     err,
		}
		p.logger.LogPrivilegeDrop(false, originalUID, "failed to drop group privileges")

		return dropErr
	}

	logger.Debugf("Calling Setuid with uid=%d", originalUID)
	// Then drop user privileges
	err = p.pm.Setuid(originalUID)
	if err != nil {
		dropErr := &PrivilegeDropError{
			TargetUID: originalUID,
			Reason:    "failed to set UID",
			Cause:     err,
		}
		p.logger.LogPrivilegeDrop(false, originalUID, "failed to drop user privileges")

		return dropErr
	}

	logger.Debug("Privilege drop completed successfully")

	// Log successful privilege drop
	p.logger.LogPrivilegeChange("drop", 0, originalUID, "dropped privileges to original user")

	return nil
}

// sanitizeArgs sanitizes command-line arguments to prevent privilege escalation attacks.
// It removes potentially dangerous arguments and validates argument format.
func sanitizeArgs(args []string) ([]string, error) {
	if len(args) == 0 {
		return args, nil
	}

	sanitized := make([]string, 0, len(args))

	for _, arg := range args {
		err := validateArg(arg)
		if err != nil {
			return nil, err
		}

		sanitized = append(sanitized, arg)
	}

	return sanitized, nil
}

// validateArg validates a single argument for security issues.
//
//nolint:cyclop // Complex validation logic requires multiple checks for security
func validateArg(arg string) error {
	// Skip empty arguments
	if strings.TrimSpace(arg) == "" {
		return nil
	}

	// Reject arguments that could be used for privilege escalation
	if strings.Contains(arg, ";") || strings.Contains(arg, "|") ||
		strings.Contains(arg, "&") || strings.Contains(arg, "`") ||
		strings.Contains(arg, "$(") || strings.Contains(arg, "${") {
		return fmt.Errorf("%w: %s", ErrDangerousCharacters, arg)
	}

	// Reject arguments that are dangerous sudo options
	if strings.HasPrefix(arg, "-") && len(arg) == 2 {
		switch arg[1] {
		case 'e', 'c', 'k', 'u', 'g', 'r', 'p', 'i', 't', 's', 'E', 'C', 'K', 'U', 'G', 'R', 'P', 'I', 'T':
			return fmt.Errorf("%w: %s", ErrDangerousSudoOption, arg)
		}
	}

	// Basic validation: ensure argument contains only printable characters
	for _, r := range arg {
		if !unicode.IsPrint(r) && !unicode.IsSpace(r) {
			return fmt.Errorf("%w: %s", ErrNonPrintableCharacters, arg)
		}
	}

	return nil
}
