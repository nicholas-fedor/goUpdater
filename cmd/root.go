// Package cmd provides the command-line interface for goUpdater.
package cmd

import (
	"os"

	"github.com/nicholas-fedor/goUpdater/internal/logger"
	"github.com/spf13/cobra"
)

// NewRootCmd creates the base command when called without any subcommands.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "goUpdater",
		Short: "A tool for automating Go version updates",
		Long: `goUpdater provides commands to download, install, update, and verify Go installations on your system.
It automates keeping Go installations up-to-date with the latest stable releases from the official Go website.`,
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
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			verbose, _ := cmd.Flags().GetBool("verbose")
			logger.SetVerbose(verbose)
		},
		PersistentPreRunE:  nil,
		PreRun:             nil,
		PreRunE:            nil,
		Run:                nil,
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
	cmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging")

	return cmd
}

// Execute runs the root command.
// This is called by main.main().
func Execute(rootCmd *cobra.Command) {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
