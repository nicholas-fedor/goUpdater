package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nicholas-fedor/goUpdater/cmd/download"
	"github.com/nicholas-fedor/goUpdater/cmd/install"
	"github.com/nicholas-fedor/goUpdater/cmd/uninstall"
	"github.com/nicholas-fedor/goUpdater/cmd/update"
	"github.com/nicholas-fedor/goUpdater/cmd/verify"
	"github.com/nicholas-fedor/goUpdater/cmd/version"
)

// RegisterCommands adds all subcommands to the root command.
// This function must be called before executing the root command.
func RegisterCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(download.NewDownloadCmd())
	rootCmd.AddCommand(install.NewInstallCmd())
	rootCmd.AddCommand(uninstall.NewUninstallCmd())
	rootCmd.AddCommand(update.NewUpdateCmd())
	rootCmd.AddCommand(verify.NewVerifyCmd())
	rootCmd.AddCommand(version.NewVersionCmd())
}
