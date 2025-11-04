// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package verify provides functionality to verify Go installations.
// It checks Go binary presence and version correctness.
package verify

import (
	"fmt"

	"github.com/nicholas-fedor/goUpdater/internal/exec"
	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
)

var (
	// defaultVerifier is a shared verifier instance to eliminate duplication
	// in public functions. It uses OS filesystem and command executor.
	defaultVerifier = NewVerifier( //nolint:gochecknoglobals // required to eliminate duplication in public functions
		&filesystem.OSFileSystem{},
		&exec.OSCommandExecutor{},
	)
)

// Installation checks if Go is properly installed and matches the expected version.
// It verifies that the go binary exists and that 'go version' returns the expected version.
func Installation(installDir, expectedVersion string) error {
	return defaultVerifier.Installation(installDir, expectedVersion)
}

// GetInstalledVersion returns the version of the currently installed Go.
// It runs 'go version' and extracts the version string without logging.
func GetInstalledVersion(installDir string) (string, error) {
	return defaultVerifier.GetInstalledVersionCore(installDir)
}

// GetVerificationInfo returns comprehensive verification information.
// It checks the Go installation and returns structured data about the verification results.
// This function is used by commands that need to display detailed verification details.
func GetVerificationInfo(installDir string) (VerificationInfo, error) {
	return defaultVerifier.GetVerificationInfo(installDir)
}

// GetInstalledVersionWithLogging returns the version of the currently installed Go.
// It runs 'go version', extracts the version string, and logs the result.
func GetInstalledVersionWithLogging(installDir string) (string, error) {
	return defaultVerifier.GetInstalledVersionWithLogging(installDir)
}

// Verify performs the complete Go verification workflow.
// It retrieves verification information for the specified install directory,
// displays the results, and returns any errors encountered.
func Verify(installDir string) error {
	v := NewVerifier(&filesystem.OSFileSystem{}, &exec.OSCommandExecutor{})

	return v.Verify(installDir)
}

// RunVerify executes the verify command logic.
// It performs the complete Go verification workflow for the specified install directory.
func RunVerify(verifyDir string) error {
	err := Verify(verifyDir)
	if err != nil {
		return fmt.Errorf("verification failed for directory %s: %w", verifyDir, err)
	}

	return nil
}
