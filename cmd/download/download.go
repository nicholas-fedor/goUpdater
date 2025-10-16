// Package download provides the download command for goUpdater.
// It handles downloading the latest Go version archive from official sources.
package download

import (
	"os"

	"github.com/nicholas-fedor/goUpdater/internal/download"
	"github.com/spf13/cobra"
)

// NewDownloadCmd creates the download command.
func NewDownloadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "download",
		Short: "Download the latest Go version archive",
		Long: `Download the latest stable Go version archive for the current platform.
The archive will be downloaded to a temporary directory and verified with its checksum.`,
		Aliases:                nil,
		SuggestFor:             nil,
		GroupID:                "",
		Example:                "",
		ValidArgs:              nil,
		ValidArgsFunction:      nil,
		Args:                   nil,
		ArgAliases:             nil,
		BashCompletionFunction: "",
		Deprecated:             "",
		Annotations:            nil,
		Version:                "",
		PersistentPreRun:       nil,
		PersistentPreRunE:      nil,
		PreRun:                 nil,
		PreRunE:                nil,
		Run: func(_ *cobra.Command, _ []string) {
			_, _, err := download.Download()
			if err != nil {
				os.Exit(1)
			}
		},
		RunE:               nil,
		PostRun:            nil,
		PostRunE:           nil,
		PersistentPostRun:  nil,
		PersistentPostRunE: nil,
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: false},
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:         false,
			DisableNoDescFlag:         false,
			DisableDescriptions:       false,
			HiddenDefaultCmd:          false,
			DefaultShellCompDirective: nil,
		},
		TraverseChildren:           false,
		Hidden:                     false,
		SilenceErrors:              false,
		SilenceUsage:               false,
		DisableFlagParsing:         false,
		DisableAutoGenTag:          false,
		DisableFlagsInUseLine:      false,
		DisableSuggestions:         false,
		SuggestionsMinimumDistance: 0,
	}
}
