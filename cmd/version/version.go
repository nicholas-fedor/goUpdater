// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package version provides the version command for goUpdater.
// It handles displaying version information in various formats.
package version

import (
	"errors"
	"fmt"

	"github.com/nicholas-fedor/goUpdater/internal/version"
	"github.com/spf13/cobra"
)

var errMultipleFormatFlags = errors.New("multiple format flags specified: only one of --json, --short, --verbose, or --format can be used") //nolint:lll

// determineDisplayFormat determines the display format based on flags.
// It returns an error if multiple conflicting format flags are set.
func determineDisplayFormat(jsonFlag, shortFlag, verboseFlag bool, format string) (string, error) {
	count := 0

	var chosenFormat string

	if jsonFlag {
		count++
		chosenFormat = "json"
	}

	if shortFlag {
		count++
		chosenFormat = "short"
	}

	if verboseFlag {
		count++
		chosenFormat = "verbose"
	}

	if format != "" {
		count++
		chosenFormat = format
	}

	if count > 1 {
		return "", errMultipleFormatFlags
	}

	if count == 0 {
		return "default", nil
	}

	return chosenFormat, nil
}

// NewVersionCmd creates the version command.
// It returns a cobra.Command that displays detailed version information of goUpdater
// when executed. The version information includes version, commit hash, build date,
// Go version, and platform, retrieved from the internal/version package and printed
// to stdout in a formatted manner based on the specified output format.
func NewVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Display detailed version information of goUpdater",
		Long: `Display detailed version information of goUpdater including version, commit hash,
build date, Go version, and platform. The version information is set at build time
using linker flags for production releases. If no version has been set, it defaults
to "dev" for development builds.

Output formats:
- default: "goUpdater <version>" only
- short: Only version number
- verbose: All available information in tree-like structure
- json: JSON formatted output`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			jsonFlag, err := cmd.Flags().GetBool("json")
			if err != nil {
				return fmt.Errorf("failed to get flag %q: %w", "json", err)
			}

			shortFlag, err := cmd.Flags().GetBool("short")
			if err != nil {
				return fmt.Errorf("failed to get flag %q: %w", "short", err)
			}

			verboseFlag, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return fmt.Errorf("failed to get flag %q: %w", "verbose", err)
			}

			format, err := cmd.Flags().GetString("format")
			if err != nil {
				return fmt.Errorf("failed to get flag %q: %w", "format", err)
			}

			displayFormat, err := determineDisplayFormat(jsonFlag, shortFlag, verboseFlag, format)
			if err != nil {
				return err
			}

			return version.RunVersion(cmd.OutOrStdout(), displayFormat)
		},
	}

	cmd.Flags().String("format", "", "Output format: default, short, verbose, json")
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format (shorthand for --format=json)")
	cmd.Flags().BoolP("short", "s", false, "Output only version number (shorthand for --format=short)")
	cmd.Flags().BoolP("verbose", "v", false,
		"Output all available information (shorthand for --format=verbose)")

	return cmd
}
