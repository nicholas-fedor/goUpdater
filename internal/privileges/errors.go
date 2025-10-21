// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package privileges

import (
	"errors"
	"fmt"
)

// ErrElevationRequired indicates that the operation requires elevated privileges.
var ErrElevationRequired = errors.New("operation requires elevated privileges")

// ErrElevationFailed indicates that privilege elevation failed.
var ErrElevationFailed = errors.New("privilege elevation failed")

// ErrSudoNotAvailable indicates that sudo is not available on the system.
var ErrSudoNotAvailable = errors.New("sudo is not available on this system")

// ErrExecutableNotFound indicates that the executable path could not be determined.
var ErrExecutableNotFound = errors.New("executable path could not be determined")

// ErrPrivilegeDropFailed indicates that dropping privileges back to the original user failed.
var ErrPrivilegeDropFailed = errors.New("failed to drop privileges to original user")

// ErrArgumentSanitizationFailed indicates that argument sanitization failed due to dangerous content.
var ErrArgumentSanitizationFailed = errors.New("argument sanitization failed due to dangerous content")

// ErrDangerousCharacters indicates that arguments contain dangerous characters.
var ErrDangerousCharacters = errors.New("arguments contain dangerous characters")

// ErrDangerousSudoOption indicates that arguments contain dangerous sudo options.
var ErrDangerousSudoOption = errors.New("arguments contain dangerous sudo options")

// ErrNonPrintableCharacters indicates that arguments contain non-printable characters.
var ErrNonPrintableCharacters = errors.New("arguments contain non-printable characters")

// ErrGoNotInstalled indicates that Go is not installed in the specified directory.
var ErrGoNotInstalled = errors.New("go is not installed in /usr/local/go. " +
	"Use --auto-install flag to install it automatically")

// ElevationError represents an error during privilege elevation with additional context.
type ElevationError struct {
	Op      string // Operation that failed
	Reason  string // Reason for failure
	Cause   error  // Underlying error
	SudoErr bool   // Whether the error is related to sudo
}

// ValidationError represents a validation error for privilege operations.
type ValidationError struct {
	Field  string // Field that failed validation
	Value  string // Value that was invalid
	Reason string // Reason for invalidity
}

// PrivilegeDropError represents an error when dropping privileges.
type PrivilegeDropError struct {
	TargetUID int    // Target UID to drop to
	Reason    string // Reason for failure
	Cause     error  // Underlying error
}

func (e *ElevationError) Error() string {
	msg := fmt.Sprintf("privilege elevation failed during %s: %s", e.Op, e.Reason)
	if e.Cause != nil {
		msg += fmt.Sprintf(" (cause: %v)", e.Cause)
	}

	if e.SudoErr {
		msg += " - ensure sudo is installed and configured correctly"
	}

	return msg
}

func (e *ElevationError) Unwrap() error {
	return e.Cause
}

func (v *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s '%s': %s", v.Field, v.Value, v.Reason)
}

func (p *PrivilegeDropError) Error() string {
	msg := fmt.Sprintf("failed to drop privileges to UID %d: %s", p.TargetUID, p.Reason)
	if p.Cause != nil {
		msg += fmt.Sprintf(" (cause: %v)", p.Cause)
	}

	return msg
}

func (p *PrivilegeDropError) Unwrap() error {
	return p.Cause
}
