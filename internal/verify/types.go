// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package verify

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nicholas-fedor/goUpdater/internal/exec"
	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

// Static test errors to satisfy err113 linter rule.
var (
	errUnexpectedGoVersionOutputFormat = errors.New("unexpected go version output format")
	errPermissionDenied                = errors.New("permission denied checking Go binary")
)

// VerificationInfo holds detailed verification information for the Go installation.
type VerificationInfo struct {
	InstallDir string `json:"installDir"`
	Version    string `json:"version"`
	Status     string `json:"status"`
}

// Verifier handles Go installation verification operations.
type Verifier struct {
	fs       filesystem.FileSystem
	executor exec.CommandExecutor
}

// NewVerifier creates a new Verifier with the provided dependencies.
func NewVerifier(fs filesystem.FileSystem, executor exec.CommandExecutor) *Verifier {
	return &Verifier{
		fs:       fs,
		executor: executor,
	}
}

// Verify performs the complete Go verification workflow.
// It retrieves verification information for the specified install directory,
// displays the results, and returns any errors encountered.
func (v *Verifier) Verify(installDir string) error {
	info, err := v.GetVerificationInfo(installDir)
	if err != nil {
		return err
	}

	v.displayVerificationInfo(info)

	return nil
}

// Installation checks if Go is properly installed and matches the expected version.
// It verifies that the go binary exists and that 'go version' returns the expected version.
func (v *Verifier) Installation(installDir, expectedVersion string) error {
	version, err := v.GetInstalledVersionCore(installDir)
	if err != nil {
		return fmt.Errorf("failed to get installed version: %w", err)
	}

	if version == "" {
		return &VerificationError{
			ExpectedVersion: expectedVersion,
			ActualVersion:   "",
			BinaryPath:      filepath.Join(installDir, "bin", "go"),
			Err:             ErrGoNotInstalled,
		}
	}

	if version != expectedVersion {
		return &VerificationError{
			ExpectedVersion: expectedVersion,
			ActualVersion:   version,
			BinaryPath:      filepath.Join(installDir, "bin", "go"),
			Err:             ErrVersionMismatch,
		}
	}

	return nil
}

// GetInstalledVersionCore returns the version of the currently installed Go.
// It runs 'go version' and extracts the version string without logging.
func (v *Verifier) GetInstalledVersionCore(installDir string) (string, error) {
	goBinPath := filepath.Join(installDir, "bin", "go")
	logger.Debugf("Checking for Go binary at path: %s", goBinPath)

	// Check if go binary exists
	_, err := v.fs.Stat(goBinPath)
	if err != nil {
		if v.fs.IsNotExist(err) {
			logger.Debugf("Go binary not found at %s: file does not exist", goBinPath)
			logger.Debugf("Go is not installed in %s", installDir)

			return "", nil // Go is not installed
		}

		if os.IsPermission(err) {
			logger.Errorf("permission denied checking Go binary at %s", goBinPath)

			return "", fmt.Errorf("%w at %s", errPermissionDenied, goBinPath)
		}

		logger.Debugf("Failed to check Go binary at %s: %v", goBinPath, err)

		return "", fmt.Errorf("failed to check Go binary at %s: %w", goBinPath, err)
	}

	logger.Debugf("Go binary found at %s, executing 'go version' command", goBinPath)

	// Run 'go version' command
	ctx := context.Background()
	cmd := v.executor.CommandContext(ctx, goBinPath, "version")

	output, err := cmd.Output()
	if err != nil {
		logger.Debugf("Failed to execute 'go version' command: %v", err)

		return "", fmt.Errorf("failed to execute 'go version' at %s: %w", goBinPath, err)
	}

	logger.Debugf("Go version command output: %s", strings.TrimSpace(string(output)))

	// Parse version from output (format: "go version go1.21.0 linux/amd64")
	parts := strings.Fields(string(output))
	if len(parts) < 3 || parts[0] != "go" || parts[1] != "version" {
		logger.Debugf("Unexpected go version output format: %s", string(output))

		return "", errUnexpectedGoVersionOutputFormat
	}

	version := parts[2]
	logger.Debugf("Extracted Go version: %s", version)

	return version, nil
}

