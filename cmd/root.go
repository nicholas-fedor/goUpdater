// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package cmd provides the command-line interface for goUpdater.
package cmd

import (
	"fmt"
	"os"

	"github.com/nicholas-fedor/goUpdater/internal/logger"
	"github.com/spf13/cobra"
)

// loggerSetter defines the interface for setting logger verbosity.
type loggerSetter interface {
	SetVerbose(verbose bool)
}

// realLoggerSetterImpl is the default implementation of the loggerSetter interface.
// This is used as the default when no custom setter is provided.
//
//nolint:gochecknoglobals // required for CLI dependency injection
var realLoggerSetterImpl = &realLoggerSetter{}

// realLoggerSetter implements the loggerSetter interface using the internal logger package.
type realLoggerSetter struct{}

// SetVerbose sets the verbose logging level.
func (r *realLoggerSetter) SetVerbose(verbose bool) {
	logger.SetVerbose(verbose)
}

// executor defines the interface for executing commands.
type executor interface {
	Execute(cmd *cobra.Command) error
}

// realExecutorImpl is the default implementation of the executor interface.
// This is used as the default when no custom executor is provided.
//
//nolint:gochecknoglobals // required for CLI dependency injection
var realExecutorImpl = &realExecutor{}

// realExecutor implements the executor interface using cobra.
type realExecutor struct{}

// Execute runs the command.
func (r *realExecutor) Execute(cmd *cobra.Command) error {
	err := cmd.Execute()
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	return nil
}

// setVerboseLogging sets the verbose logging based on the flag.
func setVerboseLogging(cmd *cobra.Command, setter loggerSetter) {
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		cmd.PrintErrf("error retrieving verbose flag: %v\n", err)
		setter.SetVerbose(false)
	} else {
		setter.SetVerbose(verbose)
	}
}

// executeRoot runs the root command.
func executeRoot(rootCmd *cobra.Command, exec executor) error {
	err := exec.Execute(rootCmd)
	if err != nil {
		return fmt.Errorf("failed to execute root command: %w", err)
	}

	return nil
}

// NewRootCmd creates the base command when called without any subcommands.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "goUpdater",
		Short: "A tool for automating Go version updates",
		Long: `goUpdater provides commands to download, install, update, and verify Go installations on your system.
It automates keeping Go installations up-to-date with the latest stable releases from the official Go website.`,
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			setVerboseLogging(cmd, realLoggerSetterImpl)
		},
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: false},
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:         false,
			DisableNoDescFlag:         false,
			DisableDescriptions:       false,
			HiddenDefaultCmd:          false,
			DefaultShellCompDirective: nil,
		},
	}
	cmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging")

	return cmd
}

// Execute runs the root command.
// This is called by main.main().
func Execute(rootCmd *cobra.Command) {
	err := executeRoot(rootCmd, realExecutorImpl)
	if err != nil {
		os.Exit(1)
	}
}
