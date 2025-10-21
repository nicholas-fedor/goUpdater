// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package update

import (
	"errors"
	"fmt"
)

var (
	// ErrGoNotInstalled indicates that Go is not installed in the specified directory.
	ErrGoNotInstalled = errors.New("go is not installed")
	// ErrUnableToParseVersion indicates that the version output could not be parsed.
	ErrUnableToParseVersion = errors.New("unable to parse version from output")
	// ErrVersionFetcherNil indicates that the version fetcher is nil.
	ErrVersionFetcherNil = errors.New("version fetcher is nil")
)

// Error represents multi-step update failures with operation phase, current step, and progress context.
type Error struct {
	OperationPhase string
	CurrentStep    string
	Progress       string
	Err            error
}

// Error implements the error interface for Error.
func (e *Error) Error() string {
	return fmt.Sprintf("update failed: phase=%s step=%s progress=%s: %v",
		e.OperationPhase, e.CurrentStep, e.Progress, e.Err)
}

// Unwrap returns the underlying error for compatibility with errors.Is and errors.As.
func (e *Error) Unwrap() error {
	return e.Err
}
