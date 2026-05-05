package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/DesmondSanctity/shipnote/internal/site"
)

func newSiteCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "site <output-dir>",
		Short: "Build/update a public static changelog site.",
		Long: "Generates release notes for the requested range and writes a\n" +
			"static, multi-version HTML site (plus JSON Feed, Atom feed, and an\n" +
			"embeddable widget) into <output-dir>. Re-running with the same\n" +
			"input produces byte-identical output.",
		Args: cobra.ExactArgs(1),
		RunE: runSite,
	}
	c.Flags().String("from", "", "Start ref (default: most recent tag)")
	c.Flags().String("to", "HEAD", "End ref")
	c.Flags().String("token", "", "GitHub token (default: $GITHUB_TOKEN)")
	c.Flags().Bool("no-network", false, "Skip GitHub; render git-only changelog")
	c.Flags().String("site-url", "", "Public base URL where the site will be hosted")
	c.Flags().String("site-title", "", "Page title (defaults to '<repo> changelog')")
	addAIFlags(c)
	return c
}

func runSite(cmd *cobra.Command, args []string) error {
	opts := runOptions{
		FromRef:    stringFlag(cmd, "from"),
		ToRef:      stringFlag(cmd, "to"),
		ConfigPath: stringFlag(cmd.Root(), "config"),
		Token:      stringFlag(cmd, "token"),
		NoNetwork:  boolFlag(cmd, "no-network"),
	}
	readAIFlags(cmd, &opts)

	out, _, root, err := executeRun(cmd.Context(), opts)
	if err != nil {
		return err
	}

	dir := args[0]
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(root, dir)
	}

	siteOpts := site.Options{
		Root:    dir,
		SiteURL: stringFlag(cmd, "site-url"),
		Title:   stringFlag(cmd, "site-title"),
	}
	if err := site.Build(siteOpts, out.Release); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "shipnote: wrote site to %s\n", dir)
	return nil
}
