// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package install provides the install command for goUpdater.
// It handles installing Go from downloaded archives to the system.
package install

import (
	"github.com/nicholas-fedor/goUpdater/internal/install"
	"github.com/spf13/cobra"
)

// NewInstallCmd creates the install command.
func NewInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [archive-path]",
		Short: "Install the latest Go version",
		Long: `Install the latest Go version by downloading it and extracting to the installation directory.
By default, Go is installed to /usr/local/go. If an archive path is provided,
it will install from that archive instead.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			installDir, _ := cmd.Flags().GetString("install-dir")

			var archivePath string
			if len(args) > 0 {
				archivePath = args[0]
			}

			return install.RunInstall(installDir, archivePath)
		},
	}

	cmd.Flags().StringP("install-dir", "d", "/usr/local/go", "Directory to install Go")

	return cmd
}
