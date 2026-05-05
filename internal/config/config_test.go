package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_MissingFileReturnsDefaults(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "does-not-exist.toml"))
	if err != nil {
		t.Fatalf("missing file should be ok, got: %v", err)
	}
	if cfg.Output.Markdown != "CHANGELOG.md" {
		t.Fatalf("default markdown path: %q", cfg.Output.Markdown)
	}
	if cfg.Packages.Mode != "single" {
		t.Fatalf("default packages.mode: %q", cfg.Packages.Mode)
	}
}

func TestLoadReader_FullExample(t *testing.T) {
	input := `
[repo]
provider = "github"
owner    = "Acme"
name     = "widgets"

[packages]
mode = "single"
root = "widgets"

[output]
markdown = "docs/CHANGELOG.md"
json     = ".shipnote/release.json"

[ignore]
authors = ["dependabot[bot]"]
labels  = ["skip-changelog"]

[[authors]]
canonical = "alice"
aliases   = ["alice@example.com", "Alice"]
name      = "Alice"
url       = "https://example.com/alice"
`
	cfg, err := LoadReader(strings.NewReader(input), "test.toml")
	if err != nil {
		t.Fatalf("LoadReader: %v", err)
	}
	if cfg.Repo.Owner != "Acme" || cfg.Repo.Name != "widgets" {
		t.Fatalf("repo: %+v", cfg.Repo)
	}
	if cfg.Output.Markdown != "docs/CHANGELOG.md" {
		t.Fatalf("markdown path: %s", cfg.Output.Markdown)
	}
	if len(cfg.Ignore.Authors) != 1 || cfg.Ignore.Authors[0] != "dependabot[bot]" {
		t.Fatalf("ignore.authors: %#v", cfg.Ignore.Authors)
	}
	if len(cfg.Authors) != 1 || cfg.Authors[0].Canonical != "alice" || len(cfg.Authors[0].Aliases) != 2 {
		t.Fatalf("authors: %#v", cfg.Authors)
	}
}

func TestValidate_RejectsMonorepoMode(t *testing.T) {
	input := `
[packages]
mode = "monorepo"
`
	_, err := LoadReader(strings.NewReader(input), "bad.toml")
	if err == nil || !strings.Contains(err.Error(), "packages.mode") {
		t.Fatalf("expected packages.mode rejection, got: %v", err)
	}
}

func TestValidate_RejectsNonGithubProvider(t *testing.T) {
	input := `
[repo]
provider = "gitlab"
`
	_, err := LoadReader(strings.NewReader(input), "bad.toml")
	if err == nil || !strings.Contains(err.Error(), "repo.provider") {
		t.Fatalf("expected repo.provider rejection, got: %v", err)
	}
}

func TestResolvedPath(t *testing.T) {
	if got := ResolvedPath("/repo", ""); got != "/repo/.shipnote.toml" {
		t.Fatalf("default: %s", got)
	}
	if got := ResolvedPath("/repo", "custom.toml"); got != "/repo/custom.toml" {
		t.Fatalf("relative: %s", got)
	}
	if got := ResolvedPath("/repo", "/abs/x.toml"); got != "/abs/x.toml" {
		t.Fatalf("absolute: %s", got)
	}
}
