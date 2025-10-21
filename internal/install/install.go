// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package install provides functionality to install Go from downloaded archives.
// It handles archive extraction, installation verification, and user interaction.
package install

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/semver"

	"github.com/nicholas-fedor/goUpdater/internal/archive"
	"github.com/nicholas-fedor/goUpdater/internal/download"
	"github.com/nicholas-fedor/goUpdater/internal/exec"
	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	httpclient "github.com/nicholas-fedor/goUpdater/internal/http"
	"github.com/nicholas-fedor/goUpdater/internal/logger"
	"github.com/nicholas-fedor/goUpdater/internal/privileges"
	"github.com/nicholas-fedor/goUpdater/internal/verify"
)

const directoryPermissions = 0755 // Default directory permissions for installation

var errVersionParseError = errors.New("version parse error")

// compare compares two Go version strings.
// It returns -1 if version1 < version2, 0 if version1 == version2, and 1 if version1 > version2.
// If either version is invalid, it returns an error.
// Both versions should be in the format "go1.x.x", "1.x.x", or "v1.x.x" (with or without "go" or "v" prefix).
func compare(version1, version2 string) (int, error) {
	if version1 == "" {
		return 0, fmt.Errorf("version1 cannot be empty: %w", errVersionParseError)
	}

	if version2 == "" {
		return 0, fmt.Errorf("version2 cannot be empty: %w", errVersionParseError)
	}

	// Normalize versions by removing "go" and "v" prefixes, then adding "v" prefix for semver
	version1Normalized := "v" + strings.TrimPrefix(strings.TrimPrefix(version1, "go"), "v")
	version2Normalized := "v" + strings.TrimPrefix(strings.TrimPrefix(version2, "go"), "v")

	if !semver.IsValid(version1Normalized) {
		return 0, fmt.Errorf("invalid version %s: %w", version1, errVersionParseError)
	}

	if !semver.IsValid(version2Normalized) {
		return 0, fmt.Errorf("invalid version %s: %w", version2, errVersionParseError)
	}

	return semver.Compare(version1Normalized, version2Normalized), nil
}

// Install installs Go to the specified directory, either from the latest version or from a provided archive.
// It handles privilege elevation and all output messaging.
// The installDir should typically be "/usr/local/go". If archivePath is empty, the latest version is installed
// only if Go is not already installed. If archivePath is provided, it will install from that archive directly
// without any checks for existing installations, download logic, validation, or verification - just direct extraction.
func (i *Installer) Install(installDir, archivePath string) error {
	logger.Debugf(

		"Starting InstallGo: installDir=%s, archivePath=%s",

		installDir, archivePath)

	logger.Debugf("Install: archivePath='%s', checking for existing installation", archivePath)

	if archivePath == "" {
		// Check if Go is already installed before proceeding with latest version installation
		installedVersion, err := i.verifyService.GetInstalledVersion(installDir)
		logger.Debugf("Install: verifyService.GetInstalledVersion returned version='%s', err='%v'", installedVersion, err)

		if err != nil {
			return fmt.Errorf("failed to check existing Go installation: %w", err)
		}

		if installedVersion != "" {
			logger.Infof("Go (%s) is already installed in %s.", installedVersion, installDir)
			logger.Infof("Use 'goUpdater update' to update or provide an archive path to force installation.")

			return nil
		}

		logger.Debugf("Install: no existing installation found, proceeding with latest version install")
		// Install latest version - Go is not installed
		err = i.privilegeService.ElevateAndExecute(func() error { return i.Latest(installDir) })
		logger.Debugf("Install: Latest install err: %v", err)

		if err != nil {
			return fmt.Errorf("failed to install latest Go: %w", err)
		}

		return nil
	}

	// Install from archive - direct extraction without validation or verification
	err := i.privilegeService.ElevateAndExecute(

		func() error { return i.DirectExtract(archivePath, installDir) })
	logger.Debugf("Archive install err: %v", err)

	if err != nil {
		return fmt.Errorf("failed to install Go from archive: %w", err)
	}

	return nil
}

// DirectExtract extracts the Go archive directly to the specified directory.
// The installDir should typically be "/usr/local/go".
func (i *Installer) DirectExtract(archivePath, installDir string) error {
	logger.Debugf("Starting direct Go extraction: archive=%s, installDir=%s", archivePath, installDir)

	err := i.prepareInstallDir(installDir)
	if err != nil {
		return err
	}

	logger.Debugf("Extracting archive to: %s", filepath.Dir(installDir))

	err = i.archiveService.Extract(archivePath, filepath.Dir(installDir))
	if err != nil {
		return &InstallError{
			Phase:     "extract",
			FilePath:  archivePath,
			Operation: "extract",
			Err:       err,
		}
	}

	expectedVersion := i.archiveService.ExtractVersion(archivePath)
	logger.Debugf("Expected version from filename: %s", expectedVersion)

	// Post-installation verification against expected version from archive filename
	err = i.verifyService.Installation(installDir, expectedVersion)
	if err != nil {
		return &InstallError{
			Phase:     "verify",
			FilePath:  installDir,
			Operation: "verify",
			Err:       err,
		}
	}

	logger.Infof("Go successfully installed to %s", installDir)

	return nil
}

