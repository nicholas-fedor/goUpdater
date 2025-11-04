// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package update

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/mod/semver"

	"github.com/nicholas-fedor/goUpdater/internal/exec"
	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	"github.com/nicholas-fedor/goUpdater/internal/logger"
	"github.com/nicholas-fedor/goUpdater/internal/uninstall"
)

// NewUpdater creates a new Updater instance with the provided dependencies.
func NewUpdater(
	fileSystem filesystem.FileSystem,
	commandExecutor exec.CommandExecutor,
	archiveDownloader ArchiveDownloader,
	installer Installer,
	uninstaller Uninstaller,
	verifier Verifier,
	privilegeManager PrivilegeManager,
	versionFetcher VersionFetcher,
) *Updater {
	return &Updater{
		fileSystem:        fileSystem,
		commandExecutor:   commandExecutor,
		versionFetcher:    versionFetcher,
		archiveDownloader: archiveDownloader,
		installer:         installer,
		uninstaller:       uninstaller,
		verifier:          verifier,
		privilegeManager:  privilegeManager,
	}
}

// Update performs a complete Go update: checks if Go is installed, compares versions,
// downloads the latest version if needed, removes the existing installation,
// installs the new version, verifies it, and logs success message.
// installDir is the directory where Go should be installed (e.g., "/usr/local/go").
// autoInstall enables automatic installation if Go is not present.
//
//nolint:funlen
func (u *Updater) Update(installDir string, autoInstall bool) error {
	logger.Debugf("Starting Go update process: installDir=%s, autoInstall=%t", installDir, autoInstall)

	installedVersion, latestVersionStr, err := u.checkAndPrepare(installDir, autoInstall)
	if err != nil {
		logger.Debugf("checkAndPrepare failed: %v", err)

		if errors.Is(err, ErrGoNotInstalled) {
			return ErrGoNotInstalled
		}

		return &Error{
			OperationPhase: "check",
			CurrentStep:    "prepare",
			Progress:       "checking installation and fetching latest version",
			Err:            err,
		}
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

	archivePath, tempDir, err := u.downloadLatest()
	if err != nil {
		logger.Debugf("downloadLatest failed: %v", err)

		return err
	}

	logger.Debugf("downloadLatest succeeded: archivePath=%s, tempDir=%s", archivePath, tempDir)

	// Clean up temporary directory after function completes
	defer func() { _ = u.fileSystem.RemoveAll(tempDir) }()

	err = u.performUpdate(archivePath, installDir, installedVersion)
	if err != nil {
		logger.Debugf("performUpdate failed: %v", err)

		return err
	}

	logger.Debug("performUpdate succeeded")

	err = u.verifier.Installation(installDir, latestVersionStr)
	if err != nil {
		logger.Debugf("u.verifier.Installation failed: %v", err)

		return &Error{
			OperationPhase: "verify",
			CurrentStep:    "check_installation",
			Progress:       "verifying Go installation",
			Err:            err,
		}
	}

	logger.Debug("verify.Installation succeeded")

	return nil
}

// UpdateWithPrivileges performs a complete Go update workflow using the injected dependencies.
// It first checks if Go is installed without requiring privileges, and only elevates if needed.
func (u *Updater) UpdateWithPrivileges(installDir string, autoInstall bool) error {
	logger.Debugf("Starting update operation: installDir=%s, autoInstall=%t", installDir, autoInstall)

	// Check installation status first without requiring elevation
	installedVersion, err := u.checkInstallation(installDir, autoInstall)
	if err != nil {
		logger.Debugf("checkInstallation failed: %v", err)

		if errors.Is(err, ErrGoNotInstalled) {
			return ErrGoNotInstalled
		}

		return fmt.Errorf("failed to check installation: %w", err)
	}

	logger.Debugf("checkInstallation succeeded: installedVersion=%s", installedVersion)

	// Proceed with elevation for the update operation
	err = u.privilegeManager.ElevateAndExecute(func() error { return u.Update(installDir, autoInstall) })
	if err != nil {
		logger.Debugf("privileges.ElevateAndExecute failed: %v", err)

		return fmt.Errorf("failed to update Go: %w", err)
	}

	logger.Debug("privileges.ElevateAndExecute succeeded")

	return nil
}

// checkAndPrepare checks if Go is installed, fetches the latest version, and determines if an update is needed.
// It returns the installed version, latest version string, and any error encountered.
func (u *Updater) checkAndPrepare(installDir string, autoInstall bool) (string, string, error) {
	installedVersion, err := u.checkInstallation(installDir, autoInstall)
	if err != nil {
		return "", "", err
	}

	if u.versionFetcher == nil {
		return "", "", &Error{
			OperationPhase: "check",
			CurrentStep:    "fetch_version",
			Progress:       "version fetcher not initialized",
			Err:            ErrVersionFetcherNil,
		}
	}

	latestVersion, err := u.versionFetcher.GetLatestVersionInfo()
	if err != nil {
		return "", "", &Error{
			OperationPhase: "check",
			CurrentStep:    "fetch_version",
			Progress:       "fetching latest Go version information",
			Err:            err,
		}
	}

	logger.Debugf("latestVersion.Version starts with 'go': %t",
		strings.HasPrefix(latestVersion.Version, "go"))

	var latestVersionStr string
	if !strings.HasPrefix(latestVersion.Version, "go") {
		latestVersionStr = "go" + latestVersion.Version
	} else {
		latestVersionStr = latestVersion.Version
	}

	logger.Debugf("Latest available version: %s", latestVersionStr)

	return installedVersion, latestVersionStr, nil
}

// checkInstallation checks if Go is installed and handles auto-install logic.
// It returns the installed version if Go is present, or an empty string if not installed.
// If autoInstall is false and Go is not installed, it returns an error.
// It uses absolute paths to the go binary to avoid PATH-based resolution issues.
func (u *Updater) checkInstallation(installDir string, autoInstall bool) (string, error) {
	goBinary := filepath.Join(installDir, "bin", "go")
	logger.Debugf("Checking for Go binary at: %s", goBinary)

	cmd := u.commandExecutor.CommandContext(context.Background(), goBinary, "version")

	output, err := cmd.Output()
	if err != nil {
		logger.Debugf("Go not found in %s: %v", installDir, err)

		if !autoInstall {
			return "", fmt.Errorf("%w", ErrGoNotInstalled)
		}

		logger.Info("Go is not installed. Proceeding with installation.")

		return "", nil
	}

	versionOutput := strings.TrimSpace(string(output))

	// Parse version from output like "go version go1.21.0 linux/amd64"
	// Expected format: ["go", "version", "go1.21.0", "linux/amd64"]
	parts := strings.Fields(versionOutput)
	if len(parts) >= 3 && parts[0] == "go" && parts[1] == "version" {
		version := parts[2]
		logger.Debugf("Found installed Go version: %s", version)

		// Validate the parsed version using semver
		if !semver.IsValid("v" + strings.TrimPrefix(version, "go")) {
			return "", fmt.Errorf("%w: invalid version format: %s", ErrUnableToParseVersion, version)
		}

		return version, nil
	}

	return "", fmt.Errorf("%w: %s", ErrUnableToParseVersion, versionOutput)
}

// downloadLatest downloads the latest Go archive to a temporary directory.
// It returns the archive path, temp directory path, and any error encountered.
// The caller is responsible for cleaning up the temporary directory.
func (u *Updater) downloadLatest() (string, string, error) {
	tempDir, err := u.fileSystem.MkdirTemp("", "goUpdater-*")
	if err != nil {
		return "", "", &Error{
			OperationPhase: "download",
			CurrentStep:    "create_temp_dir",
			Progress:       "creating temporary directory",
			Err:            err,
		}
	}

	archivePath, _, err := u.archiveDownloader.GetLatest(tempDir)
	if err != nil {
		_ = u.fileSystem.RemoveAll(tempDir)

		return "", "", &Error{
			OperationPhase: "download",
			CurrentStep:    "download_archive",
			Progress:       "downloading latest Go archive",
			Err:            err,
		}
	}

	return archivePath, tempDir, nil
}

// performUpdate handles the uninstallation of the existing Go installation and installation of the new version.
// It takes the archive path, install directory, and installed version as parameters.
// If installedVersion is empty, it skips the uninstallation step.
func (u *Updater) performUpdate(archivePath, installDir, installedVersion string) error {
	logger.Debugf("Performing update: archive=%s, installDir=%s, installedVersion=%s",
		archivePath, installDir, installedVersion)

	if installedVersion != "" {
		logger.Debug("Uninstalling existing Go installation")

		err := u.privilegeManager.ElevateAndExecute(func() error { return u.uninstaller.Remove(installDir) })
		if err != nil {
			if errors.Is(err, uninstall.ErrInstallDirEmpty) {
				return &Error{
					OperationPhase: "uninstall",
					CurrentStep:    "validate_install_dir",
					Progress:       "validating installation directory",
					Err:            err,
				}
			}

			return &Error{
				OperationPhase: "uninstall",
				CurrentStep:    "remove_existing",
				Progress:       "removing existing Go installation",
				Err:            err,
			}
		}
	}

	logger.Debug("Installing new Go version")

	err := u.installer.Extract(archivePath, installDir, installedVersion)
	if err != nil {
		return &Error{
			OperationPhase: "install",
			CurrentStep:    "extract_archive",
			Progress:       "extracting Go archive to installation directory",
			Err:            err,
		}
	}

	logger.Debug("Go installation completed successfully")

	return nil
}
