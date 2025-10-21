// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package privileges provides functions to detect privileges and request elevation using sudo.
// It handles privilege escalation for system operations that require root access.
package privileges

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"

	"github.com/nicholas-fedor/goUpdater/internal/exec"
	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

// IsRoot reports whether the current process is running as root.
// It creates a default privilege manager and checks the effective user ID.
func IsRoot() bool {
	pm := NewPrivilegeManager(&filesystem.OSFileSystem{}, OSPrivilegeManagerImpl{}, &exec.OSCommandExecutor{})

	return pm.isRoot()
}

// RequestElevation re-executes the current process with sudo if not already running as root.
// It creates a default privilege manager and delegates the elevation request.
func RequestElevation() error {
	pm := NewPrivilegeManager(&filesystem.OSFileSystem{}, OSPrivilegeManagerImpl{}, &exec.OSCommandExecutor{})

	return pm.requestElevation()
}

// HandleElevationError logs and exits with an error message for privilege elevation failures.
// It provides detailed error information and user guidance before terminating the process.
func HandleElevationError(err error) {
	var elevErr *ElevationError
	if errors.As(err, &elevErr) {
		logger.Errorf("Error: Failed to obtain elevated privileges during %s: %s", elevErr.Op, elevErr.Reason)

		if elevErr.SudoErr {
			logger.Error("Installation requires elevated privileges. Please ensure sudo is installed and configured correctly.")
			logger.Error("Try running the command with 'sudo' prefix or as root user.")
		} else {
			logger.Error("Installation requires elevated privileges. Please run with sudo or as root.")
		}

		if elevErr.Cause != nil {
			logger.Errorf("Underlying cause: %v", elevErr.Cause)
		}
	} else {
		logger.Errorf("Error: Failed to obtain elevated privileges: %v", err)
		logger.Error("Installation requires elevated privileges. Please run with sudo or as root.")
	}

	os.Exit(1)
}

// ElevateAndExecute checks for root privileges and requests elevation if necessary.
// If elevation is required and fails, it handles the error appropriately.
// Once elevated or if already running as root, it executes the provided callback function
// and returns any error from the callback.
// The callback should be a function that performs the privileged operation.
// This is a convenience function that creates a default privilege manager.
func ElevateAndExecute(callback func() error) error {
	pm := NewPrivilegeManager(&filesystem.OSFileSystem{}, OSPrivilegeManagerImpl{}, &exec.OSCommandExecutor{})

	return pm.ElevateAndExecute(callback)
}

// isElevated checks if the process is running with elevated privileges via sudo.
// It returns true if the SUDO_USER environment variable is set, indicating
// the process was started with sudo by a different user.
// This is an internal helper function.
func isElevated(pm OSPrivilegeManager) bool {
	return pm.Getenv("SUDO_USER") != ""
}

// getOriginalUserHome retrieves the original user's home directory from the SUDO_USER environment variable.
// It looks up the user by name and returns their home directory path.
// This is an internal helper function.
func getOriginalUserHome(pm OSPrivilegeManager) string {
	sudoUser := pm.Getenv("SUDO_USER")
	logger.Debugf("SUDO_USER environment variable: %s", sudoUser)

	if sudoUser == "" {
		logger.Debug("SUDO_USER environment variable is empty")

		return ""
	}

	originalUser, err := user.Lookup(sudoUser)
	if err != nil {
		logger.Debugf("Failed to lookup user %s: %v", sudoUser, err)

		return ""
	}

	logger.Debugf("Original user home directory resolved: %s", originalUser.HomeDir)

	return originalUser.HomeDir
}

// IsElevated checks if the process is running with elevated privileges via sudo.
// It returns true if the SUDO_USER environment variable is set, indicating
// the process was started with sudo by a different user.
// This is the exported version of the internal isElevated function.
func IsElevated() bool {
	pm := NewPrivilegeManager(&filesystem.OSFileSystem{}, OSPrivilegeManagerImpl{}, &exec.OSCommandExecutor{})

	return isElevated(pm.pm)
}

// GetOriginalUserHome retrieves the original user's home directory from the SUDO_USER environment variable.
// It provides access to the internal getOriginalUserHome function.
func GetOriginalUserHome() string {
	pm := NewPrivilegeManager(&filesystem.OSFileSystem{}, OSPrivilegeManagerImpl{}, &exec.OSCommandExecutor{})

	return getOriginalUserHome(pm.pm)
}

