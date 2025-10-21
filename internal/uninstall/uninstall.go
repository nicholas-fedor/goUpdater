// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package uninstall provides functionality to remove existing Go installations.
// It handles safe removal of Go directories and cleanup operations.
package uninstall

import (
	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
)

// RunUninstall removes the Go installation from the specified directory.
// It returns an error if the removal fails or if the directory does not exist.
// This function creates a DefaultUninstaller with real dependencies for backward compatibility.
// WARNING: This function performs real filesystem operations and should not be used in tests.
func RunUninstall(installDir string) error {
	uninstaller := NewDefaultUninstaller(&filesystem.OSFileSystem{})

	return uninstaller.Remove(installDir)
}
