// Package install provides functionality to install Go from downloaded archives.
// It handles archive extraction, installation verification, and user interaction.
package install

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nicholas-fedor/goUpdater/internal/archive"
	"github.com/nicholas-fedor/goUpdater/internal/download"
	"github.com/nicholas-fedor/goUpdater/internal/logger"
	"github.com/nicholas-fedor/goUpdater/internal/privileges"
	"github.com/nicholas-fedor/goUpdater/internal/verify"
	"github.com/nicholas-fedor/goUpdater/internal/version"
)

const directoryPermissions = 0755 // Default directory permissions for installation

// Install installs Go to the specified directory, either from the latest version or from a provided archive.
// It handles privilege elevation, existing installation checks, and all output messaging.
// The installDir should typically be "/usr/local/go". If archivePath is empty, the latest version is installed.
func Install(installDir, archivePath string) error {
	logger.Debugf("Starting InstallGo: installDir=%s, archivePath=%s", installDir, archivePath)

	// Check if Go is already installed
	installedVersion, err := verify.GetInstalledVersion(installDir)
	if err == nil {
		HandleExistingInstallation(installDir, installedVersion)

		return nil
	}

	if archivePath == "" {
		// Install latest version
		err = privileges.ElevateAndExecute(func() error { return Latest(installDir) })

		return fmt.Errorf("failed to install latest Go: %w", err)
	}
	// Install from archive
	err = privileges.ElevateAndExecute(func() error { return GoWithVerification(archivePath, installDir) })

	return fmt.Errorf("failed to install Go from archive: %w", err)
}

// Go extracts the Go archive to the specified installation directory.
// The installDir should typically be "/usr/local/go".
func Go(archivePath, installDir string) error {
	logger.Debugf("Starting Go installation: archive=%s, installDir=%s",
		archivePath, installDir)

	err := archive.Validate(archivePath)
	if err != nil {
		return fmt.Errorf("failed to validate archive: %w", err)
	}

	err = prepareInstallDir(installDir)
	if err != nil {
		return err
	}

	logger.Debugf("Extracting archive to: %s", filepath.Dir(installDir))

	err = archive.Extract(archivePath, filepath.Dir(installDir))
	if err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	logger.Debug("Go installation completed successfully")

	return nil
}

// GoWithVerification extracts the Go archive to the specified installation directory,
// and verifies the installation afterwards.
// The installDir should typically be "/usr/local/go".
func GoWithVerification(archivePath, installDir string) error {
	logger.Debugf("Starting Go installation with verification: archive=%s, installDir=%s",
		archivePath, installDir)

	err := Go(archivePath, installDir)
	if err != nil {
		return err
	}

	expectedVersion := archive.ExtractVersion(archivePath)
	logger.Debugf("Expected version from archive: %s", expectedVersion)

	err = verify.Installation(installDir, expectedVersion)
	if err != nil {
		return fmt.Errorf("installation verification failed: %w", err)
	}

	logger.Infof("Go successfully installed to %s", installDir)

	return nil
}

// Latest downloads the latest Go version and installs it to the specified directory.
// The installDir should typically be "/usr/local/go".
func Latest(installDir string) error {
	logger.Debugf("Starting latest Go installation: installDir=%s", installDir)

	tempDir, err := os.MkdirTemp("", "goUpdater-install-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	defer func() { _ = os.RemoveAll(tempDir) }()

	archivePath, _, err := download.GetLatest(tempDir)
	if err != nil {
		return fmt.Errorf("failed to download Go: %w", err)
	}

	logger.Debugf("Downloaded archive: %s", archivePath)

	err = GoWithVerification(archivePath, installDir)
	if err != nil {
		return err
	}

	return nil
}

// HandleExistingInstallation handles the case when Go is already installed.
// It compares versions and prompts for update if necessary.
func HandleExistingInstallation(_ string, installedVersion string) {
	// Get latest version info
	latestVersionInfo, err := download.GetLatestVersionInfo()
	if err != nil {
		logger.Errorf("Error fetching latest Go version: %v", err)
		os.Exit(1)
	}

	latestVersion := strings.TrimPrefix(latestVersionInfo.Version, "go")

	// Compare versions
	if version.Compare(installedVersion, latestVersion) >= 0 {
		logger.Infof("Go (%s) is already installed.", strings.TrimPrefix(installedVersion, "go"))

		return
	}

	// Go is installed but not latest, prompt user for confirmation
	logger.Infof("Go (%s) is already installed. Would you like to update to (%s)? (Y/n): ",
		strings.TrimPrefix(installedVersion, "go"), latestVersion)

	reader := bufio.NewReader(os.Stdin)

	response, err := reader.ReadString('\n')
	if err != nil {
		logger.Errorf("Error reading input: %v", err)
		os.Exit(1)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response == "n" || response == "no" {
		logger.Info("Update cancelled.")

		return
	}

	// Proceed with update (default is yes)
	// Note: Update functionality moved to cmd layer to avoid import cycle
	logger.Info("Update functionality should be handled by cmd layer")
}

// prepareInstallDir prepares the installation directory.
func prepareInstallDir(installDir string) error {
	logger.Debugf("Preparing installation directory: %s", installDir)

	err := os.MkdirAll(filepath.Dir(installDir), directoryPermissions)
	if err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	logger.Debug("Installation directory prepared")

	return nil
}