// Extract extracts the Go archive to the specified installation directory.
// The installDir should typically be "/usr/local/go".
func (i *Installer) Extract(archivePath, installDir, _ string) error {
	logger.Debugf("Starting Go installation: archive=%s, installDir=%s",
		archivePath, installDir)

	err := i.archiveService.Validate(archivePath)
	if err != nil {
		logger.Debugf("Extract: archiveService.Validate returned err: %v", err)

		return fmt.Errorf("failed to validate archive: %w", err)
	}

	err = i.prepareInstallDir(installDir)
	if err != nil {
		return err
	}

	logger.Debugf("Extracting archive to: %s", filepath.Dir(installDir))

	err = i.archiveService.Extract(archivePath, filepath.Dir(installDir))
	if err != nil {
		logger.Debugf("Extract: archiveService.Extract returned err: %v", err)

		return &InstallError{
			Phase:     "extract",
			FilePath:  archivePath,
			Operation: "extract",
			Err:       err,
		}
	}

	extractedVersion := i.archiveService.ExtractVersion(archivePath)
	logger.Debugf("Actual extracted version: %s", extractedVersion)

	logger.Debug("Go installation completed successfully")

	return nil
}

// ExtractWithVerification extracts the Go archive to the specified installation directory,
// and verifies the installation afterwards.
// The installDir should typically be "/usr/local/go".
func (i *Installer) ExtractWithVerification(archivePath, installDir, checksum string) error {
	logger.Debugf("Starting Go installation with verification: archive=%s, installDir=%s, checksum=%s",
		archivePath, installDir, checksum)

	// Extract expected version from filename
	expectedVersion := i.archiveService.ExtractVersion(archivePath)
	logger.Debugf("Expected version from filename: %s", expectedVersion)

	err := i.Extract(archivePath, installDir, checksum)
	if err != nil {
		return err
	}

	err = i.verifyService.Installation(installDir, expectedVersion)
	if err != nil {
		return &InstallError{
			Phase:     "verify",
			FilePath:  installDir,
			Operation: "verify",
			Err:       err,
		}
	}

	logger.Infof("Go successfully installed to %s", installDir)

	return nil
}

// Latest downloads the latest Go version and installs it to the specified directory.
// The installDir should typically be "/usr/local/go".
func (i *Installer) Latest(installDir string) error {
	logger.Debugf(

		"Starting latest Go installation: installDir=%s",

		installDir)

	tempDir, err := i.fs.MkdirTemp("", "goUpdater-install-*")
	if err != nil {
		return &InstallError{
			Phase:     "prepare",
			FilePath:  "",
			Operation: "tempdir",
			Err:       err,
		}
	}

	defer func() { _ = i.fs.RemoveAll(tempDir) }()

	archivePath, _, err := i.downloadService.GetLatest(tempDir)
	if err != nil {
		return &InstallError{
			Phase:     "download",
			FilePath:  tempDir,
			Operation: "download",
			Err:       err,
		}
	}

	logger.Debugf("Downloaded archive: %s", archivePath)

	err = i.ExtractWithVerification(archivePath, installDir, "")
	if err != nil {
		return err
	}

	return nil
}

