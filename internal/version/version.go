// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package version provides comprehensive version information for goUpdater.
// It supports both release binaries built with GoReleaser (using ldflags)
// and source builds (using debug.ReadBuildInfo for VCS information).
// The package provides version, commit hash, and build date information.
// It also handles version command execution including flag parsing and output formatting.
package version

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

// versionMutex protects access to global version variables.
var versionMutex sync.Mutex //nolint:gochecknoglobals

// versionOnce ensures debug.ReadBuildInfo is only called once.
var versionOnce sync.Once //nolint:gochecknoglobals

// version is set at build time using ldflags.
var version string

// commit is set at build time using ldflags.
var commit string //nolint:gochecknoglobals

// date is set at build time using ldflags.
var date string //nolint:gochecknoglobals

// goVersion is set at build time using ldflags.
var goVersion string //nolint:gochecknoglobals

const (
	formatDefault outputFormat = ""
	formatShort   outputFormat = "short"
	formatVerbose outputFormat = "verbose"
	formatJSON    outputFormat = "json"
)

// platform is set at build time using ldflags.
var platform string //nolint:gochecknoglobals

// VersionInfo holds detailed version information for the application.
// It includes version, commit hash, build date, Go version, and platform.
//
//revive:disable:exported
type VersionInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"goVersion"`
	Platform  string `json:"platform"`
}

// Info holds version information for the application.
// It encapsulates version, commit, date, Go version, and platform data that can be set at build time.
type Info struct {
	version   string
	commit    string
	date      string
	goVersion string
	platform  string
}

// outputFormat represents the different output formats for version information.
type outputFormat string

// SetVersion sets the version string, typically called via ldflags at build time.
func SetVersion(v string) {
	versionMutex.Lock()

	version = v

	versionMutex.Unlock()
}

// SetCommit sets the commit hash, typically called via ldflags at build time.
func SetCommit(c string) {
	versionMutex.Lock()

	commit = c

	versionMutex.Unlock()
}

// SetDate sets the build date, typically called via ldflags at build time.
func SetDate(d string) {
	versionMutex.Lock()

	date = d

	versionMutex.Unlock()
}

// SetGoVersion sets the Go version, typically called via ldflags at build time.
func SetGoVersion(gv string) {
	versionMutex.Lock()

	goVersion = gv

	versionMutex.Unlock()
}

// SetPlatform sets the platform/architecture, typically called via ldflags at build time.
func SetPlatform(p string) {
	versionMutex.Lock()

	platform = p

	versionMutex.Unlock()
}

// Get returns the current version of goUpdater.
// The version is typically set during the build process using linker flags
// to inject the actual version from CI/CD or release tags.
// Returns "dev" if no version has been set at build time.
func Get() string {
	return getInfo().version
}

// GetCommit returns the Git commit hash for the build.
// For release binaries, this is set via ldflags during build.
// For source builds, it is retrieved from VCS information if available.
// Returns an empty string if no commit information is available.
func GetCommit() string {
	return getInfo().commit
}

// GetDate returns the build date in RFC3339 format.
// For release binaries, this is set via ldflags during build.
// For source builds, it is retrieved from VCS time information.
// Returns an empty string if no date information is available.
func GetDate() string {
	return getInfo().date
}

// GetVersionInfo returns comprehensive version information.
// It provides version, commit hash, build date, Go version, and platform in a structured format.
// This function is used by commands that need to display detailed version details.
func GetVersionInfo() VersionInfo {
	info := getInfo()

	return VersionInfo{
		Version:   info.version,
		Commit:    info.commit,
		Date:      info.date,
		GoVersion: info.goVersion,
		Platform:  info.platform,
	}
}

// GetVersion displays version information with the given flags.
// It handles output formatting for the version display.
func GetVersion(w io.Writer, format string, jsonFlag, shortFlag, verboseFlag bool) {
	displayVersion(w, determineFormat(format, jsonFlag, shortFlag, verboseFlag))
}

// getInfo returns the version information, initializing it if necessary.
// This function ensures thread-safe initialization of build information.
func getInfo() Info { //nolint:gocognit,cyclop
	versionMutex.Lock()

	info := Info{
		version:   version,
		commit:    commit,
		date:      date,
		goVersion: goVersion,
		platform:  platform,
	}

	versionMutex.Unlock()

	// If version is not set via ldflags, default to "dev"
	if info.version == "" {
		info.version = "dev"
	}

	versionOnce.Do(func() {
		// Only use debug.ReadBuildInfo if commit/date are not set via ldflags
		if info.commit == "" || info.date == "" { //nolint:nestif
			if bi, ok := debug.ReadBuildInfo(); ok {
				for _, setting := range bi.Settings {
					switch setting.Key {
					case "vcs.revision":
						if info.commit == "" {
							info.commit = setting.Value
						}
					case "vcs.time":
						if info.date == "" {
							t, err := time.Parse(time.RFC3339, setting.Value)
							if err == nil {
								info.date = t.Format(time.RFC3339)
							}
						}
					}
				}
			}
		}
	})

	return info
}

// determineFormat determines the output format based on the provided flags.
// It prioritizes shorthand flags over the format flag and returns the appropriate outputFormat.
func determineFormat(format string, jsonFlag, shortFlag, verboseFlag bool) outputFormat {
	if jsonFlag {
		return formatJSON
	}

	if shortFlag {
		return formatShort
	}

	if verboseFlag {
		return formatVerbose
	}

	switch format {
	case "short":
		return formatShort
	case "verbose":
		return formatVerbose
	case "json":
		return formatJSON
	default:
		return formatDefault
	}
}

// displayVersion displays version information in the specified format.
// It handles different output formats including default tree-like structure,
// short version only, verbose with all details, and JSON format.
func displayVersion(writer io.Writer, format outputFormat) {
	info := GetVersionInfo()

	switch format {
	case formatShort:
		displayShort(writer, info)
	case formatVerbose:
		displayVerbose(writer, info)
	case formatJSON:
		displayJSON(writer, info)
	case formatDefault:
		displayDefault(writer, info)
	}
}

// displayDefault displays version information in a tree-like structure with visual hierarchy.
// It follows modern CLI patterns similar to kubectl/docker, using consistent alignment
// and separators for clean layout.
func displayDefault(writer io.Writer, info VersionInfo) {
	_, _ = fmt.Fprintf(writer, "goUpdater %s\n", info.Version)

	if info.Commit != "" {
		_, _ = fmt.Fprintf(writer, "├─ Commit: %s\n", info.Commit)
	}

	if info.Date != "" {
		t, err := time.Parse(time.RFC3339, info.Date)
		if err == nil {
			formatted := t.UTC().Format("January 2, 2006 at 3:04 PM UTC")
			_, _ = fmt.Fprintf(writer, "├─ Built: %s\n", formatted)
		} else {
			_, _ = fmt.Fprintf(writer, "├─ Built: %s\n", info.Date)
		}
	}

	if info.GoVersion != "" {
		_, _ = fmt.Fprintf(writer, "├─ Go version: %s\n", info.GoVersion)
	}

	if info.Platform != "" {
		_, _ = fmt.Fprintf(writer, "└─ Platform: %s\n", info.Platform)
	}
}

// displayShort displays only the version number.
// This is useful for scripts and automation that only need the version.
func displayShort(writer io.Writer, info VersionInfo) {
	_, _ = fmt.Fprintf(writer, "%s\n", info.Version)
}

// displayVerbose displays all available version information.
// It shows every field that has a value, providing complete information.
func displayVerbose(writer io.Writer, info VersionInfo) {
	_, _ = fmt.Fprintf(writer, "Version: %s\n", info.Version)

	if info.Commit != "" {
		_, _ = fmt.Fprintf(writer, "Commit: %s\n", info.Commit)
	}

	if info.Date != "" {
		t, err := time.Parse(time.RFC3339, info.Date)
		if err == nil {
			formatted := t.UTC().Format("January 2, 2006 at 3:04 PM UTC")
			_, _ = fmt.Fprintf(writer, "Built: %s\n", formatted)
		} else {
			_, _ = fmt.Fprintf(writer, "Built: %s\n", info.Date)
		}
	}

	if info.GoVersion != "" {
		_, _ = fmt.Fprintf(writer, "Go version: %s\n", info.GoVersion)
	}

	if info.Platform != "" {
		_, _ = fmt.Fprintf(writer, "Platform: %s\n", info.Platform)
	}
}

// Compare compares two Go version strings.
// Returns -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2.
// This function handles semantic versioning for Go versions correctly,
// treating missing version parts as 0 (e.g., "1.21" is equivalent to "1.21.0").
func Compare(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := max(len(parts2), len(parts1))

	for len(parts1) < maxLen {
		parts1 = append(parts1, "0")
	}

	for len(parts2) < maxLen {
		parts2 = append(parts2, "0")
	}

	for index := range maxLen {
		if cmp := compareVersionParts(parts1[index], parts2[index]); cmp != 0 {
			return cmp
		}
	}

	return 0
}

// compareVersionParts compares two version parts.
// Returns -1 if part1 < part2, 0 if part1 == part2, 1 if part1 > part2.
// Handles both numeric and string comparisons, with numeric taking precedence.
func compareVersionParts(part1, part2 string) int {
	num1, err1 := strconv.Atoi(part1)
	num2, err2 := strconv.Atoi(part2)

	switch {
	case err1 != nil || err2 != nil:
		if part1 < part2 {
			return -1
		} else if part1 > part2 {
			return 1
		}

		return 0
	case num1 < num2:
		return -1
	case num1 > num2:
		return 1
	default:
		return 0
	}
}

// displayJSON displays version information in JSON format.
// This is useful for programmatic consumption and integration with other tools.
func displayJSON(writer io.Writer, info VersionInfo) {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")

	err := encoder.Encode(info)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
	}
}
