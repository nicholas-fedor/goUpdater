// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package install provides the install command for goUpdater.
// It handles installing Go from downloaded archives to the system.
package install

import (
	"github.com/nicholas-fedor/goUpdater/internal/install"
	"github.com/spf13/cobra"
)

// createInstallCommand creates the cobra command with basic configuration.
// It sets up the command structure, arguments, and flags for the install command.
func createInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [archive-path]",
		Short: "Install the latest Go version",
		Long: `Install the latest Go version by downloading it and extracting to the installation directory.
By default, Go is installed to /usr/local/go. If an archive path is provided,
it will install from that archive instead.`,
		Aliases:                nil,
		SuggestFor:             nil,
		GroupID:                "",
		Example:                "",
		ValidArgs:              nil,
		ValidArgsFunction:      nil,
		Args:                   cobra.MaximumNArgs(1),
		ArgAliases:             nil,
		BashCompletionFunction: "",
		Deprecated:             "",
		Annotations:            nil,
		Version:                "",
		PersistentPreRun:       nil,
		PersistentPreRunE:      nil,
		PreRun:                 nil,
		PreRunE:                nil,
		Run:                    nil,
		RunE:                   nil,
		PostRun:                nil,
		PostRunE:               nil,
		PersistentPostRun:      nil,
		PersistentPostRunE:     nil,
		FParseErrWhitelist:     cobra.FParseErrWhitelist{UnknownFlags: false},
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

	cmd.Flags().StringP("install-dir", "d", "/usr/local/go", "Directory to install Go")

	return cmd
}

// NewInstallCmd creates the install command.
func NewInstallCmd() *cobra.Command {
	cmd := createInstallCommand()

	cmd.Run = func(cmd *cobra.Command, args []string) {
		installDir, _ := cmd.Flags().GetString("install-dir")

		var archivePath string
		if len(args) > 0 {
			archivePath = args[0]
		}

		err := install.Install(installDir, archivePath)
		if err != nil {
			// Error handling is done within InstallGo, but we need to check the return value
			return
		}
	}

	return cmd
}
