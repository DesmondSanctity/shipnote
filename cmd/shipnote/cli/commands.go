package cli

import (
	"errors"
	"fmt"

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
		RunE:  runGenerate,
	}
	c.Flags().String("from", "", "Start ref (default: most recent tag)")
	c.Flags().String("to", "HEAD", "End ref")
	c.Flags().Bool("check", false, "Exit non-zero if regenerating would change output")
	c.Flags().String("token", "", "GitHub token (default: $GITHUB_TOKEN)")
	c.Flags().Bool("no-network", false, "Skip GitHub; render git-only changelog")
	return c
}

func runGenerate(cmd *cobra.Command, _ []string) error {
	opts := runOptions{
		FromRef:    stringFlag(cmd, "from"),
		ToRef:      stringFlag(cmd, "to"),
		ConfigPath: stringFlag(cmd.Root(), "config"),
		Token:      stringFlag(cmd, "token"),
		NoNetwork:  boolFlag(cmd, "no-network"),
	}
	check := boolFlag(cmd, "check")
	out, cfg, root, err := executeRun(cmd.Context(), opts)
	if err != nil {
		return err
	}
	mdPath := absUnder(root, cfg.Output.Markdown)
	jsPath := absUnder(root, cfg.Output.JSON)
	if check {
		return checkArtifacts(cmd, mdPath, jsPath, out.Markdown, out.JSON)
	}
	if err := writeArtifact(mdPath, []byte(out.Markdown)); err != nil {
		return err
	}
	if err := writeArtifact(jsPath, out.JSON); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "shipnote: wrote %s and %s\n", cfg.Output.Markdown, cfg.Output.JSON)
	return nil
}

func newPreviewCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "preview",
		Short: "Generate release notes to stdout without writing files.",
		RunE:  runPreview,
	}
	c.Flags().String("from", "", "Start ref (default: most recent tag)")
	c.Flags().String("to", "HEAD", "End ref")
	c.Flags().String("format", "markdown", "Output format (markdown or json)")
	c.Flags().String("token", "", "GitHub token (default: $GITHUB_TOKEN)")
	c.Flags().Bool("no-network", false, "Skip GitHub; render git-only changelog")
	return c
}

func runPreview(cmd *cobra.Command, _ []string) error {
	opts := runOptions{
		FromRef:    stringFlag(cmd, "from"),
		ToRef:      stringFlag(cmd, "to"),
		ConfigPath: stringFlag(cmd.Root(), "config"),
		Token:      stringFlag(cmd, "token"),
		NoNetwork:  boolFlag(cmd, "no-network"),
	}
	format := stringFlag(cmd, "format")
	out, _, _, err := executeRun(cmd.Context(), opts)
	if err != nil {
		return err
	}
	w := cmd.OutOrStdout()
	switch format {
	case "markdown", "md", "":
		_, _ = fmt.Fprint(w, out.Markdown)
	case "json":
		_, _ = w.Write(out.JSON)
	default:
		return fmt.Errorf("unknown --format %q (want markdown or json)", format)
	}
	return nil
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
