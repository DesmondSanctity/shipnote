// Package config loads and validates .shipnote.toml.
//
// The schema mirrors the example in thinkering/10-build-plan.md. v0.1
// only supports a single-package repo, so [packages].mode is locked to
// "single" and any other value is a load-time error.
//
// All fields have sensible defaults; an empty/missing file is valid.
package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// DefaultPath is the canonical config filename, resolved relative to
// the repo root.
const DefaultPath = ".shipnote.toml"

// Config is the resolved, validated configuration record.
type Config struct {
	Repo       Repo       `toml:"repo"`
	Packages   Packages   `toml:"packages"`
	Categories Categories `toml:"categories"`
	Output     Output     `toml:"output"`
	Ignore     Ignore     `toml:"ignore"`
	Authors    []Author   `toml:"authors"`
}

// Repo identifies the source repository. "auto" means: detect from the
// origin remote at run time.
type Repo struct {
	Provider string `toml:"provider"`
	Owner    string `toml:"owner"`
	Name     string `toml:"name"`
}

// Packages controls monorepo handling. Only "single" is supported in v0.1.
type Packages struct {
	Mode string `toml:"mode"`
	Root string `toml:"root"` // package id used by the analyzer
}

// Categories maps user-defined aliases to category buckets.
// Empty fields fall back to internal defaults from package analyze.
type Categories struct {
	Features     []string `toml:"features"`
	Fixes        []string `toml:"fixes"`
	Breaking     []string `toml:"breaking"`
	Deprecations []string `toml:"deprecations"`
	Perf         []string `toml:"perf"`
	Security     []string `toml:"security"`
	Docs         []string `toml:"docs"`
	Internal     []string `toml:"internal"`
}

// Output points to the on-disk artifacts shipnote writes.
type Output struct {
	Markdown string `toml:"markdown"`
	JSON     string `toml:"json"`
}

// Ignore lists noise sources that should be excluded from the changelog.
type Ignore struct {
	Authors []string `toml:"authors"`
	Labels  []string `toml:"labels"`
	Paths   []string `toml:"paths"`
}

// Author defines an identity-merge rule. Aliases are case-insensitively
// collapsed into the canonical login.
type Author struct {
	Canonical string   `toml:"canonical"`
	Aliases   []string `toml:"aliases"`
	Name      string   `toml:"name"`
	URL       string   `toml:"url"`
}

// Default returns the zero-config defaults applied when the user has
// no .shipnote.toml or omits a section.
func Default() Config {
	return Config{
		Repo:     Repo{Provider: "github", Owner: "auto", Name: "auto"},
		Packages: Packages{Mode: "single", Root: ""},
		Output: Output{
			Markdown: "CHANGELOG.md",
			JSON:     ".shipnote/release.json",
		},
	}
}

// Load reads and validates the config at path. A missing file returns
// Default() with no error (treated as zero-config).
func Load(path string) (Config, error) {
	cfg := Default()
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, fmt.Errorf("read %s: %w", path, err)
	}
	return parseAndValidate(raw, path)
}

// LoadReader is the io.Reader-flavored sibling of Load. Useful in tests.
func LoadReader(r io.Reader, sourceLabel string) (Config, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return Default(), fmt.Errorf("read %s: %w", sourceLabel, err)
	}
	return parseAndValidate(raw, sourceLabel)
}

func parseAndValidate(raw []byte, source string) (Config, error) {
	cfg := Default()
	if _, err := toml.Decode(string(raw), &cfg); err != nil {
		return cfg, fmt.Errorf("parse %s: %w", source, err)
	}
	cfg.applyDefaults()
	if err := cfg.validate(); err != nil {
		return cfg, fmt.Errorf("invalid config %s: %w", source, err)
	}
	return cfg, nil
}

func (c *Config) applyDefaults() {
	if strings.TrimSpace(c.Repo.Provider) == "" {
		c.Repo.Provider = "github"
	}
	if strings.TrimSpace(c.Packages.Mode) == "" {
		c.Packages.Mode = "single"
	}
	if strings.TrimSpace(c.Output.Markdown) == "" {
		c.Output.Markdown = "CHANGELOG.md"
	}
	if strings.TrimSpace(c.Output.JSON) == "" {
		c.Output.JSON = ".shipnote/release.json"
	}
}

func (c Config) validate() error {
	if c.Packages.Mode != "single" {
		return fmt.Errorf("packages.mode = %q (only \"single\" supported in v0.1)", c.Packages.Mode)
	}
	if c.Repo.Provider != "github" {
		return fmt.Errorf("repo.provider = %q (only \"github\" supported in v0.1)", c.Repo.Provider)
	}
	return nil
}

// ResolvedPath returns the absolute path of the config file used. If
// path is empty, falls back to DefaultPath relative to repoRoot.
func ResolvedPath(repoRoot, path string) string {
	if path == "" {
		path = DefaultPath
	}
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(repoRoot, path)
}
