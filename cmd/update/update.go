// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package update provides the update command for goUpdater.
// It handles updating Go to the latest version with privilege escalation.
package update

import (
	"os"

	"github.com/nicholas-fedor/goUpdater/internal/logger"
	"github.com/nicholas-fedor/goUpdater/internal/update"
	"github.com/spf13/cobra"
)

// NewUpdateCmd creates the update command.
func NewUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update Go to the latest version",
		Long: `Update Go by downloading the latest version, uninstalling the current installation,
installing the new version, and verifying the installation. By default, Go is updated in /usr/local/go.`,
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
		Run: func(cmd *cobra.Command, _ []string) {
			updateDir, _ := cmd.Flags().GetString("install-dir")
			autoInstall, _ := cmd.Flags().GetBool("auto-install")
			logger.Debugf("Starting update operation: installDir=%s, autoInstall=%t", updateDir, autoInstall)

			err := update.GoWithPrivileges(updateDir, autoInstall)
			if err != nil {
				logger.Errorf("Error updating Go: %v", err)
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
	cmd.Flags().StringP("install-dir", "d", "/usr/local/go", "Directory where Go should be updated")
	cmd.Flags().BoolP("auto-install", "a", false, "Automatically install Go if not present")

	return cmd
}
