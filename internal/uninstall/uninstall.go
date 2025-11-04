// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package uninstall provides functionality to remove existing Go installations.
// It handles safe removal of Go directories and cleanup operations.
package uninstall

import (
	"fmt"
	"path/filepath"

	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

// RunUninstall removes the Go installation from the specified directory.
// It returns an error if the removal fails or if the directory does not exist.
// This function creates a DefaultUninstaller with real dependencies for backward compatibility.
// WARNING: This function performs real filesystem operations and should not be used in tests.
func RunUninstall(installDir string) error {
	uninstaller := NewDefaultUninstaller(&filesystem.OSFileSystem{})

	return uninstaller.Remove(installDir)
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
	if installDir == "" {
		return ErrInstallDirEmpty
	}

	logger.Info("Starting Go uninstallation")
	logger.Debugf("Attempting to remove Go installation from directory: %s", filepath.Base(installDir))

	// Check if the directory exists
	_, err := d.fs.Stat(installDir)
	if err != nil {
		if d.fs.IsNotExist(err) {
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
