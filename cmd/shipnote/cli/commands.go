package cli

import (
	"errors"

	"github.com/spf13/cobra"
)

// errNotImplemented is returned by stub commands. It causes a non-zero exit.
var errNotImplemented = errors.New("not implemented yet — landing in a follow-up PR")

func newInitCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "init",
		Short: "Bootstrap shipnote in the current repo.",
		Long:  "Detects packages, writes .shipnote.toml, and adds .shipnote/ to .gitignore.",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented
		},
	}
	c.Flags().Bool("yes", false, "Accept all defaults without prompting")
	c.Flags().Bool("force", false, "Overwrite existing .shipnote.toml")
	c.Flags().Bool("with-pr-template", false, "Append changelog block to .github/pull_request_template.md")
	return c
}

func newGenerateCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "generate",
		Short: "Generate release notes for a commit range.",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented
		},
	}
	c.Flags().String("from", "", "Start ref (default: most recent tag)")
	c.Flags().String("to", "HEAD", "End ref")
	c.Flags().StringSlice("format", []string{"markdown", "json"}, "Output formats")
	c.Flags().Bool("check", false, "Exit non-zero if regenerating would change output")
	c.Flags().String("token", "", "GitHub token (default: $GITHUB_TOKEN)")
	return c
}

func newPreviewCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "preview",
		Short: "Generate release notes to stdout without writing files.",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented
		},
	}
	c.Flags().String("from", "", "Start ref (default: most recent tag)")
	c.Flags().String("to", "HEAD", "End ref")
	c.Flags().String("format", "markdown", "Output format (markdown or json)")
	return c
}

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose configuration, git state, GitHub access, and detected packages.",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented
		},
	}
}
