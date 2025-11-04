// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package uninstall

import (
	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
)

// Uninstaller handles Go installation removal operations.
type Uninstaller interface {
	Remove(installDir string) error
}

// DefaultUninstaller implements Uninstaller using dependency injection.
type DefaultUninstaller struct {
	fs filesystem.FileSystem
}
