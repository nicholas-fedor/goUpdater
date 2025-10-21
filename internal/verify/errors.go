// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package verify

import (
	"errors"
	"fmt"
)

var (
	// ErrUnexpectedGoVersionOutput indicates that the go version command output format is unexpected.
	ErrUnexpectedGoVersionOutput = errors.New("unexpected go version output format")
	// ErrGoNotInstalled indicates that Go is not installed.
	ErrGoNotInstalled = errors.New("go is not installed")
	// ErrVersionMismatch indicates that the installed Go version does not match the expected version.
	ErrVersionMismatch = errors.New("version mismatch")
)

// VerificationError represents version verification failures with expected vs actual versions and binary path.
type VerificationError struct {
	ExpectedVersion string
	ActualVersion   string
	BinaryPath      string
	Err             error
}

// Error implements the error interface for VerificationError.
func (e *VerificationError) Error() string {
	if e.ActualVersion == "" {
		return fmt.Sprintf("verification failed: expected version %s at %s, but Go is not installed",
			e.ExpectedVersion, e.BinaryPath)
	}

	return fmt.Sprintf("verification failed: expected version %s, got %s at %s",
		e.ExpectedVersion, e.ActualVersion, e.BinaryPath)
}

// Unwrap returns the underlying error for compatibility with errors.Is and errors.As.
func (e *VerificationError) Unwrap() error {
	return e.Err
}
