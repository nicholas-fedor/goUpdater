// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package main provides the entry point for the goUpdater application.
package main

import "github.com/nicholas-fedor/goUpdater/cmd"

// main is the entry point of the goUpdater application.
// It creates the root command, registers all subcommands, and executes it.
func main() {
	rootCmd := cmd.NewRootCmd()
	cmd.RegisterCommands(rootCmd)
	cmd.Execute(rootCmd)
}
