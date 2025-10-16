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
			verifyDir, _ := cmd.Flags().GetString("install-dir")
			verify.Verify(verifyDir)
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
	cmd.Flags().StringP("install-dir", "d", "/usr/local/go", "Directory to verify Go installation")

	return cmd
}