// getSearchDirectories determines the directories to search for existing archives.
// When running with elevated privileges, it includes both the elevated user's directories
// and the original user's directories to find user-downloaded archives.
// It only includes directories that are readable to maintain security.
// This is an internal helper function.
func getSearchDirectories(elevatedHome, destDir string, privilegeManager OSPrivilegeManager, fileSystem filesystem.FileSystem) []string {
	var searchDirs []string

	logger.Debugf("Building search directories: elevatedHome=%s, destDir=%s, isElevated=%t",
		elevatedHome, destDir, isElevated(privilegeManager))

	// Always include elevated user's directories
	if isReadableDir(filepath.Join(elevatedHome, "Downloads"), fileSystem) {
		searchDirs = append(searchDirs, filepath.Join(elevatedHome, "Downloads"))
		logger.Debugf("Added elevated user's Downloads directory: %s", filepath.Join(elevatedHome, "Downloads"))
	} else {
		logger.Debugf("Elevated user's Downloads directory not readable: %s", filepath.Join(elevatedHome, "Downloads"))
	}

	if isReadableDir(elevatedHome, fileSystem) {
		searchDirs = append(searchDirs, elevatedHome)
		logger.Debugf("Added elevated user's home directory: %s", elevatedHome)
	} else {
		logger.Debugf("Elevated user's home directory not readable: %s", elevatedHome)
	}

	// Check if running with elevated privileges and add original user's directories
	if isElevated(privilegeManager) {
		addOriginalUserDirs(&searchDirs, privilegeManager, fileSystem)
	}

	// Always include destination directory
	searchDirs = append(searchDirs, destDir)
	logger.Debugf("Added destination directory: %s", destDir)

	logger.Debugf("Final search directories: %v", searchDirs)

	return searchDirs
}

// addOriginalUserDirs adds the original user's directories to the search list if available.
// It checks for both the Downloads subdirectory and the home directory itself.
// This is an internal helper function.
func addOriginalUserDirs(searchDirs *[]string, pm OSPrivilegeManager, fileSystem filesystem.FileSystem) {
	originalHome := getOriginalUserHome(pm)
	logger.Debugf("Original user home directory: %s", originalHome)

	if originalHome == "" {
		logger.Debug("No original user home directory found")

		return
	}

	if isReadableDir(filepath.Join(originalHome, "Downloads"), fileSystem) {
		*searchDirs = append(*searchDirs, filepath.Join(originalHome, "Downloads"))
		logger.Debugf("Added original user's Downloads directory: %s", filepath.Join(originalHome, "Downloads"))
	} else {
		logger.Debugf("Original user's Downloads directory not readable: %s", filepath.Join(originalHome, "Downloads"))
	}

	if isReadableDir(originalHome, fileSystem) {
		*searchDirs = append(*searchDirs, originalHome)
		logger.Debugf("Added original user's home directory: %s", originalHome)
	} else {
		logger.Debugf("Original user's home directory not readable: %s", originalHome)
	}
}

// isReadableDir checks if a directory exists and is readable.
// It uses the filesystem interface to check directory accessibility.
// This is an internal helper function.
func isReadableDir(dir string, fs filesystem.FileSystem) bool {
	info, err := fs.Stat(dir)
	if err != nil {
		return false
	}

	return info.IsDir()
}

// GetSearchDirectories determines the directories to search for existing archives.
// When running with elevated privileges, it includes both the elevated user's directories
// and the original user's directories to find user-downloaded archives.
// It only includes directories that are readable to maintain security.
// This is the exported version of the internal getSearchDirectories function.
func GetSearchDirectories(elevatedHome, destDir string, fileSystem filesystem.FileSystem) []string {
	pm := NewPrivilegeManager(fileSystem, OSPrivilegeManagerImpl{}, &exec.OSCommandExecutor{})

	return getSearchDirectories(elevatedHome, destDir, pm.pm, fileSystem)
}

// ElevateAndExecuteWithDrop is a convenience function that uses the default privilege manager
// to perform operations with automatic privilege dropping.
// It creates a default privilege manager and delegates the operation.
func ElevateAndExecuteWithDrop(callback func() error) error {
	pm := NewPrivilegeManager(&filesystem.OSFileSystem{}, OSPrivilegeManagerImpl{}, &exec.OSCommandExecutor{})

	return pm.ElevateAndExecuteWithDrop(callback)
}
