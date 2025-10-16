// Package version provides the version command for goUpdater.
// It handles displaying version information in various formats.
package version

import (
	"github.com/nicholas-fedor/goUpdater/internal/version"
	"github.com/spf13/cobra"
)

// NewVersionCmd creates the version command.
// It returns a cobra.Command that displays detailed version information of goUpdater
// when executed. The version information includes version, commit hash, build date,
// Go version, and platform, retrieved from the internal/version package and printed
// to stdout in a formatted manner based on the specified output format.
func NewVersionCmd() *cobra.Command {
	var (
		format                           string
		jsonFlag, shortFlag, verboseFlag bool
	)

	cmd := &cobra.Command{ //nolint:exhaustruct
		Use:   "version",
		Short: "Display detailed version information of goUpdater",
		Long: `Display detailed version information of goUpdater including version, commit hash,
build date, Go version, and platform. The version information is set at build time
using linker flags for production releases. If no version has been set, it defaults
to "dev" for development builds.

Output formats:
- default: Tree-like structure with visual hierarchy
- short: Only version number
- verbose: All available information
- json: JSON formatted output`,
		Run: func(_ *cobra.Command, _ []string) {
			version.GetVersion(format, jsonFlag, shortFlag, verboseFlag)
		},
	}

	cmd.Flags().StringVar(&format, "format", "", "Output format: default, short, verbose, json")
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output in JSON format (shorthand for --format=json)")
	cmd.Flags().BoolVar(&shortFlag, "short", false, "Output only version number (shorthand for --format=short)")
	cmd.Flags().BoolVar(&verboseFlag, "verbose", false,
		"Output all available information (shorthand for --format=verbose)")

	return cmd
}
