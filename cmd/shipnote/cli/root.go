// Package cli wires the shipnote command tree.
//
// Subcommands live in their own files but are registered here so that
// the command tree is discoverable from a single place. v0.1 ships
// with placeholder commands that print "not implemented yet" — the
// flag surface and tree shape are stable; behavior lands incrementally.
package cli

import (
	"github.com/spf13/cobra"

	"github.com/DesmondSanctity/shipnote/internal/version"
)

// NewRootCmd returns a fresh root command. Constructed per-call so tests
// can run subcommands in isolation without global state.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "shipnote",
		Short:         "Release intelligence for SDK and developer-platform teams.",
		Long:          "shipnote produces deterministic, audience-targeted release notes from your repository's PRs, labels, tags, and monorepo structure.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       buildVersionString(),
		// Bare `shipnote` invocation prints the banner then the help
		// text. --help / subcommands skip the banner so output stays
		// scriptable.
		Run: func(cmd *cobra.Command, _ []string) {
			if !noColor() {
				printBanner(cmd.OutOrStdout())
			}
			_ = cmd.Help()
		},
	}

	// Global flags. Behavior wired in subsequent PRs.
	root.PersistentFlags().String("config", "", "Path to config file (default: .shipnote.toml)")
	root.PersistentFlags().Bool("no-color", false, "Disable colored output")
	root.PersistentFlags().Bool("verbose", false, "Verbose logging")
	root.PersistentFlags().Bool("debug", false, "Debug logging (very noisy)")

	root.AddCommand(
		newInitCmd(),
		newGenerateCmd(),
		newPreviewCmd(),
		newSiteCmd(),
		newDoctorCmd(),
	)

	return root
}

func buildVersionString() string {
	return version.Version +
		"\n  commit: " + version.Commit +
		"\n  built:  " + version.Date +
		"\n  schema: " + version.SchemaVersion
}
