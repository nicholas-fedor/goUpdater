// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package uninstall

import "errors"

var (
	// ErrInstallationNotFound indicates the Go installation directory does not exist.
	ErrInstallationNotFound = errors.New("go installation directory does not exist")
	// ErrCheckInstallDir indicates failure to check the installation directory.
	ErrCheckInstallDir = errors.New("failed to check installation directory")
	// ErrInstallDirEmpty indicates that the install directory parameter is empty.
	ErrInstallDirEmpty = errors.New("install directory cannot be empty")
	// ErrRemoveFailed indicates failure to remove the installation directory.
	ErrRemoveFailed = errors.New("failed to remove installation directory")
)
