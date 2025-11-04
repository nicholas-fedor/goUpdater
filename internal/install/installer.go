package install

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

const directoryPermissions = 0755 // Default directory permissions for installation

// NewInstaller creates a new Installer with the provided dependencies.
func NewInstaller(fileSystem filesystem.FileSystem, reader io.Reader) *Installer {
	return &Installer{
		fs:               fileSystem,
		archiveService:   nil,
		downloadService:  nil,
		verifyService:    nil,
		versionService:   nil,
		privilegeService: nil,
		reader:           reader,
	}
}

// NewInstallerWithDeps creates a new Installer with all dependencies injected.
func NewInstallerWithDeps(
	fileSystem filesystem.FileSystem,
	archiveSvc ArchiveService,
	downloadSvc DownloadService,
	verifySvc VerifyService,
	versionSvc VersionService,
	privilegeSvc PrivilegeService,
	reader io.Reader,
) *Installer {
	return &Installer{
		fs:               fileSystem,
		archiveService:   archiveSvc,
		downloadService:  downloadSvc,
		verifyService:    verifySvc,
		versionService:   versionSvc,
		privilegeService: privilegeSvc,
		reader:           reader,
	}
}

// Install installs Go to the specified directory, either from the latest version or from a provided archive.
// It handles privilege elevation and all output messaging.
// The installDir should typically be "/usr/local/go". If archivePath is empty, the latest version is installed
// only if Go is not already installed. If archivePath is provided, it will install from that archive directly
// without any checks for existing installations, download logic, validation, or verification - just direct extraction.
func (i *Installer) Install(installDir, archivePath string) error {
	logger.Debugf("Starting InstallGo: installDir=%s, archivePath=%s", installDir, archivePath)

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

	err := i.archiveService.Validate(archivePath, filepath.Dir(installDir))
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
// It compares versions and prompts for update if necessary. Returns an error
// if any operation fails, allowing the caller to decide how to handle failures.
func (i *Installer) HandleExistingInstallation(_ string, installedVersion string) error { //nolint:lll // function handles complex user interaction logic
	// Get latest version info
	latestVersionInfo, err := i.downloadService.GetLatestVersionInfo()
	if err != nil {
		return fmt.Errorf("failed to fetch latest Go version info: %w", err)
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
		return fmt.Errorf("failed to parse latest version: invalid or empty version from API: %s", //nolint:err113,lll // dynamic error needed for context
			latestVersionInfo.Version)
	}

	logger.Debugf("Parsed latest version: %s", latestVersion)

	// Compare versions using semantic versioning
	result, err := i.versionService.Compare(installedVersion, latestVersion)
	if err != nil {
		return fmt.Errorf("failed to compare versions: %w", err)
	}

	if result >= 0 {
		logger.Infof("Go (%s) is already installed.", strings.TrimPrefix(installedVersion, "go"))

		return nil
	}

	// Go is installed but not latest, prompt user for confirmation
	logger.Infof("Go (%s) is already installed. Would you like to update to (%s)? (Y/n): ",
		strings.TrimPrefix(installedVersion, "go"), latestVersion)

	reader := bufio.NewReader(i.reader)

	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read user input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response == "n" || response == "no" {
		logger.Info("Update cancelled.")

		return nil
	}

	// Proceed with update (default is yes)
	logger.Info("Update functionality should be handled by cmd layer")

	return nil
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