// HandleExistingInstallation handles the case when Go is already installed.
// It compares versions and prompts for update if necessary.
//
//nolint:funlen // function handles complex user interaction logic
func (i *Installer) HandleExistingInstallation(_ string, installedVersion string) {
	// Get latest version info
	latestVersionInfo, err := i.downloadService.GetLatestVersionInfo()
	if err != nil {
		logger.Errorf("Error fetching latest Go version: %v", err)
		os.Exit(1)
	}

	logger.Debugf("Raw latest version from API: %s", latestVersionInfo.Version)

	logger.Debugf("latestVersionInfo.Version before TrimPrefix: %q (len=%d)",
		latestVersionInfo.Version, len(latestVersionInfo.Version))
	logger.Debugf("latestVersionInfo.Version starts with 'go': %t",
		strings.HasPrefix(latestVersionInfo.Version, "go"))
	latestVersion := strings.TrimPrefix(latestVersionInfo.Version, "go")
	logger.Debugf("latestVersionInfo.Version after TrimPrefix: %q (len=%d)",
		latestVersion, len(latestVersion))

	if latestVersion == "" {
		logger.Errorf(

			"Failed to parse latest version: invalid or empty version from API: %s",

			latestVersionInfo.Version)
		os.Exit(1)
	}

	logger.Debugf("Parsed latest version: %s", latestVersion)

	// Compare versions using semantic versioning
	result, err := i.versionService.Compare(installedVersion, latestVersion)
	if err != nil {
		logger.Errorf("Error comparing versions: %v", err)
		os.Exit(1)
	}

	if result >= 0 {
		logger.Infof(

			"Go (%s) is already installed.",

			strings.TrimPrefix(installedVersion, "go"))

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
	logger.Info("Update functionality should be handled by cmd layer")
}

// prepareInstallDir prepares the installation directory.
func (i *Installer) prepareInstallDir(installDir string) error {
	logger.Debugf("Preparing installation directory: %s", installDir)

	err := i.fs.MkdirAll(filepath.Dir(installDir), directoryPermissions)
	if err != nil {
		return &InstallError{
			Phase:     "prepare",
			FilePath:  filepath.Dir(installDir),
			Operation: "mkdir",
			Err:       err,
		}
	}

	logger.Debug("Installation directory prepared")

	return nil
}

// RunInstall executes the install command logic.
// It installs Go to the specified directory, either from the latest version or from a provided archive.
// installDir is the directory where Go should be installed (e.g., "/usr/local/go").
// archivePath is the path to the archive file to install from, or empty to install the latest version.
// WARNING: This function performs real filesystem operations and should not be used in tests.
func RunInstall(installDir, archivePath string) error {
	// Create service implementations for dependency injection
	archiveSvc := &archiveServiceImpl{
		extractor: archive.NewExtractor(&filesystem.OSFileSystem{}, &archive.DefaultProcessor{}),
	}
	downloadSvc := &downloadServiceImpl{
		downloader: download.NewDownloader(
			&filesystem.OSFileSystem{},
			download.NewHTTPClient(),
			&exec.OSCommandExecutor{},
			&defaultVersionFetcherImpl{},
		),
	}
	verifySvc := verify.NewVerifier(&filesystem.OSFileSystem{}, &exec.OSCommandExecutor{})
	versionSvc := &versionServiceImpl{}
	privilegeSvc := privileges.NewPrivilegeManager(
		&filesystem.OSFileSystem{},
		privileges.OSPrivilegeManagerImpl{},
		&exec.OSCommandExecutor{},
	)

	installer := NewInstallerWithDeps(
		&filesystem.OSFileSystem{},
		archiveSvc,
		downloadSvc,
		verifySvc,
		versionSvc,
		privilegeSvc,
	)

	return installer.Install(installDir, archivePath)
}

// archiveServiceImpl implements ArchiveService using archive package functions.
type archiveServiceImpl struct {
	extractor *archive.Extractor
}

func (a *archiveServiceImpl) Validate(archivePath string) error {
	err := a.extractor.Validate(archivePath)
	logger.Debugf("archiveServiceImpl.Validate: extractor.Validate returned err: %v (type: %T)", err, err)

	if err != nil {
		logger.Debugf("archiveServiceImpl.Validate: wrapping non-nil error")

		return fmt.Errorf("failed to validate archive: %w", err)
	}

	logger.Debugf("archiveServiceImpl.Validate: validation successful, returning nil")

	return nil
}

func (a *archiveServiceImpl) Extract(archivePath, destDir string) error {
	err := a.extractor.Extract(archivePath, destDir)
	if err != nil {
		logger.Debugf("archiveServiceImpl.Extract: extractor.Extract returned err: %v", err)

		return fmt.Errorf("failed to extract archive: %w", err)
	}

	return nil
}

func (a *archiveServiceImpl) ExtractVersion(archivePath string) string {
	return archive.ExtractVersion(archivePath)
}

// downloadServiceImpl implements DownloadService using download package functions.
type downloadServiceImpl struct {
	downloader *download.Downloader
}

func (d *downloadServiceImpl) GetLatest(tempDir string) (string, string, error) {
	archivePath, checksum, err := d.downloader.GetLatest(tempDir)
	if err != nil {
		return "", "", fmt.Errorf("failed to download latest: %w", err)
	}

	return archivePath, checksum, nil
}

func (d *downloadServiceImpl) GetLatestVersionInfo() (struct{ Version string }, error) {
	info, err := download.GetLatestVersionInfo(&defaultVersionFetcherImpl{})
	if err != nil {
		return struct{ Version string }{}, fmt.Errorf("failed to get latest version info: %w", err)
	}

	return struct{ Version string }{Version: info.Version}, nil
}

// defaultVersionFetcherImpl implements VersionFetcher for download service.
type defaultVersionFetcherImpl struct{}

func (d *defaultVersionFetcherImpl) GetLatestVersion() (*httpclient.GoVersionInfo, error) {
	info, err := httpclient.GetLatestVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}

	return info, nil
}

// versionServiceImpl implements VersionService using version package functions.
type versionServiceImpl struct{}

func (v *versionServiceImpl) Compare(installedVersion, latestVersion string) (int, error) {
	result, err := compare(installedVersion, latestVersion)
	if err != nil {
		return 0, fmt.Errorf("failed to compare versions: %w", err)
	}

	return result, nil
}
