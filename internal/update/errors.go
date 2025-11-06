// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package update provides error types and utilities for update operations.
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

	// ErrGetLatestVersionFuncNotSet indicates that the getLatestVersionFunc is not set.
	ErrGetLatestVersionFuncNotSet = errors.New("getLatestVersionFunc is not set")

	// ErrGoNotFound indicates that Go was not found.
	ErrGoNotFound = errors.New("go not found")
	// ErrExecutableFileNotFound indicates that the executable file was not found.
	ErrExecutableFileNotFound = errors.New("executable file not found")
	// ErrNetworkError indicates that a network error occurred.
	ErrNetworkError = errors.New("network error")
	// ErrDownloadFailed indicates that the download failed.
	ErrDownloadFailed = errors.New("download failed")
	// ErrUninstallFailed indicates that the uninstall failed.
	ErrUninstallFailed = errors.New("uninstall failed")
	// ErrInstallFailed indicates that the install failed.
	ErrInstallFailed = errors.New("install failed")
	// ErrVerificationFailed indicates that the verification failed.
	ErrVerificationFailed = errors.New("verification failed")
	// ErrElevationFailed indicates that privilege elevation failed.
	ErrElevationFailed = errors.New("elevation failed")
	// ErrMkdirTempFailed indicates that creating a temporary directory failed.
	ErrMkdirTempFailed = errors.New("mkdir temp failed")
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
