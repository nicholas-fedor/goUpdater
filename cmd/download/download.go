// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package download provides the download command for goUpdater.
// It handles downloading the latest Go version archive from official sources.
package download

import (
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
		RunE: func(_ *cobra.Command, _ []string) error {
			return download.RunDownload()
		},
	}
}
