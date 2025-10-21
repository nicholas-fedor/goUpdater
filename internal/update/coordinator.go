// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package update

import (
	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

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
