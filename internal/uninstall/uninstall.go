// Package uninstall provides functionality to remove existing Go installations.
// It handles safe removal of Go directories and cleanup operations.
package uninstall

import (
	"errors"
	"fmt"
	"os"

	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

// errInstallationNotFound indicates the Go installation was not found.
var errInstallationNotFound = errors.New("installation not found")

// Remove removes the Go installation from the specified directory.
// It returns an error if the removal fails or if the directory does not exist.
func Remove(installDir string) error {
	logger.Debugf("Starting Go uninstallation: installDir=%s", installDir)

	// Check if the directory exists
	logger.Debug("Checking if installation directory exists")

	_, err := os.Stat(installDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("go installation not found at %s: %w", installDir, errInstallationNotFound)
	}

	// Remove the entire directory
	logger.Debug("Removing installation directory")

	err = os.RemoveAll(installDir)
	if err != nil {
		return fmt.Errorf("failed to remove Go installation: %w", err)
	}

	logger.Debug("Go uninstallation completed successfully")
	logger.Infof("Successfully uninstalled Go from: %s", installDir)

	return nil
}
