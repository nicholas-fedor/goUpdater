// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package uninstall

import "errors"

// ErrInstallationNotFound indicates the Go installation directory does not exist.
var ErrInstallationNotFound = errors.New("go installation directory does not exist")

// ErrCheckInstallDir indicates failure to check the installation directory.
var ErrCheckInstallDir = errors.New("failed to check installation directory")

// ErrRemoveFailed indicates failure to remove the installation directory.
var ErrRemoveFailed = errors.New("failed to remove installation directory")