// GetVerificationInfo returns comprehensive verification information.
// It checks the Go installation and returns structured data about the verification results.
func (v *Verifier) GetVerificationInfo(installDir string) (VerificationInfo, error) {
	version, err := v.GetInstalledVersionCore(installDir)
	if err != nil {
		return VerificationInfo{
			InstallDir: installDir,
			Version:    "",
			Status:     "Failed to get version",
		}, fmt.Errorf("failed to get installed version: %w", err)
	}

	status := "Verified"
	if version == "" {
		status = "Not installed or not found"
	}

	return VerificationInfo{
		InstallDir: installDir,
		Version:    version,
		Status:     status,
	}, nil
}

// GetInstalledVersion returns the version of the currently installed Go.
// It runs 'go version' and extracts the version string without logging.
func (v *Verifier) GetInstalledVersion(installDir string) (string, error) {
	return v.GetInstalledVersionCore(installDir)
}

// GetInstalledVersionWithLogging returns the version of the currently installed Go.
// It runs 'go version', extracts the version string, logs the full command output
// along with a clear message using the verifier's logger, and returns the version
// string and any execution/parsing error. Errors include command output for debugging.
func (v *Verifier) GetInstalledVersionWithLogging(installDir string) (string, error) {
	goBinPath := filepath.Join(installDir, "bin", "go")

	// Check if go binary exists
	_, err := v.fs.Stat(goBinPath)
	if err != nil {
		if v.fs.IsNotExist(err) {
			logger.Infof("Go binary not found at %s: file does not exist", goBinPath)
			logger.Infof("Go is not installed in %s", installDir)

			return "", nil // Go is not installed
		}

		if os.IsPermission(err) {
			logger.Infof("Go binary access denied at %s: permission denied, assuming not installed", goBinPath)
			logger.Infof("Go is not installed in %s", installDir)

			return "", nil // Assume not installed due to permission error
		}

		logger.Infof("Failed to check Go binary at %s: %v", goBinPath, err)

		return "", fmt.Errorf("failed to check Go binary at %s: %w", goBinPath, err)
	}

	logger.Infof("Go binary found at %s, executing 'go version' command", goBinPath)

	// Run 'go version' command
	ctx := context.Background()
	cmd := v.executor.CommandContext(ctx, goBinPath, "version")

	output, err := cmd.Output()
	if err != nil {
		logger.Infof("Failed to execute 'go version' command: %v", err)
		logger.Infof("Command output: %s", strings.TrimSpace(string(output)))

		return "", fmt.Errorf("failed to execute 'go version' at %s: %w (output: %s)",
			goBinPath, err, strings.TrimSpace(string(output)))
	}

	logger.Infof("Go version command output: %s", strings.TrimSpace(string(output)))

	// Parse version from output (format: "go version go1.21.0 linux/amd64")
	parts := strings.Fields(string(output))
	if len(parts) < 3 || parts[0] != "go" || parts[1] != "version" {
		logger.Infof("Unexpected go version output format: %s", string(output))

		return "", fmt.Errorf("%w: %s", errUnexpectedGoVersionOutputFormat, string(output))
	}

	version := parts[2]
	logger.Infof("Extracted Go version: %s", version)

	return version, nil
}

// displayVerificationInfo displays verification information in a tree format.
// It shows the install directory, version, and status in a hierarchical structure.
func (v *Verifier) displayVerificationInfo(info VerificationInfo) {
	logger.Infof("Go Installation Verification")

	if info.InstallDir != "" {
		logger.Infof("├── Directory: %s", info.InstallDir)
	}

	if info.Version != "" {
		logger.Infof("├── Version: %s", info.Version)
	}

	if info.Status != "" {
		logger.Infof("└── Status: %s", info.Status)
	}
}
