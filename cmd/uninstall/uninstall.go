// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package uninstall provides the uninstall command for goUpdater.
// It handles removing Go installations from the system.
package uninstall

import (
	"os"

	"github.com/nicholas-fedor/goUpdater/internal/privileges"
	"github.com/nicholas-fedor/goUpdater/internal/uninstall"
	"github.com/spf13/cobra"
)

// createUninstallCommand creates the cobra command with basic configuration.
// It sets up the command structure and flags for the uninstall command.
func createUninstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall Go from the system",
		Long: `Uninstall Go by removing the installation directory.
By default, this removes Go from /usr/local/go.`,
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

	cmd.Flags().StringP("install-dir", "d", "/usr/local/go", "Directory from which to uninstall Go")

	return cmd
}

// NewUninstallCmd creates the uninstall command.
func NewUninstallCmd() *cobra.Command {
	cmd := createUninstallCommand()

	cmd.Run = func(cmd *cobra.Command, _ []string) {
		installDir, _ := cmd.Flags().GetString("install-dir")

		err := privileges.ElevateAndExecute(func() error {
			return uninstall.Remove(installDir)
		})
		if err != nil {
			cmd.PrintErrln(err)
			os.Exit(1)
		}
	}

	return cmd
}
