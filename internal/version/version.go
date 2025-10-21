// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package version provides comprehensive version information for goUpdater.
// It supports both release binaries built with GoReleaser (using ldflags)
// and source builds (using debug.ReadBuildInfo for VCS information).
// The package provides version, commit hash, and build date information.
// It also handles version command execution including flag parsing and output formatting.
package version

import (
	"fmt"
	"io"
	"runtime"

	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
)

// These variables are set at build time using ldflags.
// They provide version information for release binaries.
var (
	version   string
	commit    string
	date      string
	goVersion string
	platform  string
)

// RunVersion executes the version command logic with the specified format.
// It displays version information in the given format to the provided writer.
func RunVersion(writer io.Writer, format string) error {
	DisplayWithFormat(writer, format)

	return nil
}

// GetVersionInfo retrieves version information from build info and environment.
// It returns a VersionInfo struct populated with version details.
// It first checks for ldflag variables set at build time, then falls back to
// debug.ReadBuildInfo for development builds.
func GetVersionInfo(reader filesystem.DebugInfoReader) Info {
	info := Info{
		Version:   version,
		Commit:    commit,
		Date:      date,
		GoVersion: goVersion,
		Platform:  platform,
	}

	// If version is not set via ldflags, default to "dev"
	if info.Version == "" {
		info.Version = "dev"
	}

	// Try to get missing info from build info (for development builds)
	populateFromBuildInfo(&info, reader)

	// Set platform if not set via ldflags
	if info.Platform == "" {
		info.Platform = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	}

	return info
}

// GetClientVersion returns the current version string.
// It retrieves version information and returns the version field.
// If no version is set, it defaults to "dev".
func GetClientVersion() string {
	info := GetVersionInfo(&filesystem.OSDebugInfoReader{})
	if info.Version != "" {
		return info.Version
	}

	return "dev"
}

// populateFromBuildInfo populates version info from debug.ReadBuildInfo if fields are missing.
//
//nolint:cyclop // Function handles multiple version info fields with conditional logic
func populateFromBuildInfo(info *Info, reader filesystem.DebugInfoReader) {
	if info.Commit != "" && info.Date != "" && info.GoVersion != "" {
		return
	}

	buildInfo, ok := reader.ReadBuildInfo()
	if !ok {
		return
	}

	for _, setting := range buildInfo.Settings {
		switch setting.Key {
		case "vcs.revision":
			if info.Commit == "" {
				info.Commit = setting.Value
			}
		case "vcs.time":
			if info.Date == "" {
				info.Date = setting.Value
			}
		}
	}

	if info.GoVersion == "" {
		info.GoVersion = buildInfo.GoVersion
	}
}
