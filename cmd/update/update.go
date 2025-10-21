// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package update provides the update command for goUpdater.
// It handles updating Go to the latest version with privilege escalation.
package update

import (
	"errors"
	"fmt"

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
		RunE: func(cmd *cobra.Command, _ []string) error {
			updateDir, _ := cmd.Flags().GetString("install-dir")
			autoInstall, _ := cmd.Flags().GetBool("auto-install")

			err := update.RunUpdate(updateDir, autoInstall)
			if err != nil {
				if errors.Is(err, update.ErrGoNotInstalled) {
					//nolint:forbidigo // required for stdout output when Go is not installed
					fmt.Println("Go is not installed. Use the --auto-install flag to install it automatically.")

					return nil
				}

				return fmt.Errorf("update failed: %w", err)
			}

			return nil
		},
	}
	cmd.Flags().StringP("install-dir", "d", "/usr/local/go", "Directory where Go should be updated")
	cmd.Flags().BoolP("auto-install", "a", false, "Automatically install Go if not present")

	return cmd
}
