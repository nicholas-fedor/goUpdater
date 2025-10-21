// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package uninstall provides the uninstall command for goUpdater.
// It handles removing Go installations from the system.
package uninstall

import (
	"github.com/nicholas-fedor/goUpdater/internal/privileges"
	"github.com/nicholas-fedor/goUpdater/internal/uninstall"
	"github.com/spf13/cobra"
)

// NewUninstallCmd creates the uninstall command.
func NewUninstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall Go from the system",
		Long: `Uninstall Go by removing the installation directory.
By default, this removes Go from /usr/local/go.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			installDir, _ := cmd.Flags().GetString("install-dir")

			return privileges.ElevateAndExecute(func() error {
				return uninstall.RunUninstall(installDir)
			})
		},
	}

	cmd.Flags().StringP("install-dir", "d", "/usr/local/go", "Directory from which to uninstall Go")

	return cmd
}
