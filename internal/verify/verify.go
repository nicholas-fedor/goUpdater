// Package verify provides functionality to verify Go installations.
// It checks Go binary presence and version correctness.
package verify

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nicholas-fedor/goUpdater/internal/cli"
	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

// errVersionMismatch indicates a version mismatch.
var errVersionMismatch = errors.New("version mismatch")

// errVersionParseError indicates an error parsing the version.
var errVersionParseError = errors.New("version parse error")

// VerificationInfo holds detailed verification information for the Go installation.
// It includes the installation directory, version, and verification status.
//
//revive:disable:exported
type VerificationInfo struct {
	InstallDir string `json:"installDir"`
	Version    string `json:"version"`
	Status     string `json:"status"`
}

// Installation checks if Go is properly installed and matches the expected version.
// It verifies that the go binary exists and that 'go version' returns the expected version.
func Installation(installDir, expectedVersion string) error {
	logger.Debugf("Verifying Go installation: installDir=%s, expectedVersion=%s", installDir, expectedVersion)
	goBinary := filepath.Join(installDir, "bin", "go")

	// Check if the go binary exists
	logger.Debugf("Checking for go binary at: %s", goBinary)

	_, err := exec.LookPath(goBinary)
	if err != nil {
		return fmt.Errorf("go binary not found at %s: %w", goBinary, err)
	}

	// Run 'go version' and check the output
	// gosec: G204 - Subprocess launched with variable is acceptable here as we control the goBinary path
	// gosec: G204 - Subprocess launched with variable is acceptable here as we control the goBinary path
	logger.Debug("Running 'go version' command")

	cmd := exec.CommandContext(context.Background(), goBinary, "version") //nolint:gosec

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to run 'go version': %w", err)
	}

	versionOutput := strings.TrimSpace(string(output))
	logger.Debugf("Go version output: %s", versionOutput)

	// Check if the output contains the expected version
	if !strings.Contains(versionOutput, expectedVersion) {
		return fmt.Errorf("version mismatch: expected %s, got %s: %w", expectedVersion, versionOutput, errVersionMismatch)
	}

	logger.Debug("Installation verification successful")

	return nil
}

// GetInstalledVersion returns the version of the currently installed Go.
// It runs 'go version' and extracts the version string without logging.
func GetInstalledVersion(installDir string) (string, error) {
	return getInstalledVersionCore(installDir)
}

// GetVerificationInfo returns comprehensive verification information.
// It checks the Go installation and returns structured data about the verification results.
// This function is used by commands that need to display detailed verification details.
func GetVerificationInfo(installDir string) (VerificationInfo, error) {
	logger.Debugf("Getting verification info from: %s", installDir)

	version, err := getInstalledVersionCore(installDir)
	if err != nil {
		return VerificationInfo{
			InstallDir: installDir,
			Version:    "",
			Status:     "failed",
		}, err
	}

	logger.Debugf("Verification successful: version=%s", version)

	return VerificationInfo{
		InstallDir: installDir,
		Version:    version,
		Status:     "verified",
	}, nil
}

// GetInstalledVersionWithLogging returns the version of the currently installed Go.
// It runs 'go version', extracts the version string, and logs the result.
func GetInstalledVersionWithLogging(installDir string) (string, error) {
	logger.Debugf("Getting installed Go version from: %s", installDir)
	goBinary := filepath.Join(installDir, "bin", "go")

	logger.Debugf("Running 'go version' for binary: %s", goBinary)

	version, err := getInstalledVersionCore(installDir)
	if err != nil {
		return "", err
	}

	logger.Debugf("Parsed version: %s", version)
	logger.Info("Go installation verified. Version: " + version)

	return version, nil
}

// Verify performs the complete Go verification workflow.
// It retrieves verification information for the specified install directory,
// displays the results, and handles any errors by logging and exiting.
func Verify(installDir string) {
	logger.Debugf("Starting verification: installDir=%s", installDir)

	info, err := GetVerificationInfo(installDir)
	if err != nil {
		logger.Errorf("Error verifying Go installation: %v", err)
		os.Exit(1)
	}

	logger.Debugf("Verification completed: version=%s, status=%s", info.Version, info.Status)

	var items []string

	if info.InstallDir != "" {
		items = append(items, "Directory: "+info.InstallDir)
	}

	if info.Version != "" {
		items = append(items, "Version: "+info.Version)
	}

	if info.Status != "" {
		items = append(items, "Status: "+info.Status)
	}

	_, _ = fmt.Fprint(os.Stdout, cli.TreeFormat("Go Installation Verification", items))
}

// getInstalledVersionCore returns the version of the currently installed Go without logging.
// It runs 'go version' and extracts the version string.
func getInstalledVersionCore(installDir string) (string, error) {
	goBinary := filepath.Join(installDir, "bin", "go")

	cmd := exec.CommandContext(context.Background(), goBinary, "version") //nolint:gosec

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run 'go version': %w", err)
	}

	versionOutput := strings.TrimSpace(string(output))

	// Parse version from output like "go version go1.21.0 linux/amd64"
	parts := strings.Fields(versionOutput)
	if len(parts) >= 3 && parts[0] == "go" && parts[1] == "version" {
		version := parts[2]

		return version, nil
	}

	return "", fmt.Errorf("unable to parse version from output: %s: %w", versionOutput, errVersionParseError)
}
