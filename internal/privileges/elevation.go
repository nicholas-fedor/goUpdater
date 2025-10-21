// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package privileges

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

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

	sanitizedArgs, err := sanitizeArgs(os.Args[1:])
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
