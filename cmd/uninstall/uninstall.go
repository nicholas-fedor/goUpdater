// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package uninstall provides the uninstall command for goUpdater.
// It handles removing Go installations from the system.
package uninstall

import (
	"errors"
	"fmt"

	"github.com/nicholas-fedor/goUpdater/internal/logger"
	"github.com/nicholas-fedor/goUpdater/internal/privileges"
	"github.com/nicholas-fedor/goUpdater/internal/uninstall"
	"github.com/spf13/cobra"
)

var errInstallDirEmpty = errors.New("install directory cannot be empty")

// NewUninstallCmd creates the uninstall command.
func NewUninstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall Go from the system",
		Long: `Uninstall Go by removing the installation directory.
By default, this removes Go from /usr/local/go.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			installDir, err := cmd.Flags().GetString("install-dir")
			if err != nil {
				return fmt.Errorf("failed to get install-dir flag: %w", err)
			}

			if installDir == "" {
				return fmt.Errorf("%w", errInstallDirEmpty)
			}

			return privileges.ElevateAndExecute(func() error {
				err := uninstall.RunUninstall(installDir)
				if errors.Is(err, uninstall.ErrInstallDirEmpty) {
					return fmt.Errorf("%w", errInstallDirEmpty)
				}

				logger.Debugf("Uninstall result: err=%v, err==nil=%t", err, err == nil)

				if err != nil {
					return fmt.Errorf("uninstall failed: %w", err)
				}

				return nil
			})
		},
	}

	cmd.Flags().StringP("install-dir", "d", "/usr/local/go", "Directory from which to uninstall Go")

	return cmd
}
