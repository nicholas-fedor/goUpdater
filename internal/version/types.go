// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package version

import (
	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
)

// Info holds detailed version information for the application.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"goVersion"`
	Platform  string `json:"platform"`
}

// DisplayManager handles version information operations.
type DisplayManager struct {
	reader  filesystem.DebugInfoReader
	parser  filesystem.TimeParser
	encoder filesystem.JSONEncoder
	writer  filesystem.ErrorWriter
}
