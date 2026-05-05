package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/DesmondSanctity/shipnote/internal/config"
	"github.com/DesmondSanctity/shipnote/internal/runner"
)

// runOptions captures the flags shared by `generate` and `preview`.
type runOptions struct {
	FromRef    string
	ToRef      string
	ConfigPath string
	Token      string
	CacheDir   string
	NoNetwork  bool
}

func resolveRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return cwd, nil
}

func loadConfigForRoot(root, configPath string) (config.Config, string, error) {
	path := config.ResolvedPath(root, configPath)
	cfg, err := config.Load(path)
	return cfg, path, err
}

func defaultToken(flagToken string) string {
	if flagToken != "" {
		return flagToken
	}
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return t
	}
	return os.Getenv("GH_TOKEN")
}

func defaultCacheDir(root string) string {
	return filepath.Join(root, ".shipnote", "cache")
}

func executeRun(ctx context.Context, opts runOptions) (runner.Outputs, config.Config, string, error) {
	root, err := resolveRepoRoot()
	if err != nil {
		return runner.Outputs{}, config.Config{}, "", err
	}
	cfg, _, err := loadConfigForRoot(root, opts.ConfigPath)
	if err != nil {
		return runner.Outputs{}, cfg, root, err
	}
	in := runner.Inputs{
		RepoDir:   root,
		FromRef:   opts.FromRef,
		ToRef:     opts.ToRef,
		Token:     defaultToken(opts.Token),
		Cfg:       cfg,
		CacheDir:  defaultCacheDir(root),
		NoNetwork: opts.NoNetwork,
	}
	out, err := runner.Run(ctx, in)
	return out, cfg, root, err
}

// stringFlag pulls a string flag, panicking on developer error
// (missing flag definitions). Keeps callers concise.
func stringFlag(c *cobra.Command, name string) string {
	v, _ := c.Flags().GetString(name)
	return v
}

func boolFlag(c *cobra.Command, name string) bool {
	v, _ := c.Flags().GetBool(name)
	return v
}

func writeArtifact(absPath string, body []byte) error {
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(absPath), err)
	}
	if err := os.WriteFile(absPath, body, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", absPath, err)
	}
	return nil
}

func absUnder(root, rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Join(root, rel)
}
