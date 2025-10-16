// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package update provides orchestration logic for updating Go installations.
// It coordinates downloading, installing, and verifying Go updates.
package update

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/nicholas-fedor/goUpdater/internal/download"
	"github.com/nicholas-fedor/goUpdater/internal/install"
	"github.com/nicholas-fedor/goUpdater/internal/logger"
	"github.com/nicholas-fedor/goUpdater/internal/privileges"
	"github.com/nicholas-fedor/goUpdater/internal/uninstall"
	"github.com/nicholas-fedor/goUpdater/internal/verify"
	"github.com/nicholas-fedor/goUpdater/internal/version"
)

var (
	// ErrGoNotInstalled indicates that Go is not installed in the specified directory.
	ErrGoNotInstalled = errors.New("Go is not installed")
)

// Go performs a complete Go update: checks if Go is installed, compares versions,
// downloads the latest version if needed, removes the existing installation,
// installs the new version, verifies it, and logs success message.
// installDir is the directory where Go should be installed (e.g., "/usr/local/go").
// autoInstall enables automatic installation if Go is not present.
func Go(installDir string, autoInstall bool) error {
	logger.Debugf("Starting Go update process: installDir=%s, autoInstall=%t", installDir, autoInstall)

	installedVersion, latestVersionStr, err := checkAndPrepare(installDir, autoInstall)
	if err != nil {
		logger.Debugf("checkAndPrepare failed: %v", err)

		return err
	}

	logger.Debugf("checkAndPrepare succeeded: installedVersion=%s, latestVersionStr=%s",
		installedVersion, latestVersionStr)

	needsUpdateResult := needsUpdate(installedVersion, latestVersionStr)
	logger.Debugf("needsUpdate result: %t", needsUpdateResult)

	if !needsUpdateResult {
		logger.Debug("No update needed, returning nil")

		return nil
	}

	logger.Debug("Update needed, proceeding to download")

	archivePath, tempDir, err := downloadLatest()
	if err != nil {
		logger.Debugf("downloadLatest failed: %v", err)

		return err
	}

	logger.Debugf("downloadLatest succeeded: archivePath=%s, tempDir=%s", archivePath, tempDir)

	defer func() { _ = os.RemoveAll(tempDir) }()

	err = performUpdate(archivePath, installDir, installedVersion)
	if err != nil {
		logger.Debugf("performUpdate failed: %v", err)

		return err
	}

	logger.Debug("performUpdate succeeded")

	err = verify.Installation(installDir, latestVersionStr)
	if err != nil {
		logger.Debugf("verify.Installation failed: %v", err)

		return fmt.Errorf("failed to verify installation: %w", err)
	}

	logger.Debug("verify.Installation succeeded")

	return nil
}

// GoWithPrivileges performs a complete Go update workflow including privilege checking,
// version comparison, user prompts, and success/error messaging.
// It wraps the existing update logic and handles all display/output logic.
// installDir is the directory where Go should be installed (e.g., "/usr/local/go").
// autoInstall enables automatic installation if Go is not present.
func GoWithPrivileges(installDir string, autoInstall bool) error {
	logger.Debugf("Starting update operation: installDir=%s, autoInstall=%t", installDir, autoInstall)

	err := privileges.ElevateAndExecute(func() error { return Go(installDir, autoInstall) })
	if err != nil {
		logger.Debugf("privileges.ElevateAndExecute failed: %v", err)

		return fmt.Errorf("failed to update Go: %w", err)
	}

	logger.Debug("privileges.ElevateAndExecute succeeded")

	return nil
}

// checkAndPrepare checks if Go is installed, fetches the latest version, and determines if an update is needed.
// It returns the installed version, latest version string, and any error encountered.
func checkAndPrepare(installDir string, autoInstall bool) (string, string, error) {
	installedVersion, err := checkInstallation(installDir, autoInstall)
	if err != nil {
		return "", "", err
	}

	latestVersion, err := download.GetLatestVersionInfo()
	if err != nil {
		return "", "", fmt.Errorf("failed to get latest version info: %w", err)
	}

	latestVersionStr := strings.TrimPrefix(latestVersion.Version, "go")
	logger.Debugf("Latest available version: %s", latestVersionStr)

	return installedVersion, latestVersionStr, nil
}

// checkInstallation checks if Go is installed and handles auto-install logic.
func checkInstallation(installDir string, autoInstall bool) (string, error) {
	installedVersion, err := verify.GetInstalledVersion(installDir)
	if err != nil {
		logger.Debugf("Go not found in %s: %v", installDir, err)

		if !autoInstall {
			return "", fmt.Errorf("%w in %s. Use --auto-install flag to install it automatically", ErrGoNotInstalled, installDir)
		}

		logger.Info("Go is not installed. Proceeding with installation.")
	} else {
		logger.Debugf("Found installed Go version: %s", installedVersion)
	}

	return installedVersion, nil
}

// downloadLatest downloads the latest Go archive to a temporary directory.
// It returns the archive path, temp directory path, and any error encountered.
func downloadLatest() (string, string, error) {
	tempDir, err := os.MkdirTemp("", "goUpdater-*")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	archivePath, _, err := download.GetLatest(tempDir)
	if err != nil {
		_ = os.RemoveAll(tempDir)

		return "", "", fmt.Errorf("failed to download Go: %w", err)
	}

	return archivePath, tempDir, nil
}

// performUpdate handles the uninstallation of the existing Go installation and installation of the new version.
// It takes the archive path, install directory, and installed version as parameters.
func performUpdate(archivePath, installDir, installedVersion string) error {
	logger.Debugf("Performing update: archive=%s, installDir=%s, installedVersion=%s",
		archivePath, installDir, installedVersion)

	if installedVersion != "" {
		logger.Debug("Uninstalling existing Go installation")

		err := privileges.ElevateAndExecute(func() error { return uninstall.Remove(installDir) })
		if err != nil {
			return fmt.Errorf("failed to uninstall existing Go: %w", err)
		}
	}

	logger.Debug("Installing new Go version")

	err := install.Go(archivePath, installDir)
	if err != nil {
		return fmt.Errorf("failed to install Go: %w", err)
	}

	logger.Debug("Go installation completed successfully")

	return nil
}

// needsUpdate determines if an update is required based on version comparison.
func needsUpdate(installedVersion, latestVersionStr string) bool {
	if installedVersion == "" {
		return true
	}

	if version.Compare(installedVersion, latestVersionStr) >= 0 {
		logger.Infof("Latest Go version (%s) already installed.", latestVersionStr)

		return false
	}

	logger.Infof("Updating Go from %s to %s", installedVersion, latestVersionStr)

	return true
}
