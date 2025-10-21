// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package uninstall

import (
	"fmt"

	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

// Uninstaller handles Go installation removal operations.
type Uninstaller interface {
	Remove(installDir string) error
}

// DefaultUninstaller implements Uninstaller using dependency injection.
type DefaultUninstaller struct {
	fs filesystem.FileSystem
}

// NewDefaultUninstaller creates a new DefaultUninstaller with the provided filesystem.
func NewDefaultUninstaller(fs filesystem.FileSystem) *DefaultUninstaller {
	return &DefaultUninstaller{fs: fs}
}

// Remove removes the Go installation from the specified directory.
// It returns an error if the removal fails. If the directory does not exist,
// it returns nil as the uninstallation is already complete (idempotent behavior).
// The method first checks if the directory exists before attempting removal.
func (d *DefaultUninstaller) Remove(installDir string) error {
	logger.Info("Starting Go uninstallation")
	logger.Debugf("Attempting to remove Go installation from directory: %s", installDir)

	// Check if the directory exists
	_, err := d.fs.Stat(installDir)
	if err != nil {
		if d.fs.IsNotExist(err) && installDir != "" {
			logger.Info("Go is already uninstalled")

			return nil
		}

		logger.Debugf("Failed to check Go installation directory %s: %v", installDir, err)

		return fmt.Errorf("uninstall failed: %w", ErrCheckInstallDir)
	}

	logger.Info("Removing Go installation directory")

	// Remove the directory and all its contents
	err = d.fs.RemoveAll(installDir)
	if err != nil {
		logger.Debugf("Failed to remove Go installation directory %s: %v", installDir, err)

		return fmt.Errorf("uninstall failed: %w", ErrRemoveFailed)
	}

	logger.Info("Go successfully uninstalled")

	return nil
}
