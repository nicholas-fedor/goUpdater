// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package verify provides the verify command for goUpdater.
// It handles verifying Go installations and displaying version information.
package verify

import (
	"github.com/nicholas-fedor/goUpdater/internal/verify"
	"github.com/spf13/cobra"
)

// NewVerifyCmd creates the verify command.
func NewVerifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify the installed Go version",
		Long: `Verify that Go is properly installed by checking the version.
Displays the currently installed Go version. By default, checks /usr/local/go.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			verifyDir, _ := cmd.Flags().GetString("install-dir")

			return verify.RunVerify(verifyDir)
		},
	}
	cmd.Flags().StringP("install-dir", "d", "/usr/local/go", "Directory to verify Go installation")

	return cmd
}
